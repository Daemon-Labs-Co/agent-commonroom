# Durable worker artifact-reader journal

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


This internal journal records one isolated reader attempt before launch and retains
its container-capacity charge until cleanup is independently acknowledged. It is
part of the mandatory execution boundary, not an artifact-download feature. It
adds no public endpoint, UI, provider operation or permission to accept a task.

## Configuration and trusted dependencies

`newWorkerObjectReadJournal` requires an active retained-workspace creator lease,
a finite `objectread.Request`, reviewed image digest, exact custody-volume identity
and a non-nil `workerObjectReadJournalGuard`. The job, repository and room must match
the pinned worker record. Pure `toolbroker.ValidObjectReadMetadata` validates finite
receipt/image/volume identities without fabricating an observed Docker data root.
The actual reader runtime still requires independently authenticated admission.

This initial worker adapter deliberately requires the original workspace to remain
retained and unchanged. It cannot serve an artifact after workspace disposal, even
if a custody object still exists. Independent object-only ownership and controller
integration remain future work; this restriction must not be bypassed with a host
read or a fake retained-workspace observation.

The guard MUST independently authenticate current human export or cleanup authority,
exact sealed receipt/custody, approved worker/image/daemon/volume, physical quotas,
writer exclusion and account-wide capacity. Its acknowledgment must durably mirror
the exact local transition before returning. There is no permissive or deployed
default guard/transport. Callbacks must honor cancellation, must not reenter the
creator/admission locks, and must not treat worker metadata as authentication.
Mutable event slices and worker records are detached before callbacks.

No feature flag disables journaling, creator fencing, audit or capacity checks.
This is mandatory security infrastructure. Optional task/editing flags grant no
reader authority; retained reads still require current human/room/export permission
through the independently assembled controller release guard. This journal alone
cannot provide those permissions. Source forks can replace guards/persistence or
skip the isolated runtime and invalidate the security contract; no supported host
or direct-SSH fallback exists.

## Durable records and capacity

Records live in the private worker state directory as
`.object-read-<job>-<challenge>.record`. They contain versioned immutable request,
image and volume metadata plus a finite lifecycle sequence. Reads require canonical
JSON, private regular files, matching names/identities, valid transition order and
bounded sizes/counts: 16 KiB per record, six lifecycle events, 1024 reader records
and 16384 total directory entries. Reaching those bounds refuses new admission;
no automatic record deletion or retention bypass is provided. Unexpected/malformed
reader records block admission rather than silently disappearing from accounting.

The actual creator lease remains held during every transition. A shared admission
lock serializes readers with normal jobs, probes and capture helpers. Before begin,
the journal checks aggregate worker capacity and rejects unresolved same-job reader
or capture attempts. Capture admission symmetrically rejects an unresolved reader.
Existing worker directories without reader records retain their prior accounting.

The normal sequence is `begin -> bound -> finish -> confirmed`. Begin is locally
durable and controller-acknowledged before the runtime may create a container;
bound records the exact immutable ID before start. Finish records the observed
absence/uncertainty, then obtains durable controller acknowledgment and fresh current
authority. Only after those succeed is a final local confirmation written and the
helper slot released. This releases no object-storage charge and proves neither
user delivery nor human acceptance. A new authorized read requires a new challenge.

Unknown/lost IDs, failed cleanup, failed local writes, failed controller writes,
revocation or incomplete confirmation retain the slot. Local/controller persistence
failure writes the shared worker integrity-stop marker and independently revokes
live guard authority. Failure to persist the marker never grants permission to
continue. Invalid requests/ordinary permission refusals do not invent a persistence
failure. A failed confirmation write likewise retains the unresolved record.

## Cancellation and recovery

A canceled read cannot begin, bind or claim confirmed cleanup. While the actual
creator lease remains active, a finish may still persist and acknowledge cleanup
uncertainty using the runtime's fresh bounded finish context. It then refuses to
confirm/release the slot if the pinned read lifetime was canceled. Recording this
metadata grants no export permission. Once the creator lease closes, even finish
refuses; an escaped journal cannot regain ownership with a fresh context.

Known-ID recovery requires fresh current cleanup authority and the same exact
request/image/volume/ID. If cleanup fails or times out after recovery begins, a new
supervisor may resume that exact already-begun recovery. Each retry repeats private
record validation/persistence, current authorization before and after controller
acknowledgment, and durable controller acknowledgment. It preserves the same finite
local history and capacity charge; it neither appends duplicate recovery stages nor
reuses prior permission. Guards must support re-acknowledging the same logical
recovery checkpoint under current authority. A different ID or changed intent
refuses; settled recoveries cannot be reopened. Only a previously bound recorded
ID can enter
`recovery -> recovery-finish -> recovered`. An unconfirmed recovery stays charged;
no name discovery, unknown-ID adoption, container start, object read, volume deletion
or storage refund is available. The guard must independently verify actual cleanup;
a boolean supplied by a model is not absence evidence.

## Verification and remaining work

Tests use temporary synthetic worker records, private creator/admission locks,
mocked guards and injected persistence failures. They verify persisted-before-ack
ordering, detached callback data, shared capacity, mutual reader/capture exclusion,
normal confirmation, unknown-ID retention, exact-ID recovery, canceled finish,
escaped lease refusal, identity/room substitution, malformed records, local/controller
write failures and fresh-attempt rules. No real database, Docker daemon, provider,
private message or credential is used. Existing capture tests exercise compatibility.

The [controller reader ledger](controller-object-read-ledger.md) now provides
private account-wide lifecycle accounting with mandatory independent evidence.
It is not an authenticated transport or deployed guard.

The [authority bridge](reader-authority-bridge.md) now supplies pinned sessions,
private control frames, local/physical evidence and worker/runtime orchestration.
Still required: deployment authentication, real physical quota/custody/release
witnesses and reviewed starter deployment,
object-only ownership after source disposal, unknown-ID incident reconciliation,
reviewed rootless deployment/live acceptance and public artifact delivery/human
review. File-task dispatch remains blocked. Synthetic journal tests are not live
release-gate evidence.

See [reader supervision](object-reader-supervision.md),
[reader admission](object-reader-admission.md),
[controller read sessions](controller-artifact-read-session.md) and
[mandatory isolation](architecture.md#mandatory-container-isolation-design-contract).
