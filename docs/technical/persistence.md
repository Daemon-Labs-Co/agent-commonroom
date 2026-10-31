# Events, state projection and atomic operations

## Input

Typed command payload, authenticated actor attribution, expected revision/event count and bounded persistence context.

## Output

Atomic append-only event batch, operation receipt, replayed state or rollback/conflict/refusal.

## Functionality

appendBatch serializes audit-chain writes using a PostgreSQL advisory transaction lock. Payloads use canonical JSON with protected metadata. load currently replays the whole event history; a projection cache/checkpoint is not implemented. Automated DB interactions use mocks.

## Properties and state changes

Committed events append history; ordinary edits/deletes of retained events are rejected. Supported revisions/receipts change atomically. Integrity failures latch execution off; privileged database rewrite/rollback resistance requires independent evidence.

Source: [event_store.go](../../cmd/server/event_store.go) · [state_projection.go](../../cmd/server/state_projection.go) · [atomic_operations.go](../../cmd/server/atomic_operations.go).
