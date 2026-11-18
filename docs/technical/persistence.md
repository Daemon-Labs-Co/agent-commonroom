# Events, state projection and atomic operations

## Input

Typed command payload, authenticated actor attribution, expected revision/event count and bounded persistence context.

## Output

Atomic append-only event batch, operation receipt, replayed state or rollback/conflict/refusal.

## Functionality

appendBatch serializes audit-chain writes using a PostgreSQL advisory transaction lock. Payloads use canonical JSON with protected metadata. load currently replays the whole event history; a projection cache/checkpoint is not implemented. Automated DB interactions use mocks.

Workspace storage (`migrations/002.sql`, applied by `--migrate` after `001.sql`) adds
`events.workspace_id` and forces every read and write through row-level security keyed
to the transaction setting `app.workspace_id`. The serving process wraps its database
in a scoped database that begins each transaction with `set_config('app.workspace_id',
$1, true)`; the setting is local to the transaction. With no setting, or an empty one,
no row is visible and an insert fails. The hash chain, last-hash lookup, count check,
load and integrity verification therefore operate on one workspace's rows with
unchanged SQL, and the append lock is keyed by workspace. A workspace id is `default`
or a lowercase DNS label of up to 63 characters; the table also enforces the format.
Startup refuses to serve once the column exists unless row-level security is enabled and
the runtime role neither owns `events` nor holds SUPERUSER or BYPASSRLS. Before the
migration the previous single-history layout is served unchanged, and existing events
become workspace `default` and keep verifying. `scripts/check-workspace-isolation.sql`
proves the database behaviour against a scratch database. Sequence numbers are shared
across workspaces (see ADR 0004).

## Properties and state changes

Committed events append history; ordinary edits/deletes of retained events are rejected. Supported revisions/receipts change atomically. Integrity failures latch execution off; privileged database rewrite/rollback resistance requires independent evidence.

Source: [event_store.go](../../cmd/server/event_store.go) · [state_projection.go](../../cmd/server/state_projection.go) · [atomic_operations.go](../../cmd/server/atomic_operations.go).
