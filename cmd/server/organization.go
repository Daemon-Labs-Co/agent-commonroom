package main

import (
	"agentroom/internal/room"
	"errors"
	"math"
	"net/http"
	"strings"
)

// validateOrganization checks a bounded company graph, independently of layout.
// Reporting lines replace numbered authority, never account roles or room access.
func validateOrganization(chart room.OrganizationChart, actors []room.Actor) error {
	if len(chart.Nodes) > 500 {
		return errors.New("organization charts support up to 500 identities")
	}
	known := map[string]bool{}
	for _, a := range actors {
		known[a.ID] = true
	}
	nodes := map[string]string{}
	for _, n := range chart.Nodes {
		if !known[n.ActorID] || n.ActorID == "" {
			return errors.New("choose an existing company identity")
		}
		if _, exists := nodes[n.ActorID]; exists {
			return errors.New("duplicate company identity")
		}
		if math.IsNaN(n.X) || math.IsInf(n.X, 0) || math.IsNaN(n.Y) || math.IsInf(n.Y, 0) || n.X < 0 || n.Y < 0 || n.X > 50000 || n.Y > 50000 {
			return errors.New("node coordinates must be finite and between 0 and 50000")
		}
		if n.ReportsTo == n.ActorID || n.ReportsTo != "" && !known[n.ReportsTo] {
			return errors.New("choose a different existing reporting identity")
		}
		nodes[n.ActorID] = n.ReportsTo
	}
	for _, parent := range nodes {
		if parent != "" {
			if _, exists := nodes[parent]; !exists {
				return errors.New("reporting identity must be included in the chart")
			}
		}
	}
	visited := map[string]bool{}
	for actor := range nodes {
		path := map[string]bool{}
		for cursor := actor; cursor != "" && !visited[cursor]; cursor = nodes[cursor] {
			if path[cursor] {
				return errors.New("reporting relationships cannot contain a cycle")
			}
			path[cursor] = true
		}
		for v := range path {
			visited[v] = true
		}
	}
	return nil
}

// organization exposes only company identity metadata to the owner. It omits
// keys, endpoints, conversations and agent instructions/context.
func (s *server) organization(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", 405)
		return
	}
	if requestActor(r) != ownerID {
		http.Error(w, "Owner permission required", 403)
		return
	}
	st, err := s.load(r.Context())
	if err != nil {
		http.Error(w, "Organization unavailable", 503)
		return
	}
	if enabled, exists := st.FeatureFlags["companyOrgChart"]; exists && !enabled {
		http.Error(w, "Company organization chart is disabled", 403)
		return
	}
	type identity struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Title   string `json:"title"`
		Kind    string `json:"kind"`
		Role    string `json:"role"`
		Enabled bool   `json:"enabled"`
	}
	type profile struct {
		ActorID string `json:"actorId"`
		Name    string `json:"name"`
		Role    string `json:"role"`
		Room    string `json:"room"`
	}
	identities := []identity{}
	for _, a := range st.Actors {
		identities = append(identities, identity{a.ID, a.Name, a.Title, a.Kind, a.Role, a.Enabled})
	}
	profiles := []profile{}
	for _, p := range st.Participants {
		name := "Room"
		for _, r := range st.Rooms {
			if r.ID == p.RoomID {
				name = r.Name
			}
		}
		role := p.Role
		if p.Discipline != "" {
			role = strings.TrimSpace(p.Level + " " + p.Speciality + " " + p.Discipline)
		}
		profiles = append(profiles, profile{p.ActorID, p.Name, role, name})
	}
	reply(w, map[string]any{"company": st.Company, "actors": identities, "profiles": profiles, "organization": st.Organization})
}

// organizationGuidance supplies only relationships between identities currently
// participating in this room. It is context, not automatic conflict arbitration.
func organizationGuidance(st room.State, participant room.Participant, rid string) string {
	if enabled, exists := st.FeatureFlags["companyOrgChart"]; exists && !enabled {
		return ""
	}
	actor := participant.ActorID
	if actor == "" {
		actor = "bot-" + participant.ConnectionID
	}
	names := map[string]string{}
	for _, p := range st.Participants {
		if p.RoomID == rid {
			a := p.ActorID
			if a == "" {
				a = "bot-" + p.ConnectionID
			}
			names[a] = p.Name
		}
	}
	text := "Decision authority follows company reporting relationships, not numbered seniority. Managers outrank direct and indirect reports. Peers and unrelated branches must reach consensus. Explicit human directions are final; immutable history and security boundaries remain enforced. Unresolved disagreements must be reported, never treated as approval.\n"
	for other, name := range names {
		if room.OrganizationallyOutranks(st.Organization, other, actor) {
			text += name + " outranks you in this room.\n"
		}
		if room.OrganizationallyOutranks(st.Organization, actor, other) {
			text += "You outrank " + name + " in this room.\n"
		}
	}
	return text
}
