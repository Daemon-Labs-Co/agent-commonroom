# Managed storage accounting

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


CommonRoom's free-account storage budget is **5,000,000,000 bytes**, shared by
retained development workspaces, staged source bundles and stored artifacts.
Container slots and storage capacity are separate: stopping containers must not
make retained source disappear from the storage budget.

The controller now implements durable reservation accounting. Physical quota
provisioning, a complete live inventory adapter, launch/write integration and
production allocation/disposal integration remain unfinished. This is an internal foundation, not an
available storage-management screen or a completed hard disk quota. Production
repository development remains gated.

## Owner and user behavior

There is no new setting, public endpoint or model tool. The budget has no paid,
owner or feature-flag override; paid account limits have not been established.
Existing data is preserved even if an initial inventory exceeds the budget. Such
an account cannot reserve additional capacity. Existing accounts without a trusted
complete inventory have **unknown usage**, rather than zero usage, and reservation
requests are refused.

A reservation is a maximum capacity charge recorded before a future allocation.
For example, 3 GB of retained workspace capacity, 1 GB of staged bundles and 1 GB
of artifacts fill the account. Another byte is refused. Once capacity is reserved,
restarts, expired jobs and container shutdown cannot refund it. Unused reserved
capacity also remains charged unless its exact bound workspace is explicitly
approved for disposal and its physical absence is verified. Reservations cannot
be shrunk or refunded by stopping a job.

Owner-approved disposal is controlled by the default-off `workspaceDisposal`
flag (owner-only, immediate, no restart). Turning it off blocks new approvals but
does not cancel an already committed claim or prevent its verified accounting
outcome from being recorded. Ordinary shutdown never releases capacity.

This is mandatory resource-accounting infrastructure, so there is no optional
on/off product flag. Optional diagnostics/background-work flags cannot disable it.
Disabling development admission prevents new development; it does not erase
storage records, retained data or history.

## Trusted initialization

`initializeManagedStorage` requires an authenticated enabled human owner and an
explicit matching company. Its trusted inventory adapter must stop competing
writers and account for **all** retained volumes, staged bundles and artifacts
across every approved account worker. A partial directory scan, a model's report,
a stale usage poll or a caller-supplied claim of completeness is insufficient.
The implementation provides no default adapter and no empty-account fallback.

The adapter returns immutable nonsecret object identities, reserved/occupied byte
counts and a SHA-256 inventory digest. The controller validates this evidence and
atomically records `storage.inventory.recorded`. An explicit empty list is required
for a verified empty account; omitted/null inventory is rejected. The digest binds
the baseline evidence; it does not certify a kernel quota or ongoing filesystem
usage. Initialization is single-use. There is no reset, baseline replacement or
baseline reset that could erase an earlier charge. Verified disposal records a
separate release while retaining the original reservation.

Inventory adapters are trusted infrastructure, never model/plugin dependencies.
Changing source to supply false inventory bypasses this contract. No production
inventory adapter is shipped yet, so no operator should manually synthesize these
events or edit the database to enable development.

## Reservation and audit contract

`reserveManagedStorage` validates the current enabled human, room access and the
exact company/job/worker/repository binding. A non-owner initiator must match the
job's recorded human initiator. Expired jobs, stopped execution, uncertain/finished
jobs, duplicate reservation IDs, unknown categories, missing inventory and invalid
byte counts are refused. Only workspace, bundle and artifact categories exist.

Reservations and inventory use the controller's append-only account journal.
Each append carries the loaded journal revision. The transactional audit-chain
lock checks that revision before commit; separate controllers cannot both reserve
the final byte from stale snapshots. A conflict is refused rather than silently
retried. Failed or ambiguous persistence grants no allocation authority and does
not authorize an automatic refund. Reload and reconcile the journal first.

All arithmetic is checked before addition. Negative/zero allocations and values
that exceed the ceiling are refused. Malformed, foreign, duplicated or overflowing
replayed evidence blocks state loading. Existing over-budget inventory can be
recorded without deleting data, but cannot grow through reservations.

