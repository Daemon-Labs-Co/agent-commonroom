package main

// Immutable-event replay into current application state; no persistence or provider calls.

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

func (s *server) load(ctx context.Context) (room.State, error) {
	ctx, cancel := s.storageContext(ctx)
	defer cancel()
	activities := map[string]string{}
	background := map[string][]room.BackgroundActivity{}
	activityStarts := map[string]time.Time{}
	failures := map[string]string{}
	acknowledged := map[string]string{}
	st := room.State{Company: room.Company{ID: "company", Name: "My company"}, HumanCompanyID: "company", Workflows: []room.Workflow{}, Rooms: []room.Room{{ID: "engineering", Name: "Engineering room"}}, Tasks: []room.Task{}, Evidence: []room.Evidence{}, Connections: []room.Connection{}, Participants: []room.Participant{}, Memories: []room.Memory{}, Messages: []room.Message{}}
	rows, err := s.db.Query(ctx, "SELECT kind,payload FROM events ORDER BY sequence")
	if err != nil {
		return st, err
	}
	defer rows.Close()
	for rows.Next() {
		if err := ctx.Err(); err != nil {
			return st, err
		}
		var kind string
		var data []byte
		if err = rows.Scan(&kind, &data); err != nil {
			return st, err
		}
		st.EventCount++
		switch kind {
		case "workspace.disposal.proposed", "workspace.disposal.claimed", "workspace.disposal.revoked", "workspace.disposal.succeeded", "workspace.disposal.unconfirmed":
			if err := projectWorkspaceDisposal(&st, kind, data); err != nil {
				return st, err
			}
		case "storage.inventory.recorded", "storage.reserved", "storage.released":
			if err := projectManagedStorage(&st, kind, data); err != nil {
				return st, err
			}
		case "worker.tool.intent", "worker.tool.outcome":
			if err := projectToolAudit(&st, kind, data); err != nil {
				return st, err
			}
		case "operation.receipt":
			var receipt room.OperationReceipt
			if json.Unmarshal(data, &receipt) != nil || len(receipt.Key) != 64 || len(receipt.Fingerprint) != 64 || receipt.Status < 200 || receipt.Status >= 300 || !json.Valid(receipt.Response) {
				return st, errors.New("Invalid operation receipt")
			}
			if st.OperationReceipts == nil {
				st.OperationReceipts = map[string]room.OperationReceipt{}
			}
			st.OperationReceipts[receipt.Key] = receipt
		case "tool.call.intent":
			var intent toolbroker.Intent
			// ValidAuditIntent's deadline check is admission-time-relative; replay
			// happens long after that deadline passed, so validate shape as of the
			// intent's own deadline instant rather than wall-clock now (see its doc
			// comment: "outcome persistence may instead use the original intent's
			// admission instant" — the same reasoning applies to replaying history).
			if json.Unmarshal(data, &intent) != nil || !toolbroker.ValidAuditIntent(intent, intent.Scope.Deadline.Add(-time.Nanosecond)) {
				return st, errors.New("invalid tool call intent event")
			}
			if st.ToolCalls == nil {
				st.ToolCalls = map[string]room.ToolCallAudit{}
			}
			st.ToolCalls[intent.CallID] = room.ToolCallAudit{
				CallID: intent.CallID, JobID: intent.Scope.JobID, RoomID: intent.Scope.Room,
				ParticipantID: intent.Scope.Participant, ConnectionID: intent.Scope.Connection,
				Provider: intent.Scope.Provider, Operation: intent.Operation,
				PayloadDigest: intent.PayloadDigest, Status: "pending",
			}
		case "tool.call.outcome":
			var record toolCallOutcomeRecord
			if json.Unmarshal(data, &record) != nil || !toolbroker.ValidAuditOutcome(record.Outcome) {
				return st, errors.New("invalid tool call outcome event")
			}
			existing, ok := st.ToolCalls[record.Intent.CallID]
			if !ok {
				return st, errors.New("tool call outcome has no recorded intent")
			}
			existing.Status, existing.OutputBytes, existing.OutputDigest = record.Outcome.Status, record.Outcome.OutputBytes, record.Outcome.OutputDigest
			st.ToolCalls[record.Intent.CallID] = existing
		case "worker.job.updated":
			var job room.ManagedJob
			if json.Unmarshal(data, &job) != nil || !validManagedJob(job) {
				return st, errors.New("invalid managed job event")
			}
			if st.ManagedJobs == nil {
				st.ManagedJobs = map[string]room.ManagedJob{}
			}
			if prior, exists := st.ManagedJobs[job.JobID]; exists {
				binding := job
				binding.Status, binding.DiagnosticCode, binding.UpdatedAt, binding.ShutdownConfirmed = prior.Status, prior.DiagnosticCode, prior.UpdatedAt, prior.ShutdownConfirmed
				if binding != prior || prior.ShutdownConfirmed && !job.ShutdownConfirmed {
					return st, errors.New("managed job binding changed")
				}
			}
			st.ManagedJobs[job.JobID] = job
		case "actor.help.updated":
			var v struct {
				ID   string `json:"id"`
				Mode string `json:"mode"`
			}
			if err := json.Unmarshal(data, &v); err != nil {
				return st, err
			}
			projectActors(&st)
			for i := range st.Actors {
				if st.Actors[i].ID == v.ID {
					st.Actors[i].HelpDisplay = v.Mode
				}
			}
		case "actor.image.updated":
			var v struct {
				ID    string `json:"id"`
				Image string `json:"image"`
			}
			if err := json.Unmarshal(data, &v); err != nil {
				return st, err
			}
			projectActors(&st)
			for i := range st.Actors {
				if st.Actors[i].ID == v.ID {
					st.Actors[i].Image = v.Image
				}
			}
		case "actor.name.updated":
			var v struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}
			if e := json.Unmarshal(data, &v); e != nil {
				return st, e
			}
			for i, a := range st.Actors {
				if a.ID == v.ID {
					st.Actors[i].Name = v.Name
				}
			}
		case "actor.revised":
			var v room.Actor
			if e := json.Unmarshal(data, &v); e != nil {
				return st, e
			}
			found := false
			for i, a := range st.Actors {
				if a.ID == v.ID {
					v.HelpDisplay = a.HelpDisplay
					st.Actors[i] = v
					found = true
				}
			}
			if !found {
				st.Actors = append(st.Actors, v)
			}
		case "company.organization.updated":
			if err := json.Unmarshal(data, &st.Organization); err != nil {
				return st, err
			}
		case "company.updated":
			json.Unmarshal(data, &st.Company)
		case "human.profile.updated":
			var v struct {
				Name  string  `json:"name"`
				Title *string `json:"title"`
			}
			json.Unmarshal(data, &v)
			st.HumanRevision++
			st.HumanName = v.Name
			if v.Title != nil {
				st.HumanTitle = *v.Title
			}
		case "workflow.run.updated":
			var run room.WorkflowRun
			if json.Unmarshal(data, &run) != nil {
				return st, fmt.Errorf("Invalid workflow run event")
			}
			found := false
			for i, old := range st.WorkflowRuns {
				if old.ID == run.ID {
					st.WorkflowRuns[i] = run
					found = true
					break
				}
			}
			if !found {
				st.WorkflowRuns = append(st.WorkflowRuns, run)
			}
		case "workflow.revised":
			var v room.Workflow
			json.Unmarshal(data, &v)
			found := false
			for i, old := range st.Workflows {
				if old.ID == v.ID {
					st.Workflows[i] = v
					found = true
				}
			}
			if !found {
				st.Workflows = append(st.Workflows, v)
			}

		case "room.created", "room.updated":
			var v room.Room
			json.Unmarshal(data, &v)
			if kind == "room.created" {
				st.Rooms = append(st.Rooms, v)
			} else {
				for i, old := range st.Rooms {
					if old.ID == v.ID {
						st.Rooms[i] = v
					}
				}
			}
		case "task.revised":
			var v room.Task
			json.Unmarshal(data, &v)
			found := false
			for i, t := range st.Tasks {
				if t.ID == v.ID {
					st.Tasks[i] = v
					found = true
				}
			}
			if !found {
				st.Tasks = append(st.Tasks, v)
			}
		case "evidence.created":
			var v room.Evidence
			json.Unmarshal(data, &v)
			st.Evidence = append(st.Evidence, v)

		case "connection.created":
			var v room.Connection
			json.Unmarshal(data, &v)
			st.Connections = append(st.Connections, v)
			st.Actors = append(st.Actors, room.Actor{ID: "bot-" + v.ID, Name: v.Name, Kind: "agent", Role: "Agent", Enabled: true, Revision: 1})

		case "connection.deleted":
			var v struct {
				ID string `json:"id"`
			}
			json.Unmarshal(data, &v)
			connections := st.Connections[:0]
			for _, c := range st.Connections {
				if c.ID != v.ID {
					connections = append(connections, c)
				}
			}
			st.Connections = connections
			participants := st.Participants[:0]
			for _, member := range st.Participants {
				if member.ConnectionID != v.ID {
					participants = append(participants, member)
				}
			}
			st.Participants = participants
		case "connection.updated":
			var v room.Connection
			json.Unmarshal(data, &v)
			for i, old := range st.Connections {
				if old.ID != v.ID {
					continue
				}
				st.Connections[i] = v
				if old.Endpoint != v.Endpoint || old.Protocol != v.Protocol {
					for j, member := range st.Participants {
						if member.ConnectionID == v.ID {
							st.Participants[j].ThreadID = ""
							st.Participants[j].ExecutionMode = ""
							st.Participants[j].Status = ""
							if room.SupportsAgent(v) {
								st.Participants[j].Status = "Idle"
							}
						}
					}
				}
			}
		case "participant.created":
			var v room.Participant
			json.Unmarshal(data, &v)
			if v.RoomID == "" {
				v.RoomID = "engineering"
			}
			st.Participants = append(st.Participants, v)
		case "agent.activity":
			var activity struct {
				ParticipantID string    `json:"participantId"`
				Activity      string    `json:"activity"`
				StartedAt     time.Time `json:"startedAt"`
			}
			json.Unmarshal(data, &activity)
			activities[activity.ParticipantID] = activity.Activity
			activityStarts[activity.ParticipantID] = activity.StartedAt
		case "agent.background":
			var event room.BackgroundActivity
			if json.Unmarshal(data, &event) == nil && backgroundLabels[event.Stage] != "" {
				event.Label = backgroundLabels[event.Stage]
				background[event.ParticipantID] = append(background[event.ParticipantID], event)
				if len(background[event.ParticipantID]) > 64 {
					background[event.ParticipantID] = background[event.ParticipantID][1:]
				}
			}
		case "agent.error":
			var failure struct {
				ParticipantID string `json:"participantId"`
				Error         string `json:"error"`
			}
			json.Unmarshal(data, &failure)
			failures[failure.ParticipantID] = failure.Error
		case "agent.recovered":
			var recovered struct {
				Participant room.Participant `json:"participant"`
			}
			if json.Unmarshal(data, &recovered) == nil {
				for i, p := range st.Participants {
					if p.ID == recovered.Participant.ID && p.RoomID == recovered.Participant.RoomID {
						st.Participants[i] = recovered.Participant
						delete(failures, p.ID)
					}
				}
			}
		case "participant.updated", "participant.session":
			var v room.Participant
			json.Unmarshal(data, &v)
			if v.RoomID == "" {
				v.RoomID = "engineering"
			}
			for i, p := range st.Participants {
				if p.ID == v.ID {
					st.Participants[i] = v
					if kind == "participant.session" && v.Status != "Blocked" {
						delete(failures, v.ID)
					}
					if v.ConnectionID != p.ConnectionID {
						delete(failures, v.ID)
						delete(activities, v.ID)
						delete(activityStarts, v.ID)
					}
				}
			}
		case "room.context_export.updated":
			var v struct {
				RoomID string                   `json:"roomId"`
				Policy room.ContextExportPolicy `json:"policy"`
			}
			json.Unmarshal(data, &v)
			for i := range st.Rooms {
				if st.Rooms[i].ID == v.RoomID {
					st.Rooms[i].ContextExport = &v.Policy
					st.Rooms[i].Revision++
				}
			}
		case "memory.revised":
			var v room.Memory
			json.Unmarshal(data, &v)
			if v.RoomID == "" {
				v.RoomID = "engineering"
			}
			found := false
			for i, m := range st.Memories {
				if m.ID == v.ID {
					st.Memories[i] = v
					found = true
				}
			}
			if !found {
				st.Memories = append(st.Memories, v)
			}
		case "participant.removed":
			var v struct {
				ID string `json:"id"`
			}
			json.Unmarshal(data, &v)
			kept := st.Participants[:0]
			for _, p := range st.Participants {
				if p.ID != v.ID {
					kept = append(kept, p)
				}
			}
			st.Participants = kept
		case "message.attention.acknowledged":
			var ack struct {
				MessageID string `json:"messageId"`
				ActorID   string `json:"actorId"`
			}
			json.Unmarshal(data, &ack)
			acknowledged[ack.MessageID] = ack.ActorID
		case "message.created":
			var v room.Message
			json.Unmarshal(data, &v)
			if v.RoomID == "" {
				v.RoomID = "engineering"
			}
			if v.Provider == "" && v.Author != "You" {
				for _, p := range st.Participants {
					if p.Name == v.Author && p.RoomID == v.RoomID {
						for _, c := range st.Connections {
							if c.ID == p.ConnectionID {
								v.Provider = c.Provider
								if v.Provider == "" && room.SupportsAgent(c) {
									v.Provider = providerForEndpoint(c.Endpoint)
								}
							}
						}
					}
				}
			}
			st.Messages = append(st.Messages, v)
		case "agents.idle_policy.updated":
			if err := projectIdleSettings(&st, data); err != nil {
				return st, err
			}
		case "feature.flag.updated":
			var v struct {
				Name    string `json:"name"`
				Enabled bool   `json:"enabled"`
			}
			if e := json.Unmarshal(data, &v); e != nil {
				return st, e
			}
			if st.FeatureFlags == nil {
				st.FeatureFlags = map[string]bool{}
			}
			st.FeatureFlags[v.Name] = v.Enabled
		case "agents.stop", "agents.resume":
			var v struct {
				Scope  string `json:"scope"`
				RoomID string `json:"roomId"`
			}
			if e := json.Unmarshal(data, &v); e != nil {
				return st, e
			}
			if v.Scope == "application" {
				st.AgentsStopped = kind == "agents.stop"
			} else if v.Scope == "room" {
				kept := []string{}
				for _, rid := range st.AgentStopRooms {
					if rid != v.RoomID {
						kept = append(kept, rid)
					}
				}
				if kind == "agents.stop" {
					kept = append(kept, v.RoomID)
				}
				st.AgentStopRooms = kept
			}
		case "incident.opened":
			st.Incident = true
		case "incident.recovered":
			st.Incident = false
		}
	}
	if err := ctx.Err(); err != nil {
		return st, err
	}
	projectAgentStatusDetails(&st, activities, failures)
	projectAgentElapsed(&st, activityStarts, time.Now().UTC())
	projectBackground(&st, background)
	for i := range st.Rooms {
		st.Rooms[i].CompanyID = st.Company.ID
	}
	for i := range st.Connections {
		st.Connections[i].CompanyID = st.Company.ID
	}
	for i := range st.Participants {
		st.Participants[i].CompanyID = st.Company.ID
	}
	projectActors(&st)
	projectHumanAttention(&st, acknowledged)
	if err := ctx.Err(); err != nil {
		return st, err
	}
	return st, rows.Err()
}
