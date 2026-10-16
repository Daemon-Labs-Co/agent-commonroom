package main

import (
	"agentroom/internal/toolbroker"
	"context"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestOwnedCodexPreparationFailuresFenceAndRelease(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"credential missing", "unsafe gateway", "invalid image"} {
		t.Run(mode, func(t *testing.T) {
			_, s, _ := workspacePair(t)
			r := s.record
			scope := toolbroker.Scope{Version: 1, JobID: r.JobID, Room: r.RoomID, Participant: r.ParticipantID, Connection: r.ConnectionID, Provider: r.Provider, RepositoryDigest: r.Pair.Repository, Deadline: r.Deadline}
			ctx := context.WithValue(s.ctx, workerLocalJobContextKey{}, scope.JobID)
			channel, peer := net.Pipe()
			defer peer.Close()
			runner := newWorkspaceCleanup(t, s)
			runner.volume = false
			session, token, err := newWorkerToolSession(ctx, scope, []string{"read"}, "ghcr.io/example/tool@sha256:"+strings.Repeat("a", 64), "/trusted/docker", time.Now(), channel, runner)
			if err != nil {
				t.Fatal(err)
			}
			defer session.Close()
			auth := t.TempDir()
			private := filepath.Join(auth, connectionAuthKey(scope.Connection), scope.Provider)
			if err = os.MkdirAll(private, 0700); err != nil {
				t.Fatal(err)
			}
			if mode != "credential missing" {
				if err = os.WriteFile(filepath.Join(private, "auth.json"), []byte(syntheticEngineCredential), 0600); err != nil {
					t.Fatal(err)
				}
			}
			gateway := shortGatewayRoot(t)
			image := "ghcr.io/example/engine@sha256:" + strings.Repeat("a", 64)
			if mode == "unsafe gateway" {
				if err = os.Chmod(gateway, 0755); err != nil {
					t.Fatal(err)
				}
			}
			if mode == "invalid image" {
				image = "unreviewed:latest"
			}
			bundle := &pairBundleFixture{Reader: strings.NewReader("synthetic-bundle")}
			_, record, err := s.RunCodex(t.Context(), session, token, bundle, workerCodexRunConfig{EngineImage: image, AuthRoot: auth, GatewayRoot: gateway, Prompt: "synthetic", Model: "model"}, runner)
			if err == nil || !record.ShutdownConfirmed || !s.closed {
				t.Fatalf("preparation failure did not fence: %v %+v", err, record)
			}
			if s.engineCredential != nil {
				t.Fatal("settled credential retained by supervisor")
			}
			if bundle.closes.Load() != 1 {
				t.Fatal("bundle not closed exactly once", bundle.closes.Load())
			}
			unlock, err := lockWorkerCredential(private)
			if err != nil {
				t.Fatal("connection remained locked", err)
			}
			unlock()
			if entries, err := os.ReadDir(gateway); err != nil || len(entries) != 0 {
				t.Fatal("gateway directory leaked", err)
			}
			if err = s.BeginCreate(t.Context(), "engine"); err == nil {
				t.Fatal("failed job resumed")
			}
		})
	}
}

func TestOwnedCodexValidationRefusalRetainsCallerOwnership(t *testing.T) {
	t.Parallel()
	_, s, _ := workspacePair(t)
	bundle := &pairBundleFixture{Reader: strings.NewReader("synthetic-bundle")}
	if _, _, err := s.RunCodex(t.Context(), nil, "", bundle, workerCodexRunConfig{}, &pairFakeDocker{}); err == nil {
		t.Fatal("invalid dependencies accepted")
	}
	if s.closed || s.codexRunStarted || bundle.closes.Load() != 0 {
		t.Fatal("refused request consumed caller resources")
	}
}

func TestSupervisorRetainsCredentialUntilBothRolesAreFenced(t *testing.T) {
	t.Parallel()
	_, s, adapter := workspacePair(t)
	r := s.record
	scope := toolbroker.Scope{Version: 1, JobID: r.JobID, Room: r.RoomID, Participant: r.ParticipantID, Connection: r.ConnectionID, Provider: r.Provider, RepositoryDigest: r.Pair.Repository, Deadline: r.Deadline}
	private := t.TempDir()
	unlock, err := lockWorkerCredential(private)
	if err != nil {
		t.Fatal(err)
	}
	lease := &workerEngineCredential{scope: scope, original: s.ctx, raw: []byte(syntheticEngineCredential), unlock: unlock}
	s.engineCredential = lease
	defer unlock()
	if err = s.Start(t.Context(), adapter); err != nil {
		t.Fatal(err)
	}
	runner := newWorkspaceCleanup(t, s)
	for role, id := range map[string]string{"engine": s.record.Pair.Engine.ID, "tool": s.record.Pair.Tool.ID} {
		runner.containers[id] = pairFakeContainer{name: pairContainerName(s.record.JobID, role), execution: "separated-" + role + "-v1", repository: s.record.Pair.Repository, job: s.record.JobID}
	}
	runner.failRemove = s.record.Pair.Engine.ID
	record, err := s.Shutdown(t.Context(), runner)
	if err == nil || record.Pair.Engine.Absent || s.engineCredential != lease || lease.released {
		t.Fatal("unresolved engine released credential custody")
	}
	if release, e := lockWorkerCredential(private); e == nil {
		release()
		t.Fatal("unresolved engine no longer locked connection")
	}
	runner.failRemove = ""
	record, err = s.Shutdown(t.Context(), runner)
	if err == nil || record.ShutdownConfirmed || !record.Pair.Engine.Absent || !record.Pair.Tool.Absent {
		t.Fatal("retained workspace or independent role proof lost", err)
	}
	if s.engineCredential != nil || !lease.released {
		t.Fatal("fenced engines retained credential lock")
	}
	release, err := lockWorkerCredential(private)
	if err != nil {
		t.Fatal(err)
	}
	release()
}

type failedCodexBundleClose struct {
	io.Reader
	err error
}

func (b *failedCodexBundleClose) Close() error { return b.err }
func TestOwnedBundlePreservesFirstCloseError(t *testing.T) {
	t.Parallel()
	expected := errors.New("synthetic bundle close failure")
	bundle := &workerPairBundle{ReadCloser: &failedCodexBundleClose{Reader: strings.NewReader("synthetic"), err: expected}}
	for i := 0; i < 2; i++ {
		if !errors.Is(bundle.Close(), expected) {
			t.Fatal("repeated close hid resource failure")
		}
	}
}
