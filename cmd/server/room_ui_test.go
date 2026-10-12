package main

import (
	"agentroom/internal/room"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRoomPresentationFlags(t *testing.T) {
	t.Parallel()
	for _, flag := range []string{"squadProfileActions", "agentInvitationLayout", "standardHelpIcons", "featureFlagTabs", "historyDateGroups", "settingsTabs", "roomSections", "humanProfileDetails", "topAlignedViews", "compactProfileOverlays", "companyOrgChart", "newestMessagesFirst", "roomViewportFit", "featureFlagSwitches"} {
		t.Run(flag, func(t *testing.T) {
			t.Parallel()
			s, db, h := mockServer(t)
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, "presentation.js"), []byte("/* synthetic presentation */"), 0600); err != nil {
				t.Fatal(err)
			}
			handler := s.roomPresentationScript(root, "presentation.js", flag)
			check := func(expected int) {
				t.Helper()
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, httptest.NewRequest("GET", "/presentation.js", nil))
				if w.Code != expected {
					t.Fatal(w.Code, w.Body.String())
				}
				if w.Header().Get("Cache-Control") != "no-store" {
					t.Fatal("flag delivery must not be cached")
				}
			}
			if flag == "newestMessagesFirst" {
				check(204)
			} else {
				check(200)
			}
			for _, enabled := range []bool{false, true} {
				value := "false"
				expected := 204
				if enabled {
					value = "true"
					expected = 200
				}
				w := mockRequest(h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"`+flag+`","enabled":`+value+`}}`)
				if w.Code != 200 {
					t.Fatal(w.Code, w.Body.String())
				}
				check(expected)
			}
			db.seed("actor.revised", room.Actor{ID: "developer", Name: "Dev", Kind: "human", Role: "Developer", Enabled: true, RoomIDs: []string{"engineering"}})
			s.auth.sessions["owner"] = ownerSession{ActorID: "developer", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
			// An authenticated invited human cannot change cosmetic flags.
			if w := mockRequest(h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"`+flag+`","enabled":false}}`); w.Code == 200 {
				t.Fatal("non-owner changed flag")
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, httptest.NewRequest("POST", "/presentation.js", nil))
			if w.Code != 405 {
				t.Fatal(w.Code)
			}
		})
	}
}
