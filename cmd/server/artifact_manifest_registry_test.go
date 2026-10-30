//go:build linux || darwin

package main

import (
	"agentroom/internal/artifactmanifest"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestArtifactManifestRecordedOnceAndReplayed(t *testing.T) {
	t.Parallel()
	s, db, ctx, scope := artifactJournalFixture(t, true)
	journal, _ := newControllerArtifactJournal(s, &syntheticArtifactGuard{})
	store, _ := artifactStore(t, journal)
	receipt, err := store.Put(ctx, scope, 128, strings.NewReader("synthetic persistent output"))
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := s.recordStoredArtifactManifest(ctx, scope.ExecutionScope, "report", store)
	if err != nil {
		t.Fatal(err)
	}
	restarted := &server{db: db}
	st := artifactState(t, restarted, ctx)
	key := artifactManifestKey(manifest)
	if st.ArtifactManifests[key].Digest != manifest.Digest || st.ArtifactManifests[key].Files[0].Receipt != receipt {
		t.Fatal("manifest linkage lost on replay")
	}
	manifest.Files[0].Name = "caller mutation"
	again := artifactState(t, restarted, ctx)
	if again.ArtifactManifests[key].Files[0].Name != "report.txt" {
		t.Fatal("caller changed immutable history")
	}
	if m, err := s.recordStoredArtifactManifest(ctx, scope.ExecutionScope, "report", store); err == nil || m.Files != nil {
		t.Fatal("record overwritten/adopted")
	}
	count := 0
	for _, e := range db.snapshot() {
		if e.kind == "artifact.manifest.recorded" {
			count++
		}
	}
	if count != 1 || again.Tasks[0].Status == "Accepted" {
		t.Fatal("manifest became acceptance/duplicate", count)
	}
	public, _ := json.Marshal(again)
	if strings.Contains(string(public), "artifactManifests") || strings.Contains(string(public), receipt.Intent.ObjectID) {
		t.Fatal("internal manifest leaked through generic state")
	}
}
func TestArtifactManifestCommitFailureRetainsCustodyWithoutSuccess(t *testing.T) {
	t.Parallel()
	s, db, ctx, scope := artifactJournalFixture(t, true)
	journal, _ := newControllerArtifactJournal(s, &syntheticArtifactGuard{})
	wrapped := &manifestAckJournal{Journal: journal}
	store, _ := artifactStore(t, wrapped)
	receipt, err := store.Put(ctx, scope, 128, strings.NewReader("synthetic retained output"))
	if err != nil {
		t.Fatal(err)
	}
	wrapped.after = func() { db.failAt = "commit" }
	manifest, err := s.recordStoredArtifactManifest(ctx, scope.ExecutionScope, "report", store)
	if err == nil || manifest.Files != nil || !s.integrityFault {
		t.Fatal("uncertain commit claimed success")
	}
	st := artifactState(t, &server{db: db}, t.Context())
	if len(st.ArtifactManifests) != 0 || st.ArtifactCustody[receipt.Intent.ObjectID].Receipt == nil || st.ArtifactCustody[receipt.Intent.ObjectID].ReadAcknowledgments != 1 {
		t.Fatal("uncertain registry lost custody/history")
	}
	if total := artifactTotal(t, &server{db: db}, t.Context()); total != 128 {
		t.Fatal("uncertain metadata refunded bytes", total)
	}
}
func TestArtifactManifestProjectionRejectsForgeryAndClientInjection(t *testing.T) {
	t.Parallel()
	s, db, ctx, scope := artifactJournalFixture(t, true)
	journal, _ := newControllerArtifactJournal(s, &syntheticArtifactGuard{})
	store, _ := artifactStore(t, journal)
	if _, err := store.Put(ctx, scope, 128, strings.NewReader("synthetic checked bytes")); err != nil {
		t.Fatal(err)
	}
	manifest, err := s.resolveStoredArtifactManifest(ctx, scope.ExecutionScope, "report", store)
	if err != nil {
		t.Fatal(err)
	}
	base := artifactState(t, s, ctx)
	for _, mode := range []string{"good", "actor", "missing stamp", "digest", "format", "missing member", "no read ack", "unknown field", "duplicate"} {
		t.Run(mode, func(t *testing.T) {
			state := artifactState(t, s, ctx)
			m := manifest
			m.Files = append([]artifactmanifest.File(nil), m.Files...)
			event := artifactManifestEvent{Version: 1, ActorID: ownerID, PerformedBy: ownerID, Manifest: m}
			switch mode {
			case "actor":
				event.ActorID = "bot-openai"
				event.PerformedBy = event.ActorID
			case "missing stamp":
				event.PerformedBy = ""
			case "digest":
				event.Manifest.Digest = strings.Repeat("1", 64)
			case "format":
				event.Manifest.Files[0].StructuralCheck = "invented"
			case "missing member":
				event.Manifest.Files = nil
			case "no read ack":
				for id, object := range state.ArtifactCustody {
					object.ReadAcknowledgments = 0
					state.ArtifactCustody[id] = object
				}
			}
			raw, _ := json.Marshal(event)
			if mode == "unknown field" {
				raw = append(raw[:len(raw)-1], []byte(`,"extra":true}`)...)
			}
			err := projectArtifactManifest(&state, raw)
			if mode == "good" || mode == "duplicate" {
				if err != nil {
					t.Fatal(err)
				}
				if mode == "duplicate" && projectArtifactManifest(&state, raw) == nil {
					t.Fatal("duplicate replay accepted")
				}
			} else if err == nil || len(state.ArtifactManifests) != 0 {
				t.Fatal("forged metadata projected", mode)
			}
		})
	}
	handler := s.auth.protect(s.actorGuard(http.HandlerFunc(s.events)))
	payload, _ := json.Marshal(artifactManifestEvent{Version: 1, ActorID: ownerID, PerformedBy: ownerID, Manifest: manifest})
	before := len(db.snapshot())
	response := mockRequest(handler, "POST", "/api/events", `{"kind":"artifact.manifest.recorded","payload":`+string(payload)+`}`)
	if response.Code != 400 || !strings.Contains(response.Body.String(), "Unsupported operation") || len(db.snapshot()) != before {
		t.Fatal("client injected trusted manifest", response.Code, response.Body.String())
	}
	if len(base.ArtifactManifests) != 0 {
		t.Fatal("test prestate already recorded")
	}
}

func TestArtifactManifestRecordingAfterEditingDisabled(t *testing.T) {
	t.Parallel()
	s, db, ctx, scope := artifactJournalFixture(t, true)
	journal, _ := newControllerArtifactJournal(s, &syntheticArtifactGuard{})
	store, _ := artifactStore(t, journal)
	if _, err := store.Put(ctx, scope, 128, strings.NewReader("synthetic historical bytes")); err != nil {
		t.Fatal(err)
	}
	db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": false})
	manifest, err := s.recordStoredArtifactManifest(ctx, scope.ExecutionScope, "report", store)
	if err != nil || artifactState(t, s, ctx).ArtifactManifests[artifactManifestKey(manifest)].Digest != manifest.Digest {
		t.Fatal("editing flag disabled mandatory historical verification", err)
	}
}

func TestArtifactManifestConcurrentRecordHasOneWinner(t *testing.T) {
	t.Parallel()
	s, db, ctx, scope := artifactJournalFixture(t, true)
	journal, _ := newControllerArtifactJournal(s, &syntheticArtifactGuard{})
	store, _ := artifactStore(t, journal)
	if _, err := store.Put(ctx, scope, 128, strings.NewReader("synthetic concurrent bytes")); err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	results := make(chan error, 2)
	for range 2 {
		go func() {
			<-start
			_, err := s.recordStoredArtifactManifest(ctx, scope.ExecutionScope, "report", store)
			results <- err
		}()
	}
	close(start)
	winners := 0
	for range 2 {
		if <-results == nil {
			winners++
		}
	}
	records := 0
	for _, event := range db.snapshot() {
		if event.kind == "artifact.manifest.recorded" {
			records++
		}
	}
	if winners != 1 || records != 1 || len(artifactState(t, s, ctx).ArtifactManifests) != 1 {
		t.Fatal("duplicate concurrent publication", winners, records)
	}
}
