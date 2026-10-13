package main

// Checksummed append-only event persistence behind the eventDatabase interface.

import (
	"agentroom/internal/room"
	"context"
	"encoding/json"
	"errors"
	"time"
)

type auditChange struct {
	Kind    string
	Payload any
}

func (s *server) append(ctx context.Context, kind string, payload any) error {
	return s.appendBatch(ctx, []auditChange{{Kind: kind, Payload: payload}})
}

// appendBatch commits a whole business operation under the existing audit-chain
// lock. No dispatch may start until it succeeds; failure rolls back every event.
func (s *server) appendBatch(ctx context.Context, changes []auditChange) error {
	if len(changes) == 0 {
		return nil
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(878701)"); err != nil {
		return err
	}
	previous := ""
	err = tx.QueryRow(ctx, "SELECT hash FROM events ORDER BY sequence DESC LIMIT 1").Scan(&previous)
	if err != nil && err.Error() != "no rows in result set" {
		return err
	}
	if expected, ok := ctx.Value(expectedEventCountKey{}).(int64); ok {
		var count int64
		if err = tx.QueryRow(ctx, "SELECT count(*) FROM events").Scan(&count); err != nil {
			return err
		}
		if count != expected {
			return errAuditConflict
		}
	}
	if receipt, ok := ctx.Value(operationReceiptKey{}).(room.OperationReceipt); ok && receipt.Status != 0 {
		changes = append(append([]auditChange(nil), changes...), auditChange{Kind: "operation.receipt", Payload: receipt})
	}
	for _, change := range changes {
		kind, payload := change.Kind, change.Payload
		// Background work is a projection of separate telemetry events, not part of
		// a mutable profile/session snapshot. Do not duplicate it into those records.
		if participant, ok := payload.(room.Participant); ok {
			participant.Background = nil
			payload = participant
		}
		// PostgreSQL canonical JSON text is used for verification and chain calculation.
		raw, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		if actor, ok := ctx.Value(actorContextKey{}).(string); ok && actor != "" {
			var fields map[string]json.RawMessage
			if json.Unmarshal(raw, &fields) != nil || fields == nil {
				return errors.New("audit payload must be an object")
			}
			fields["performedBy"], _ = json.Marshal(actor)
			raw, err = json.Marshal(fields)
			if err != nil {
				return err
			}
		}
		var canonical string
		var sequence int64
		if err = tx.QueryRow(ctx, "SELECT nextval('public.events_sequence_seq')").Scan(&sequence); err != nil {
			return err
		}
		meta := auditRecordMetadata{Version: 1, ID: id(), Sequence: sequence, CommittedAt: time.Now().UTC().Truncate(time.Microsecond)}
		raw, err = protectAuditMetadata(raw, meta)
		if err != nil {
			return err
		}
		if err = tx.QueryRow(ctx, "SELECT $1::jsonb::text", string(raw)).Scan(&canonical); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, "INSERT INTO events(sequence,id,kind,payload,committed_at,previous_hash,hash) VALUES($1,$2,$3,$4::jsonb,$5,$6,$7)", meta.Sequence, meta.ID, kind, canonical, meta.CommittedAt, previous, checksum(previous, kind, canonical))
		if err != nil {
			return err
		}

		previous = checksum(previous, kind, canonical)
	}
	return tx.Commit(ctx)
}
