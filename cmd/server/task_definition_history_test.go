package main

import (
	"agentroom/internal/room"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func definitionHistoryHandler(s *server) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/task-definition-history", s.taskDefinitionHistory)
	return s.auth.protect(s.actorGuard(mux))
}
func TestDefinitionHistoryPreservesUnexecutedVersionsAndDeduplicatesStateEdits(t *testing.T) {
	t.Parallel()
	s, db, h, task := structuredTaskFixture(t)
	for i := 2; i <= 12; i++ {
		st, err := s.load(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		task = st.Tasks[0]
		task.Definition.Task = fmt.Sprintf("Synthetic revision %d <safe>", i)
		raw, _ := json.Marshal(task)
		w := mockRequest(h, "POST", "/api/events", fmt.Sprintf(`{"kind":"task.revised","payload":%s}`, raw))
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	task = st.Tasks[0]
	task.Status = "Draft"
	raw, _ := json.Marshal(task)
	if w := mockRequest(h, "POST", "/api/events", fmt.Sprintf(`{"kind":"task.revised","payload":%s}`, raw)); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	history := definitionHistoryHandler(s)
	var page struct {
		Definitions []room.TaskDefinitionRevision `json:"definitions"`
		Next        string                        `json:"nextCursor"`
	}
	w := mockRequest(history, "GET", "/api/task-definition-history?room=engineering&taskId=task", "")
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &page) != nil || len(page.Definitions) != 10 || page.Definitions[0].Definition.Revision != 12 || page.Next != "3" {
		t.Fatal(w.Code, w.Body.String())
	}
	w = mockRequest(history, "GET", "/api/task-definition-history?room=engineering&taskId=task&before="+page.Next, "")
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &page) != nil || len(page.Definitions) != 2 || page.Definitions[1].Definition.Revision != 1 || page.Next != "" {
		t.Fatal(w.Code, w.Body.String())
	}
	if page.Definitions[1].Definition.Task != "Return a synthetic summary" {
		t.Fatal("old definition mutated")
	}
	db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": false})
	if w := mockRequest(history, "GET", "/api/task-definition-history?room=engineering&taskId=task", ""); w.Code != 200 {
		t.Fatal("off erased retained versions", w.Code)
	}
	st, err = s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(st)
	if strings.Contains(string(encoded), "taskDefinitionHistory") || strings.Contains(string(encoded), "Synthetic revision 2 ") {
		t.Fatal("global state disclosed private versions")
	}
	for _, path := range []string{"?room=engineering&taskId=task&before=999", "?room=engineering&taskId=missing", "?room=foreign&taskId=task"} {
		if w := mockRequest(history, "GET", "/api/task-definition-history"+path, ""); w.Code == 200 {
			t.Fatal("invalid scope/cursor accepted", path)
		}
	}
	db.seed("actor.revised", room.Actor{ID: "outsider", Kind: "human", Enabled: true, RoomIDs: []string{}})
	s.auth.sessions["owner"] = ownerSession{ActorID: "outsider", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	if w := mockRequest(history, "GET", "/api/task-definition-history?room=engineering&taskId=task", ""); w.Code == 200 {
		t.Fatal("foreign member read versions")
	}
}
func TestDefinitionHistoryPrivateReadsRequireAuditAndLegacyIsEmpty(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	db.seed("actor.revised", room.Actor{ID: "member", Name: "Synthetic member", Kind: "human", Enabled: true})
	db.seed("task.revised", room.Task{ID: "private", RoomID: "personal-member", Title: "Legacy task", Status: "Draft", Revision: 1})
	h := definitionHistoryHandler(s)
	before := len(db.snapshot())
	w := mockRequest(h, "GET", "/api/task-definition-history?room=personal-member&taskId=private", "")
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"definitions":[]`) || len(db.snapshot()) != before+1 || db.snapshot()[before].kind != "room.owner_access" {
		t.Fatal(w.Code, w.Body.String())
	}
	db.failInsertAt = 1
	if w := mockRequest(h, "GET", "/api/task-definition-history?room=personal-member&taskId=private", ""); w.Code != 503 || strings.Contains(w.Body.String(), "definitions") {
		t.Fatal("unaudited history", w.Code)
	}
}

func TestDefinitionHistoryDoesNotInventFailedWritesOrAcceptMutation(t *testing.T) {
	t.Parallel()
	s, db, h, task := structuredTaskFixture(t)
	task.Definition.Task = "Unacknowledged change"
	raw, _ := json.Marshal(task)
	db.failInsertAt = 1
	w := mockRequest(h, "POST", "/api/events", fmt.Sprintf(`{"kind":"task.revised","payload":%s}`, raw))
	if w.Code == 200 {
		t.Fatal("failed write accepted")
	}
	db.failInsertAt = 0
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	records := st.TaskDefinitionHistory[task.ID]
	if len(records) != 1 || records[0].Definition.Task == task.Definition.Task {
		t.Fatal("unacknowledged version exposed", records)
	}
	history := definitionHistoryHandler(s)
	if w := mockRequest(history, "POST", "/api/task-definition-history?room=engineering&taskId=task", `{}`); w.Code != 405 {
		t.Fatal("history endpoint accepted mutation", w.Code)
	}
}
