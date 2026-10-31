# Paired worker lifecycle

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


## Status and purpose

The internal Codex path joins a fresh engine/tool pair, credential-free workspace
bootstrap, private audit transport and callback execution. The supervisor journals
both roles and reconciles them by immutable ID after every turn. Synthetic tests
exercise this combined path; public production dispatch remains gated and live
provider acceptance is outstanding.

This is mandatory free/core safety infrastructure. It has no isolation-off switch,
host fallback or paid exemption. Commands and trusted artifacts remain unfinished.

No operator action or product flag enables paired development through this
component. Existing chat execution and legacy worker journals retain their
version-2 contract. Do not delete old journals, rewrite history or use an older
worker to recover new paired work.

## Durable identity and admission

New paired starts use lifecycle **version 4**, execution
`separated-pair-workspace-v1`, and nested pair metadata **version 1**. Historical
version-3 `separated-pair-v1` records remain recovery-only. Version 4 additionally
journals fresh workspace identity and requires its cleanup proof. A version-2 record must not contain pair
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
This accounts for logical jobs. The internal startup module also requires
reported inspection limits for both roles: each at most 512 MiB, half a CPU and
64 processes, with swap equal to memory. Missing or unbounded limits and
inspection identities different from the journal are rejected before either
start. Live daemon enforcement still requires runtime verification.

The internal `Start` interface now owns the following sequence behind a trusted
launch adapter:

1. Persist the version-4 workspace intent, allocate a fresh volume and durably
   bind its nonce/daemon creation time before container creation.
2. Persist `BeginCreate` for a role before creating its fixed-name container.
3. Use Docker **create**, not create-and-start `run`, with reviewed images,
   mounts and aggregate resource policy.
4. Inspect identity/policy and durably `BindCreated` before starting that container.
5. Bind both IDs and verify their aggregate resource inspection before starting
   either role or claiming running state. Never reuse an attempted role or
   substitute an ID. Engine and tool must have different IDs.
6. Start the idle tool, bootstrap its approved snapshot and validate the
   job/repository/commit/branch receipt before starting the engine.
7. Stop on cancellation, expiry or journal failure. Closing the supervisor releases
   leases, but does not itself prove containers stopped or settle the journal.

Both identities are journalled before either container starts. Adapter errors,
cancellation or a durable cancellation latch stop the sequence without retries;
uncertain creation remains available to fenced recovery. Only successful starts,
a matching bootstrap receipt and final journal persistence publish `running`. Closing the supervisor cancels
its original lifetime, then waits for any in-flight creation/start adapter call
to return before releasing leases. Concurrent close callers all wait for this
ownership barrier. A cancellation signal alone is never shutdown proof.

The adapter receives fixed job, role and repository identity, never model launch
arguments. It must create without starting, inspect full image/mount/credential
and aggregate resource policy, and start only the supplied immutable ID with
cancellation. The internal Codex operation supplies that adapter; public production dispatch
is not wired. A source fork **CAN** skip journalling or start before identity publication, **BUT**
then crash recovery and the mandatory credential boundary are no longer proven.

## Cancellation and crash reconciliation

`cancel-pair-v1` retains the historical container-only cancellation contract.
New version-4 jobs require `cancel-pair-workspace-v1` and workspace removal proof.
The historical request is
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
and repository digest as immutable versioned job metadata before dispatch.
Registration policy changes or missing provenance fail closed. This path creates
no process and does not bypass development admission.

Owner recovery preserves the recorded version-3 or version-4 cancellation protocol.
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

Live rootless pair creation/shutdown and resource enforcement remain unverified.
The Codex engine, controlled configuration, workspace bootstrap, private audit
transport and callback driver are implemented internally. Public dispatch,
live refreshed-token custody acceptance, other provider drivers, commands/artifacts and live
acceptance remain P0 work. See [release readiness](release-readiness.md)
and [provider/tool separation](provider-tool-separation.md).

The startup tests cross this same interface and verify durable ordering, failure
at creation/inspection/start, no uncertain replay, context cancellation and the
durable cancellation latch. Pair cleanup shares immutable-ID shutdown mechanics
with the tool executor while retaining pair-specific name, label and lease proof.
This keeps lifecycle ordering in one module instead of duplicating Docker
sequencing in callers.

The [constrained tool runtime](tool-container-runtime.md) now supplies the fixed
tool-side Docker create/inspect/start implementation for later adapter assembly.
It supplies workspace bootstrap; the separate Codex engine runtime supplies the
engine role. Public production dispatch remains gated. Both role IDs must still be durably bound and aggregate
policy verified before startup.

[Workspace bootstrap](workspace-bootstrap.md) explains streamed, credential-free
initialization. [Worker pair assembly](worker-pair-assembly.md) explains how the
internal Codex operation connects both roles.

[Fresh workspace volumes and cleanup](workspace-volumes.md) specifies version-4 creation, creator-lease ownership and remaining daemon limitations.

## Free container capacity

Accounts and workers each enforce a four-container reservation ceiling. A pair
reserves two slots before creation; unresolved work retains capacity. The
[container quota contract](container-quotas.md) explains durable admission,
proof-based release and account scope. The separate hard 5 GB storage limit is
not yet implemented.

See [worker pair assembly](worker-pair-assembly.md) for exact session/runtime
binding, snapshot ownership, startup ordering and remaining engine integration.
