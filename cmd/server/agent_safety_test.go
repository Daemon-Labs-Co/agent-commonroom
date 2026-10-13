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

func safetyHarness(t *testing.T) (*server, *mockDB, http.Handler) {
	s, db, _ := mockServer(t)
	s.recoveryIdle = func(context.Context, string) (bool, error) { return true, nil }
	mux := http.NewServeMux()
	mux.HandleFunc("/api/agent-safety", s.agentSafety)
	mux.HandleFunc("/api/agents", s.agents)
	return s, db, s.auth.protect(s.actorGuard(mux))
}

func TestSafetyResumeRefusesUnknownWorkerOrIntegrityStop(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"busy", "unknown", "active", "integrity"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			s, db, h := safetyHarness(t)
			s.stopAll = true
			s.recoveryIdle = func(context.Context, string) (bool, error) { return false, nil }
			if mode == "unknown" {
				s.recoveryIdle = func(context.Context, string) (bool, error) { return false, context.DeadlineExceeded }
			}
			if mode == "active" {
				s.active["synthetic"] = true
			}
			if mode == "integrity" {
				s.integrityFault = true
			}
			w := mockRequest(h, "POST", "/api/agent-safety", `{"action":"resume","scope":"application"}`)
			if w.Code < 400 || !s.stopAll || len(db.snapshot()) != 0 {
				t.Fatal("unverified resume accepted", mode, w.Code)
			}
		})
	}
}

