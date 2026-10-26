package main

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"context"
	"encoding/json"
	"errors"
)

var errControllerProbe = errors.New("controller probe authority unavailable")

const controllerProbeEventKind = "workspace.probe.acknowledged"

// controllerProbeEvidence MUST authenticate the exact approved worker, creator
// ownership, reviewed image and physical volume/container observations for each
// event. Recovery requires fresh cleanup authority; absence must be independently
// verified before release. This cannot trust model/worker claims alone. No
// production transport/default exists. It must not re-enter server/admission locks.
// Source forks CAN falsify evidence and invalidate managed isolation guarantees.
type controllerProbeEvidence interface {
	ValidateProbe(context.Context, room.State, room.WorkspaceProbeReservation, workerProbeEvent) error
}
type controllerProbeLedger struct {
	server   *server
	evidence controllerProbeEvidence
}
type controllerProbeEvent struct {
	Version   int                             `json:"version"`
	ActorID   string                          `json:"actorId"`
	CompanyID string                          `json:"companyId"`
	RoomID    string                          `json:"roomId"`
	Worker    string                          `json:"worker"`
	Intent    toolbroker.WorkspaceProbeIntent `json:"intent"`
	Event     workerProbeEvent                `json:"event"`
}

func newControllerProbeLedger(s *server, e controllerProbeEvidence) (*controllerProbeLedger, error) {
	if s == nil || nilWorkerObservationDependency(e) {
		return nil, errControllerProbe
	}
	return &controllerProbeLedger{s, e}, nil
}
func controllerProbeKey(i toolbroker.WorkspaceProbeIntent) string { return i.JobID + "/" + i.Challenge }
func releasedProbeEvents(events []workerProbeEvent) bool {
	for _, e := range events {
		if (e.Stage == "finish" || e.Stage == "recovery-finish") && e.Absent {
			return true
		}
	}
	return false
}
func validControllerProbeRecord(r room.WorkspaceProbeReservation) bool {
	if r.Version != 1 || !membershipPartPattern.MatchString(r.CompanyID) || !membershipPartPattern.MatchString(r.RoomID) || !workerAliasPattern.MatchString(r.Worker) || !membershipPartPattern.MatchString(r.ActorID) || len(r.Events) == 0 || len(r.Events) > 5 || r.Released != releasedProbeEvents(r.Events) {
		return false
	}
	for _, e := range r.Events {
		if e.Stage == "confirmed" || e.Stage == "recovered" {
			return false
		}
	}
	return validWorkerProbeRecord(workerProbeRecord{Version: 1, Intent: r.Intent, Events: r.Events})
}
func controllerProbeSlots(st room.State) (int, error) {
	count := 0
	for key, r := range st.WorkspaceProbes {
		if key != controllerProbeKey(r.Intent) || !validControllerProbeRecord(r) {
			return 0, errControllerProbe
		}
		job, ok := st.ManagedJobs[r.Intent.JobID]
		if !ok || !validManagedJob(job) || job.Version != workerWorkspaceLifecycleVersion || job.CompanyID != r.CompanyID || job.RoomID != r.RoomID || job.Worker != r.Worker || job.RepositoryDigest != r.Intent.Repository {
			return 0, errControllerProbe
		}
		if r.CompanyID == st.Company.ID && !r.Released {
			count++
		}
	}
	return count, nil
}
func admitControllerProbeQuota(st room.State) error {
	slots, err := controllerProbeSlots(st)
	if err != nil {
		return err
	}
	slots++
	for key, job := range st.ManagedJobs {
		if key != job.JobID || !validManagedJob(job) {
			return errControllerProbe
		}
		if (job.CompanyID == "" || job.CompanyID == st.Company.ID) && !job.ShutdownConfirmed {
			slots += managedContainerSlots(job)
		}
	}
	if slots > managedContainerLimit {
		return errControllerProbe
	}
	return nil
}
func controllerProbeJob(st room.State, v controllerProbeEvent) (room.ManagedJob, bool) {
	job, ok := st.ManagedJobs[v.Intent.JobID]
	return job, ok && validManagedJob(job) && job.Version == workerWorkspaceLifecycleVersion && job.CompanyID == st.Company.ID && v.CompanyID == job.CompanyID && v.RoomID == job.RoomID && v.Worker == job.Worker && v.Intent.Repository == job.RepositoryDigest
}
func validateControllerProbeEvent(st room.State, v controllerProbeEvent) (room.WorkspaceProbeReservation, bool, error) {
	var zero room.WorkspaceProbeReservation
	job, ok := controllerProbeJob(st, v)
	if !ok || v.Version != 1 || !validWorkerProbeIntent(v.Intent) || !artifactRecordedActor(st, v.ActorID, v.RoomID) || (v.ActorID != ownerID && v.ActorID != job.InitiatorID) {
		return zero, false, errControllerProbe
	}
	if v.Event.Stage == "begin" || v.Event.Stage == "bound" {
		if job.ShutdownConfirmed || job.Status != "shutdown_unconfirmed" || !managedTaskCurrent(st, job) || !coordinationJobAllowed(st, job) || !delegatedJobAllowed(st, job) {
			return zero, false, errControllerProbe
		}
	}
	key := controllerProbeKey(v.Intent)
	old, exists := st.WorkspaceProbes[key]
	if exists {
		if !validControllerProbeRecord(old) || old.Intent != v.Intent || old.CompanyID != v.CompanyID || old.RoomID != v.RoomID || old.Worker != v.Worker {
			return zero, false, errControllerProbe
		}
		for _, e := range old.Events {
			if e == v.Event {
				return old, true, nil
			}
		}
	} else {
		if v.Event.Stage != "begin" || job.Status != "shutdown_unconfirmed" || job.ShutdownConfirmed || admitControllerProbeQuota(st) != nil {
			return zero, false, errControllerProbe
		}
		for _, r := range st.WorkspaceProbes {
			if r.Intent.JobID == v.Intent.JobID && !r.Released {
				return zero, false, errControllerProbe
			}
		}
		old = room.WorkspaceProbeReservation{Version: 1, CompanyID: v.CompanyID, RoomID: v.RoomID, Worker: v.Worker, ActorID: v.ActorID, Intent: v.Intent}
	}
	old.Events = append(append([]workerProbeEvent(nil), old.Events...), v.Event)
	old.Released = releasedProbeEvents(old.Events)
	if !validControllerProbeRecord(old) {
		return zero, false, errControllerProbe
	}
	return old, false, nil
}
func projectControllerProbe(st *room.State, raw []byte) error {
	var v controllerProbeEvent
	if json.Unmarshal(raw, &v) != nil {
		return errControllerProbe
	}
	record, duplicate, err := validateControllerProbeEvent(*st, v)
	if err != nil || duplicate {
		return errControllerProbe
	}
	if st.WorkspaceProbes == nil {
		st.WorkspaceProbes = map[string]room.WorkspaceProbeReservation{}
	}
	st.WorkspaceProbes[controllerProbeKey(v.Intent)] = record
	return nil
}
func (j *controllerProbeLedger) current(ctx context.Context, i toolbroker.WorkspaceProbeIntent, e workerProbeEvent) (room.State, controllerProbeEvent, error) {
	st, err := j.server.load(ctx)
	var v controllerProbeEvent
	if err != nil {
		return st, v, err
	}
	actor, _ := ctx.Value(actorContextKey{}).(string)
	job, ok := st.ManagedJobs[i.JobID]
	if !ok || st.Incident || j.server.integrityFault || !artifactActor(st, ctx, job.RoomID) || (actor != ownerID && actor != job.InitiatorID) {
		return st, v, errControllerProbe
	}
	// Cleanup may be acknowledged while agent launches are halted. It grants no
	// new execution or bytes and still requires current human/evidence authority.
	if e.Stage == "begin" || e.Stage == "bound" {
		if job.ShutdownConfirmed || job.Status != "shutdown_unconfirmed" || j.server.agentsBlockedLocked(st, job.RoomID) || !managedTaskCurrent(st, job) || !coordinationJobAllowed(st, job) || !delegatedJobAllowed(st, job) {
			return st, v, errControllerProbe
		}
	}
	v = controllerProbeEvent{1, actor, st.Company.ID, job.RoomID, job.Worker, i, e}
	return st, v, nil
}

