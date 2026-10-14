package main

import (
	"agentroom/internal/room"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func dispatchFixture(t *testing.T) (*server, *mockDB, http.Handler, chan struct{}) {
	s, db, _ := mockServer(t)
	db.seed("connection.created", room.Connection{ID: "connection", Provider: "openai", Protocol: "CLI", Endpoint: "ssh://synthetic-worker"})
	db.seed("participant.created", room.Participant{ID: "dev", RoomID: "engineering", ConnectionID: "connection", Name: "Synthetic developer"})
	db.seed("task.revised", room.Task{ID: "task", RoomID: "engineering", Title: "Synthetic feature", Criteria: "Synthetic acceptance", Assignee: "dev", Reviewer: "dev", Status: "Assigned", Revision: 1})
	release := make(chan struct{})
	s.taskReady = func(string) error { return nil }
	s.taskTurn = func(ctx context.Context, _ room.Participant, _, _ string) (string, error) {
		select {
		case <-release:
			return "Synthetic result", nil
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/task-dispatch", s.taskDispatch)
	mux.HandleFunc("/api/events", s.events)
	return s, db, s.auth.protect(s.actorGuard(mux)), release
}
func dispatchRequest(h http.Handler, key, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest("POST", "http://127.0.0.1:8787/api/task-dispatch?room=engineering", strings.NewReader(body))
	r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: "owner"})
	r.Header.Set("X-CSRF-Token", "csrf")
	r.Header.Set("Idempotency-Key", key)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

const dispatchBody = `{"taskId":"task","revision":1,"mode":"work"}`

func waitDispatch(t *testing.T, s *server) room.State {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		s.mu.Lock()
		active := s.active["dev"]
		st, err := s.load(context.Background())
		s.mu.Unlock()
		if err != nil {
			t.Fatal(err)
		}
		if !active {
			return st
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("dispatch did not finish")
	return room.State{}
}
func TestTaskDispatchAtomicReplayAndCompletion(t *testing.T) {
	t.Parallel()
	s, db, h, release := dispatchFixture(t)
	base := len(db.snapshot())
	w := dispatchRequest(h, "dispatch-key", dispatchBody)
	if w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	st, err := s.load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if st.Tasks[0].Status != "Working" || st.Tasks[0].DispatchStatus != "pending" || len(db.snapshot()) != base+5 {
		t.Fatal("intent not committed atomically")
	}
	replay := dispatchRequest(h, "dispatch-key", dispatchBody)
	if replay.Code != 202 || replay.Header().Get("Idempotent-Replay") != "true" || len(db.snapshot()) != base+5 {
		t.Fatal("replay duplicated work")
	}
	conflict := dispatchRequest(h, "dispatch-key", `{"taskId":"task","revision":1,"mode":"review"}`)
	if conflict.Code != 409 {
		t.Fatal("key reuse accepted")
	}
	close(release)
	st = waitDispatch(t, s)
	if st.Tasks[0].DispatchStatus != "completed" || st.Tasks[0].Revision != 3 || st.Tasks[0].Status == "Accepted" {
		t.Fatal("incorrect task result")
	}
}
func TestTaskDispatchRollbackNeverLaunches(t *testing.T) {
	for _, stage := range []string{"first", "second", "third", "fourth", "receipt", "commit"} {
		t.Run(stage, func(t *testing.T) {
			t.Parallel()
			s, db, h, _ := dispatchFixture(t)
			base := len(db.snapshot())
			switch stage {
			case "first":
				db.failInsertAt = 1
			case "second":
				db.failInsertAt = 2
			case "third":
				db.failInsertAt = 3
			case "fourth":
				db.failInsertAt = 4
			case "receipt":
				db.failInsertAt = 5
			case "commit":
				db.failAt = "commit"
			}
			w := dispatchRequest(h, "rollback-key", dispatchBody)
			if w.Code != 500 || len(db.snapshot()) != base || s.active["dev"] {
				t.Fatal("partial dispatch", w.Code)
			}
		})
	}
}
func TestTaskDispatchRejections(t *testing.T) {
	for _, kind := range []string{"disabled", "stale", "foreign", "busy", "stopped", "unready", "forged", "csrf", "nonowner", "accepted", "local", "capacity", "disclosure"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()
			s, db, h, _ := dispatchFixture(t)
			body := dispatchBody
			code := 409
			switch kind {
			case "disabled":
				db.seed("feature.flag.updated", map[string]any{"name": "taskDispatch", "enabled": false})
				code = 403
			case "stale":
				body = `{"taskId":"task","revision":2,"mode":"work"}`
			case "foreign":
				body = `{"taskId":"other-room-task","revision":1,"mode":"work"}`
				code = 404
			case "busy":
				s.active["dev"] = true
			case "stopped":
				s.stopAll = true
			case "unready":
				s.taskReady = func(string) error { return errors.New("unavailable") }
				code = 503
			case "forged":
				body = `{"taskId":"task","revision":1,"mode":"work","participantId":"other"}`
				code = 400
			case "accepted":
				db.seed("task.revised", room.Task{ID: "task", RoomID: "engineering", Title: "Synthetic", Assignee: "dev", Status: "Accepted", Revision: 1})
			case "local":
				db.seed("connection.updated", room.Connection{ID: "connection", Provider: "openai", Protocol: "CLI", Endpoint: "codex"})
				code = 503
			case "capacity":
				for i := 0; i < 8; i++ {
					s.active[string(rune('a'+i))] = true
				}
				code = 429
			case "disclosure":
				db.seed("room.updated", room.Room{ID: "engineering", Name: "Synthetic", ContextExport: &room.ContextExportPolicy{Providers: []string{"claude"}}})
				code = 413
			case "csrf", "nonowner":
				code = 403
			}
			base := len(db.snapshot())
			var w *httptest.ResponseRecorder
			if kind == "csrf" || kind == "nonowner" {
				r := httptest.NewRequest("POST", "http://127.0.0.1:8787/api/task-dispatch", strings.NewReader(body))
				cookie := "owner"
				if kind == "nonowner" {
					db.seed("actor.revised", room.Actor{ID: "human", Kind: "human", Enabled: true, Role: "Developer"})
					base = len(db.snapshot())
					s.auth.sessions["human"] = ownerSession{ActorID: "human", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
					cookie = "human"
					r.Header.Set("X-CSRF-Token", "csrf")
				}
				r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: cookie})
				w = httptest.NewRecorder()
				h.ServeHTTP(w, r)
			} else {
				w = dispatchRequest(h, "reject-key", body)
			}
			if w.Code != code || len(db.snapshot()) != base {
				t.Fatal(kind, w.Code, w.Body.String())
			}
		})
	}
}
func TestTaskFailureAndReconciliation(t *testing.T) {
	t.Parallel()
	s, db, h, release := dispatchFixture(t)
	s.taskTurn = func(context.Context, room.Participant, string, string) (string, error) {
		return "", errors.New("Synthetic provider failure")
	}
	if w := dispatchRequest(h, "failure-key", dispatchBody); w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	st := waitDispatch(t, s)
	if st.Tasks[0].DispatchStatus != "failed" || st.Participants[0].Status != "Blocked" {
		t.Fatal("failure not visible")
	}
	close(release)
	task := st.Tasks[0]
	task.Revision++
	task.DispatchStatus = "pending"
	db.seed("task.revised", task)
	s.recoveryIdle = func(context.Context, string) (bool, error) { return false, nil }
	body := `{"taskId":"task","revision":4,"mode":"reconcile"}`
	if w := dispatchRequest(h, "reconcile-key", body); w.Code != 503 {
		t.Fatal(w.Code, w.Body.String())
	}
	s.recoveryIdle = func(context.Context, string) (bool, error) { return true, nil }
	if w := dispatchRequest(h, "reconcile-key", body); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
}

