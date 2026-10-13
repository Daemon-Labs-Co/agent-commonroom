package main

import (
	"agentroom/internal/room"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func keyedEvent(h http.Handler, key, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest("POST", "http://127.0.0.1:8787/api/events", strings.NewReader(body))
	r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: "owner"})
	r.Header.Set("X-CSRF-Token", "csrf")
	r.Header.Set("Idempotency-Key", key)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}
func TestOperationReplayAndKeyConflict(t *testing.T) {
	t.Parallel()
	_, db, h := mockServer(t)
	body := `{"kind":"room.created","payload":{"name":"Synthetic"}}`
	a := keyedEvent(h, "operation-123", body)
	if a.Code != 200 {
		t.Fatal(a.Code, a.Body.String())
	}
	count := len(db.snapshot())
	b := keyedEvent(h, "operation-123", body)
	if b.Code != 200 || referenceJSON(json.RawMessage(b.Body.Bytes())) != referenceJSON(json.RawMessage(a.Body.Bytes())) || b.Header().Get("Idempotent-Replay") != "true" {
		t.Fatal(b.Code, b.Body.String(), a.Body.String())
	}
	if len(db.snapshot()) != count {
		t.Fatal("duplicate creation")
	}
	c := keyedEvent(h, "operation-123", `{"kind":"room.created","payload":{"name":"Different"}}`)
	if c.Code != 409 || len(db.snapshot()) != count {
		t.Fatal("key reused")
	}
}
func TestBatchRollbackAtEveryInsertAndCommit(t *testing.T) {
	t.Parallel()
	for _, stage := range []string{"first", "second", "receipt", "commit"} {
		t.Run(stage, func(t *testing.T) {
			t.Parallel()
			db := newMockDB(t)
			s := server{db: db}
			switch stage {
			case "first":
				db.failInsertAt = 1
			case "second":
				db.failInsertAt = 2
			case "receipt":
				db.failInsertAt = 3
			case "commit":
				db.failAt = "commit"
			}
			ctx := context.WithValue(t.Context(), operationReceiptKey{}, room.OperationReceipt{Key: strings.Repeat("a", 64), Fingerprint: strings.Repeat("b", 64), Status: 200, Response: json.RawMessage(`{"ok":true}`)})
			err := s.appendBatch(ctx, []auditChange{{Kind: "connection.created", Payload: room.Connection{ID: "c"}}, {Kind: "participant.created", Payload: room.Participant{ID: "p", ConnectionID: "c"}}})
			if err == nil || len(db.snapshot()) != 0 {
				t.Fatal("partial operation persisted", err)
			}
		})
	}
}
func TestExpectedHeadRejectsConcurrentBusinessChange(t *testing.T) {
	t.Parallel()
	db := newMockDB(t)
	s := server{db: db}
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if err = s.append(t.Context(), "message.created", room.Message{Text: "intervening"}); err != nil {
		t.Fatal(err)
	}
	ctx := context.WithValue(t.Context(), expectedEventCountKey{}, st.EventCount)
	if err = s.appendBatch(ctx, []auditChange{{Kind: "room.created", Payload: room.Room{Name: "stale"}}}); err != errAuditConflict || len(db.snapshot()) != 1 {
		t.Fatal("stale operation committed", err)
	}
}
func TestBatchMaintainsCanonicalAuditChain(t *testing.T) {
	t.Parallel()
	db := newMockDB(t)
	s := server{db: db}
	if err := s.appendBatch(t.Context(), []auditChange{{Kind: "room.created", Payload: room.Room{Name: "one"}}, {Kind: "room.created", Payload: room.Room{Name: "two"}}}); err != nil {
		t.Fatal(err)
	}
	events := db.snapshot()
	if len(events) != 2 || events[1].sequence != events[0].sequence+1 || events[1].previous != events[0].hash {
		t.Fatal("broken batch chain")
	}
}
func TestStaleMutableAggregateEdits(t *testing.T) {
	t.Parallel()
	for _, item := range []struct {
		kind    string
		value   any
		payload string
	}{{"company.updated", room.Company{ID: "company", Name: "current", Revision: 3}, `{"name":"stale","revision":2}`}, {"room.updated", room.Room{ID: "engineering", Name: "current", Revision: 3}, `{"id":"engineering","name":"stale","revision":2}`}, {"connection.updated", room.Connection{ID: "c", Name: "current", Protocol: "CLI", Endpoint: "codex", Revision: 3}, `{"id":"c","name":"stale","protocol":"CLI","endpoint":"codex","revision":2}`}, {"participant.updated", room.Participant{ID: "p", RoomID: "engineering", Name: "current", ConnectionID: "c", Rank: 1, Revision: 3}, `{"id":"p","name":"stale","connectionId":"c","rank":1,"revision":2}`}} {
		t.Run(item.kind, func(t *testing.T) {
			t.Parallel()
			_, db, h := mockServer(t)
			db.seed("connection.created", room.Connection{ID: "c", Protocol: "CLI", Endpoint: "codex"})
			if item.kind == "participant.updated" {
				db.seed("participant.created", item.value)
			} else {
				db.seed(item.kind, item.value)
			}
			before := len(db.snapshot())
			w := mockRequest(h, "POST", "/api/events", `{"kind":"`+item.kind+`","payload":`+item.payload+`}`)
			if w.Code != 409 || len(db.snapshot()) != before {
				t.Fatal("stale edit accepted", w.Code, w.Body.String())
			}
		})
	}
}
func TestFailedOperationCanRetrySameKeyWithoutPartialReceipt(t *testing.T) {
	t.Parallel()
	_, db, h := mockServer(t)
	db.failInsertAt = 2
	body := `{"kind":"room.created","payload":{"name":"Retry"}}`
	if w := keyedEvent(h, "retry-operation", body); w.Code != 500 {
		t.Fatal(w.Code)
	}
	if len(db.snapshot()) != 0 {
		t.Fatal("partial receipt")
	}
	db.failInsertAt = 0
	if w := keyedEvent(h, "retry-operation", body); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
}
