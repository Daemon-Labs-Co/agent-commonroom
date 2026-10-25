package main

import (
	"agentroom/internal/room"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func reviewFixture(t *testing.T, skipTasks ...bool) (*server, *mockDB, http.Handler) {
	t.Helper()
	s, db, _ := mockServer(t)
	db.seed("feature.flag.updated", map[string]any{"name": "roomOutcomes", "enabled": true})
	db.seed("feature.flag.updated", map[string]any{"name": "roomOutcomeReview", "enabled": true})
	db.seed("room.updated", room.Room{ID: "engineering", Name: "Synthetic room", Outcome: "Synthetic delivery", AcceptanceCriteria: "Owner verifies synthetic evidence", Revision: 1})
	if len(skipTasks) == 0 || !skipTasks[0] {
		db.seed("task.revised", room.Task{ID: "task", RoomID: "engineering", Title: "Synthetic step", Criteria: "Synthetic task criterion", Status: "Accepted", Revision: 1})
	}
	db.seed("evidence.created", room.Evidence{ID: "evidence", RoomID: "engineering", TaskID: "task", Kind: "Test", Body: "Synthetic fixture"})
	mux := http.NewServeMux()
	mux.HandleFunc("/api/room-outcome-review", s.roomOutcomeReview)
	mux.HandleFunc("/api/events", s.events)
	return s, db, s.auth.protect(s.actorGuard(mux))
}
func readRoomReview(t *testing.T, h http.Handler) map[string]json.RawMessage {
	t.Helper()
	w := mockRequest(h, "GET", "/api/room-outcome-review?room=engineering", "")
	var data map[string]json.RawMessage
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &data) != nil {
		t.Fatal(w.Code, w.Body.String())
	}
	return data
}
func roomReviewBody(t *testing.T, h http.Handler, action string) string {
	t.Helper()
	data := readRoomReview(t, h)
	var snapshot string
	var review room.RoomOutcomeReview
	_ = json.Unmarshal(data["snapshot"], &snapshot)
	_ = json.Unmarshal(data["review"], &review)
	raw, _ := json.Marshal(map[string]any{"action": action, "revision": review.Revision, "snapshot": snapshot, "note": "Synthetic evidence reviewed by owner"})
	return string(raw)
}
func submitRoomReview(h http.Handler, key, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest("POST", "http://127.0.0.1:8787/api/room-outcome-review?room=engineering", strings.NewReader(body))
	r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: "owner"})
	r.Header.Set("X-CSRF-Token", "csrf")
	r.Header.Set("Idempotency-Key", key)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}
