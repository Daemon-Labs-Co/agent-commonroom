package main

import (
	"agentroom/internal/toolbroker"
	"bytes"
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const syntheticEngineCredential = `{"auth_mode":"chatgpt","OPENAI_API_KEY":null,"tokens":{"access_token":"synthetic-not-a-real-token","refresh_token":"synthetic-refresh"}}`

func credentialLeaseFixture(t *testing.T) (context.Context, context.CancelFunc, toolbroker.Scope, string, string, workerJobRecord, string) {
	t.Helper()
	journal, s, _ := workspacePair(t)
	r := s.record
	scope := toolbroker.Scope{Version: 1, JobID: r.JobID, Room: r.RoomID, Participant: r.ParticipantID, Connection: r.ConnectionID, Provider: r.Provider, RepositoryDigest: r.Pair.Repository, Deadline: r.Deadline}
	ctx, cancel := context.WithCancel(context.WithValue(t.Context(), workerLocalJobContextKey{}, r.JobID))
	t.Cleanup(cancel)
	root := t.TempDir()
	private := filepath.Join(root, connectionAuthKey(scope.Connection), scope.Provider)
	if err := os.MkdirAll(private, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(private, "auth.json"), []byte(syntheticEngineCredential), 0600); err != nil {
		t.Fatal(err)
	}
	return ctx, cancel, scope, root, private, r, journal
}
func TestEngineCredentialLeaseSerializesAndRequiresFencing(t *testing.T) {
	t.Parallel()
	ctx, cancel, scope, root, private, record, journal := credentialLeaseFixture(t)
	lease, err := acquireWorkerEngineCredential(ctx, scope, root, journal)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { lease.Release(record) }()
	var out bytes.Buffer
	if err = lease.Provision(ctx, &out); err != nil || !bytes.Equal(out.Bytes(), framedSyntheticCredential()) {
		t.Fatal("snapshot mismatch", err)
	}
	out.Reset()
	if _, err = acquireWorkerEngineCredential(ctx, scope, root, journal); err == nil {
		t.Fatal("parallel engine acquired same connection")
	}
	unlock, err := lockWorkerCredential(private)
	if err == nil {
		unlock()
		t.Fatal("sign-in/sign-out bypassed engine lease")
	}
	// An uncertain creation must keep the lock even after original cancellation.
	uncertain := record
	pair := *record.Pair
	uncertain.Pair = &pair
	uncertain.Pair.Engine.CreationStarted = true
	if lease.Release(uncertain) == nil {
		t.Fatal("uncertain engine released credentials")
	}
	cancel()
	fresh := context.WithValue(t.Context(), workerLocalJobContextKey{}, scope.JobID)
	if lease.Provision(fresh, &out) == nil || out.Len() != 0 {
		t.Fatal("fresh context revived cancelled snapshot")
	}
	unlock, err = lockWorkerCredential(private)
	if err == nil {
		unlock()
		t.Fatal("cancellation released live credentials")
	}
	foreign := record
	foreign.JobID = strings.Repeat("f", 32)
	if lease.Release(foreign) == nil {
		t.Fatal("foreign fence released lease")
	}
	if err = lease.Release(record); err != nil {
		t.Fatal(err)
	}
	if len(lease.raw) != 0 || !lease.released {
		t.Fatal("released bytes retained")
	}
	if err = lease.Release(record); err != nil {
		t.Fatal("release not idempotent")
	}
	unlock, err = lockWorkerCredential(private)
	if err != nil {
		t.Fatal("fenced lease not released", err)
	}
	unlock()
	if lease.Provision(fresh, &out) == nil {
		t.Fatal("released snapshot reused")
	}
}
func TestEngineCredentialRejectsUnsafeInputAndUnlocksFailure(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"api", "duplicate", "alias", "unknown", "null tokens", "invalid encoding", "world readable", "symlink", "oversize", "wrong job"} {
		t.Run(mode, func(t *testing.T) {
			ctx, _, scope, root, private, record, journal := credentialLeaseFixture(t)
			path := filepath.Join(private, "auth.json")
			raw := syntheticEngineCredential
			switch mode {
			case "api":
				raw = `{"auth_mode":"api","OPENAI_API_KEY":"synthetic-api"}`
			case "duplicate":
				raw = `{"auth_mode":"api","auth_mode":"chatgpt","tokens":{"access_token":"synthetic"}}`
			case "alias":
				raw = `{"Auth_Mode":"chatgpt","tokens":{"access_token":"synthetic"}}`
			case "unknown":
				raw = `{"auth_mode":"chatgpt","tokens":{"access_token":"synthetic","unreviewed":true}}`
			case "null tokens":
				raw = `{"auth_mode":"chatgpt","tokens":null}`
			case "invalid encoding":
				raw = "{\"auth_mode\":\"chatgpt\",\"tokens\":{\"access_token\":\"\xff\"}}"
			case "world readable":
				if err := os.Chmod(path, 0644); err != nil {
					t.Fatal(err)
				}
			case "symlink":
				if err := os.Rename(path, path+".target"); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(path+".target", path); err != nil {
					t.Fatal(err)
				}
			case "oversize":
				raw = strings.Repeat("x", 65537)
			case "wrong job":
				ctx = context.WithValue(t.Context(), workerLocalJobContextKey{}, strings.Repeat("f", 32))
			}
			if mode != "symlink" {
				if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if lease, err := acquireWorkerEngineCredential(ctx, scope, root, journal); err == nil {
				lease.Release(record)
				t.Fatal("unsafe credential accepted")
			}
			unlock, err := lockWorkerCredential(private)
			if err != nil {
				t.Fatal("failed admission leaked lock", err)
			}
			unlock()
		})
	}
}

func TestEngineCredentialReleaseRequiresBothCreatedRolesAbsent(t *testing.T) {
	t.Parallel()
	ctx, _, scope, root, _, record, journal := credentialLeaseFixture(t)
	lease, err := acquireWorkerEngineCredential(ctx, scope, root, journal)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Release(record)
	proof := record
	pair := *record.Pair
	proof.Pair = &pair
	proof.Status = "shutdown_unconfirmed"
	proof.DiagnosticCode = "container_shutdown_unverified"
	proof.Pair.Engine.CreationStarted = true
	proof.Pair.Tool.CreationStarted = true
	for _, absent := range []string{"neither", "engine", "tool"} {
		proof.Pair.Engine.Absent = absent == "engine"
		proof.Pair.Tool.Absent = absent == "tool"
		if lease.Release(proof) == nil {
			t.Fatal("partial absence released credentials", absent)
		}
	}
	proof.Pair.Engine.Absent = true
	proof.Pair.Tool.Absent = true
	// A retained workspace may keep the overall job unresolved; absent engines
	// no longer justify blocking this connection's authentication operations.
	if err = lease.Release(proof); err != nil {
		t.Fatal("both-role proof refused", err)
	}
}

func framedSyntheticCredential() []byte {
	raw := []byte(syntheticEngineCredential)
	header := []byte("COMMONROOM-AUTH-V1\n")
	size := make([]byte, 4)
	binary.BigEndian.PutUint32(size, uint32(len(raw)))
	return append(append(header, size...), raw...)
}

type syntheticShortCredentialWriter struct{ calls int }

func (w *syntheticShortCredentialWriter) Write(p []byte) (int, error) {
	w.calls++
	return len(p) - 1, nil
}
func TestEngineCredentialProvisionNeverReplaysPartialFrame(t *testing.T) {
	t.Parallel()
	ctx, _, scope, root, _, record, journal := credentialLeaseFixture(t)
	lease, err := acquireWorkerEngineCredential(ctx, scope, root, journal)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Release(record)
	writer := &syntheticShortCredentialWriter{}
	if lease.Provision(ctx, writer) == nil {
		t.Fatal("short write accepted")
	}
	var retry bytes.Buffer
	if lease.Provision(ctx, &retry) == nil || retry.Len() != 0 || writer.calls != 1 {
		t.Fatal("partial credential frame replayed")
	}
}

func TestEngineCredentialJournalBlocksCrashSurvivors(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"unresolved engine", "unresolved tool", "expired unresolved", "settled roles", "other connection", "missing current", "current creation", "malformed journal", "foreign current scope"} {
		t.Run(mode, func(t *testing.T) {
			ctx, _, scope, root, private, current, journal := credentialLeaseFixture(t)
			prior := current
			pair := *current.Pair
			prior.Pair = &pair
			prior.JobID = strings.Repeat("f", 32)
			prior.Status = "shutdown_unconfirmed"
			prior.DiagnosticCode = "container_shutdown_unverified"
			prior.Pair.Engine.CreationStarted = true
			switch mode {
			case "unresolved tool":
				prior.Pair.Engine.CreationStarted = false
				prior.Pair.Tool.CreationStarted = true
			case "expired unresolved":
				prior.StartedAt = prior.StartedAt.Add(-time.Hour)
				prior.Deadline = prior.Deadline.Add(-time.Hour)
				prior.UpdatedAt = prior.StartedAt
			case "settled roles":
				prior.Pair.Engine.Absent = true
			case "other connection":
				prior.ConnectionID = "other-connection"
				_, _, _, prior.Scope, _ = workerMembership(prior.RoomID + "/" + prior.ParticipantID + "/" + prior.ConnectionID + "/worker")
			case "missing current":
				if err := os.Remove(filepath.Join(journal, current.JobID+".json")); err != nil {
					t.Fatal(err)
				}
			case "current creation":
				current.Pair.Engine.CreationStarted = true
				if err := writeWorkerRecord(journal, current, false); err != nil {
					t.Fatal(err)
				}
			case "malformed journal":
				if err := os.WriteFile(filepath.Join(journal, strings.Repeat("e", 32)+".json"), []byte("{}"), 0600); err != nil {
					t.Fatal(err)
				}
			case "foreign current scope":
				scope.Room = "foreign-room"
			}
			if mode == "unresolved engine" || mode == "unresolved tool" || mode == "expired unresolved" || mode == "settled roles" || mode == "other connection" {
				if err := writeWorkerRecord(journal, prior, true); err != nil {
					t.Fatal(err)
				}
			}
			lease, err := acquireWorkerEngineCredential(ctx, scope, root, journal)
			allow := mode == "settled roles" || mode == "other connection"
			if allow {
				if err != nil {
					t.Fatal("settled or independent connection refused", err)
				}
				if err = lease.Release(current); err != nil {
					t.Fatal(err)
				}
			} else if err == nil {
				lease.Release(current)
				t.Fatal("unresolved or invalid journal admitted")
			}
			unlock, err := lockWorkerCredential(private)
			if err != nil {
				t.Fatal("journal refusal leaked connection lock", err)
			}
			unlock()
		})
	}
}
