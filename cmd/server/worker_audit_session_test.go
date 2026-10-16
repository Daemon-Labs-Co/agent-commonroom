package main

import (
	"agentroom/internal/toolbroker"
	"context"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestWatchedWorkerAuditLinkDetectsIdleEOF(t *testing.T) {
	t.Parallel()
	raw, peer := net.Pipe()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	link := newWatchedWorkerAuditLink(ctx, raw, cancel)
	defer link.Close()
	peer.Close()
	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("idle EOF did not cancel lifetime")
	}
	awaitAuditResult(t, closedAuditDone(link.done))
	if _, err := link.Write([]byte("after-drop")); err == nil {
		t.Fatal("disconnected link wrote")
	}
}
func closedAuditDone(done <-chan struct{}) <-chan error {
	result := make(chan error, 1)
	go func() { <-done; result <- nil }()
	return result
}
func TestWatchedWorkerAuditLinkPreservesFramedAcknowledgements(t *testing.T) {
	t.Parallel()
	scope := sampleToolScope(t)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	raw, peer := net.Pipe()
	defer peer.Close()
	var failures atomic.Int32
	link := newWatchedWorkerAuditLink(ctx, raw, func() { failures.Add(1); cancel() })
	defer link.Close()
	auditor := &transportTestAuditor{}
	served := make(chan error, 1)
	go func() { served <- toolbroker.ServeAudit(ctx, peer, scope, auditor) }()
	client, err := toolbroker.NewAuditClient(link, scope)
	if err != nil {
		t.Fatal(err)
	}
	intent := toolbroker.Intent{Scope: scope, CallID: "synthetic-call", Operation: "read", PayloadDigest: strings.Repeat("a", 64)}
	if err = client.RecordIntent(ctx, intent); err != nil {
		t.Fatal(err)
	}
	if err = client.RecordOutcome(ctx, intent, toolbroker.Outcome{Status: "succeeded", OutputDigest: strings.Repeat("b", 64)}); err != nil {
		t.Fatal(err)
	}
	if auditor.calls.Load() != 2 || failures.Load() != 0 {
		t.Fatal("acknowledgements lost or falsely disconnected")
	}
	peer.Close()
	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("idle disconnect not detected")
	}
	link.Close()
	link.Close()
	if failures.Load() != 1 {
		t.Fatal("failure repeated", failures.Load())
	}
	awaitAuditResult(t, served)
}
func TestWatchedWorkerAuditLinkCancellationJoinsBlockedForwarding(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	raw, peer := net.Pipe()
	defer peer.Close()
	link := newWatchedWorkerAuditLink(ctx, raw, cancel)
	written := make(chan error, 1)
	go func() { _, err := peer.Write([]byte("unconsumed-synthetic-frame")); written <- err }()
	// The raw write can finish while forwarding waits for the audit frame consumer.
	awaitAuditResult(t, written)
	cancel()
	closed := make(chan error, 1)
	go func() { closed <- link.Close() }()
	awaitAuditResult(t, closed)
}

