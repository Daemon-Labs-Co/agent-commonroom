package main

import (
	"agentroom/internal/toolbroker"
	"agentroom/internal/workspacebootstrap"
	"context"
	"io"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type assembledToolFake struct{ *pairLaunchFake }

func (f assembledToolFake) Create(ctx context.Context) (string, error) {
	return f.pairLaunchFake.Create(ctx, workerPairIdentity{Job: f.supervisor.record.JobID, Role: "tool", Repository: f.supervisor.record.Pair.Repository})
}
func (f assembledToolFake) VerifyCreated(ctx context.Context, id string) error {
	return f.Verify(ctx, workerPairIdentity{Job: f.supervisor.record.JobID, Role: "tool", Repository: f.supervisor.record.Pair.Repository}, id)
}
func (f assembledToolFake) Bootstrap(ctx context.Context, in io.ReadCloser) (workspacebootstrap.Receipt, error) {
	defer in.Close()
	return f.pairLaunchFake.Bootstrap(ctx, strings.Repeat("c", 64))
}

type pairBundleFixture struct {
	io.Reader
	closes atomic.Int32
}

func (b *pairBundleFixture) Close() error { b.closes.Add(1); return nil }

func assembledPairFixture(t *testing.T) (*workerPairSupervisor, *workerAssembledPair, *pairLaunchFake, *pairBundleFixture) {
	t.Helper()
	_, s, _ := workspacePair(t)
	fake := &pairLaunchFake{t: t, supervisor: s}
	scope := toolbroker.Scope{Version: 1, JobID: s.record.JobID, Room: s.record.RoomID, Participant: s.record.ParticipantID, Connection: s.record.ConnectionID, Provider: s.record.Provider, RepositoryDigest: s.record.Pair.Repository, Deadline: s.record.Deadline}
	input := &pairBundleFixture{Reader: strings.NewReader("synthetic-bundle")}
	a := &workerAssembledPair{ctx: s.ctx, scope: scope, engine: fake, tool: assembledToolFake{fake}, bundle: &workerPairBundle{ReadCloser: input}}
	return s, a, fake, input
}
func TestAssembledPairRoutesSupervisorStartupToSharedToolRuntime(t *testing.T) {
	t.Parallel()
	s, a, fake, input := assembledPairFixture(t)
	if err := s.Start(t.Context(), a); err != nil {
		t.Fatal(err)
	}
	if s.record.Status != "running" || !a.bootstrapped || input.closes.Load() != 1 {
		t.Fatal("bootstrap not consumed exactly once")
	}
	if len(fake.calls) < 3 || fake.calls[len(fake.calls)-1] != "start:d" {
		t.Fatal("engine not started last", fake.calls)
	}
	for _, id := range []string{a.toolID, a.engineID} {
		if a.Start(t.Context(), id) == nil {
			t.Fatal("start replayed", id)
		}
	}
	if _, err := a.Bootstrap(t.Context(), a.toolID); err == nil {
		t.Fatal("bootstrap replayed")
	}
}
func TestAssembledPairCannotStartEngineAfterFailedBootstrap(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"bootstrap", "bootstrap-binding"} {
		t.Run(mode, func(t *testing.T) {
			s, a, fake, _ := assembledPairFixture(t)
			fake.fail = mode
			if s.Start(t.Context(), a) == nil {
				t.Fatal("failed bootstrap accepted")
			}
			if a.Start(t.Context(), a.engineID) == nil {
				t.Fatal("engine started after failed bootstrap")
			}
			for _, call := range fake.calls {
				if call == "start:d" {
					t.Fatal("engine launch reached")
				}
			}
		})
	}
}
func TestAssembledPairRejectsForeignScopeAndUnknownIdentities(t *testing.T) {
	t.Parallel()
	_, a, fake, _ := assembledPairFixture(t)
	for _, identity := range []workerPairIdentity{{Job: "foreign", Role: "tool", Repository: a.scope.RepositoryDigest}, {Job: a.scope.JobID, Role: "engine", Repository: strings.Repeat("f", 64)}, {Job: a.scope.JobID, Role: "other", Repository: a.scope.RepositoryDigest}} {
		if _, err := a.Create(t.Context(), identity); err == nil {
			t.Fatal("foreign scope created")
		}
	}
	unknown := strings.Repeat("f", 64)
	if _, err := a.Inspect(t.Context(), unknown); err == nil {
		t.Fatal("unknown container inspected")
	}
	if a.Start(t.Context(), unknown) == nil {
		t.Fatal("unknown container started")
	}
	if _, err := a.Bootstrap(t.Context(), unknown); err == nil {
		t.Fatal("unknown container bootstrapped")
	}
	if len(fake.calls) != 0 {
		t.Fatal("invalid request reached runtime", fake.calls)
	}
}
func TestAssembledPairOriginalCancellationBlocksFreshContext(t *testing.T) {
	t.Parallel()
	s, a, fake, _ := assembledPairFixture(t)
	s.Close()
	if _, err := a.WorkspaceIntent(context.Background()); err == nil {
		t.Fatal("cancelled assembly revived")
	}
	if a.Start(context.Background(), strings.Repeat("d", 64)) == nil {
		t.Fatal("cancelled engine started")
	}
	if len(fake.calls) != 0 {
		t.Fatal("cancelled request reached adapter")
	}
}
func TestAssemblePairBindsRealAuditedSessionAndClosesUnusedBundle(t *testing.T) {
	t.Parallel()
	_, s, _ := workspacePair(t)
	r := s.record
	scope := toolbroker.Scope{Version: 1, JobID: r.JobID, Room: r.RoomID, Participant: r.ParticipantID, Connection: r.ConnectionID, Provider: r.Provider, RepositoryDigest: r.Pair.Repository, Deadline: r.Deadline}
	client, peer := net.Pipe()
	defer peer.Close()
	ctx := context.WithValue(s.ctx, workerLocalJobContextKey{}, r.JobID)
	session, _, err := newWorkerToolSession(ctx, scope, []string{"read"}, "ghcr.io/example/tool@sha256:"+strings.Repeat("a", 64), "/trusted/docker", time.Now(), client, &pairFakeDocker{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	input := &pairBundleFixture{Reader: strings.NewReader("synthetic")}
	assembly, err := s.AssemblePair(session, &pairLaunchFake{t: t, supervisor: s}, input)
	if err != nil {
		t.Fatal(err)
	}
	if assembly.tool != session.Runtime || session.Scope() != scope {
		t.Fatal("different tool runtime or scope bound")
	}
	otherInput := &pairBundleFixture{Reader: strings.NewReader("second-bundle")}
	if _, err = s.AssemblePair(session, &pairLaunchFake{t: t, supervisor: s}, otherInput); err == nil {
		t.Fatal("supervisor assembled twice")
	}
	if otherInput.closes.Load() != 0 {
		t.Fatal("refused assembly took caller input ownership")
	}
	s.Close()
	deadline := time.After(time.Second)
	for input.closes.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("unused bundle not closed")
		case <-time.After(time.Millisecond):
		}
	}
	if input.closes.Load() != 1 {
		t.Fatal("bundle closed repeatedly")
	}
	inputScope := session.Scope()
	inputScope.JobID = "foreign"
	if session.Scope().JobID != scope.JobID {
		t.Fatal("scope mutated through copy")
	}
	if _, err = s.AssemblePair(session, &pairLaunchFake{t: t, supervisor: s}, input); err == nil {
		t.Fatal("closed supervisor assembled")
	}
}

