package main

import (
	"agentroom/internal/room"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func recoveryServer(t *testing.T) (*server, *mockDB, http.Handler) {
	s, db, _ := mockServer(t)
	db.seed("connection.created", room.Connection{ID: "c", Protocol: "CLI", Endpoint: "codex"})
	db.seed("participant.created", room.Participant{ID: "p", RoomID: "engineering", ConnectionID: "c", Status: "Blocked", ThreadID: "synthetic-session", ExecutionMode: "container-v1"})
	s.recoveryIdle = func(context.Context, string) (bool, error) { return true, nil }
	mux := http.NewServeMux()
	mux.HandleFunc("/api/agent-recovery", s.agentRecovery)
	return s, db, s.auth.protect(s.actorGuard(mux))
}

func TestRecoveryPreservesOrClearsOnlySessionBinding(t *testing.T) {
	t.Parallel()
	for _, action := range []string{"check", "resume", "fresh"} {
		t.Run(action, func(t *testing.T) {
			t.Parallel()
			s, db, h := recoveryServer(t)
			before := len(db.snapshot())
			w := mockRequest(h, "POST", "/api/agent-recovery", `{"action":"`+action+`","participantId":"p"}`)
			if w.Code != 200 {
				t.Fatal(w.Code, w.Body.String())
			}
			st, _ := s.load(context.Background())
			p := st.Participants[0]
			if action == "check" {
				if p.Status != "Blocked" || len(db.snapshot()) != before {
					t.Fatal("check mutated history")
				}
			} else if p.Status != "Idle" {
				t.Fatal(p)
			}
			if action == "fresh" {
				if p.ThreadID != "" || p.ExecutionMode != "" {
					t.Fatal(p)
				}
			} else if p.ThreadID != "synthetic-session" {
				t.Fatal("saved context lost")
			}
			if len(st.Messages) != 0 || len(s.active) != 0 {
				t.Fatal("recovery launched or replayed work")
			}
		})
	}
}

func TestRecoveryFailurePathsLeaveSessionUntouched(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"busy", "unknown", "active", "stop", "flag", "wrong room", "legacy"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			s, db, h := recoveryServer(t)
			path := "/api/agent-recovery"
			code := 409
			switch name {
			case "busy":
				s.recoveryIdle = func(context.Context, string) (bool, error) { return false, nil }
				code = 200
			case "unknown":
				s.recoveryIdle = func(context.Context, string) (bool, error) { return false, errors.New("synthetic transport failure") }
				code = 503
			case "active":
				s.active["other"] = true
			case "stop":
				s.stopAll = true
			case "flag":
				db.seed("feature.flag.updated", map[string]any{"name": "agentRecovery", "enabled": false})
				code = 403
			case "wrong room":
				path += "?room=other"
				code = 404
			case "legacy":
				db.seed("participant.session", room.Participant{ID: "p", RoomID: "engineering", ConnectionID: "c", Status: "Blocked", ThreadID: "synthetic-session"})
			}
			before := len(db.snapshot())
			w := mockRequest(h, "POST", path, `{"action":"resume","participantId":"p"}`)
			if w.Code != code || len(db.snapshot()) != before {
				t.Fatal(name, w.Code, w.Body.String())
			}
		})
	}
}

func TestWorkerIdleScanFailsClosed(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	job := strings.Repeat("a", 32)
	if err := writeWorkerRecord(dir, workerJobRecord{JobID: job, Status: "starting"}, true); err != nil {
		t.Fatal(err)
	}
	if idle, _ := workerRecordsIdle(dir); idle {
		t.Fatal("nonterminal job ignored")
	}
	if err := writeWorkerRecord(dir, workerJobRecord{JobID: job, Status: "cancelled"}, false); err != nil {
		t.Fatal(err)
	}
	if idle, _ := workerRecordsIdle(dir); idle {
		t.Fatal("legacy terminal label accepted without shutdown proof")
	}
	confirmed := fakeWorkerRecord(t, job, "room/agent/connection/worker")
	confirmed.Status, confirmed.ShutdownConfirmed = "cancelled", true
	if err := writeWorkerRecord(dir, confirmed, false); err != nil {
		t.Fatal(err)
	}
	if idle, err := workerRecordsIdle(dir); !idle || err != nil {
		t.Fatal(idle, err)
	}
	if err := os.WriteFile(filepath.Join(dir, job+".json"), []byte("invalid"), 0600); err != nil {
		t.Fatal(err)
	}
	if idle, err := workerRecordsIdle(dir); idle || err == nil {
		t.Fatal("unknown record accepted")
	}
}

func TestRecoveryAuthenticationAndCSRF(t *testing.T) {
	t.Parallel()
	s, db, h := recoveryServer(t)
	for _, loggedIn := range []bool{false, true} {
		r := httptest.NewRequest("POST", "http://127.0.0.1:8787/api/agent-recovery", strings.NewReader(`{"action":"fresh","participantId":"p"}`))
		if loggedIn {
			r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: "owner"})
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		expected := 401
		if loggedIn {
			expected = 403
		}
		if w.Code != expected {
			t.Fatal(w.Code)
		}
	}
	db.seed("actor.revised", room.Actor{ID: "developer", Name: "Developer", Kind: "human", Role: "Developer", Enabled: true, RoomIDs: []string{"engineering"}})
	s.auth.sessions["owner"] = ownerSession{ActorID: "developer", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	w := mockRequest(h, "POST", "/api/agent-recovery", `{"action":"fresh","participantId":"p"}`)
	if w.Code != 403 {
		t.Fatal("non-owner recovered", w.Code)
	}
}

func TestRecoveryAuditFailureDoesNotResetSession(t *testing.T) {
	t.Parallel()
	s, db, h := recoveryServer(t)
	db.failAt = "commit"
	before := len(db.snapshot())
	w := mockRequest(h, "POST", "/api/agent-recovery", `{"action":"fresh","participantId":"p"}`)
	if w.Code != 503 || len(db.snapshot()) != before {
		t.Fatal("failed commit changed recovery", w.Code)
	}
	db.failAt = ""
	st, _ := s.load(context.Background())
	if st.Participants[0].Status != "Blocked" || st.Participants[0].ThreadID != "synthetic-session" {
		t.Fatal("failed audit reset binding")
	}
}
