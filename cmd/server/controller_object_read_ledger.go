package main

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"context"
	"encoding/json"
	"errors"
	"slices"
)

var errControllerObjectRead = errors.New("controller reader authority unavailable")

const controllerObjectReadEventKind = "artifact.reader.acknowledged"

// controllerObjectReadEvidence MUST independently authenticate the approved worker,
// exact private local journal/creator ownership, reviewed image, custody volume,
// physical quota and operation-specific export or cleanup authority. Absence is
// independently observed, never accepted from model/worker metadata alone. It must
// not reenter server/admission locks. No default or authenticated transport exists.
// Source forks can falsify these checks and invalidate mandatory isolation.
type controllerObjectReadEvidence interface {
	ValidateObjectReadTransition(context.Context, room.State, room.ObjectReadReservation, workerProbeEvent) error
}
type controllerObjectReadLedger struct {
	server   *server
	evidence controllerObjectReadEvidence
}
type controllerObjectReadEvent struct {
	Version   int                    `json:"version"`
	ActorID   string                 `json:"actorId"`
	CompanyID string                 `json:"companyId"`
	RoomID    string                 `json:"roomId"`
	Worker    string                 `json:"worker"`
	Intent    workerObjectReadIntent `json:"intent"`
	Event     workerProbeEvent       `json:"event"`
}

func newControllerObjectReadLedger(s *server, e controllerObjectReadEvidence) (*controllerObjectReadLedger, error) {
	if s == nil || nilWorkerObservationDependency(e) {
		return nil, errControllerObjectRead
	}
	return &controllerObjectReadLedger{s, e}, nil
}
func controllerObjectReadKey(i workerObjectReadIntent) string {
	return i.Request.Receipt.Intent.Scope.JobID + "/" + i.Request.Challenge
}
func validControllerObjectReadRecord(r room.ObjectReadReservation) bool {
	if r.Version != 1 || !membershipPartPattern.MatchString(r.CompanyID) || !membershipPartPattern.MatchString(r.RoomID) || !workerAliasPattern.MatchString(r.Worker) || !membershipPartPattern.MatchString(r.ActorID) || len(r.Events) == 0 || len(r.Events) > 5 || r.Released != releasedProbeEvents(r.Events) {
		return false
	}
	for _, e := range r.Events {
		if e.Stage == "confirmed" || e.Stage == "recovered" {
			return false
		}
	}
	return validWorkerObjectReadRecord(workerObjectReadRecord{Version: 1, Intent: r.Intent, Events: r.Events})
}

