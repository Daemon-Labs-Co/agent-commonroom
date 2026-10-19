package main

import (
	"encoding/json"
	"strings"
	"unicode"
)

// auditSummary derives a display description without changing the signed event
// payload or exposing message bodies, credentials, instructions, or raw errors.
func auditSummary(kind string, payload []byte) string {
	var p map[string]json.RawMessage
	if json.Unmarshal(payload, &p) != nil {
		p = nil
	}
	label := func(key, fallback string) string {
		var value string
		json.Unmarshal(p[key], &value)
		value = strings.TrimSpace(strings.Map(func(r rune) rune {
			if unicode.IsControl(r) {
				return ' '
			}
			return r
		}, value))
		chars := []rune(value)
		if len(chars) > 120 {
			value = string(chars[:120]) + "…"
		}
		if value == "" {
			return fallback
		}
		return value
	}
	name := label("name", "Agent")
	title := label("title", "Untitled")
	switch kind {
	case "worker.tool.intent":
		return "Agent tool action requested · " + label("operation", "action")
	case "worker.tool.outcome":
		return "Agent tool action · " + label("status", "outcome recorded")
	case "message.attention.acknowledged":
		return "Human acknowledged an agent attention request"
	case "agent.activity":
		return "Agent activity started"
	case "agent.background":
		return "Agent background work · " + label("stage", "update")
	case "workspace.disposal.proposed":
		return "Owner proposed permanent workspace disposal"
	case "workspace.disposal.claimed":
		return "Owner approved permanent workspace disposal"
	case "workspace.disposal.revoked":
		return "Owner revoked workspace disposal"
	case "workspace.disposal.succeeded":
		return "Workspace disposal verified"
	case "workspace.disposal.unconfirmed":
		return "Workspace disposal outcome unconfirmed"
	case "storage.inventory.recorded":
		return "Account storage inventory recorded"
	case "storage.released":
		return "Workspace storage capacity released after verified disposal"
	case "storage.reserved":
		return "Account storage capacity reserved"
	case "worker.job.updated":
		return "Managed worker job · " + label("status", "lifecycle update")
	case "worker.job.recovery_requested":
		return "Owner requested worker job recovery · " + label("action", "review")
	case "agent.recovered":
		return "Owner recovered an agent session · " + label("action", "recovery")
	case "workflow.run.updated":
		return "Workflow execution · " + label("status", "updated")
	case "room.created":
		return "Room created: " + label("name", "Room")
	case "room.updated":
		return "Room settings updated: " + label("name", "Room")
	case "room.owner_access":
		return "Owner viewed a personal room"
	case "connection.created":
		return "Connection added: " + label("name", "Connection")
	case "connection.updated":
		return "Connection settings updated: " + label("name", "Connection")
	case "connection.deleted":
		return "Connection removed"
	case "participant.created":
		return name + " joined a room as " + label("role", "an agent")
	case "participant.updated":
		return name + "’s room profile was updated"
	case "participant.removed":
		return name + " removed from a room"
	case "participant.session":
		return name + "’s session updated · " + label("status", "status recorded")
	case "participant.history_attached":
		return "An existing conversation was attached to an agent"
	case "message.created":
		return label("author", "A participant") + " sent a room message"
	case "memory.revised":
		return "Shared memory updated: " + title + " · " + label("scope", "room") + " scope"
	case "task.revised":
		return "Task updated: " + title + " · " + label("status", "status recorded")
	case "evidence.created":
		return "Evidence added: " + title
	case "workflow.revised":
		return "Workflow updated: " + label("name", "Workflow")
	case "actor.revised":
		return "Member account updated: " + label("name", "Member") + " · " + label("role", "role recorded")
	case "actor.image.updated":
		return "Human updated their profile photo"
	case "actor.name.updated":
		return "Human name updated: " + label("name", "Member")
	case "human.profile.updated":
		return "Human profile updated"
	case "company.organization.updated":
		return "Owner updated company organization chart and reporting relationships"
	case "company.updated":
		return "Company settings updated: " + label("name", "Company")
	case "provider.logout_requested":
		return "Connection sign-out requested"
	case "provider.signed_out":
		return "Connection signed out; saved worker credentials cleared"
	case "provider.login_launched":
		return "Provider sign-in launched"
	case "agent.error":
		return name + " turn failed; check the agent connection and provider diagnostics"
	case "agent.stopped":
		return name + " stopped: " + label("reason", "turn cancelled")
	case "feature.flag.updated":
		return "Human updated a feature flag"
	case "agents.stop":
		return "Human activated emergency agent stop · " + label("scope", "scope recorded")
	case "agents.resume":
		return "Human resumed agent dispatch · " + label("scope", "scope recorded")
	case "incident.opened":
		return "Integrity failure detected; human investigation required"
	case "incident.recovered":
		return "Human recorded an integrity recovery decision"
	default:
		return "Activity recorded: " + strings.ReplaceAll(kind, ".", " ")
	}
}
