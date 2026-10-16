package main

import (
	"agentroom/internal/strictjson"
	"agentroom/internal/toolbroker"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"path/filepath"
	"slices"
	"sync"
	"time"
)

// workerEngineCredential is worker-only custody, not a model tool or controller
// payload. It serializes engine use with the existing sign-in/sign-out lock.
// Cancellation does NOT unlock: a still-running container may retain credentials.
// The supervisor must fence both roles before releasing. A source fork can leak
// these bytes or forge evidence; the worker account remains a trusted boundary.
// No host credential mount is used. Refreshed-token persistence remains separate.
type workerEngineCredential struct {
	mu          sync.Mutex
	scope       toolbroker.Scope
	original    context.Context
	raw         []byte
	unlock      func()
	released    bool
	provisioned bool
}

// acquireWorkerEngineCredential takes a bounded private snapshot while holding
// the connection lock. authRoot is trusted worker configuration, never a request
// path. The configured root and its ancestry must be secured by deployment.
func acquireWorkerEngineCredential(ctx context.Context, scope toolbroker.Scope, authRoot, journalDir string) (*workerEngineCredential, error) {
	job, ok := ctxJob(ctx)
	if ctx == nil || ctx.Err() != nil || !ok || job != scope.JobID || scope.Provider != "codex" || !toolbroker.ValidAuditScope(scope, time.Now()) || !workerAliasPattern.MatchString(scope.Connection) || !filepath.IsAbs(authRoot) || filepath.Clean(authRoot) != authRoot || authRoot == string(filepath.Separator) {
		return nil, toolbroker.ErrRejected
	}
	scope.Deadline = scope.Deadline.UTC().Round(0)
	root := filepath.Join(authRoot, connectionAuthKey(scope.Connection), scope.Provider)
	unlock, err := lockWorkerCredential(root)
	if err != nil {
		return nil, toolbroker.ErrRejected
	}
	accepted := false
	defer func() {
		if !accepted {
			unlock()
		}
	}()
	// Check after taking the shared connection lock. A process crash releases
	// flock, but cannot erase an unresolved creation intent in the journal.
	if !workerCredentialJournalReady(journalDir, scope) {
		return nil, toolbroker.ErrRejected
	}
	raw, err := readPrivateBounded(filepath.Join(root, providerCredentialName(scope.Provider)), 65536)
	if err != nil {
		return nil, toolbroker.ErrRejected
	}
	if !validEngineSubscriptionCredential(raw) || ctx.Err() != nil {
		clear(raw)
		return nil, toolbroker.ErrRejected
	}
	accepted = true
	return &workerEngineCredential{scope: scope, original: ctx, raw: raw, unlock: unlock}, nil
}

func workerCredentialJournalReady(dir string, scope toolbroker.Scope) bool {
	records, err := boundedWorkerRecords(dir)
	if err != nil {
		return false
	}
	found := false
	for _, record := range records {
		if !validWorkerRecord(record) {
			return false
		}
		if record.JobID == scope.JobID {
			if found || !validWorkerAuditScope(record, scope) || record.Pair.Engine.CreationStarted || record.Pair.Tool.CreationStarted {
				return false
			}
			found = true
			continue
		}
		if record.ConnectionID != scope.Connection || record.Provider != scope.Provider {
			continue
		}
		if record.Pair != nil {
			if !pairRoleSettled(record.Pair.Engine) || !pairRoleSettled(record.Pair.Tool) {
				return false
			}
		} else if !workerJobSettled(record) {
			return false
		}
	}
	return found
}

func validEngineSubscriptionCredential(raw []byte) bool {
	if len(raw) < 2 || len(raw) > 65536 || !strictjson.ValidEncoding(raw) {
		return false
	}
	fields, ok := engineCredentialFields(raw, []string{"auth_mode", "OPENAI_API_KEY", "tokens", "last_refresh"})
	if !ok {
		return false
	}
	if _, ok = engineCredentialFields(fields["tokens"], []string{"access_token", "refresh_token", "id_token", "account_id"}); !ok {
		return false
	}

	var auth struct {
		Mode   string          `json:"auth_mode"`
		APIKey json.RawMessage `json:"OPENAI_API_KEY"`
		Tokens struct {
			Access string `json:"access_token"`
		} `json:"tokens"`
	}
	if json.Unmarshal(raw, &auth) != nil || auth.Mode != "chatgpt" || auth.Tokens.Access == "" {
		return false
	}
	return len(auth.APIKey) == 0 || string(auth.APIKey) == "null"
}

// Provision sends one bounded private authentication frame before the callback
// protocol begins. Only pass the inspected engine attachment, never a model,
// audit, controller or log writer. Even partial failure consumes this attempt.
func (c *workerEngineCredential) Provision(ctx context.Context, out io.Writer) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	job, ok := ctxJob(ctx)
	if !ok || job != c.scope.JobID || c.original.Err() != nil || out == nil || c.released || c.provisioned || !toolbroker.ValidAuditScope(c.scope, time.Now()) {
		return toolbroker.ErrRejected
	}
	c.provisioned = true
	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, uint32(len(c.raw)))
	for _, data := range [][]byte{[]byte("COMMONROOM-AUTH-V1\n"), header, c.raw} {
		n, err := out.Write(data)
		if err != nil || n != len(data) {
			return toolbroker.ErrExecution
		}
	}
	return nil
}

// Release requires durable pair-fencing evidence for the original identity.
// Retained source volumes do not require retaining provider credentials once
// both execution roles are confirmed absent (or never began creation). No timer,
// cancelled context or process exit can substitute this evidence.
func (c *workerEngineCredential) Release(record workerJobRecord) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !validWorkerRecord(record) || !matchesWorkerCredentialScope(record, c.scope) || record.Pair == nil || !pairRoleSettled(record.Pair.Engine) || !pairRoleSettled(record.Pair.Tool) {
		return toolbroker.ErrRejected
	}
	if !c.released {
		clear(c.raw)
		c.raw = nil
		c.released = true
		c.unlock()
	}
	return nil
}

func matchesWorkerCredentialScope(record workerJobRecord, scope toolbroker.Scope) bool {
	return record.Version == workerWorkspaceLifecycleVersion && record.JobID == scope.JobID && record.RoomID == scope.Room && record.ParticipantID == scope.Participant && record.ConnectionID == scope.Connection && record.Provider == scope.Provider && record.Pair != nil && record.Pair.Repository == scope.RepositoryDigest && record.Deadline.Equal(scope.Deadline)
}

// Exact top-level and token objects reject duplicate and case-aliased fields.
// Unknown schema fields require review before a newer client can be admitted.
func engineCredentialFields(raw []byte, allowed []string) (map[string]json.RawMessage, bool) {
	d := json.NewDecoder(bytes.NewReader(raw))
	start, err := d.Token()
	if err != nil || start != json.Delim('{') {
		return nil, false
	}
	fields := map[string]json.RawMessage{}
	for d.More() {
		key, err := d.Token()
		name, ok := key.(string)
		if err != nil || !ok || !slices.Contains(allowed, name) {
			return nil, false
		}
		if _, exists := fields[name]; exists {
			return nil, false
		}
		var value json.RawMessage
		if d.Decode(&value) != nil {
			return nil, false
		}
		fields[name] = value
	}
	end, err := d.Token()
	if err != nil || end != json.Delim('}') {
		return nil, false
	}
	_, err = d.Token()
	return fields, err == io.EOF
}
