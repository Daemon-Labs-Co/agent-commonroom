package main

import (
	"agentroom/internal/room"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"
	"strings"
)

type proposalDecision struct {
	ProposalID string    `json:"proposalId"`
	Revision   int       `json:"revision"`
	Task       room.Task `json:"task"`
}

func proposalSource(st room.State, rid, mid string) (room.Message, bool) {
	for _, m := range st.Messages {
		if m.ID == mid && m.RoomID == rid {
			return m, true
		}
	}
	return room.Message{}, false
}
func proposalText(title, criteria string) bool {
	return strings.TrimSpace(title) != "" && len(title) <= 200 && len(criteria) <= 4000
}

// Only the authenticated human capture/review endpoint can create these facts.
// A source message is untrusted content, never an approval or execution grant.
func projectTaskProposal(st *room.State, kind string, data []byte) error {
	// Personal rooms derive from actors/connections already replayed, including
	// the implicit owner. Materialize them before validating historical scope;
	// never rely on unrelated profile/help events or later journal records.
	projectActors(st)
	var actor struct {
		PerformedBy string `json:"performedBy"`
	}
	if json.Unmarshal(data, &actor) != nil || actor.PerformedBy != ownerID || !st.FeatureFlags["taskProposals"] {
		return errors.New("invalid task proposal authority")
	}
	if kind == "task.proposal.captured" {
		var p room.TaskProposal
		if json.Unmarshal(data, &p) != nil {
			return errors.New("invalid proposal")
		}
		source, ok := proposalSource(*st, p.RoomID, p.SourceMessageID)
		if !ok || !hasRoom(*st, p.RoomID) || p.CompanyID != st.Company.ID || !membershipPartPattern.MatchString(p.ID) || p.CapturedBy != ownerID || p.ReviewedBy != "" || p.TaskID != "" || p.Status != "Proposed" || p.Revision != 1 || !proposalText(p.Title, p.Criteria) || p.SourceActorID != source.ActorID || p.SourceActorType != source.ActorType || p.SourceAuthor != source.Author {
			return errors.New("invalid proposal source or scope")
		}
		if st.TaskProposals == nil {
			st.TaskProposals = map[string]room.TaskProposal{}
		}
		if _, exists := st.TaskProposals[p.ID]; exists {
			return errors.New("duplicate task proposal")
		}
		st.TaskProposals[p.ID] = p
		return nil
	}
	var decision proposalDecision
	if json.Unmarshal(data, &decision) != nil {
		return errors.New("invalid proposal decision")
	}
	p, ok := st.TaskProposals[decision.ProposalID]
	if !ok || p.Status != "Proposed" || decision.Revision != p.Revision || p.CompanyID != st.Company.ID || !hasRoom(*st, p.RoomID) {
		return errors.New("stale proposal decision")
	}
	if _, ok = proposalSource(*st, p.RoomID, p.SourceMessageID); !ok {
		return errors.New("proposal source unavailable")
	}
	p.ReviewedBy = ownerID
	p.Revision++
	if kind == "task.proposal.approved" {
		task := decision.Task
		if !membershipPartPattern.MatchString(task.ID) || task.ProposalID != p.ID || task.RoomID != p.RoomID || task.Revision != 1 || task.Status != "Draft" || task.Assignee != "" || task.Reviewer != "" || task.DispatchID != "" || task.DispatchStatus != "" || task.DispatchDetail != "" || !proposalText(task.Title, task.Criteria) || strings.TrimSpace(task.Criteria) == "" {
			return errors.New("invalid proposal draft")
		}
		for _, old := range st.Tasks {
			if old.ID == task.ID {
				return errors.New("duplicate draft task")
			}
		}
		p.Status = "Approved"
		p.TaskID = task.ID
		st.Tasks = append(st.Tasks, task)
	} else {
		if decision.Task != (room.Task{}) {
			return errors.New("rejected proposal cannot create task")
		}
		p.Status = "Rejected"
	}
	st.TaskProposals[p.ID] = p
	return nil
}

