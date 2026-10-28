# Supervised isolated artifact reads and recovery

This infrastructure supervises one credential-free object-reader container for an
exact sealed artifact receipt. It advances the path from task output to human
review, but **does not enable artifact downloads or file-task dispatch**. There is
no new user interface, public endpoint or repository permission. A reference URL
is still not an artifact or acceptance evidence.

## Configuration and authority

`NewObjectReadRuntime` requires an admitted `ObjectReadContainerPlan`, authenticated
rootless Docker runner, private attached-process starter, durable lifecycle journal
and pinned `objectcustody.ReadReleaseGuard`. Every dependency is mandatory; nil and
typed-nil dependencies refuse. There is no permissive production implementation.
The plan fixes the reviewed image digest, finite receipt/challenge and retained
custody-volume nonce/creation identity. It accepts no host file path or model command.

The lifecycle journal must independently authenticate the worker and exclusive
creator, fence writers, enforce physical quotas/account capacity and current export
policy. `BeginObjectRead` durably reserves capacity before any daemon query or
creation. `BindObjectRead` durably records the returned immutable container ID
before inspection/start. `FinishObjectRead` records observed absence or uncertainty
under current authority; persistence failure must trip the controller integrity
stop. Implementations must not treat labels or synthetic observations as proof of
physical storage, authentication or quota enforcement.

The release guard pins the human session, worker and exact receipt. It must recheck
current authorization and durably acknowledge release, including after cleanup.
Disabling task-definition editing or an unrelated optional feature does not erase
retained artifacts; current room/export permission, integrity, emergency stops and
retention remain mandatory. Existing controller-session tests enforce that contract.
This supervisor grants no independent human permission and adds no flag override.

Mandatory isolation is security infrastructure and has **no feature flag, off
switch, restart option or host/direct-SSH fallback**. It does not enable any optional
product capability. A source fork can replace the trusted runner, starter, journal
or guard and invalidate containment; these seams are not an untrusted plugin API.

## Successful read sequence

1. Reserve lifecycle authority and check the pinned current read permission.
2. Observe the exact retained volume and two independent empty job/name/volume
   reference inventories. A missing/replaced volume or pre-existing helper refuses;
   Docker cannot implicitly create a missing custody volume.
3. Issue the plan's fixed `create` arguments and bind the exact returned ID durably.
   Verify complete created policy twice, then start only that ID.
4. Verify running admission twice. Receive one finite, rehashed artifact over the
   private supervisor stream, including source authorization and durable source
   acknowledgment through the journal relay. Recheck actual admission before and
   after authorization/acknowledgment callbacks.
5. Wait for actual successful helper exit, then verify exited policy twice. Protocol
   EOF can coincide with successful process exit: the final protocol acknowledgment
   accepts only fully admitted running or successfully exited policy, never a failed
   exit, missing container or changed identity.
6. Close and join the attached process, remove only the exact owned ID, then verify
   repeated absence plus unchanged volume/catalog and no returning references.
7. Durably finish the lifecycle, recheck current release authority through its final
   acknowledgment, and only then return bytes to the trusted caller.

`ReadSupervisedChannel` borrows the attached process stream and leaves it open on
protocol success so process `Wait` is not canceled prematurely. Protocol failure or
cancellation still closes promptly. The original `ReadChannel` keeps its standalone
close-on-return ownership contract. Neither API authenticates a peer from JSON.
The caller owns trusted private session setup; model/provider pipes are forbidden.

`ObjectReadProcess.Wait` must observe actual success and join I/O pumps. `Close`
must safely tolerate concurrent cancellation/closure, interrupt I/O and join pumps.
A process returned alongside a start error is still closed by the supervisor.
Callbacks must cooperate with cancellation and must not reenter their stream.

## Failure and recovery

The runtime is one-shot. Active reads are bounded to 35 seconds; each admission
has a five-second bound, byte transfer has a 30-second bound and cleanup has a
separate five-second context so caller cancellation cannot skip cleanup. Lifecycle
finish gets a fresh, independent five-second background context: an exhausted
cleanup deadline must still allow the journal to durably record uncertainty. A
cleanup timeout does not itself prove journal persistence failed. A genuine finish
failure remains an integrity failure; neither case releases bytes or refunds storage.
Cleanup and finish can together add up to ten seconds after active read cancellation.
The underlying adapters must honor these deadlines. All errors withhold bytes and stop
the affected read authority. Errors after lifecycle admission retain an unconfirmed
outcome; failed finish/cleanup/final acknowledgment cannot turn verified bytes into
a successful release. Stored objects and their charges remain retained.

Lost/invalid creation replies leave an unknown ID: the supervisor records uncertainty
and never discovers/adopts a container by name or deletes an unrecorded identity.
A pre-existing helper is likewise not adopted or removed. Operational reconciliation
of unknown-ID incidents remains required; do not retry by guessing an ID.

For an already **durably recorded exact ID**, construct `ObjectReadRecovery` with
its original plan and an independent `ObjectReadRecoveryJournal`. Begin must
validate the interrupted record's worker/request/image/volume/ID and current cleanup
authority before any daemon query. Recovery checks identity/references/catalog twice,
removes only a present owned ID and reconfirms absence. An owned policy-drifted helper
is eligible for cleanup only; foreign/replaced/conflicting identities refuse.
Finish durably records absence/uncertainty. Recovery is one-shot and cannot start,
read, publish bytes, delete volumes, refund quota or reconstruct a successful read.
Recovery has a ten-second operation bound and a separate five-second finish bound.
An absent helper does not prove delivery, task completion or human acceptance.

## Verification and remaining integration

Tests use synthetic complete metadata, mocked Docker/lifecycle/permission boundaries
and joined private pipes carrying real finite protocol bytes with the source journal
relay. Coverage includes success ordering, EOF/exit timing, all lifecycle failures,
permission refusal, cancellation, invalid/lost IDs, corrupted bytes, policy drift, cleanup and
final acknowledgment failures, retained uncertainty and known-ID recovery conflicts.
Deterministically expired/canceled cleanup contexts verify that finish still receives
a live bounded context and records uncertainty; actual finish failures remain visible.
The existing controller-session tests separately cover human room/export authority
and retained reads with optional flags off. No database, Docker daemon or provider
is launched; no private data or credentials are used. These tests do not certify a
deployed container or live P0 acceptance.

The [worker journal](worker-object-read-journal.md) now supplies local durable
records and shared capacity under a retained-workspace creator lease.
The [controller reader ledger](controller-object-read-ledger.md) adds private
account-wide lifecycle accounting, with independent evidence still mandatory.
Still required: authenticated production runner/starter and controller session
assembly, actual physical quota/account/release witnesses and storage provisioning,
unknown-ID incident handling, reviewed rootless images and live acceptance, public
artifact delivery and human review integration. Required file-task dispatch remains
blocked until the full execution and delivery boundary is assembled and verified.

See [reader admission](object-reader-admission.md), [fixed helper](isolated-object-reader.md),
[controller read sessions](controller-artifact-read-session.md) and
[mandatory isolation](architecture.md#mandatory-container-isolation-design-contract).
