package main

import (
	"agentroom/internal/toolbroker"
	"context"
	"encoding/json"
	"path/filepath"
	"time"
)

// Recovery requires separate current authenticated cleanup authority and durable
// recovery audit, never inference from an old export/probe approval. Controller
// transport/implementation remains mandatory and is not supplied by models.
type workerProbeRecoveryGuard interface {
	workerWorkspaceObservationGuard
	AuthorizeRecovery(context.Context, toolbroker.WorkspaceProbeIntent, string) error
	Acknowledge(context.Context, toolbroker.WorkspaceProbeIntent, workerProbeEvent) error
	StopAuthority()
}
type guardedProbeRecovery struct {
	absentRequired bool
	active         bool
	intent         toolbroker.WorkspaceProbeIntent
	id             string
	lease          *workerWorkspaceObservation
	probe          *toolbroker.WorkspaceMountProbe
	guard          workerProbeRecoveryGuard
}

func (g *guardedProbeRecovery) Authorize(ctx context.Context, i toolbroker.WorkspaceProbeIntent) error {
	if !g.active || i != g.intent {
		return errWorkerProbeJournal
	}
	var err error
	if g.absentRequired {
		err = g.lease.Check(ctx)
	} else {
		err = g.lease.checkRecovery(ctx, g.probe, g.id)
	}
	if err != nil {
		return errWorkerProbeJournal
	}
	return g.guard.AuthorizeRecovery(ctx, i, g.id)
}
func (g *guardedProbeRecovery) Acknowledge(ctx context.Context, i toolbroker.WorkspaceProbeIntent, e workerProbeEvent) error {
	if !g.active || i != g.intent || e.ID != g.id {
		return errWorkerProbeJournal
	}
	return g.guard.Acknowledge(ctx, i, e)
}
func (g *guardedProbeRecovery) StopAuthority() { g.guard.StopAuthority() }
func recordedProbeID(r workerProbeRecord) string {
	for _, e := range r.Events {
		if (e.Stage == "bound" || e.Stage == "finish") && e.ID != "" {
			return e.ID
		}
	}
	return ""
}

// recoveryTransition preserves monotonic history. Explicit retries may re-ack an
// existing recovery intent/outcome; they never delete evidence or repeat launch.
func (j *workerProbeJournal) recoveryTransition(ctx context.Context, i toolbroker.WorkspaceProbeIntent, id string, complete bool) error {
	if j == nil || ctx == nil || ctx.Err() != nil || nilWorkerObservationDependency(j.guard) || j.persist == nil || !validWorkerProbeIntent(i) || !probeHex(id, 32) {
		return errWorkerProbeJournal
	}
	lock, err := workerWaitLease(ctx, j.dir, ".admission.lease")
	if err != nil {
		return err
	}
	defer lock.Close()
	records, err := readWorkerProbeRecords(j.dir)
	if err != nil {
		return j.fail()
	}
	r, ok := records[probeRecordName(i)]
	if !ok || r.Intent != i || recordedProbeID(r) != id || workerProbeSettled(r) || workerProbeStopped(j.dir) || !probeRecordMatchesWorker(j.dir, i) || j.guard.Authorize(ctx, i) != nil || ctx.Err() != nil {
		return errWorkerProbeJournal
	}
	stage := r.Events[len(r.Events)-1].Stage
	event := workerProbeEvent{Stage: "recovery", ID: id}
	appendEvent := false
	if complete {
		event = workerProbeEvent{Stage: "recovery-finish", ID: id, Absent: true}
		if stage != "recovery" && stage != "recovery-finish" {
			return errWorkerProbeJournal
		}
		appendEvent = stage == "recovery"
	} else {
		if stage != "bound" && stage != "finish" && stage != "recovery" && stage != "recovery-finish" {
			return errWorkerProbeJournal
		}
		appendEvent = stage == "bound" || stage == "finish"
	}
	persist := func() error {
		if !validWorkerProbeRecord(r) {
			return errWorkerProbeJournal
		}
		raw, _ := json.Marshal(r)
		if j.persist(filepath.Join(j.dir, probeRecordName(i)), raw, syncParentDirectory) != nil {
			return j.fail()
		}
		return nil
	}
	if appendEvent {
		r.Events = append(r.Events, event)
		if err = persist(); err != nil {
			return err
		}
	}
	if j.guard.Acknowledge(ctx, i, event) != nil || ctx.Err() != nil {
		return j.fail()
	}
	if j.guard.Authorize(ctx, i) != nil || ctx.Err() != nil {
		return errWorkerProbeJournal
	}
	if complete {
		r.Events = append(r.Events, workerProbeEvent{Stage: "recovered", ID: id, Absent: true})
		if err = persist(); err != nil {
			return err
		}
	}
	if j.guard.Authorize(ctx, i) != nil || ctx.Err() != nil {
		return errWorkerProbeJournal
	}
	return nil
}

// recoverRetainedWorkspaceProbe reconciles only a durable exact helper identity
// while holding creator ownership. It never adopts unknown create replies,
// starts a helper, clears integrity stops, disposes volumes or returns source
// bytes/proof/FrozenLease. Source forks can falsify trusted guards/runners and
// invalidate these guarantees. No host/uncontained execution fallback exists.
func recoverRetainedWorkspaceProbe(ctx context.Context, dir string, record workerJobRecord, intent toolbroker.WorkspaceProbeIntent, dockerRoot string, runner toolbroker.DockerRunner, guard workerProbeRecoveryGuard) error {
	if ctx == nil || ctx.Err() != nil || nilWorkerObservationDependency(guard) || !validWorkerProbeIntent(intent) {
		return errWorkerProbeJournal
	}
	child, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return withWorkerWorkspaceOwnership(child, dir, record, dockerRoot, runner, guard, func(lease *workerWorkspaceObservation) error {
		if intent.JobID != lease.record.JobID || intent.Repository != lease.record.Pair.Repository || intent.Volume != lease.record.Pair.Workspace.Identity {
			return errWorkerProbeJournal
		}
		records, err := readWorkerProbeRecords(dir)
		if err != nil {
			return err
		}
		r, ok := records[probeRecordName(intent)]
		if !ok || r.Intent != intent || workerProbeSettled(r) {
			return errWorkerProbeJournal
		}
		id := recordedProbeID(r)
		if !probeHex(id, 32) {
			return errWorkerProbeJournal
		}
		g := &guardedProbeRecovery{active: true, intent: intent, id: id, lease: lease, guard: guard}
		defer func() { g.active = false }()
		journal, err := newWorkerProbeJournal(dir, g)
		if err != nil {
			return err
		}
		probe, err := toolbroker.NewWorkspaceMountProbe(intent, runner, journal, dockerRoot)
		if err != nil {
			return err
		}
		g.probe = probe
		if err = lease.checkRecovery(child, probe, id); err != nil {
			return err
		}
		if err = journal.recoveryTransition(child, intent, id, false); err != nil {
			return err
		}
		if err = probe.RemoveRecorded(child, id); err != nil {
			return err
		}
		g.absentRequired = true
		return journal.recoveryTransition(child, intent, id, true)
	})
}
