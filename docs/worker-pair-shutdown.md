# Paired turn shutdown and reconciliation

An admitted isolated turn must always end with fenced reconciliation, whether it
finishes normally, fails startup, returns a driver error, loses audit access,
is cancelled or panics. `workerPairSupervisor.Run` now owns that internal lifecycle:
it invokes startup, runs the trusted reviewed-engine turn driver and defers
`Shutdown`. This is mandatory isolation infrastructure without an off switch.
The internal Codex operation uses this lifecycle; public launch and live engine
acceptance remain unfinished.

## Stop behavior

`Shutdown` cancels the original supervisor lifetime, waits for both active model
work and in-flight startup to unwind, then releases the creator leases. Concurrent
shutdown attempts on the same supervisor are serialized. The Run wrapper supplies a fresh cleanup lifetime independent of the model
request; Shutdown caps observation at 15 seconds within its caller context. Thus
a cancelled model or audit connection cannot skip cleanup merely by cancelling
its own execution context. An explicit already-cancelled cleanup context still
refuses observations and cannot become absence proof.

The existing fence path validates any [exact disposal binding](workspace-disposal-binding.md),
persists a cancellation tombstone, acquires the original
job lease, validates durable version/job/provider/repository identity and
attempts both engine and tool removal. Only inspected immutable container IDs
are removed. Successful absence observations and final name/job inventory checks
remain required; unknown/replaced containers or daemon failures retain unresolved
state. One failed role does not skip cleanup of the other role.

`Close` remains cancellation and lease ownership handling, not shutdown proof.
It now waits for an active turn as well as startup. A trusted adapter or turn
driver that ignores cancellation can keep closure waiting; never release its
lease early, fabricate absence or dispatch replacement work. Turn callbacks must
return before invoking Close/Shutdown themselves because those operations wait
for the callback to unwind. They may request cancellation through their driver
and return; they must never launch provider work outside the reviewed engine.

## Retention and results

Ordinary stop and normal completion do not authorize workspace discard.
`Shutdown` deliberately leaves the discard bit unset. A present workspace is
retained, even after both containers have been removed, and full lifecycle
shutdown remains unconfirmed. Its reservation stays unresolved until [owner-reviewed disposal](workspace-disposal.md)
and reconciliation complete. Storage capacity remains charged until its separate
exact bound credit release is durably verified. No restart, workspace
adoption or owner-role destructive bypass is provided.

`Run` preserves both model/startup failure and cleanup failure in its returned
error. Original audit cancellation is observed even if callback cancellation
propagation has not yet run. It cannot report a successful lifecycle when cleanup
is unverified. A driver panic still executes deferred cleanup and then propagates;
it is not swallowed or treated as a successful turn. Source/output sanitization
remains the reviewed driver's responsibility; this helper does not log raw errors.

The optional [idle shutdown guard](worker-idle-shutdown.md) uses this same
non-destructive fence. Disabling its timer cannot disable ordinary shutdown.

## Verification and remaining integration

Tests use temporary journals and mocked Docker observations, never a real daemon,
database or provider. They cover both-role removal, partial failure, retained
workspaces, cancelled cleanup observations, repeat shutdown, normal/error/panic
turn exit, startup refusal, audit cancellation and concurrent stop while a turn
unwinds. Creator leases remain held until work returns. Synthetic successful
absence observations are not evidence of actual workspace deletion or retention
in a deployed system.

`RunCodex` already uses this lifecycle around the internal engine and audited
session. Public worker dispatch must preserve the same ownership and ordering. Live rootless/SSH/provider shutdown,
artifact retention and approved discard remain P0 work. The development admission
barrier stays enforced until those requirements and the other P0 blockers are
verified.
