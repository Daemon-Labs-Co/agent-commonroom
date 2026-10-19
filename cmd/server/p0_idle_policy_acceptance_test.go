package main

import (
	"agentroom/internal/room"
	"encoding/json"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func driveP0IdleUnauthorizedConfig(t *testing.T, scenario pendingP0Case) pendingP0Observation {
	t.Helper()
	s, db, h := idleSettingsHTTP(t)
	db.seed("actor.revised", room.Actor{ID: "developer", Kind: "human", Enabled: true})
	s.auth.sessions["owner"] = ownerSession{ActorID: "developer", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	seconds, err := strconv.ParseFloat(scenario.Given["requested_seconds"], 64)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(map[string]any{"kind": "agents.idle_policy.updated", "payload": map[string]any{"minutes": seconds / 60, "revision": 0}})
	if err != nil {
		t.Fatal(err)
	}
	before := db.snapshot()
	response := mockRequest(h, "POST", "/api/events", string(payload))
	out := pendingP0Observation{State: "allow", Values: map[string]string{}}
	if response.Code == 403 {
		out.State = "deny"
	}
	after := db.snapshot()
	if !reflect.DeepEqual(before, after) {
		out.Effects = append(out.Effects, "idle_config_changed")
	}
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if st.AgentIdleSettings.Revision != 0 || st.AgentIdleSettings.Minutes != 0 {
		out.Effects = append(out.Effects, "idle_config_changed")
	}
	return out
}
