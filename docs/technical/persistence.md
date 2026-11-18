# Events, state projection and atomic operations

## Input

Typed command payload, authenticated actor attribution, expected revision/event count and bounded persistence context. Settings (read at startup): `COMMONROOM_PROJECTION_CACHE` (`off` default, `on`), `COMMONROOM_INTEGRITY_INCREMENTAL_EVERY`, `COMMONROOM_INTEGRITY_FULL_EVERY`, `COMMONROOM_WORKSPACE_IDLE_MINUTES` (the last three take effect only with the cache on).

## Output

Atomic append-only event batch, operation receipt, replayed state or rollback/conflict/refusal.

## Functionality

appendBatch serializes audit-chain writes using a PostgreSQL advisory transaction lock. Payloads use canonical JSON with protected metadata. load replays the event history through one reducer per event kind (`projection.go`, `projection_*.go`) and then derives the views the application shows (agent status, elapsed time, attention, suggestions). By default it replays the whole history on every call. Automated DB interactions use mocks.

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
across workspaces (see ADR 0004). Upgrade order is part of the contract: deploy the controller
first, then stop it, migrate, grant and start it again (see the deployment guide); an older
controller, or this controller left running across the migration, cannot insert once 002 applies.

Incremental projection cache (`COMMONROOM_PROJECTION_CACHE=on`, default `off`). The server keeps the replayed projection and the `(sequence, hash)` it has reached. Each load reads only the events from that position (`SELECT ... WHERE sequence >= $1`); the first row returned is the position already applied and is compared, not applied. If it is missing or its hash differs (history purged, restored or altered), the cache logs the mismatch, asks for a full integrity verification at once, and rebuilds from the full history. An event that cannot be applied empties the cache and fails the call exactly as the full replay does; a failed read leaves the cache as it was. Each load returns a deep copy of the replayed state with the derived views computed on the copy, so callers can change it freely and later events continue from the unchanged projection. State size, not history size, now bounds the cost of a request.

With the cache on, chain verification is scheduled instead of running in full every 30 seconds: only the records after the last verified checkpoint every `COMMONROOM_INTEGRITY_INCREMENTAL_EVERY` (default 1h), and the whole chain at startup, on request from the cache, and every `COMMONROOM_INTEGRITY_FULL_EVERY` (default 24h; 1m to 720h, never shorter than the incremental interval). Each workspace's schedule is shifted earlier by up to a tenth of the interval. The owner's integrity view and recovery always verify the whole chain. Detection windows with the cache on: a forged or altered new record is caught by the next incremental check, and by the next write (every append reads the last stored hash under the audit lock) or the next cache refresh if it is the latest record; an altered older record is caught by the next full verification, so up to the full interval. With the cache off the 30-second full verification is unchanged.

In multi mode with the cache on, a workspace nobody has used for `COMMONROOM_WORKSPACE_IDLE_MINUTES` (default 30, 1 to 10080, 0 never) is stopped and forgotten and rebuilt from history on its next request. It is never released while a request is running, while agent or workflow work is running, or when its sessions are not saved to disk.

## Properties and state changes

Committed events append history; ordinary edits/deletes of retained events are rejected. Supported revisions/receipts change atomically. The cache holds no state of its own beyond the replayed history and is discarded on restart. Integrity failures latch execution off; privileged database rewrite/rollback resistance requires independent evidence.

Source: [event_store.go](../../cmd/server/event_store.go) · [state_projection.go](../../cmd/server/state_projection.go) · [atomic_operations.go](../../cmd/server/atomic_operations.go) · [projection.go](../../cmd/server/projection.go) · [projection_cache.go](../../cmd/server/projection_cache.go) · [integrity_schedule.go](../../cmd/server/integrity_schedule.go) · [workspace_idle.go](../../cmd/server/workspace_idle.go).
