package main

import (
	"agentroom/internal/room"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func harnessFixture(t *testing.T) (harnessReview, ed25519.PublicKey, ed25519.PrivateKey, time.Time) {
	t.Helper()
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	return harnessReview{Schema: 1, Sequence: 2, Provider: "codex", Version: "synthetic-reviewed-2", Image: "registry.example/engine@sha256:" + strings.Repeat("a", 64), Profile: "codex-separated-callback-v1", AuthProtocol: 1, AuditProtocol: 2, WorkspaceProtocol: 4, IssuedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Hour)}, public, private, now
}
func signHarness(t *testing.T, review harnessReview, key ed25519.PrivateKey) []byte {
	t.Helper()
	payload, err := json.Marshal(review)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(harnessReviewEnvelope{Review: payload, Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(key, payload))})
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
func TestHarnessReviewSelectionAndStartup(t *testing.T) {
	t.Parallel()
	review, key, private, now := harnessFixture(t)
	raw := signHarness(t, review, private)
	selected, err := selectHarnessReview("true", func() ([]byte, error) { return raw, nil }, key, now, nil)
	if err != nil || selected.Image != review.Image || selected.Sequence != 2 {
		t.Fatal(selected, err)
	}
	// The same reviewed version can be selected for another fresh allocation.
	prior := map[string]room.ManagedJob{"previous": {Harness: selected}}
	if _, err := selectHarnessReview("true", func() ([]byte, error) { return raw, nil }, key, now, prior); err != nil {
		t.Fatal(err)
	}
	job := quotaJob(1, true)
	job.Provider = "codex"
	job.AllocatedAt = now.Add(-time.Second)
	job.UpdatedAt = now
	job.Harness = selected
	if !validManagedJob(job) {
		t.Fatal("reviewed allocation invalid")
	}
	if image, err := selectedHarnessImage(job, "", now); err != nil || image != review.Image {
		t.Fatal(image, err)
	}
	if _, err := selectedHarnessImage(job, "other@sha256:"+strings.Repeat("b", 64), now); err == nil {
		t.Fatal("accepted worker override")
	}
	if _, err := selectedHarnessImage(job, "", review.ExpiresAt); err == nil {
		t.Fatal("accepted expired startup")
	}
	if _, err := selectedHarnessImage(job, "", now.Add(-time.Second)); err == nil {
		t.Fatal("accepted startup before review")
	}
	job.Provider = "claude"
	if validManagedJob(job) {
		t.Fatal("accepted unsupported provider")
	}
}
func TestHarnessReviewRejectsUnsignedIncompatibleAndAmbiguousInput(t *testing.T) {
	t.Parallel()
	cases := []string{"schema", "sequence", "provider", "profile", "auth", "audit", "workspace", "future", "expired", "tag", "digest", "version", "signature", "key", "duplicate", "unknown", "case", "oversized"}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			review, key, private, now := harnessFixture(t)
			switch name {
			case "schema":
				review.Schema = 2
			case "sequence":
				review.Sequence = 0
			case "provider":
				review.Provider = "claude"
			case "profile":
				review.Profile = "native"
			case "auth":
				review.AuthProtocol = 2
			case "audit":
				review.AuditProtocol = 1
			case "workspace":
				review.WorkspaceProtocol = 3
			case "future":
				review.IssuedAt = now.Add(time.Second)
			case "expired":
				review.ExpiresAt = now
			case "tag":
				review.Image = "registry.example/engine:latest"
			case "digest":
				review.Image = "registry.example/engine@sha256:invalid"
			case "version":
				review.Version = "version\nunsafe"
			}
			raw := signHarness(t, review, private)
			switch name {
			case "signature":
				raw = []byte(strings.Replace(string(raw), "synthetic-reviewed-2", "synthetic-reviewed-3", 1))
			case "key":
				key = make(ed25519.PublicKey, 32)
			case "duplicate":
				raw = []byte(strings.Replace(string(raw), `"schema":1`, `"schema":1,"schema":1`, 1))
			case "unknown":
				raw = []byte(strings.Replace(string(raw), `"schema":1`, `"unknown":1,"schema":1`, 1))
			case "case":
				raw = []byte(strings.Replace(string(raw), `"schema":1`, `"Schema":1`, 1))
			case "oversized":
				raw = append(raw, make([]byte, 4097)...)
			}
			if _, err := verifyHarnessReview(raw, key, now); err == nil {
				t.Fatal("accepted invalid review")
			}
		})
	}
}
func TestHarnessReviewFlagAvailabilityAndRollback(t *testing.T) {
	t.Parallel()
	review, key, private, now := harnessFixture(t)
	raw := signHarness(t, review, private)
	for _, flag := range []string{"", "false"} {
		selected, err := selectHarnessReview(flag, func() ([]byte, error) { t.Fatal("disabled check read catalog"); return nil, nil }, nil, now, nil)
		if err != nil || selected != (room.HarnessSelection{}) {
			t.Fatal(selected, err)
		}
	}
	if _, err := selectHarnessReview("TRUE", nil, key, now, nil); err == nil {
		t.Fatal("invalid flag accepted")
	}
	if _, err := selectHarnessReview("true", func() ([]byte, error) { return nil, errors.New("synthetic unavailable") }, key, now, nil); err == nil {
		t.Fatal("unavailable catalog accepted")
	}
	for _, prior := range []room.HarnessSelection{{Sequence: 3}, {Sequence: 2, ReviewDigest: strings.Repeat("f", 64)}} {
		if _, err := selectHarnessReview("true", func() ([]byte, error) { return raw, nil }, key, now, map[string]room.ManagedJob{"previous": {Harness: prior}}); err == nil {
			t.Fatal("rollback/equivocation accepted")
		}
	}
	legacy := quotaJob(1, true)
	if image, err := selectedHarnessImage(legacy, "explicit", now); err != nil || image != "explicit" {
		t.Fatal("disabled legacy contract changed")
	}
}
func TestHarnessReviewBindingCannotChangeOnReplay(t *testing.T) {
	t.Parallel()
	s, db, _, registration := managedPairFixtureVersion(t, workerWorkspaceLifecycleVersion)
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	job := st.ManagedJobs[registration.JobID]
	review, key, private, _ := harnessFixture(t)
	selection, err := verifyHarnessReview(signHarness(t, review, private), key, job.UpdatedAt)
	if err != nil {
		t.Fatal(err)
	}
	// A later lifecycle event cannot attach or replace the admission's review.
	job.Harness = selection
	job.UpdatedAt = job.UpdatedAt.Add(time.Second)
	db.seed("worker.job.updated", job)
	st, err = s.load(t.Context())
	if err == nil && st.ManagedJobs[registration.JobID].Harness != (room.HarnessSelection{}) {
		t.Fatal("replay changed immutable image binding")
	}
}

