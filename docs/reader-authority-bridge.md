# Artifact-reader authority bridge and worker assembly

This internal bridge connects the isolated reader's local journal, controller
account ledger and read-only daemon observations. It preserves authenticated
identity over a supervisor-owned private stream and assembles the fixed container
reader with mandatory lifecycle and byte-release guards. It does not add a public
download endpoint, make file-task dispatch ready, or establish network authentication.

## User and operator behavior

No new UI, public API, environment setting or feature switch is exposed. Retained
artifact permissions and human acceptance are unchanged. A successful internal read
still means only that verified bytes reached the trusted caller; it is not proof
that a user downloaded them or accepted the task.

An interrupted exchange can leave a reader charged locally, at the controller, or
both. Keep its private local record and controller audit history. Never delete a
record or reset a stream sequence to clear a refusal. A known-ID recovery requires
fresh current cleanup authority and the exact retained identity. Lost-ID incidents,
stop-marker reconciliation and network authentication still require production
integration; reconnecting an uncertain session is not a supported recovery shortcut.

For a normal read, the sequence is:

1. Trusted setup authenticates the human and worker independently and holds the
   retained-workspace creator lease. No model/container may choose these identities.
2. Construct local/physical evidence with the approved same-daemon runner and an
   independent quota/export/creator witness; construct the pinned controller session.
3. Connect the private authority client/server on a supervisor-owned authenticated
   stream. Its directory/socket must never be exposed to model or tool containers.
4. Construct `newWorkerObjectReadRuntime` with the retained lease, finite intent,
   same-daemon fixed-ID starter, authority client and independently pinned byte-release
   guard. Missing dependencies refuse before daemon effects.
5. `Run` reserves local/controller capacity, binds the actual created ID before
   start, validates actual container policy and verifies the finite byte protocol.
6. Successful exit, joined stream, exact owned cleanup, repeated absence checks,
   both journals and current final release permission precede returned bytes.
7. The caller closes and joins its owned authority transport on every path. Success
   neither refunds persistent object storage nor updates task acceptance.

## Pinned controller session

`newControllerObjectReadSession` captures the original authenticated principal,
worker alias, complete retained worker binding, exact receipt/image/volume intent
and a non-nil `controllerObjectReadSessionEvidence`. Later caller contexts cannot
replace that principal. Candidate transitions must be the actual latest valid
checkpoint, including complete local history at acknowledgment. Truncated histories,
substituted requests and old begin events cannot become new authority.

Operation authorization and durable acknowledgment are separate. Authorization
checks current room/human/receipt policy and independently observed evidence, then
reloads the audit head; it grants no capacity by itself. Acknowledgment compares the
exact private local history with the controller candidate, performs independent
lifecycle checks and commits under the same audit revision fence. A stale revision
refuses; genuine controller persistence failure trips the global integrity latch.

Session revocation is scoped, nonblocking and permanent for reads and recovery.
Ordinary stream closure does not create a global integrity incident. A canceled
session may still record a `finish` with **absence false**, using a fresh bounded
call context and current human/creator authority. This preserves uncertainty only:
it cannot confirm absence, release capacity, export bytes or revive recovery.
An unavailable control connection cannot deliver even this metadata; local charges
remain retained. Fresh recovery must be independently authorized.

## Private control protocol

The transport carries metadata only, never artifact bytes, provider messages,
credentials, model commands or host paths to execute. It requires an already
authenticated supervisor-owned `io.ReadWriteCloser`; it is not itself a handshake,
listener, credential store, public HTTP endpoint or remote worker deployment.

Frames use a four-byte big-endian length followed by canonical JSON, bounded to
32 KiB. Requests and responses carry `protocol: "object-read-v1"`, version1 and a
strictly increasing sequence. Unknown/duplicate/noncanonical fields, cross-protocol
frames, wrong versions/sequences, substituted intents/events/histories and unexpected
operations refuse. Only `authorize` and `acknowledge` operations exist. The session
allows at most64 exchanges over two minutes; each active client exchange is bounded
to five seconds. Calls are serialized; queued calls do not gain permission from an
older acknowledgment. Complete finite histories are independently validated.

Denied, lost, malformed or canceled exchanges poison and close the one-shot link
and revoke scoped authority. A single bounded read pump also observes idle EOF
without competing with frame parsing. The assembled runtime follows both the
control authority lifetime and actual creator lifetime: losing either cancels
active I/O and withholds bytes, even between protocol exchanges. Stop callbacks must be local, nonblocking and must not
reenter the link. The controller guard must not reenter server/worker admission
locks or call the same stream while the worker awaits its response. The client owns
its stream; invalid construction does not transfer that ownership. The caller must
join the serving goroutine rather than leave it behind after a failure.

## Private local and physical evidence

`newWorkspaceObjectReadEvidence` requires an active retained-workspace creator
lease, approved worker/runner/root/image/volume intent and non-nil
`objectReadSessionWitness`. It checks recorded job/task/execution binding, unchanged
retained worker metadata/cancellation latch and exact private local reader records.
Checks compare local state before and after callbacks/daemon inspection, with
independent witness validation repeated on both sides. Callback records are detached.

