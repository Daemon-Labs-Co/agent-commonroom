package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestHistoryPresentationSafeContextAndSeverity(t *testing.T) {
	t.Parallel()
	d := describeAuditEvent("agent.error", []byte(`{"name":"Synthetic agent","roomId":"engineering","error":"context deadline exceeded secret-token","endpoint":"https://secret.example","text":"private message"}`))
	raw, _ := json.Marshal(d)
	if d.Category != "Agents" || d.Tone != "error" || d.Routine || !strings.Contains(d.Explanation, "time limit") {
		t.Fatal(d)
	}
	for _, secret := range []string{"secret-token", "secret.example", "private message"} {
		if strings.Contains(string(raw), secret) {
			t.Fatal("raw payload exposed", string(raw))
		}
	}
	d = describeAuditEvent("workflow.run.updated", []byte(`{"actorId":"owner","status":"awaiting_human","workflow":{"name":"Delivery"},"detail":"private instruction"}`))
	if d.Subject != "Delivery" || d.Tone != "warning" || d.ActorID != "owner" {
		t.Fatal(d)
	}
	d = describeAuditEvent("participant.session", []byte(`{"name":"Agent","actorId":"subject-identity","status":"Idle"}`))
	if !d.Routine || d.ActorID != "" {
		t.Fatal("subject confused with performer", d)
	}
	d = describeAuditEvent("participant.session", []byte(`{"status":"Blocked"}`))
	if d.Routine || d.Tone != "error" {
		t.Fatal("blocked event hidden as routine", d)
	}
	d = describeAuditEvent("feature.flag.updated", []byte(`{"name":"historyRefinement","enabled":false}`))
	if d.Outcome != "Off" || d.Subject != "historyRefinement" {
		t.Fatal(d)
	}
}

func TestHistoryRefinementServerFlagBothStates(t *testing.T) {
	t.Parallel()
	_, db, h := mockServer(t)
	db.seed("message.created", map[string]string{"author": "Synthetic actor", "roomId": "engineering", "text": "private body"})
	for _, enabled := range []bool{true, false, true} {
		db.seed("feature.flag.updated", map[string]any{"name": "historyRefinement", "enabled": enabled})
		w := mockRequest(h, "GET", "/api/events?page=true", "")
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		var page struct {
			Events []map[string]any `json:"events"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &page); err != nil {
			t.Fatal(err)
		}
		for _, e := range page.Events {
			_, exists := e["presentation"]
			if exists != enabled {
				t.Fatal("metadata flag not enforced", e)
			}
		}
		if strings.Contains(w.Body.String(), "private body") {
			t.Fatal("message body leaked")
		}
	}
	db2 := newMockDB(t)
	fresh := &server{db: db2}
	enabled, err := fresh.historyPresentationEnabled(context.Background())
	if err != nil || !enabled {
		t.Fatal("default must be on")
	}
}
