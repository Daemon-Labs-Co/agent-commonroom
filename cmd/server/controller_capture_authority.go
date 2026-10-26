package main

import (
	"agentroom/internal/room"
	"agentroom/internal/workspacecapture"
	"agentroom/internal/workspacemount"
	"agentroom/internal/workspacesnapshot"
	"context"
	"encoding/json"
	"errors"
	"slices"
	"time"
)

const controllerCaptureAuthorizationKind = "workspace.capture.authorized"

// controllerCaptureEvidence MUST independently authenticate creator ownership,
// exact reviewed running helper/source/object volumes, actual frozen mount proof,
// writer exclusion and physically enforced storage accounting. It must not reenter
// server locks or this capture stream. No permissive/default implementation exists.
type controllerCaptureEvidence interface {
	ValidateCapture(context.Context, room.State, workerJobRecord, workspacecapture.Request, workspacemount.Proof) error
}
type controllerCaptureAuthorization struct {
	Version      int                      `json:"version"`
	ActorID      string                   `json:"actorId"`
	Worker       string                   `json:"worker"`
	RecordDigest string                   `json:"recordDigest"`
	Request      workspacecapture.Request `json:"request"`
	Proof        workspacemount.Proof     `json:"proof"`
}
type controllerCaptureAuthority struct {
	server        *server
	lifetime      context.Context
	actor, worker string
	record        workerJobRecord
	request       workspacecapture.Request
	evidence      controllerCaptureEvidence
}

func copyCaptureRequest(q workspacecapture.Request) workspacecapture.Request {
	q.Members = slices.Clone(q.Members)
	return q
}

// Historical audit validation intentionally has no current deadline/connection or
// physical callback: replay must not turn a past observation into a fresh grant.
func captureDefinitionBound(st room.State, q workspacecapture.Request, worker string) bool {
	b := q.Binding.Execution
	j, ok := st.ManagedJobs[b.JobID]
	e, exists := st.TaskExecutions[b.ExecutionID]
	return ok && exists && validManagedJob(j) && j.Version == workerWorkspaceLifecycleVersion && worker == j.Worker && b.CompanyID == st.Company.ID && b.CompanyID == j.CompanyID && b.RoomID == j.RoomID && b.TaskID == j.TaskID && b.ExecutionID == j.TaskDispatchID && b.RepositoryDigest == j.RepositoryDigest && b.DefinitionDigest == j.TaskDefinitionDigest && e.ID == b.ExecutionID && e.CompanyID == b.CompanyID && e.RoomID == b.RoomID && e.TaskID == b.TaskID && e.Mode == "work" && e.ParticipantID == j.ParticipantID && e.InitiatorID == j.InitiatorID && e.DefinitionDigest == b.DefinitionDigest && workspacecapture.ValidateContractRequest(e.Definition, q) == nil
}
func validControllerCaptureAuthorization(st room.State, v controllerCaptureAuthorization) bool {
	j := st.ManagedJobs[v.Request.Binding.Execution.JobID]
	return v.Version == 1 && probeHex(v.RecordDigest, 32) && captureDefinitionBound(st, v.Request, v.Worker) && workspacemount.ValidProof(v.Proof, v.Request.Challenge) && artifactRecordedActor(st, v.ActorID, j.RoomID) && (v.ActorID == ownerID || v.ActorID == j.InitiatorID)
}
func projectControllerCaptureAuthorization(st *room.State, raw []byte) error {
	var v controllerCaptureAuthorization
	if json.Unmarshal(raw, &v) != nil || !validControllerCaptureAuthorization(*st, v) {
		return errArtifactCustody
	}
	return nil
}

