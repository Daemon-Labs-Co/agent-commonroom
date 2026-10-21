package main

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"context"
	"encoding/base64"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Exercise signed controller allocation through the owned Codex assembly path.
// Docker and the provider protocol peer are synthetic; no real provider runs.
func driveP0HarnessFreshPair(t *testing.T, scenario pendingP0Case) pendingP0Observation {
	t.Helper()
	current := scenario.ID == "upgrade-already-current"
	if (!current && scenario.ID != "upgrade-reviewed") || (current && scenario.Given["available"] != "current-reviewed") || (!current && scenario.Given["available"] != "new-compatible-reviewed") {
		t.Fatal("unknown fresh harness fixture")
	}
	review, key, private, _ := harnessFixture(t)
	priorReview := review
	if !current {
		priorReview.Sequence = 1
		priorReview.Version = "synthetic-reviewed-1"
		priorReview.Image = "registry.example/engine@sha256:" + strings.Repeat("b", 64)
	}
	path := filepath.Join(t.TempDir(), "review.json")
	if err := os.WriteFile(path, signHarness(t, priorReview, private), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("COMMONROOM_HARNESS_UPGRADE_CHECKS", "true")
	t.Setenv("COMMONROOM_HARNESS_REVIEW_FILE", path)
	t.Setenv("COMMONROOM_HARNESS_REVIEW_KEY", base64.StdEncoding.EncodeToString(key))
	s, db, _, priorRegistration := managedPairFixtureVersion(t, workerWorkspaceLifecycleVersion)
	state, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	prior := state.ManagedJobs[priorRegistration.JobID]
	if prior.Harness.Image != priorReview.Image || prior.Harness.Version != priorReview.Version {
		t.Fatal("initial allocation lost reviewed version")
	}
	if err := os.WriteFile(path, signHarness(t, review, private), 0600); err != nil {
		t.Fatal(err)
	}
	p := room.Participant{ID: "fresh-agent", RoomID: prior.RoomID, ConnectionID: prior.ConnectionID, Revision: 1}
	db.seed("participant.created", p)
	ctx := context.WithValue(t.Context(), actorContextKey{}, ownerID)
	ctx = s.managedExecutionJobContextVersion(ctx, p, prior.RepositoryDigest, func() (string, error) { return prior.RepositoryDigest, nil }, workerWorkspaceLifecycleVersion)
	registration := ctx.Value(managedJobContextKey{}).(managedJobRegistration)
	if err := registration.Register(prior.Worker, prior.Provider); err != nil {
		t.Fatal("new reviewed allocation refused", err)
	}
	// Replay through another controller before starting any worker resources.
	state, err = (&server{db: db}).load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	job := state.ManagedJobs[registration.JobID]
	out := pendingP0Observation{State: "allow", Values: map[string]string{}}
	if job.Harness.CheckedAt.Before(job.AllocatedAt) || job.Harness.ReviewDigest == "" || job.JobID == prior.JobID {
		t.Fatal("fresh allocation lacks durable review evidence")
	}
	out.Effects = append(out.Effects, "upgrade_checked")
	if !current && validManagedHarness(job) && job.Harness.Sequence > prior.Harness.Sequence {
		out.Effects = append(out.Effects, "compatibility_verified")
	}
	if job.Harness.Image != review.Image || job.Harness.Version != review.Version {
		t.Fatal("controller did not select reviewed release")
	}
	out.Effects = append(out.Effects, "reviewed_image_selected")
	if state.ManagedJobs[prior.JobID] != prior {
		out.Effects = append(out.Effects, "running_container_mutated")
	}
	request := pairRequest()
	request.Version, request.JobID, request.Provider = job.Version, job.JobID, job.Provider
	request.Key, request.Repository, request.Deadline = managedJobFence(job).Key, job.RepositoryDigest, job.Deadline
	worker, err := admitWorkerPair(t.Context(), t.TempDir(), request, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer worker.Close()
	scope := toolbroker.Scope{Version: 1, JobID: job.JobID, Room: job.RoomID, Participant: job.ParticipantID, Connection: job.ConnectionID, Provider: job.Provider, RepositoryDigest: job.RepositoryDigest, Deadline: job.Deadline}
	counter := &transportTestAuditor{}
	runner := &codexIntegrationDocker{containers: map[string]map[string]any{}, scope: scope, engineDone: make(chan error, 1), audit: counter, mode: "success"}
	client, peer := net.Pipe()
	defer peer.Close()
	auditDone := make(chan error, 1)
	go func() {
		auditDone <- toolbroker.ServeAudit(worker.ctx, peer, scope, codexIntegrationAuditor{counter, "success"})
	}()
	workerCtx := context.WithValue(worker.ctx, workerLocalJobContextKey{}, job.JobID)
	session, token, err := newWorkerToolSession(workerCtx, scope, []string{"read"}, "ghcr.io/example/tool@sha256:"+strings.Repeat("a", 64), syntheticDockerRoot, time.Now(), client, runner)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	auth := t.TempDir()
	credentialDir := filepath.Join(auth, connectionAuthKey(job.ConnectionID), job.Provider)
	if err := os.MkdirAll(credentialDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(credentialDir, "auth.json"), []byte(syntheticEngineCredential), 0600); err != nil {
		t.Fatal(err)
	}
	bundle := &pairBundleFixture{Reader: strings.NewReader("synthetic approved bundle")}
	result, record, runErr := worker.RunCodex(t.Context(), session, token, bundle, workerCodexRunConfig{ManagedJob: &job, AuthRoot: auth, GatewayRoot: shortGatewayRoot(t), Prompt: "Synthetic upgrade acceptance", Model: "reviewed-model"}, runner)
	if runErr == nil || record.ShutdownConfirmed || record.Pair == nil || record.Pair.Workspace == nil || record.Pair.Workspace.Removed || runner.volume == nil {
		t.Fatal("retained source was discarded or mislabeled settled", runErr)
	}
	if runner.creates == 2 && runner.starts == 2 && runner.bootstrapped && record.Pair.Engine.ID != record.Pair.Tool.ID && record.Pair.Engine.ID != "" && record.Pair.Tool.ID != "" && bundle.closes.Load() == 1 {
		out.Effects = append(out.Effects, "fresh_pair")
	} else {
		t.Fatal("fresh isolated pair did not bootstrap exactly once")
	}
	if runner.engineImage != job.Harness.Image {
		t.Fatal("owned startup bypassed reviewed image", runner.engineImage)
	}
	if result.Text != "Synthetic turn completed" || counter.calls.Load() != 2 || runner.helpers != 1 {
		t.Fatal("synthetic reviewed engine did not complete callback")
	}
	out.Effects = append(out.Effects, "callback_dispatched")
	if !record.Pair.Engine.Absent || !record.Pair.Tool.Absent || runner.removed != 2 || len(runner.containers) != 0 {
		t.Fatal("reviewed pair cleanup lacked both-role absence")
	}
	out.Effects = append(out.Effects, "fenced_pair", "workspace_retained")
	stored, err := readWorkerRecord(worker.dir, job.JobID)
	if err != nil || !reflect.DeepEqual(stored, record) {
		t.Fatal("worker journal lost final pair/source evidence", err)
	}
	replayed, err := (&server{db: db}).load(t.Context())
	if err != nil || replayed.ManagedJobs[prior.JobID] != prior || replayed.ManagedJobs[job.JobID] != job {
		t.Fatal("worker run changed immutable controller allocations", err)
	}
	select {
	case err := <-runner.engineDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("synthetic engine peer did not stop")
	}
	awaitAuditResult(t, auditDone)
	if !current {
		out.Values["running_version"] = job.Harness.Version
		out.Values["version_audited"] = strconv.FormatBool(validManagedHarness(job))
	}
	return out
}
