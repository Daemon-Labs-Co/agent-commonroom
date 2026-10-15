# Paired worker lifecycle

## Status and purpose

The worker now journals a logical engine/tool pair and can reconcile or cancel
its containers using immutable IDs. This is mandatory free/core safety
infrastructure. It does **not** launch a provider pair, initialize a workspace,
install an engine, enable commands, or authenticate the future audit transport.
Development admission remains blocked by the existing credential boundary.

No operator action or product flag enables paired development through this
component. Existing chat execution and legacy worker journals retain their
version-2 contract. Do not delete old journals, rewrite history or use an older
worker to recover new paired work.

## Durable identity and admission

Paired records use lifecycle **version 3**, execution `separated-pair-v1`, and
nested pair metadata **version 1**. A version-2 record must not contain pair
metadata. This ensures older binaries, which reject unknown lifecycle versions,
cannot silently ignore the extra fields and declare one-container shutdown.

Each record binds room, participant, connection, provider, original deadline and
approved repository digest. It stores separate engine/tool creation intents,
immutable container IDs and absence observations. No provider secret, model
prompt, command content or host path is stored in pair metadata. Company/human
provenance remains the controller audit contract; the worker journal does not
invent human authority.

`admitWorkerPair` is an internal seam, not a public launch endpoint. It shares
the existing admission, job and membership leases. One pair consumes one logical
job slot under the eight-global/four-per-room limits, including unresolved work.
Both expected container names are recognized together; orphan, legacy-version,
unknown or already-settled container associations block admission. Duplicate job
IDs, a held membership lease and persisted cancellation all refuse admission.
This accounts for logical jobs; actual engine/tool resource partitioning and
inspection remain requirements for the future launcher.

The owning supervisor must:

1. Persist `BeginCreate` for a role before creating its fixed-name container.
2. Use Docker **create**, not create-and-start `run`, with reviewed images,
   mounts and aggregate resource policy.
3. Inspect identity/policy and durably `BindCreated` before starting that container.
4. Bind both IDs before claiming running state. Never reuse an attempted role or
   substitute an ID. Engine and tool must have different IDs.
5. Stop on cancellation, expiry or journal failure. Closing the supervisor releases
   leases, but does not itself prove containers stopped or settle the journal.

These methods create no containers and grant no host-execution authority. A
source fork **CAN** skip journalling or start before identity publication, **BUT**
then crash recovery and the mandatory credential boundary are no longer proven.

## Cancellation and crash reconciliation

`cancel-pair-v1` is the explicit worker cancellation command. Its request is
version 3, bound to job/provider/membership and repository digest. Legacy
`cancel-v2` cannot settle a pair. A version-2 or changed-scope/repository reply
cannot acknowledge pair shutdown. The control-plane transport remains the
approved authenticated SSH command path; no new network listener exists.

Cancellation is durably tombstoned first. Cleanup attempts both roles even when
one removal fails. It verifies fixed name, role, managed label, job and repository,
and removes only by immutable container ID. If creation succeeded before the ID
was journalled, cleanup may discover the fixed-name candidate, verify those
bindings, and use its immutable ID strictly for shutdown; it never resumes it.

Successful exact-ID inventory must prove every observed role absent. A name
replacement, extra container carrying the job label, malformed/ambiguous inventory,
daemon error or incomplete cleanup refuses confirmation. Unrecognized containers
are not removed automatically. A role whose creation never began is checked for
unexpected containers too. An interrupted creation with no observable container
is settled only after the creator lease is released and inventory succeeds.

The cancelling process must acquire the original supervisor lease before it
persists `shutdownConfirmed`. While that lease is held it may attempt cleanup,
but missing containers alone cannot acknowledge shutdown. Partial cleanup remains
`shutdown_unconfirmed`, retaining known IDs/observations for explicit reconciliation.
No action, provider session or uncertain side effect is replayed automatically.

The controller has an explicit internal paired registration path. It requires
an initiating human with current room access, a company binding, enabled matching
repository policy and the bounded original deadline. It persists execution kind
and repository digest as immutable version-3 job metadata before dispatch.
Registration policy changes or missing provenance fail closed. This path creates
no process and does not bypass development admission.

Owner recovery of recorded version-3 jobs sends the paired cancellation protocol.
Snapshot version, execution kind, repository digest, deadline and membership must
match; stale/legacy evidence or held ownership cannot authorize reconciliation.
Diagnostic UI flags do not disable recovery checks. Ordinary chat remains
version 2, and the legacy launcher explicitly refuses paired registrations.
Scoped audit checks also bind the original repository, so a later policy change
cannot silently retarget a job. Automatic lifecycle records retain the initiating
human's attribution. Existing job versions remain readable but cannot be promoted
or downgraded by rewriting lifecycle metadata.

## Tests, flags and limits

Tests use temporary worker journals, real local file leases and a mocked Docker
runner. They cover creation-before-ID publication, immutable/unique IDs, partial
creation, held leases, cancellation-before-admission, one-role failure, daemon
failure, name reuse, foreign labels/repository, extra job containers, logical
admission limits, unknown versions and protocol downgrade refusal. Existing
supervisor-crash tests continue exercising kernel lease release with a synthetic
test subprocess. Controller tests additionally exercise the real owner recovery handler with mocked
worker observations, permissions/diagnostic flags, repository/deadline mismatches,
failed shutdown and immutable event projection. No provider, real Docker workload
or database is used.

There is no pair-safety-off flag, owner exemption, paid bypass or uncontained
fallback. Optional diagnostics flags do not disable journal/fencing checks.

Live rootless pair creation/shutdown has not been verified for these changes.
Actual pair startup, fresh controlled provider configuration, credential-free
workspace bootstrap, aggregate runtime resource verification, authenticated audit
substreams, callback-only provider drivers, command execution and artifacts remain
open P0 integration work. See [the execution progress](p0-execution-progress-2026-10-15.md)
and [provider/tool separation](provider-tool-separation.md).
