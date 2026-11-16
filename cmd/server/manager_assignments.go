package main

// Manager-assigned work. An agent may hand a piece of work to an agent that
// reports to it in the company organization chart. The assignment starts one
// ordinary managed turn for the report: the report's existing permissions and
// container apply and nothing new is granted. It can never approve, accept or
// publish work. A per-room daily limit, set by the owner, bounds how much work
// agent text can start.

import (
	"agentroom/internal/room"
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	defaultManagerAssignmentLimit = 5
	maxManagerAssignmentLimit     = 50
)

// managerAssignmentsEnabled defaults off; while off [ASSIGN] is plain text.
func managerAssignmentsEnabled(st room.State) bool {
	return st.FeatureFlags["managerAssignedTasks"]
}

type assignmentLimit struct {
	RoomID string `json:"roomId"`
	Limit  int    `json:"limit"`
}

type managerAssignment struct {
	RoomID   string    `json:"roomId"`
	FromID   string    `json:"fromId"`
	ToID     string    `json:"toId"`
	Title    string    `json:"title"`
	Assigned time.Time `json:"assigned"`
}

var assignMarkerPattern = regexp.MustCompile(`(?i)^\s*\[ASSIGN(?:\s*:\s*|\s+)([^\]\n]{1,100})\]\s*(.*)$`)

// assignMarker parses "[ASSIGN name] title" from the first line of a reply.
func assignMarker(text string) (name, title string) {
	parts := strings.SplitN(strings.TrimSpace(text), "\n", 2)
	m := assignMarkerPattern.FindStringSubmatch(parts[0])
	if m == nil {
		return "", ""
	}
	return strings.TrimSpace(m[1]), strings.TrimSpace(m[2])
}

// roomAssignmentLimit is the owner's daily limit for a room; an unset room uses
// the default and zero refuses every assignment.
func roomAssignmentLimit(st room.State, roomID string) int {
	for _, r := range st.Rooms {
		if r.ID == roomID && r.AssignmentLimit != nil {
			return *r.AssignmentLimit
		}
	}
	return defaultManagerAssignmentLimit
}

func assignmentsToday(st room.State, roomID string, now time.Time) int {
	day := now.UTC().Format("2006-01-02")
	for _, r := range st.Rooms {
		if r.ID == roomID && r.AssignmentDay == day {
			return r.AssignmentsToday
		}
	}
	return 0
}

func assignmentInstruction(from, title, detail string) string {
	return fmt.Sprintf("%s, who you report to, assigned you this work: %s\n%s\nDo it within your existing permissions and report the result. If you think it is wrong or unsafe, begin your first line with [OBJECT] and your reason; if it needs a decision you cannot make, begin with [ESCALATE] and the question. You cannot approve, accept or publish work.", from, title, detail)
}

// reactToAssignment starts the report's turn when a committed reply from its
// manager begins with [ASSIGN name]. It returns true when the marker was present,
// whether or not the assignment was allowed.
func (s *server) reactToAssignment(ctx context.Context, p room.Participant, text string) bool {
	name, title := assignMarker(text)
	if name == "" {
		return false
	}
	s.mu.Lock()
	all, err := s.load(context.Background())
	if err != nil || !managerAssignmentsEnabled(all) || all.Incident || s.agentsBlockedLocked(all, p.RoomID) {
		s.mu.Unlock()
		return true
	}
	refuse := func(reason string) bool {
		_ = s.appendBatch(context.Background(), []auditChange{{Kind: "message.created", Payload: discussionNotice(p.RoomID, "Assignment from "+p.Name+" was not started: "+reason+".")}})
		s.mu.Unlock()
		return true
	}
	st := scoped(all, p.RoomID)
	committed := false
	for i := len(st.Messages) - 1; i >= 0; i-- {
		if m := st.Messages[i]; m.ActorType == "agent" && m.Author == p.Name {
			committed = m.Text == text
			break
		}
	}
	if !committed {
		s.mu.Unlock()
		return true
	}
	if title == "" {
		return refuse("it has no title")
	}
	var report room.Participant
	matches := 0
	for _, candidate := range st.Participants {
		if candidate.RoomID == p.RoomID && candidate.ID != p.ID && strings.EqualFold(strings.TrimSpace(candidate.Name), name) {
			report = candidate
			matches++
		}
	}
	if matches != 1 {
		return refuse("no single agent named " + name + " is in this room")
	}
	if !room.OrganizationallyOutranks(all.Organization, participantActor(p), participantActor(report)) {
		return refuse(report.Name + " does not report to " + p.Name)
	}
	limit := roomAssignmentLimit(all, p.RoomID)
	now := time.Now().UTC()
	if assignmentsToday(all, p.RoomID, now) >= limit {
		return refuse(fmt.Sprintf("this room's daily limit of %d manager-assigned tasks is reached", limit))
	}
	detail := ""
	if parts := strings.SplitN(strings.TrimSpace(text), "\n", 2); len(parts) == 2 {
		detail = strings.TrimSpace(parts[1])
	}
	record := managerAssignment{RoomID: p.RoomID, FromID: p.ID, ToID: report.ID, Title: title, Assigned: now}
	if s.appendBatch(context.Background(), []auditChange{
		{Kind: "agent.assignment", Payload: record},
		{Kind: "message.created", Payload: discussionNotice(p.RoomID, p.Name+" assigned "+report.Name+": "+title)},
	}) != nil {
		s.mu.Unlock()
		return true
	}
	s.mu.Unlock()
	if _, stop := s.agentInitiatedTurn(detachedInitiatorContext(ctx), p.RoomID, report.ID, assignmentInstruction(p.Name, title, detail), "Working on an assignment from "+p.Name, nil); stop != "" {
		s.mu.Lock()
		_ = s.appendBatch(context.Background(), []auditChange{{Kind: "message.created", Payload: discussionNotice(p.RoomID, "Assignment to "+report.Name+" did not run: "+stop+".")}})
		s.mu.Unlock()
	}
	return true
}
