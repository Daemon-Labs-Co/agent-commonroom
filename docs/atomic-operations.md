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

## Workflow outcome persistence

The background runner commits each successful step's next run revision and the
agent's Idle session in one transaction. Failure outcomes commit the failed run,
agent error and Blocked session together. Human-gate and completion transitions
also check persistence before ending the runner. No next node is dispatched after
an outcome write fails.

`persistWorkflowTransitionLocked` owns this background transition boundary. It
requires the controller mutex, advances the caller's revision only after commit,
and latches the existing integrity execution stop if persistence fails. Runner
state-read or dispatch-intent persistence failures also latch that stop. Controller
logs emit a fixed diagnostic without provider output or raw database errors.

If the UI still shows Working or Running after a storage outage, that is the last
committed state, not evidence that a job continues. Repair storage, inspect history
and worker state, and use the existing owner-only incident recovery procedure.
Recovery does not automatically retry the failed action. Stop and cancellation
cannot undo an external action already performed before its result failed to save.

This is mandatory security/correctness behavior with no independent off switch.
The existing `workflowExecution` flag governs whether workflows may execute;
neither state bypasses audit persistence or the integrity stop. Single-controller
coordination remains the supported model. Explicit request/response workflow
mutations and external provider effects are not distributed transactions.

Regression tests inject begin, every outcome insert and commit failures using the
in-memory database mock. A runner-level test verifies that a failed result batch
leaves the prior run/session intact and prevents the next agent from launching.
No database connection or live provider is used.

## Task dispatch

[Task dispatch](task-dispatch.md) now commits task state, dispatch message, agent
activity/session and retry receipt before launch, and task/agent outcomes together.
Pending intents require verified worker quiescence for explicit owner reconciliation.
