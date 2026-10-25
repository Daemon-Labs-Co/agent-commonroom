package main

import (
	"agentroom/internal/objectcustody"
	"agentroom/internal/room"
	"agentroom/internal/taskcontract"
	"context"
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"time"
)

// artifactCustodyGuard is trusted worker infrastructure, never a model/plugin.
// ValidateSource must attest a frozen isolated source, exact repository/revision,
// output membership and current export policy. EstablishQuota must enforce the
// exact durable reservation on the actual private object volume before writes.
// ValidateSeal binds actual stored content/size to that frozen source and checks
// the quota/root identity is still enforced. It cannot trust the model receipt.
// AuthorizeRelease checks current authenticated export policy and source identity,
// including completed work. No production implementation/default is supplied.
// A source fork CAN lie through these callbacks and invalidate the boundary.
type artifactCustodyGuard interface {
	ValidateSource(context.Context, room.State, objectcustody.Intent) error
	EstablishQuota(context.Context, room.StorageReservation) error
	ValidateSeal(context.Context, room.State, objectcustody.Receipt) error
	AuthorizeRelease(context.Context, room.State, objectcustody.Receipt) error
}

type controllerArtifactJournal struct {
	server *server
	guard  artifactCustodyGuard
}

var errArtifactCustody = errors.New("artifact custody authority unavailable")

// Internal only: no HTTP/model constructor or optional security bypass. A missing
// guard refuses; the existing managed file/bundle dispatch barrier is unchanged.
func newControllerArtifactJournal(s *server, guard artifactCustodyGuard) (*controllerArtifactJournal, error) {
	if s == nil || nilArtifactGuard(guard) {
		return nil, errArtifactCustody
	}
	return &controllerArtifactJournal{server: s, guard: guard}, nil
}

type artifactCustodyEvent struct {
	Version     int                     `json:"version"`
	ActorID     string                  `json:"actorId"`
	Intent      objectcustody.Intent    `json:"intent"`
	Reservation room.StorageReservation `json:"reservation"`
	Receipt     *objectcustody.Receipt  `json:"receipt,omitempty"`
}

func artifactActor(st room.State, ctx context.Context, rid string) bool {
	actor, _ := ctx.Value(actorContextKey{}).(string)
	a, ok := member(st, actor)
	return ok && a.Enabled && a.Kind == "human" && canAccess(st, actor, rid)
}
func artifactSourceJob(st room.State, intent objectcustody.Intent) (room.ManagedJob, bool) {
	scope := intent.Scope
	j, ok := st.ManagedJobs[scope.JobID]
	if !objectcustody.ValidIntent(intent) || !ok || !validManagedJob(j) || j.Version != workerWorkspaceLifecycleVersion || j.ShutdownConfirmed || j.Status != "submitted" || !time.Now().Before(j.Deadline) || j.CompanyID != scope.CompanyID || j.CompanyID != st.Company.ID || j.RoomID != scope.RoomID || j.RepositoryDigest != scope.RepositoryDigest || j.TaskID != scope.TaskID || j.TaskDispatchID != scope.ExecutionID || j.TaskDefinitionDigest != scope.DefinitionDigest || (!managedTaskCurrent(st, j) || !coordinationJobAllowed(st, j) || !delegatedJobAllowed(st, j)) {
		return j, false
	}
	if !artifactRecordedSource(st, intent) {
		return j, false
	}
	return j, true
}