func TestRoomOutcomeAcceptanceReplayReopenAndRestart(t *testing.T) {
	t.Parallel()
	s, db, h := reviewFixture(t)
	body := roomReviewBody(t, h, "accept")
	w := submitRoomReview(h, "review-accept", body)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	before := len(db.snapshot())
	w = submitRoomReview(h, "review-accept", body)
	if w.Code != 200 || w.Header().Get("Idempotent-Replay") != "true" || len(db.snapshot()) != before {
		t.Fatal("duplicate review", w.Code)
	}
	if string(readRoomReview(t, h)["status"]) != `"Accepted"` {
		t.Fatal("acceptance lost")
	}
	st, err := (&server{db: db}).load(context.Background())
	if err != nil || st.OutcomeReviews["engineering"].ReviewedBy != ownerID {
		t.Fatal("restart lost human review", err)
	}
	// Any room revision requires a fresh review, including a later text reversion.
	db.seed("room.updated", room.Room{ID: "engineering", Name: "Renamed", Outcome: "Synthetic delivery", AcceptanceCriteria: "Owner verifies synthetic evidence", Revision: 2})
	if string(readRoomReview(t, h)["status"]) != `"Needs review"` {
		t.Fatal("room revision did not invalidate review")
	}
	db.seed("task.revised", room.Task{ID: "later", RoomID: "engineering", Title: "New work", Status: "Draft", Revision: 1})
	if string(readRoomReview(t, h)["status"]) != `"Needs review"` {
		t.Fatal("new task did not invalidate acceptance")
	}
	if w = submitRoomReview(h, "stale-accept", body); w.Code != 409 {
		t.Fatal("stale review accepted", w.Code)
	}
	if w = submitRoomReview(h, "review-reopen", roomReviewBody(t, h, "reopen")); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if string(readRoomReview(t, h)["status"]) != `"Open"` || len(s.active) != 0 {
		t.Fatal("reopen launched work or retained acceptance")
	}
}
func TestRoomReviewRefusalsAndSnapshotChanges(t *testing.T) {
	for _, scenario := range []string{"undefined", "no tasks", "unaccepted", "pending", "personal", "flag off", "metadata off", "stale evidence", "stale definition", "member", "persistence", "concurrent"} {
		t.Run(scenario, func(t *testing.T) {
			t.Parallel()
			s, db, h := reviewFixture(t, scenario == "no tasks")
			body := roomReviewBody(t, h, "accept")
			want := 409
			switch scenario {
			case "undefined":
				db.seed("room.updated", room.Room{ID: "engineering", Name: "Legacy", Revision: 2})
				body = roomReviewBody(t, h, "accept")
			case "no tasks":
				// Fixture intentionally contains no tasks.
				body = roomReviewBody(t, h, "accept")
			case "unaccepted":
				db.seed("task.revised", room.Task{ID: "task", RoomID: "engineering", Title: "Synthetic", Status: "Submitted", Revision: 2})
				body = roomReviewBody(t, h, "accept")
			case "pending":
				db.seed("task.revised", room.Task{ID: "task", RoomID: "engineering", Title: "Synthetic", Status: "Accepted", Revision: 2, DispatchID: "pending", DispatchStatus: "pending"})
				body = roomReviewBody(t, h, "accept")
			case "personal":
				db.seed("room.updated", room.Room{ID: "engineering", PersonalActorID: ownerID, Name: "Personal", Outcome: "Synthetic", AcceptanceCriteria: "Synthetic", Revision: 2})
				want = 404
			case "flag off":
				db.seed("feature.flag.updated", map[string]any{"name": "roomOutcomeReview", "enabled": false})
				want = 403
			case "metadata off":
				db.seed("feature.flag.updated", map[string]any{"name": "roomOutcomes", "enabled": false})
				want = 403
			case "stale evidence":
				db.seed("evidence.created", room.Evidence{ID: "new", RoomID: "engineering", Body: "Changed evidence"})
			case "stale definition":
				db.seed("room.updated", room.Room{ID: "engineering", Name: "Synthetic", Outcome: "Changed", AcceptanceCriteria: "Changed", Revision: 2})
			case "member":
				db.seed("actor.revised", room.Actor{ID: "dev", Kind: "human", Enabled: true, RoomIDs: []string{"engineering"}})
				s.auth.sessions["owner"] = ownerSession{ActorID: "dev", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
				want = 403
			case "persistence":
				db.failInsertAt = 1
				want = 500
			case "concurrent":
				s.db = &interleavedEventDB{eventDatabase: db, beforeBegin: func(context.Context) {
					db.seed("feature.flag.updated", map[string]any{"name": "roomOutcomeReview", "enabled": false})
				}}
			}
			before := len(db.snapshot())
			w := submitRoomReview(h, "refuse-review", body)
			if w.Code != want {
				t.Fatal(scenario, w.Code, w.Body.String())
			}
			expected := before
			if scenario == "concurrent" {
				expected++
			}
			if len(db.snapshot()) != expected || len(s.active) != 0 {
				t.Fatal("refused decision had effects")
			}
		})
	}
}
func TestRoomReviewHistoryOffAndEventForgery(t *testing.T) {
	t.Parallel()
	_, db, h := reviewFixture(t)
	if w := submitRoomReview(h, "history-accept", roomReviewBody(t, h, "accept")); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	db.seed("feature.flag.updated", map[string]any{"name": "roomOutcomeReview", "enabled": false})
	state := readRoomReview(t, h)
	if string(state["status"]) != `"Accepted"` || string(state["enabled"]) != "false" {
		t.Fatal("off erased history")
	}
	if w := mockRequest(h, "POST", "/api/events", `{"kind":"room.outcome.reviewed","payload":{"roomId":"engineering","status":"Accepted"}}`); w.Code == 200 {
		t.Fatal("generic event forged acceptance")
	}
	_, err := (&server{db: db}).load(t.Context())
	if err != nil {
		t.Fatal("disabled history replay failed", err)
	}
}

func TestRoomReviewConfigurationAndMalformedRequests(t *testing.T) {
	t.Parallel()
	s, db, h := reviewFixture(t)
	w := mockRequest(h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"roomOutcomeReview","enabled":false}}`)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	w = mockRequest(h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"roomOutcomeReview","enabled":true}}`)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	good := roomReviewBody(t, h, "accept")
	var fields map[string]any
	_ = json.Unmarshal([]byte(good), &fields)
	for _, scenario := range []string{"blank note", "long note", "unknown field", "missing snapshot", "wrong action", "negative revision", "missing key", "csrf"} {
		t.Run(scenario, func(t *testing.T) {
			body := map[string]any{}
			for k, v := range fields {
				body[k] = v
			}
			key := "malformed-review"
			want := 400
			switch scenario {
			case "blank note":
				body["note"] = "  "
			case "long note":
				body["note"] = strings.Repeat("x", 4001)
			case "unknown field":
				body["reviewedBy"] = "agent"
			case "missing snapshot":
				delete(body, "snapshot")
			case "wrong action":
				body["action"] = "auto-accept"
			case "negative revision":
				body["revision"] = -1
			case "missing key":
				key = ""
			case "csrf":
				want = 403
			}
			raw, _ := json.Marshal(body)
			before := len(db.snapshot())
			var response *httptest.ResponseRecorder
			if scenario == "csrf" {
				r := httptest.NewRequest("POST", "http://127.0.0.1:8787/api/room-outcome-review?room=engineering", strings.NewReader(string(raw)))
				r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: "owner"})
				response = httptest.NewRecorder()
				h.ServeHTTP(response, r)
			} else {
				response = submitRoomReview(h, key, string(raw))
			}
			if response.Code != want || len(db.snapshot()) != before {
				t.Fatal(scenario, response.Code, response.Body.String())
			}
		})
	}
	// Authorized members can read only rooms assigned to them, without writes.
	db.seed("room.created", room.Room{ID: "private", Name: "Other room", Revision: 1})
	db.seed("actor.revised", room.Actor{ID: "member", Kind: "human", Enabled: true, RoomIDs: []string{"private"}})
	s.auth.sessions["owner"] = ownerSession{ActorID: "member", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	if w = mockRequest(h, "GET", "/api/room-outcome-review?room=engineering", ""); w.Code == 200 {
		t.Fatal("foreign review disclosed")
	}
}
