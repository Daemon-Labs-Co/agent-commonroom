package main

import (
	"agentroom/internal/toolbroker"
	"agentroom/internal/workspacebootstrap"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const workerPairVersion = 1
const workerPairLifecycleVersion = 3
const workerPairExecution = "separated-pair-v1"

// workerPairRecord is additive, explicitly versioned lifecycle evidence, not an
// execution capability. Both roles share one job lease and the old logical-job
// admission slot. Actual engine/tool resource policy still needs launch verification.
// No provider, token, command or host path belongs here.
// A source fork CAN skip creation journalling, BUT invalidates crash recovery.
type workerPairRecord struct {
	Version    int                 `json:"version"`
	Repository string              `json:"repository"`
	Engine     workerPairContainer `json:"engine"`
	Tool       workerPairContainer `json:"tool"`
}
type workerPairContainer struct {
	CreationStarted bool   `json:"creationStarted"`
	ID              string `json:"id,omitempty"`
	Absent          bool   `json:"absent"`
}

func pairContainerName(job, role string) string { return "commonroom-" + role + "-" + job }
func validPairContainerID(id string) bool {
	return len(id) == 64 && repositoryDigestPattern.MatchString(id)
}
func validWorkerPair(v workerJobRecord) bool {
	if v.Execution == "" {
		return v.Pair == nil && v.Version != workerPairLifecycleVersion
	}
	if v.Version != workerPairLifecycleVersion || v.Execution != workerPairExecution || v.Pair == nil {
		return false
	}
	p := v.Pair
	if p.Version != workerPairVersion || !repositoryDigestPattern.MatchString(p.Repository) {
		return false
	}
	for _, c := range []workerPairContainer{p.Engine, p.Tool} {
		if c.ID != "" && (!c.CreationStarted || !validPairContainerID(c.ID)) {
			return false
		}
		if c.Absent && !c.CreationStarted {
			return false
		}
	}
	if p.Engine.ID != "" && p.Engine.ID == p.Tool.ID {
		return false
	}
	if v.Status == "running" && (p.Engine.ID == "" || p.Tool.ID == "" || p.Engine.Absent || p.Tool.Absent) {
		return false
	}
	if v.ShutdownConfirmed && (!pairRoleSettled(p.Engine) || !pairRoleSettled(p.Tool)) {
		return false
	}
	return true
}
func pairRoleSettled(c workerPairContainer) bool { return !c.CreationStarted || c.Absent }

// workerPairSupervisor holds existing job/membership leases. The future launcher
// owns create, inspect, journal and start ordering through its internal adapter.
// No production launch adapter or public entry point is wired.
// Its caller must authenticate the controller and validate image/mount policy.
type workerPairSupervisor struct {
	ctx     context.Context
	cancel  context.CancelFunc
	startup sync.Mutex
	mu      sync.Mutex
	dir     string
	record  workerJobRecord
	leases  *workerJobLeases
	closed  bool
}

// admitWorkerPair is an internal worker seam, not a new run-v2 variant or public
// endpoint. The legacy development barrier remains enforced at all entry points.
func admitWorkerPair(ctx context.Context, dir string, request workerRunRequest, containers []string) (*workerPairSupervisor, error) {
	if ctx == nil || ctx.Err() != nil || request.Version != workerPairLifecycleVersion || !validWorkerJobID(request.JobID) || !repositoryDigestPattern.MatchString(request.Repository) || (request.Provider != "codex" && request.Provider != "claude") {
		return nil, errors.New("invalid pair admission")
	}
	now := time.Now()
	if request.IssuedAt.IsZero() || request.IssuedAt.After(now.Add(5*time.Second)) || !request.Deadline.After(now) || request.Deadline.Sub(request.IssuedAt) > 3*time.Minute || !request.Deadline.After(request.IssuedAt) {
		return nil, errors.New("invalid pair lifetime")
	}
	record, err := newWorkerJobRecord(request)
	if err != nil {
		return nil, err
	}
	record.Version = workerPairLifecycleVersion
	record.Execution = workerPairExecution
	record.Pair = &workerPairRecord{Version: workerPairVersion, Repository: request.Repository}
	record, leases, err := admitWorkerRecord(ctx, dir, record, containers)
	if err != nil {
		return nil, err
	}
	lifetime, cancel := context.WithDeadline(ctx, record.Deadline)
	return &workerPairSupervisor{ctx: lifetime, cancel: cancel, dir: dir, record: record, leases: leases}, nil
}

// Close cancels first, then retains ownership until in-flight adapter work has
// returned. Cancellation alone cannot prove a create/start operation has ended.
// Every concurrent Close waits too; none can report released ownership early.
func (s *workerPairSupervisor) Close() {
	s.mu.Lock()
	s.closed = true
	s.cancel()
	s.mu.Unlock()
	s.startup.Lock()
	defer s.startup.Unlock()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.leases.Close()
}

// pairUpdate is under the owning lease, and refuses any post-cancellation create
// or bind. Intent is durable before Docker create, immutable ID before start.
func (s *workerPairSupervisor) pairUpdate(ctx context.Context, role, id string, begin bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if ctx == nil || (role != "engine" && role != "tool") {
		return errors.New("pair supervisor unavailable")
	}
	if err := s.available(ctx); err != nil {
		return err
	}
	record := s.record
	pair := *record.Pair
	current := &pair.Engine
	if role == "tool" {
		current = &pair.Tool
	}
	if begin {
		if current.CreationStarted || id != "" {
			return errors.New("pair role creation already attempted")
		}
		current.CreationStarted = true
	} else {
		if !current.CreationStarted || current.ID != "" || current.Absent || !validPairContainerID(id) {
			return errors.New("pair role identity unavailable")
		}
		other := pair.Tool
		if role == "tool" {
			other = pair.Engine
		}
		if id == other.ID {
			return errors.New("pair roles cannot share a container")
		}
		current.ID = id
	}
	record.Pair = &pair
	record.UpdatedAt = time.Now().UTC()
	if !validWorkerPair(record) {
		return errors.New("invalid pair transition")
	}
	if err := writeWorkerRecord(s.dir, record, false); err != nil {
		s.closed = true
		s.cancel()
		s.leases.Close()
		return err
	}
	s.record = record
	return nil
}
func (s *workerPairSupervisor) BeginCreate(ctx context.Context, role string) error {
	return s.pairUpdate(ctx, role, "", true)
}
func (s *workerPairSupervisor) BindCreated(ctx context.Context, role, id string) error {
	return s.pairUpdate(ctx, role, id, false)
}

// workerPairLaunchAdapter is trusted worker policy, never model-supplied launch
// arguments. Create MUST create without starting, Verify MUST inspect reviewed
// image, mounts, credentials and aggregate resource policy, and Start MUST use
// the immutable ID and respect cancellation. Bootstrap must return a job-bound
// receipt before engine startup. No production adapter is wired yet.
// A source fork CAN violate this contract, BUT loses container isolation proof.
type workerPairLaunchAdapter interface {
	Create(context.Context, workerPairIdentity) (string, error)
	Verify(context.Context, workerPairIdentity, string) error
	Inspect(context.Context, string) ([]byte, error)
	Start(context.Context, string) error
	Bootstrap(context.Context, string) (workspacebootstrap.Receipt, error)
}

type workerPairIdentity struct{ Job, Role, Repository string }

// Start owns the lifecycle ordering behind one interface. Both identities are
// durable before either process starts. Errors retain the journal for explicit
// fenced recovery; startup never retries uncertain creation or starts.
// This internal seam neither authenticates callers nor enables development.
func (s *workerPairSupervisor) Start(ctx context.Context, adapter workerPairLaunchAdapter) error {
	s.startup.Lock()
	defer s.startup.Unlock()
	if ctx == nil || adapter == nil {
		return errors.New("pair startup unavailable")
	}
	s.mu.Lock()
	original := s.record
	s.mu.Unlock()
	run, cancel := context.WithDeadline(ctx, original.Deadline)
	stop := context.AfterFunc(s.ctx, cancel)
	defer stop()
	defer cancel()
	for _, role := range []string{"tool", "engine"} {
		if err := s.BeginCreate(run, role); err != nil {
			return err
		}
		identity := workerPairIdentity{Job: original.JobID, Role: role, Repository: original.Pair.Repository}
		id, err := adapter.Create(run, identity)
		if err != nil || !validPairContainerID(id) {
			return errors.New("pair creation unconfirmed")
		}
		if err = adapter.Verify(run, identity, id); err != nil {
			return errors.New("pair policy unverified")
		}
		if err = s.BindCreated(run, role, id); err != nil {
			return err
		}
	}
	inspections := map[string][]byte{}
	for _, role := range []string{"tool", "engine"} {
		s.mu.Lock()
		err := s.available(run)
		id := s.record.Pair.Tool.ID
		if role == "engine" {
			id = s.record.Pair.Engine.ID
		}
		s.mu.Unlock()
		if err != nil {
			return err
		}
		data, err := adapter.Inspect(run, id)
		var observed struct {
			ID string `json:"Id"`
		}
		if err != nil || len(data) > 65536 || json.Unmarshal(data, &observed) != nil || observed.ID != id {
			return errors.New("pair resources unverified")
		}
		inspections[role] = data
	}
	if err := toolbroker.VerifyPairResources(inspections["engine"], inspections["tool"]); err != nil {
		return errors.New("pair resources unverified")
	}
	for _, role := range []string{"tool", "engine"} {
		s.mu.Lock()
		err := s.available(run)
		id := s.record.Pair.Tool.ID
		if role == "engine" {
			id = s.record.Pair.Engine.ID
		}
		s.mu.Unlock()
		if err != nil {
			return err
		}
		if err = adapter.Start(run, id); err != nil {
			return errors.New("pair start unconfirmed")
		}
		if role == "tool" {
			s.mu.Lock()
			err = s.available(run)
			s.mu.Unlock()
			if err != nil {
				return err
			}
			receipt, err := adapter.Bootstrap(run, id)
			if err != nil || !workspacebootstrap.ValidReceipt(receipt, original.JobID, original.Pair.Repository) {
				return errors.New("pair workspace bootstrap unverified")
			}
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.available(run); err != nil {
		return err
	}
	record := s.record
	record.Status = "running"
	record.DiagnosticCode = "running"
	record.UpdatedAt = time.Now().UTC()
	if err := writeWorkerRecord(s.dir, record, false); err != nil {
		s.closed = true
		s.cancel()
		s.leases.Close()
		return err
	}
	s.record = record
	return nil
}

// available is called with mu held so Close and state publication serialize.
func (s *workerPairSupervisor) available(ctx context.Context) error {
	if ctx.Err() != nil || s.ctx.Err() != nil || s.closed || !s.record.Deadline.After(time.Now()) {
		return errors.New("pair supervisor unavailable")
	}
	if _, err := os.Lstat(filepath.Join(s.dir, s.record.JobID+".cancel")); !os.IsNotExist(err) {
		return errors.New("pair cancelled or cancellation state unavailable")
	}
	return nil
}

// pairCapture never interprets shell input, inherits credentials or returns raw
// diagnostics. The production runner revalidates the approved rootless daemon.
func pairCapture(ctx context.Context, runner toolbroker.DockerRunner, args []string) (string, error) {
	if ctx == nil || ctx.Err() != nil || runner == nil {
		return "", errors.New("pair runtime unavailable")
	}
	child, cancel := context.WithCancel(ctx)
	defer cancel()
	out := &workerPairOutput{cancel: cancel}
	err := runner.Run(child, args, nil, out)
	if err != nil || child.Err() != nil || out.overflow {
		return "", errors.New("pair runtime observation failed")
	}
	return strings.TrimSpace(out.data.String()), nil
}

type workerPairOutput struct {
	data     strings.Builder
	cancel   context.CancelFunc
	overflow bool
}

func (w *workerPairOutput) Write(p []byte) (int, error) {
	if w.data.Len()+len(p) > 8192 {
		w.overflow = true
		w.cancel()
		return 0, io.ErrShortBuffer
	}
	return w.data.Write(p)
}

func pairInventory(ctx context.Context, runner toolbroker.DockerRunner, filter string) (string, error) {
	raw, err := pairCapture(ctx, runner, []string{"ps", "-a", "--no-trunc", "--filter", filter, "--format", "{{.ID}}"})
	if err != nil {
		return "", err
	}
	if raw != "" && !validPairContainerID(raw) {
		return "", errors.New("ambiguous pair inventory")
	}
	return raw, nil
}
func inspectPairIdentity(ctx context.Context, runner toolbroker.DockerRunner, v workerJobRecord, role, id string) error {
	format := `{{.Id}}|{{.Name}}|{{index .Config.Labels "commonroom.managed"}}|{{index .Config.Labels "commonroom.job"}}|{{index .Config.Labels "commonroom.execution"}}|{{index .Config.Labels "commonroom.repository"}}`
	got, err := pairCapture(ctx, runner, []string{"inspect", "--type", "container", "--format", format, id})
	expected := id + "|/" + pairContainerName(v.JobID, role) + "|true|" + v.JobID + "|separated-" + role + "-v1|" + v.Pair.Repository
	if err != nil || got != expected {
		return errors.New("pair container binding unverified")
	}
	return nil
}

// stopWorkerPair attempts both roles even if one fails. Docker errors/EOF are
// never absence proof. Missing IDs from a create/journal crash are discovered by
// fixed name, inspected for exact role/job/digest, then removed by immutable ID.
// Callers must also wait for the supervisor lease; this function alone cannot
// prove a still-running creator won't launch later. It grants no restart authority.
func stopWorkerPair(ctx context.Context, v workerJobRecord, runner toolbroker.DockerRunner) (workerJobRecord, error) {
	if v.Pair == nil || !validWorkerPair(v) || !validWorkerJobID(v.JobID) {
		return v, errors.New("invalid pair shutdown binding")
	}
	pair := *v.Pair
	v.Pair = &pair
	var failed bool
	for _, role := range []string{"engine", "tool"} {
		current := &pair.Engine
		if role == "tool" {
			current = &pair.Tool
		}
		// Check fixed-name collisions even for a role whose creation never began.
		byName, err := pairInventory(ctx, runner, "name=^/"+pairContainerName(v.JobID, role)+"$")
		if err != nil {
			failed = true
			continue
		}
		id := current.ID
		if byName != "" {
			if id != "" && id != byName {
				failed = true
				continue
			}
			if !current.CreationStarted || inspectPairIdentity(ctx, runner, v, role, byName) != nil {
				failed = true
				continue
			}
			id = byName
			current.ID = id
		}
		if id != "" {
			if err := toolbroker.StopContainer(ctx, runner, id, func(ctx context.Context, id string) error {
				return inspectPairIdentity(ctx, runner, v, role, id)
			}); err != nil {
				failed = true
				continue
			}
		}
		// Name reuse must not be converted into absence of this role.
		if remaining, err := pairInventory(ctx, runner, "name=^/"+pairContainerName(v.JobID, role)+"$"); err != nil || remaining != "" {
			failed = true
			continue
		}
		if current.CreationStarted {
			current.Absent = true
		}
	}
	// No extra managed container with this job label may survive unnoticed.
	if remaining, err := pairCapture(ctx, runner, []string{"ps", "-a", "--no-trunc", "--filter", "label=commonroom.job=" + v.JobID, "--format", "{{.ID}}"}); err != nil || remaining != "" {
		failed = true
	}
	if failed {
		return v, errors.New("pair shutdown unverified")
	}
	return v, nil
}

// Pair acknowledgements bind execution kind and repository as well as scope.
// Version-2 replies cannot settle a version-3 cancellation request.
type workerFenceAck struct {
	Version               int    `json:"version"`
	JobID                 string `json:"jobId"`
	Scope                 string `json:"scope"`
	Stopped               bool   `json:"stopped"`
	CancellationRequested bool   `json:"cancellationRequested"`
	Execution             string `json:"execution,omitempty"`
	Repository            string `json:"repository,omitempty"`
}

func workerFenceAcknowledgement(request workerFenceRequest, record workerJobRecord) workerFenceAck {
	ack := workerFenceAck{Version: request.Version, JobID: record.JobID, Scope: record.Scope, Stopped: record.ShutdownConfirmed, CancellationRequested: true, Execution: record.Execution}
	if record.Pair != nil {
		ack.Repository = record.Pair.Repository
	}
	return ack
}
func validWorkerFenceAck(request workerFenceRequest, expected workerJobRecord, ack workerFenceAck) bool {
	if ack.Version != request.Version || ack.JobID != expected.JobID || ack.Scope != expected.Scope || !ack.Stopped {
		return false
	}
	if request.Version == 3 {
		return ack.CancellationRequested && ack.Execution == workerPairExecution && ack.Repository == request.Repository
	}
	return ack.Execution == "" && ack.Repository == ""
}

// Version 3 cannot be interpreted by old binaries as a version-2 single job.
// Reuse common scalar validation only after checking the actual pair evidence.
func validWorkerRecord(v workerJobRecord) bool {
	if v.Version == workerLifecycleVersion {
		return validV2WorkerRecord(v)
	}
	if v.Version != workerPairLifecycleVersion || v.Pair == nil || !validWorkerPair(v) {
		return false
	}
	base := v
	base.Version = workerLifecycleVersion
	base.Pair = nil
	base.Execution = ""
	return validV2WorkerRecord(base)
}
