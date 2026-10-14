package main

import (
	"agentroom/internal/room"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func verificationFixture(t *testing.T) (*server, *mockDB, room.Connection) {
	t.Helper()
	s, db, _ := mockServer(t)
	c := room.Connection{ID: "synthetic-connection", Name: "Synthetic", Protocol: "CLI", Provider: "openai", Endpoint: "ssh://worker", Revision: 1}
	db.seed("connection.created", c)
	return s, db, c
}
func journalVerification(ctx context.Context, status string) error {
	r := ctx.Value(managedJobContextKey{}).(managedJobRegistration)
	if err := r.Register("worker", "codex"); err != nil {
		return err
	}
	if err := r.Update("submitted"); err != nil {
		return err
	}
	return r.Update(status)
}
func TestManagedVerificationRequiresIntentAndShutdownProof(t *testing.T) {
	t.Setenv("COMMONROOM_REMOTE_WORKER_ALIAS", "worker")
	t.Setenv("COMMONROOM_REMOTE_WORKER", "true")
	for _, kind := range []string{"success", "wrong reply", "provider failed", "no journal", "uncertain", "intent failed", "flag off", "stopped", "capacity", "cancelled", "changed connection"} {
		t.Run(kind, func(t *testing.T) {
			s, db, c := verificationFixture(t)
			called := false
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			switch kind {
			case "flag off":
				db.seed("feature.flag.updated", map[string]any{"name": "connectionTesting", "enabled": false})
			case "stopped":
				s.stopAll = true
			case "capacity":
				for i := 0; i < 8; i++ {
					s.active[string(rune('a'+i))] = true
				}
			case "cancelled":
				cancel()
			}
			s.verificationTurn = func(ctx context.Context) (string, error) {
				called = true
				if kind == "no journal" {
					return "COMMONROOM_CONNECTION_OK", nil
				}
				if kind == "intent failed" {
					db.failAt = "insert"
				}
				if kind == "changed connection" {
					changed := c
					changed.Revision++
					db.seed("connection.updated", changed)
				}
				status := "shutdown_confirmed"
				if kind == "uncertain" {
					status = "shutdown_unconfirmed"
				}
				if err := journalVerification(ctx, status); err != nil {
					return "", err
				}
				if kind == "wrong reply" {
					return "some other result", nil
				}
				if kind == "provider failed" {
					return "", errors.New("synthetic failure")
				}
				return "COMMONROOM_CONNECTION_OK", nil
			}
			raw, err := s.managedVerification(ctx, c, "engineering")
			if kind == "success" || kind == "wrong reply" {
				var v struct {
					Verified bool `json:"verified"`
				}
				if err != nil || json.Unmarshal(raw, &v) != nil || v.Verified != (kind == "success") {
					t.Fatal(kind, err, string(raw))
				}
			} else if err == nil {
				t.Fatal("unsafe verification accepted", kind)
			}
			if (kind == "flag off" || kind == "stopped" || kind == "capacity" || kind == "cancelled") && called {
				t.Fatal("blocked check launched provider")
			}
			if kind == "intent failed" && !s.integrityFault {
				t.Fatal("audit outage did not latch stop")
			}
			if kind == "uncertain" && !s.stopAll {
				t.Fatal("unconfirmed shutdown did not stop agents")
			}
			if s.active[verificationParticipant("engineering", c.ID).ID] || len(s.verificationRooms) != 0 {
				t.Fatal("verification reservation leaked")
			}
		})
	}
}
func TestVerificationCanBeCancelledAndBlocksConnectionChanges(t *testing.T) {
	t.Setenv("COMMONROOM_REMOTE_WORKER_ALIAS", "worker")
	t.Setenv("COMMONROOM_REMOTE_WORKER", "true")
	s, db, c := verificationFixture(t)
	started := make(chan struct{})
	done := make(chan error, 1)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	s.verificationTurn = func(ctx context.Context) (string, error) {
		r := ctx.Value(managedJobContextKey{}).(managedJobRegistration)
		if err := r.Register("worker", "codex"); err != nil {
			return "", err
		}
		close(started)
		<-ctx.Done()
		r.Update("shutdown_confirmed")
		return "", ctx.Err()
	}
	go func() { _, err := s.managedVerification(ctx, c, "engineering"); done <- err }()
	select {
	case <-started:
	case <-ctx.Done():
		t.Fatal("verification did not start")
	}
	s.mu.Lock()
	st, err := s.load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !connectionHasActiveAgent(st, s.active, c.ID) {
		t.Fatal("verification invisible to credential controls")
	}
	// Both edit and delete paths must reject the controller reservation.
	for _, kind := range []string{"connection.updated", "connection.deleted"} {
		raw, _ := json.Marshal(c)
		w := httptest.NewRecorder()
		_, ok := s.prepareAgentsEvent(w, httptest.NewRequest("POST", "/api/events", nil), st, "engineering", eventRequest{Kind: kind, Payload: raw})
		if ok || w.Code != 409 {
			t.Fatal("connection mutation allowed", kind, w.Code)
		}
	}
	s.stopAll = true
	for _, stop := range s.cancels {
		stop()
	}
	s.mu.Unlock()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("cancelled verification claimed working")
		}
	case <-time.After(time.Second):
		t.Fatal("verification ignored cancellation")
	}
	_ = db
}
func TestLegacyVerificationLauncherCannotRun(t *testing.T) {
	t.Parallel()
	var out, log bytes.Buffer
	if workerVerifyConnection(workerAuthRequest{}, &out, &log) != 1 || out.Len() != 0 {
		t.Fatal("legacy unjournalled launch allowed")
	}
}