// Historical identity checks have no clock, physical callback or current
// connection requirement. Retained completed objects remain interpretable.
func artifactRecordedSource(st room.State, intent objectcustody.Intent) bool {
	scope := intent.Scope
	j, ok := st.ManagedJobs[scope.JobID]
	if !ok || !validManagedJob(j) || j.Version != workerWorkspaceLifecycleVersion || j.CompanyID != scope.CompanyID || j.RoomID != scope.RoomID || j.RepositoryDigest != scope.RepositoryDigest || j.TaskID != scope.TaskID || j.TaskDispatchID != scope.ExecutionID || j.TaskDefinitionDigest != scope.DefinitionDigest {
		return false
	}
	e, ok := st.TaskExecutions[scope.ExecutionID]
	if !ok || e.ID != scope.ExecutionID || e.CompanyID != scope.CompanyID || e.RoomID != scope.RoomID || e.TaskID != scope.TaskID || e.Mode != "work" || e.ParticipantID != j.ParticipantID || e.InitiatorID != j.InitiatorID || e.DefinitionDigest != scope.DefinitionDigest || taskcontract.Digest(e.Definition) != scope.DefinitionDigest {
		return false
	}
	// Only a declared finite file/member can be reserved. Value strings, paths in
	// provider messages and undeclared outputs never become artifact sources.
	for _, c := range e.Definition.DesiredOutputs {
		if c.Name != scope.OutputName {
			continue
		}
		if c.Type == "file" && c.FileName == scope.MemberName && intent.MaxBytes <= c.MaxBytes {
			return true
		}
		if c.Type == "bundle" {
			for _, m := range c.Members {
				if m.Name == scope.MemberName && intent.MaxBytes <= m.MaxBytes {
					return true
				}
			}
		}
	}
	return false
}
func (j *controllerArtifactJournal) current(ctx context.Context, intent objectcustody.Intent, source bool) (room.State, error) {
	s := j.server
	st, err := s.load(ctx)
	if err != nil {
		return st, err
	}
	if st.Incident || !artifactActor(st, ctx, intent.Scope.RoomID) || st.Company.ID != intent.Scope.CompanyID || s.agentsBlockedLocked(st, intent.Scope.RoomID) {
		return st, errArtifactCustody
	}
	if source {
		job, ok := artifactSourceJob(st, intent)
		actor, _ := ctx.Value(actorContextKey{}).(string)
		if !ok || (actor != ownerID && job.InitiatorID != actor) {
			return st, errArtifactCustody
		}
		if err = j.guard.ValidateSource(ctx, st, intent); err != nil || ctx.Err() != nil {
			return st, errArtifactCustody
		}
	}
	return st, nil
}
func (j *controllerArtifactJournal) Reserve(ctx context.Context, intent objectcustody.Intent) error {
	if j == nil || j.server == nil || nilArtifactGuard(j.guard) || ctx == nil || ctx.Err() != nil || !objectcustody.ValidIntent(intent) {
		return errArtifactCustody
	}
	s := j.server
	ctx, cancel := s.storageContext(ctx)
	defer cancel()
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := j.current(ctx, intent, true)
	if err != nil {
		return err
	}
	job, _ := artifactSourceJob(st, intent)
	reservation := room.StorageReservation{Version: 1, ID: intent.ObjectID, CompanyID: job.CompanyID, JobID: job.JobID, RoomID: job.RoomID, Worker: job.Worker, RepositoryDigest: job.RepositoryDigest, Category: "artifact", Bytes: intent.MaxBytes}
	actor, _ := ctx.Value(actorContextKey{}).(string)
	event := artifactCustodyEvent{Version: 1, ActorID: actor, Intent: intent, Reservation: reservation}
	if err = validateArtifactTransition(st, "artifact.custody.reserved", event); err != nil {
		return err
	}
	if err = j.append(ctx, st.EventCount, "artifact.custody.reserved", event); err != nil {
		return err
	}
	// A failure/timeout here retains the maximum charge and reserved identity.
	// Neither a retry of this object nor a shutdown can silently refund it.
	if err = j.guard.EstablishQuota(ctx, reservation); err != nil || ctx.Err() != nil {
		return objectcustody.ErrUnconfirmed
	}
	// Another controller can revoke source authority while physical setup runs.
	st, err = j.current(ctx, intent, true)
	if err != nil {
		return err
	}
	event.Reservation = room.StorageReservation{}
	if err = validateArtifactTransition(st, "artifact.custody.ready", event); err != nil {
		return err
	}
	return j.append(ctx, st.EventCount, "artifact.custody.ready", event)
}
func (j *controllerArtifactJournal) Seal(ctx context.Context, receipt objectcustody.Receipt) error {
	return j.record(ctx, "artifact.custody.sealed", receipt)
}
func (j *controllerArtifactJournal) AuthorizeRead(ctx context.Context, scope objectcustody.Scope, receipt objectcustody.Receipt) error {
	if scope != receipt.Intent.Scope {
		return errArtifactCustody
	}
	return j.record(ctx, "", receipt)
}
func (j *controllerArtifactJournal) AcknowledgeRead(ctx context.Context, receipt objectcustody.Receipt) error {
	return j.record(ctx, "artifact.custody.read", receipt)
}
func (j *controllerArtifactJournal) record(ctx context.Context, kind string, receipt objectcustody.Receipt) error {
	if j == nil || j.server == nil || nilArtifactGuard(j.guard) || ctx == nil || ctx.Err() != nil || !objectcustody.ValidReceipt(receipt) {
		return errArtifactCustody
	}
	s := j.server
	ctx, cancel := s.storageContext(ctx)
	defer cancel()
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := j.current(ctx, receipt.Intent, kind == "artifact.custody.sealed")
	if err != nil {
		return err
	}
	actor, _ := ctx.Value(actorContextKey{}).(string)
	event := artifactCustodyEvent{Version: 1, ActorID: actor, Intent: receipt.Intent, Receipt: &receipt}
	transition := kind
	if transition == "" {
		transition = "artifact.custody.read"
	}
	if err = validateArtifactTransition(st, transition, event); err != nil {
		return err
	}
	if kind == "artifact.custody.sealed" {
		if err = j.guard.ValidateSeal(ctx, st, receipt); err != nil || ctx.Err() != nil {
			return errArtifactCustody
		}
	} else {
		if err = j.guard.AuthorizeRelease(ctx, st, receipt); err != nil || ctx.Err() != nil {
			return errArtifactCustody
		}
	}
	if kind == "" {
		return nil
	} // Metadata/byte check is not a release acknowledgment.
	return j.append(ctx, st.EventCount, kind, event)
}
func validateArtifactTransition(st room.State, kind string, v artifactCustodyEvent) error {
	if v.Version != 1 || !membershipPartPattern.MatchString(v.ActorID) || !objectcustody.ValidIntent(v.Intent) || v.Intent.Scope.CompanyID != st.Company.ID {
		return errArtifactCustody
	}
	if !artifactRecordedActor(st, v.ActorID, v.Intent.Scope.RoomID) || !artifactRecordedSource(st, v.Intent) {
		return errArtifactCustody
	}
	if kind == "artifact.custody.reserved" || kind == "artifact.custody.ready" || kind == "artifact.custody.sealed" {
		job := st.ManagedJobs[v.Intent.Scope.JobID]
		execution := st.TaskExecutions[v.Intent.Scope.ExecutionID]
		actor := v.ActorID
		if job.Status != "submitted" || job.ShutdownConfirmed || (actor != ownerID && actor != job.InitiatorID) {
			return errArtifactCustody
		}
		pending := false
		for _, task := range st.Tasks {
			if task.ID == v.Intent.Scope.TaskID && taskExecutionPendingCurrent(st, task, execution) {
				pending = true
				break
			}
		}
		if !pending {
			return errArtifactCustody
		}
	}
	old, exists := st.ArtifactCustody[v.Intent.ObjectID]
	if kind == "artifact.custody.reserved" {
		for _, prior := range st.ArtifactCustody {
			if prior.Intent.Scope.CompanyID == v.Intent.Scope.CompanyID && prior.Intent.Scope.ExecutionID == v.Intent.Scope.ExecutionID && prior.Intent.Scope.OutputName == v.Intent.Scope.OutputName && prior.Intent.Scope.MemberName == v.Intent.Scope.MemberName {
				return errArtifactCustody
			}
		}
		if exists || v.Receipt != nil {
			return errArtifactCustody
		}
		r := v.Reservation
		if r.ID != v.Intent.ObjectID || r.Category != "artifact" || r.Bytes != v.Intent.MaxBytes || r.CompanyID != v.Intent.Scope.CompanyID || r.JobID != v.Intent.Scope.JobID || r.RoomID != v.Intent.Scope.RoomID || r.RepositoryDigest != v.Intent.Scope.RepositoryDigest || r.Worker != st.ManagedJobs[v.Intent.Scope.JobID].Worker || r.WorkspaceName != "" {
			return errArtifactCustody
		}
		_, err := admitManagedStorage(st, r)
		return err
	}
	if !exists || old.Version != 1 || old.Intent != v.Intent || old.ReservationID != v.Intent.ObjectID || v.Reservation != (room.StorageReservation{}) {
		return errArtifactCustody
	}
	if st.ManagedStorage == nil {
		return errArtifactCustody
	}
	r, ok := st.ManagedStorage.Reservations[old.ReservationID]
	if !ok || r.Bytes != v.Intent.MaxBytes || r.Category != "artifact" || r.JobID != v.Intent.Scope.JobID || r.CompanyID != v.Intent.Scope.CompanyID || r.RoomID != v.Intent.Scope.RoomID || r.RepositoryDigest != v.Intent.Scope.RepositoryDigest || r.Worker != st.ManagedJobs[v.Intent.Scope.JobID].Worker {
		return errArtifactCustody
	}
	if _, released := st.ManagedStorage.Releases[old.ReservationID]; released {
		return errArtifactCustody
	}
	if kind == "artifact.custody.ready" {
		if !old.QuotaEstablished && old.Receipt == nil && v.Receipt == nil {
			return nil
		}
		return errArtifactCustody
	}
	if !old.QuotaEstablished || v.Receipt == nil || !objectcustody.ValidReceipt(*v.Receipt) || v.Receipt.Intent != v.Intent {
		return errArtifactCustody
	}
	switch kind {
	case "artifact.custody.sealed":
		if old.Receipt == nil {
			return nil
		}
	case "artifact.custody.read":
		if old.Receipt != nil && *old.Receipt == *v.Receipt && old.ReadAcknowledgments >= 0 && old.ReadAcknowledgments < math.MaxInt64 {
			return nil
		}
	}
	return errArtifactCustody
}
func projectArtifactCustody(st *room.State, kind string, raw []byte) error {
	var v artifactCustodyEvent
	if json.Unmarshal(raw, &v) != nil {
		return errArtifactCustody
	}
	if err := validateArtifactTransition(*st, kind, v); err != nil {
		return err
	}
	if st.ArtifactCustody == nil {
		st.ArtifactCustody = map[string]room.ArtifactCustodyObject{}
	}
	switch kind {
	case "artifact.custody.reserved":
		// One event owns both charge and intent. There is no partial metadata append.
		data, _ := json.Marshal(v.Reservation)
		if err := projectManagedStorage(st, "storage.reserved", data); err != nil {
			return err
		}
		st.ArtifactCustody[v.Intent.ObjectID] = room.ArtifactCustodyObject{Version: 1, Intent: v.Intent, ReservationID: v.Reservation.ID}
	case "artifact.custody.ready":
		old := st.ArtifactCustody[v.Intent.ObjectID]
		old.QuotaEstablished = true
		st.ArtifactCustody[v.Intent.ObjectID] = old
	case "artifact.custody.sealed":
		old := st.ArtifactCustody[v.Intent.ObjectID]
		r := *v.Receipt
		old.Receipt = &r
		st.ArtifactCustody[v.Intent.ObjectID] = old
	case "artifact.custody.read":
		old := st.ArtifactCustody[v.Intent.ObjectID]
		old.ReadAcknowledgments++
		st.ArtifactCustody[v.Intent.ObjectID] = old
	}
	return nil
}

func nilArtifactGuard(g artifactCustodyGuard) bool {
	if g == nil {
		return true
	}
	v := reflect.ValueOf(g)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	}
	return false
}

// Caller holds the server lock. A genuine persistence failure blocks later
// authority; an ordinary stale-controller conflict does not create an incident.
func (j *controllerArtifactJournal) append(ctx context.Context, revision int64, kind string, event artifactCustodyEvent) error {
	err := j.server.append(context.WithValue(ctx, expectedEventCountKey{}, revision), kind, event)
	if err != nil && !errors.Is(err, errAuditConflict) {
		j.server.tripIntegrityLocked()
	}
	return err
}

var _ objectcustody.Journal = (*controllerArtifactJournal)(nil)

// The owner is intrinsic when its default actor has not yet been projected at
// this event position. Explicit historical disabling still takes precedence.
func artifactRecordedActor(st room.State, actor, rid string) bool {
	a, ok := member(st, actor)
	if !ok {
		return actor == ownerID
	}
	return a.Enabled && a.Kind == "human" && canAccess(st, actor, rid)
}
