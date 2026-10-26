package main

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"agentroom/internal/workspacecapture"
	"agentroom/internal/workspacemount"
	"context"
	"strings"
)

// capturePhysicalBindingGuard is mandatory independent custody infrastructure.
// It MUST authenticate exact durable local/controller lifecycle binding, helper
// image/object volume, account charges and actual physical source/object quotas,
// plus the proof's origin on this helper's exclusively owned private channel.
// It returns only the already durably bound immutable ID, never name discovery.
// Calls must not reenter server locks or the capture stream. No default exists.
type capturePhysicalBindingGuard interface {
	BoundCapture(context.Context, room.State, string, workerJobRecord, workspacecapture.Request, string, toolbroker.CaptureObjectVolume, workspacemount.Proof) (string, error)
}

// workspaceCapturePhysicalEvidence observes worker-local metadata under the actual
// creator ownership lifetime. It never launches, reads source/object bytes or
// changes accounting. A dishonest runner/binding guard invalidates isolation.
type workspaceCapturePhysicalEvidence struct {
	lifetime                 context.Context
	dir, root, worker, image string
	record                   workerJobRecord
	request                  workspacecapture.Request
	objects                  toolbroker.CaptureObjectVolume
	runner                   toolbroker.DockerRunner
	binding                  capturePhysicalBindingGuard
	plan                     *toolbroker.CaptureContainerPlan
}

func newWorkspaceCapturePhysicalEvidence(l *workerWorkspaceObservation, worker, image string, request workspacecapture.Request, objects toolbroker.CaptureObjectVolume, binding capturePhysicalBindingGuard) (*workspaceCapturePhysicalEvidence, error) {
	if l == nil || !workerAliasPattern.MatchString(worker) || !strings.HasPrefix(image, "sha256:") || !probeHex(strings.TrimPrefix(image, "sha256:"), 32) || nilWorkerObservationDependency(binding) {
		return nil, errArtifactCustody
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.active || l.lifetime == nil || l.lifetime.Err() != nil || l.ownership == nil || l.ownership.file == nil || !retainedObservationRecord(l.record) || nilWorkerObservationDependency(l.runner) {
		return nil, errArtifactCustody
	}
	b := request.Binding
	v := l.record.Pair.Workspace.Identity
	if b.Execution.JobID != l.record.JobID || b.Execution.RepositoryDigest != l.record.Pair.Repository || b.VolumeName != v.Name || b.VolumeNonce != v.Nonce || b.VolumeCreatedAt != v.CreatedAt {
		return nil, errArtifactCustody
	}
	plan, err := toolbroker.NewCaptureContainerPlan(request, image, l.dockerRoot, objects)
	if err != nil {
		return nil, err
	}
	return &workspaceCapturePhysicalEvidence{l.lifetime, l.dir, l.dockerRoot, worker, image, detachedProbeWorkerRecord(l.record), copyCaptureRequest(request), objects, l.runner, binding, plan}, nil
}
func (e *workspaceCapturePhysicalEvidence) ValidateCapture(ctx context.Context, st room.State, record workerJobRecord, request workspacecapture.Request, proof workspacemount.Proof) error {
	if e == nil || e.lifetime == nil || e.lifetime.Err() != nil || ctx == nil || ctx.Err() != nil || e.plan == nil || nilWorkerObservationDependency(e.binding) || probeWorkerRecordDigest(record) != probeWorkerRecordDigest(e.record) || !workspacemount.ValidProof(proof, e.request.Challenge) {
		return errArtifactCustody
	}
	actual, err := workspacecapture.RequestDigest(request)
	expected, _ := workspacecapture.RequestDigest(e.request)
	if err != nil || actual != expected {
		return errArtifactCustody
	}
	job, ok := st.ManagedJobs[e.record.JobID]
	b := e.request.Binding.Execution
	if !ok || !validManagedJob(job) || job.CompanyID != st.Company.ID || job.CompanyID != b.CompanyID || job.Worker != e.worker || job.RoomID != b.RoomID || job.TaskID != b.TaskID || job.TaskDispatchID != b.ExecutionID || job.TaskDefinitionDigest != b.DefinitionDigest || !sameWorkerBinding(job, workerJobView{workerJobRecord: e.record}) {
		return errArtifactCustody
	}
	child, cancel := context.WithCancel(ctx)
	defer cancel()
	stop := context.AfterFunc(e.lifetime, cancel)
	defer stop()
	if _, err = readRetainedWorkspaceObservation(e.dir, e.record); err != nil {
		return err
	}
	id, err := e.binding.BoundCapture(child, st, e.worker, detachedProbeWorkerRecord(e.record), copyCaptureRequest(e.request), e.image, e.objects, proof)
	if err != nil || !probeHex(id, 32) {
		return errArtifactCustody
	}
	if e.plan.VerifyAdmission(child, e.runner, id, "running") != nil {
		return errArtifactCustody
	}
	if _, err = readRetainedWorkspaceObservation(e.dir, e.record); err != nil {
		return err
	}
	again, err := e.binding.BoundCapture(child, st, e.worker, detachedProbeWorkerRecord(e.record), copyCaptureRequest(e.request), e.image, e.objects, proof)
	if err != nil || again != id || child.Err() != nil || e.lifetime.Err() != nil {
		return errArtifactCustody
	}
	return nil
}

var _ controllerCaptureEvidence = (*workspaceCapturePhysicalEvidence)(nil)