// Private session context supplies the authenticated human and worker. Peer/model
// frames may not choose either identity. The record/request and lifetime are pinned;
// canceled authority cannot be revived with a new operation context. This source
// observation is not lifecycle admission, custody sealing or public export authority.
func newControllerCaptureAuthority(ctx context.Context, s *server, worker string, r workerJobRecord, q workspacecapture.Request, e controllerCaptureEvidence) (*controllerCaptureAuthority, error) {
	if ctx == nil || ctx.Err() != nil || s == nil || !workerAliasPattern.MatchString(worker) || !retainedObservationRecord(r) || nilWorkerObservationDependency(e) {
		return nil, errArtifactCustody
	}
	actor, _ := ctx.Value(actorContextKey{}).(string)
	if !membershipPartPattern.MatchString(actor) || q.Binding.Execution.JobID != r.JobID || q.Binding.Execution.RepositoryDigest != r.Pair.Repository || q.Binding.VolumeName != r.Pair.Workspace.Identity.Name || q.Binding.VolumeNonce != r.Pair.Workspace.Identity.Nonce || q.Binding.VolumeCreatedAt != r.Pair.Workspace.Identity.CreatedAt {
		return nil, errArtifactCustody
	}
	if _, err := workspacecapture.RequestDigest(q); err != nil {
		return nil, errArtifactCustody
	}
	return &controllerCaptureAuthority{s, ctx, actor, worker, detachedProbeWorkerRecord(r), copyCaptureRequest(q), e}, nil
}
func liveControllerCaptureBinding(st room.State, v controllerCaptureAuthorization, r workerJobRecord) bool {
	j := st.ManagedJobs[v.Request.Binding.Execution.JobID]
	return !st.Incident && validControllerCaptureAuthorization(st, v) && artifactCaptureJobStatus(j) && time.Now().Before(j.Deadline) && sameWorkerBinding(j, workerJobView{workerJobRecord: r}) && managedTaskCurrent(st, j) && coordinationJobAllowed(st, j) && delegatedJobAllowed(st, j)
}
func (g *controllerCaptureAuthority) current(ctx context.Context, p workspacemount.Proof) (room.State, controllerCaptureAuthorization, error) {
	st, err := g.server.load(ctx)
	v := controllerCaptureAuthorization{1, g.actor, g.worker, probeWorkerRecordDigest(g.record), copyCaptureRequest(g.request), p}
	j := st.ManagedJobs[g.request.Binding.Execution.JobID]
	if err != nil || ctx.Err() != nil || g.lifetime.Err() != nil || g.server.integrityFault || st.Incident || !artifactActor(st, ctx, j.RoomID) || g.server.agentsBlockedLocked(st, j.RoomID) || !liveControllerCaptureBinding(st, v, g.record) {
		return st, v, errArtifactCustody
	}
	return st, v, nil
}
func (g *controllerCaptureAuthority) AuthorizeFrozenSource(ctx context.Context, b workspacesnapshot.Binding, p workspacemount.Proof) error {
	if g == nil || g.server == nil || g.lifetime == nil || nilWorkerObservationDependency(g.evidence) || ctx == nil || ctx.Err() != nil || g.lifetime.Err() != nil || b != g.request.Binding || !workspacemount.ValidProof(p, g.request.Challenge) {
		return errArtifactCustody
	}
	child, cancel := context.WithCancel(context.WithValue(ctx, actorContextKey{}, g.actor))
	defer cancel()
	stop := context.AfterFunc(g.lifetime, cancel)
	defer stop()
	child, storageCancel := g.server.storageContext(child)
	defer storageCancel()
	s := g.server
	s.mu.Lock()
	defer s.mu.Unlock()
	st, _, err := g.current(child, p)
	if err != nil {
		return err
	}
	if g.evidence.ValidateCapture(child, st, detachedProbeWorkerRecord(g.record), copyCaptureRequest(g.request), p) != nil || child.Err() != nil {
		return errArtifactCustody
	}
	fresh, v, err := g.current(child, p)
	if err != nil {
		return err
	}
	if fresh.EventCount != st.EventCount {
		return errAuditConflict
	}
	err = s.append(context.WithValue(child, expectedEventCountKey{}, fresh.EventCount), controllerCaptureAuthorizationKind, v)
	if err != nil && !errors.Is(err, errAuditConflict) {
		s.tripIntegrityLocked()
	}
	if err == nil && (child.Err() != nil || g.lifetime.Err() != nil) {
		return errArtifactCustody
	}
	return err
}