func TestAssembledPairUncertainCreateAndStartAreNeverRetried(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"create:tool", "create:engine", "start:c", "start:d"} {
		t.Run(mode, func(t *testing.T) {
			s, a, fake, _ := assembledPairFixture(t)
			fake.fail = mode
			if s.Start(t.Context(), a) == nil {
				t.Fatal("uncertain startup accepted")
			}
			calls := len(fake.calls)
			role := "tool"
			if mode == "create:engine" {
				role = "engine"
			}
			if strings.HasPrefix(mode, "create:") {
				if _, err := a.Create(t.Context(), workerPairIdentity{Job: a.scope.JobID, Role: role, Repository: a.scope.RepositoryDigest}); err == nil {
					t.Fatal("uncertain creation retried")
				}
			} else {
				id := a.toolID
				if mode == "start:d" {
					id = a.engineID
				}
				if a.Start(t.Context(), id) == nil {
					t.Fatal("uncertain start retried")
				}
			}
			if len(fake.calls) != calls {
				t.Fatal("retry reached backend")
			}
		})
	}
}

func TestAssemblePairRejectsDifferentAuditedSessionScope(t *testing.T) {
	t.Parallel()
	_, s, _ := workspacePair(t)
	r := s.record
	scope := toolbroker.Scope{Version: 1, JobID: r.JobID, Room: r.RoomID, Participant: r.ParticipantID, Connection: "foreign-connection", Provider: r.Provider, RepositoryDigest: r.Pair.Repository, Deadline: r.Deadline}
	client, peer := net.Pipe()
	defer peer.Close()
	ctx := context.WithValue(s.ctx, workerLocalJobContextKey{}, r.JobID)
	session, _, err := newWorkerToolSession(ctx, scope, []string{"read"}, "ghcr.io/example/tool@sha256:"+strings.Repeat("a", 64), "/trusted/docker", time.Now(), client, &pairFakeDocker{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	input := &pairBundleFixture{Reader: strings.NewReader("synthetic")}
	if _, err = s.AssemblePair(session, &pairLaunchFake{t: t, supervisor: s}, input); err == nil {
		t.Fatal("different session scope bound")
	}
	if input.closes.Load() != 0 || s.assembled {
		t.Fatal("refused session consumed assembly ownership")
	}
}
