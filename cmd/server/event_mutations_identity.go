package main

// Identity event validation and typed payload preparation; persistence stays in the handler.

import (
	"agentroom/internal/room"
	"encoding/json"
	"net/http"
	"strings"
)

func (s *server) prepareIdentityEvent(w http.ResponseWriter, r *http.Request, st room.State, roomID string, req eventRequest) (any, bool) {
	var p any
	switch req.Kind {
	case "agents.idle_policy.updated":
		return prepareIdleSettings(w, req, st)
	case "feature.flag.updated":
		var v struct {
			Name    string `json:"name"`
			Enabled *bool  `json:"enabled"`
		}
		if json.Unmarshal(req.Payload, &v) != nil || (v.Name != "taskRepositoryLinks" && v.Name != "taskProgress" && v.Name != "taskAcceptanceReview" && v.Name != "taskValueDependencies" && v.Name != "taskDefinitions" && v.Name != "roomOutcomeReview" && v.Name != "roomOutcomeContext" && v.Name != "roomOutcomes" && v.Name != "automaticAgentRooms" && v.Name != "taskDelegation" && v.Name != "taskProposals" && v.Name != "agentCoordination" && v.Name != "idleStopNotices" && v.Name != "idleAgentShutdown" && v.Name != "workspaceDisposal" && v.Name != "taskDispatch" && v.Name != "contextPolicyEditor" && v.Name != "jobDiagnostics" && v.Name != "agentRecovery" && v.Name != "agentBackgroundWork" && v.Name != "agentSafetyControls" && v.Name != "humanProfileImages" && v.Name != "connectionSignIn" && v.Name != "connectionTesting" && v.Name != "helpWindowChoice" && v.Name != "squadProfileActions" && v.Name != "agentInvitationLayout" && v.Name != "standardHelpIcons" && v.Name != "featureFlagTabs" && v.Name != "historyDateGroups" && v.Name != "settingsTabs" && v.Name != "roomSections" && v.Name != "humanProfileDetails" && v.Name != "topAlignedViews" && v.Name != "compactProfileOverlays" && v.Name != "companyOrgChart" && v.Name != "newestMessagesFirst" && v.Name != "roomViewportFit" && v.Name != "featureFlagSwitches" && v.Name != "workflowExecution" && v.Name != "agentStatusDetails" && v.Name != "squadMenuLayout" && v.Name != "humanAttention" && v.Name != "historyRefinement") || v.Enabled == nil {
			http.Error(w, "Choose a supported feature flag and an enabled value", 400)
			return nil, false
		}
		p = v
	case "actor.help.updated":
		var v struct {
			ID   string `json:"id"`
			Mode string `json:"mode"`
		}
		if json.Unmarshal(req.Payload, &v) != nil || (v.Mode != "same-window" && v.Mode != "separate-window") {
			http.Error(w, "Choose same-window or separate-window", 400)
			return nil, false
		}
		if enabled, ok := st.FeatureFlags["helpWindowChoice"]; ok && !enabled {
			http.Error(w, "Help window choice is disabled", 403)
			return nil, false
		}
		v.ID = requestActor(r)
		if a, ok := member(st, v.ID); !ok || a.Kind != "human" {
			http.Error(w, "Human profile required", 403)
			return nil, false
		}
		p = v
	case "actor.image.updated":
		var v struct {
			ID    string  `json:"id"`
			Image *string `json:"image"`
		}
		if json.Unmarshal(req.Payload, &v) != nil || v.Image == nil {
			http.Error(w, "Image required (empty removes photo)", 400)
			return nil, false
		}
		v.ID = requestActor(r)
		if a, ok := member(st, v.ID); !ok || a.Kind != "human" {
			http.Error(w, "Human profile required", 403)
			return nil, false
		}
		if !humanProfileImagesEnabled(st) && *v.Image != "" {
			http.Error(w, "Profile photos are disabled", 403)
			return nil, false
		}
		image, err := normalizeProfileImage(*v.Image)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return nil, false
		}
		v.Image = &image
		p = v
	case "actor.name.updated":
		var v struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		if json.Unmarshal(req.Payload, &v) != nil || strings.TrimSpace(v.Name) == "" || len(v.Name) > 100 {
			http.Error(w, "Name required (maximum 100 characters)", 400)
			return nil, false
		}
		v.ID = requestActor(r)
		if v.ID == ownerID {
			http.Error(w, "Use owner profile settings", 400)
			return nil, false
		}
		v.Name = strings.TrimSpace(v.Name)
		p = v
	case "company.organization.updated":
		if enabled, exists := st.FeatureFlags["companyOrgChart"]; exists && !enabled {
			http.Error(w, "Company organization chart is disabled", 403)
			return nil, false
		}
		var v room.OrganizationChart
		if json.Unmarshal(req.Payload, &v) != nil {
			http.Error(w, "Invalid organization chart", 400)
			return nil, false
		}
		if v.Revision != st.Organization.Revision {
			http.Error(w, "Organization changed; reload before saving", 409)
			return nil, false
		}
		if err := validateOrganization(v, st.Actors); err != nil {
			http.Error(w, err.Error(), 400)
			return nil, false
		}
		v.Revision++
		p = v
	case "company.updated":
		var v room.Company
		if json.Unmarshal(req.Payload, &v) != nil || strings.TrimSpace(v.Name) == "" || len(v.Name) > 100 || len(v.Description) > 2000 {
			http.Error(w, "Company name required (100 characters); description maximum 2000 characters", 400)
			return nil, false
		}
		if v.Revision != st.Company.Revision {
			http.Error(w, "Company changed; reload", 409)
			return nil, false
		}
		v.Revision++
		v.ID = st.Company.ID
		v.Name = strings.TrimSpace(v.Name)
		v.Description = strings.TrimSpace(v.Description)
		p = v
	case "human.profile.updated":
		var v struct {
			Revision int     `json:"revision"`
			Name     string  `json:"name"`
			Title    *string `json:"title"`
		}
		if json.Unmarshal(req.Payload, &v) != nil || strings.TrimSpace(v.Name) == "" || len(v.Name) > 100 {
			http.Error(w, "Human name required (maximum 100 characters)", 400)
			return nil, false
		}
		if v.Revision != st.HumanRevision {
			http.Error(w, "Profile changed; reload", 409)
			return nil, false
		}
		if v.Title == nil {
			title := st.HumanTitle
			v.Title = &title
		}
		title := strings.TrimSpace(*v.Title)
		if len(title) > 100 {
			http.Error(w, "Title must be at most 100 characters", 400)
			return nil, false
		}
		v.Title = &title
		v.Name = strings.TrimSpace(v.Name)
		p = v
	}
	return p, true
}
