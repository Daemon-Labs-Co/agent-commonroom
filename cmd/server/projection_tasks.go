package main

// Event reducers for task definitions, proposals, delegations, acceptance and execution. Each reducer applies one kind of recorded event to
// the projection and returns an error only for an event that must never have been recorded.

import (
	"agentroom/internal/room"
	"encoding/json"
	"errors"
)

// reduceTaskRepositoryLinksUpdated reduces task.repository_links.updated.
func reduceTaskRepositoryLinksUpdated(proj *projection, kind string, data []byte) error {
	st := &proj.st
	if err := projectTaskRepositoryLinks(st, data); err != nil {
		return err
	}
	return nil
}

// reduceTaskAcceptanceReviewed reduces task.acceptance.reviewed.
func reduceTaskAcceptanceReviewed(proj *projection, kind string, data []byte) error {
	st := &proj.st
	if err := projectTaskAcceptanceReview(st, data); err != nil {
		return err
	}
	return nil
}

// reduceTaskExecutionStarted reduces task.execution.started, task.execution.finished.
func reduceTaskExecutionStarted(proj *projection, kind string, data []byte) error {
	st := &proj.st
	if err := projectTaskExecution(st, kind, data); err != nil {
		return err
	}
	return nil
}

// reduceTaskDelegationProposed reduces task.delegation.proposed, task.delegation.approved, task.delegation.rejected.
func reduceTaskDelegationProposed(proj *projection, kind string, data []byte) error {
	st := &proj.st
	if err := projectDelegation(st, kind, data); err != nil {
		return err
	}
	return nil
}

// reduceTaskProposalCaptured reduces task.proposal.captured, task.proposal.approved, task.proposal.rejected.
func reduceTaskProposalCaptured(proj *projection, kind string, data []byte) error {
	st := &proj.st
	if err := projectTaskProposal(st, kind, data); err != nil {
		return err
	}
	return nil
}

// reduceTaskRevised reduces task.revised.
func reduceTaskRevised(proj *projection, kind string, data []byte) error {
	st := &proj.st
	var v room.Task
	json.Unmarshal(data, &v)
	if err := validateTaskDefinitionProjection(*st, v, data); err != nil {
		return err
	}
	if v.DelegationID != "" {
		d, ok := st.Delegations[v.DelegationID]
		if !ok || d.Status != "Approved" || d.TaskID != v.ID || d.RoomID != v.RoomID || d.TaskDigest != delegationTaskDigest(v) || v.Assignee != d.DelegateID {
			return errors.New("invalid delegated task scope")
		}
	}
	if v.ProposalID != "" {
		proposal, ok := st.TaskProposals[v.ProposalID]
		if !ok || proposal.Status != "Approved" || proposal.TaskID != v.ID || proposal.RoomID != v.RoomID {
			return errors.New("invalid task proposal provenance")
		}
	}
	found := false
	for i, t := range st.Tasks {
		if t.ID == v.ID {
			if t.DelegationID != v.DelegationID || t.ProposalID != v.ProposalID {
				return errors.New("task proposal provenance changed")
			}
			st.Tasks[i] = v
			found = true
		}
	}
	if !found {
		st.Tasks = append(st.Tasks, v)
	}
	recordTaskDefinitionRevision(st, v)
	return nil
}

// reduceTaskSuggestionDismissed reduces task.suggestion.dismissed.
func reduceTaskSuggestionDismissed(proj *projection, kind string, data []byte) error {
	var v suggestionDismissed
	json.Unmarshal(data, &v)
	proj.dismissedSuggestions[v.ID] = true
	return nil
}
