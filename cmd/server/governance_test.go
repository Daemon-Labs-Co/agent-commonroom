package main

import (
	"agentroom/internal/room"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func governanceFixture(t *testing.T, enabled bool) (*server, *mockDB, http.Handler) {
	t.Helper()
	s, db, _ := mockServer(t)
	db.seed("feature.flag.updated", map[string]any{"name": "agentCoordination", "enabled": enabled})
	db.seed("connection.created", room.Connection{ID: "c", Protocol: "CLI", Endpoint: "codex"})
	for _, pid := range []string{"a", "b"} {
		db.seed("participant.created", room.Participant{ID: pid, RoomID: "engineering", ConnectionID: "c", Status: "Idle"})
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/governance", s.governance)
	mux.HandleFunc("/api/events", s.events)
	return s, db, s.auth.protect(s.actorGuard(mux))
}
func governanceRequest(h http.Handler, actorCookie, key, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest("POST", "http://127.0.0.1:8787/api/governance", strings.NewReader(body))
	r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: actorCookie})
	r.Header.Set("X-CSRF-Token", "csrf")
	r.Header.Set("Idempotency-Key", key)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}
func TestGovernanceTransferReplayAndRestart(t *testing.T) {
	t.Parallel()
	s, db, h := governanceFixture(t, true)
	for _, step := range []struct{ key, body string }{{"appoint-a", `{"participantId":"a","revision":0}`}, {"appoint-b", `{"participantId":"b","revision":1}`}} {
		w := governanceRequest(h, "owner", step.key, step.body)
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		count := len(db.snapshot())
		w = governanceRequest(h, "owner", step.key, step.body)
		if w.Code != 200 || w.Header().Get("Idempotent-Replay") != "true" || len(db.snapshot()) != count {
			t.Fatal("replay transferred twice", w.Code)
		}
	}
	st, err := (&server{db: db}).load(t.Context())
	if err != nil || st.Governance["engineering"].ParticipantID != "b" || st.Governance["engineering"].Revision != 2 {
		t.Fatal("restart lost authority", err)
	}
	// Appointments must not mutate actor privileges or issue tool/job grants.
	if len(st.ManagedJobs) != 0 || len(st.ToolAudits) != 0 {
		t.Fatal("appointment granted execution")
	}
	s.auth.sessions["agent"] = ownerSession{ActorID: "agent", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	db.seed("actor.created", room.Actor{ID: "agent", Kind: "agent", Enabled: true})
	count := len(db.snapshot())
	if w := governanceRequest(h, "agent", "forged-op", `{"participantId":"a","revision":2}`); w.Code != 403 || len(db.snapshot()) != count {
		t.Fatal("agent usurped governance", w.Code)
	}
}
func TestGovernanceRefusalsAndAtomicFailure(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, body, key string
		enabled         bool
		code            int
	}{
		{"off", `{"participantId":"a","revision":0}`, "valid-key", false, 403},
		{"foreign", `{"participantId":"foreign","revision":0}`, "valid-key", true, 400},
		{"stale", `{"participantId":"a","revision":1}`, "valid-key", true, 409},
		{"missing-operation", `{"participantId":"a","revision":0}`, "", true, 400},
		{"forged-approval", `{"participantId":"a","revision":0,"approvedBy":"local-owner"}`, "valid-key", true, 400},
		{"missing-revision", `{"participantId":"a"}`, "valid-key", true, 400},
		{"duplicate", `{"participantId":"a","participantId":"b","revision":0}`, "valid-key", true, 400},
		{"null-revision", `{"participantId":"a","revision":null}`, "valid-key", true, 400},
		{"alias", `{"ParticipantId":"a","revision":0}`, "valid-key", true, 400},
		{"trailing", `{"participantId":"a","revision":0} {}`, "valid-key", true, 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, db, h := governanceFixture(t, tc.enabled)
			count := len(db.snapshot())
			w := governanceRequest(h, "owner", tc.key, tc.body)
			if w.Code != tc.code || len(db.snapshot()) != count {
				t.Fatal(w.Code, w.Body.String())
			}
		})
	}
	t.Run("audit-failure", func(t *testing.T) {
		s, db, h := governanceFixture(t, true)
		db.failAt = "insert"
		count := len(db.snapshot())
		w := governanceRequest(h, "owner", "valid-key", `{"participantId":"a","revision":0}`)
		if w.Code != 500 || len(db.snapshot()) != count {
			t.Fatal("partial appointment", w.Code)
		}
		db.failAt = ""
		st, err := s.load(t.Context())
		if err != nil || len(st.Governance) != 0 {
			t.Fatal("failed appointment survived", err)
		}
	})
	t.Run("stop", func(t *testing.T) {
		s, db, h := governanceFixture(t, true)
		s.stopAll = true
		count := len(db.snapshot())
		if w := governanceRequest(h, "owner", "valid-key", `{"participantId":"a","revision":0}`); w.Code != 409 || len(db.snapshot()) != count {
			t.Fatal("stop bypass", w.Code)
		}
	})
}
func TestGovernanceFirstRegistrationAndNoAutomaticReplacement(t *testing.T) {
	t.Parallel()
	s, db, h := governanceFixture(t, true)
	// Seeded legacy profiles do not acquire an appointment retroactively.
	for i := 0; i < 2; i++ {
		w := mockRequest(h, "POST", "/api/events", `{"kind":"participant.created","payload":{"name":"New","rank":1,"connectionId":"c"}}`)
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	st, err := s.load(t.Context())
	if err != nil || st.Governance["engineering"].Revision != 1 {
		t.Fatal(err, st.Governance)
	}
	initial := st.Governance["engineering"]
	db.seed("participant.removed", map[string]string{"id": initial.ParticipantID, "roomId": "engineering"})
	w := mockRequest(h, "GET", "/api/governance", "")
	var view struct {
		Available   bool                       `json:"available"`
		Appointment room.GovernanceAppointment `json:"appointment"`
	}
	json.Unmarshal(w.Body.Bytes(), &view)
	if w.Code != 200 || view.Available || view.Appointment != initial {
		t.Fatal("silent replacement", w.Body.String())
	}
	// Turning off retains records and denies changes without disabling human stop.
	db.seed("feature.flag.updated", map[string]any{"name": "agentCoordination", "enabled": false})
	st, err = s.load(t.Context())
	if err != nil || st.Governance["engineering"] != initial {
		t.Fatal("history lost", err)
	}
}
func TestGovernanceConcurrentTransfers(t *testing.T) {
	t.Parallel()
	_, db, h := governanceFixture(t, true)
	var wg sync.WaitGroup
	codes := make(chan int, 2)
	for _, pid := range []string{"a", "b"} {
		wg.Add(1)
		go func(pid string) {
			defer wg.Done()
			codes <- governanceRequest(h, "owner", "appoint-"+pid, `{"participantId":"`+pid+`","revision":0}`).Code
		}(pid)
	}
	wg.Wait()
	close(codes)
	counts := map[int]int{}
	for c := range codes {
		counts[c]++
	}
	if counts[200] != 1 || counts[409] != 1 {
		t.Fatal(counts)
	}
	st, err := (&server{db: db}).load(context.Background())
	if err != nil || st.Governance["engineering"].Revision != 1 {
		t.Fatal(err)
	}
}
func TestGovernanceProjectionRefusesForgedAuthority(t *testing.T) {
	t.Parallel()
	s, _, _ := governanceFixture(t, true)
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	for _, actor := range []string{"", "agent"} {
		raw, _ := json.Marshal(map[string]any{"companyId": st.Company.ID, "roomId": "engineering", "participantId": "a", "revision": 1, "performedBy": actor})
		if projectGovernance(&st, raw) == nil {
			t.Fatal("forged audit authority accepted")
		}
	}
}

func TestAutomaticAppointmentFailureLatchesStop(t *testing.T) {
	t.Parallel()
	for _, failure := range []string{"begin", "insert", "commit"} {
		t.Run(failure, func(t *testing.T) {
			s, db, h := governanceFixture(t, true)
			before := len(db.snapshot())
			db.failAt = failure
			w := mockRequest(h, "POST", "/api/events", `{"kind":"participant.created","payload":{"name":"New","rank":1,"connectionId":"c"}}`)
			if w.Code != 500 || !s.integrityFault || len(db.snapshot()) != before {
				t.Fatal("failed registration did not latch atomically", w.Code, s.integrityFault)
			}
			db.failAt = ""
			if w = governanceRequest(h, "owner", "retry-after-fault", `{"participantId":"a","revision":0}`); w.Code != 409 {
				t.Fatal("fault allowed later appointment", w.Code)
			}
		})
	}
}

func TestAppointmentBatchesStopOnFailureButNotConflict(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"create", "attach", "conflict"} {
		t.Run(mode, func(t *testing.T) {
			s, db, _ := governanceFixture(t, true)
			st, err := s.load(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			p := room.Participant{ID: "new", RoomID: "engineering", ConnectionID: "fresh", Status: "Idle"}
			c := room.Connection{ID: "fresh", Protocol: "CLI", Endpoint: "claude"}
			changes := []auditChange{{Kind: "connection.created", Payload: c}, {Kind: "participant.created", Payload: p}}
			if mode == "attach" {
				changes = append(changes, auditChange{Kind: "participant.history_attached", Payload: map[string]string{"participantId": p.ID}})
			}
			changes = append(changes, initialGovernanceChanges(st, p, c)...)
			ctx := context.WithValue(t.Context(), actorContextKey{}, ownerID)
			if mode == "conflict" {
				ctx = context.WithValue(ctx, expectedEventCountKey{}, st.EventCount-1)
			} else {
				db.failAt = "commit"
			}
			before := len(db.snapshot())
			s.mu.Lock()
			err = s.appendAppointmentBatch(ctx, changes)
			s.mu.Unlock()
			if err == nil || len(db.snapshot()) != before || s.integrityFault != (mode != "conflict") {
				t.Fatal("incorrect failure/stop policy", err, s.integrityFault)
			}
		})
	}
}
