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

func TestStoredArtifactManifestUsesControllerReceiptsAndActualBytes(t *testing.T) {
	for _, off := range []bool{false, true} {
		t.Run(map[bool]string{false: "normal", true: "editing off"}[off], func(t *testing.T) {
			t.Parallel()
			s, db, ctx, scope := artifactJournalFixture(t, true)
			guard := &syntheticArtifactGuard{}
			journal, _ := newControllerArtifactJournal(s, guard)
			store, _ := artifactStore(t, journal)
			receipt, err := store.Put(ctx, scope, 128, strings.NewReader("actual synthetic manifest bytes"))
			if err != nil {
				t.Fatal(err)
			}
			if off {
				db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": false})
			}
			manifest, err := s.resolveStoredArtifactManifest(ctx, scope.ExecutionScope, "report", store)
			if err != nil || manifest.OutputType != "file" || len(manifest.Files) != 1 || manifest.Files[0].Receipt != receipt || manifest.TotalBytes != receipt.Bytes || manifest.Digest == "" {
				t.Fatal(manifest, err)
			}
			st := artifactState(t, s, ctx)
			if st.ArtifactCustody[receipt.Intent.ObjectID].ReadAcknowledgments != 1 {
				t.Fatal("actual byte release not acknowledged")
			}
			// Manifest output is metadata only, never task acceptance or provider success.
			if st.Tasks[0].Status == "Accepted" {
				t.Fatal("manifest accepted task")
			}
		})
	}
}
func TestStoredArtifactManifestRefusesUntrustedOrIncompleteState(t *testing.T) {
	for _, mode := range []string{"foreign company", "foreign room", "wrong revision", "unknown output", "agent", "unsealed", "release denied", "tampered bytes", "incident", "integrity", "canceled"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			s, db, ctx, scope := artifactJournalFixture(t, true)
			guard := &syntheticArtifactGuard{}
			journal, _ := newControllerArtifactJournal(s, guard)
			store, dir := artifactStore(t, journal)
			var receipt objectcustody.Receipt
			if mode == "unsealed" {
				if err := journal.Reserve(ctx, objectcustody.Intent{ObjectID: strings.Repeat("6", 32), Scope: scope, MaxBytes: 128}); err != nil {
					t.Fatal(err)
				}
			} else {
				var err error
				receipt, err = store.Put(ctx, scope, 128, strings.NewReader("synthetic manifest bytes"))
				if err != nil {
					t.Fatal(err)
				}
			}
			switch mode {
			case "foreign company":
				scope.CompanyID = "foreign"
			case "foreign room":
				scope.RoomID = "foreign"
			case "wrong revision":
				scope.RevisionDigest = strings.Repeat("7", 64)
			case "agent":
				ctx = context.WithValue(ctx, actorContextKey{}, "bot-openai")
			case "release denied":
				guard.release = func(context.Context, room.State, objectcustody.Receipt) error { return errors.New("synthetic refusal") }
			case "tampered bytes":
				if err := os.WriteFile(filepath.Join(dir, "object-"+receipt.Intent.ObjectID+".blob"), []byte("changed"), 0600); err != nil {
					t.Fatal(err)
				}
			case "incident":
				db.seed("incident.opened", map[string]any{"reason": "synthetic stop"})
			case "integrity":
				s.integrityFault = true
				defer func() { s.integrityFault = false }()
			case "canceled":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			output := "report"
			if mode == "unknown output" {
				output = "unknown"
			}
			manifest, err := s.resolveStoredArtifactManifest(ctx, scope.ExecutionScope, output, store)
			if err == nil || manifest.Digest != "" || manifest.Files != nil {
				t.Fatal("uncertain output escaped", manifest, err)
			}
			if receipt.Intent.ObjectID != "" {
				st := artifactState(t, s, t.Context())
				if st.ArtifactCustody[receipt.Intent.ObjectID].ReadAcknowledgments != 0 {
					t.Fatal("refused assembly released bytes")
				}
			}
		})
	}
}

// Hook runs after the real journal has released its server lock. This models a
// concurrent policy change after byte authorization, never reentrant callbacks.
type manifestAckJournal struct {
	objectcustody.Journal
	after func()
}

