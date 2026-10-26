# Recorded workspace probe recovery

The internal `recoverRetainedWorkspaceProbe` operation reconciles an interrupted
mount helper whose exact container ID was durably recorded. It preserves the
retained workspace and lifecycle history. It never launches a helper, returns a
mount report/source bytes, disposes a volume or implements `FrozenLease`.

## Authority and invocation

Trusted worker/controller assembly supplies the exact retained creator record,
recorded probe intent, private worker directory, approved rootless Docker runner
and data root, plus a mandatory `workerProbeRecoveryGuard`. The intent must match
the existing job/challenge/repository/image/volume nonce/time record exactly.
There is no HTTP/model/plugin entry point or permissive production guard.

`AuthorizeRecovery` must enforce current authenticated human cleanup permission
for that exact intent and container ID; an old export approval is insufficient.
The guard also supplies durable observation/audit acknowledgments, account-wide
cleanup accounting and integrity-stop propagation. Audit must deduplicate by
intent/stage: explicitly resumed recovery can acknowledge the same historical
intent or cleanup outcome again. Original authority must stay bound across
contexts. Guard callbacks must not re-enter creator/admission locks. Production
controller transport and authenticated invocation remain outstanding.

This mandatory internal security infrastructure has no optional feature flag or
host/uncontained fallback. Existing task flags never grant cleanup permission.
Operators cannot enable file export by invoking this operation.

## Exact-ID cleanup boundary

Recovery holds the real creator job lock throughout inspection, durable recovery
intent, removal and final checks. Busy creators refuse; recovery never stops or
steals a live creator. The unchanged reconciled role record and cancellation
latch remain required. Both creator role names must be absent.

Only IDs in original `bound` or known-ID `finish` events qualify. A begin or
lost-create finish with no ID remains unresolved and charged; a name or label
match cannot turn it into removal authority. Already settled records refuse.

Before removal, the trusted runner repeatedly checks exact volume metadata and
job/name/volume reference inventory. Only the recorded ID or complete absence is
allowed. A present ID must have the exact probe name, image and ownership labels.
An owned helper may be running or have policy drift: recovery only stops/removes
it, never executes it. This is narrower than permission to observe source bytes.
Foreign ownership/image, extra references, returning roles, changed or missing
volume identity and uncertain daemon replies refuse removal.

`StopContainer` removes only the immutable recorded ID, validates ownership again
if present and observes absence. Recovery then rechecks retained volume/reference
identity. Before confirming recovery, all subsequent observation checks require
zero references; a returning owned helper is not an allowed attachment anymore.
The final ordinary observation also runs before creator ownership is released.
The operation has a bounded ten-second lifetime; cancellation/failure leaves
uncertain recovery state for another explicitly authorized attempt.

## Journal and restart behavior

Existing version-1 histories retain their original events. Valid recovery extends
an unresolved known-ID history with:

1. `recovery`: durable exact-ID cleanup intent and controller acknowledgment
   before any removal.
2. `recovery-finish`: observed absence, durably recorded before cleanup audit.
3. `recovered`: acknowledged local slot-release commit after current authority
   and zero-reference checks.

Histories remain canonical/private and bounded to six events. No supported
transition edits or deletes earlier evidence. Interrupted `recovery` and
`recovery-finish` can resume explicitly with fresh authorization; already recorded
stages are re-acknowledged without duplicate history or relaunch. Removal is
idempotent when the exact ID is already absent. Unknown IDs and corrupt records
are never repaired or adopted automatically.

Only `confirmed` or `recovered` histories release the local probe slot. Pending
recovery continues to count during worker admission and blocks another challenge
for that creator. Account-wide release still requires the trusted controller.
A final authorization/observation failure after `recovered` withholds success but
does not undo acknowledged actual absence or re-charge a removed helper.

Write/sync/audit failures use the existing durable integrity-stop requirement and
live revocation callback. Recovery never clears an integrity-stop marker or
bypasses it; marked incidents require separate administrator investigation. It
never refunds retained workspace/storage charges or deletes confirmed history.
A source fork can replace trusted runners/guards or skip fencing, invalidating
managed security claims. Worker parents, filesystem and daemon administrators
remain trusted; snapshot inspections cannot exclude administrator races forever.

## Verification and remaining gaps

Tests use real temporary private files/creator locks and the actual journal and
removal runtime with synthetic controller/Docker responses. Fixtures first seed
and reconcile both actual recorded creator role IDs. They exercise running,
stopped, already absent and known-unbound helpers; exact removal ordering;
explicit retry after revocation/interruption; unknown/foreign identities;
returning references; cancellation; busy ownership; integrity stops; audit/removal
failure and final revocation. No live database, provider, Docker daemon, SSH,
credentials or private fixtures are used.

Production authority transport/invocation, continuous reference enforcement,
live Linux/helper verification, recovery of lost create replies without a known
ID, integrity-incident remediation, retention/compaction, account-wide physical
admission, frozen source-root authority, physical storage quota and authenticated
file delivery remain outstanding. Required file/bundle dispatch remains blocked.
No P0 release gate or production readiness is certified by these synthetic tests.
