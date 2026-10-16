package main

import (
	"agentroom/internal/toolbroker"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type transportTestAuditor struct{ calls atomic.Int32 }

func (a *transportTestAuditor) RecordIntent(context.Context, toolbroker.Intent) error {
	a.calls.Add(1)
	return nil
}
func (a *transportTestAuditor) RecordOutcome(context.Context, toolbroker.Intent, toolbroker.Outcome) error {
	a.calls.Add(1)
	return nil
}

func auditTransportFixture(t *testing.T) (string, *workerPairSupervisor, toolbroker.Scope) {
	t.Helper()
	dir, err := os.MkdirTemp("/tmp", "cr-audit-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	request := pairRequest()
	request.Version = workerWorkspaceLifecycleVersion
	s, err := admitWorkerPair(t.Context(), dir, request, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Close)
	r := s.record
	scope := toolbroker.Scope{Version: 1, JobID: r.JobID, Room: r.RoomID, Participant: r.ParticipantID, Connection: r.ConnectionID, Provider: r.Provider, RepositoryDigest: r.Pair.Repository, Deadline: r.Deadline}
	return dir, s, scope
}
func awaitAuditResult(t *testing.T, ch <-chan error) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("audit transport did not stop")
	}
}
func TestWorkerAuditTransportFramedRoundTripAndCancellation(t *testing.T) {
	t.Parallel()
	dir, _, scope := auditTransportFixture(t)
	ctx, cancel := context.WithTimeout(context.WithValue(t.Context(), workerLocalJobContextKey{}, scope.JobID), 3*time.Second)
	defer cancel()
	listener, err := newWorkerAuditListener(ctx, dir, scope)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	controller, relay := net.Pipe()
	defer controller.Close()
	defer relay.Close()
	attachDone := make(chan error, 1)
	go func() { attachDone <- workerAttachAudit(ctx, dir, scope, relay, relay) }()
	link, err := listener.Accept(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer link.Close()
	auditor := &transportTestAuditor{}
	serveDone := make(chan error, 1)
	go func() { serveDone <- toolbroker.ServeAudit(ctx, controller, scope, auditor) }()
	client, err := toolbroker.NewAuditClient(link, scope)
	if err != nil {
		t.Fatal(err)
	}
	intent := toolbroker.Intent{Scope: scope, CallID: "synthetic-call", Operation: "read", PayloadDigest: strings.Repeat("a", 64)}
	if err = client.RecordIntent(ctx, intent); err != nil {
		t.Fatal("intent", err)
	}
	outcome := toolbroker.Outcome{Status: "succeeded", OutputDigest: strings.Repeat("b", 64), OutputBytes: 0}
	if err = client.RecordOutcome(ctx, intent, outcome); err != nil {
		t.Fatal("outcome", err)
	}
	if auditor.calls.Load() != 2 {
		t.Fatal("missing durable callbacks")
	}
	if _, err = listener.Accept(ctx); err == nil {
		t.Fatal("socket reused")
	}
	cancel()
	awaitAuditResult(t, attachDone)
	awaitAuditResult(t, serveDone)
	if client.RecordIntent(context.Background(), intent) == nil {
		t.Fatal("cancelled link revived")
	}
}
func TestWorkerAuditSocketRefusesPublicDirectoryAndExistingPath(t *testing.T) {
	t.Parallel()
	dir, _, scope := auditTransportFixture(t)
	ctx := context.WithValue(t.Context(), workerLocalJobContextKey{}, scope.JobID)
	os.Chmod(dir, 0755)
	if l, err := newWorkerAuditListener(ctx, dir, scope); err == nil {
		l.Close()
		t.Fatal("public directory accepted")
	}
	os.Chmod(dir, 0700)
	path := workerAuditSocket(dir, scope.JobID)
	if err := os.WriteFile(path, []byte("retain"), 0600); err != nil {
		t.Fatal(err)
	}
	if l, err := newWorkerAuditListener(ctx, dir, scope); err == nil {
		l.Close()
		t.Fatal("existing socket path adopted")
	}
	raw, _ := os.ReadFile(path)
	if string(raw) != "retain" {
		t.Fatal("existing path modified")
	}
	symlink := filepath.Join(dir, "alias")
	os.Symlink(dir, symlink)
	if l, err := newWorkerAuditListener(ctx, symlink, scope); err == nil {
		l.Close()
		t.Fatal("symlink directory accepted")
	}
}
func TestWorkerAuditScopeRequiresExactRecordAndCanonicalRequest(t *testing.T) {
	t.Parallel()
	_, s, scope := auditTransportFixture(t)
	if !validWorkerAuditScope(s.record, scope) {
		t.Fatal("valid scope refused")
	}
	for _, field := range []string{"job", "room", "participant", "connection", "provider", "repository", "deadline"} {
		changed := scope
		switch field {
		case "job":
			changed.JobID = strings.Repeat("f", 32)
		case "room":
			changed.Room = "other"
		case "participant":
			changed.Participant = "other"
		case "connection":
			changed.Connection = "other"
		case "provider":
			changed.Provider = "claude"
		case "repository":
			changed.RepositoryDigest = strings.Repeat("f", 64)
		case "deadline":
			changed.Deadline = changed.Deadline.Add(time.Second)
		}
		if validWorkerAuditScope(s.record, changed) {
			t.Fatal("scope mismatch accepted", field)
		}
	}
	raw, _ := json.Marshal(scope)
	encoded := base64.RawURLEncoding.EncodeToString(raw)
	if _, err := decodeWorkerAuditScope(encoded); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{string(raw) + " ", strings.Replace(string(raw), `"version":1`, `"version":1,"version":1`, 1), strings.Replace(string(raw), `"version":1`, `"version":1,"extra":true`, 1)} {
		if _, err := decodeWorkerAuditScope(base64.RawURLEncoding.EncodeToString([]byte(bad))); err == nil {
			t.Fatal("noncanonical scope accepted")
		}
	}
	r := s.record
	r.Version = workerPairLifecycleVersion
	r.Execution = workerPairExecution
	if validWorkerAuditScope(r, scope) {
		t.Fatal("legacy pair attached")
	}
}
func TestWorkerAuditAttachRequiresHeldCreatorLease(t *testing.T) {
	t.Parallel()
	dir, s, scope := auditTransportFixture(t)
	s.Close()
	input, output := net.Pipe()
	defer input.Close()
	defer output.Close()
	if workerAttachAudit(t.Context(), dir, scope, input, input) == nil {
		t.Fatal("abandoned supervisor attached")
	}
}
func TestWorkerAuditWrongHandshakeConsumesLink(t *testing.T) {
	t.Parallel()
	dir, _, scope := auditTransportFixture(t)
	ctx, cancel := context.WithTimeout(context.WithValue(t.Context(), workerLocalJobContextKey{}, scope.JobID), time.Second)
	defer cancel()
	l, err := newWorkerAuditListener(ctx, dir, scope)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	done := make(chan error, 1)
	go func() {
		conn, err := l.Accept(ctx)
		if conn != nil {
			conn.Close()
		}
		done <- err
	}()
	peer, err := net.Dial("unix", workerAuditSocket(dir, scope.JobID))
	if err != nil {
		t.Fatal(err)
	}
	defer peer.Close()
	peer.Write(make([]byte, 32))
	select {
	case err = <-done:
		if err == nil {
			t.Fatal("forged scope proof accepted")
		}
	case <-ctx.Done():
		t.Fatal("handshake blocked")
	}
	if _, err = l.Accept(ctx); err == nil {
		t.Fatal("failed handshake retried")
	}
}
func TestWorkerAuditRelayCancelsBlockedBothDirections(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	local, peer := net.Pipe()
	defer peer.Close()
	input, other := net.Pipe()
	defer other.Close()
	done := make(chan error, 1)
	go func() { done <- relayWorkerAudit(ctx, local, input, input) }()
	cancel()
	awaitAuditResult(t, done)
}

var _ io.ReadWriteCloser = (*workerAuditConn)(nil)

func TestWorkerAuditClosedSocketCannotBeRecreated(t *testing.T) {
	t.Parallel()
	_, s, scope := auditTransportFixture(t)
	l, err := s.OpenAuditListener(scope)
	if err != nil {
		t.Fatal(err)
	}
	l.Close()
	if next, err := s.OpenAuditListener(scope); err == nil {
		next.Close()
		t.Fatal("consumed audit job reopened")
	}
}
func TestWorkerAuditListenerOriginalLifetimeCannotBeRevived(t *testing.T) {
	t.Parallel()
	dir, _, scope := auditTransportFixture(t)
	parent, cancel := context.WithCancel(context.WithValue(t.Context(), workerLocalJobContextKey{}, scope.JobID))
	l, err := newWorkerAuditListener(parent, dir, scope)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	accepted := make(chan io.ReadWriteCloser, 1)
	go func() { conn, _ := l.Accept(context.Background()); accepted <- conn }()
	peer, err := net.Dial("unix", workerAuditSocket(dir, scope.JobID))
	if err != nil {
		t.Fatal(err)
	}
	defer peer.Close()
	peer.Write(auditScopeProof(scope))
	ack := make([]byte, 1)
	if _, err = io.ReadFull(peer, ack); err != nil {
		t.Fatal(err)
	}
	conn := <-accepted
	if conn == nil {
		t.Fatal("valid handshake failed")
	}
	defer conn.Close()
	cancel()
	peer.SetReadDeadline(time.Now().Add(time.Second))
	if _, err = peer.Read(ack); err == nil {
		t.Fatal("original lifetime revived")
	}
}
func TestAuditCommandChild(t *testing.T) {
	if os.Getenv("COMMONROOM_TEST_AUDIT_COMMAND") != "true" {
		return
	}
	io.Copy(os.Stdout, os.Stdin)
	os.Exit(0)
}
func TestCommandAuditLinkOwnsAndReapsDedicatedProcess(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestAuditCommandChild$")
	command.Env = append(os.Environ(), "COMMONROOM_TEST_AUDIT_COMMAND=true")
	link, err := startCommandAuditLink(command, cancel)
	if err != nil {
		t.Fatal(err)
	}
	defer link.Close()
	done := make(chan error, 1)
	go func() { _, err := link.Write([]byte("synthetic-frame")); done <- err }()
	raw := make([]byte, len("synthetic-frame"))
	if _, err = io.ReadFull(link, raw); err != nil || string(raw) != "synthetic-frame" {
		t.Fatal("dedicated stream corrupted", err)
	}
	awaitAuditResult(t, done)
	link.Close()
	link.Close()
	if command.ProcessState == nil {
		t.Fatal("SSH substitute process not reaped")
	}
	if _, err = link.Write([]byte("after-close")); err == nil {
		t.Fatal("closed process revived")
	}
}

func TestWorkerAuditListenerCloseInterruptsAcceptedRead(t *testing.T) {
	t.Parallel()
	dir, _, scope := auditTransportFixture(t)
	ctx := context.WithValue(t.Context(), workerLocalJobContextKey{}, scope.JobID)
	l, err := newWorkerAuditListener(ctx, dir, scope)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	accepted := make(chan io.ReadWriteCloser, 1)
	go func() { conn, _ := l.Accept(ctx); accepted <- conn }()
	peer, err := net.Dial("unix", workerAuditSocket(dir, scope.JobID))
	if err != nil {
		t.Fatal(err)
	}
	defer peer.Close()
	peer.Write(auditScopeProof(scope))
	ack := make([]byte, 1)
	if _, err = io.ReadFull(peer, ack); err != nil {
		t.Fatal(err)
	}
	conn := <-accepted
	if conn == nil {
		t.Fatal("valid handshake refused")
	}
	defer conn.Close()
	done := make(chan error, 1)
	go func() { _, err := conn.Read(make([]byte, 1)); done <- err }()
	l.Close()
	awaitAuditResult(t, done)
}
func TestWorkerAuditListenerDeadlineBoundsUnattachedJob(t *testing.T) {
	t.Parallel()
	dir, _, scope := auditTransportFixture(t)
	scope.Deadline = time.Now().Add(30 * time.Millisecond)
	ctx := context.WithValue(t.Context(), workerLocalJobContextKey{}, scope.JobID)
	l, err := newWorkerAuditListener(ctx, dir, scope)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	done := make(chan error, 1)
	go func() {
		conn, err := l.Accept(context.Background())
		if conn != nil {
			conn.Close()
		}
		done <- err
	}()
	awaitAuditResult(t, done)
}
