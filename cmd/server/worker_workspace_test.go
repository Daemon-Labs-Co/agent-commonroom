package main

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"context"
	"encoding/json"
	"errors"
	"io"
	"path"
	"strings"
	"testing"
	"time"
)

type workspaceLaunchFake struct {
	*pairLaunchFake
	failWorkspace string
}

func (f *workspaceLaunchFake) WorkspaceIntent(ctx context.Context) (toolbroker.WorkspaceVolumeIdentity, error) {
	f.calls = append(f.calls, "workspace-intent")
	return toolbroker.WorkspaceVolumeIdentity{Version: 1, Name: "commonroom-tool-work-" + f.supervisor.record.JobID, Nonce: strings.Repeat("c", 32)}, nil
}
func (f *workspaceLaunchFake) PrepareVolume(ctx context.Context, intent toolbroker.WorkspaceVolumeIdentity) (toolbroker.WorkspaceVolumeIdentity, error) {
	f.calls = append(f.calls, "prepare-volume")
	record, err := readWorkerRecord(f.supervisor.dir, f.supervisor.record.JobID)
	if err != nil || record.Pair.Workspace == nil || !record.Pair.Workspace.CreationStarted || record.Pair.Workspace.Identity != intent || record.Pair.Engine.CreationStarted || record.Pair.Tool.CreationStarted {
		f.t.Fatal("volume create preceded journal or followed container creation")
	}
	if f.failWorkspace == "create" {
		return intent, errors.New("synthetic volume create failure")
	}
	intent.CreatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	if f.failWorkspace == "binding" {
		intent.Nonce = strings.Repeat("d", 32)
	}
	return intent, nil
}
func workspacePair(t *testing.T) (string, *workerPairSupervisor, *workspaceLaunchFake) {
	t.Helper()
	dir := t.TempDir()
	request := pairRequest()
	request.Version = workerWorkspaceLifecycleVersion
	s, err := admitWorkerPair(t.Context(), dir, request, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Close)
	f := &workspaceLaunchFake{pairLaunchFake: &pairLaunchFake{t: t, supervisor: s}}
	return dir, s, f
}
func TestWorkspacePairJournalsVolumeBeforeContainers(t *testing.T) {
	t.Parallel()
	dir, s, f := workspacePair(t)
	if err := s.Start(t.Context(), f); err != nil {
		t.Fatal(err)
	}
	record, err := readWorkerRecord(dir, s.record.JobID)
	if err != nil || record.Version != 4 || record.Execution != workerWorkspaceExecution || record.Pair.Workspace == nil || record.Pair.Workspace.Identity.CreatedAt == "" || record.Status != "running" {
		t.Fatal("workspace lifecycle not bound", err)
	}
	if len(f.calls) < 3 || f.calls[0] != "workspace-intent" || f.calls[1] != "prepare-volume" || f.calls[2] != "create:tool" {
		t.Fatalf("bad ordering %v", f.calls)
	}
	downgraded := record
	downgraded.Version = 3
	downgraded.Execution = workerPairExecution
	if validWorkerRecord(downgraded) {
		t.Fatal("workspace record downgraded to container-only proof")
	}
	record.ShutdownConfirmed = true
	record.Status = "cancelled"
	record.Pair.Engine.Absent = true
	record.Pair.Tool.Absent = true
	if validWorkerRecord(record) || workerJobSettled(record) {
		t.Fatal("container absence settled remaining workspace")
	}
}
func TestWorkspacePairFailureCannotStartContainers(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"create", "binding", "missing-adapter"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			dir, s, f := workspacePair(t)
			f.failWorkspace = mode
			var adapter workerPairLaunchAdapter = f
			if mode == "missing-adapter" {
				adapter = containerOnlyAdapter{f.pairLaunchFake}
			}
			if err := s.Start(t.Context(), adapter); err == nil {
				t.Fatal("workspace failure accepted")
			}
			record, err := readWorkerRecord(dir, s.record.JobID)
			if err != nil || record.Pair.Engine.CreationStarted || record.Pair.Tool.CreationStarted || record.ShutdownConfirmed {
				t.Fatal("containers started after failed workspace")
			}
			before := len(f.calls)
			if mode != "missing-adapter" {
				if err = s.Start(t.Context(), adapter); err == nil || len(f.calls) != before {
					t.Fatal("uncertain workspace replayed")
				}
			}
		})
	}
}