func (s *server) taskProposals(w http.ResponseWriter, r *http.Request) {
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
	if !hasRoom(st, rid) || !canAccess(st, requestActor(r), rid) {
		http.Error(w, "Room unavailable", 404)
		return
	}
	if r.Method == "GET" {
		type entry struct {
			Proposal        room.TaskProposal `json:"proposal"`
			Source          room.Message      `json:"source"`
			SourceAvailable bool              `json:"sourceAvailable"`
		}
		entries := []entry{}
		for _, p := range st.TaskProposals {
			if p.RoomID == rid && p.CompanyID == st.Company.ID {
				source, ok := proposalSource(st, rid, p.SourceMessageID)
				entries = append(entries, entry{p, source, ok})
			}
		}
		sort.Slice(entries, func(i, j int) bool { return entries[i].Proposal.ID < entries[j].Proposal.ID })
		reply(w, map[string]any{"enabled": st.FeatureFlags["taskProposals"], "proposals": entries})
		return
	}
	if !st.FeatureFlags["taskProposals"] {
		http.Error(w, "Task proposals are disabled", 403)
		return
	}
	var req struct {
		Action          string `json:"action"`
		SourceMessageID string `json:"sourceMessageId"`
		ProposalID      string `json:"proposalId"`
		Revision        int    `json:"revision"`
		Title           string `json:"title"`
		Criteria        string `json:"criteria"`
	}
	raw, e := io.ReadAll(http.MaxBytesReader(w, r.Body, 8192))
	fields, fieldErr := fenceObjectFields(raw, "action", "sourceMessageId", "proposalId", "revision", "title", "criteria")
	if e != nil || fieldErr != nil || json.Unmarshal(raw, &req) != nil || r.Header.Get("Idempotency-Key") == "" {
		http.Error(w, "Valid proposal action and idempotency key required", 400)
		return
	}
	allowed := map[string]bool{"action": true}
	switch req.Action {
	case "capture":
		allowed["sourceMessageId"] = true
		allowed["title"] = true
		allowed["criteria"] = true
	case "approve":
		allowed["proposalId"] = true
		allowed["revision"] = true
		allowed["title"] = true
		allowed["criteria"] = true
	case "reject":
		allowed["proposalId"] = true
		allowed["revision"] = true
	default:
		http.Error(w, "Choose capture, approve or reject", 400)
		return
	}
	if len(fields) != len(allowed) {
		http.Error(w, "Action fields required", 400)
		return
	}
	for name := range fields {
		if !allowed[name] {
			http.Error(w, "Unexpected action field", 400)
			return
		}
	}
	ctx, replay := prepareOperation(w, r, st, req)
	if replay {
		return
	}
	if err = s.workflowSafety(ctx, st, rid); err != nil {
		http.Error(w, "Proposal changes blocked by safety state", 409)
		return
	}
	var response room.TaskProposal
	var kind string
	var payload any
	if req.Action == "capture" {
		source, ok := proposalSource(st, rid, req.SourceMessageID)
		if !ok {
			http.Error(w, "Source message unavailable", 404)
			return
		}
		if !proposalText(req.Title, req.Criteria) {
			http.Error(w, "Title required (maximum 200), criteria maximum 4000 bytes", 400)
			return
		}
		response = room.TaskProposal{ID: id(), CompanyID: st.Company.ID, RoomID: rid, SourceMessageID: source.ID, SourceActorID: source.ActorID, SourceActorType: source.ActorType, SourceAuthor: source.Author, CapturedBy: ownerID, Title: strings.TrimSpace(req.Title), Criteria: strings.TrimSpace(req.Criteria), Status: "Proposed", Revision: 1}
		kind = "task.proposal.captured"
		payload = response
	} else {
		p, ok := st.TaskProposals[req.ProposalID]
		if !ok || p.RoomID != rid || p.CompanyID != st.Company.ID {
			http.Error(w, "Proposal unavailable", 404)
			return
		}
		if p.Status != "Proposed" || p.Revision != req.Revision {
			http.Error(w, "Proposal changed; reload", 409)
			return
		}
		if _, ok = proposalSource(st, rid, p.SourceMessageID); !ok {
			http.Error(w, "Source message unavailable", 409)
			return
		}
		decision := proposalDecision{ProposalID: p.ID, Revision: p.Revision}
		response = p
		response.Revision++
		response.ReviewedBy = ownerID
		if req.Action == "approve" {
			if !proposalText(req.Title, req.Criteria) || strings.TrimSpace(req.Criteria) == "" {
				http.Error(w, "Review a title and acceptance criteria before approval", 400)
				return
			}
			decision.Task = room.Task{ID: id(), ProposalID: p.ID, RoomID: rid, Title: strings.TrimSpace(req.Title), Criteria: strings.TrimSpace(req.Criteria), Status: "Draft", Revision: 1}
			kind = "task.proposal.approved"
			response.Status = "Approved"
			response.TaskID = decision.Task.ID
		} else {
			kind = "task.proposal.rejected"
			response.Status = "Rejected"
		}
		payload = decision
	}
	if err = s.append(withOperationResponse(ctx, 200, response), kind, payload); err != nil {
		if !errors.Is(err, errAuditConflict) {
			s.tripIntegrityLocked()
		}
		operationError(w, err)
		return
	}
	reply(w, response)
}
