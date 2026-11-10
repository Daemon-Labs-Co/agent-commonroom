package main

// Agent stance markers and objection records. A marker on the first line of an
// agent's reply is a request the server can act on; it is never an approval,
// permission or execution grant. Objections block only the task or proposal they
// name, and only a human decision (or the objecting agent's withdrawal) closes one.

import (
	"agentroom/internal/room"
	"regexp"
	"strings"
)

// agentObjectionsEnabled defaults off: until enabled, markers stay plain text and
// nothing is blocked. Disabling it stops blocking but keeps the messages.
func agentObjectionsEnabled(st room.State) bool {
	return st.FeatureFlags["agentObjections"]
}

var stanceMarkerPattern = regexp.MustCompile(`(?i)^\s*\[(AGREE|OBJECT|ESCALATE|PROPOSE TASK|DISCUSS|WITHDRAW)(?:\s*:\s*([^\]\n]{1,200}))?\]\s*(.*)$`)

// stanceMarker parses the first line only, so markers quoted later in a reply or
// inside reference material are ignored.
func stanceMarker(text string) (stance, target, detail string) {
	first := strings.SplitN(strings.TrimSpace(text), "\n", 2)[0]
	m := stanceMarkerPattern.FindStringSubmatch(first)
	if m == nil {
		return "", "", ""
	}
	stance = map[string]string{"AGREE": "agree", "OBJECT": "object", "ESCALATE": "escalate", "PROPOSE TASK": "propose", "DISCUSS": "discuss", "WITHDRAW": "withdraw"}[strings.ToUpper(m[1])]
	return stance, strings.TrimSpace(m[2]), strings.TrimSpace(m[3])
}

type objectionResolution struct {
	MessageID string `json:"messageId"`
	RoomID    string `json:"roomId"`
	ActorID   string `json:"actorId"`
	Decision  string `json:"decision"`
	Note      string `json:"note"`
}

// objectionTarget resolves a named target to exactly one task or pending proposal
// in the room. An ambiguous or unknown name yields an untargeted objection, which
// is recorded and visible but blocks nothing.
func objectionTarget(st room.State, roomID, name string) (kind, id, title string) {
	if name == "" {
		return "", "", ""
	}
	matches := 0
	for _, t := range st.Tasks {
		if t.RoomID == roomID && strings.EqualFold(strings.TrimSpace(t.Title), name) {
			kind, id, title = "task", t.ID, t.Title
			matches++
		}
	}
	for _, p := range st.TaskProposals {
		if p.RoomID == roomID && p.Status == "pending" && strings.EqualFold(strings.TrimSpace(p.Title), name) {
			kind, id, title = "proposal", p.ID, p.Title
			matches++
		}
	}
	if matches != 1 {
		return "", "", name
	}
	return kind, id, title
}

// projectStances derives stance labels and objection records from agent messages
// in order. Stored messages are never rewritten.
func projectStances(st *room.State, resolved map[string]objectionResolution) {
	st.Objections = []room.Objection{}
	for i := range st.Messages {
		st.Messages[i].Stance = ""
	}
	if !agentObjectionsEnabled(*st) {
		return
	}
	for i := range st.Messages {
		m := &st.Messages[i]
		if m.ActorType != "agent" {
			continue
		}
		stance, target, detail := stanceMarker(m.Text)
		m.Stance = stance
		switch stance {
		case "object":
			kind, id, title := objectionTarget(*st, m.RoomID, target)
			o := room.Objection{MessageID: m.ID, RoomID: m.RoomID, ActorID: m.ActorID, Author: m.Author, Reason: detail, TargetKind: kind, TargetID: id, TargetTitle: title, Status: "open"}
			if r, ok := resolved[m.ID]; ok {
				o.Status, o.Decision, o.Note, o.ResolvedBy = "resolved", r.Decision, r.Note, r.ActorID
			}
			st.Objections = append(st.Objections, o)
		case "withdraw":
			for j := range st.Objections {
				o := &st.Objections[j]
				if o.Status == "open" && o.RoomID == m.RoomID && o.ActorID == m.ActorID && (target == "" || strings.EqualFold(o.TargetTitle, target)) {
					o.Status, o.Decision, o.ResolvedBy = "resolved", "withdrawn", m.ActorID
				}
			}
		}
	}
}

// openObjection reports the first open objection against a task or proposal.
func openObjection(st room.State, kind, id string) (room.Objection, bool) {
	if !agentObjectionsEnabled(st) {
		return room.Objection{}, false
	}
	for _, o := range st.Objections {
		if o.Status == "open" && o.TargetKind == kind && o.TargetID == id {
			return o, true
		}
	}
	return room.Objection{}, false
}

func objectionBlockMessage(o room.Objection) string {
	return "An open objection from " + o.Author + " blocks this " + o.TargetKind + "; resolve the objection first"
}

const stanceGuidance = "\nYou may begin the first line of a reply with one marker: [AGREE] when you accept the current proposal; [OBJECT: exact task or proposal title] followed by your reason when you disagree with it (omit the title for a general objection); [WITHDRAW: title] to withdraw your earlier objection; [ESCALATE] followed by the question when the disagreement needs a decision from whoever you report to. A marker is a request recorded by the room, never an approval or permission. Use one only when it applies; ordinary replies need none.\n"
