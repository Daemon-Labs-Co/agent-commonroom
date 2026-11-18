package main

import (
	"context"
	"fmt"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"agentroom/internal/room"
)

var scheduleStart = time.Date(2026, 11, 18, 0, 0, 0, 0, time.UTC)

func TestIntegrityScheduleHourlyIncrementalAndDailyFull(t *testing.T) {
	t.Parallel()
	cfg := integritySettings{incrementalEvery: time.Hour, fullEvery: 24 * time.Hour}
	var st integrityState
	if st.next(scheduleStart, cfg) != integrityFull {
		t.Fatal("the first verification must be full")
	}
	st.record(integrityCheckpoint{sequence: 10, hash: "h"}, true, scheduleStart)
	steps := []struct {
		after time.Duration
		want  integrityMode
	}{
		{time.Minute, integritySkip}, {59 * time.Minute, integritySkip}, {time.Hour, integrityIncremental},
	}
	for _, step := range steps {
		if got := st.next(scheduleStart.Add(step.after), cfg); got != step.want {
			t.Fatalf("after %v: %v, want %v", step.after, got, step.want)
		}
	}
	// Hourly increments continue until a full verification is due after 24 hours.
	now := scheduleStart
	incrementals := 0
	for i := 0; i < 24*60; i++ {
		now = now.Add(time.Minute)
		switch st.next(now, cfg) {
		case integrityIncremental:
			incrementals++
			st.record(*st.checkpoint, false, now)
		case integrityFull:
			if now.Sub(scheduleStart) < 24*time.Hour {
				t.Fatalf("full verification after only %v", now.Sub(scheduleStart))
			}
			if incrementals != 23 {
				t.Fatalf("%d incremental verifications before the daily full one, want 23", incrementals)
			}
			return
		}
	}
	t.Fatal("no full verification within 24 hours")
}

func TestIntegrityScheduleFullIsDueWhenRequestedOrWithoutACheckpoint(t *testing.T) {
	t.Parallel()
	cfg := integritySettings{incrementalEvery: time.Hour, fullEvery: 24 * time.Hour}
	var st integrityState
	st.record(integrityCheckpoint{sequence: 3, hash: "h"}, true, scheduleStart)
	st.fullDue = true
	if st.next(scheduleStart.Add(time.Second), cfg) != integrityFull {
		t.Fatal("a requested full verification was not due")
	}
	st.record(integrityCheckpoint{sequence: 3, hash: "h"}, true, scheduleStart.Add(time.Second))
	if st.fullDue {
		t.Fatal("a full verification did not clear the request")
	}
	st.checkpoint = nil
	if st.next(scheduleStart.Add(2*time.Second), cfg) != integrityFull {
		t.Fatal("no checkpoint must mean a full verification")
	}
}

func TestIntegrityScheduleJitterOnlyBringsChecksForward(t *testing.T) {
	t.Parallel()
	cfg := integritySettings{incrementalEvery: time.Hour, fullEvery: 24 * time.Hour}
	for i := 0; i < 50; i++ {
		full, inc := newIntegrityJitter(cfg)
		if full < 0 || full > 144*time.Minute || inc < 0 || inc > 6*time.Minute {
			t.Fatal(full, inc)
		}
	}
	st := integrityState{jitterIncremental: 5 * time.Minute}
	st.record(integrityCheckpoint{sequence: 1, hash: "h"}, true, scheduleStart)
	if st.next(scheduleStart.Add(54*time.Minute), cfg) != integritySkip || st.next(scheduleStart.Add(55*time.Minute), cfg) != integrityIncremental {
		t.Fatal("jitter did not shift the incremental time")
	}
}

