package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// tripIntegrityLocked is independent of storage availability. Its latch must
// never clear solely because a later probe succeeds or a UI flag is disabled.
func (s *server) tripIntegrityLocked() {
	s.integrityFault = true
	if s.integrityStopPath != "" {
		file, err := os.OpenFile(s.integrityStopPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err == nil {
			_, err = file.WriteString("Audit verification or required persistence failed. Explicit owner recovery required.\n")
			if err == nil {
				err = file.Sync()
			}
			if closeErr := file.Close(); err == nil {
				err = closeErr
			}
			if err == nil {
				err = syncParentDirectory(s.integrityStopPath)
			}
		}
		if os.IsExist(err) {
			err = syncParentDirectory(s.integrityStopPath)
		}
		if err != nil {
			log.Print("Integrity stop marker could not be persisted; do not restart until storage is repaired")
		}
	}
	for _, cancel := range s.cancels {
		cancel()
	}
	for _, cancel := range s.workflowLive {
		cancel()
	}
}

// The marker preserves an outage stop across restarts even when the database
// cannot record an incident. It is operational state, not an independent archive.
func (s *server) restoreIntegrityStop() {
	if s.integrityStopPath == "" {
		return
	}
	if _, err := os.Lstat(s.integrityStopPath); err == nil || !os.IsNotExist(err) {
		s.integrityFault = true
	}
}

// recoverIntegrityLocked never repairs or rewrites history. It records owner
// approval only after a fresh valid chain and verified worker shutdown.
func (s *server) recoverIntegrityLocked(w http.ResponseWriter, r *http.Request) {
	if requestActor(r) != ownerID {
		http.Error(w, "Owner permission required", 403)
		return
	}
	if !s.validCSRF(r) {
		http.Error(w, "Invalid request token", 403)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 2048)
	var req struct {
		Reason string `json:"reason"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil || strings.TrimSpace(req.Reason) == "" || len(req.Reason) > 500 {
		http.Error(w, "Record a recovery reason of at most 500 characters", 400)
		return
	}
	for _, active := range s.active {
		if active {
			http.Error(w, "Wait for cancellation and worker shutdown first", 409)
			return
		}
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	rec := &bufferResponse{header: http.Header{}, body: &bytes.Buffer{}}
	s.integrityLocked(rec, (&http.Request{Method: "GET"}).WithContext(ctx))
	var result struct {
		Valid bool `json:"valid"`
	}
	if rec.status >= 400 || json.Unmarshal(rec.body.Bytes(), &result) != nil || !result.Valid {
		http.Error(w, "A valid audit chain is required; history cannot be rewritten through recovery", 409)
		return
	}
	probe := s.recoveryIdle
	if probe == nil {
		probe = remoteWorkerIdle
	}
	idle, err := probe(ctx, "ssh://"+os.Getenv("COMMONROOM_REMOTE_WORKER_ALIAS"))
	if err != nil || !idle {
		http.Error(w, "Approved worker shutdown could not be verified", 503)
		return
	}
	if s.append(ctx, "incident.recovered", map[string]string{"actorId": ownerID, "reason": strings.TrimSpace(req.Reason)}) != nil {
		http.Error(w, "Recovery audit failed; execution remains stopped", 503)
		return
	}
	if s.integrityStopPath != "" {
		if err := os.Remove(s.integrityStopPath); err != nil && !os.IsNotExist(err) {
			http.Error(w, "Recovery marker could not be cleared; execution remains stopped", 503)
			return
		} else if err == nil {
			if syncParentDirectory(s.integrityStopPath) != nil {
				s.tripIntegrityLocked()
				http.Error(w, "Recovery marker removal could not be persisted; execution remains stopped", 503)
				return
			}
		}
	}
	s.integrityFault = false
	reply(w, map[string]any{"recovered": true, "jobsRestarted": false})
}

func (s *server) checkIntegrityNow() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.mu.Lock()
	defer s.mu.Unlock()
	rec := &bufferResponse{header: http.Header{}, body: &bytes.Buffer{}}
	s.integrityLocked(rec, (&http.Request{Method: "GET"}).WithContext(ctx))
}

func (s *server) watchIntegrity(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.checkIntegrityNow()
		}
	}
}
