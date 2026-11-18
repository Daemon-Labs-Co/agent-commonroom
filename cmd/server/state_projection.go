package main

// Immutable-event replay into current application state; no persistence or provider calls.

import (
	"agentroom/internal/room"
	"context"
	"time"
)

// load replays every recorded event into the current state.
func (s *server) load(ctx context.Context) (room.State, error) {
	if s.projection.active() {
		return s.loadCached(ctx, time.Now().UTC())
	}
	return s.loadAt(ctx, time.Now().UTC())
}

// loadAt is load with the clock as a parameter.
func (s *server) loadAt(ctx context.Context, now time.Time) (room.State, error) {
	ctx, cancel := s.storageContext(ctx)
	defer cancel()
	p := newProjection()
	rows, err := s.db.Query(ctx, "SELECT kind,payload FROM events ORDER BY sequence")
	if err != nil {
		return p.st, err
	}
	defer rows.Close()
	for rows.Next() {
		if err := ctx.Err(); err != nil {
			return p.st, err
		}
		var kind string
		var data []byte
		if err = rows.Scan(&kind, &data); err != nil {
			return p.st, err
		}
		if err = p.apply(kind, data); err != nil {
			return p.st, err
		}
	}
	if err := ctx.Err(); err != nil {
		return p.st, err
	}
	p.finish(&p.st, now)
	return p.st, rows.Err()
}