func (j *manifestAckJournal) AcknowledgeRead(ctx context.Context, r objectcustody.Receipt) error {
	if err := j.Journal.AcknowledgeRead(ctx, r); err != nil {
		return err
	}
	if j.after != nil {
		j.after()
	}
	return nil
}
func TestStoredArtifactManifestRechecksAfterAllReads(t *testing.T) {
	for _, mode := range []string{"access revoked", "incident", "integrity"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			s, db, ctx, scope := artifactJournalFixture(t, true)
			guard := &syntheticArtifactGuard{}
			journal, _ := newControllerArtifactJournal(s, guard)
			wrapped := &manifestAckJournal{Journal: journal}
			store, _ := artifactStore(t, wrapped)
			receipt, err := store.Put(ctx, scope, 128, strings.NewReader("synthetic sealed output"))
			if err != nil {
				t.Fatal(err)
			}
			actor := room.Actor{ID: "reader", Kind: "human", Enabled: true, Revision: 1, RoomIDs: []string{scope.RoomID}}
			db.seed("actor.revised", actor)
			wrapped.after = func() {
				switch mode {
				case "access revoked":
					actor.Enabled = false
					actor.Revision++
					db.seed("actor.revised", actor)
				case "incident":
					db.seed("incident.opened", map[string]any{"reason": "synthetic stop"})
				case "integrity":
					s.mu.Lock()
					s.integrityFault = true
					s.mu.Unlock()
				}
			}
			readCtx := context.WithValue(ctx, actorContextKey{}, actor.ID)
			manifest, err := s.resolveStoredArtifactManifest(readCtx, scope.ExecutionScope, "report", store)
			if err == nil || manifest.Digest != "" || manifest.Files != nil {
				t.Fatal("revoked complete manifest returned", manifest, err)
			}
			if mode == "integrity" {
				s.mu.Lock()
				s.integrityFault = false
				s.mu.Unlock()
			}
			st := artifactState(t, s, t.Context())
			if st.ArtifactCustody[receipt.Intent.ObjectID].ReadAcknowledgments != 1 {
				t.Fatal("authorized read history lost")
			}
		})
	}
}
func TestStoredArtifactBundleRequiresCompleteSealedMembers(t *testing.T) {
	for _, mode := range []string{"healthy", "missing", "invalid content"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			output := taskcontract.DataContract{Name: "report", Type: "bundle", MediaType: "application/vnd.commonroom.bundle+json", MaxBytes: 128, Members: []taskcontract.BundleMember{{Name: "report.txt", MediaType: "text/plain", MaxBytes: 64}, {Name: "checks.json", MediaType: "application/json", MaxBytes: 64}}}
			s, _, ctx, scope := artifactJournalOutputFixture(t, true, output)
			journal, _ := newControllerArtifactJournal(s, &syntheticArtifactGuard{})
			store, _ := artifactStore(t, journal)
			first, err := store.Put(ctx, scope, 64, strings.NewReader("synthetic report\n"))
			if err != nil {
				t.Fatal(err)
			}
			if mode != "missing" {
				secondScope := scope
				secondScope.MemberName = "checks.json"
				content := "{}"
				if mode == "invalid content" {
					content = "not json"
				}
				if _, err := store.Put(ctx, secondScope, 64, strings.NewReader(content)); err != nil {
					t.Fatal(err)
				}
			}
			manifest, err := s.resolveStoredArtifactManifest(ctx, scope.ExecutionScope, "report", store)
			st := artifactState(t, s, ctx)
			if mode == "healthy" {
				if err != nil || manifest.OutputType != "bundle" || len(manifest.Files) != 2 || manifest.Files[0].Name != "checks.json" || manifest.Files[1].Name != "report.txt" {
					t.Fatal(manifest, err)
				}
			} else {
				if err == nil || manifest.Digest != "" || manifest.Files != nil {
					t.Fatal("partial/invalid bundle returned")
				}
			}
			if mode == "missing" && st.ArtifactCustody[first.Intent.ObjectID].ReadAcknowledgments != 0 {
				t.Fatal("incomplete set read before preflight refusal")
			}
		})
	}
}
