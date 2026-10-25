# Controller artifact custody journal

The controller now provides the durable metadata side of the internal
[object custody store](object-custody.md). An object intent and its maximum storage
charge commit together, physical quota setup requires its own acknowledgment,
sealing records an immutable receipt, and every successful byte read requires a
new durable release acknowledgment. This closes the previous gap where tests
could supply only an in-memory journal callback.

**This is internal infrastructure, not available artifact delivery.** There is
no new UI, HTTP endpoint, provider tool or operator switch. Required file/bundle
task dispatch continues to refuse. Optional declarations do not establish that
an artifact was produced. No P0 acceptance observation is created by these records.

## Operator and user behavior

There is no installation command or manual event to enable this component. Do
not insert custody events or synthesize quota/source evidence in the database.
Production assembly still needs complete account inventory, an approved isolated
source adapter and a real quota-backed private object volume. These mandatory
checks have no feature-flag override. The controller supplies no default guard,
empty-inventory fallback, host source or automatic uncontained execution path.

The entire reservation remains charged after a failed setup, failed seal,
shutdown, expiration, failed read or unused capacity. For example, storing a
23-byte report under a 128-byte reservation charges 128 bytes. Missing or uncertain
setup acknowledgment does not authorize a write or a later automatic adoption.
There is no object retry, replacement, refund, retention deletion or disposal
operation. A new task attempt is a separate immutable execution and may reserve
new capacity only through all normal admission checks.

Existing task flags continue to apply to source work: turning off Task definitions,
changing participant/connection/policy/input authority, coordination revocation,
delegation revocation, emergency stop or ending the submitted job prevents new
source reservation/sealing. Optional diagnostics flags cannot disable custody
accounting. Stored objects and charges remain retained. Reading a sealed retained
object uses current human room authorization and the trusted current export
policy; it does not require the original container to remain running or a task
editing flag to stay enabled. Emergency/integrity stops still block release.

## Trusted assembly contract

`newControllerArtifactJournal(server, guard)` implements `objectcustody.Journal`.
It refuses a nil or typed-nil guard. This constructor is internal controller code,
never a model/plugin extension point. Four mandatory guard methods remain a
trusted production integration boundary:

- `ValidateSource(ctx, state, intent)` must establish the actual credential-free
  isolated tool-volume identity, frozen repository/revision, declared output/member
  and current source/export authority. A model path, provider response or supplied
  digest is metadata, not evidence. The callback runs again after physical setup
  and before seal admission.
- `EstablishQuota(ctx, reservation)` must enforce the exact maximum allocation on
  the actual private object root after the durable account charge and before any
  source read or file creation. A quota label, counter or directory mode is not
  physical enforcement. Uncertain outcomes retain the charge.
- `ValidateSeal(ctx, state, receipt)` must bind actual stored digest and byte count
  to the frozen source and revalidate quota/root identity. The byte store verifies
  its private inode independently; the guard supplies the missing source proof.
- `AuthorizeRelease(ctx, state, receipt)` must verify current authenticated export
  policy and exact sealed source/root identity, including retained completed work.
  It runs before the store reads the inode and again before recording release.

Callbacks must respect the bounded storage context, never call back into a locked
server mutation, and never treat a synthetic test guard as production evidence.
A source fork **can** supply dishonest/no-op callbacks, replace the root or call
internal seal methods with fabricated receipts. That invalidates the security
contract; the supported product exposes no such bypass. The real source, quota
and release-policy guard is still unimplemented. The journal cannot itself attest
kernel quotas, source volume identity or byte delivery to a remote client.

## State transitions and persistence

Every event carries version 1, the authenticated human actor, the complete intent,
and the applicable reservation/receipt. Actual bytes, credentials, raw errors and
host paths are excluded. The generated object ID is also its storage reservation
ID; logical output names never select controller filesystem paths.

1. `artifact.custody.reserved` atomically adds the exact maximum `artifact` charge
   and object intent in one account audit event. It uses the complete storage
   inventory and mandatory decimal 5,000,000,000-byte account ceiling. Unknown
   inventory, exhausted capacity and malformed/foreign bindings refuse. Only a
   currently submitted workspace job bound to the exact active work execution,
   task definition and declared finite file/member may reserve. Review jobs and
   undeclared/value outputs cannot create artifacts. A human initiator or owner
   must have current room access.
2. The trusted quota callback runs. `artifact.custody.ready` commits only after
   physical setup succeeds and source authority is revalidated. If setup, current
   authority or acknowledgment fails, capacity stays charged and source/file I/O
   is withheld. A restarted controller cannot seal an unready object.
3. `artifact.custody.sealed` stores an exact, valid, immutable receipt only for
   that ready, unreleased reservation. Current source/job/contract authority and
   the mandatory seal guard must pass. The store fsyncs and hashes actual private
   inode bytes before calling this method. There is no second seal or replacement.
4. `AuthorizeRead` requires an exact existing sealed receipt and current human
   access/export policy, but grants no durable release by itself. The store then
   validates actual inode bytes. `artifact.custody.read` rechecks current authority
   and records release acknowledgment before any byte slice is returned. A read
   failure returns no bytes and makes no automatic refund. Acknowledgment means
   bytes were authorized for release, not proof that a client received them.

One output member can be reserved only once per execution, even by changing its
object ID, revision digest or declared maximum. The complete intent, receipt,
reservation linkage and read counter are validated during replay. Recorded human
membership/room access, retained job/execution/definition/output relationships
and pending source phase are checked at each source event position. The default
owner is intrinsic if its derived actor has not yet been projected; an explicit
historical disabled owner still refuses. Unknown kinds,
foreign metadata, missing readiness, forged/mismatched receipts, unknown/agent/disabled actors, missing source records, duplicate seals,
released/missing charges and overflowing read counters refuse. Replay retains
historical metadata; it does not rerun time-dependent source checks or physical
callbacks and therefore does not claim current runtime authority.

Every append carries the loaded event count into the audit-chain transaction.
Independent controllers cannot both commit from a stale head. Conflicts refuse
without automatic retry or a false integrity incident; genuine persistence
failures latch the controller's integrity stop. A reservation or ready append
failure grants no source I/O, a seal failure grants no successful receipt, and a
release append failure grants no returned bytes. Another controller must reload
and honor the durable state; no failed callback implies permission to recover by
adopting, deleting or shrinking an object.

The private `State.ArtifactCustody` projection is omitted from global JSON state.
Immutable event history retains actor attribution and receipts; ordinary audit
summaries describe transitions without exposing artifact bytes or logical names.

## Verification and remaining integration

Automated tests use the actual object store, a mocked event database, synthetic
task turns, synthetic guard callbacks and private temporary files. They cover
charge-before-quota-before-source ordering, actual bytes across controller restart,
read acknowledgments and actor attribution, current human room permissions,
revocation between authorization and release, unknown/full inventories, missing
source evidence, quota/setup acknowledgment failure, sealing failure, immutable
slots/receipts, malformed replay, public event injection refusal, cancellation,
stale and concurrent controllers, shutdown retention and private state omission.

These tests do not connect to PostgreSQL or launch providers, SSH, Docker or real
workers. They do not certify physical quota, complete live inventory, approved
worker volume identity, source quiescence, crash recovery, retention/disposal,
public authenticated download, trusted format/content checks or delivery. Those
production integrations remain required before managed file/bundle dispatch can
be enabled. See [finite manifests](artifact-manifests.md),
[storage accounting](storage-accounting.md) and [release readiness](release-readiness.md).
