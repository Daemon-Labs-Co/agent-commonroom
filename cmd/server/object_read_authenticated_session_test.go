package main

import (
	"agentroom/internal/room"
	"context"
	"net"
	"strings"
	"testing"
	"time"
)

func boundObjectReadFixture(t *testing.T, g *controllerObjectReadSession) (objectReadPeerBinding, objectReadPeerBinding) {
	t.Helper()
	client, server := objectReadTestPeers(t)
	b := objectReadPeerBinding{Identity: client, Worker: g.worker, RecordDigest: probeWorkerRecordDigest(g.record), Intent: g.intent}
	s := b
	s.Identity = server
	return b, s
}
func TestAuthenticatedObjectReadSessionLifecycle(t *testing.T) {
	t.Parallel()
	for _, denied := range []bool{false, true} {
		t.Run(map[bool]string{false: "settled", true: "permission denied"}[denied], func(t *testing.T) {
			t.Parallel()
			g, db, ctx, e := objectReadSessionFixture(t)
			if denied {
				e.operation = func(context.Context, room.State, workerJobRecord, workerObjectReadIntent, workerProbeEvent) error {
					return errObjectReadPeer
				}
			}
			before := len(db.snapshot())
			client, server := boundObjectReadFixture(t, g)
			a, b := net.Pipe()
			defer a.Close()
			defer b.Close()
			done := make(chan error, 1)
			go func() { done <- serveAuthenticatedObjectReadSession(ctx, b, server, g) }()
			c, err := newAuthenticatedObjectReadClient(ctx, a, client, g.record, g.intent, func() {})
			if err != nil {
				t.Fatal(err)
			}
			defer c.StopObjectReadAuthority()
			local := workerObjectReadRecord{Version: 1, Intent: g.intent}
			for _, event := range []workerProbeEvent{{Stage: "begin"}, {Stage: "bound", ID: strings.Repeat("1", 64)}, {Stage: "finish", ID: strings.Repeat("1", 64), Absent: true}} {
				err = c.AuthorizeObjectRead(ctx, g.record, g.intent, event)
				if denied {
					if err == nil {
						t.Fatal("TLS identity bypassed permission denial")
					}
					break
				}
				if err != nil {
					t.Fatal(err)
				}
				local.Events = append(local.Events, event)
				if err = c.AcknowledgeObjectRead(ctx, local, event); err != nil {
					t.Fatal(err)
				}
			}
			c.StopObjectReadAuthority()
			<-done
			if g.lifetime.Err() == nil {
				t.Fatal("controller authority survived closed channel")
			}
			want := before + 3
			if denied {
				want = before
			}
			if len(db.snapshot()) != want {
				t.Fatal("wrong durable audit count", len(db.snapshot()), want)
			}
			if !denied {
				st := artifactState(t, g.server, ctx)
				saved := st.ObjectReads[controllerObjectReadKey(g.intent)]
				if saved.ActorID != g.actor || !saved.Released {
					t.Fatal("lost pinned actor or settlement")
				}
			}
		})
	}
}
func TestAuthenticatedObjectReadSessionBindingRefusals(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"worker", "record", "intent", "certificate", "controller worker", "controller record"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()
			g, db, ctx, e := objectReadSessionFixture(t)
			before := len(db.snapshot())
			client, server := boundObjectReadFixture(t, g)
			switch kind {
			case "worker":
				client.Worker = "other-worker"
			case "record":
				client.RecordDigest = strings.Repeat("0", 64)
			case "intent":
				client.Intent.Request.Challenge = strings.Repeat("8", 32)
				if client.Intent == g.intent {
					client.Intent.Request.Challenge = strings.Repeat("9", 32)
				}
			case "certificate":
				client.Identity.PeerSHA256[0] ^= 1
			case "controller worker":
				server.Worker = "other-worker"
			case "controller record":
				server.RecordDigest = strings.Repeat("0", 64)
			}
			a, b := net.Pipe()
			defer a.Close()
			defer b.Close()
			child, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()
			done := make(chan error, 1)
			go func() { done <- serveAuthenticatedObjectReadSession(child, b, server, g) }()
			c, err := newAuthenticatedObjectReadClient(child, a, client, g.record, g.intent, func() {})
			if c != nil {
				c.StopObjectReadAuthority()
			}
			serverErr := <-done
			if err == nil || c != nil || serverErr == nil {
				t.Fatal("substituted binding accepted", err, serverErr)
			}
			if e.calls != 0 || e.acks != 0 || len(db.snapshot()) != before {
				t.Fatal("invalid binding reached authority or persistence")
			}
		})
	}
}
func TestAuthenticatedObjectReadSessionIdleCancellation(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"worker parent", "controller lifetime", "connection loss"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()
			g, db, ctx, _ := objectReadSessionFixture(t)
			before := len(db.snapshot())
			client, server := boundObjectReadFixture(t, g)
			a, b := net.Pipe()
			defer a.Close()
			defer b.Close()
			child, cancel := context.WithCancel(ctx)
			defer cancel()
			done := make(chan error, 1)
			go func() { done <- serveAuthenticatedObjectReadSession(ctx, b, server, g) }()
			c, err := newAuthenticatedObjectReadClient(child, a, client, g.record, g.intent, func() {})
			if err != nil {
				t.Fatal(err)
			}
			defer c.StopObjectReadAuthority()
			switch kind {
			case "worker parent":
				cancel()
			case "controller lifetime":
				g.StopObjectReadAuthority()
			case "connection loss":
				_ = a.Close()
			}
			select {
			case <-c.ObjectReadAuthorityLifetime().Done():
			case <-time.After(time.Second):
				t.Fatal("idle worker authority survived cancellation")
			}
			c.StopObjectReadAuthority()
			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("controller did not join")
			}
			if g.lifetime.Err() == nil || len(db.snapshot()) != before {
				t.Fatal("idle loss changed audit or retained authority")
			}
		})
	}
}
