package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"agentroom/internal/room"
)

func mockServer(t *testing.T) (*server, *mockDB, http.Handler) {
	db := newMockDB(t)
	auth, err := newOwnerAuth(strings.Repeat("a", 64), "127.0.0.1:8787")
	if err != nil {
		t.Fatal(err)
	}
	auth.sessions["owner"] = ownerSession{ActorID: ownerID, CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	s := &server{db: db, auth: auth, active: map[string]bool{}, cancels: map[string]context.CancelFunc{}}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/events", s.events)
	mux.HandleFunc("/api/state", s.state)
	mux.HandleFunc("/api/integrity", s.integrity)
	return s, db, auth.protect(s.actorGuard(mux))
}
func mockRequest(h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, "http://127.0.0.1:8787"+path, strings.NewReader(body))
	r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: "owner"})
	r.Header.Set("X-CSRF-Token", "csrf")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}
func TestEventValidationNoWrites(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, kind, payload string
		code                int
	}{
		{"room name", "room.created", `{"name":" "}`, 400},
		{"human rank", "room.created", `{"name":"QA","humanRank":12}`, 400},
		{"protocol", "connection.created", `{"name":"Bad","protocol":"shell","endpoint":"x"}`, 400},
		{"endpoint", "connection.created", `{"name":"Bad","protocol":"CLI"}`, 400},
		{"missing connection", "connection.updated", `{"id":"missing","name":"Agent","protocol":"CLI","endpoint":"codex"}`, 404},
		{"agent rank", "participant.created", `{"name":"Alex","rank":11,"connectionId":"c"}`, 400},
		{"agent challenge", "participant.created", `{"name":"Alex","rank":5,"challenge":5,"connectionId":"c"}`, 400},
		{"memory scope", "memory.revised", `{"title":"Notes","scope":"Public"}`, 400},
		{"stale memory", "memory.revised", `{"id":"m","title":"Notes","scope":"Room","revision":1}`, 409},
		{"foreign memory", "memory.revised", `{"id":"foreign","title":"Notes","scope":"Room","revision":1}`, 404},
		{"stale task", "task.revised", `{"id":"t","title":"Build","status":"Working","revision":1}`, 409},
		{"foreign assignee", "task.revised", `{"title":"Build","status":"Assigned","assignee":"foreign"}`, 400},
		{"task status", "task.revised", `{"title":"Build","status":"Magic"}`, 400},
		{"evidence task", "evidence.created", `{"title":"Test","body":"Pass","taskId":"missing","kind":"Checks"}`, 404},
		{"evidence kind", "evidence.created", `{"title":"Test","body":"Pass","taskId":"t","kind":"Unverified"}`, 400},
		{"immutable rewrite", "message.updated", `{"text":"replace history"}`, 400},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			_, db, h := mockServer(t)
			db.seed("memory.revised", room.Memory{ID: "m", RoomID: "engineering", Title: "Notes", Scope: "Room", Revision: 2})
			db.seed("memory.revised", room.Memory{ID: "foreign", RoomID: "qa", Title: "Private", Scope: "Room", Revision: 1})
			db.seed("task.revised", room.Task{ID: "t", RoomID: "engineering", Title: "Build", Status: "Assigned", Revision: 2})
			db.seed("participant.created", room.Participant{ID: "foreign", RoomID: "qa", Name: "QA", Rank: 5})
			before := len(db.snapshot())
			w := mockRequest(h, "POST", "/api/events", `{"kind":"`+c.kind+`","payload":`+c.payload+`}`)
			if w.Code != c.code {
				t.Fatalf("got %d: %s", w.Code, w.Body.String())
			}
			if len(db.snapshot()) != before {
				t.Fatal("rejected request wrote an event")
			}
		})
	}
}
func TestEventProjectionRoundTrip(t *testing.T) {
	t.Parallel()
	s, db, h := mockServer(t)
	operations := []string{
		`{"kind":"company.updated","payload":{"name":"Example Studio"}}`,
		`{"kind":"human.profile.updated","payload":{"name":"Mark","title":"Owner"}}`,
		`{"kind":"room.created","payload":{"name":"QA"}}`,
		`{"kind":"connection.created","payload":{"name":"Claude","protocol":"CLI","endpoint":"claude"}}`,
		`{"kind":"message.created","payload":{"author":"Impersonated","actorId":"attacker","roomId":"qa","text":"Hello"}}`,
		`{"kind":"memory.revised","payload":{"title":"Design","scope":"Room","body":"First revision"}}`,
		`{"kind":"task.revised","payload":{"title":"Build","status":"Assigned"}}`,
		`{"kind":"workflow.revised","payload":{"name":"Review","maxIterations":3,"nodes":[],"edges":[]}}`,
	}
	for _, body := range operations {
		w := mockRequest(h, "POST", "/api/events", body)
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	st, err := s.load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if st.Company.Name != "Example Studio" || st.HumanName != "Mark" || len(st.Connections) != 1 || len(st.Messages) != 1 || len(st.Memories) != 1 || len(st.Tasks) != 1 || len(st.Workflows) != 1 {
		t.Fatalf("projection lost changes: %+v", st)
	}
	m := st.Messages[0]
	if m.Author != "Mark" || m.ActorID != ownerID || m.ActorType != "human" || m.RoomID != "engineering" {
		t.Fatal("message impersonation or room spoofing accepted", m)
	}
	var qa room.Room
	for _, r := range st.Rooms {
		if r.Name == "QA" {
			qa = r
		}
	}
	if qa.HumanRank == nil || *qa.HumanRank != 11 || qa.HumanPresent == nil || !*qa.HumanPresent {
		t.Fatal("new room human defaults missing")
	}
	memory := st.Memories[0]
	memory.Body = "Second revision"
	payload, _ := json.Marshal(memory)
	w := mockRequest(h, "POST", "/api/events", `{"kind":"memory.revised","payload":`+string(payload)+`}`)
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	st, err = s.load(context.Background())
	if err != nil || len(st.Memories) != 1 || st.Memories[0].Revision != 2 {
		t.Fatal("revision update failed")
	}
	if len(db.snapshot()) != 9 {
		t.Fatal("history was overwritten")
	}
}
func TestAuditReadAndIntegrityFailures(t *testing.T) {
	t.Parallel()
	t.Run("readable and private", func(t *testing.T) {
		t.Parallel()
		s, db, _ := mockServer(t)
		db.seed("message.created", room.Message{Author: "Mark", Text: "private-body"})
		db.seed("connection.created", room.Connection{Name: "Claude", Endpoint: "credential-in-endpoint"})
		w := httptest.NewRecorder()
		s.events(w, httptest.NewRequest("GET", "/api/events", nil))
		var entries []struct {
			Sequence      int
			Summary, Hash string
		}
		if json.Unmarshal(w.Body.Bytes(), &entries) != nil || len(entries) != 2 || entries[0].Sequence != 2 || entries[1].Summary != "Mark sent a room message" {
			t.Fatal(w.Body.String())
		}
		for _, private := range []string{"private-body", "credential-in-endpoint", "payload"} {
			if strings.Contains(w.Body.String(), private) {
				t.Fatal("private payload leaked")
			}
		}
	})
	for _, stage := range []string{"query", "scan", "stream"} {
		t.Run(stage, func(t *testing.T) {
			t.Parallel()
			s, db, _ := mockServer(t)
			db.seed("message.created", room.Message{Text: "Hello"})
			if stage == "stream" {
				db.streamErr = errors.New("interrupted stream")
			} else {
				db.failAt = stage
			}
			for _, handler := range []http.HandlerFunc{s.events, s.integrity} {
				w := httptest.NewRecorder()
				handler(w, httptest.NewRequest("GET", "/api/events", nil))
				if w.Code < 500 {
					t.Fatal("DB failure reported as success", w.Code, w.Body.String())
				}
			}
			if _, err := s.load(context.Background()); err == nil {
				t.Fatal("projection discarded read failure")
			}
		})
	}
	t.Run("tampering", func(t *testing.T) {
		t.Parallel()
		s, db, _ := mockServer(t)
		db.seed("message.created", room.Message{Text: "original"})
		db.events[0].payload = `{"text":"altered"}`
		for i := 0; i < 2; i++ {
			w := httptest.NewRecorder()
			s.integrity(w, httptest.NewRequest("GET", "/api/integrity", nil))
			if w.Code != 200 || !strings.Contains(w.Body.String(), `"valid":false`) {
				t.Fatal("tampering undetected")
			}
		}
		count := 0
		for _, e := range db.snapshot() {
			if e.kind == "incident.opened" {
				count++
			}
		}
		if count != 1 {
			t.Fatal("incident missing or duplicated", count)
		}
	})
	t.Run("valid chain", func(t *testing.T) {
		t.Parallel()
		s, db, _ := mockServer(t)
		db.seed("message.created", room.Message{Text: "Hello"})
		w := httptest.NewRecorder()
		s.integrity(w, httptest.NewRequest("GET", "/api/integrity", nil))
		if !strings.Contains(w.Body.String(), `"valid":true`) || len(db.snapshot()) != 1 {
			t.Fatal("valid chain altered")
		}
	})
}