func TestConnectionTestHTTPUsesJournalledVerification(t *testing.T) {
	t.Setenv("COMMONROOM_REMOTE_WORKER_ALIAS", "worker")
	t.Setenv("COMMONROOM_REMOTE_WORKER", "true")
	s, db, c := verificationFixture(t)
	calls := 0
	s.connectionProbe = func(_ context.Context, _, action, _ string) ([]byte, error) {
		calls++
		switch action {
		case "ready-v1":
			return []byte(`{"version":1,"lifecycleVersion":2,"configured":true,"runtimeReady":true}`), nil
		case "auth-status-v1":
			return []byte(`{"credentialPresent":true}`), nil
		}
		t.Fatal("legacy provider probe invoked", action)
		return nil, nil
	}
	s.verificationTurn = func(ctx context.Context) (string, error) {
		return "COMMONROOM_CONNECTION_OK", journalVerification(ctx, "shutdown_confirmed")
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/connections/test", s.connectionTest)
	h := s.auth.protect(s.actorGuard(mux))
	w := mockRequest(h, "POST", "/api/connections/test?room=engineering", `{"connectionId":"`+c.ID+`"}`)
	var result connectionTestResult
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &result) != nil || result.Status != "working" || calls != 2 {
		t.Fatal(w.Code, w.Body.String(), calls)
	}
	st, err := s.load(context.Background())
	if err != nil || len(st.ManagedJobs) != 1 {
		t.Fatal("verification omitted journal", err)
	}
	for _, job := range st.ManagedJobs {
		if !job.ShutdownConfirmed || job.ConnectionID != c.ID || job.RoomID != "engineering" {
			t.Fatal("incorrect verification job", job)
		}
	}
	_ = db
}

func TestVerificationReservationsConsumeRoomCapacity(t *testing.T) {
	t.Parallel()
	s, _, _ := verificationFixture(t)
	s.verificationRooms = map[string]string{}
	for i := 0; i < 4; i++ {
		pid := string(rune('a' + i))
		s.active[pid] = true
		s.verificationRooms[pid] = "engineering"
	}
	if err := s.agentAdmission(room.State{}, "engineering", 1); err == nil {
		t.Fatal("verification bypassed room capacity")
	}
	if err := s.agentAdmission(room.State{}, "other", 1); err != nil {
		t.Fatal("other room incorrectly counted", err)
	}
}
