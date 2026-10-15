package main

// Human integrity inspection and fail-closed execution response.

import (
	"net/http"
	"time"
)

func (s *server) integrity(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if r.Method == "POST" {
		s.recoverIntegrityLocked(w, r)
		return
	}
	s.integrityLocked(w, r)
}

// integrityLocked requires s.mu; every verification failure stops execution,
// including failures that prevent the incident from being committed.
func (s *server) integrityLocked(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := s.storageContext(r.Context())
	defer cancel()
	r = r.WithContext(ctx)
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	rows, err := s.db.Query(r.Context(), "SELECT id,sequence,committed_at,kind,payload::text,previous_hash,hash FROM events ORDER BY sequence")
	if err != nil {
		s.tripIntegrityLocked()
		http.Error(w, "Integrity unverified", 503)
		return
	}
	previous := ""
	count := 0
	valid := true
	protectedCount := 0
	protectedFormatSeen := false
	for rows.Next() {
		if ctx.Err() != nil {
			rows.Close()
			s.tripIntegrityLocked()
			http.Error(w, "Integrity verification timed out", 503)
			return
		}
		var kind, payload, prev, hash string
		var eventID string
		var sequence int64
		var committedAt time.Time
		if err = rows.Scan(&eventID, &sequence, &committedAt, &kind, &payload, &prev, &hash); err != nil {
			rows.Close()
			s.tripIntegrityLocked()
			http.Error(w, "Integrity unverified", 503)
			return
		}
		if prev != previous || hash != checksum(prev, kind, payload) {
			valid = false
		}
		protected, metadataValid := verifyAuditMetadata(payload, eventID, sequence, committedAt)
		if protected {
			protectedFormatSeen = true
			protectedCount++
		} else if protectedFormatSeen {
			valid = false
		}
		if !metadataValid {
			valid = false
		}
		previous = hash
		count++
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		s.tripIntegrityLocked()
		http.Error(w, "Integrity unverified", 503)
		return
	}
	if !valid {
		s.tripIntegrityLocked()
	}
	st, e := s.load(r.Context())
	if e != nil {
		s.tripIntegrityLocked()
		http.Error(w, "Integrity state unverified", 503)
		return
	}
	if st.Incident {
		s.tripIntegrityLocked()
	}
	if s.integrityFault && !st.Incident {
		if s.append(r.Context(), "incident.opened", map[string]string{"reason": "audit verification failed or was unavailable; owner review required"}) != nil {
			http.Error(w, "Incident audit unavailable; execution remains stopped", 503)
			return
		}
	}
	reply(w, map[string]any{"valid": valid, "records": count, "metadataProtectedRecords": protectedCount, "legacyMetadataRecords": count - protectedCount, "executionBlocked": s.integrityFault, "protection": "Local chain only; independent immutable archive and rollback anchor still required"})
}
