# Durable worker jobs, shutdown and diagnostics

**Tier: Free.** Execution fencing, durable intent, shutdown verification and crash
reconciliation are required correctness/security boundaries. They have no off
switch and no paid entitlement. The optional detailed diagnostics UI is also Free.

## What is implemented

Normal managed chat/workflow turns record a `worker.job.updated` intent in the
append-only database before SSH dispatch. The record binds job, room, participant,
connection, approved worker and provider. It contains only fixed lifecycle codes
and timestamps: no prompts, commands, provider output, credentials or reasoning.
Allocation/update audit failure prevents launch or stops execution. Journal updates
are independent of the background-work and diagnostics presentation flags.

The worker uses protocol **v2**, durable private outcome files and kernel file
leases. A membership lease covers a room member across connection/provider
changes; an exclusive job lease covers the supervisor's lifetime. A shared
admission lease serializes worker admission. Duplicate job IDs are refused,
unsettled member work prevents replacement, and the worker enforces eight
unsettled jobs overall and four per room independently of controller counts.
Untracked managed containers or unresolved legacy records block new dispatch.

Cancellation first persists a latch, waits for the supervisor lease to be
released, removes only the exact managed container and confirms its absence.
A missing container alone is not proof of shutdown. If cancellation arrives
before a job record exists, a durable tombstone prevents that job ID from being
launched later. Expired requests and requests extending the three-minute budget
are rejected; the original controller deadline travels to the worker. UTC clocks
must be synchronized: excessive skew is refused rather than extending budgets.

After a supervisor crash, its kernel lease is released but its record remains.
The owner can explicitly reconcile that job. Reconciliation proves shutdown,
records it and preserves partial files and provider-session bindings. It never
replays a task, restarts an agent or declares work accepted. Shutdown uncertainty
latches an application stop. A later uncertain callback cannot overwrite an
already recorded verified shutdown.

Startup checks both worker quiescence and unconfirmed controller intents. A
worker with no containers does not make an unconfirmed controller job safe.
Emergency resume refuses unconfirmed jobs. Resolve/reconcile them, review partial
work, explicitly resume the application if stopped, and recover the agent session
if needed. A reconciled stale Working membership becomes Stopped, not Idle.

## Owner workflow

1. Open an agent's three-dot menu → **Job diagnostics**, or **Worker jobs** in a
   room toolbar for application-wide tracked jobs, including archived-room jobs.
2. Inspect the job ID, controller phase, last recorded worker transition,
   supervisor ownership, model deadline and snapshot check time.
3. **Cancel and verify shutdown** explicitly cancels the selected job. It cannot
   cancel a newer controller turn just because that turn uses the same member.
4. **Reconcile abandoned job** is available only when current worker inspection
   does not show a held supervisor lease. Unreachable, malformed, incorrectly
   bound or incomplete worker state is never assumed stopped.
5. Supply a reason. The request is audited before a worker action is sent. Inspect
   partial work; use application Resume and session recovery explicitly afterward.

The snapshot is refreshed on demand, not a complete process trace. Lease ownership
proves an active supervisor holds its coordination lock; it does not prove useful
model progress. Worker transition timestamps are not continuous heartbeats.
Deadlines never serve as shutdown proof. Background work retains its separate
bounded tool/lifecycle timeline. Debug output consists of identifiers, fixed
labels and timestamps, and controller logs record safe job-ID lifecycle lines.
Do not publish operational identifiers without reviewing them for sensitivity.

The view distinguishes unverified/missing worker observations from confirmed
shutdown. Historical/untracked records are reported separately. Nonterminal legacy
or untracked work requires administrator review because this controller cannot
prove its ownership. Normal completed legacy records do not require task replay.

## Flag and permissions

`jobDiagnostics` defaults **on** under Settings → Features → Security. Only the
owner may change it. Off hides the menu/toolbar view, gates `job-diagnostics.js`,
and rejects the diagnostics GET API. Existing records remain. Reload to load an
enabled optional script; stored flag changes do not require controller restart.

All diagnostics and recovery APIs require an authenticated owner. Recovery POST
also requires CSRF and a nonempty reason (maximum 500 characters). The flag does
**not** disable mandatory execution fencing, journal recording, cancellation,
startup checks, integrity controls or the recovery API. Other humans and agents
cannot inspect these owner diagnostics or perform reconciliation. Application-wide
scope exposes operational bindings only, never private-room message content.

## Deployment and compatibility

Install the matching controller and Linux/rootless-Docker worker together during
an idle maintenance window. The controller refuses old worker lifecycle versions;
there is no `run-v1` fallback and controller-local model launching is unsupported.
A local Linux worker must be reached through the approved SSH worker transport.
Containers, pinned images, credentials, session volumes and approved development
snapshots retain the existing mandatory isolation contract.

