package main

// Idle eviction (multi mode, only with COMMONROOM_PROJECTION_CACHE=on): a workspace nobody has
// used for COMMONROOM_WORKSPACE_IDLE_MINUTES is stopped and forgotten, which frees its cached
// projection and its integrity watcher. The next request rebuilds it from the recorded history.
// A workspace is never released while a request is running, while agent or workflow work is
// running, or when its sessions are not saved to disk (people would be signed out).

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"
)

const (
	defaultWorkspaceIdle = 30 * time.Minute
	minimumWorkspaceIdle = time.Minute
	maximumWorkspaceIdle = 7 * 24 * time.Hour
)

// quiet reports whether the workspace can be released without losing anything that is not
// stored: no running turns, no cancellable work, and sessions that survive a rebuild.
func (rt *workspaceRuntime) quiet() bool {
	s := rt.s
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, running := range s.active {
		if running {
			return false
		}
	}
	return len(s.cancels) == 0 && len(s.workflowLive) == 0 && s.auth != nil && s.auth.sessionPath != ""
}

// evictIdle releases every built workspace that has been idle for at least idle and is quiet.
// It returns the ids it released.
func (h *workspaceHost) evictIdle(now time.Time, idle time.Duration) []string {
	if h.tenancy.Mode != tenancyMulti || idle <= 0 {
		return nil
	}
	h.mu.Lock()
	candidates := []string{}
	for id, e := range h.entries {
		if e.runtime != nil && e.inflight == 0 && !e.lastUsed.IsZero() && now.Sub(e.lastUsed) >= idle {
			candidates = append(candidates, id)
		}
	}
	h.mu.Unlock()
	released := []string{}
	for _, id := range candidates {
		h.mu.Lock()
		e := h.entries[id]
		stillIdle := e != nil && e.runtime != nil && e.inflight == 0 && now.Sub(e.lastUsed) >= idle
		var rt *workspaceRuntime
		if stillIdle {
			rt = e.runtime
		}
		h.mu.Unlock()
		if rt == nil || !rt.quiet() {
			continue
		}
		h.mu.Lock()
		// Checked again under the lock that routes requests: it was used or entered meanwhile.
		if cur := h.entries[id]; cur == e && cur.inflight == 0 && now.Sub(cur.lastUsed) >= idle {
			delete(h.entries, id)
			delete(h.missing, id)
			h.mu.Unlock()
			rt.stop()
			released = append(released, id)
			continue
		}
		h.mu.Unlock()
	}
	return released
}

// runIdleJanitor releases idle workspaces until ctx ends.
func (h *workspaceHost) runIdleJanitor(ctx context.Context, idle time.Duration) {
	interval := idle / 4
	if interval < 30*time.Second {
		interval = 30 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.evictIdle(h.now(), idle)
		}
	}
}

// parseWorkspaceIdle reads COMMONROOM_WORKSPACE_IDLE_MINUTES: 1 to 10080, or 0 to never release.
func parseWorkspaceIdle(raw string) (time.Duration, error) {
	if raw == "" {
		return defaultWorkspaceIdle, nil
	}
	minutes, err := parsePositiveMinutes(raw)
	if err != nil || (minutes != 0 && (time.Duration(minutes)*time.Minute < minimumWorkspaceIdle || time.Duration(minutes)*time.Minute > maximumWorkspaceIdle)) {
		return 0, errWorkspaceIdle
	}
	return time.Duration(minutes) * time.Minute, nil
}

var errWorkspaceIdle = errors.New("COMMONROOM_WORKSPACE_IDLE_MINUTES must be a whole number of minutes from 1 to 10080, or 0 to never release idle workspaces")

func parsePositiveMinutes(raw string) (int, error) {
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 || strings.TrimSpace(raw) != raw {
		return 0, errWorkspaceIdle
	}
	return n, nil
}
