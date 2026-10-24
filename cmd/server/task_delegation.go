package main

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"sort"
	"time"
)

type taskJobContextKey struct{}
type taskJobBinding struct{ TaskID, DispatchID, DelegationID string }

func managedTaskCurrent(st room.State, j room.ManagedJob) bool {
	if j.TaskID == "" {
		return j.TaskDispatchID == "" && j.DelegationID == ""
	}
	for _, t := range st.Tasks {
		if t.ID == j.TaskID && t.RoomID == j.RoomID && t.DispatchID == j.TaskDispatchID && t.DispatchStatus == "pending" && t.Assignee == j.ParticipantID && t.DelegationID == j.DelegationID {
			return true
		}
	}
	return false
}
func delegationTaskDigest(t room.Task) string {
	raw, _ := json.Marshal([4]string{t.ID, t.RoomID, t.Title, t.Criteria})
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
func delegationPeople(st room.State, d room.TaskDelegation) bool {
	source, ok := member(st, ownerID)
	if !ok || !source.Enabled || !canAccess(st, ownerID, d.RoomID) {
		return false
	}
	for _, binding := range []struct {
		pid, cid   string
		pr, cr, ar int
	}{{d.SourceParticipantID, d.SourceConnectionID, d.SourceParticipantRevision, d.SourceConnectionRevision, d.SourceActorRevision}, {d.DelegateID, d.DelegateConnectionID, d.DelegateRevision, d.DelegateConnectionRevision, d.DelegateActorRevision}} {
		agent, ok := member(st, "bot-"+binding.cid)
		if !ok || !agent.Enabled || agent.Kind != "agent" || agent.Revision != binding.ar {
			return false
		}
		found := false
		for _, p := range st.Participants {
			if p.ID == binding.pid && p.RoomID == d.RoomID && p.ConnectionID == binding.cid && p.Revision == binding.pr && (p.Status == "Idle" || p.Status == "Working") {
				for _, c := range st.Connections {
					if c.ID == binding.cid && c.Revision == binding.cr && room.SupportsAgent(c) {
						if _, _, err := contextDestination(st, p); err == nil {
							found = true
						}
					}
				}
			}
		}
		if !found {
			return false
		}
	}
	return true
}
func delegationCurrent(st room.State, d room.TaskDelegation) bool {
	j, ok := st.ManagedJobs[d.ParentJobID]
	if !ok || !st.FeatureFlags["taskDelegation"] || d.Epoch != st.DelegationEpoch || d.CompanyID != st.Company.ID || !d.ExpiresAt.After(time.Now()) || !j.ShutdownConfirmed || j.InitiatorID != ownerID || j.RepositoryDigest != d.RepositoryDigest || j.TaskID != d.TaskID || j.TaskDispatchID != d.DispatchID || !coordinationJobAllowed(st, j) || !delegationPeople(st, d) {
		return false
	}
	for _, t := range st.Tasks {
		if t.ID == d.TaskID && t.RoomID == d.RoomID && delegationTaskDigest(t) == d.TaskDigest && t.Status != "Accepted" {
			return true
		}
	}
	return false
}
func delegatedJobAllowed(st room.State, j room.ManagedJob) bool {
	if j.DelegationID == "" {
		return true
	}
	d, ok := st.Delegations[j.DelegationID]
	return ok && d.Status == "Approved" && delegationCurrent(st, d) && j.TaskID == d.TaskID && j.ParticipantID == d.DelegateID && j.ConnectionID == d.DelegateConnectionID && j.RepositoryDigest == d.RepositoryDigest && !j.Deadline.After(d.ExpiresAt)
}
func delegationContextAllowed(ctx context.Context, st room.State, p room.Participant) bool {
	b, _ := ctx.Value(taskJobContextKey{}).(taskJobBinding)
	if b.DelegationID == "" {
		return true
	}
	d, ok := st.Delegations[b.DelegationID]
	return ok && d.Status == "Approved" && d.DelegateID == p.ID && d.RoomID == p.RoomID && delegationCurrent(st, d)
}
func (s *server) delegationRepositoryFor(rid string) (string, error) {
	if s.delegationRepository != nil {
		return s.delegationRepository(rid)
	}
	return controllerDevelopmentRepository(rid, os.Getenv("COMMONROOM_DEVELOPMENT_WORKSPACES"), os.Getenv("COMMONROOM_DEVELOPMENT_REPOSITORIES"))
}

func (a *controllerToolAuditor) RecordDelegation(ctx context.Context, scope toolbroker.Scope, request toolbroker.DelegationRequest) error {
	if ctx == nil || scope != a.scope || !toolbroker.ValidDelegation(request) {
		return toolbroker.ErrAudit
	}
	s := a.server
	s.mu.Lock()
	defer s.mu.Unlock()
	st, _, p, c, err := a.check(ctx, true)
	if err != nil {
		return err
	}
	job := st.ManagedJobs[scope.JobID]
	if !st.FeatureFlags["taskDelegation"] || job.InitiatorID != ownerID || job.TaskID != request.TaskID || job.DelegationID != "" || !pairedLifecycleVersion(job.Version) || job.RepositoryDigest != scope.RepositoryDigest {
		return toolbroker.ErrAudit
	}
	var task room.Task
	for _, t := range st.Tasks {
		if t.ID == job.TaskID && t.RoomID == scope.Room {
			task = t
		}
	}
	if task.ID == "" || task.Revision != request.TaskRevision || task.Status != "Working" || request.DelegateID == p.ID {
		return toolbroker.ErrAudit
	}
	digest := sha256.Sum256([]byte(job.JobID + "/" + request.RequestID))
	key := hex.EncodeToString(digest[:])
	if prior, exists := st.Delegations[key]; exists {
		if prior.TaskID == request.TaskID && prior.TaskRevision == request.TaskRevision && prior.ParentOperations == request.ParentOperations && prior.DelegateID == request.DelegateID && prior.Epoch == st.DelegationEpoch {
			return nil
		}
		return toolbroker.ErrAudit
	}
	count := 0
	for _, d := range st.Delegations {
		if d.ParentJobID == job.JobID {
			count++
		}
	}
	if count >= toolbroker.MaxCalls {
		return toolbroker.ErrAudit
	}
	d := room.TaskDelegation{ParentOperations: request.ParentOperations, ID: key, CompanyID: st.Company.ID, RoomID: scope.Room, ParentJobID: job.JobID, SourceParticipantID: p.ID, SourceConnectionID: c.ID, SourceParticipantRevision: p.Revision, SourceConnectionRevision: c.Revision, TaskID: task.ID, TaskRevision: task.Revision, TaskDigest: delegationTaskDigest(task), DispatchID: job.TaskDispatchID, DelegateID: request.DelegateID, RepositoryDigest: scope.RepositoryDigest, ExpiresAt: job.Deadline, Epoch: st.DelegationEpoch, Status: "Proposed", Revision: 1}
	for _, target := range st.Participants {
		if target.ID == request.DelegateID && target.RoomID == scope.Room {
			d.DelegateConnectionID = target.ConnectionID
			d.DelegateRevision = target.Revision
			for _, connection := range st.Connections {
				if connection.ID == target.ConnectionID {
					d.DelegateConnectionRevision = connection.Revision
				}
			}
		}
	}
	sourceActor, _ := member(st, "bot-"+d.SourceConnectionID)
	delegateActor, _ := member(st, "bot-"+d.DelegateConnectionID)
	d.SourceActorRevision = sourceActor.Revision
	d.DelegateActorRevision = delegateActor.Revision
	if !delegationPeople(st, d) {
		return toolbroker.ErrAudit
	}
	if err = s.append(context.WithValue(context.WithValue(ctx, actorContextKey{}, ownerID), expectedEventCountKey{}, st.EventCount), "task.delegation.proposed", d); err != nil {
		if !errors.Is(err, errAuditConflict) {
			s.tripIntegrityLocked()
		}
		return toolbroker.ErrAudit
	}
	return nil
}

type delegationDecision struct {
	ID       string `json:"id"`
	Revision int    `json:"revision"`
}

func projectDelegation(st *room.State, kind string, raw []byte) error {
	projectActors(st)
	var actor struct {
		PerformedBy string `json:"performedBy"`
	}
	if json.Unmarshal(raw, &actor) != nil || actor.PerformedBy != ownerID || !st.FeatureFlags["taskDelegation"] {
		return errors.New("invalid delegation authority")
	}
	if kind == "task.delegation.proposed" {
		var d room.TaskDelegation
		if json.Unmarshal(raw, &d) != nil {
			return errors.New("invalid delegation")
		}
		j, ok := st.ManagedJobs[d.ParentJobID]
		var task room.Task
		for _, t := range st.Tasks {
			if t.ID == d.TaskID && t.RoomID == d.RoomID {
				task = t
			}
		}
		if !ok || d.ParentOperations == 0 || d.ParentOperations > 15 || !validWorkerJobID(j.JobID) || !repositoryDigestPattern.MatchString(d.ID) || j.InitiatorID != ownerID || j.Status != "submitted" || !pairedLifecycleVersion(j.Version) || j.DelegationID != "" || d.SourceParticipantID != j.ParticipantID || d.SourceConnectionID != j.ConnectionID || d.CompanyID != st.Company.ID || d.RoomID != j.RoomID || d.RepositoryDigest != j.RepositoryDigest || d.TaskID != j.TaskID || d.DispatchID != j.TaskDispatchID || !managedTaskCurrent(*st, j) || d.TaskRevision != task.Revision || d.TaskDigest != delegationTaskDigest(task) || d.DelegateID == d.SourceParticipantID || d.Revision != 1 || d.Status != "Proposed" || d.ReviewedBy != "" || !d.ExpiresAt.Equal(j.Deadline) || d.Epoch != st.DelegationEpoch || !delegationPeople(*st, d) {
			return errors.New("invalid delegation scope")
		}
		if st.Delegations == nil {
			st.Delegations = map[string]room.TaskDelegation{}
		}
		if _, exists := st.Delegations[d.ID]; exists {
			return errors.New("duplicate delegation")
		}
		st.Delegations[d.ID] = d
		return nil
	}
	var review delegationDecision
	if json.Unmarshal(raw, &review) != nil {
		return errors.New("invalid delegation decision")
	}
	d, ok := st.Delegations[review.ID]
	if !ok || d.Status != "Proposed" || d.Revision != review.Revision || d.Epoch != st.DelegationEpoch {
		return errors.New("stale delegation decision")
	}
	if kind == "task.delegation.approved" {
		job := st.ManagedJobs[d.ParentJobID]
		if !job.ShutdownConfirmed || !coordinationJobAllowed(*st, job) || !delegationPeople(*st, d) {
			return errors.New("delegation parent not reconciled")
		}
		found := false
		for i, t := range st.Tasks {
			if t.ID == d.TaskID && t.RoomID == d.RoomID && t.DispatchID == d.DispatchID && t.Revision == d.TaskRevision+1 && t.DispatchStatus != "pending" && (t.DispatchStatus == "completed" || t.DispatchStatus == "failed") && t.Assignee == d.SourceParticipantID && t.DelegationID == "" && t.Status != "Accepted" && delegationTaskDigest(t) == d.TaskDigest {
				t.Revision++
				t.Assignee = d.DelegateID
				t.Status = "Assigned"
				t.DelegationID = d.ID
				st.Tasks[i] = t
				found = true
			}
		}
		if !found {
			return errors.New("delegation task changed")
		}
		d.Status = "Approved"
	} else {
		d.Status = "Rejected"
	}
	d.Revision++
	d.ReviewedBy = ownerID
	st.Delegations[d.ID] = d
	return nil
}

func (s *server) taskDelegations(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	if requestActor(r) != ownerID {
		http.Error(w, "Owner permission required", 403)
		return
	}
	if r.Method == "POST" && !s.validCSRF(r) {
		http.Error(w, "Invalid request token", 403)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.load(r.Context())
	if err != nil {
		http.Error(w, "State unavailable", 503)
		return
	}
	rid := requestedRoom(r)
	if !hasRoom(st, rid) || !canAccess(st, ownerID, rid) {
		http.Error(w, "Room unavailable", 404)
		return
	}
	if r.Method == "GET" {
		list := []room.TaskDelegation{}
		for _, d := range st.Delegations {
			if d.RoomID == rid && d.CompanyID == st.Company.ID {
				list = append(list, d)
			}
		}
		sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
		reply(w, map[string]any{"enabled": st.FeatureFlags["taskDelegation"], "delegations": list})
		return
	}
	var req struct {
		Action   string `json:"action"`
		ID       string `json:"id"`
		Revision int    `json:"revision"`
	}
	raw, e := io.ReadAll(http.MaxBytesReader(w, r.Body, 2048))
	fields, ferr := fenceObjectFields(raw, "action", "id", "revision")
	if e != nil || ferr != nil || len(fields) != 3 || json.Unmarshal(raw, &req) != nil || (req.Action != "approve" && req.Action != "reject") || req.Revision < 1 || r.Header.Get("Idempotency-Key") == "" {
		http.Error(w, "Decision, identity, revision and operation key required", 400)
		return
	}
	if !st.FeatureFlags["taskDelegation"] {
		http.Error(w, "Task delegation is disabled", 403)
		return
	}
	ctx, replay := prepareOperation(w, r, st, req)
	if replay {
		return
	}
	d, ok := st.Delegations[req.ID]
	if !ok || d.RoomID != rid || d.CompanyID != st.Company.ID {
		http.Error(w, "Delegation unavailable", 404)
		return
	}
	if d.Status != "Proposed" || d.Revision != req.Revision || d.Epoch != st.DelegationEpoch {
		http.Error(w, "Delegation changed; reload", 409)
		return
	}
	if s.workflowSafety(ctx, st, rid) != nil {
		http.Error(w, "Delegation blocked by safety state", 409)
		return
	}
	if req.Action == "approve" {
		repository, repoErr := s.delegationRepositoryFor(rid)
		if repoErr != nil || repository != d.RepositoryDigest || !delegationCurrent(st, d) {
			http.Error(w, "Parent unfinished, expired or scope changed", 409)
			return
		}
		valid := false
		for _, t := range st.Tasks {
			if t.ID == d.TaskID && t.RoomID == rid && t.Revision == d.TaskRevision+1 && t.DispatchID == d.DispatchID && t.DispatchStatus != "pending" && (t.DispatchStatus == "completed" || t.DispatchStatus == "failed") && t.Assignee == d.SourceParticipantID && t.DelegationID == "" && t.Status != "Accepted" && delegationTaskDigest(t) == d.TaskDigest {
				valid = true
			}
		}
		if !valid {
			http.Error(w, "Task changed or parent work unreconciled", 409)
			return
		}
	}
	response := map[string]any{"committed": true, "id": d.ID, "decision": req.Action}
	if err = s.append(withOperationResponse(ctx, 200, response), "task.delegation."+map[string]string{"approve": "approved", "reject": "rejected"}[req.Action], delegationDecision{d.ID, d.Revision}); err != nil {
		if !errors.Is(err, errAuditConflict) {
			s.tripIntegrityLocked()
		}
		operationError(w, err)
		return
	}
	reply(w, response)
}
