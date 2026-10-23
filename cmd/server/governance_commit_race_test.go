package main

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

// Commits through an independent controller immediately before the original
// controller's append transaction starts. No goroutine timing or live DB needed.
type interleavedEventDB struct {
	eventDatabase
	beforeBegin func(context.Context)
}

func (d *interleavedEventDB) Begin(ctx context.Context) (pgx.Tx, error) {
	d.beforeBegin(ctx)
	return d.eventDatabase.Begin(ctx)
}

func TestCoordinationWritesRevalidateAfterConcurrentCommit(t *testing.T) {
	t.Parallel()
	for _, path := range []string{"result", "session", "submitted"} {
		for _, mutation := range []string{"transfer", "disable", "stop", "unrelated", "continued-conflicts"} {
			t.Run(path+"/"+mutation, func(t *testing.T) {
				t.Parallel()
				s, db, _, p, appointment := coordinationFixture(t)
				ctx := coordinatedContext(appointment)
				var registration managedJobRegistration
				if path == "submitted" {
					registration = s.managedJobContext(ctx, p).Value(managedJobContextKey{}).(managedJobRegistration)
					if err := registration.Register("worker", "codex"); err != nil {
						t.Fatal(err)
					}
				}
				other := &server{db: db, auth: s.auth}
				mux := http.NewServeMux()
				mux.HandleFunc("/api/governance", other.governance)
				mux.HandleFunc("/api/events", other.events)
				mux.HandleFunc("/api/agent-safety", other.agentSafety)
				h := other.auth.protect(other.actorGuard(mux))
				begins := 0
				s.db = &interleavedEventDB{eventDatabase: db, beforeBegin: func(appendCtx context.Context) {
					begins++
					if _, ok := appendCtx.Value(expectedEventCountKey{}).(int64); !ok {
						t.Error("validated event head not carried to append")
					}
					if begins > 1 && mutation != "continued-conflicts" {
						return
					}
					if mutation == "transfer" {
						replaceCoordinator(t, h)
					} else if mutation == "disable" || mutation == "stop" {
						path := "/api/events"
						body := `{"kind":"feature.flag.updated","payload":{"name":"agentCoordination","enabled":false}}`
						if mutation == "stop" {
							path = "/api/agent-safety"
							body = `{"action":"stop","scope":"application"}`
						}
						w := mockRequest(h, "POST", path, body)
						if w.Code != 200 {
							t.Fatal(w.Code, w.Body.String())
						}
					} else {
						w := mockRequest(h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"jobDiagnostics","enabled":false}}`)
						if w.Code != 200 {
							t.Fatal(w.Code, w.Body.String())
						}
					}
				}}
				var err error
				switch path {
				case "result":
					s.finishAgentTurn(ctx, p, "synthetic result", "codex", nil)
				case "session":
					err = s.recordAgentSession(ctx, p, "chat", "synthetic-thread")
				case "submitted":
					err = registration.Update("submitted")
				}
				if mutation == "unrelated" && (err != nil || begins != 2) {
					t.Fatal("healthy conflict did not retry", err, begins)
				}
				if (mutation == "transfer" || mutation == "disable" || mutation == "stop") && path != "result" && err == nil {
					t.Fatal("revoked metadata accepted")
				}
				if mutation == "continued-conflicts" && (begins != 3 || path != "result" && !errors.Is(err, errAuditConflict)) {
					t.Fatal("retry limit not enforced", err, begins)
				}
				if s.integrityFault {
					t.Fatal("ordinary conflict falsely latched integrity fault")
				}
				st, err := s.load(t.Context())
				if err != nil {
					t.Fatal(err)
				}
				succeeded := false
				switch path {
				case "result":
					for _, m := range st.Messages {
						if m.Text == "synthetic result" {
							succeeded = true
						}
					}
					if mutation == "transfer" || mutation == "disable" || mutation == "stop" {
						found := false
						for _, current := range st.Participants {
							if current.ID == p.ID {
								found = current.Status == "Stopped"
							}
						}
						if !found {
							t.Fatal("revoked completion not recorded as stopped")
						}
					}
				case "session":
					for _, current := range st.Participants {
						if current.ID == p.ID {
							succeeded = current.ThreadID == "synthetic-thread"
						}
					}
				case "submitted":
					job := st.ManagedJobs[registration.JobID]
					succeeded = job.Status == "submitted"
					if job.ShutdownConfirmed {
						t.Fatal("race manufactured shutdown proof")
					}
				}
				if succeeded != (mutation == "unrelated") {
					t.Fatal("stale write committed or healthy write lost", succeeded)
				}
				// A failed/colliding append must leave no partially saved session/result.
				for _, e := range db.snapshot() {
					if mutation != "unrelated" && (e.kind == "message.created" && strings.Contains(e.payload, "synthetic result") || e.kind == "participant.session" && strings.Contains(e.payload, "synthetic-thread")) {
						t.Fatal("stale event persisted")
					}
				}
			})
		}
	}
}

// Existing independent human work is not revoked by a coordinator replacement.
func TestIndependentResultSurvivesConcurrentCoordinatorTransfer(t *testing.T) {
	t.Parallel()
	s, db, h, p, _ := coordinationFixture(t)
	other := &server{db: db, auth: s.auth}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/governance", other.governance)
	h = other.auth.protect(other.actorGuard(mux))
	once := false
	s.db = &interleavedEventDB{eventDatabase: db, beforeBegin: func(context.Context) {
		if !once {
			once = true
			replaceCoordinator(t, h)
		}
	}}
	s.finishAgentTurn(context.Background(), p, "independent result", "codex", nil)
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range st.Messages {
		if m.Text == "independent result" {
			return
		}
	}
	t.Fatal("independent human result was discarded")
}
