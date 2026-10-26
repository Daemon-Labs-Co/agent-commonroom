package main

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"context"
	"encoding/json"
	"errors"
	"slices"
	"time"
)

var errControllerCapture = errors.New("controller capture authority unavailable")

const controllerCaptureEventKind = "workspace.capture.acknowledged"

// controllerCaptureLedgerEvidence MUST authenticate the exact approved worker, creator
// ownership, reviewed image and physical volume/container observations for each
// event. Recovery requires fresh cleanup authority; absence must be independently
// verified before release. This cannot trust model/worker claims alone. No
// production transport/default exists. It must not re-enter server/admission locks.
// Source forks CAN falsify evidence and invalidate managed isolation guarantees.
type controllerCaptureLedgerEvidence interface {
	ValidateCaptureTransition(context.Context, room.State, room.WorkspaceCaptureReservation, workerProbeEvent) error
}
type controllerCaptureLedger struct {
	server   *server
	evidence controllerCaptureLedgerEvidence
}
type controllerCaptureEvent struct {
	Version   int                 `json:"version"`
	ActorID   string              `json:"actorId"`
	CompanyID string              `json:"companyId"`
	RoomID    string              `json:"roomId"`
	Worker    string              `json:"worker"`
	Intent    workerCaptureIntent `json:"intent"`
	Event     workerProbeEvent    `json:"event"`
}

func newControllerCaptureLedger(s *server, e controllerCaptureLedgerEvidence) (*controllerCaptureLedger, error) {
	if s == nil || nilWorkerObservationDependency(e) {
		return nil, errControllerCapture
	}
	return &controllerCaptureLedger{s, e}, nil
}
func controllerCaptureKey(i workerCaptureIntent) string {
	return i.Request.Binding.Execution.JobID + "/" + i.Request.Challenge
}
func validControllerCaptureRecord(r room.WorkspaceCaptureReservation) bool {
	if r.Version != 1 || !membershipPartPattern.MatchString(r.CompanyID) || !membershipPartPattern.MatchString(r.RoomID) || !workerAliasPattern.MatchString(r.Worker) || !membershipPartPattern.MatchString(r.ActorID) || len(r.Events) == 0 || len(r.Events) > 5 || r.Released != releasedProbeEvents(r.Events) {
		return false
	}
	for _, e := range r.Events {
		if e.Stage == "confirmed" || e.Stage == "recovered" {
			return false
		}
	}
	return validWorkerCaptureRecord(workerCaptureRecord{Version: 1, Intent: r.Intent, Events: r.Events})
}
func controllerCaptureSlots(st room.State) (int, error) {
	count := 0
	for key, r := range st.WorkspaceCaptures {
		if key != controllerCaptureKey(r.Intent) || !validControllerCaptureRecord(r) {
			return 0, errControllerCapture
		}
		job, ok := st.ManagedJobs[r.Intent.Request.Binding.Execution.JobID]
		if !ok || !captureDefinitionBound(st, r.Intent.Request, r.Worker) || !validManagedJob(job) || job.Version != workerWorkspaceLifecycleVersion || job.CompanyID != r.CompanyID || job.RoomID != r.RoomID || job.Worker != r.Worker || job.RepositoryDigest != r.Intent.Request.Binding.Execution.RepositoryDigest {
			return 0, errControllerCapture
		}
		if r.CompanyID == st.Company.ID && !r.Released {
			count++
		}
	}
	return count, nil
}
func admitControllerCaptureQuota(st room.State) error { return admitControllerProbeQuota(st) }
func controllerCaptureJob(st room.State, v controllerCaptureEvent) (room.ManagedJob, bool) {
	job, ok := st.ManagedJobs[v.Intent.Request.Binding.Execution.JobID]
	return job, captureDefinitionBound(st, v.Intent.Request, v.Worker) && ok && validManagedJob(job) && job.Version == workerWorkspaceLifecycleVersion && job.CompanyID == st.Company.ID && v.CompanyID == job.CompanyID && v.RoomID == job.RoomID && v.Worker == job.Worker && v.Intent.Request.Binding.Execution.RepositoryDigest == job.RepositoryDigest
}
func validateControllerCaptureEvent(st room.State, v controllerCaptureEvent) (room.WorkspaceCaptureReservation, bool, error) {
	var zero room.WorkspaceCaptureReservation
	job, ok := controllerCaptureJob(st, v)
	if !ok || v.Version != 1 || !toolbroker.ValidCaptureMetadata(v.Intent.Request, v.Intent.Image, v.Intent.Objects) || !artifactRecordedActor(st, v.ActorID, v.RoomID) || (v.ActorID != ownerID && v.ActorID != job.InitiatorID) {
		return zero, false, errControllerCapture
	}
	if v.Event.Stage == "begin" || v.Event.Stage == "bound" {
		if job.ShutdownConfirmed || !artifactCaptureJobStatus(job) || !managedTaskCurrent(st, job) || !coordinationJobAllowed(st, job) || !delegatedJobAllowed(st, job) {
			return zero, false, errControllerCapture
		}
	}
	key := controllerCaptureKey(v.Intent)
	old, exists := st.WorkspaceCaptures[key]
	if exists {
		if !validControllerCaptureRecord(old) || !captureIntentEqual(old.Intent, v.Intent) || old.CompanyID != v.CompanyID || old.RoomID != v.RoomID || old.Worker != v.Worker {
			return zero, false, errControllerCapture
		}
		for _, e := range old.Events {
			if e == v.Event {
				return old, true, nil
			}
		}
	} else {
		if v.Event.Stage != "begin" || !artifactCaptureJobStatus(job) || job.ShutdownConfirmed || admitControllerCaptureQuota(st) != nil {
			return zero, false, errControllerCapture
		}
		for _, r := range st.WorkspaceCaptures {
			if r.Intent.Request.Binding.Execution.JobID == v.Intent.Request.Binding.Execution.JobID && !r.Released {
				return zero, false, errControllerCapture
			}
		}
		old = room.WorkspaceCaptureReservation{Version: 1, CompanyID: v.CompanyID, RoomID: v.RoomID, Worker: v.Worker, ActorID: v.ActorID, Intent: v.Intent}
	}
	old.Events = append(append([]workerProbeEvent(nil), old.Events...), v.Event)
	old.Released = releasedProbeEvents(old.Events)
	if !validControllerCaptureRecord(old) {
		return zero, false, errControllerCapture
	}
	return old, false, nil
}
func projectControllerCapture(st *room.State, raw []byte) error {
	var v controllerCaptureEvent
	if json.Unmarshal(raw, &v) != nil {
		return errControllerCapture
	}
	record, duplicate, err := validateControllerCaptureEvent(*st, v)
	if err != nil || duplicate {
		return errControllerCapture
	}
	if st.WorkspaceCaptures == nil {
		st.WorkspaceCaptures = map[string]room.WorkspaceCaptureReservation{}
	}
	st.WorkspaceCaptures[controllerCaptureKey(v.Intent)] = record
	return nil
}
func (j *controllerCaptureLedger) current(ctx context.Context, i workerCaptureIntent, e workerProbeEvent) (room.State, controllerCaptureEvent, error) {
	st, err := j.server.load(ctx)
	var v controllerCaptureEvent
	if err != nil {
		return st, v, err
	}
	actor, _ := ctx.Value(actorContextKey{}).(string)
	job, ok := st.ManagedJobs[i.Request.Binding.Execution.JobID]
	if !ok || st.Incident || j.server.integrityFault || !artifactActor(st, ctx, job.RoomID) || (actor != ownerID && actor != job.InitiatorID) {
		return st, v, errControllerCapture
	}
	// Cleanup may be acknowledged while agent launches are halted. It grants no
	// new execution or bytes and still requires current human/evidence authority.
	if e.Stage == "begin" || e.Stage == "bound" {
		if !time.Now().Before(job.Deadline) || job.ShutdownConfirmed || !artifactCaptureJobStatus(job) || j.server.agentsBlockedLocked(st, job.RoomID) || !managedTaskCurrent(st, job) || !coordinationJobAllowed(st, job) || !delegatedJobAllowed(st, job) {
			return st, v, errControllerCapture
		}
	}
	v = controllerCaptureEvent{1, actor, st.Company.ID, job.RoomID, job.Worker, i, e}
	return st, v, nil
}

