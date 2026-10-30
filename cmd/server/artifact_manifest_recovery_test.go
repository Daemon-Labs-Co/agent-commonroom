//go:build linux || darwin

package main

import (
	"agentroom/internal/objectcustody"
	"agentroom/internal/room"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Reopening descriptors and replaying immutable metadata must not carry the old
// controller's release authority across restart. Physical boundaries are mocked;
// this verifies recovery protocol behavior, never deployed custody readiness.
func TestRecordedArtifactRecoveryRequiresCurrentCustodyAndAuthority(t *testing.T) {
	for _, mode := range []string{"healthy", "editing disabled", "missing bytes", "changed bytes", "unsafe root", "physical denial", "incident", "room revoked", "ack failure"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			s, db, ctx, scope := artifactJournalFixture(t, true)
			journal, err := newControllerArtifactJournal(s, &syntheticArtifactGuard{})
			if err != nil {
				t.Fatal(err)
			}
			store, dir := artifactStore(t, journal)
			receipt, err := store.Put(ctx, scope, 128, strings.NewReader("retained synthetic result"))
			if err != nil {
				t.Fatal(err)
			}
			manifest, err := s.recordStoredArtifactManifest(ctx, scope.ExecutionScope, "report", store)
			if err != nil {
				t.Fatal(err)
			}
			actor := room.Actor{ID: "recovery-reader", Kind: "human", Enabled: true, Revision: 1, RoomIDs: []string{scope.RoomID}}
			db.seed("actor.revised", actor)
			readCtx := context.WithValue(ctx, actorContextKey{}, actor.ID)
			// The restarted controller gets neither the old journal nor its guard.
			restarted := &server{db: db, integrityStopPath: filepath.Join(t.TempDir(), "integrity-stop")}
			guard := &syntheticArtifactGuard{}
			releases := 0
			guard.release = func(context.Context, room.State, objectcustody.Receipt) error {
				releases++
				if mode == "physical denial" {
					return errors.New("synthetic current physical refusal")
				}
				return nil
			}
			freshJournal, err := newControllerArtifactJournal(restarted, guard)
			if err != nil {
				t.Fatal(err)
			}
			root, err := os.OpenRoot(dir)
			if err != nil {
				t.Fatal(err)
			}
			defer root.Close()
			freshStore, err := objectcustody.New(root, freshJournal)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(dir, "object-"+receipt.Intent.ObjectID+".blob")
			switch mode {
			case "editing disabled":
				db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": false})
			case "missing bytes":
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
			case "changed bytes":
				if err := os.WriteFile(path, []byte("changed synthetic result"), 0600); err != nil {
					t.Fatal(err)
				}
			case "unsafe root":
				if err := os.Chmod(dir, 0755); err != nil {
					t.Fatal(err)
				}
				defer os.Chmod(dir, 0700)
			case "incident":
				db.seed("incident.opened", map[string]any{"reason": "synthetic recovery incident"})
			case "room revoked":
				actor.RoomIDs = nil
				actor.Revision++
				db.seed("actor.revised", actor)
			case "ack failure":
				db.failInsertAt = 1
			}
			raw, err := restarted.readRecordedArtifactMember(readCtx, scope.ExecutionID, "report", manifest.Digest, "report.txt", freshStore)
			healthy := mode == "healthy" || mode == "editing disabled"
			if healthy {
				if err != nil || string(raw) != "retained synthetic result" || releases < 2 {
					t.Fatal("restart did not reauthorize actual bytes", err, releases)
				}
			} else if err == nil || raw != nil {
				t.Fatal("historical manifest bypassed current recovery refusal", mode)
			}
			if mode == "ack failure" && !restarted.integrityFault {
				t.Fatal("uncertain read acknowledgment did not stop controller")
			}
			// A second replay must preserve charges and the original immutable
			// association even if actual bytes or current authority are unavailable.
			state := artifactState(t, &server{db: db}, ctx)
			acks := int64(1)
			if healthy {
				acks = 3
			}
			if state.ArtifactCustody[receipt.Intent.ObjectID].ReadAcknowledgments != acks || state.ArtifactManifests[artifactManifestKey(manifest)].Digest != manifest.Digest || artifactTotal(t, &server{db: db}, ctx) != 128 {
				t.Fatal("recovery erased evidence, replaced manifest or refunded capacity")
			}
		})
	}
}