Audit records contain the authenticated actor, immutable nonsecret binding,
category and byte count. Workspace reservations may additionally prebind the
complete nonsecret volume identity: `workspaceName`, `workspaceNonce` and
`workspaceCreatedAt`. These fields are all-or-none, workspace-only, validated
against the job and unique across the account, including released reservations.
The trusted allocation/inventory adapter must know the exact worker-generated
identity; callers must not fabricate or retrofit it to reclaim old charges. They contain no host paths, credentials, file contents,
command output or prompts. Raw reservation/inventory state is excluded from the
public state projection. Public `/api/events` requests cannot create, replace or
release these internal records, even for the owner.

## Physical enforcement still required

A durable reservation prevents accounting overbooking. It cannot stop a compiler
or detached process from writing past its reserved capacity. Before enabling
production allocation, the worker needs a reviewed physical quota backend that
bounds every writable workspace, staging area and artifact store, including data
retained after job exit. Multiple workers must share the authoritative account
ledger. Missing/ambiguous enforcement must block allocation before effects.

A filesystem usage check or Docker volume label is not a hard quota. Rootless
Docker's plain local named volumes currently have no verified per-volume byte
ceiling. [Disposal identity binding](workspace-disposal-binding.md) now rejects stale or
boolean-only internal cleanup requests. An [owner review workflow](workspace-disposal.md)
now records explicit approval and one-shot execution. Its proposal captures the
unique matching reservation ID, if one exists. After exact worker absence proof,
`workspace.disposal.succeeded`, container-slot release and `storage.released` commit
in one transaction with the expected journal revision. Replay validates the
successful disposal, full immutable scope and confirmed shutdown before excluding
that charge from the total. Duplicate releases and reuse of a released reservation
ID or physical volume identity are refused. The reservation and release remain
in history. Unbound legacy workspaces, bundles and artifacts receive no refund.
Failures or ambiguous audit outcomes keep the full charge; the action cannot be
repeated to obtain a credit. Production allocation binding and live verification
remain P0 work.

## Verification

Fast automated tests mock database access and exercise real controller methods,
journal appends/replay, restart, concurrent controllers, stale transaction
revisions, retained categories, audit failure, current permissions, stopped/expired
work, public event injection and malformed inventory. No real database, Docker
workload or provider runs in these tests.

Nine P0 accounting scenarios now have production-backed hermetic drivers: exact
boundary, one byte over, concurrent final-byte reservations, retained categories,
unavailable metering, overflow decoding, independent-account accounting, negative
reservation and no owner/paid/diagnostic bypass. Other storage acceptance
cases and physical/live verification remain pending. See [P0 acceptance](p0-acceptance-tests.md),
[container quotas](container-quotas.md), and [release readiness](release-readiness.md).

Unavailable-metering acceptance checks the actual initializer with an inventory
source returning an error alongside a structurally valid empty result. That
result must never become authoritative zero usage. The controller must preserve
the journal and refuse even one-byte reservations for workspace, bundle and
artifact categories with unknown usage, including after restart. Recovery requires
an explicit successful complete inventory; its retained charges still count, and
the capacity bound remains enforced. These are mocked controller accounting
checks, not physical allocation or engine-start evidence. They introduce no new
product capability or verification toggle. Mandatory accounting and existing
permission checks remain in force; production inventory/launch integration and
hard worker quotas are still pending.

Internal overflow acceptance exercises the journal projection decoder with
`18446744073709551616` bytes. It must reject that evidence without mutating the
ledger; malformed signed-range, fractional, exponent and string inputs are also
refused. A valid integer updates the projection, and invalid evidence prevents
journal replay. This is not a new public numeric input API: reservation methods
use int64. Account acceptance uses two independent synthetic installation journals,
observes B's durable reservation while A is full, and rejects foreign company
identity. Shared-database multi-tenancy and physical worker allocation remain
outside this coverage.

Retained workspace reconciliation has a combined worker/controller regression:
the cleanup fake begins with both journal-bound role containers present and must
observe their exact-ID removals. Both roles become absent, source and its exact full charge remain, and replay
keeps shutdown unconfirmed and the pair's container reservations occupied. The
catalog scenario requiring container capacity release with retained source stays
pending. Distinguishing process absence from complete version-4 job settlement
is still an architectural gap; tests do not bypass that contract or add disposal
authority to cancellation.

A [read-only physical project-quota observer](physical-project-quota-observer.md) now checks pre-provisioned XFS enforcement on supported Linux x86-64 hosts. It does not provision quotas or replace complete inventory, account admission, rootless containment or live acceptance.
