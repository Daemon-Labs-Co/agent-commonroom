package main

import (
	"agentroom/internal/room"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestParseTaskSuggestions(t *testing.T) {
	t.Parallel()
	m := room.Message{ID: "m", RoomID: "engineering", ActorID: "bot-a", Author: "Avery", ActorType: "agent", Text: "Here is the plan.\n[PROPOSE TASK] Add login :: Users can sign in and out :: Blake\n  [propose task] Write tests\n[PROPOSE TASK]   \nnot [PROPOSE TASK] inline\n" + strings.Repeat("[PROPOSE TASK] filler\n", 6)}
	got := parseTaskSuggestions(m)
	if len(got) != maxSuggestionsPerMessage {
		t.Fatal("per-message bound", len(got))
	}
	if got[0].ID != "m#1" || got[0].Title != "Add login" || got[0].Criteria != "Users can sign in and out" || got[0].Assignee != "Blake" || got[0].Status != "open" {
		t.Fatalf("first suggestion: %+v", got[0])
	}
	if got[1].Title != "Write tests" || got[1].Criteria != "" || got[2].Title != "filler" {
		t.Fatalf("optional parts or empty title handling: %+v %+v", got[1], got[2])
	}
}

func TestTaskSuggestionsFlow(t *testing.T) {
	// Synthetic worker configuration; the scripted turn never reaches a provider.
	t.Setenv("COMMONROOM_REMOTE_WORKER", "true")
	t.Setenv("COMMONROOM_REMOTE_WORKER_ALIAS", "worker")
	withProposals := func(s *server, h http.Handler) http.Handler {
		mux := http.NewServeMux()
		mux.Handle("/", h)
		mux.Handle("/api/task-proposals", s.auth.protect(s.actorGuard(http.HandlerFunc(s.taskProposals))))
		return mux
	}

	t.Run("ask, capture through the human proposal path, dismiss", func(t *testing.T) {
		s, db, h, script := discussionFixture(t, false)
		h = withProposals(s, h)
		if w := mockRequest(h, "POST", "/api/agents", `{"action":"suggest-tasks"}`); w.Code != 403 {
			t.Fatal("disabled feature accepted", w.Code)
		}
		db.seed("feature.flag.updated", map[string]any{"name": "governingSuggestions", "enabled": true})
		db.seed("feature.flag.updated", map[string]any{"name": "taskProposals", "enabled": true})
		script.reply = func(string, int) string {
			return "Two things are settled.\n[PROPOSE TASK] Add login :: Users can sign in :: Blake\n[PROPOSE TASK] Write tests :: Tests pass"
		}
		if w := mockRequest(h, "POST", "/api/agents", `{"action":"suggest-tasks"}`); w.Code != 202 || !strings.Contains(w.Body.String(), "Avery") {
			t.Fatal(w.Code, w.Body.String())
		}
		st := awaitIdle(t, s)
		if len(script.calls) != 1 || !strings.Contains(script.prompts[0], "[PROPOSE TASK] short title :: acceptance criteria") {
			t.Fatal("agent not asked for suggestions", script.calls)
		}
		if len(st.TaskSuggestions) != 2 || st.TaskSuggestions[0].Status != "open" || len(st.TaskProposals) != 0 || len(st.Tasks) != 0 {
			t.Fatalf("suggestions must be derived without creating proposals or tasks: %+v", st.TaskSuggestions)
		}
		first := st.TaskSuggestions[0]
		capture := fmt.Sprintf(`{"action":"capture","sourceMessageId":%q,"title":%q,"criteria":%q}`, first.MessageID, first.Title, first.Criteria)
		if w := proposalRequest(h, "owner", "capture-suggestion-key", capture); w.Code != 200 && w.Code != 201 {
			t.Fatal("capture through the existing human endpoint failed", w.Code, w.Body.String())
		}
		second := st.TaskSuggestions[1]
		if w := mockRequest(h, "POST", "/api/events", fmt.Sprintf(`{"kind":"task.suggestion.dismissed","payload":{"id":%q}}`, second.ID)); w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		if w := mockRequest(h, "POST", "/api/events", fmt.Sprintf(`{"kind":"task.suggestion.dismissed","payload":{"id":%q}}`, second.ID)); w.Code != 404 {
			t.Fatal("suggestion dismissed twice", w.Code)
		}
		st, _ = s.load(t.Context())
		if st.TaskSuggestions[0].Status != "captured" || st.TaskSuggestions[1].Status != "dismissed" || len(st.TaskProposals) != 1 {
			t.Fatalf("status not derived: %+v", st.TaskSuggestions)
		}
		db.seed("feature.flag.updated", map[string]any{"name": "governingSuggestions", "enabled": false})
		if st, _ = s.load(t.Context()); len(st.TaskSuggestions) != 0 || len(st.TaskProposals) != 1 {
			t.Fatal("turning the feature off must hide suggestions and keep captured proposals")
		}
	})

	t.Run("suggestions follow a discussion that ends in agreement", func(t *testing.T) {
		s, db, h, script := discussionFixture(t, true)
		db.seed("feature.flag.updated", map[string]any{"name": "governingSuggestions", "enabled": true})
		script.reply = func(name string, call int) string {
			if call <= 2 {
				return "[AGREE] " + name
			}
			return "[PROPOSE TASK] Build it :: It works"
		}
		if w := mockRequest(h, "POST", "/api/agents", `{"action":"discuss","text":"Plan the build"}`); w.Code != 202 {
			t.Fatal(w.Code, w.Body.String())
		}
		st := awaitIdle(t, s)
		if strings.Join(script.calls, ",") != "Avery,Blake,Avery" || len(st.TaskSuggestions) != 1 || st.Rooms[0].AgentInitiated != 1 {
			t.Fatal("agreement did not lead to one suggestion turn", script.calls, len(st.TaskSuggestions))
		}
	})
}
