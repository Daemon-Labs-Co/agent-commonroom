# Atomic operations and safe retries (Free)

A business operation must either save every related event or save none of them.
The event store now supports a checksummed `appendBatch` transaction under its
existing PostgreSQL advisory lock. No model dispatch starts before its entire
intent operation commits.

## What is protected

- Agent connection/participant creation and history-attachment records.
- Broadcast/direct message, activity and every target's Working state.
- Final agent output/error and session state.
- Workflow attempt intent, activity and agent Working state.
- Disclosure policies/memory revisions and native session resets.
- User event writes and workflow start/gate/cancel requests with retry receipts.

Company, room, connection and participant profile updates now require the current
`revision`; successful edits increment it. Tasks, memory, workflows and the
organization chart retain their existing revision checks. Legacy objects begin
at revision zero. The UI sends the saved revision; API clients must do the same.
A stale edit returns **409** and writes nothing. Reload and review the change;
do not blindly resubmit a stale form.

## HTTP retries

Send `Idempotency-Key` (8–80 characters, no whitespace) on POST `/api/events`,
`/api/agents` for create/attach/send/broadcast, or `/api/workflow-runs`. Use the
same key only when retrying the same operation with identical input.

The binding is hashed with authenticated human ID, room and endpoint. A committed
response receipt is written in the same transaction as its business events.
Repeats return the recorded JSON/status and `Idempotent-Replay: true`; they never
start another agent or workflow. Reusing the key with different input returns
409. Failed transactions leave no receipt and may be retried. Authentication,
room access, CSRF and request limits still apply to retries.

The browser assigns a UUID per API invocation and reuses it across its existing
CSRF refresh retry. A separate click is a new operation. There is no automatic
network retry or hidden replay of agent work. External clients should retain the
key when a response is lost. Receipts remain effective while their immutable
history records are retained; authorized retention purge can remove old receipts.

## Concurrency and limits

User operation commits compare the loaded event count under the database audit
lock. Another committed event causes 409 rather than writing against stale state.
This is intentionally conservative: even an unrelated background event may
require a reload. It is not a substitute for per-aggregate database projections.
No new schema or runtime privileges are required.

The supported deployment remains **one controller**. Distributed workflow leases
and multiple-controller ownership remain separate work. A database transaction
cannot undo an external provider call, fork, file write, or completed model turn.
A crash after intent commit may leave an interrupted job requiring the existing
owner reconciliation flow; it never authorizes silent replay. Worker job fencing
remains independently mandatory.

These correctness protections have no disable flag: turning off an optional
feature cannot permit partial writes, stale overwrites or duplicate dispatch.
The optional workflow runner's existing `workflowExecution` flag still controls
whether new runs may start; it does not disable receipts or transaction safeguards.

## Technical boundaries and tests

`auditChange` is a typed event payload entry. `appendBatch` normalizes participant
telemetry, stamps authenticated attribution and protected metadata, canonicalizes
each payload through PostgreSQL, advances the chain within the transaction and
commits once. Rollback includes any response receipt. Sequence gaps after rollback
are valid; contiguous sequence numbers are not required across transactions.
`prepareOperation` validates request keys and resolves prior receipts;
`withOperationResponse` attaches the safe response to the commit context.

In-memory database mocks support staged multi-event transactions, injected insert
and commit failures, canonical JSON, and concurrent state checks. Tests exercise
rollback at each batch position, receipt failure, replay/conflicting input,
retry after failure, stale entity edits, audit-chain continuity and CSRF retry
key reuse. Tests never access PostgreSQL or launch provider agents.
