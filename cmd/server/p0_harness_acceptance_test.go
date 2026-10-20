package main

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

type harnessDigestDocker struct {
	calls [][]string
	image string
}

func (d *harnessDigestDocker) Run(ctx context.Context, args []string, _ []byte, out io.Writer) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	d.calls = append(d.calls, append([]string(nil), args...))
	switch args[0] {
	case "ps":
		return nil
	case "image":
		return json.NewEncoder(out).Encode(map[string]any{
			"Id":          "sha256:" + strings.Repeat("c", 64),
			"RepoDigests": []string{strings.Split(d.image, "@sha256:")[0] + "@sha256:" + strings.Repeat("f", 64)},
			"Config":      map[string]any{"User": "1000:1000", "WorkingDir": "/tmp", "Entrypoint": []string{"/usr/local/bin/node", "/opt/commonroom/engine-entrypoint.cjs"}},
		})
	}
	return errors.New("unexpected container mutation")
}

func driveP0HarnessDigestRefusal(t *testing.T, scenario pendingP0Case) pendingP0Observation {
	t.Helper()
	if scenario.Given["available"] != "digest-mismatch" {
		t.Fatal("unknown image refusal input")
	}
	review, key, private, now := harnessFixture(t)
	selected, err := selectHarnessReview("true", func() ([]byte, error) { return signHarness(t, review, private), nil }, key, now, nil)
	if err != nil {
		t.Fatal(err)
	}
	runner := &harnessDigestDocker{image: selected.Image}
	scope := toolbroker.Scope{Version: 1, JobID: strings.Repeat("a", 32), Room: "room", Participant: "agent", Connection: "connection", Provider: "codex", RepositoryDigest: strings.Repeat("b", 64), Deadline: now.Add(time.Minute)}
	runtime, err := toolbroker.NewEngineContainerRuntime(t.Context(), scope, selected.Image, "/private/synthetic/gateway.sock", runner)
	if err != nil {
		t.Fatal(err)
	}
	defer runtime.Close()
	_, err = runtime.Create(t.Context())
	actual := pendingP0Observation{State: "allow", Values: map[string]string{}}
	if err != nil {
		actual.State = "block"
	}
	if errors.Is(err, toolbroker.ErrRejected) {
		actual.Effects = append(actual.Effects, "upgrade_failure_reported")
	}
	for _, call := range runner.calls {
		switch call[0] {
		case "create":
			actual.Effects = append(actual.Effects, "engine_created")
		case "start":
			actual.Effects = append(actual.Effects, "engine_started")
		}
	}
	return actual
}
