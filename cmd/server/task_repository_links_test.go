package main

import (
	"agentroom/internal/room"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func repositoryLinksFixture(t *testing.T) (*server, *mockDB, http.Handler, room.Task) {
	t.Helper()
	s, db, _, task := structuredTaskFixture(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/task-repository-links", s.taskRepositoryLinks)
	mux.HandleFunc("/api/events", s.events)
	return s, db, s.auth.protect(s.actorGuard(mux)), task
}
func repositoryLinksCall(h http.Handler, method, actor, key, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, "http://127.0.0.1:8787/api/task-repository-links?room=engineering&taskId=task", strings.NewReader(body))
	if actor != "" {
		r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: actor})
	}
	r.Header.Set("X-CSRF-Token", "csrf")
	if key != "" {
		r.Header.Set("Idempotency-Key", key)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}
func repositoryLinksBody(task room.Task, revision int, links []room.TaskRepositoryLink) string {
	raw, _ := json.Marshal(taskRepositoryLinksRequest{TaskID: task.ID, TaskRevision: task.Revision, Revision: revision, Links: links})
	return string(raw)
}
func TestTaskRepositoryLinksLifecycleFlagsAndReplay(t *testing.T) {
	t.Parallel()
	s, db, h, task := repositoryLinksFixture(t)
	if w := repositoryLinksCall(h, "GET", "owner", "", ""); w.Code != 403 {
		t.Fatal("not default off", w.Code)
	}
	body := repositoryLinksBody(task, 0, []room.TaskRepositoryLink{{Kind: "issue", URL: "https://github.com/example/synthetic/issues/1"}})
	if w := repositoryLinksCall(h, "POST", "owner", "off", body); w.Code != 403 {
		t.Fatal("disabled write", w.Code)
	}
	if w := mockRequest(h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"taskRepositoryLinks","enabled":true}}`); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	w := repositoryLinksCall(h, "POST", "owner", "save-links", body)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	afterSave := len(db.snapshot())
	again := repositoryLinksCall(h, "POST", "owner", "save-links", body)
	var firstJSON, replayJSON any
	if again.Code != 200 || json.Unmarshal(w.Body.Bytes(), &firstJSON) != nil || json.Unmarshal(again.Body.Bytes(), &replayJSON) != nil || !reflect.DeepEqual(firstJSON, replayJSON) || len(db.snapshot()) != afterSave {
		t.Fatal("retry changed response or appended history", again.Code)
	}
	if stale := repositoryLinksCall(h, "POST", "owner", "stale-links", body); stale.Code != 409 {
		t.Fatal("stale revision accepted", stale.Code)
	}
	changed := repositoryLinksBody(task, 1, []room.TaskRepositoryLink{})
	if conflict := repositoryLinksCall(h, "POST", "owner", "save-links", changed); conflict.Code != 409 {
		t.Fatal("key payload substitution accepted", conflict.Code)
	}
	if clear := repositoryLinksCall(h, "POST", "owner", "clear-links", changed); clear.Code != 200 {
		t.Fatal(clear.Code, clear.Body.String())
	}
	for _, enabled := range []bool{false, true} {
		db.seed("feature.flag.updated", map[string]any{"name": "taskRepositoryLinks", "enabled": enabled})
		got := repositoryLinksCall(h, "GET", "owner", "", "")
		if !enabled {
			if got.Code != 403 {
				t.Fatal(got.Code)
			}
		} else {
			var data struct {
				Record  room.TaskRepositoryLinks `json:"record"`
				CanEdit bool                     `json:"canEdit"`
			}
			if got.Code != 200 || json.Unmarshal(got.Body.Bytes(), &data) != nil || data.Record.Revision != 2 || len(data.Record.Links) != 0 || !data.CanEdit {
				t.Fatal("retained record not restored", got.Body.String())
			}
		}
	}
	st, err := s.load(t.Context())
	if err != nil || st.Tasks[0].Revision != task.Revision || st.TaskRepositoryLinks[taskRepositoryLinksKey(st.Company.ID, "engineering", task.ID)].Revision != 2 {
		t.Fatal("metadata changed task or failed replay", err)
	}
}
func TestTaskRepositoryLinksPermissionsAndIntegrity(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"member-read", "member-write", "outsider", "agent", "disabled-human", "anonymous", "csrf", "flag-owner-only", "incident", "persistence"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			s, db, h, task := repositoryLinksFixture(t)
			db.seed("feature.flag.updated", map[string]any{"name": "taskRepositoryLinks", "enabled": true})
			actor := room.Actor{ID: "reader", Name: "Synthetic reader", Kind: "human", Enabled: true, RoomIDs: []string{"engineering"}}
			if mode == "outsider" {
				actor.RoomIDs = []string{"elsewhere"}
			}
			if mode == "agent" {
				actor.Kind = "agent"
			}
			if mode == "disabled-human" {
				actor.Enabled = false
			}
			db.seed("actor.revised", actor)
			s.auth.sessions["reader"] = ownerSession{ActorID: actor.ID, CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
			body := repositoryLinksBody(task, 0, []room.TaskRepositoryLink{})
			cookie := "reader"
			method := "POST"
			if mode == "member-read" || mode == "outsider" || mode == "agent" || mode == "disabled-human" || mode == "anonymous" {
				method = "GET"
			}
			if mode == "anonymous" {
				cookie = ""
			}
			if mode == "incident" || mode == "persistence" || mode == "csrf" {
				cookie = "owner"
			}
			if mode == "incident" {
				db.seed("incident.opened", map[string]string{"reason": "synthetic stop"})
			}
			before := len(db.snapshot())
			var w *httptest.ResponseRecorder
			switch mode {
			case "csrf":
				r := httptest.NewRequest("POST", "http://127.0.0.1:8787/api/task-repository-links?room=engineering", strings.NewReader(body))
				r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: "owner"})
				w = httptest.NewRecorder()
				h.ServeHTTP(w, r)
			case "flag-owner-only":
				s.auth.sessions["owner"] = s.auth.sessions["reader"]
				w = mockRequest(h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"taskRepositoryLinks","enabled":false}}`)
			default:
				if mode == "persistence" {
					db.failAt = "commit"
				}
				w = repositoryLinksCall(h, method, cookie, "permission-check", body)
				db.failAt = ""
			}
			if mode == "member-read" {
				if w.Code != 200 || !strings.Contains(w.Body.String(), `"canEdit":false`) {
					t.Fatal(w.Code, w.Body.String())
				}
			} else if w.Code >= 200 && w.Code < 300 {
				t.Fatal("unauthorized operation succeeded", mode, w.Code)
			}
			if len(db.snapshot()) != before {
				t.Fatal("denied/read operation wrote history", mode)
			}
			if mode == "persistence" {
				if !s.integrityFault {
					t.Fatal("write failure did not stop integrity")
				}
				s.integrityFault = false
			}
		})
	}
}
func TestTaskRepositoryLinkURLBoundaries(t *testing.T) {
	t.Parallel()
	for _, link := range []room.TaskRepositoryLink{{Kind: "issue", URL: "https://github.com/example/repo/issues/1"}, {Kind: "pull-request", URL: "https://github.com/example/repo/pull/2"}, {Kind: "branch", URL: "https://github.com/example/repo/tree/Feature/branch"}, {Kind: "commit", URL: "https://github.com/example/repo/commit/" + strings.Repeat("a", 40)}} {
		if !validTaskRepositoryLink(link) {
			t.Fatal(link)
		}
	}
	for _, url := range []string{"http://github.com/example/repo/issues/1", "https://github.com.evil.test/example/repo/issues/1", "https://user:secret@github.com/example/repo/issues/1", "https://github.com/example/repo/issues/0", "https://github.com/example/repo/issues/1?token=synthetic", "https://github.com/example/repo/issues/1#fragment", "https://github.com/example/repo/issues/%31", "https://github.com/example/repo/issues/../1", strings.Repeat("x", 1025)} {
		if validTaskRepositoryLink(room.TaskRepositoryLink{Kind: "issue", URL: url}) {
			t.Fatal("unsafe URL accepted", url)
		}
	}
}
func TestTaskRepositoryLinksInvalidWritesHaveNoEffects(t *testing.T) {
	t.Parallel()
	_, db, h, task := repositoryLinksFixture(t)
	db.seed("feature.flag.updated", map[string]any{"name": "taskRepositoryLinks", "enabled": true})
	link := room.TaskRepositoryLink{Kind: "issue", URL: "https://github.com/example/repo/issues/1"}
	tooMany := make([]room.TaskRepositoryLink, 17)
	for i := range tooMany {
		tooMany[i] = link
	}
	cases := []struct {
		body, key string
		code      int
	}{
		{repositoryLinksBody(task, 0, nil), "nil", 400},
		{repositoryLinksBody(task, 0, tooMany), "many", 400},
		{repositoryLinksBody(task, 0, []room.TaskRepositoryLink{link, link}), "duplicate", 400},
		{repositoryLinksBody(task, 0, []room.TaskRepositoryLink{link}), "", 400},
		{`{"taskId":"task","taskRevision":1,"revision":0,"links":[],"extra":true}`, "unknown", 400},
		{`{"taskId":"task","taskId":"other","taskRevision":1,"revision":0,"links":[]}`, "duplicate-field", 400},
		{repositoryLinksBody(room.Task{ID: "task", Revision: task.Revision + 1}, 0, []room.TaskRepositoryLink{}), "task-stale", 409},
		{strings.Repeat("x", 24577), "oversize", 400},
	}
	before := len(db.snapshot())
	for _, c := range cases {
		w := repositoryLinksCall(h, "POST", "owner", c.key, c.body)
		if w.Code != c.code {
			t.Fatal(c.key, w.Code, w.Body.String())
		}
	}
	if len(db.snapshot()) != before {
		t.Fatal("invalid requests wrote history")
	}
	if w := repositoryLinksCall(h, "DELETE", "owner", "", ""); w.Code != 405 {
		t.Fatal(w.Code)
	}
	if w := mockRequest(h, "POST", "/api/events", `{"kind":"task.repository_links.updated","payload":{}}`); w.Code < 400 {
		t.Fatal("public event injection")
	}
}