// Acknowledge is internal authenticated controller admission/audit, never an HTTP
// model event sink. The original principal MUST be supplied by trusted middleware
// or transport even for cleanup; absence never implies owner. Local worker journal
// and physical evidence remain mandatory. No host execution/fallback is introduced.
func (j *controllerCaptureLedger) Acknowledge(ctx context.Context, i workerCaptureIntent, e workerProbeEvent) error {
	if j == nil || j.server == nil || nilWorkerObservationDependency(j.evidence) || ctx == nil || ctx.Err() != nil || !toolbroker.ValidCaptureMetadata(i.Request, i.Image, i.Objects) {
		return errControllerCapture
	}
	i = copyCaptureIntent(i)
	s := j.server
	ctx, cancel := s.storageContext(ctx)
	defer cancel()
	s.mu.Lock()
	defer s.mu.Unlock()
	st, v, err := j.current(ctx, i, e)
	if err != nil {
		return err
	}
	record, _, err := validateControllerCaptureEvent(st, v)
	if err != nil {
		return err
	}
	if j.evidence.ValidateCaptureTransition(ctx, st, copyControllerCaptureReservation(record), e) != nil || ctx.Err() != nil {
		return errControllerCapture
	}
	// Reload after the trusted callback: neither callback mutation nor concurrent
	// controllers can replace the exact authority/capacity snapshot that was checked.
	fresh, again, err := j.current(ctx, i, e)
	if err != nil {
		return err
	}
	if fresh.EventCount != st.EventCount {
		return errAuditConflict
	}
	_, duplicate, err := validateControllerCaptureEvent(fresh, again)
	if err != nil {
		return err
	}
	if duplicate {
		return nil
	}
	err = s.append(context.WithValue(ctx, expectedEventCountKey{}, fresh.EventCount), controllerCaptureEventKind, again)
	if err != nil && !errors.Is(err, errAuditConflict) {
		s.tripIntegrityLocked()
	}
	return err
}

func controllerHelperSlots(st room.State) (int, error) {
	probes, err := controllerProbeSlots(st)
	if err != nil {
		return 0, err
	}
	captures, err := controllerCaptureSlots(st)
	if err != nil {
		return 0, err
	}
	return probes + captures, nil
}

func copyControllerCaptureReservation(r room.WorkspaceCaptureReservation) room.WorkspaceCaptureReservation {
	r.Intent = copyCaptureIntent(r.Intent)
	r.Events = slices.Clone(r.Events)
	return r
}
