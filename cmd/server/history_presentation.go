package main

import (
	"context"
	"encoding/json"
	"strings"
	"unicode"
)

const historyFlagSQL = "SELECT payload FROM events WHERE kind='feature.flag.updated' AND payload->>'name'=$1 ORDER BY sequence DESC LIMIT 1"

// historyPresentationEnabled reads only the latest setting, not the full history.
// Missing means on. Read/decoding failures stop the enriched history response.
func (s *server) historyPresentationEnabled(ctx context.Context) (bool, error) {
	ctx, cancel := s.storageContext(ctx)
	defer cancel()
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	rows, err := s.db.Query(ctx, historyFlagSQL, "historyRefinement")
	if err != nil {
		return false, err
	}
	defer rows.Close()
	enabled := true
	if rows.Next() {
		var raw []byte
		if err = rows.Scan(&raw); err != nil {
			return false, err
		}
		var flag struct {
			Enabled bool `json:"enabled"`
		}
		if err = json.Unmarshal(raw, &flag); err != nil {
			return false, err
		}
		enabled = flag.Enabled
	}
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	return enabled, rows.Err()
}

// auditPresentation contains allowlisted human-readable metadata, never a raw
// event payload, provider stderr, message body, endpoint or credential.
type auditPresentation struct {
	Category    string `json:"category"`
	Tone        string `json:"tone"`
	Routine     bool   `json:"routine"`
	Actor       string `json:"actor,omitempty"`
	ActorID     string `json:"actorId,omitempty"`
	Subject     string `json:"subject,omitempty"`
	RoomID      string `json:"roomId,omitempty"`
	Outcome     string `json:"outcome,omitempty"`
	Explanation string `json:"explanation,omitempty"`
	Next        string `json:"next,omitempty"`
}

func auditSafeLabel(value string) string {
	value = strings.TrimSpace(strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, value))
	chars := []rune(value)
	if len(chars) > 120 {
		return string(chars[:120]) + "…"
	}
	return value
}
func describeAuditEvent(kind string, raw []byte) auditPresentation {
	var p map[string]json.RawMessage
	_ = json.Unmarshal(raw, &p)
	label := func(key string) string {
		var value string
		_ = json.Unmarshal(p[key], &value)
		return auditSafeLabel(value)
	}
	d := auditPresentation{Category: "Other", Tone: "info", ActorID: label("performedBy"), RoomID: label("roomId"), Subject: label("title"), Outcome: label("status")}
	if d.Subject == "" {
		d.Subject = label("name")
	}
	if d.ActorID == "" && (strings.HasPrefix(kind, "message.") || strings.HasPrefix(kind, "provider.") || strings.HasPrefix(kind, "incident.") || strings.HasPrefix(kind, "agents.") || kind == "workflow.run.updated" || kind == "room.owner_access") {
		d.ActorID = label("actorId")
	}
	switch {
	case strings.HasPrefix(kind, "message."):
		d.Category = "Messages"
		d.Actor = label("author")
	case strings.HasPrefix(kind, "agent.") || strings.HasPrefix(kind, "participant."):
		d.Category = "Agents"
		if strings.HasPrefix(kind, "agent.") {
			d.Actor = label("name")
		}
	case strings.HasPrefix(kind, "workflow."):
		d.Category = "Workflows"
	case strings.HasPrefix(kind, "task.") || strings.HasPrefix(kind, "evidence."):
		d.Category = "Tasks & evidence"
	case strings.HasPrefix(kind, "memory."):
		d.Category = "Memory"
	case strings.HasPrefix(kind, "connection.") || strings.HasPrefix(kind, "provider."):
		d.Category = "Connections"
	case strings.HasPrefix(kind, "incident.") || strings.HasPrefix(kind, "agents."):
		d.Category = "Security"
	case strings.HasPrefix(kind, "room."):
		d.Category = "Rooms"
	case strings.HasPrefix(kind, "feature."):
		d.Category = "Settings"
		d.Subject = label("name")
	case strings.HasPrefix(kind, "worker.job."):
		d.Category = "Agents"
	case strings.HasPrefix(kind, "actor.") || strings.HasPrefix(kind, "company.") || strings.HasPrefix(kind, "human."):
		d.Category = "People & company"
	}
	d.Routine = kind == "participant.session" || kind == "agent.activity" || kind == "agent.background"
	if kind == "agent.background" {
		switch label("stage") {
		case "failed", "turn_failed", "timed_out":
			d.Outcome = "failed"
		case "unreachable", "shutdown_unconfirmed", "stop_requested", "cancelled":
			d.Outcome = "stopped"
		case "finished", "completed":
			d.Outcome = "completed"
		}
		if d.Outcome != "" {
			d.Routine = false
		}
	}
	if kind == "workflow.run.updated" {
		var workflow struct {
			Name string `json:"name"`
		}
		_ = json.Unmarshal(p["workflow"], &workflow)
		d.Subject = auditSafeLabel(workflow.Name)
		d.Routine = d.Outcome == "running"
	}
	switch strings.ToLower(d.Outcome) {
	case "blocked", "failed", "interrupted":
		d.Tone = "error"
	case "awaiting_human", "changes requested", "stopped":
		d.Tone = "warning"
	case "completed", "accepted":
		d.Tone = "success"
	}
	if kind == "agent.error" {
		d.Tone = "error"
		d.Routine = false
		d.Outcome = "Failed"
		var failure string
		_ = json.Unmarshal(p["error"], &failure)
		d.Explanation, d.Next = safeAgentFailure(failure)
	}
	if kind == "incident.opened" {
		d.Tone = "error"
		d.Explanation = "Human investigation is required before agent execution resumes."
	}
	if kind == "agents.stop" || kind == "agent.stopped" {
		d.Tone = "warning"
	}
	if kind == "message.attention.acknowledged" {
		d.Tone = "success"
		d.Outcome = "Acknowledged"
	}
	if kind == "feature.flag.updated" {
		var enabled *bool
		_ = json.Unmarshal(p["enabled"], &enabled)
		if enabled != nil {
			d.Outcome = "Off"
			if *enabled {
				d.Outcome = "On"
			}
		}
	}
	if d.Tone == "error" || d.Tone == "warning" {
		d.Routine = false
	}
	return d
}
