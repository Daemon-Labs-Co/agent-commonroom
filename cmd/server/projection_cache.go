package main

// An optional incremental cache of the replayed state (COMMONROOM_PROJECTION_CACHE=on).
// Instead of reading and replaying the whole history for every request, the cache keeps the
// projection and the position it has reached, reads only the events after that position and
// applies them. The recorded history stays the only source of truth: the cache is rebuilt
// from scratch whenever it cannot prove it is still a prefix of that history.

import (
	"agentroom/internal/room"
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

// The first row returned is the position the cache has already applied (the anchor); it is
// compared, not applied. Everything after it is new.
const projectionDeltaSQL = "SELECT sequence,kind,payload,hash FROM events WHERE sequence >= $1 ORDER BY sequence"

var errHistoryChanged = errors.New("recorded history changed while it was being read")

// projectionCacheForTests lets the whole test suite run with the cache on. It is set once,
// by TestMain, and never changed by the program.
var projectionCacheForTests = false

type projectionCache struct {
	enabled bool
	// onHistoryChanged is called (without the lock held) when the cache finds that the history
	// it had applied is no longer there or no longer the same.
	onHistoryChanged func()

	mu   sync.Mutex
	p    *projection
	seq  int64
	hash string
}

func (c *projectionCache) active() bool { return c.enabled || projectionCacheForTests }

func (c *projectionCache) reset() { c.p, c.seq, c.hash = nil, 0, "" }

// loadCached returns the same state loadAt would, reading only events it has not applied.
func (s *server) loadCached(ctx context.Context, now time.Time) (room.State, error) {
	ctx, cancel := s.storageContext(ctx)
	defer cancel()
	c := &s.projection
	changed := false
	c.mu.Lock()
	state, err := s.refreshLocked(ctx, now, &changed)
	c.mu.Unlock()
	if changed && c.onHistoryChanged != nil {
		c.onHistoryChanged()
	}
	return state, err
}

func (s *server) refreshLocked(ctx context.Context, now time.Time, changed *bool) (room.State, error) {
	c := &s.projection
	for attempt := 0; ; attempt++ {
		if c.p == nil {
			c.p, c.seq, c.hash = newProjection(), 0, ""
		}
		rows, err := s.db.Query(ctx, projectionDeltaSQL, c.seq)
		if err != nil {
			return newProjection().st, err // nothing was applied; the cache is still valid
		}
		anchored := c.seq == 0
		stale := false
		var failure error
		for rows.Next() {
			if failure = ctx.Err(); failure != nil {
				break
			}
			var seq int64
			var kind, hash string
			var data []byte
			if failure = rows.Scan(&seq, &kind, &data, &hash); failure != nil {
				break
			}
			if !anchored {
				anchored = true
				if seq != c.seq || hash != c.hash {
					stale = true
					break
				}
				continue
			}
			if failure = c.p.apply(kind, data); failure != nil {
				break
			}
			c.seq, c.hash = seq, hash
		}
		if failure == nil && !stale {
			failure = rows.Err()
		}
		rows.Close()
		switch {
		case failure != nil:
			partial := c.p.st
			c.reset() // a half-applied projection must never be reused
			return partial, failure
		case !anchored || stale:
			// The history the cache had applied is missing or different: start again from nothing.
			*changed = true
			log.Printf("Projection cache: recorded history no longer matches what was applied (sequence %d); rebuilding from the full history", c.seq)
			c.reset()
			if attempt == 0 {
				continue
			}
			return newProjection().st, errHistoryChanged
		}
		return c.p.snapshot(now), nil
	}
}