// Acknowledge is internal authenticated controller admission/audit, never an HTTP
// model event sink. The original principal MUST be supplied by trusted middleware
// or transport even for cleanup; absence never implies owner. Local worker journal
// and physical evidence remain mandatory. No host execution/fallback is introduced.
func (j *controllerProbeLedger) Acknowledge(ctx context.Context, i toolbroker.WorkspaceProbeIntent, e workerProbeEvent) error {
	if j == nil || j.server == nil || nilWorkerObservationDependency(j.evidence) || ctx == nil || ctx.Err() != nil || !validWorkerProbeIntent(i) {
		return errControllerProbe
	}
	s := j.server
	ctx, cancel := s.storageContext(ctx)
	defer cancel()
	s.mu.Lock()
	defer s.mu.Unlock()
	st, v, err := j.current(ctx, i, e)
	if err != nil {
		return err
	}
	record, _, err := validateControllerProbeEvent(st, v)
	if err != nil {
		return err
	}
	if j.evidence.ValidateProbe(ctx, st, record, e) != nil || ctx.Err() != nil {
		return errControllerProbe
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
	_, duplicate, err := validateControllerProbeEvent(fresh, again)
	if err != nil {
		return err
	}
	if duplicate {
		return nil
	}
	err = s.append(context.WithValue(ctx, expectedEventCountKey{}, fresh.EventCount), controllerProbeEventKind, again)
	if err != nil && !errors.Is(err, errAuditConflict) {
		s.tripIntegrityLocked()
	}
	return err
}
