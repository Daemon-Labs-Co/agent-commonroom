# Sending tasks to agents

Task dispatch is a Free owner-only capability. In a room's Tasks tab, select
**Send to assignee** or **Request review** after assigning an executable agent.
The server reads the current task, criteria and same-room evidence. The browser
cannot substitute the recipient or supply a different prompt. Accepted tasks
cannot be dispatched. Only the owner can accept work; an agent reply never does.

## What happens

The controller checks your login, request token, room access, task revision,
feature flag, integrity, emergency stops, agent availability, capacity, isolated
worker configuration and context disclosure policy. It commits the task's Working
or Reviewing state, human dispatch message, agent activity, Working session and
retry receipt in one transaction before starting the turn. Failure rolls back all
of those writes and launches nothing.

The task shows a separate dispatch state:

- **pending:** committed intent is waiting for a worker result. Task editing and
  additional dispatch are blocked while the intent is pending.
- **completed:** the agent replied. Inspect the conversation and evidence; this
  does not prove the task is complete or accepted.
- **failed:** the turn stopped or failed. Inspect agent diagnostics before retrying.

Agent result/error, participant status and task dispatch outcome are committed
together. If result storage fails, the mandatory integrity stop is latched and the
last committed pending state remains visible. Repair storage and follow owner
incident recovery before proceeding. No automatic provider replay is attempted.

## Interrupted work

After resolving any integrity incident or emergency stop through the existing
owner recovery controls, **Check pending dispatch** verifies there are no active controller turns and the
approved worker is idle. Unknown/busy worker state leaves the task pending. Only
confirmed quiescence marks the dispatch failed, allowing an explicit retry.
This records uncertainty rather than claiming the previous action never occurred.
Reconciliation does not launch an agent or undo edits. Inspect actual files and
results before retrying. It conservatively requires whole-worker quiescence.

## Feature flag

`taskDispatch` defaults **on**, requires owner permission to change and appears in
Settings → Features → Workflows. Off hides send/review/reconciliation controls and
rejects all new dispatch/reconciliation commands on the server. Existing turns
finish normally; tasks, receipts and history remain. Re-enable it to reconcile a
pending intent. No restart is needed after refreshing settings. Emergency stop,
audit checks, context restrictions and container isolation remain enforced in
both states. There is no fallback to the former pair of browser mutations.

## API and technical contract

Authenticated, owner-only, CSRF-protected `POST /api/task-dispatch?room=ROOM`:

```json
{"taskId":"TASK","revision":1,"mode":"work"}
```

Mode is `work`, `review` or `reconcile`. Work/review returns 202 with `taskId`, new
`revision`, `dispatchId` and `queued:1`. Reconcile returns 200 with `reconciled:true`.
Use `Idempotency-Key` to safely retry the same request: a committed receipt returns
its original response without dispatching again, including after completion.
Reusing a key with different input returns 409. An original 202 replay acknowledges
previous admission, not current health; refresh state for the current outcome.

403 indicates permissions/CSRF/disabled capability; 404 missing room/task; 409
stale/pending/accepted task, stopped execution or busy member; 429 capacity;
413 context policy/bounds; 503 unavailable worker or state; 500 failed atomic commit.
Unknown request fields are rejected. Recipient, text and task state are server-owned.

`taskDispatch` holds the controller lock through validation and intent persistence.
The managed turn uses the mandatory isolated worker and existing durable job
lifecycle. `taskOutcome` binds completion to the dispatch ID, increments revision
and cannot accept work. Public task revisions preserve server-owned dispatch
fields and reject edits while pending. Existing records need no SQL migration;
new optional fields are replayed from task events.

Support remains single-controller. Worker readiness at admission validates approved
configuration; runtime/authentication failure can still occur afterward and is
recorded as failed. There is no exactly-once claim for external provider effects.
A manual ordinary room message is not automatically a task assignment.

Tests mock all database, readiness, idle and provider interactions. They cover
transaction insert/commit failure, replay, stale revision, permissions, feature
states, disclosure/capacity/stops, result failure and reconciliation. Live provider
and deployment acceptance are not covered by these tests.
