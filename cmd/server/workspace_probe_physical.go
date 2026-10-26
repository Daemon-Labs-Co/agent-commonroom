package main

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"context"
	"strings"
)

// workspaceProbePhysicalEvidence is a worker-local metadata observer scoped to
// the actual creator ownership callback. It reads Docker and private lifecycle
// metadata only. It is not a FrozenLease, remote attestation or model authority.
// Trusted setup must supply an authenticated worker/approved daemon and a reviewed
// immutable helper image. A source fork substituting those loses these claims.
type workspaceProbePhysicalEvidence struct {
	lifetime                 context.Context
	dir, root, worker, image string
	record                   workerJobRecord
	runner                   toolbroker.DockerRunner
}

func newWorkspaceProbePhysicalEvidence(l *workerWorkspaceObservation, worker, image string) (*workspaceProbePhysicalEvidence, error) {
	if l == nil || !workerAliasPattern.MatchString(worker) || !strings.HasPrefix(image, "sha256:") || !probeHex(strings.TrimPrefix(image, "sha256:"), 32) {
		return nil, errControllerProbe
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.active || l.lifetime == nil || l.lifetime.Err() != nil || l.ownership == nil || l.ownership.file == nil || !retainedObservationRecord(l.record) || nilWorkerObservationDependency(l.runner) {
		return nil, errControllerProbe
	}
	return &workspaceProbePhysicalEvidence{lifetime: l.lifetime, dir: l.dir, root: l.dockerRoot, worker: worker, image: image, record: detachedProbeWorkerRecord(l.record), runner: l.runner}, nil
}

// Inspection never executes a probe; its journal refuses every effect. This
// prevents a metadata validator from becoming a create/start/remove capability.
type inspectionOnlyProbeJournal struct{}

func (inspectionOnlyProbeJournal) Begin(context.Context, toolbroker.WorkspaceProbeIntent) error {
	return errControllerProbe
}
func (inspectionOnlyProbeJournal) Bound(context.Context, toolbroker.WorkspaceProbeIntent, string) error {
	return errControllerProbe
}
func (inspectionOnlyProbeJournal) Finish(context.Context, toolbroker.WorkspaceProbeIntent, string, bool) error {
	return errControllerProbe
}

func (e *workspaceProbePhysicalEvidence) inspect(ctx context.Context, st room.State, i toolbroker.WorkspaceProbeIntent, check func(context.Context, *toolbroker.WorkspaceMountProbe) error) error {
	if e == nil || e.lifetime == nil || !retainedObservationRecord(e.record) || ctx == nil || ctx.Err() != nil || e.lifetime.Err() != nil || check == nil || !validWorkerProbeIntent(i) || i.JobID != e.record.JobID || i.Repository != e.record.Pair.Repository || i.Volume != e.record.Pair.Workspace.Identity || i.Image != e.image {
		return errControllerProbe
	}
	job, ok := st.ManagedJobs[i.JobID]
	if !ok || !validManagedJob(job) || job.CompanyID != st.Company.ID || job.Worker != e.worker || !sameWorkerBinding(job, workerJobView{workerJobRecord: e.record}) {
		return errControllerProbe
	}
	child, cancel := context.WithCancel(ctx)
	defer cancel()
	stop := context.AfterFunc(e.lifetime, cancel)
	defer stop()
	if _, err := readRetainedWorkspaceObservation(e.dir, e.record); err != nil {
		return err
	}
	probe, err := toolbroker.NewWorkspaceMountProbe(i, e.runner, inspectionOnlyProbeJournal{}, e.root)
	if err != nil {
		return err
	}
	if err = check(child, probe); err != nil {
		return err
	}
	// Re-read durable fences after physical calls; callback expiry, replacement or
	// cancellation cannot turn a previously valid observation into current evidence.
	if _, err = readRetainedWorkspaceObservation(e.dir, e.record); err != nil {
		return err
	}
	if child.Err() != nil || e.lifetime.Err() != nil {
		return errControllerProbe
	}
	return nil
}
func (e *workspaceProbePhysicalEvidence) zeroReferences(ctx context.Context, probe *toolbroker.WorkspaceMountProbe) error {
	return verifyRetainedWorkspaceReferences(ctx, e.record, e.root, e.runner, probe, "", false)
}
func (e *workspaceProbePhysicalEvidence) ValidateWorkspaceProbe(ctx context.Context, st room.State, r workerJobRecord, i toolbroker.WorkspaceProbeIntent, operation string) error {
	if e == nil || (operation != "authorize" && operation != "observe") || probeWorkerRecordDigest(r) != probeWorkerRecordDigest(e.record) {
		return errControllerProbe
	}
	return e.inspect(ctx, st, i, func(ctx context.Context, probe *toolbroker.WorkspaceMountProbe) error {
		candidate, err := pairInventory(ctx, e.runner, "name=^/commonroom-export-"+i.JobID+"-"+i.Challenge+"$")
		if err != nil {
			return err
		}
		if candidate == "" {
			return e.zeroReferences(ctx, probe)
		}
		reservation, ok := st.WorkspaceProbes[controllerProbeKey(i)]
		if !ok || !validControllerProbeRecord(reservation) || reservation.Intent != i || reservation.Worker != e.worker || reservation.Released {
			return errControllerProbe
		}
		known := recordedProbeID(workerProbeRecord{Version: 1, Intent: i, Events: reservation.Events})
		if known != "" && candidate != known {
			return errControllerProbe
		}
		// A pre-bound candidate is inspected only. It is never adopted into history,
		// launched or removed by name. Bound acknowledgment still needs the journal's
		// exact ID, complete admission checks and durable controller persistence.
		return verifyRetainedWorkspaceReferences(ctx, e.record, e.root, e.runner, probe, candidate, false)
	})
}
func (e *workspaceProbePhysicalEvidence) ValidateProbe(ctx context.Context, st room.State, reservation room.WorkspaceProbeReservation, event workerProbeEvent) error {
	i := reservation.Intent
	return e.inspect(ctx, st, i, func(ctx context.Context, probe *toolbroker.WorkspaceMountProbe) error {
		if reservation.Worker != e.worker || !validControllerProbeRecord(reservation) || len(reservation.Events) == 0 || reservation.Events[len(reservation.Events)-1] != event {
			return errControllerProbe
		}
		records, err := readWorkerProbeRecords(e.dir)
		local, ok := records[probeRecordName(i)]
		if err != nil || !ok || local.Intent != i || len(local.Events) == 0 || local.Events[len(local.Events)-1] != event {
			return errControllerProbe
		}
		switch event.Stage {
		case "begin":
			return e.zeroReferences(ctx, probe)
		case "bound":
			return verifyRetainedWorkspaceReferences(ctx, e.record, e.root, e.runner, probe, event.ID, false)
		case "finish", "recovery-finish":
			if event.Absent {
				if !probeHex(event.ID, 32) {
					return errControllerProbe
				}
				found, err := pairInventory(ctx, e.runner, "id="+event.ID)
				if err != nil || found != "" {
					return errControllerProbe
				}
				return e.zeroReferences(ctx, probe)
			}
			if event.ID == "" {
				return e.zeroReferences(ctx, probe)
			}
			_, err := probe.InspectRecovery(ctx, event.ID)
			return err
		case "recovery":
			_, err := probe.InspectRecovery(ctx, event.ID)
			return err
		default:
			return errControllerProbe
		}
	})
}

var _ controllerWorkspaceProbeEvidence = (*workspaceProbePhysicalEvidence)(nil)
