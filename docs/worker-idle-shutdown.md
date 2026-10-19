# Worker idle shutdown guard

The internal paired-worker supervisor can stop an unused started pair after an
administrator-configured idle interval. The default is five minutes. It keeps
workspace source and never restarts an old container or releases storage credit.
This is an internal lifecycle component; public development dispatch and the full
idle-agent product feature remain gated.

## Owner settings

Open **Settings → Security → Idle agents** as the enabled human owner. The
`idleAgentShutdown` feature flag defaults **on**. Use its switch to enable or
disable idle shutdown for new paired-job allocations, and **Set idle interval**
to choose 1–60 whole minutes (default 5). Settings remain editable while off.
Only the owner may save either value. No restart is needed for new allocations;
already allocated/running jobs retain their durable snapshot. A stale interval
form is refused; refresh before saving again.

These controls configure admission policy, not public development availability.
The production dispatch barrier remains in force. Disabling idle shutdown cannot
disable deadlines, normal cleanup, isolation, audit integrity or retention.

## Configure the approved worker fallback

Set these values in the trusted worker process environment, owned by the worker
administrator. Do not accept them from a room, model, plugin, browser request or
SSH client environment forwarding.

```dotenv
COMMONROOM_AGENT_IDLE_SHUTDOWN=true
COMMONROOM_AGENT_IDLE_MINUTES=5
```

| Setting | Default | Accepted values |
| --- | --- | --- |
| `COMMONROOM_AGENT_IDLE_SHUTDOWN` | `true` | Exact `true` or `false`; omitted uses the default. |
| `COMMONROOM_AGENT_IDLE_MINUTES` | `5` | Integer from 1 to 60. |

Only the deployment administrator with permission to change the worker environment
can configure this fallback. A registered controller job snapshot takes precedence
when provided to reviewed Codex assembly; these environment settings apply to
legacy internal calls without that snapshot. There is no model configuration tool.
Restart the worker deployment to apply environment changes reliably. Each newly
admitted `Run` snapshots policy before creating containers; existing jobs keep
their policy. Invalid values refuse paired execution before startup, including
invalid intervals when the guard is disabled. This does not validate or change
controller startup configuration.

Setting the flag to `false` disables only the optional idle timer for new pairs.
Normal turn completion, cancellation, hard deadlines, isolation, audit integrity,
container fencing and source retention remain enforced. Disabling the timer does
not resume a closed pair or erase a recorded intention.

## What counts as idle

The timer starts after durable startup finishes. Startup, active turns and queued
turns block idle shutdown. A completed queued turn resets a full interval. Output
silence is not idleness: a long-running build or model turn remains active.
Arrival and expiry serialize under the supervisor locks. Once expiry wins, new
work is refused; the caller must retain/reconcile it rather than restart this pair.

Current production development admission remains blocked. The internal paired
path is single-turn, stops immediately when its turn returns, and has a hard
three-minute lifetime. That deadline takes precedence over the default five-minute
idle interval. This guard does not extend the lifetime or create a resident pool;
it protects a shorter configured idle gap between startup and work, and future
reviewed assembly that deliberately waits for work. Existing chat/login paths
keep their own one-shot cleanup.

## Retention and uncertainty

Before cancellation or Docker cleanup, the supervisor persists
`idleShutdownRequestedAt` and `idleTimeoutMinutes` with an
`idle_shutdown_requested` diagnostic in its private worker journal. Failure to
persist closes admission without sending an unrecorded Docker cleanup action;
ordinary recovery must investigate the retained resources.

After a committed intention, the existing bounded `Shutdown` path cancels the
supervisor, fences creator ownership, removes both roles by immutable identity
and checks absence. It never sets workspace-disposal authority. Retained source
keeps full job shutdown unresolved even when both containers are absent. Failed
cleanup remains unresolved, with no automatic retry, replacement or capacity
refund. Use [worker recovery](worker-jobs.md) and separately
[owner-reviewed disposal](workspace-disposal.md) when appropriate.

The intention is durable worker-local evidence, not a controller append-only
announcement or complete enterprise audit chain. Owner settings and immutable controller policy snapshots are implemented.
Controller idle-stop notification, production dispatch integration and live acceptance remain
P0 work. The worker journal must not be edited to manufacture shutdown proof.

## Technical contract and verification

`Run` validates deployment policy and queues the turn before taking the work lock.
`Start` arms a configured guard only after it commits running state. The timer and
work admission share supervisor state; expiry must acquire both turn and startup
ownership without blocking active work. `Close` and the original hard-deadline
context terminate the timer. Cleanup uses a fresh bounded context.

Low-level `Start` has no default runner; trusted internal callers must bind the
policy and reviewed cleanup runner with `configureIdle` before startup. Production
assembly must use `Run`/`RunCodex` rather than calling provider work outside these
ownership boundaries. Replacing the runner, omitting configuration, or invoking
work outside the turn lock in a source fork invalidates the idle guarantee and
may invalidate mandatory container isolation. There is no host-execution fallback.

Fast tests use temporary worker journals, virtual expiry instants and mocked
Docker observations. They cover policy defaults/on/off/invalid values, Run
admission, the exact boundary, startup/active/queued work, reset, durable-intention
failure, both-role removal, retained source, cleanup uncertainty and refusal to
revive or automatically retry a closed pair. No real provider, Docker daemon or
database is used. Full idle catalog and live runtime acceptance remain pending.

## Controller and API contract

`POST /api/events` accepts owner-only `agents.idle_policy.updated` with exactly
`{"minutes":5,"revision":0}`. Both fields are required integers. The server
commits revision 1; later saves supply the current revision. Missing CSRF,
unauthorized/disabled actors, duplicate/unknown fields, invalid minutes, stale
revisions and failed audit writes cannot change policy. State exposes
`agentIdleSettings`; revision zero/minutes zero represents the unsaved default.
The existing owner-only `feature.flag.updated` operation changes
`idleAgentShutdown`. Neither operation resumes or cancels an existing job.

Paired registration snapshots `idlePolicyVersion:1`, `idlePolicyRevision`,
`idleShutdownEnabled` and `idleMinutes` into its mandatory `worker.job.updated`
allocation event. Policy changes cannot rewrite those immutable fields. Legacy
allocations with no snapshot remain readable; no history is backfilled.

Reviewed authenticated assembly passes that exact registered allocation as
`workerCodexRunConfig.ManagedJob`. `configureJobIdle` checks lifecycle, job,
provider, membership, repository and deadline against the admitted supervisor
before binding the guard. A foreign snapshot is refused before acquiring
credentials or consuming the bundle. A valid snapshot also remains enforced if
later preparation fails. This trusted internal handoff is not an authenticated
wire protocol; production controller-to-worker admission still must transport and
verify it. Omitting the snapshot in old/internal source paths uses the deployment
fallback and does not establish enforcement of the controller policy.

Additional mocked tests cover owner/non-owner/CSRF checks, strict request shape,
stale forms, audit failure, durable settings/replay, default/on/off snapshots,
immutable allocations, exact worker binding and Codex assembly refusal/cleanup.
Synthetic DOM tests cover settings visibility, revisioned saves and re-enabling
an off policy. One P0 driver verifies the unauthorized configuration boundary;
remaining idle drivers and live browser/worker acceptance are still pending.
