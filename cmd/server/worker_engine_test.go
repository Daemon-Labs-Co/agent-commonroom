package main

import (
	"agentroom/internal/toolbroker"
	"context"
	"net"
	"strings"
	"testing"
	"time"
)

func TestWorkerCodexEngineRequiresOriginalWorkerJobAndRole(t *testing.T) {
	t.Parallel()
	scope := sampleToolScope(t)
	image := "ghcr.io/example/engine@sha256:" + strings.Repeat("a", 64)
	for _, mode := range []string{"missing job", "wrong job", "valid"} {
		t.Run(mode, func(t *testing.T) {
			ctx := t.Context()
			if mode != "missing job" {
				job := scope.JobID
				if mode == "wrong job" {
					job = strings.Repeat("f", 32)
				}
				ctx = context.WithValue(ctx, workerLocalJobContextKey{}, job)
			}
			var gateway *workerEngineGateway
			if mode == "valid" {
				gateway = engineGatewayFixture(t, ctx, scope)
			}
			engine, err := newWorkerCodexEngine(ctx, scope, image, gateway, &workerEngineCredential{scope: scope, original: ctx, raw: []byte(syntheticEngineCredential)}, &pairFakeDocker{})
			if mode != "valid" {
				if err == nil {
					t.Fatal("unbound engine constructed")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			defer engine.runtime.Close()
			for _, identity := range []workerPairIdentity{{Job: scope.JobID, Role: "tool", Repository: scope.RepositoryDigest}, {Job: "foreign", Role: "engine", Repository: scope.RepositoryDigest}, {Job: scope.JobID, Role: "engine", Repository: strings.Repeat("f", 64)}} {
				if _, err = engine.Create(t.Context(), identity); err == nil {
					t.Fatal("foreign engine role created")
				}
				if engine.Verify(t.Context(), identity, strings.Repeat("d", 64)) == nil {
					t.Fatal("foreign engine verified")
				}
			}
		})
	}
}

func TestPairAssemblyRejectsEngineFromDifferentCompleteScope(t *testing.T) {
	t.Parallel()
	_, s, _ := workspacePair(t)
	r := s.record
	scope := toolbroker.Scope{Version: 1, JobID: r.JobID, Room: r.RoomID, Participant: r.ParticipantID, Connection: r.ConnectionID, Provider: r.Provider, RepositoryDigest: r.Pair.Repository, Deadline: r.Deadline}
	client, peer := net.Pipe()
	defer peer.Close()
	ctx := context.WithValue(s.ctx, workerLocalJobContextKey{}, scope.JobID)
	session, _, err := newWorkerToolSession(ctx, scope, []string{"read"}, "ghcr.io/example/tool@sha256:"+strings.Repeat("a", 64), "/trusted/docker", time.Now(), client, &pairFakeDocker{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	other := scope
	other.Room = "foreign-room"
	engine, err := newWorkerCodexEngine(ctx, other, "ghcr.io/example/engine@sha256:"+strings.Repeat("a", 64), engineGatewayFixture(t, ctx, other), &workerEngineCredential{scope: other, original: ctx, raw: []byte(syntheticEngineCredential)}, &pairFakeDocker{})
	if err != nil {
		t.Fatal(err)
	}
	defer engine.runtime.Close()
	if _, err = engine.Turn(ctx, s, session, strings.Repeat("a", 64), "synthetic", "model"); err == nil {
		t.Fatal("engine turn accepted a different session scope")
	}
	if s.ctx.Err() != nil || s.codexTurnStarted {
		t.Fatal("refused scope consumed or cancelled the original supervisor")
	}
	input := &pairBundleFixture{Reader: strings.NewReader("synthetic-bundle")}
	if _, err = s.AssemblePair(session, engine, input); err == nil {
		t.Fatal("same job/repository with foreign room accepted")
	}
	if s.assembled || input.closes.Load() != 0 {
		t.Fatal("refused engine consumed assembly ownership")
	}
}