func TestTaskEditCannotForgeOrChangePendingDispatch(t *testing.T) {
	t.Parallel()
	_, db, h, _ := dispatchFixture(t)
	body := `{"kind":"task.revised","payload":{"id":"task","revision":1,"title":"Synthetic","status":"Assigned","dispatchId":"forged","dispatchStatus":"completed"}}`
	w := mockRequest(h, "POST", "/api/events?room=engineering", body)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var task room.Task
	events := db.snapshot()
	if err := json.Unmarshal([]byte(events[len(events)-1].payload), &task); err != nil {
		t.Fatal(err)
	}
	if task.DispatchID != "" || task.DispatchStatus != "" {
		t.Fatal("client forged dispatch provenance")
	}
	task.DispatchID = "pending-id"
	task.DispatchStatus = "pending"
	db.seed("task.revised", task)
	w = mockRequest(h, "POST", "/api/events?room=engineering", `{"kind":"task.revised","payload":{"id":"task","revision":2,"title":"Changed","status":"Accepted"}}`)
	if w.Code != 409 {
		t.Fatal("pending task was edited", w.Code)
	}
}

func TestTaskResultWriteFailureLeavesIntentPendingAndStopsExecution(t *testing.T) {
	t.Parallel()
	s, db, h, release := dispatchFixture(t)
	if w := dispatchRequest(h, "result-key", dispatchBody); w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	s.mu.Lock()
	db.failInsertAt = 3
	s.mu.Unlock() // Fail task result after staged message/session.
	close(release)
	st := waitDispatch(t, s)
	if !s.integrityFault || st.Tasks[0].DispatchStatus != "pending" || st.Participants[0].Status != "Working" {
		t.Fatal("partial result or unsafe persistence failure")
	}
}

func TestTaskDispatchFlagOffPreservesInflightResult(t *testing.T) {
	t.Parallel()
	s, db, h, release := dispatchFixture(t)
	if w := dispatchRequest(h, "flag-key", dispatchBody); w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	s.mu.Lock()
	db.seed("feature.flag.updated", map[string]any{"name": "taskDispatch", "enabled": false})
	s.mu.Unlock()
	close(release)
	st := waitDispatch(t, s)
	if st.Tasks[0].DispatchStatus != "completed" {
		t.Fatal("turning off capability lost existing result")
	}
}
