package main

import (
	"agentroom/internal/room"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func executionHistoryHandler(s *server) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/task-executions", s.taskExecutions)
	return s.auth.protect(s.actorGuard(mux))
}
func TestTaskExecutionHistoryPaginationAndScopedAccess(t *testing.T) {
	t.Parallel()
	s, db, h, task := structuredTaskFixture(t)
	s.taskTurn = func(context.Context, room.Participant, string, string) (string, error) {
		return `{"result":"<stored value>"}`, nil
	}
	for i := 0; i < 12; i++ {
		if w := dispatchRequest(h, fmt.Sprintf("history-attempt-%d", i), definedDispatchBody(task)); w.Code != 202 {
			t.Fatal(w.Code, w.Body.String())
		}
		task = waitDispatch(t, s).Tasks[0]
	}
	history := executionHistoryHandler(s)
	w := mockRequest(history, "GET", "/api/task-executions?room=engineering&taskId=task", "")
	var page struct {
		Executions []room.TaskExecution `json:"executions"`
		NextCursor string               `json:"nextCursor"`
	}
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &page) != nil || len(page.Executions) != 10 || page.NextCursor == "" {
		t.Fatal(w.Code, w.Body.String())
	}
	firstID := page.Executions[0].ID
	w = mockRequest(history, "GET", "/api/task-executions?room=engineering&taskId=task&before="+page.NextCursor, "")
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &page) != nil || len(page.Executions) != 2 || page.NextCursor != "" || page.Executions[0].ID == firstID {
		t.Fatal("cursor lost or duplicated history", w.Code, w.Body.String())
	}
	db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": false})
	if w = mockRequest(history, "GET", "/api/task-executions?room=engineering&taskId=task", ""); w.Code != 200 || !strings.Contains(w.Body.String(), "validated") {
		t.Fatal("flag off erased history", w.Code)
	}
	db.seed("actor.revised", room.Actor{ID: "foreign-member", Kind: "human", Enabled: true, RoomIDs: []string{}})
	s.auth.sessions["owner"] = ownerSession{ActorID: "foreign-member", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	if w = mockRequest(history, "GET", "/api/task-executions?room=engineering&taskId=task", ""); w.Code == 200 {
		t.Fatal("foreign actor obtained execution values")
	}
}
func TestTaskExecutionPersonalRoomReadsRetainOwnerAudit(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	db.seed("actor.revised", room.Actor{ID: "member", Name: "Synthetic member", Kind: "human", Enabled: true})
	db.seed("task.revised", room.Task{ID: "private-task", RoomID: "personal-member", Title: "Synthetic personal task", Status: "Draft", Revision: 1})
	h := executionHistoryHandler(s)
	before := len(db.snapshot())
	w := mockRequest(h, "GET", "/api/task-executions?room=personal-member&taskId=private-task", "")
	if w.Code != 200 || len(db.snapshot()) != before+1 || db.snapshot()[before].kind != "room.owner_access" {
		t.Fatal("owner read bypassed personal audit", w.Code, w.Body.String())
	}
	db.failInsertAt = 1
	w = mockRequest(h, "GET", "/api/task-executions?room=personal-member&taskId=private-task", "")
	if w.Code != 503 || strings.Contains(w.Body.String(), "executions") {
		t.Fatal("unaudited read disclosed history", w.Code)
	}
}
func TestTaskContractPersistenceFailuresDoNotInventExecutionsOrOutputs(t *testing.T) {
	for _, stage := range []string{"start", "finish"} {
		t.Run(stage, func(t *testing.T) {
			t.Parallel()
			s, db, h, task := structuredTaskFixture(t)
			called := make(chan struct{}, 1)
			s.taskTurn = func(context.Context, room.Participant, string, string) (string, error) {
				called <- struct{}{}
				db.failInsertAt = 1
				return `{"result":"not acknowledged"}`, nil
			}
			if stage == "start" {
				db.failInsertAt = 2
			}
			before := len(db.snapshot())
			w := dispatchRequest(h, "persistence-attempt", definedDispatchBody(task))
			if stage == "start" {
				if w.Code != 500 || len(db.snapshot()) != before || len(s.active) != 0 {
					t.Fatal("partial task execution admitted", w.Code)
				}
				select {
				case <-called:
					t.Fatal("provider started after failed allocation")
				default:
				}
				return
			}
			if w.Code != 202 {
				t.Fatal(w.Code, w.Body.String())
			}
			st := waitDispatch(t, s)
			attempt := st.TaskExecutions[st.Tasks[0].DispatchID]
			if attempt.Status != "running" || len(attempt.Outputs) != 0 || st.Tasks[0].DispatchStatus != "pending" || !s.integrityFault {
				t.Fatal("unacknowledged output became deliverable", attempt)
			}
		})
	}
}
func TestTaskDefinitionFlagDoesNotPoisonLegacyJobEpoch(t *testing.T) {
	t.Parallel()
	s, db, h, _ := dispatchFixture(t)
	db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": false})
	done := make(chan error, 1)
	s.taskTurn = func(ctx context.Context, p room.Participant, _, _ string) (string, error) {
		jobCtx := s.managedJobContext(ctx, p)
		reg := jobCtx.Value(managedJobContextKey{}).(managedJobRegistration)
		err := reg.Register("synthetic-worker", "codex")
		if err == nil {
			err = reg.Update("shutdown_confirmed")
		}
		done <- err
		return "Synthetic legacy result", err
	}
	if w := dispatchRequest(h, "legacy-after-flag", dispatchBody); w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatal("legacy job inherited structured epoch", err)
		}
	case <-time.After(time.Second):
		t.Fatal("legacy did not run")
	}
	_ = waitDispatch(t, s)
}