Operation checks admit only the next valid local transition or the exact latest
one after persistence. Lifecycle acknowledgment requires already-persisted matching
history. Local-only `confirmed`/`recovered` markers are stripped only for comparing
the controller history; they cannot authorize a new lifecycle. Malformed, changed,
foreign, missing or stale records refuse. No lease/admission lock is reacquired by
callbacks while the journal holds it.

Daemon observations are read-only:

- Begin requires the exact retained custody volume with no reader/job/volume/name
  references, including independent catalog checks.
- Bound requires the exact recorded ID in the fixed created-container policy,
  checked in two independent passes before it may start.
- Finish/recovery inspect only the separately recorded immutable ID and exact owned
  labels/image/volume references. Claimed absence must actually be observed. Owned
  policy drift can be inspected for cleanup but grants no read permission.
- `ObjectReadContainerPlan.InspectRecorded` never creates, starts, executes, removes,
  discovers or adopts a helper. Unknown IDs cannot be substituted by name.

Inspection is not physical quota enforcement or authentication. The independent
witness must still attest actual creator ownership, approved daemon/worker/image,
writer exclusion, custody, current operation-specific export/cleanup rights, real
physical quotas and account admission. No permissive implementation is provided.
Evidence lifetime follows the actual creator lease; it cannot be reused afterward.

## Cleanup-only worker assembly

`recoverRetainedWorkerObjectRead` binds a fresh authority lifetime, actual retained
creator and the existing private known-ID journal to the cleanup-only runtime.
It receives no starter, reader, export guard, volume-deletion or storage-refund
capability. Present and already-absent recorded helpers can settle under repeated
independent evidence. A failed cleanup retains the same recovery checkpoint and
capacity; a newly authenticated cleanup session can retry that exact known ID.
Wrong/unknown IDs, denied authority and changed ownership refuse before effects.
A worker integrity-stop marker still requires independent incident reconciliation;
this helper cannot clear it or turn a cleanup session into new read permission.

The outer creator callback must join runtime/recovery work before relinquishing
its actual lease. Cancellation is not an ownership substitute or proof of absence.
Private control transport closure/join remains the caller’s responsibility.

## Flags and execution boundary

This is mandatory security infrastructure with no optional-feature flag, off switch,
per-connection exception or restart setting. Task-definition editing flags may be
off while current authorized retained reads and cleanup remain interpretable. They
never bypass authentication, audit, retention or export restrictions. The runtime
cannot fall back to host file reads, local/SSH model execution or a different daemon.
Source forks that replace guards, witnesses, starters or persistence invalidate
these guarantees. A starter must attach only the exact durably bound ID on the
same authenticated rootless daemon. `dockerCLIRunner.StartObjectRead` now provides
the fixed-ID `docker start --attach --interactive` adapter after per-call rootless
local-daemon validation, with reduced environment, sanitized errors/stderr, retained
final stdout through Wait and joined process ownership on Close. It shares the
existing fixed-helper process machinery; it cannot select a command from a request.
Its deployment and reviewed-image behavior still require live verification.

## Verification and remaining integration

Tests exercise pinned identities, current policy/revocation, complete history,
independent operation/acknowledgment checks, canceled uncertainty without release,
failed audit, private framing/canonical sizes/sequences/protocol isolation, poison/
cancellation/closure, frame budget and detached callback data. Physical tests cover
pre/post stages, exact IDs, false absence, writable mounts, volume replacement,
extra references/catalog entries, changed/corrupt local history, dead creator
lifetimes and independent witness denial/revocation. Observer tests verify no
lifecycle effects or byte-release callbacks occur.

An assembled synthetic journey connects the actual local journal, private control
protocol, controller session/account ledger, local/physical evidence and fixed
reader runtime with current controller byte-release policy. It verifies actual
finite protocol bytes and withheld data after cleanup/exit/Close/export/audit
failures. Recovery integration covers present/already-absent IDs, failed cleanup
with fresh-session retry, denied/wrong/unknown IDs, foreign observations and editing
flags off, without exporting bytes or refunding storage. Idle control EOF and creator
cancellation withhold active reads. Synthetic subprocess tests cover final buffered
stdout, sanitized failure and cancel/Close joining for the fixed reader CLI stream.
All DB and Docker boundaries are mocked; channel helpers and filesystem fixtures
are synthetic/temporary, with no real provider, credential or private user data.

Still required: deployment-authenticated transport setup, real physical quota/
custody/export/creator witnesses, reviewed-image and same-daemon starter deployment
verification, production service assembly, independent object-only ownership after
workspace disposal, unknown-ID/
stop-marker incident reconciliation, public delivery/human review wiring and live
rootless acceptance. Required file-task dispatch remains blocked. These synthetic
checks are not live P0 gate observations and certify no production readiness.

See [controller ledger](controller-object-read-ledger.md),
[worker journal](worker-object-read-journal.md),
[reader supervision](object-reader-supervision.md),
[release sessions](controller-artifact-read-session.md) and
[mandatory isolation](architecture.md#mandatory-container-isolation-design-contract).