func TestParseIntegritySettings(t *testing.T) {
	t.Parallel()
	env := func(m map[string]string) func(string) string { return func(k string) string { return m[k] } }
	cfg, err := parseIntegritySettings(env(nil))
	if err != nil || cfg.incrementalEvery != time.Hour || cfg.fullEvery != 24*time.Hour {
		t.Fatal(cfg, err)
	}
	cfg, err = parseIntegritySettings(env(map[string]string{"COMMONROOM_INTEGRITY_INCREMENTAL_EVERY": "15m", "COMMONROOM_INTEGRITY_FULL_EVERY": "12h"}))
	if err != nil || cfg.incrementalEvery != 15*time.Minute || cfg.fullEvery != 12*time.Hour {
		t.Fatal(cfg, err)
	}
	for name, m := range map[string]map[string]string{
		"below minimum":       {"COMMONROOM_INTEGRITY_INCREMENTAL_EVERY": "30s"},
		"zero":                {"COMMONROOM_INTEGRITY_FULL_EVERY": "0"},
		"not a duration":      {"COMMONROOM_INTEGRITY_FULL_EVERY": "daily"},
		"negative":            {"COMMONROOM_INTEGRITY_INCREMENTAL_EVERY": "-1h"},
		"too long":            {"COMMONROOM_INTEGRITY_FULL_EVERY": "9999h"},
		"full before inc":     {"COMMONROOM_INTEGRITY_INCREMENTAL_EVERY": "2h", "COMMONROOM_INTEGRITY_FULL_EVERY": "1h"},
		"full before default": {"COMMONROOM_INTEGRITY_FULL_EVERY": "30m"},
	} {
		if _, err := parseIntegritySettings(env(m)); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
}

// --- behaviour against the (mock) database ---

func chainServer(t *testing.T, events int) (*server, *mockDB, *recordingDB) {
	t.Helper()
	db := newMockDB(t)
	seedMessages(db, events, "a")
	rec := &recordingDB{eventDatabase: db}
	s := &server{db: rec}
	s.projection.enabled = true
	s.integritySettings = integritySettings{incrementalEvery: time.Hour, fullEvery: 24 * time.Hour}
	return s, db, rec
}

func (r *recordingDB) count(sql string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := 0
	for _, q := range r.queries {
		if q.sql == sql {
			n++
		}
	}
	return n
}

func (r *recordingDB) fromArgs() []int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []int64{}
	for _, q := range r.queries {
		if q.sql == integrityFromSQL {
			out = append(out, q.args[0].(int64))
		}
	}
	return out
}

func TestScheduledChecksVerifyOnlyWhatIsDue(t *testing.T) {
	t.Parallel()
	s, db, rec := chainServer(t, 5)
	s.checkIntegrityScheduledAt(scheduleStart) // startup: full
	if rec.count(integrityFullSQL) != 1 || s.integrityFault {
		t.Fatal("startup verification was not a clean full verification")
	}
	seedMessages(db, 3, "b")
	s.checkIntegrityScheduledAt(scheduleStart.Add(30 * time.Minute)) // nothing due
	if rec.count(integrityFullSQL) != 1 || len(rec.fromArgs()) != 0 {
		t.Fatal("a check ran before it was due")
	}
	s.checkIntegrityScheduledAt(scheduleStart.Add(time.Hour)) // incremental, from the checkpoint at sequence 5
	if rec.count(integrityFullSQL) != 1 || fmt.Sprint(rec.fromArgs()) != "[5]" || s.integrityFault {
		t.Fatalf("incremental verification: full=%d from=%v fault=%v", rec.count(integrityFullSQL), rec.fromArgs(), s.integrityFault)
	}
	if s.integrityPlan.checkpoint.sequence != 8 || s.integrityPlan.checkpoint.records != 8 {
		t.Fatalf("checkpoint %+v", s.integrityPlan.checkpoint)
	}
	s.checkIntegrityScheduledAt(scheduleStart.Add(24*time.Hour + time.Minute)) // daily full
	if rec.count(integrityFullSQL) != 2 {
		t.Fatal("the daily full verification did not run")
	}
}

