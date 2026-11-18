package main

// Human integrity inspection and fail-closed execution response.

import (
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
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

// integrityLocked verifies the whole chain. It requires s.mu; every verification failure stops
// execution, including failures that prevent the incident from being committed. The owner's
// view and recovery always use this full verification.
func (s *server) integrityLocked(w http.ResponseWriter, r *http.Request) {
	s.integrityCheckLocked(w, r, nil, time.Now())
}

const (
	integrityFullSQL = "SELECT id,sequence,committed_at,kind,payload::text,previous_hash,hash FROM events ORDER BY sequence"
	// The first row is the checkpoint already verified: its hash is compared, the rest are verified.
	integrityFromSQL = "SELECT id,sequence,committed_at,kind,payload::text,previous_hash,hash FROM events WHERE sequence >= $1 ORDER BY sequence"
)

// integrityCheckLocked verifies the chain from the start, or, when from is given, only the
// records after that verified checkpoint, as of at. It requires s.mu.
func (s *server) integrityCheckLocked(w http.ResponseWriter, r *http.Request, from *integrityCheckpoint, at time.Time) {
	ctx, cancel := s.storageContext(r.Context())
	defer cancel()
	r = r.WithContext(ctx)
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	if from != nil && from.sequence <= 0 {
		from = nil
	}
	var rows pgx.Rows
	var err error
	if from == nil {
		rows, err = s.db.Query(r.Context(), integrityFullSQL)
	} else {
		rows, err = s.db.Query(r.Context(), integrityFromSQL, from.sequence)
	}
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
	last := integrityCheckpoint{}
	anchorPending := false
	if from != nil {
		previous, count, protectedCount, protectedFormatSeen = from.hash, from.records, from.protected, from.protectedSeen
		last = *from
		anchorPending = true
	}
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
		if anchorPending {
			// The checkpoint record itself: it must still be the record that was verified.
			anchorPending = false
			if sequence != from.sequence || hash != from.hash {
				valid = false
			}
			continue
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
		last.sequence = sequence
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		s.tripIntegrityLocked()
		http.Error(w, "Integrity unverified", 503)
		return
	}
	if anchorPending {
		valid = false // the verified checkpoint record is gone
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
	if valid {
		last.hash, last.records, last.protected, last.protectedSeen = previous, count, protectedCount, protectedFormatSeen
		s.integrityPlan.record(last, from == nil, at)
	}
	reply(w, map[string]any{"valid": valid, "records": count, "metadataProtectedRecords": protectedCount, "legacyMetadataRecords": count - protectedCount, "executionBlocked": s.integrityFault, "protection": "Local chain only; independent immutable archive and rollback anchor still required"})
}
