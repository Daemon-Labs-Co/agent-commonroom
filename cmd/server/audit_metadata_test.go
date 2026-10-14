package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestNewAuditMetadataDetectsIdentityOrderAndTimeTampering(t *testing.T) {
	t.Parallel()
	at := time.Date(2026, 1, 1, 0, 0, 0, 123456000, time.UTC)
	meta := auditRecordMetadata{Version: 1, ID: "synthetic-event", Sequence: 7, CommittedAt: at}
	raw, err := protectAuditMetadata([]byte(`{"value":"Synthetic","_auditRecord":{"id":"client-forged"}}`), meta)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		id    string
		seq   int64
		at    time.Time
		valid bool
	}{{meta.ID, 7, at, true}, {"changed", 7, at, false}, {meta.ID, 8, at, false}, {meta.ID, 7, at.Add(time.Second), false}} {
		protected, valid := verifyAuditMetadata(string(raw), tc.id, tc.seq, tc.at)
		if !protected || valid != tc.valid {
			t.Fatal(tc, protected, valid)
		}
	}
	if p, v := verifyAuditMetadata(`{"value":"Legacy synthetic"}`, "legacy", 1, at); p || !v {
		t.Fatal("legacy data rejected or claimed protected")
	}
}

func TestNewAppendOverwritesClaimedActorAndChecksMetadata(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	ctx := context.WithValue(context.Background(), actorContextKey{}, ownerID)
	if err := s.append(ctx, "company.updated", map[string]any{"name": "Synthetic", "performedBy": "forged", "_auditRecord": map[string]any{"id": "forged"}}); err != nil {
		t.Fatal(err)
	}
	entry := db.snapshot()[0]
	var fields map[string]json.RawMessage
	_ = json.Unmarshal([]byte(entry.payload), &fields)
	if string(fields["performedBy"]) != `"`+ownerID+`"` || strings.Contains(string(fields["_auditRecord"]), "forged") {
		t.Fatal("client forged audit provenance")
	}
	if p, v := verifyAuditMetadata(entry.payload, entry.id, entry.sequence, entry.at); !p || !v {
		t.Fatal("append did not protect metadata")
	}
	db.events[0].at = entry.at.Add(time.Hour)
	s.checkIntegrityNow()
	if !s.integrityFault {
		t.Fatal("metadata tampering did not stop execution")
	}
}

func TestProtectedAuditFormatCannotDowngradeToLegacy(t *testing.T) {
	t.Parallel()
	s, db, h := mockServer(t)
	db.seedLegacy("message.created", map[string]string{"text": "Synthetic historical prefix"})
	if err := s.append(context.Background(), "company.updated", map[string]string{"name": "Synthetic protected"}); err != nil {
		t.Fatal(err)
	}
	// Correctly linked/hash-consistent data with missing metadata must still fail.
	db.seedLegacy("message.created", map[string]string{"text": "Synthetic downgrade"})
	active, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.cancels["synthetic"] = cancel
	w := mockRequest(h, "GET", "/api/integrity", "")
	var result struct {
		Valid bool `json:"valid"`
	}
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &result) != nil || result.Valid || !s.integrityFault || active.Err() == nil {
		t.Fatal("audit format downgrade accepted", w.Code, w.Body.String())
	}
	s.recoveryIdle = func(context.Context, string) (bool, error) {
		t.Error("worker queried before invalid audit rejected")
		return true, nil
	}
	w = mockRequest(h, "POST", "/api/integrity", `{"reason":"Synthetic owner attempt"}`)
	if w.Code != 409 || !s.integrityFault {
		t.Fatal("owner bypassed format barrier", w.Code)
	}
}
func TestLegacyPrefixTransitionsToProtectedAuditWithoutRewriting(t *testing.T) {
	t.Parallel()
	s, db, h := mockServer(t)
	db.seedLegacy("message.created", map[string]string{"text": "Synthetic historical prefix"})
	original := db.snapshot()[0]
	if err := s.append(context.Background(), "company.updated", map[string]string{"name": "Synthetic protected"}); err != nil {
		t.Fatal(err)
	}
	db.seed("message.created", map[string]string{"text": "Synthetic new fixture"})
	w := mockRequest(h, "GET", "/api/integrity", "")
	var result struct {
		Valid     bool `json:"valid"`
		Protected int  `json:"metadataProtectedRecords"`
		Legacy    int  `json:"legacyMetadataRecords"`
	}
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &result) != nil || !result.Valid || result.Protected != 2 || result.Legacy != 1 || db.snapshot()[0] != original {
		t.Fatal("legitimate transition failed", w.Code, w.Body.String())
	}
}
