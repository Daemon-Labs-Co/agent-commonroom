//go:build linux || darwin

package main

import (
	"agentroom/internal/objectcustody"
	"agentroom/internal/room"
	"agentroom/internal/taskcontract"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRecordedArtifactMemberRequiresFreshBytesAndExactManifest(t *testing.T) {
	for _, mode := range []string{"healthy", "editing off", "wrong digest", "unknown member", "foreign execution", "agent", "tampered", "revoked after read", "release revoked", "canceled", "canceled after read"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			s, db, ctx, scope := artifactJournalFixture(t, true)
			guard := &syntheticArtifactGuard{}
			journal, _ := newControllerArtifactJournal(s, guard)
			wrapped := &manifestAckJournal{Journal: journal}
			store, dir := artifactStore(t, wrapped)
			receipt, err := store.Put(ctx, scope, 128, strings.NewReader("synthetic immutable member"))
			if err != nil {
				t.Fatal(err)
			}
			manifest, err := s.recordStoredArtifactManifest(ctx, scope.ExecutionScope, "report", store)
			if err != nil {
				t.Fatal(err)
			}
			actor := room.Actor{ID: "reader", Kind: "human", Enabled: true, Revision: 1, RoomIDs: []string{scope.RoomID}}
			db.seed("actor.revised", actor)
			readCtx := context.WithValue(ctx, actorContextKey{}, actor.ID)
			execution, digest, member := scope.ExecutionID, manifest.Digest, "report.txt"
			switch mode {
			case "editing off":
				db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": false})
			case "wrong digest":
				digest = strings.Repeat("0", 64)
			case "unknown member":
				member = "../private-key"
			case "foreign execution":
				execution = strings.Repeat("0", 32)
			case "agent":
				readCtx = context.WithValue(ctx, actorContextKey{}, "bot-openai")
			case "tampered":
				if err := os.WriteFile(filepath.Join(dir, "object-"+receipt.Intent.ObjectID+".blob"), []byte("changed bytes"), 0600); err != nil {
					t.Fatal(err)
				}
			case "revoked after read":
				calls := 0
				wrapped.after = func() {
					calls++
					if calls == 2 {
						actor.Enabled = false
						actor.Revision++
						db.seed("actor.revised", actor)
					}
				}
			case "release revoked":
				calls := 0
				guard.release = func(context.Context, room.State, objectcustody.Receipt) error {
					calls++
					if calls == 3 {
						return errors.New("synthetic release revocation")
					}
					return nil
				}
			case "canceled after read":
				var cancel context.CancelFunc
				readCtx, cancel = context.WithCancel(readCtx)
				defer cancel()
				calls := 0
				wrapped.after = func() {
					calls++
					if calls == 2 {
						cancel()
					}
				}
			case "canceled":
				var cancel context.CancelFunc
				readCtx, cancel = context.WithCancel(readCtx)
				cancel()
			}
			raw, err := s.readRecordedArtifactMember(readCtx, execution, "report", digest, member, store)
			if mode == "healthy" || mode == "editing off" {
				if err != nil || string(raw) != "synthetic immutable member" {
					t.Fatal("fresh member denied", err)
				}
			} else if err == nil || raw != nil {
				t.Fatal("unverified/stale member released", mode, raw, err)
			}
			st := artifactState(t, s, ctx)
			acks := st.ArtifactCustody[receipt.Intent.ObjectID].ReadAcknowledgments
			if mode == "healthy" || mode == "editing off" || (mode == "revoked after read" || mode == "canceled after read") {
				if acks != 3 {
					t.Fatal("missing actual read history", acks)
				}
			} else if mode == "release revoked" {
				if acks != 2 {
					t.Fatal("release revocation lost prior verification history", acks)
				}
			} else if acks != 1 {
				t.Fatal("refused preflight caused extra reads", mode, acks)
			}
			if artifactTotal(t, s, ctx) != 128 || len(st.ArtifactManifests) != 1 {
				t.Fatal("read refunded or replaced custody")
			}
		})
	}
}
func TestRecordedArtifactBundleMemberRefusesChangedSibling(t *testing.T) {
	t.Parallel()
	output := taskcontract.DataContract{Name: "report", Type: "bundle", MediaType: "application/vnd.commonroom.bundle+json", MaxBytes: 128, Members: []taskcontract.BundleMember{{Name: "report.txt", MediaType: "text/plain", MaxBytes: 64}, {Name: "checks.json", MediaType: "application/json", MaxBytes: 64}}}
	s, _, ctx, scope := artifactJournalOutputFixture(t, true, output)
	journal, _ := newControllerArtifactJournal(s, &syntheticArtifactGuard{})
	store, dir := artifactStore(t, journal)
	if _, err := store.Put(ctx, scope, 64, strings.NewReader("synthetic report")); err != nil {
		t.Fatal(err)
	}
	sibling := scope
	sibling.MemberName = "checks.json"
	receipt, err := store.Put(ctx, sibling, 64, strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := s.recordStoredArtifactManifest(ctx, scope.ExecutionScope, "report", store)
	if err != nil {
		t.Fatal(err)
	}
	if raw, err := s.readRecordedArtifactMember(ctx, scope.ExecutionID, "report", manifest.Digest, "report.txt", store); err != nil || string(raw) != "synthetic report" {
		t.Fatal("healthy bundle member denied", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "object-"+receipt.Intent.ObjectID+".blob"), []byte("tampered"), 0600); err != nil {
		t.Fatal(err)
	}
	if raw, err := s.readRecordedArtifactMember(ctx, scope.ExecutionID, "report", manifest.Digest, "report.txt", store); err == nil || raw != nil {
		t.Fatal("valid member released from invalid bundle")
	}
}

var _ objectcustody.Journal = (*manifestAckJournal)(nil)
