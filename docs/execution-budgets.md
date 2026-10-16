# Daily managed-turn budget

**Tier: Free security safeguard.** Agent CommonRoom limits managed provider launches
per UTC calendar day using its immutable controller job journal. Restarting the
controller, signing out or disabling optional diagnostics does not reset usage.
This is a launch-volume budget, not a billing estimate or dollar spend guarantee.

## Configuration

The default is **200 allocated managed turns per application per UTC day**.
Set `COMMONROOM_DAILY_TURN_LIMIT` before starting the controller to choose an
integer from 1 to 100000. Restart the controller to apply a new operator limit.
Zero, negative values, non-integers and out-of-range values prevent startup;
there is no unlimited or off setting. Example:

```sh
COMMONROOM_DAILY_TURN_LIMIT=500 ./bin/commonroom
```

This limit is application-wide, not per room/person/provider. An operator may
raise it deliberately; changing the configured limit does not erase recorded
usage. Provider prices, tokens, retries and tool behavior vary: this policy does
not establish a monetary ceiling. It complements mandatory eight-global/four-room
concurrency bounds, bounded prompts, worker deadlines and resource limits.

## Usage and recovery

Owner **Worker jobs → Job diagnostics** shows today's allocated count, remaining
units and reset time. The time is displayed in the browser's local timezone;
the window resets at 00:00 UTC. Exhaustion rejects new chat/task batches before
launch where possible. Workflows or a race that reaches the final allocation gate
report an explicit budget blocker. Wait for the reset, or intentionally adjust
the deployment limit. No task is automatically replayed when the budget resets.

Every unique committed job allocation consumes one unit, including connection
verification and turns that later fail, time out, stop or never reach the provider.
Lifecycle updates, shutdown reconciliation and retry receipts do not consume new
units. A rolled-back allocation consumes none. Prior-day unresolved work still
requires shutdown reconciliation; a calendar reset is not permission to replace it.

Already admitted jobs can finish after exhaustion. Emergency stop, cancellation,
owner recovery, sign-out and security inspection do not require a daily turn unit.

## Flag and permission behavior

Enforcement is mandatory with no product off switch. This is an execution security
boundary, not an optional premium feature. The existing owner-controlled
`jobDiagnostics` flag defaults on and controls only the usage/details view. Off
hides/rejects that view; the same durable allocation enforcement remains active.
`connectionTesting`, `taskDispatch` and `workflowExecution` can disable their
respective capabilities but cannot bypass the budget or isolation safeguards.

## Technical contract

`dailyTurnBudget` counts unique projected `ManagedJobs` by their original
`allocatedAt`, using a half-open UTC-day window. `agentAdmission` performs an
upfront check; `managedJobContext.Register` rechecks the full journal under the
controller mutex before appending allocation intent. That append also checks the
loaded event count transactionally. Failed persistence prevents launch and trips
the existing integrity stop. Durable managed-job updates retain allocation time,
so completion does not refund usage. No SQL migration or standalone mutable
counter is required.

`GET /api/job-diagnostics` includes application-wide `turnBudget`:
`{limit,used,remaining,resetsAt}`. It is owner-only and uses the existing optional
view flag; values remain available even if the worker snapshot cannot be verified.

Supported scope is one controller and one intact audit journal. Concurrent
controllers are unsupported; the state-count check detects conflicting writes
but is not a distributed quota service. Host administrators can change policy or
clock settings. A source fork **CAN** bypass the gate, **BUT** then restart-based
cost abuse and uncontrolled repeated launches become possible. Independent audit
rollback protection is still separate work. Existing pre-journal/legacy turns,
provider-internal calls and human sign-in are not counted. Storage/input/output
quotas and provider monetary budgets remain outstanding.

## Verification

Fast, parallel-safe tests use only mocked database events and synthetic timestamps.
They cover UTC rollover, completed/cancelled allocations, unique-job projection,
restart persistence, disabled diagnostic flags, batch limits, concurrent attempts
at the final slot, invalid configuration and rollback. UI tests check readable,
escaped usage. No database, live SSH worker or provider is contacted. Deployment
and live provider acceptance are not established by these checks.

## Free container capacity

Worker admission now reserves at most four physical managed container slots.
An engine/tool pair reserves two slots before either container exists, including
unresolved and partially created jobs. Thus at most two development pairs fit
on a worker. Existing eight-job/four-room bounds are additional ceilings, not
permission to exceed four containers. Changing rooms or omitting containers from
inventory does not release durable reservations. Unknown containers block launch.
There is no toggle or owner override for this safety ceiling.

Controller registration also enforces four physical slots across workers and
rooms using durable reservations and an expected-revision allocation append.
See [managed container quotas](container-quotas.md) for account scope and recovery.
The hard 5 GB storage ceiling remains pending. Paid limits have not been
established and cannot bypass the current guard.
