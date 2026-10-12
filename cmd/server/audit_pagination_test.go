package main

import (
	"agentroom/internal/room"
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestAuditPaginationBounds(t *testing.T) {
	t.Parallel()
	for _, query := range []string{"before=-1", "before=oops", "from=invalid", "from=2026-10-12T00:00:00Z&until=2026-10-11T00:00:00Z"} {
		v, _ := url.ParseQuery(query)
		if _, err := auditPageBounds(v); err == nil {
			t.Fatal(query)
		}
	}
}
func TestAuditPagesAndDateFilters(t *testing.T) {
	t.Parallel()
	s, db, h := mockServer(t)
	for i := 0; i < 205; i++ {
		db.seed("message.created", room.Message{Author: "Fixture", Text: "never expose synthetic body"})
	}
	var first, second struct {
		Events []struct {
			Sequence int
			Summary  string
		}
		NextBefore *int
	}
	w := mockRequest(h, "GET", "/api/events?page=true", "")
	if json.Unmarshal(w.Body.Bytes(), &first) != nil || len(first.Events) != 100 || *first.NextBefore != 106 {
		t.Fatal(w.Body.String())
	}
	w = mockRequest(h, "GET", "/api/events?page=true&before=106", "")
	json.Unmarshal(w.Body.Bytes(), &second)
	if len(second.Events) != 100 || second.Events[0].Sequence != 105 || *second.NextBefore != 6 {
		t.Fatal(w.Body.String())
	}
	w = mockRequest(h, "GET", "/api/events?page=true&before=6", "")
	json.Unmarshal(w.Body.Bytes(), &second)
	if len(second.Events) != 5 || second.NextBefore != nil {
		t.Fatal(w.Body.String())
	}
	w = mockRequest(h, "GET", "/api/events?page=true&from=1970-01-01T00:00:02Z&until=1970-01-01T00:00:04Z", "")
	json.Unmarshal(w.Body.Bytes(), &second)
	if len(second.Events) != 2 || second.Events[0].Sequence != 4 || second.Events[1].Sequence != 3 {
		t.Fatal(w.Body.String())
	}
	w = mockRequest(h, "GET", "/api/events?page=true&before=oops", "")
	if w.Code != 400 {
		t.Fatal(w.Code)
	}
	if len(db.snapshot()) != 205 {
		t.Fatal("read changed audit records")
	}
	w = httptest.NewRecorder()
	s.events(w, httptest.NewRequest("GET", "/api/events?page=true&from=2026-10-11T00:00:00Z", nil))
	json.Unmarshal(w.Body.Bytes(), &second)
	if len(second.Events) != 0 {
		t.Fatal("empty today must stay empty")
	}
}
func TestHistoryPagesKeepOwnerAuthorization(t *testing.T) {
	t.Parallel()
	_, _, h := mockServer(t)
	r := httptest.NewRequest("GET", "/api/events?page=true", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 401 && w.Code != 403 {
		t.Fatal(w.Code)
	}
}
