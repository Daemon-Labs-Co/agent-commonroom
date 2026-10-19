# Worker idle shutdown guard

The internal paired-worker supervisor can stop an unused started pair after an
administrator-configured idle interval. The default is five minutes. It keeps
workspace source and never restarts an old container or releases storage credit.
This is an internal lifecycle component; public development dispatch and the full
idle-agent product feature remain gated.

## Configure the approved worker

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
can configure this policy. There is no model tool or public configuration API.
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
announcement or complete enterprise audit chain. Controller notification, owner
application settings, production dispatch integration and live acceptance remain
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
