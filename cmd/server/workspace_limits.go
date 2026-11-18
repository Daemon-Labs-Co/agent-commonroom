package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"
)

// Per-workspace limits keep one workspace from consuming the installation. A
// reached limit is a visible refusal; it never extends another limit. The
// platform operator sets them through the platform API; a single-workspace
// installation has no limits (nil).
type workspaceLimits struct {
	Members           int `json:"members"`
	Rooms             int `json:"rooms"`
	Agents            int `json:"agents"`
	RequestsPerMinute int `json:"requestsPerMinute"`
	DailyTurns        int `json:"dailyTurns"`
	ConcurrentTurns   int `json:"concurrentTurns"`
}

var defaultWorkspaceLimits = workspaceLimits{Members: 5, Rooms: 20, Agents: 10, RequestsPerMinute: 600, DailyTurns: 200, ConcurrentTurns: 2}

// parseWorkspaceLimits overlays the supplied fields on the defaults. Unknown
// fields and out-of-range values are refused so a typo cannot silently lift a cap.
func parseWorkspaceLimits(raw []byte) (workspaceLimits, error) {
	limits := defaultWorkspaceLimits
	if len(bytes.TrimSpace(raw)) == 0 {
		return limits, nil
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&limits); err != nil || dec.More() {
		return workspaceLimits{}, errors.New("limits must be a JSON object with members, rooms, agents, requestsPerMinute, dailyTurns and concurrentTurns")
	}
	for _, c := range []struct {
		name          string
		value, lo, hi int
	}{
		{"members", limits.Members, 1, 1000}, {"rooms", limits.Rooms, 1, 1000}, {"agents", limits.Agents, 0, 200},
		{"requestsPerMinute", limits.RequestsPerMinute, 10, 100000}, {"dailyTurns", limits.DailyTurns, 0, 100000}, {"concurrentTurns", limits.ConcurrentTurns, 1, 16},
	} {
		if c.value < c.lo || c.value > c.hi {
			return workspaceLimits{}, errors.New(c.name + " is outside its allowed range")
		}
	}
	return limits, nil
}

// setLimits applies new limits to a running workspace; nil removes them.
func (s *server) setLimits(l *workspaceLimits) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.limits = l
	if l != nil {
		s.dailyTurnLimit = l.DailyTurns
	}
}

// limitRefusal reports why creating one more of something is refused, or "".
// The caller holds s.mu.
func (s *server) limitRefusal(kind string, current int) string {
	if s.limits == nil {
		return ""
	}
	var max int
	switch kind {
	case "members":
		max = s.limits.Members
	case "rooms":
		max = s.limits.Rooms
	case "agents":
		max = s.limits.Agents
	}
	if current >= max {
		return "This workspace has reached its limit of " + itoa(max) + " " + kind + ". Ask the platform operator to raise it"
	}
	return ""
}

func jsonMarshal(v any) ([]byte, error) { return json.Marshal(v) }

func itoa(n int) string {
	b, _ := json.Marshal(n)
	return string(b)
}

// requestMeter counts API requests per workspace in a sliding minute.
type requestMeter struct {
	mu   sync.Mutex
	seen map[string][]time.Time
}

// admit records one request and reports whether it is within the limit.
func (m *requestMeter) admit(id string, limit int, now time.Time) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.seen == nil {
		m.seen = map[string][]time.Time{}
	}
	recent := m.seen[id][:0]
	for _, at := range m.seen[id] {
		if now.Sub(at) < time.Minute {
			recent = append(recent, at)
		}
	}
	if len(recent) >= limit {
		m.seen[id] = recent
		return false
	}
	m.seen[id] = append(recent, now)
	return true
}

func tooManyRequests(w http.ResponseWriter) {
	w.Header().Set("Retry-After", "60")
	http.Error(w, "This workspace has reached its request limit; try again in a minute", http.StatusTooManyRequests)
}
