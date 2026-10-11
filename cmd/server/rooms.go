package main

import (
	"agentroom/internal/room"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"strings"
)

func requestedRoom(r *http.Request) string {
	v := r.URL.Query().Get("room")
	if v == "" {
		return "engineering"
	}
	return v
}
func hasRoom(st room.State, id string) bool {
	for _, v := range st.Rooms {
		if v.ID == id {
			return true
		}
	}
	return false
}
func hasMember(st room.State, rid, pid string) bool {
	for _, v := range st.Participants {
		if v.ID == pid && v.RoomID == rid {
			return true
		}
	}
	return false
}
func scoped(st room.State, rid string) room.State {
	ws := []room.Workflow{}
	for _, w := range st.Workflows {
		if w.RoomID == rid {
			ws = append(ws, w)
		}
	}
	st.Workflows = ws
	ps := []room.Participant{}
	for _, p := range st.Participants {
		if p.RoomID == rid {
			ps = append(ps, p)
		}
	}
	st.Participants = ps
	ms := []room.Message{}
	for _, m := range st.Messages {
		if m.RoomID == rid {
			ms = append(ms, m)
		}
	}
	st.Messages = ms
	mem := []room.Memory{}
	for _, m := range st.Memories {
		if m.RoomID == rid || m.Scope == "Project" || m.Scope == "Application" {
			mem = append(mem, m)
		}
	}
	st.Memories = mem
	ts := []room.Task{}
	for _, t := range st.Tasks {
		if t.RoomID == rid {
			ts = append(ts, t)
		}
	}
	st.Tasks = ts
	es := []room.Evidence{}
	for _, e := range st.Evidence {
		if e.RoomID == rid {
			es = append(es, e)
		}
	}
	st.Evidence = es
	return st
}
func profileGuidance(p room.Participant) string {
	n := 1
	if p.Challenge != nil {
		n = *p.Challenge
	}
	modes := []string{"Supportive: help progress; raise clear errors and risks.", "Constructive: question unclear assumptions and suggest alternatives.", "Critical: examine tradeoffs and request evidence.", "Adversarial: stress-test proposals and seek failure cases.", "Red team: try to disprove claims and find exploitable weaknesses within authorised scope."}
	if n < 0 || n > 4 {
		n = 1
	}
	return fmt.Sprintf("Discipline: %s; level: %s; speciality: %s; assignment: %s.\nChallenge level %d: %s\nChallenge work with reasons, never personal hostility. Titles and challenge grant no permissions or authority.", p.Discipline, p.Level, p.Speciality, p.Assignment, n, modes[n])
}
func personalName(st room.State) string {
	names := []string{"Alex", "Amelia", "Arjun", "Ava", "Benjamin", "Chloe", "Daniel", "Elena", "Ethan", "Freya", "Gabriel", "Grace", "Hannah", "Isaac", "Isla", "James", "Leila", "Liam", "Maya", "Noah", "Oliver", "Priya", "Samuel", "Sofia", "Theo", "Zoe", "Codex"}
	used := map[string]bool{}
	for _, p := range st.Participants {
		used[p.Name] = true
	}
	offset, _ := rand.Int(rand.Reader, big.NewInt(int64(len(names))))
	for i := range names {
		n := names[(i+int(offset.Int64()))%len(names)]
		if !used[n] {
			return n
		}
	}
	return "Alex " + id()[:4]
}

func providerForEndpoint(endpoint string) string {
	if endpoint == "claude" || strings.HasSuffix(endpoint, "/claude") {
		return "claude"
	}
	return "openai"
}