// Historical identity checks do not depend on current task state or editing flags.
// Cleanup must remain interpretable after task completion and access revocation.
func controllerObjectReadJob(st room.State, r room.ObjectReadReservation) bool {
	scope := r.Intent.Request.Receipt.Intent.Scope
	job, ok := st.ManagedJobs[scope.JobID]
	return ok && validManagedJob(job) && job.Version == workerWorkspaceLifecycleVersion && r.CompanyID == st.Company.ID && r.CompanyID == scope.CompanyID && r.CompanyID == job.CompanyID && r.RoomID == scope.RoomID && r.RoomID == job.RoomID && r.Worker == job.Worker && artifactRecordedSource(st, r.Intent.Request.Receipt.Intent)
}
func controllerObjectReadSlots(st room.State) (int, error) {
	count := 0
	if len(st.ObjectReads) > workerProbeRecordLimit {
		return 0, errControllerObjectRead
	}
	for key, r := range st.ObjectReads {
		if key != controllerObjectReadKey(r.Intent) || !validControllerObjectReadRecord(r) || !controllerObjectReadJob(st, r) {
			return 0, errControllerObjectRead
		}
		if !r.Released {
			count++
		}
	}
	return count, nil
}
func validateControllerObjectReadEvent(st room.State, v controllerObjectReadEvent) (room.ObjectReadReservation, bool, error) {
	zero := room.ObjectReadReservation{}
	candidate := room.ObjectReadReservation{Version: v.Version, ActorID: v.ActorID, CompanyID: v.CompanyID, RoomID: v.RoomID, Worker: v.Worker, Intent: v.Intent}
	if v.Version != 1 || !toolbroker.ValidObjectReadMetadata(v.Intent.Request, v.Intent.Image, v.Intent.Objects) || !controllerObjectReadJob(st, candidate) || !artifactRecordedActor(st, v.ActorID, v.RoomID) {
		return zero, false, errControllerObjectRead
	}
	// New read admission binds the actual sealed receipt and unreleased reservation.
	// Cleanup only binds recorded immutable intent; it grants no read/export rights.
	if v.Event.Stage == "begin" || v.Event.Stage == "bound" {
		receipt := v.Intent.Request.Receipt
		if validateArtifactTransition(st, "artifact.custody.read", artifactCustodyEvent{Version: 1, ActorID: v.ActorID, Intent: receipt.Intent, Receipt: &receipt}) != nil {
			return zero, false, errControllerObjectRead
		}
	}
	key := controllerObjectReadKey(v.Intent)
	old, exists := st.ObjectReads[key]
	if exists {
		if !validControllerObjectReadRecord(old) || old.Intent != v.Intent || old.CompanyID != v.CompanyID || old.RoomID != v.RoomID || old.Worker != v.Worker || (v.ActorID != old.ActorID && v.ActorID != ownerID) {
			return zero, false, errControllerObjectRead
		}
		// Only an exact latest transition may be re-acknowledged under fresh authority.
		// Earlier begin/bound checkpoints cannot reopen a finished or recovered reader.
		if old.Events[len(old.Events)-1] == v.Event {
			return old, true, nil
		}
		if old.Released {
			return zero, false, errControllerObjectRead
		}
	} else {
		if v.Event.Stage != "begin" || len(st.ObjectReads) >= workerProbeRecordLimit || admitControllerProbeQuota(st) != nil {
			return zero, false, errControllerObjectRead
		}
		jobID := v.Intent.Request.Receipt.Intent.Scope.JobID
		for _, r := range st.ObjectReads {
			if r.Intent.Request.Receipt.Intent.Scope.JobID == jobID && !r.Released {
				return zero, false, errControllerObjectRead
			}
		}
		for _, r := range st.WorkspaceCaptures {
			if r.Intent.Request.Binding.Execution.JobID == jobID && !r.Released {
				return zero, false, errControllerObjectRead
			}
		}
		old = candidate
	}
	old.Events = append(slices.Clone(old.Events), v.Event)
	old.Released = releasedProbeEvents(old.Events)
	if !validControllerObjectReadRecord(old) {
		return zero, false, errControllerObjectRead
	}
	return old, false, nil
}
func projectControllerObjectRead(st *room.State, raw []byte) error {
	var v controllerObjectReadEvent
	if json.Unmarshal(raw, &v) != nil {
		return errControllerObjectRead
	}
	r, duplicate, err := validateControllerObjectReadEvent(*st, v)
	if err != nil || duplicate {
		return errControllerObjectRead
	}
	if st.ObjectReads == nil {
		st.ObjectReads = map[string]room.ObjectReadReservation{}
	}
	st.ObjectReads[controllerObjectReadKey(v.Intent)] = r
	return nil
}
func (j *controllerObjectReadLedger) current(ctx context.Context, i workerObjectReadIntent, e workerProbeEvent) (room.State, controllerObjectReadEvent, error) {
	st, err := j.server.load(ctx)
	var v controllerObjectReadEvent
	if err != nil {
		return st, v, err
	}
	actor, _ := ctx.Value(actorContextKey{}).(string)
	scope := i.Request.Receipt.Intent.Scope
	job, ok := st.ManagedJobs[scope.JobID]
	if !ok || st.Incident || j.server.integrityFault || !artifactActor(st, ctx, scope.RoomID) {
		return st, v, errControllerObjectRead
	}
	if (e.Stage == "begin" || e.Stage == "bound") && j.server.agentsBlockedLocked(st, scope.RoomID) {
		return st, v, errControllerObjectRead
	}
	v = controllerObjectReadEvent{1, actor, st.Company.ID, job.RoomID, job.Worker, i, e}
	return st, v, nil
}

// Acknowledge is private authenticated controller infrastructure, never a public
// event sink. Caller identity must come from trusted middleware/transport; missing
// identity never implies owner. This ledger cannot launch helpers or release bytes.
func (j *controllerObjectReadLedger) Acknowledge(ctx context.Context, i workerObjectReadIntent, e workerProbeEvent) error {
	if j == nil || j.server == nil || nilWorkerObservationDependency(j.evidence) || ctx == nil || ctx.Err() != nil || !toolbroker.ValidObjectReadMetadata(i.Request, i.Image, i.Objects) {
		return errControllerObjectRead
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
	record, _, err := validateControllerObjectReadEvent(st, v)
	if err != nil {
		return err
	}
	record.Events = slices.Clone(record.Events)
	if j.evidence.ValidateObjectReadTransition(ctx, st, record, e) != nil || ctx.Err() != nil {
		return errControllerObjectRead
	}
	fresh, again, err := j.current(ctx, i, e)
	if err != nil {
		return err
	}
	if fresh.EventCount != st.EventCount {
		return errAuditConflict
	}
	_, duplicate, err := validateControllerObjectReadEvent(fresh, again)
	if err != nil {
		return err
	}
	if duplicate {
		return nil
	}
	err = s.append(context.WithValue(ctx, expectedEventCountKey{}, fresh.EventCount), controllerObjectReadEventKind, again)
	if err != nil && !errors.Is(err, errAuditConflict) {
		s.tripIntegrityLocked()
	}
	return err
}
