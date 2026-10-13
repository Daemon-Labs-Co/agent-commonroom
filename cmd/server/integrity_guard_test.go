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

func TestIntegrityStopSurvivesRestartAndRequiresVerifiedRecovery(t *testing.T) {
	t.Parallel()
	s, _, h := mockServer(t)
	s.integrityStopPath = filepath.Join(t.TempDir(), "integrity-stop")
	s.tripIntegrityLocked()
	info, err := os.Stat(s.integrityStopPath)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatal("private stop marker missing", err)
	}
	restarted := &server{integrityStopPath: s.integrityStopPath}
	restarted.restoreIntegrityStop()
	if !restarted.integrityFault {
		t.Fatal("restart forgot outage")
	}
	s.recoveryIdle = func(context.Context, string) (bool, error) { return false, nil }
	w := mockRequest(h, "POST", "/api/integrity", `{"reason":"Synthetic recovery"}`)
	if w.Code != 503 {
		t.Fatal(w.Code)
	}
	if _, err := os.Stat(s.integrityStopPath); err != nil {
		t.Fatal("unverified recovery removed stop")
	}
	s.recoveryIdle = func(context.Context, string) (bool, error) { return true, nil }
	w = mockRequest(h, "POST", "/api/integrity", `{"reason":"Verified synthetic recovery"}`)
	if w.Code != 200 || s.integrityFault {
		t.Fatal(w.Code)
	}
	if _, err := os.Stat(s.integrityStopPath); !os.IsNotExist(err) {
		t.Fatal("verified recovery retained marker", err)
	}
}

func TestIntegrityOutageCancelsJobsAndDoesNotAutomaticallyResume(t *testing.T) {
	t.Parallel()
	for _, failure := range []string{"query", "scan", "stream"} {
		t.Run(failure, func(t *testing.T) {
			t.Parallel()
			s, db, h := mockServer(t)
			db.seed("message.created", room.Message{Text: "Synthetic"})
			model, cancelModel := context.WithCancel(context.Background())
			defer cancelModel()
			flow, cancelFlow := context.WithCancel(context.Background())
			defer cancelFlow()
			s.cancels["p"] = cancelModel
			s.workflowLive = map[string]context.CancelFunc{"run": cancelFlow}
			if failure == "stream" {
				db.streamErr = errors.New("synthetic unavailable")
			} else {
				db.failAt = failure
			}
			s.checkIntegrityNow()
			if !s.integrityFault || model.Err() == nil || flow.Err() == nil {
				t.Fatal("integrity did not cancel execution")
			}
			db.failAt = ""
			db.streamErr = nil
			s.checkIntegrityNow()
			st, _ := s.load(context.Background())
			if !st.Incident || !s.agentsBlockedLocked(st, "engineering") {
				t.Fatal("storage restoration resumed agents")
			}
			s.recoveryIdle = func(context.Context, string) (bool, error) { return true, nil }
			w := mockRequest(h, "POST", "/api/integrity", `{"reason":"Inspected restored synthetic audit and verified worker shutdown"}`)
			if w.Code != 200 || s.integrityFault {
				t.Fatal(w.Code, w.Body.String())
			}
			st, _ = s.load(context.Background())
			if st.Incident {
				t.Fatal("approved recovery not projected")
			}
		})
	}
}

func TestIntegrityTamperingOrFailedRecoveryCannotClearLatch(t *testing.T) {
	t.Parallel()
	s, db, h := mockServer(t)
	db.seed("message.created", room.Message{Text: "Original synthetic"})
	db.events[0].payload = `{"text":"Altered synthetic"}`
	s.recoveryIdle = func(context.Context, string) (bool, error) {
		t.Error("worker checked before audit verified")
		return true, nil
	}
	w := mockRequest(h, "POST", "/api/integrity", `{"reason":"Attempted recovery"}`)
	if w.Code != 409 || !s.integrityFault {
		t.Fatal("invalid audit accepted", w.Code)
	}
	fresh, db2, h2 := mockServer(t)
	fresh.tripIntegrityLocked()
	fresh.recoveryIdle = func(context.Context, string) (bool, error) { return false, nil }
	w = mockRequest(h2, "POST", "/api/integrity", `{"reason":"Synthetic shutdown check"}`)
	if w.Code != 503 || !fresh.integrityFault {
		t.Fatal("busy worker cleared stop")
	}
	db2.failAt = "commit"
	fresh.recoveryIdle = func(context.Context, string) (bool, error) { return true, nil }
	w = mockRequest(h2, "POST", "/api/integrity", `{"reason":"Synthetic audit commit failure"}`)
	if w.Code < 400 || !fresh.integrityFault {
		t.Fatal("failed audit cleared latch")
	}
}

func TestNonHumanAuthenticatedActorCannotReadOrMutateAPI(t *testing.T) {
	t.Parallel()
	s, db, h := mockServer(t)
	db.seed("actor.revised", room.Actor{ID: "synthetic-agent", Kind: "agent", Role: "Developer", Enabled: true, RoomIDs: []string{"engineering"}})
	s.auth.sessions["owner"] = ownerSession{ActorID: "synthetic-agent", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	for _, method := range []string{"GET", "POST"} {
		w := mockRequest(h, method, "/api/state", "")
		if w.Code != 403 {
			t.Fatal(method, w.Code)
		}
	}
}

func TestIntegrityRecoveryRequiresOwnerReasonAndCSRF(t *testing.T) {
	t.Parallel()
	s, _, h := mockServer(t)
	w := mockRequest(h, "POST", "/api/integrity", `{"reason":""}`)
	if w.Code != 400 {
		t.Fatal(w.Code)
	}
	r := httptest.NewRequest("POST", "http://127.0.0.1:8787/api/integrity", strings.NewReader(`{"reason":"Synthetic"}`))
	r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: "owner"})
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 403 || s.integrityFault {
		t.Fatal("csrf accepted or caused incident")
	}
}