The wrapper still invokes `commonroom-worker-bin --agent-worker`. Replace that
binary, not provider credentials or session volumes. See [worker installation](connection-sign-in.md).
Before replacing it, verify idle using the installed worker. After replacing both
binaries, inspect the new protocols through the approved alias:

```sh
ssh commonroom-worker '~/.local/bin/commonroom-worker idle-v1'
ssh commonroom-worker '~/.local/bin/commonroom-worker jobs-v2'
```

`idle-v1` retains envelope version 1 and now includes `lifecycleVersion: 2`.
`jobs-v2` reports bounded jobs, held/free/legacy lease state, container/unresolved
counts and whether the snapshot is truncated. `run-v2` uses a strict base64 JSON
request with version, issue time, deadline, job identity, membership key, provider,
exact allowed arguments and optional approved repository digest. `cancel-v2`
uses a strict version/job/provider/membership binding and returns a verified
version-2 acknowledgment. `status-v1` remains a read-only per-job transport;
modern job records inside it carry version 2.

Job files are in the worker's private `~/.local/state/commonroom-worker/jobs`.
Use a reliable local filesystem supporting file locks, atomic rename and fsync.
The worker enforces bounded scans: at most 4,096 records and 16,384 directory
entries. It fails closed rather than silently dropping records. No automatic
job-record purge or tombstone maintenance is implemented; plan operator-managed
maintenance before reaching these limits. Never delete an active lease inode or
remove unresolved records to force a launch. Existing immutable controller audit
history must not be edited as part of operational maintenance.

## Technical contracts

- `managedJobContext` allocates one immutable identity per controller turn;
  `Register` audits intent before command construction; `Update` records submission
  and verified/unverified shutdown. `recordManagedJobLocked` preserves proof
  monotonically and stops execution on audit/storage uncertainty.
- `room.ManagedJob` stores version, IDs/bindings, provider, lifecycle/diagnostic
  codes, allocation/update/deadline timestamps and `ShutdownConfirmed`.
  `State.ManagedJobs` is an internal replay map, not ordinary room JSON.
- `workerJobRecord` stores corresponding worker timestamps, scope digest and
  shutdown outcome. Only supported bounded fields/codes enter diagnostics.
- `workerTryLease`, `workerWaitLease` and `workerJobLeases` manage persistent
  private coordination inodes. `admitWorkerJob` checks durable records and live
  containers under the shared admission lock before creating an exclusive record.
- `fenceWorkerJob` validates binding, persists cancellation, fences late launches
  and confirms supervisor/container shutdown. `stopWorkerContainer` refuses
  removal of an unmanaged container with a colliding name.
- `workerSnapshot`/`remoteJobsSnapshot` bound and validate current observations.
  `remoteFenceJob` accepts only matching version/job/scope/stopped acknowledgments.
- Controller snapshot and fence functions have mock seams for deterministic tests;
  production uses only the approved SSH alias with strict host-key verification.

An operator/source fork **CAN** bypass these contracts, change coordination files
or alter the writer. Doing so is a bad idea: it reintroduces delayed launches,
overlapping workspace writers, false shutdown reports and missing evidence.
Trusted host/worker administrators remain outside this local fencing boundary;
independent administrator-resistant audit anchoring is a separate Paid capability.

## Verification and remaining gaps

Fast tests mock every database and transport interaction and use temporary files.
A synthetic test supervisor is killed to verify kernel lease release and durable
crash recovery. Tests cover pre-launch cancellation, held supervisors despite
absent containers, duplicate identities, membership/room/global admission limits,
unknown containers, binding mismatches, deadline replay, unsafe files, audit
failures, both diagnostic flag states, owner/CSRF permissions and UI escaping/stale
responses. No provider agent or database is used in automated tests.

An explicitly opted-in maintenance smoke test uses a credential-free,
network-disabled Node timer in the existing pinned image to verify actual rootless
container removal. It creates no database data and stores synthetic job fixtures
only in a temporary directory; it is skipped in the fast suite. It must run on an
idle worker and does not verify provider interoperability. The Debian smoke run
on 13 October 2026 passed in 2.45 seconds; it used a synthetic container and
temporary records, with no provider credentials mounted and no database access.

This is a managed-job lifecycle foundation, not complete distributed application
support. HTTP operation idempotency, atomic multi-event business updates,
distributed workflow cursor ownership and transactional conflict checks remain
Free follow-up work. Controllers must not run concurrently against shared business
state merely because worker execution is fenced. Provider connection verification
also uses worker leases/outcomes, but does not yet have controller journal intent;
an abandoned verification job needs operator reconciliation. Unresolved legacy
jobs are not automatically adopted or erased. Provider requests/file changes
already performed cannot be undone by killing a container.
