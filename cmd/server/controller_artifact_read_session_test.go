//go:build linux || darwin

package main

import (
	"agentroom/internal/objectcustody"
	"agentroom/internal/room"
	"context"
	"errors"
	"net"
	"os"
	"strings"
	"testing"
)

func artifactReadSessionFixture(t *testing.T) (*controllerArtifactReadSession, *mockDB, context.Context, *objectcustody.Store, *syntheticArtifactGuard) {
	t.Helper()
	s, db, ctx, scope := artifactJournalFixture(t, true)
	e := &syntheticArtifactGuard{}
	journal, err := newControllerArtifactJournal(s, e)
	if err != nil {
		t.Fatal(err)
	}
	store, _ := artifactStore(t, journal)
	receipt, err := store.Put(ctx, scope, 128, strings.NewReader("synthetic sealed bytes"))
	if err != nil {
		t.Fatal(err)
	}
	g, err := newControllerArtifactReadSession(ctx, s, "synthetic-worker", receipt, e)
	if err != nil {
		t.Fatal(err)
	}
	return g, db, ctx, store, e
}
func TestControllerArtifactReadSessionPrivateByteTransfer(t *testing.T) {
	t.Parallel()
	for _, off := range []bool{false, true} {
		t.Run(map[bool]string{false: "enabled", true: "editing-off"}[off], func(t *testing.T) {
			t.Parallel()
			g, db, ctx, store, _ := artifactReadSessionFixture(t)
			if off {
				db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": false})
				defer db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": true})
			}
			left, right := net.Pipe()
			done := make(chan error, 1)
			go func() { done <- objectcustody.ServeReadChannel(ctx, right, g.receipt.Intent.Scope, g.receipt, store) }()
			// The private session pins the original authenticated owner, not this caller value.
			raw, err := objectcustody.ReadChannel(context.WithValue(ctx, actorContextKey{}, "untrusted"), left, g.receipt.Intent.Scope, g.receipt, g)
			if err != nil || string(raw) != "synthetic sealed bytes" {
				t.Fatal(err)
			}
			if err = <-done; err != nil {
				t.Fatal(err)
			}
			st := artifactState(t, g.server, ctx)
			saved := st.ArtifactCustody[g.receipt.Intent.ObjectID]
			if saved.ReadAcknowledgments != 2 {
				t.Fatal("missing source/receiver acknowledgments", saved.ReadAcknowledgments)
			}
			events := db.snapshot()
			var found int
			for _, v := range events {
				if v.kind == "artifact.custody.read" {
					found++
				}
			}
			if found != 2 {
				t.Fatal("wrong release audit", found)
			}
		})
	}
}
func TestControllerArtifactReadSessionCurrentPolicy(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"worker", "receipt", "scope", "canceled", "agent", "missing", "denied", "stale", "revoked", "incident", "integrity"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			g, db, ctx, _, e := artifactReadSessionFixture(t)
			scope := g.receipt.Intent.Scope
			r := g.receipt
			switch mode {
			case "worker":
				g.worker = "foreign"
			case "receipt":
				r.Digest = strings.Repeat("9", 64)
			case "scope":
				scope.MemberName = "other.txt"
			case "canceled":
				c, cancel := context.WithCancel(ctx)
				cancel()
				g.lifetime = c
			case "agent":
				g.actor = "bot-openai"
			case "missing":
				g.actor = "missing-human"
			case "denied":
				e.release = func(context.Context, room.State, objectcustody.Receipt) error {
					return errors.New("synthetic release denial")
				}
			case "stale":
				e.release = func(context.Context, room.State, objectcustody.Receipt) error {
					db.seed("agent.background", map[string]string{"stage": "synthetic"})
					return nil
				}
			case "revoked":
				e.release = func(context.Context, room.State, objectcustody.Receipt) error {
					db.seed("incident.opened", map[string]any{"reason": "synthetic stop"})
					return nil
				}
			case "incident":
				db.seed("incident.opened", map[string]any{"reason": "synthetic stop"})
			case "integrity":
				g.server.integrityFault = true
				defer func() { g.server.integrityFault = false }()
			}
			before := len(db.snapshot())
			err := g.AuthorizeRead(ctx, scope, r)
			if err == nil {
				t.Fatal("denied read authorized", mode)
			}
			extra := 0
			if mode == "stale" || mode == "revoked" {
				extra = 1
			}
			if len(db.snapshot()) != before+extra {
				t.Fatal("metadata precheck appended release")
			}
			// The journal translates physical policy refusals to a generic denial. No stale
			// callback observation grants metadata access or an acknowledgment.
			if mode != "scope" && g.AcknowledgeRead(ctx, r) == nil {
				t.Fatal("denied acknowledgment granted")
			}
		})
	}
}
func TestControllerArtifactReadSessionAuditFailureAndStop(t *testing.T) {
	t.Parallel()
	g, db, ctx, _, _ := artifactReadSessionFixture(t)
	if g.AuthorizeRead(ctx, g.receipt.Intent.Scope, g.receipt) != nil {
		t.Fatal("setup")
	}
	db.failAt = "commit"
	if g.AcknowledgeRead(ctx, g.receipt) == nil || !g.server.integrityFault {
		t.Fatal("failed release persistence granted")
	}
	db.failAt = ""
	g.server.integrityFault = false
	g.StopReadAuthority()
	if g.AuthorizeRead(ctx, g.receipt.Intent.Scope, g.receipt) == nil || g.AcknowledgeRead(ctx, g.receipt) == nil {
		t.Fatal("stopped session revived")
	}
	if g.ValidateSource(ctx, room.State{}, g.receipt.Intent) == nil || g.EstablishQuota(ctx, room.StorageReservation{}) == nil || g.ValidateSeal(ctx, room.State{}, g.receipt) == nil {
		t.Fatal("read session gained write authority")
	}
	if _, err := newControllerArtifactReadSession(t.Context(), g.server, g.worker, g.receipt, &syntheticArtifactGuard{}); err == nil {
		t.Fatal("missing authenticated principal")
	}
	if _, err := newControllerArtifactReadSession(ctx, g.server, g.worker, g.receipt, nil); err == nil {
		t.Fatal("missing release evidence")
	}
}

