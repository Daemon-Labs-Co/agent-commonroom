package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"agentroom/internal/room"
)

func TestWorkflowTemplateFlagAndPermissions(t *testing.T) {
	t.Parallel()
	body := `{"kind":"workflow.revised","templateId":"ui-refinement","payload":{"name":"UI refinement","maxIterations":3,"nodes":[],"edges":[]}}`
	for _, c := range []struct {
		name, actor, role, target string
		enabled                   bool
		want                      int
	}{
		{"owner", "local-owner", "Owner", "engineering", true, 200},
		{"disabled", "local-owner", "Owner", "engineering", false, 403},
		{"developer", "developer", "Developer", "engineering", true, 200},
		{"qa", "qa", "QA", "engineering", true, 403},
		{"room denied", "developer", "Developer", "other", true, 403},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s, db, h := mockServer(t)
			s.templatesEnabled = c.enabled
			db.seed("room.created", room.Room{ID: "other", Name: "Other room"})
			if c.actor != ownerID {
				db.seed("actor.revised", room.Actor{ID: c.actor, Name: "Member", Kind: "human", Role: c.role, Enabled: true, RoomIDs: []string{"engineering"}})
				s.auth.sessions["owner"] = ownerSession{ActorID: c.actor, CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
			}
			before := len(db.snapshot())
			w := mockRequest(h, "POST", "/api/events?room="+c.target, body)
			if w.Code != c.want {
				t.Fatal(w.Code, w.Body.String())
			}
			if c.want != 200 {
				if len(db.snapshot()) != before {
					t.Fatal("rejected application persisted data")
				}
				return
			}
			if !strings.Contains(w.Body.String(), "workflowId") {
				t.Fatal("created workflow ID missing")
			}
			st, err := s.load(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if len(st.Workflows) != 1 || st.Workflows[0].RoomID != c.target || st.Workflows[0].Revision != 1 {
				t.Fatal("workflow not applied to target room")
			}
			w = mockRequest(h, "POST", "/api/events?room="+c.target, body)
			if w.Code != 200 {
				t.Fatal(w.Body.String())
			}
			st, err = s.load(context.Background())
			if err != nil || len(st.Workflows) != 2 || st.Workflows[0].ID == st.Workflows[1].ID {
				t.Fatal("reapplication replaced an existing workflow")
			}
		})
	}
	t.Run("unknown", func(t *testing.T) {
		t.Parallel()
		s, db, h := mockServer(t)
		s.templatesEnabled = true
		w := mockRequest(h, "POST", "/api/events", strings.Replace(body, "ui-refinement", "unknown", 1))
		if w.Code != 400 || len(db.snapshot()) != 0 {
			t.Fatal("unknown template accepted")
		}
	})
}
func TestWorkflowTemplateAssetFlag(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "workflow-templates.js"), []byte("/* templates */"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, enabled := range []bool{true, false} {
		w := httptest.NewRecorder()
		optionalScriptHandler(enabled, root, "workflow-templates.js").ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/workflow-templates.js", nil))
		if enabled && w.Code != 200 || !enabled && (w.Code != 204 || w.Body.Len() != 0) {
			t.Fatal("asset flag failed", w.Code)
		}
	}
}

func TestAdditionalTemplateApplications(t *testing.T) {
	t.Parallel()
	for _, id := range []string{"documenter", "disaster-planning", "cooperative-adversarial"} {
		t.Run(id, func(t *testing.T) {
			t.Parallel()
			for _, enabled := range []bool{true, false} {
				s, db, h := mockServer(t)
				s.templatesEnabled = enabled
				body := `{"kind":"workflow.revised","templateId":"` + id + `","payload":{"name":"` + id + `","maxIterations":3,"nodes":[],"edges":[]}}`
				w := mockRequest(h, "POST", "/api/events", body)
				if !enabled {
					if w.Code != 403 || len(db.snapshot()) != 0 {
						t.Fatal("disabled template accepted")
					}
					continue
				}
				if w.Code != 200 {
					t.Fatal("new template rejected", w.Code, w.Body.String())
				}
				st, err := s.load(context.Background())
				if err != nil || len(st.Workflows) != 1 || st.Workflows[0].Name != id {
					t.Fatal("template not persisted in mock")
				}
			}
		})
	}
}