func TestSupervisorAcceptToolSessionRevokesOnAuthenticatedLinkLoss(t *testing.T) {
	t.Parallel()
	dir, s, scope := auditTransportFixture(t)
	listener, err := s.OpenAuditListener(scope)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	controller, relay := net.Pipe()
	defer controller.Close()
	defer relay.Close()
	attached := make(chan error, 1)
	go func() { attached <- workerAttachAudit(s.ctx, dir, scope, relay, relay) }()
	session, token, err := s.AcceptToolSession(t.Context(), listener, []string{"read"}, "ghcr.io/example/tool@sha256:"+strings.Repeat("a", 64), "/trusted/docker", &pairFakeDocker{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if token == "" || session.Scope() != scope {
		t.Fatal("session not scope-bound")
	}
	controller.Close()
	select {
	case <-s.ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("controller loss did not revoke supervisor")
	}
	if _, err = session.Runtime.WorkspaceIntent(context.Background()); err == nil {
		t.Fatal("fresh context revived revoked runtime")
	}
	lease, free, err := workerTryLease(dir, ".job-"+scope.JobID+".lease")
	if lease != nil {
		lease.Close()
	}
	if err != nil || free {
		t.Fatal("disconnect released creator ownership", err)
	}
	awaitAuditResult(t, attached)
}
func TestSupervisorAcceptToolSessionRefusesForeignListenerWithoutConsumingInput(t *testing.T) {
	t.Parallel()
	dir, s, scope := auditTransportFixture(t)
	ctx := context.WithValue(s.ctx, workerLocalJobContextKey{}, scope.JobID)
	// Same scope is insufficient: only the listener minted by this supervisor is owned.
	foreign, err := newWorkerAuditListener(ctx, dir, scope)
	if err != nil {
		t.Fatal(err)
	}
	defer foreign.Close()
	if session, _, err := s.AcceptToolSession(t.Context(), foreign, []string{"read"}, "unused", "/trusted/docker", &pairFakeDocker{}); err == nil {
		session.Close()
		t.Fatal("unowned listener accepted")
	}
	if s.ctx.Err() != nil {
		t.Fatal("refused caller cancelled unrelated supervisor")
	}
}
func TestSupervisorAcceptToolSessionCancellationInterruptsHandshake(t *testing.T) {
	t.Parallel()
	_, s, scope := auditTransportFixture(t)
	listener, err := s.OpenAuditListener(scope)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() {
		session, _, err := s.AcceptToolSession(ctx, listener, []string{"read"}, "unused", "/trusted/docker", &pairFakeDocker{})
		if session != nil {
			session.Close()
		}
		done <- err
	}()
	cancel()
	awaitAuditResult(t, done)
	// Pre-admission cancellation need not consume the supervisor; once attachment
	// starts, cancellation closes the one-shot listener and aborts the lifetime.
}
func TestSupervisorFailedToolSessionConstructionRevokesLifetime(t *testing.T) {
	t.Parallel()
	dir, s, scope := auditTransportFixture(t)
	listener, err := s.OpenAuditListener(scope)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	controller, relay := net.Pipe()
	defer controller.Close()
	defer relay.Close()
	attached := make(chan error, 1)
	go func() { attached <- workerAttachAudit(s.ctx, dir, scope, relay, relay) }()
	session, _, err := s.AcceptToolSession(t.Context(), listener, []string{"read"}, "unreviewed:latest", "/trusted/docker", &pairFakeDocker{})
	if err == nil {
		session.Close()
		t.Fatal("invalid image accepted")
	}
	select {
	case <-s.ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("failed assembly left supervisor active")
	}
	awaitAuditResult(t, attached)
}

var _ io.ReadWriteCloser = (*watchedWorkerAuditLink)(nil)

type auditSessionBlockedDocker struct {
	entered chan struct{}
	once    sync.Once
}

func (d *auditSessionBlockedDocker) Run(ctx context.Context, _ []string, _ []byte, _ io.Writer) error {
	d.once.Do(func() { close(d.entered) })
	<-ctx.Done()
	return ctx.Err()
}
func TestAuditLinkLossInterruptsRealToolRuntimeBeforeEngineLaunch(t *testing.T) {
	t.Parallel()
	dir, s, scope := auditTransportFixture(t)
	listener, err := s.OpenAuditListener(scope)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	controller, relay := net.Pipe()
	defer controller.Close()
	defer relay.Close()
	attached := make(chan error, 1)
	go func() { attached <- workerAttachAudit(s.ctx, dir, scope, relay, relay) }()
	docker := &auditSessionBlockedDocker{entered: make(chan struct{})}
	session, _, err := s.AcceptToolSession(t.Context(), listener, []string{"read"}, "ghcr.io/example/tool@sha256:"+strings.Repeat("a", 64), "/trusted/docker", docker)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	input := &pairBundleFixture{Reader: strings.NewReader("synthetic-bundle")}
	engine := &pairLaunchFake{t: t, supervisor: s}
	assembly, err := s.AssemblePair(session, engine, input)
	if err != nil {
		t.Fatal(err)
	}
	started := make(chan error, 1)
	go func() { started <- s.Start(t.Context(), assembly) }()
	select {
	case <-docker.entered:
	case <-time.After(time.Second):
		t.Fatal("real runtime not reached")
	}
	controller.Close()
	select {
	case err = <-started:
		if err == nil {
			t.Fatal("startup survived link loss")
		}
	case <-time.After(time.Second):
		t.Fatal("startup did not cancel")
	}
	if len(engine.calls) != 0 {
		t.Fatal("engine launched after disconnect", engine.calls)
	}
	record, err := readWorkerRecord(dir, scope.JobID)
	if err != nil || record.Pair.Workspace == nil || record.ShutdownConfirmed {
		t.Fatal("lost allocation evidence or forged shutdown", err)
	}
	lease, free, err := workerTryLease(dir, ".job-"+scope.JobID+".lease")
	if lease != nil {
		lease.Close()
	}
	if err != nil || free {
		t.Fatal("failed startup released creator lease", err)
	}
	awaitAuditResult(t, attached)
}