func TestControllerArtifactReadSessionPrivateJournalRelay(t *testing.T) {
	t.Parallel()
	s, _, ctx, scope := artifactJournalFixture(t, true)
	evidence := &syntheticArtifactGuard{}
	journal, err := newControllerArtifactJournal(s, evidence)
	if err != nil {
		t.Fatal(err)
	}
	store, dir := artifactStore(t, journal)
	receipt, err := store.Put(ctx, scope, 128, strings.NewReader("synthetic relayed artifact"))
	if err != nil {
		t.Fatal(err)
	}
	session, err := newControllerArtifactReadSession(ctx, s, "synthetic-worker", receipt, evidence)
	if err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	left, right := net.Pipe()
	readerCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	relay, err := objectcustody.NewReadJournalChannel(readerCtx, right, receipt, cancel)
	if err != nil {
		t.Fatal(err)
	}
	defer relay.Close()
	readStore, err := objectcustody.New(root, relay)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- objectcustody.ServeReadChannel(readerCtx, right, scope, receipt, readStore) }()
	raw, err := objectcustody.ReadChannel(ctx, left, scope, receipt, session)
	if err != nil || string(raw) != "synthetic relayed artifact" {
		t.Fatal(err)
	}
	if err = <-done; err != nil {
		t.Fatal(err)
	}
	st := artifactState(t, s, ctx)
	if st.ArtifactCustody[receipt.Intent.ObjectID].ReadAcknowledgments != 2 {
		t.Fatal("source or receiving release audit missing")
	}
}