type workspaceCleanupDocker struct {
	*pairFakeDocker
	t                *testing.T
	record           workerJobRecord
	volume           bool
	removed          int
	mode             string
	observedCreation string
}

func (f *workspaceCleanupDocker) Run(ctx context.Context, args []string, in []byte, out io.Writer) error {
	if args[0] == "info" {
		_, err := io.WriteString(out, "/home/synthetic/.local/share/docker")
		return err
	}
	if args[0] != "volume" {
		if args[0] == "ps" && len(args) > 4 && strings.HasPrefix(args[4], "volume=") && f.mode == "referenced" {
			_, err := io.WriteString(out, strings.Repeat("d", 64))
			return err
		}
		return f.pairFakeDocker.Run(ctx, args, in, out)
	}
	if f.mode == "daemon" {
		return errors.New("synthetic unavailable")
	}
	name := "commonroom-tool-work-" + f.record.JobID
	switch args[1] {
	case "ls":
		value := ""
		if f.volume {
			value = name
		}
		if f.mode == "extra" && strings.HasPrefix(args[3], "label=") {
			value = "unknown-workspace"
		}
		_, err := io.WriteString(out, value)
		return err
	case "inspect":
		if !f.volume || f.record.Pair.Workspace == nil {
			return errors.New("missing")
		}
		identity := f.record.Pair.Workspace.Identity
		if identity.CreatedAt == "" {
			if f.observedCreation == "" {
				f.observedCreation = time.Now().UTC().Format(time.RFC3339Nano)
			}
			identity.CreatedAt = f.observedCreation
		}
		if f.mode == "replacement" {
			identity.CreatedAt = time.Now().Add(time.Hour).UTC().Format(time.RFC3339Nano)
		}
		raw := map[string]any{"Name": name, "Driver": "local", "Scope": "local", "Mountpoint": path.Join("/home/synthetic/.local/share/docker", "volumes", name, "_data"), "CreatedAt": identity.CreatedAt, "Labels": map[string]string{"commonroom.managed": "true", "commonroom.execution": "separated-tool-v1", "commonroom.job": f.record.JobID, "commonroom.repository": f.record.Pair.Repository, "commonroom.workspace.nonce": identity.Nonce}}
		return json.NewEncoder(out).Encode(raw)
	case "rm":
		owned, _ := ctx.Value(workerCleanupOwnershipKey{}).(bool)
		if !owned {
			f.t.Fatal("workspace removed without creator ownership")
		}
		if f.mode == "remove" {
			return errors.New("synthetic removal failure")
		}
		if args[len(args)-1] != name {
			f.t.Fatal("foreign volume removed")
		}
		f.removed++
		f.volume = false
		_, err := io.WriteString(out, name)
		return err
	}
	return errors.New("unexpected volume action")
}
func workspaceFence(s *workerPairSupervisor) workerFenceRequest {
	return workerFenceRequest{Version: 4, DiscardWorkspace: true, Repository: s.record.Pair.Repository, JobID: s.record.JobID, Provider: s.record.Provider, Key: pairRequest().Key}
}
func newWorkspaceCleanup(t *testing.T, s *workerPairSupervisor) *workspaceCleanupDocker {
	t.Helper()
	return &workspaceCleanupDocker{pairFakeDocker: &pairFakeDocker{containers: map[string]pairFakeContainer{}}, t: t, record: s.record, volume: true}
}
func fenceWorkspace(ctx context.Context, dir string, s *workerPairSupervisor, f *workspaceCleanupDocker) (workerJobRecord, error) {
	return fenceWorkerJobContainers(ctx, dir, workspaceFence(s), func(ctx context.Context, record workerJobRecord) (workerJobRecord, error) {
		return stopWorkerPair(ctx, record, f)
	})
}
func TestWorkspaceShutdownRequiresVolumeAndLeaseProof(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"success", "daemon", "remove", "replacement", "referenced", "extra"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			dir, s, adapter := workspacePair(t)
			if err := s.Start(t.Context(), adapter); err != nil {
				t.Fatal(err)
			}
			f := newWorkspaceCleanup(t, s)
			f.mode = mode
			s.Close()
			record, err := fenceWorkspace(t.Context(), dir, s, f)
			success := mode == "success"
			if (err == nil) != success || record.ShutdownConfirmed != success {
				t.Fatalf("bad shutdown evidence %v", err)
			}
			stored, readErr := readWorkerRecord(dir, s.record.JobID)
			if readErr != nil || stored.ShutdownConfirmed != success || stored.Pair.Workspace.Removed != success {
				t.Fatal("workspace result not durably recorded", readErr)
			}
			if success {
				expected, _ := workspaceFence(s).record()
				ack := workerFenceAcknowledgement(workspaceFence(s), stored)
				if !validWorkerFenceAck(workspaceFence(s), expected, ack) || !ack.WorkspaceRemoved {
					t.Fatal("missing workspace acknowledgement")
				}
				ack.WorkspaceRemoved = false
				if validWorkerFenceAck(workspaceFence(s), expected, ack) {
					t.Fatal("container-only ack settled workspace")
				}
			}
			if mode == "replacement" && f.removed != 0 {
				t.Fatal("replacement removed")
			}
		})
	}
}
func TestWorkspaceCleanupWaitsForCreatorRelease(t *testing.T) {
	t.Parallel()
	dir, s, adapter := workspacePair(t)
	if err := s.Start(t.Context(), adapter); err != nil {
		t.Fatal(err)
	}
	f := newWorkspaceCleanup(t, s)
	ctx, cancel := context.WithTimeout(t.Context(), 25*time.Millisecond)
	defer cancel()
	if record, err := fenceWorkspace(ctx, dir, s, f); err == nil || record.ShutdownConfirmed || f.removed != 0 {
		t.Fatal("held creator allowed workspace removal")
	}
	s.Close()
	if record, err := fenceWorkspace(t.Context(), dir, s, f); err != nil || !record.ShutdownConfirmed || f.removed != 1 {
		t.Fatal("released creator failed reconciliation", err)
	}
}
func TestWorkspaceCancellationCannotDowngrade(t *testing.T) {
	t.Parallel()
	dir, s, adapter := workspacePair(t)
	if err := s.Start(t.Context(), adapter); err != nil {
		t.Fatal(err)
	}
	s.Close()
	request := workspaceFence(s)
	request.Version = 3
	request.DiscardWorkspace = false
	if _, err := fenceWorkerJobContainers(t.Context(), dir, request, func(context.Context, workerJobRecord) (workerJobRecord, error) {
		t.Fatal("legacy fence reached cleanup")
		return workerJobRecord{}, nil
	}); err == nil {
		t.Fatal("legacy cancellation accepted")
	}
	st, err := readWorkerRecord(dir, s.record.JobID)
	if err != nil || st.Version != 4 || st.ShutdownConfirmed {
		t.Fatal("downgrade changed journal")
	}
}
func TestWorkspaceManagedRegistrationAndRecoveryBinding(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	p := room.Participant{ID: "workspace-agent", RoomID: "engineering", ConnectionID: "workspace-connection"}
	db.seed("participant.created", p)
	repository := strings.Repeat("b", 64)
	ctx := context.WithValue(t.Context(), actorContextKey{}, ownerID)
	ctx = s.managedExecutionJobContextVersion(ctx, p, repository, func() (string, error) { return repository, nil }, workerWorkspaceLifecycleVersion)
	registration := ctx.Value(managedJobContextKey{}).(managedJobRegistration)
	if err := registration.Register("worker", "codex"); err != nil {
		t.Fatal(err)
	}
	state, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	job := state.ManagedJobs[registration.JobID]
	if !validManagedJob(job) || job.Version != 4 || job.Execution != workerWorkspaceExecution || managedJobFence(job).Version != 4 || job.InitiatorID != ownerID {
		t.Fatal("workspace controller provenance/protocol lost")
	}
}