func TestApplicationStopCancelsWorkflowEvenWhenDatabaseReadFails(t *testing.T) {
	t.Parallel()
	s, db, h := safetyHarness(t)
	db.failAt = "query"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.workflowLive = map[string]context.CancelFunc{"synthetic": cancel}
	w := mockRequest(h, "POST", "/api/agent-safety", `{"action":"stop","scope":"application"}`)
	if w.Code != 200 || ctx.Err() == nil {
		t.Fatal("application stop missed workflow", w.Code)
	}
}
func TestEmergencyStopAndExplicitResume(t *testing.T) {
	t.Parallel()
	s, db, h := safetyHarness(t)
	db.seed("room.created", room.Room{ID: "qa", Name: "QA"})
	db.seed("participant.created", room.Participant{ID: "dev", Name: "Builder", RoomID: "engineering"})
	db.seed("participant.created", room.Participant{ID: "qa", Name: "Reviewer", RoomID: "qa"})
	dev, cancelDev := context.WithCancel(context.Background())
	qa, cancelQA := context.WithCancel(context.Background())
	defer cancelDev()
	defer cancelQA()
	s.cancels["dev"] = cancelDev
	s.cancels["qa"] = cancelQA
	w := mockRequest(h, "POST", "/api/agent-safety", `{"action":"stop","scope":"room"}`)
	if w.Code != 200 || dev.Err() != context.Canceled || qa.Err() != nil {
		t.Fatal("room stop cancellation incorrect", w.Code)
	}
	st, err := s.load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	fresh := server{db: db}
	if !fresh.agentsBlockedLocked(st, "engineering") || fresh.agentsBlockedLocked(st, "qa") {
		t.Fatal("persisted scope incorrect")
	}
	w = mockRequest(h, "POST", "/api/agents", `{"action":"send","text":"Hello"}`)
	if w.Code != 409 {
		t.Fatal("stop allowed dispatch")
	}
	w = mockRequest(h, "POST", "/api/agent-safety", `{"action":"stop","scope":"application"}`)
	if w.Code != 200 || qa.Err() != context.Canceled {
		t.Fatal("global stop did not cancel all")
	}
	w = mockRequest(h, "POST", "/api/agent-safety", `{"action":"resume","scope":"room"}`)
	if w.Code != 409 {
		t.Fatal("room bypassed global stop")
	}
	w = mockRequest(h, "POST", "/api/agent-safety", `{"action":"resume","scope":"application"}`)
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	st, err = s.load(context.Background())
	if err != nil || !s.agentsBlockedLocked(st, "engineering") || s.agentsBlockedLocked(st, "qa") {
		t.Fatal("global resume cleared room stop")
	}
	w = mockRequest(h, "POST", "/api/agent-safety", `{"action":"resume","scope":"room"}`)
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	st, err = s.load(context.Background())
	if err != nil || s.agentsBlockedLocked(st, "engineering") {
		t.Fatal("explicit room resume failed")
	}
}
func TestEmergencyStopFailureAndAccess(t *testing.T) {
	t.Parallel()
	t.Run("database read outage", func(t *testing.T) {
		t.Parallel()
		s, db, h := safetyHarness(t)
		db.failAt = "query"
		ctx, cancel := context.WithCancel(context.Background())
		s.cancels["p"] = cancel
		w := mockRequest(h, "POST", "/api/agent-safety", `{"action":"stop","scope":"application"}`)
		if w.Code != 200 || ctx.Err() != context.Canceled || !s.stopAll {
			t.Fatal("read outage prevented emergency stop")
		}
	})
	t.Run("audit failure", func(t *testing.T) {
		t.Parallel()
		s, db, h := safetyHarness(t)
		db.failAt = "insert"
		ctx, cancel := context.WithCancel(context.Background())
		s.cancels["p"] = cancel
		w := mockRequest(h, "POST", "/api/agent-safety", `{"action":"stop","scope":"application"}`)
		if w.Code != 503 || ctx.Err() != context.Canceled || !s.stopAll || len(db.snapshot()) != 0 {
			t.Fatal("audit failure did not preserve local stop")
		}
	})
	t.Run("member denied", func(t *testing.T) {
		t.Parallel()
		s, db, h := safetyHarness(t)
		db.seed("actor.revised", room.Actor{ID: "developer", Name: "Dev", Kind: "human", Role: "Developer", Enabled: true})
		s.auth.sessions["owner"] = ownerSession{ActorID: "developer", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
		w := mockRequest(h, "POST", "/api/agent-safety", `{"action":"stop","scope":"application"}`)
		if w.Code != 403 || s.stopAll {
			t.Fatal("member controlled stop")
		}
	})
	t.Run("csrf denied", func(t *testing.T) {
		t.Parallel()
		s, _, h := safetyHarness(t)
		r := httptest.NewRequest("POST", "http://127.0.0.1:8787/api/agent-safety", strings.NewReader(`{"action":"stop","scope":"application"}`))
		r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: "owner"})
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 403 || s.stopAll {
			t.Fatal("missing csrf accepted")
		}
	})
	t.Run("incident cannot resume", func(t *testing.T) {
		t.Parallel()
		s, db, h := safetyHarness(t)
		s.stopAll = true
		db.seed("incident.opened", map[string]string{"reason": "mismatch"})
		w := mockRequest(h, "POST", "/api/agent-safety", `{"action":"resume","scope":"application"}`)
		if w.Code != 409 || !s.stopAll {
			t.Fatal("resume bypassed incident")
		}
	})
}
func TestCancelledTurnCannotPostAfterResume(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	p := room.Participant{ID: "p", Name: "Agent", RoomID: "engineering"}
	db.seed("participant.created", p)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	s.finishAgentTurn(ctx, p, "late successful result", "codex", nil)
	st, err := s.load(context.Background())
	if err != nil || len(st.Messages) != 0 || st.Participants[0].Status != "Stopped" {
		t.Fatal("cancelled result leaked after resume")
	}
}
func TestEmergencyUIFlagDoesNotDisableBackend(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "agent-safety.js"), []byte("/* controls */"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, enabled := range []bool{true, false} {
		w := httptest.NewRecorder()
		optionalScriptHandler(enabled, root, "agent-safety.js").ServeHTTP(w, httptest.NewRequest("GET", "/agent-safety.js", nil))
		if enabled && w.Code != 200 || !enabled && w.Code != 204 {
			t.Fatal("UI flag failed")
		}
		s, _, h := safetyHarness(t)
		w = mockRequest(h, "POST", "/api/agent-safety", `{"action":"stop","scope":"application"}`)
		if w.Code != 200 || !s.stopAll {
			t.Fatal("UI flag disabled backend safety")
		}
	}
}

func TestSafetyResumeVerifiesChainBeforeWorkerOrApproval(t *testing.T) {
	t.Parallel()
	s, db, h := safetyHarness(t)
	db.seed("message.created", room.Message{Text: "Synthetic original"})
	db.events[0].payload = `{"text":"Synthetic altered"}`
	s.stopAll = true
	s.recoveryIdle = func(context.Context, string) (bool, error) {
		t.Error("worker probed before audit verification")
		return true, nil
	}
	w := mockRequest(h, "POST", "/api/agent-safety", `{"action":"resume","scope":"application"}`)
	if w.Code != 409 || !s.stopAll || !s.integrityFault {
		t.Fatal("tampered chain resumed", w.Code)
	}
}
