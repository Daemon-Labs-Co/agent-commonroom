package main

import (
	"agentroom/internal/room"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func connectionRaceFixture(t *testing.T) (*server, *mockDB, room.Participant) {
	t.Helper()
	s, db, _ := mockServer(t)
	db.seed("connection.created", room.Connection{ID: "connection", Name: "Synthetic original", Protocol: "CLI", Endpoint: "ssh://worker/codex", Revision: 1})
	db.seed("connection.created", room.Connection{ID: "other", Name: "Synthetic other", Protocol: "CLI", Endpoint: "ssh://other/codex", Revision: 1})
	p := room.Participant{ID: "agent", RoomID: "engineering", ConnectionID: "connection", Name: "Synthetic", Rank: 1, Revision: 1}
	db.seed("participant.created", p)
	return s, db, p
}
func connectionRaceBody(operation string) string {
	var payload any = room.Connection{ID: "connection", Name: "Synthetic edited", Protocol: "CLI", Endpoint: "ssh://other/codex", Revision: 1}
	if operation == "connection.deleted" {
		payload = map[string]string{"id": "connection"}
	} else if operation == "participant.updated" {
		payload = room.Participant{ID: "agent", RoomID: "engineering", ConnectionID: "other", Name: "Synthetic", Rank: 1, Revision: 1}
	}
	raw, _ := json.Marshal(payload)
	return fmt.Sprintf(`{"kind":%q,"payload":%s}`, operation, raw)
}
func connectionRaceHandler(s *server) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/events", s.events)
	return s.auth.protect(s.actorGuard(mux))
}
func TestConnectionMutationAndWorkerAllocationCannotBothCommitStaleChecks(t *testing.T) {
	for _, operation := range []string{"connection.updated", "connection.deleted", "participant.updated"} {
		for _, winner := range []string{"allocation", "mutation"} {
			t.Run(operation+"/"+winner, func(t *testing.T) {
				t.Parallel()
				s, db, p := connectionRaceFixture(t)
				other := &server{db: db, auth: s.auth}
				begins := 0
				if winner == "allocation" {
					reg := other.managedJobContext(context.Background(), p).Value(managedJobContextKey{}).(managedJobRegistration)
					s.db = &interleavedEventDB{eventDatabase: db, beforeBegin: func(ctx context.Context) {
						begins++
						if _, ok := ctx.Value(expectedEventCountKey{}).(int64); !ok {
							t.Error("mutation lost audit head fence")
						}
						if err := reg.Register("worker", "codex"); err != nil {
							t.Fatal(err)
						}
					}}
					w := mockRequest(connectionRaceHandler(s), "POST", "/api/events", connectionRaceBody(operation))
					if w.Code != 409 {
						t.Fatal("stale mutation committed", w.Code, w.Body.String())
					}
					st, err := other.load(t.Context())
					if err != nil || len(st.ManagedJobs) != 1 {
						t.Fatal("winning allocation lost", err)
					}
					if !connectionHasActiveAgent(st, nil, "connection") {
						t.Fatal("durable original connection lost")
					}
					if err := reg.Update("shutdown_confirmed"); err != nil {
						t.Fatal(err)
					}
				} else {
					reg := s.managedJobContext(context.Background(), p).Value(managedJobContextKey{}).(managedJobRegistration)
					s.db = &interleavedEventDB{eventDatabase: db, beforeBegin: func(ctx context.Context) {
						begins++
						if _, ok := ctx.Value(expectedEventCountKey{}).(int64); !ok {
							t.Error("allocation lost audit head fence")
						}
						w := mockRequest(connectionRaceHandler(other), "POST", "/api/events", connectionRaceBody(operation))
						if w.Code != 200 {
							t.Fatal(w.Code, w.Body.String())
						}
					}}
					if err := reg.Register("worker", "codex"); !errors.Is(err, errAuditConflict) {
						t.Fatal("stale allocation not classified as conflict", err)
					}
					st, err := other.load(t.Context())
					if err != nil || len(st.ManagedJobs) != 0 || len(s.activeJobs) != 0 {
						t.Fatal("refused allocation journaled", err)
					}
					if reg.Update("submitted") == nil {
						t.Fatal("refused allocation submitted")
					}
				}
				if begins != 1 || s.integrityFault || other.integrityFault {
					t.Fatal("ordinary conflict retried effects or became integrity incident", begins)
				}
			})
		}
	}
}
