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
// No host credential mount is used. A durable recovery marker blocks stale-token
// reuse unless a valid final snapshot is saved after durable fencing, or an
// authorized human explicitly recovers the connection.
type workerEngineCredential struct {
	mu          sync.Mutex
	scope       toolbroker.Scope
	root        string
	journal     string
	refreshed   []byte
	original    context.Context
	raw         []byte
	unlock      func()
	released    bool
	provisioned bool
	releaseErr  error
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
	if !workerCredentialJournalReady(journalDir, scope) || !workerCredentialReady(root) {
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
	return &workerEngineCredential{scope: scope, root: root, journal: journalDir, original: ctx, raw: raw, unlock: unlock}, nil
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
		Mode        string          `json:"auth_mode"`
		APIKey      json.RawMessage `json:"OPENAI_API_KEY"`
		LastRefresh *string         `json:"last_refresh"`
		Tokens      struct {
			Access  string `json:"access_token"`
			Refresh string `json:"refresh_token"`
			ID      string `json:"id_token"`
			Account string `json:"account_id"`
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
	// A crash or partial write must not make the original snapshot reusable.
	if c.root == "" || markWorkerCredentialUncertain(c.root) != nil {
		return toolbroker.ErrExecution
	}
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

// StageRefresh holds one strict subscription snapshot from the private engine
// exporter. The lease, job, connection and account identity cannot be substituted.
// It is not persisted or exposed until both roles have durable absence proof.
func (c *workerEngineCredential) StageRefresh(scope toolbroker.Scope, raw []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.released || !c.provisioned || c.scope != scope || c.refreshed != nil || !validEngineSubscriptionCredential(raw) {
		return toolbroker.ErrRejected
	}
	var before, after struct {
		Tokens struct {
			Account string `json:"account_id"`
		} `json:"tokens"`
	}
	if json.Unmarshal(c.raw, &before) != nil || json.Unmarshal(raw, &after) != nil || before.Tokens.Account != after.Tokens.Account {
		return toolbroker.ErrRejected
	}
	c.refreshed = append([]byte(nil), raw...)
	return nil
}

// Release requires durable pair-fencing evidence for the original identity.
// Retained source volumes do not require retaining provider credentials once
// both execution roles are confirmed absent (or never began creation). No timer,
// cancelled context or process exit can substitute this evidence.
func (c *workerEngineCredential) Release(record workerJobRecord) error {
	return c.releaseWithPersistence(record, func(path string, raw []byte) error { return replacePrivateDurably(path, raw, syncParentDirectory) })
}

// Dependency injection is hermetic fault-test infrastructure, never deployment
// configuration or a model-selected persistence backend.
func (c *workerEngineCredential) releaseWithPersistence(record workerJobRecord, persist func(string, []byte) error) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if persist == nil || !validWorkerRecord(record) || !matchesWorkerCredentialScope(record, c.scope) || record.Pair == nil || !pairRoleSettled(record.Pair.Engine) || !pairRoleSettled(record.Pair.Tool) {
		return toolbroker.ErrRejected
	}
	if !c.released {
		if c.refreshed != nil {
			durable, readErr := readWorkerRecord(c.journal, c.scope.JobID)
			if readErr != nil || !validWorkerRecord(durable) || !matchesWorkerCredentialScope(durable, c.scope) || durable.Pair.Engine != record.Pair.Engine || durable.Pair.Tool != record.Pair.Tool {
				return toolbroker.ErrRejected
			}
		}
		// Persistence is attempted only after the original engine AND tool are
		// settled. Marker retention protects every failed/ambiguous operation.
		var custodyErr error
		if c.refreshed != nil {
			current, readErr := readPrivateBounded(filepath.Join(c.root, providerCredentialName(c.scope.Provider)), 65536)
			if readErr != nil || !bytes.Equal(current, c.raw) || workerCredentialReady(c.root) {
				custodyErr = toolbroker.ErrRejected
			} else if persist(filepath.Join(c.root, providerCredentialName(c.scope.Provider)), c.refreshed) != nil {
				custodyErr = toolbroker.ErrExecution
			} else if clearWorkerCredentialRecovery(c.root, syncParentDirectory) != nil {
				custodyErr = toolbroker.ErrExecution
			}
			clear(current)
		}
		clear(c.refreshed)
		c.refreshed = nil
		clear(c.raw)
		c.raw = nil
		c.released = true
		c.releaseErr = custodyErr
		c.unlock()
		return custodyErr
	}
	return c.releaseErr
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