func TestIncrementalFindsATamperedNewRecordButFullFindsAnOldOne(t *testing.T) {
	t.Parallel()
	// A new record that was altered after it was written: caught by the next incremental check.
	s, db, _ := chainServer(t, 5)
	s.checkIntegrityScheduledAt(scheduleStart)
	seedMessages(db, 2, "b")
	db.mu.Lock()
	db.events[6].payload = strings.Replace(db.events[6].payload, "hello", "forged", 1)
	db.mu.Unlock()
	s.checkIntegrityScheduledAt(scheduleStart.Add(time.Hour))
	if !s.integrityFault {
		t.Fatal("a tampered new record was not caught by the incremental check")
	}

	// An old record altered in place: the incremental check does not read it, the full one does.
	s, db, _ = chainServer(t, 6)
	s.checkIntegrityScheduledAt(scheduleStart)
	seedMessages(db, 1, "b")
	db.mu.Lock()
	db.events[1].payload = strings.Replace(db.events[1].payload, "hello", "forged", 1)
	db.mu.Unlock()
	s.checkIntegrityScheduledAt(scheduleStart.Add(time.Hour))
	if s.integrityFault {
		t.Fatal("the incremental check read an old record; the test no longer shows the documented window")
	}
	s.checkIntegrityScheduledAt(scheduleStart.Add(24*time.Hour + time.Minute))
	if !s.integrityFault {
		t.Fatal("the daily full verification missed a tampered old record")
	}
}

func TestIncrementalFindsAChangedOrRemovedCheckpointRecord(t *testing.T) {
	t.Parallel()
	for name, mutate := range map[string]func(*mockDB){
		"checkpoint hash changed": func(db *mockDB) { db.events[4].hash = strings.Repeat("0", 64) },
		"checkpoint removed":      func(db *mockDB) { db.events = append(append([]mockEvent{}, db.events[:4]...), db.events[5:]...) },
		"all history removed":     func(db *mockDB) { db.events = nil },
	} {
		s, db, _ := chainServer(t, 5)
		s.checkIntegrityScheduledAt(scheduleStart)
		seedMessages(db, 1, "b")
		db.mu.Lock()
		mutate(db)
		db.mu.Unlock()
		s.checkIntegrityScheduledAt(scheduleStart.Add(time.Hour))
		if !s.integrityFault {
			t.Errorf("%s: not caught", name)
		}
	}
}

func TestWithTheCacheOffEveryCheckIsFull(t *testing.T) {
	t.Parallel()
	if projectionCacheForTests {
		t.Skip("the whole package is running with the cache forced on")
	}
	s, _, rec := chainServer(t, 4)
	s.projection.enabled = false
	for i := 0; i < 3; i++ {
		s.checkIntegrityNow()
	}
	if rec.count(integrityFullSQL) != 3 || len(rec.fromArgs()) != 0 {
		t.Fatalf("full=%d incremental=%v", rec.count(integrityFullSQL), rec.fromArgs())
	}
}

func TestACacheAnchorMismatchRequestsAFullVerificationAtOnce(t *testing.T) {
	t.Parallel()
	s, db, rec := chainServer(t, 4)
	s.projection.onHistoryChanged = s.requestFullIntegrity
	s.checkIntegrityScheduledAt(scheduleStart)
	before := rec.count(integrityFullSQL)
	if _, err := s.load(context.Background()); err != nil {
		t.Fatal(err)
	}
	db.mu.Lock()
	db.events[3].hash = strings.Repeat("0", 64)
	db.mu.Unlock()
	seedMessages(db, 1, "b")
	if _, err := s.load(context.Background()); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for rec.count(integrityFullSQL) == before && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if rec.count(integrityFullSQL) == before {
		t.Fatal("no full verification followed the anchor mismatch")
	}
	var tripped atomic.Bool
	for time.Now().Before(deadline) {
		s.mu.Lock()
		tripped.Store(s.integrityFault)
		s.mu.Unlock()
		if tripped.Load() {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if !tripped.Load() {
		t.Fatal("the changed history was not treated as an integrity failure")
	}
	_ = room.State{}
}

func TestOwnerVerificationIsAlwaysFull(t *testing.T) {
	t.Parallel()
	s, _, rec := chainServer(t, 4)
	s.checkIntegrityScheduledAt(scheduleStart)
	before := rec.count(integrityFullSQL)
	w := httptest.NewRecorder()
	s.integrity(w, httptest.NewRequest("GET", "/api/integrity", nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"valid":true`) || rec.count(integrityFullSQL) != before+1 || len(rec.fromArgs()) != 0 {
		t.Fatalf("%d %s full=%d from=%v", w.Code, w.Body.String(), rec.count(integrityFullSQL), rec.fromArgs())
	}
}