func TestWorkspaceCrashAfterCreateDiscoversOriginalVolume(t *testing.T) {
	t.Parallel()
	dir, s, adapter := workspacePair(t)
	adapter.failWorkspace = "create"
	if err := s.Start(t.Context(), adapter); err == nil {
		t.Fatal("uncertain create accepted")
	}
	if s.record.Pair.Workspace == nil || s.record.Pair.Workspace.Identity.CreatedAt != "" {
		t.Fatal("lost pending creation intent")
	}
	f := newWorkspaceCleanup(t, s)
	s.Close()
	record, err := fenceWorkspace(t.Context(), dir, s, f)
	if err != nil || !record.ShutdownConfirmed || record.Pair.Workspace.Identity.CreatedAt == "" || f.removed != 1 {
		t.Fatal("pending workspace not safely reconciled", err)
	}
}
func TestWorkspaceCancellationBeforeCreationRefusesUnknownData(t *testing.T) {
	t.Parallel()
	for _, unexpected := range []bool{false, true} {
		t.Run(map[bool]string{false: "absent", true: "unknown volume"}[unexpected], func(t *testing.T) {
			t.Parallel()
			dir, s, _ := workspacePair(t)
			f := newWorkspaceCleanup(t, s)
			f.volume = unexpected
			s.Close()
			record, err := fenceWorkspace(t.Context(), dir, s, f)
			if (err == nil) == unexpected || record.ShutdownConfirmed == unexpected || f.removed != 0 {
				t.Fatal("never-created workspace adopted/deleted or absence refused", err)
			}
		})
	}
}
func TestWorkspaceOwnerRecoveryKeepsVersionAndProof(t *testing.T) {
	t.Setenv("COMMONROOM_REMOTE_WORKER", "true")
	t.Setenv("COMMONROOM_REMOTE_WORKER_ALIAS", "worker")
	for _, mode := range []string{"success", "diagnostics-off", "legacy-worker", "workspace-failure"} {
		t.Run(mode, func(t *testing.T) {
			s, db, h, registration := managedPairFixtureVersion(t, workerWorkspaceLifecycleVersion)
			state, err := s.load(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			job := state.ManagedJobs[registration.JobID]
			view, _ := managedJobFence(job).record()
			view.Deadline = job.Deadline
			if mode == "diagnostics-off" {
				db.seed("feature.flag.updated", map[string]any{"name": "jobDiagnostics", "enabled": false})
			}
			if mode == "legacy-worker" {
				view.Version = 3
				view.Execution = workerPairExecution
			}
			s.jobSnapshot = func(context.Context, string) (workerJobsSnapshot, error) {
				return workerJobsSnapshot{Version: 2, Jobs: []workerJobView{{workerJobRecord: view, LeaseState: "free"}}}, nil
			}
			calls := 0
			s.jobFence = func(ctx context.Context, alias string, request workerFenceRequest) error {
				calls++
				if request.DiscardWorkspace {
					t.Fatal("HTTP recovery inferred workspace deletion authority")
				}
				if request.Version != 4 || request.Repository != job.RepositoryDigest || alias != "worker" {
					t.Fatal("workspace recovery downgraded")
				}
				if mode == "workspace-failure" {
					return errors.New("synthetic workspace cleanup uncertain")
				}
				return nil
			}
			response := mockRequest(h, "POST", "/api/job-recovery", `{"action":"reconcile","jobId":"`+job.JobID+`","reason":"Reviewed synthetic workspace"}`)
			want := 200
			if mode == "legacy-worker" {
				want = 409
			}
			if mode == "workspace-failure" {
				want = 503
			}
			if response.Code != want {
				t.Fatalf("response %d expected %d: %s", response.Code, want, response.Body.String())
			}
			if mode == "legacy-worker" && calls != 0 {
				t.Fatal("legacy worker contacted")
			}
			latest, err := s.load(t.Context())
			if err != nil || latest.ManagedJobs[job.JobID].ShutdownConfirmed != (want == 200) {
				t.Fatal("invalid durable recovery result", err)
			}
		})
	}
}

// Deliberately hides workspace methods to test mixed-version adapters.
type containerOnlyAdapter struct{ workerPairLaunchAdapter }

func TestLegacyPairRecordsAreRecoveryOnly(t *testing.T) {
	t.Parallel()
	s, err := admitWorkerPair(t.Context(), t.TempDir(), pairRequest(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	f := &pairLaunchFake{t: t, supervisor: s}
	if err = s.Start(t.Context(), f); err == nil || len(f.calls) != 0 {
		t.Fatal("legacy pair launched without workspace evidence")
	}
}

func TestWorkspaceRetainedWithoutExplicitDiscard(t *testing.T) {
	t.Parallel()
	dir, s, adapter := workspacePair(t)
	if err := s.Start(t.Context(), adapter); err != nil {
		t.Fatal(err)
	}
	f := newWorkspaceCleanup(t, s)
	s.Close()
	request := workspaceFence(s)
	request.DiscardWorkspace = false
	record, err := fenceWorkerJobContainers(t.Context(), dir, request, func(ctx context.Context, v workerJobRecord) (workerJobRecord, error) {
		return stopWorkerPair(ctx, v, f)
	})
	if err == nil || record.ShutdownConfirmed || f.removed != 0 || !f.volume {
		t.Fatal("ordinary cancellation erased unapproved workspace")
	}
}
