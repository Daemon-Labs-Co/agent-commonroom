package main

// With the incremental projection cache on, the audit chain is verified on a schedule instead
// of in full every 30 seconds: only the records added since the last verified checkpoint every
// hour, and the whole chain at startup and every 24 hours. Both intervals are settings. The
// recorded history is protected in between by two checks that do not wait for the schedule:
// every write reads the last stored hash under the audit lock, and the cache compares the
// checkpoint record it has applied on each refresh and asks for a full verification at once
// if it differs. With the cache off the previous behaviour (a full check every 30 seconds)
// is unchanged.

import (
	"bytes"
	"context"
	"errors"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"
)

const (
	defaultIntegrityIncrementalEvery = time.Hour
	defaultIntegrityFullEvery        = 24 * time.Hour
	minimumIntegrityInterval         = time.Minute
	scheduledFullVerificationLimit   = 60 * time.Second
)

// integrityCheckpoint is a verified position in the chain, with the running counts and format
// state needed to continue verifying from it.
type integrityCheckpoint struct {
	sequence      int64
	hash          string
	records       int
	protected     int
	protectedSeen bool
}

type integritySettings struct {
	incrementalEvery, fullEvery time.Duration
}

type integrityState struct {
	checkpoint      *integrityCheckpoint
	lastFull        time.Time
	lastIncremental time.Time
	fullDue         bool
	// jitter spreads the schedules of many workspaces so they do not all verify together.
	jitterFull, jitterIncremental time.Duration
}

type integrityMode int

const (
	integritySkip integrityMode = iota
	integrityIncremental
	integrityFull
)

func (m integrityMode) String() string {
	return [...]string{"skip", "incremental", "full"}[m]
}

// record notes a successful verification.
func (st *integrityState) record(checkpoint integrityCheckpoint, full bool, now time.Time) {
	st.checkpoint = &checkpoint
	st.lastIncremental = now
	if full {
		st.lastFull = now
		st.fullDue = false
	}
}

func (cfg integritySettings) withDefaults() integritySettings {
	if cfg.incrementalEvery <= 0 {
		cfg.incrementalEvery = defaultIntegrityIncrementalEvery
	}
	if cfg.fullEvery <= 0 {
		cfg.fullEvery = defaultIntegrityFullEvery
	}
	return cfg
}

// next says which verification is due. A full verification is due when none has run, when a
// checkpoint is missing, when one was requested, or when the full interval has passed.
func (st *integrityState) next(now time.Time, cfg integritySettings) integrityMode {
	cfg = cfg.withDefaults()
	if st.checkpoint == nil || st.lastFull.IsZero() || st.fullDue || !now.Before(st.lastFull.Add(cfg.fullEvery-st.jitterFull)) {
		return integrityFull
	}
	if !now.Before(st.lastIncremental.Add(cfg.incrementalEvery - st.jitterIncremental)) {
		return integrityIncremental
	}
	return integritySkip
}

// newJitter returns up to a tenth of each interval, taken off it.
func newIntegrityJitter(cfg integritySettings) (full, incremental time.Duration) {
	cfg = cfg.withDefaults()
	return time.Duration(rand.Int64N(int64(cfg.fullEvery/10) + 1)), time.Duration(rand.Int64N(int64(cfg.incrementalEvery/10) + 1))
}

// parseIntegritySettings reads COMMONROOM_INTEGRITY_INCREMENTAL_EVERY and
// COMMONROOM_INTEGRITY_FULL_EVERY (Go durations such as 1h or 90m).
func parseIntegritySettings(getenv func(string) string) (integritySettings, error) {
	cfg := integritySettings{incrementalEvery: defaultIntegrityIncrementalEvery, fullEvery: defaultIntegrityFullEvery}
	for _, v := range []struct {
		name string
		dst  *time.Duration
	}{{"COMMONROOM_INTEGRITY_INCREMENTAL_EVERY", &cfg.incrementalEvery}, {"COMMONROOM_INTEGRITY_FULL_EVERY", &cfg.fullEvery}} {
		raw := getenv(v.name)
		if raw == "" {
			continue
		}
		d, err := time.ParseDuration(raw)
		if err != nil || d < minimumIntegrityInterval || d > 30*24*time.Hour {
			return cfg, errors.New(v.name + " must be a duration from 1m to 720h, such as 1h")
		}
		*v.dst = d
	}
	if cfg.fullEvery < cfg.incrementalEvery {
		return cfg, errors.New("COMMONROOM_INTEGRITY_FULL_EVERY (" + strconv.Quote(cfg.fullEvery.String()) + ") must not be shorter than COMMONROOM_INTEGRITY_INCREMENTAL_EVERY")
	}
	return cfg, nil
}

// checkIntegrityScheduledAt runs whichever verification is due at now, if any.
func (s *server) checkIntegrityScheduledAt(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	mode := s.integrityPlan.next(now, s.integritySettings)
	if mode == integritySkip {
		return
	}
	limit := 5 * time.Second
	var from *integrityCheckpoint
	if mode == integrityFull {
		limit = scheduledFullVerificationLimit
	} else {
		from = s.integrityPlan.checkpoint
	}
	ctx, cancel := context.WithTimeout(context.Background(), limit)
	defer cancel()
	rec := &bufferResponse{header: http.Header{}, body: &bytes.Buffer{}}
	s.integrityCheckLocked(rec, (&http.Request{Method: "GET"}).WithContext(ctx), from, now)
}

// requestFullIntegrity asks for a full verification now; it never blocks the caller, which
// may hold the server lock.
func (s *server) requestFullIntegrity() {
	go func() {
		s.mu.Lock()
		s.integrityPlan.fullDue = true
		s.mu.Unlock()
		s.checkIntegrityScheduledAt(time.Now())
	}()
}
