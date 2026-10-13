package main

import (
	"encoding/json"
	"errors"
	"time"
)

type auditRecordMetadata struct {
	Version     int       `json:"version"`
	ID          string    `json:"id"`
	Sequence    int64     `json:"sequence"`
	CommittedAt time.Time `json:"committedAt"`
}

func protectAuditMetadata(raw []byte, meta auditRecordMetadata) ([]byte, error) {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return nil, errors.New("audit payload must be an object")
	}
	fields["_auditRecord"], _ = json.Marshal(meta)
	return json.Marshal(fields)
}

// New records bind identity/order/time inside their existing canonical payload
// checksum. Legacy records stay byte-for-byte unchanged and explicitly lack
// this protection. Neither version alone can detect prefix rollback.
func verifyAuditMetadata(payload string, id string, sequence int64, at time.Time) (protected, valid bool) {
	var fields map[string]json.RawMessage
	if json.Unmarshal([]byte(payload), &fields) != nil {
		return false, false
	}
	raw, exists := fields["_auditRecord"]
	if !exists {
		return false, true
	}
	var meta auditRecordMetadata
	if json.Unmarshal(raw, &meta) != nil || meta.Version != 1 {
		return true, false
	}
	return true, meta.ID == id && meta.Sequence == sequence && !meta.CommittedAt.IsZero() && meta.CommittedAt.Equal(at)
}
