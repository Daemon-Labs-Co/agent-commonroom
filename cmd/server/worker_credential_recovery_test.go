package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type credentialOrderWriter struct {
	t    *testing.T
	root string
	bytes.Buffer
}

func (w *credentialOrderWriter) Write(p []byte) (int, error) {
	w.t.Helper()
	if workerCredentialReady(w.root) {
		w.t.Fatal("credential frame preceded durable recovery intent")
	}
	return w.Buffer.Write(p)
}

func TestCredentialRecoveryIntentPrecedesProvisionAndSurvivesFencing(t *testing.T) {
	t.Parallel()
	ctx, _, scope, root, private, record, journal := credentialLeaseFixture(t)
	lease, err := acquireWorkerEngineCredential(ctx, scope, root, journal)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Release(record)
	out := &credentialOrderWriter{t: t, root: private}
	if err = lease.Provision(ctx, out); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out.Bytes(), framedSyntheticCredential()) {
		t.Fatal("credential frame changed")
	}
	marker := filepath.Join(private, workerCredentialRecoveryName)
	raw, err := readPrivateBounded(marker, 8192)
	if err != nil || strings.Contains(string(raw), "synthetic") {
		t.Fatal("invalid or secret-bearing recovery marker", err)
	}
	if credentialPresent(filepath.Join(private, "auth.json")) {
		t.Fatal("uncertain credential reported usable")
	}
	if err = lease.Release(record); err != nil {
		t.Fatal(err)
	}
	if workerCredentialReady(private) {
		t.Fatal("container fencing cleared token uncertainty")
	}
	if _, err = acquireWorkerEngineCredential(ctx, scope, root, journal); err == nil {
		t.Fatal("crash/fencing allowed stale credential reuse")
	}
	if !workerCredentialConnectionQuiescent(journal, scope.Connection, scope.Provider) {
		t.Fatal("quiescent roles cannot be reauthenticated")
	}
}

func TestCredentialRecoveryMarkerFailureSendsNoBytes(t *testing.T) {
	t.Parallel()
	ctx, _, scope, root, private, record, journal := credentialLeaseFixture(t)
	lease, err := acquireWorkerEngineCredential(ctx, scope, root, journal)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Release(record)
	// A directory cannot be replaced by the private file; errors must not expose
	// a credential or allow retry to bypass the refused persistence step.
	if err = os.Mkdir(filepath.Join(private, workerCredentialRecoveryName), 0700); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if lease.Provision(ctx, &out) == nil || out.Len() != 0 {
		t.Fatal("marker failure supplied credentials")
	}
	if lease.Provision(ctx, &out) == nil || out.Len() != 0 {
		t.Fatal("marker failure was replayed")
	}
}

func TestCredentialRecoveryFailuresRemainScopedAndClosed(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"file", "directory", "dangling symlink", "corrupt"} {
		t.Run(mode, func(t *testing.T) {
			ctx, _, scope, root, private, _, journal := credentialLeaseFixture(t)
			marker := filepath.Join(private, workerCredentialRecoveryName)
			var err error
			switch mode {
			case "directory":
				err = os.Mkdir(marker, 0700)
			case "dangling symlink":
				err = os.Symlink(filepath.Join(private, "absent"), marker)
			default:
				err = os.WriteFile(marker, []byte("unknown recovery state"), 0600)
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, err = acquireWorkerEngineCredential(ctx, scope, root, journal); err == nil {
				t.Fatal("unrecognized marker permitted acquisition")
			}
			other := filepath.Join(root, connectionAuthKey("other-connection"), "codex")
			if err = os.MkdirAll(other, 0700); err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(filepath.Join(other, "auth.json"), []byte(syntheticEngineCredential), 0600); err != nil {
				t.Fatal(err)
			}
			if !credentialPresent(filepath.Join(other, "auth.json")) {
				t.Fatal("foreign connection was blocked")
			}
		})
	}
}

