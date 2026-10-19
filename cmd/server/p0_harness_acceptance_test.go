package main

import (
	"agentroom/internal/room"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These negative cases exercise the real allocation gate with mocked audit
// storage. They establish rejection before launch authority exists, not live
// image compatibility or successful provider execution.
func driveP0HarnessReviewRefusal(t *testing.T, scenario pendingP0Case) pendingP0Observation {
	t.Helper()
	review, key, private, _ := harnessFixture(t)
	path := filepath.Join(t.TempDir(), "review.json")
	switch scenario.Given["available"] {
	case "incompatible":
		review.WorkspaceProtocol = 3
	case "unreviewed":
		_, other, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			t.Fatal(err)
		}
		private = other
	case "check-fails":
	default:
		t.Fatal("unknown harness refusal input")
	}
	if scenario.Given["available"] != "check-fails" {
		if err := os.WriteFile(path, signHarness(t, review, private), 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("COMMONROOM_HARNESS_UPGRADE_CHECKS", "true")
	t.Setenv("COMMONROOM_HARNESS_REVIEW_FILE", path)
	t.Setenv("COMMONROOM_HARNESS_REVIEW_KEY", base64.StdEncoding.EncodeToString(key))
	s, db, _ := mockServer(t)
	db.seed("connection.created", room.Connection{ID: "pair-connection", Name: "Synthetic", Protocol: "CLI", Provider: "openai", Endpoint: "ssh://worker", Revision: 1})
	p := room.Participant{ID: "pair-agent", RoomID: "engineering", ConnectionID: "pair-connection", Revision: 1}
	db.seed("participant.created", p)
	ctx := context.WithValue(t.Context(), actorContextKey{}, ownerID)
	repo := strings.Repeat("b", 64)
	ctx = s.managedExecutionJobContextVersion(ctx, p, repo, func() (string, error) { return repo, nil }, workerWorkspaceLifecycleVersion)
	registration := ctx.Value(managedJobContextKey{}).(managedJobRegistration)
	before := len(db.snapshot())
	err := registration.Register("worker", "codex")
	actual := pendingP0Observation{State: "allow", Values: map[string]string{}}
	if err != nil {
		actual.State = "block"
	}
	if errors.Is(err, errHarnessReview) {
		actual.Effects = append(actual.Effects, "upgrade_failure_reported")
	}
	if len(db.snapshot()) != before {
		actual.Effects = append(actual.Effects, "allocation_recorded")
	}
	if len(s.activeJobs) != 0 {
		actual.Effects = append(actual.Effects, "allocation_activated")
	}
	return actual
}
