package main

import (
	"agentroom/internal/room"
	"context"
	"encoding/json"
	"math"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestOrganizationValidation(t *testing.T) {
	t.Parallel()
	actors := []room.Actor{{ID: "a"}, {ID: "b"}, {ID: "c"}}
	for _, tc := range []struct {
		name  string
		nodes []room.OrganizationNode
		valid bool
	}{
		{"valid", []room.OrganizationNode{{ActorID: "a"}, {ActorID: "b", ReportsTo: "a"}}, true},
		{"unknown", []room.OrganizationNode{{ActorID: "unknown"}}, false},
		{"duplicate", []room.OrganizationNode{{ActorID: "a"}, {ActorID: "a"}}, false},
		{"self", []room.OrganizationNode{{ActorID: "a", ReportsTo: "a"}}, false},
		{"cycle", []room.OrganizationNode{{ActorID: "a", ReportsTo: "b"}, {ActorID: "b", ReportsTo: "a"}}, false},
		{"missing manager node", []room.OrganizationNode{{ActorID: "a", ReportsTo: "b"}}, false},
		{"nonfinite", []room.OrganizationNode{{ActorID: "a", X: math.NaN()}}, false},
		{"outside", []room.OrganizationNode{{ActorID: "a", Y: 50001}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if (validateOrganization(room.OrganizationChart{Nodes: tc.nodes}, actors) == nil) != tc.valid {
				t.Fatal(tc.name)
			}
		})
	}
}
func TestOrganizationSavePermissionsAndProjection(t *testing.T) {
	t.Parallel()
	s, db, h := mockServer(t)
	db.seed("actor.revised", room.Actor{ID: "dev", Name: "Synthetic developer", Kind: "human", Role: "Developer", Enabled: true, RoomIDs: []string{"engineering"}, PublicKey: "synthetic-never-a-real-key"})
	post := `{"kind":"company.organization.updated","payload":{"revision":0,"nodes":[{"actorId":"dev","reportsTo":"local-owner","x":25,"y":160},{"actorId":"local-owner","x":25,"y":20}]}}`
	if w := mockRequest(h, "POST", "/api/events", post); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	st, err := s.load(context.Background())
	if err != nil || st.Organization.Revision != 1 || len(st.Organization.Nodes) != 2 {
		t.Fatal(st.Organization, err)
	}
	if w := mockRequest(h, "POST", "/api/events", post); w.Code != 409 {
		t.Fatal("stale revision", w.Code)
	}
	endpoint := s.auth.protect(s.actorGuard(http.HandlerFunc(s.organization)))
	w := mockRequest(endpoint, "GET", "/api/company/organization", "")
	if w.Code != 200 || strings.Contains(w.Body.String(), "synthetic-never-a-real-key") {
		t.Fatal(w.Code, w.Body.String())
	}
	if w := mockRequest(endpoint, "POST", "/api/company/organization", `{}`); w.Code != 405 {
		t.Fatal(w.Code)
	}
	db.seed("feature.flag.updated", map[string]any{"name": "companyOrgChart", "enabled": false})
	if w := mockRequest(endpoint, "GET", "/api/company/organization", ""); w.Code != 403 {
		t.Fatal("disabled read", w.Code)
	}
	if w := mockRequest(h, "POST", "/api/events", strings.Replace(post, `"revision":0`, `"revision":1`, 1)); w.Code != 403 {
		t.Fatal("disabled write", w.Code)
	}
	db.seed("feature.flag.updated", map[string]any{"name": "companyOrgChart", "enabled": true})
	s.auth.sessions["owner"] = ownerSession{ActorID: "dev", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	if w := mockRequest(endpoint, "GET", "/api/company/organization", ""); w.Code != 403 {
		t.Fatal("non-owner read", w.Code)
	}
	if w := mockRequest(h, "POST", "/api/events", post); w.Code != 403 {
		t.Fatal("non-owner write", w.Code)
	}
	w = mockRequest(h, "GET", "/api/state", "")
	var visible room.State
	if err := json.Unmarshal(w.Body.Bytes(), &visible); err != nil {
		t.Fatal(err)
	}
	if len(visible.Organization.Nodes) != 0 {
		t.Fatal("company graph leaked to invited user")
	}
}
func TestOrganizationGuidanceScopedToRoom(t *testing.T) {
	t.Parallel()
	st := room.State{Organization: room.OrganizationChart{Nodes: []room.OrganizationNode{{ActorID: "bot-lead"}, {ActorID: "bot-dev", ReportsTo: "bot-lead"}}}, Participants: []room.Participant{{ActorID: "bot-lead", RoomID: "dev", Name: "Lead"}, {ActorID: "bot-other", RoomID: "private", Name: "Private identity"}}}
	p := room.Participant{ConnectionID: "dev"}
	g := organizationGuidance(st, p, "dev")
	if !strings.Contains(g, "Lead outranks you") || strings.Contains(g, "Private identity") || !strings.Contains(g, "Explicit human directions are final") {
		t.Fatal(g)
	}
	st.FeatureFlags = map[string]bool{"companyOrgChart": false}
	if organizationGuidance(st, p, "dev") != "" {
		t.Fatal("disabled guidance")
	}
}