// Environment configuration is serialized by testing.Setenv; all other review
// tests use explicit inputs and can run in parallel without deployment mutation.
func TestHarnessReviewRegistrationUsesDeploymentPolicy(t *testing.T) {
	review, key, private, _ := harnessFixture(t)
	path := filepath.Join(t.TempDir(), "review.json")
	if err := os.WriteFile(path, signHarness(t, review, private), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("COMMONROOM_HARNESS_UPGRADE_CHECKS", "true")
	t.Setenv("COMMONROOM_HARNESS_REVIEW_FILE", path)
	t.Setenv("COMMONROOM_HARNESS_REVIEW_KEY", base64.StdEncoding.EncodeToString(key))
	s, _, _, registration := managedPairFixtureVersion(t, workerWorkspaceLifecycleVersion)
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	job := st.ManagedJobs[registration.JobID]
	if job.Harness.Image != review.Image || job.Harness.Version != review.Version {
		t.Fatal("review not durably recorded", job.Harness)
	}
	for _, mode := range []string{"unavailable", "unauthorized", "audit failure"} {
		t.Run(mode, func(t *testing.T) {
			s, db, _ := mockServer(t)
			db.seed("connection.created", room.Connection{ID: "pair-connection", Name: "Synthetic", Protocol: "CLI", Provider: "openai", Endpoint: "ssh://worker", Revision: 1})
			p := room.Participant{ID: "pair-agent", RoomID: "engineering", ConnectionID: "pair-connection", Revision: 1}
			db.seed("participant.created", p)
			actor := ownerID
			if mode == "unauthorized" {
				actor = "unknown-human"
			}
			ctx := context.WithValue(t.Context(), actorContextKey{}, actor)
			repo := strings.Repeat("b", 64)
			ctx = s.managedExecutionJobContextVersion(ctx, p, repo, func() (string, error) { return repo, nil }, workerWorkspaceLifecycleVersion)
			if mode == "unavailable" {
				t.Setenv("COMMONROOM_HARNESS_REVIEW_FILE", filepath.Join(t.TempDir(), "absent"))
			}
			if mode == "audit failure" {
				db.failAt = "insert"
			}
			registration := ctx.Value(managedJobContextKey{}).(managedJobRegistration)
			if err := registration.Register("worker", "codex"); err == nil {
				t.Fatal("unsafe registration accepted")
			}
			if len(s.activeJobs) != 0 {
				t.Fatal("failed allocation became active")
			}
		})
	}
}