func TestCredentialRecoveryNeedsDurableCredentialBeforeMarkerClear(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"success", "credential sync failure", "marker sync failure", "invalid export"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			source := filepath.Join(t.TempDir(), "synthetic-export.json")
			if err := markWorkerCredentialUncertain(root); err != nil {
				t.Fatal(err)
			}
			raw := syntheticEngineCredential
			if mode == "invalid export" {
				raw = "not JSON"
			}
			if err := os.WriteFile(source, []byte(raw), 0600); err != nil {
				t.Fatal(err)
			}
			credentialSynced := false
			cleared := false
			sync := func(path string) error {
				if filepath.Base(path) == "auth.json" {
					if workerCredentialReady(root) {
						t.Fatal("recovery opened before credential sync")
					}
					if mode == "credential sync failure" {
						return errors.New("synthetic backend detail")
					}
					credentialSynced = true
				} else {
					if !credentialSynced {
						t.Fatal("marker clear preceded credential durability")
					}
					cleared = true
					if mode == "marker sync failure" {
						return errors.New("synthetic backend detail")
					}
				}
				return syncParentDirectory(path)
			}
			err := persistWorkerCredentialDurably(root, "codex", source, sync)
			if mode == "success" {
				if err != nil || !credentialPresent(filepath.Join(root, "auth.json")) || !cleared {
					t.Fatal("durable recovery refused", err)
				}
			} else {
				if err == nil || strings.Contains(err.Error(), "synthetic backend detail") {
					t.Fatal("failed custody acknowledged or diagnostic leaked", err)
				}
				if mode != "marker sync failure" && workerCredentialReady(root) {
					t.Fatal("failed credential persistence cleared marker")
				}
			}
		})
	}
}

func TestCredentialRecoveryJournalRequiresBothRolesQuiescent(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"engine intent", "tool intent", "fenced retained source", "foreign connection", "corrupt journal"} {
		t.Run(mode, func(t *testing.T) {
			_, _, scope, _, _, record, journal := credentialLeaseFixture(t)
			record.Pair.Engine.CreationStarted = true
			record.Pair.Tool.CreationStarted = true
			switch mode {
			case "engine intent":
				record.Pair.Tool.Absent = true
			case "tool intent":
				record.Pair.Engine.Absent = true
			case "fenced retained source":
				record.Pair.Engine.Absent = true
				record.Pair.Tool.Absent = true
			case "foreign connection":
				record.ConnectionID = "foreign-connection"
			}
			if mode == "corrupt journal" {
				if err := os.WriteFile(filepath.Join(journal, record.JobID+".json"), []byte("corrupt"), 0600); err != nil {
					t.Fatal(err)
				}
			} else if err := writeWorkerRecord(journal, record, false); err != nil {
				t.Fatal(err)
			}
			got := workerCredentialConnectionQuiescent(journal, scope.Connection, scope.Provider)
			want := mode == "fenced retained source" || mode == "foreign connection"
			if got != want {
				t.Fatalf("quiescent=%v want %v", got, want)
			}
		})
	}
}

func TestCredentialRecoverySyncFailureRemainsBlocked(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	synced := false
	err := markWorkerCredentialUncertainDurably(root, func(path string) error {
		synced = true
		if workerCredentialReady(root) {
			t.Fatal("marker not published before sync")
		}
		return errors.New("synthetic sensitive backend error")
	})
	if err == nil || !synced || workerCredentialReady(root) || strings.Contains(err.Error(), "synthetic sensitive") {
		t.Fatal("uncertain marker accepted or error leaked", err)
	}
}

func TestWorkerCredentialStatusNeverOpensUncertainCustody(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		raw                      string
		present, recovery, valid bool
	}{
		{`{"credentialPresent":true}`, true, false, true},
		{`{"credentialPresent":false}`, false, false, true},
		{`{"credentialPresent":true,"recoveryRequired":true}`, false, true, true},
		{`{"credentialPresent":false,"recoveryRequired":true}`, false, true, true},
		{`{"credentialPresent":true,"recoveryRequired":"synthetic-secret"}`, false, false, false},
		{`{"recoveryRequired":false}`, false, false, false},
	} {
		got, recovery, err := decodeWorkerCredentialStatus([]byte(tc.raw))
		if (err == nil) != tc.valid || got != tc.present || recovery != tc.recovery {
			t.Fatal("incorrect worker custody status", got, recovery, err)
		}
		if err != nil && strings.Contains(err.Error(), "synthetic-secret") {
			t.Fatal("raw credential status leaked")
		}
	}
}
