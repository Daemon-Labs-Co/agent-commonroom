# Guarded retained-workspace probe

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


`probeRetainedWorkerWorkspace` assembles the real creator job lock, retained-volume
admission, durable local probe journal and isolated read-only helper into one
internal operation. It returns only a completed mount-observation report. It does
not export source bytes, return an open mount/root or implement `FrozenLease`.
Managed required file/bundle dispatch remains blocked.

## Invocation and permissions

Trusted worker/controller assembly supplies the exact reconciled worker record,
private worker job directory, approved rootless Docker runner/data root, reviewed
local helper image ID and a mandatory `workerWorkspaceProbeGuard`. The operation
generates its own fresh challenge; callers cannot reuse an earlier probe report.
The image is an immutable local `sha256:` ID, never a model-supplied image/path or
an automatic pull. Build/review instructions are in the
[helper guide](readonly-workspace-mount-probe.md).

The guard combines current authenticated human/task/export authorization, durable
bounded-observation acknowledgments, controller lifecycle audit, account-wide
admission and integrity-stop propagation. Missing/typed-nil guards refuse; no
permissive default or production transport is assembled. Acknowledgments must
bind the original authority, even when cleanup uses a fresh context. Callbacks
must not re-enter creator/admission locks. See the
[journal contract](workspace-probe-journal.md) for durable guard requirements.

No public route, UI setting or optional isolation switch is added. Existing task
flags cannot bypass these checks. Users cannot enable file export merely by
constructing the operation or changing configuration.

## Complete guarded lifecycle

1. Acquire the actual creator job lock without waiting for or stopping a live
   creator. Require reconciled exact role absence, the cancellation tombstone and
   unchanged retained volume with no references.
2. Create the local journal and fixed probe inside that ownership interval.
   Persist/acknowledge intent and capacity before Docker creation.
3. Bind the exact created helper ID before execution. At authorization boundaries,
   allow that ID alone in job/name/volume inventory, inspect its complete stopped
   read-only policy and repeat exact volume identity/reference checks. Both fixed
   creator role names must remain absent. Ordinary observation `Check` still
   requires zero references and never gains this narrow exception.
4. Run the fixed isolated kernel-metadata helper, verify its bounded report and
   exited policy, then remove and observe absence of its exact container ID.
5. Persist cleanup outcome and independently recheck current authority, creator
   record, cancellation, role absence and zero references. Confirm local cleanup
   according to journal semantics. Run a final ordinary observation check before
   releasing creator ownership and returning the report.

Every Docker call uses the same supplied trusted runner. Creator ownership is
held through journal acknowledgment and fresh-context cleanup, including parent
cancellation. No private ownership handle or journal adapter escapes to callers.
Failures return an empty report and retain the workspace/history. Unknown create
replies are never adopted by name; their pending slot and explicit uncertainty
remain for trusted recovery. Audit persistence failure retains the integrity stop.

The ordinary final check can deny a report after cleanup was confirmed. That
withholds report authority without undoing acknowledged physical absence or
re-charging a removed helper. A role returning, another volume reference, changed
volume nonce/time, writable/root policy drift, malformed report or cleanup failure
all refuse the operation. It never removes or refunds the workspace itself.

## Verification and limits

Integration tests execute the real admission lock, observation, journal and probe
runtime with mocked Docker/controller guards and temporary synthetic worker files.
The fixture first seeds and reconciles both real recorded role IDs; it does not
start from vacuous cleanup. Each mocked Docker/controller call verifies creator
ownership is still held. Tests cover durable ordering, healthy confirmed cleanup,
revocation including the final observation, extra references/returning roles,
policy/volume drift, cancellation, lost create replies and audit/cleanup failures.
Broker tests independently retain the ordinary zero-reference contract and
require an exact stopped read-only helper. No live database, Docker daemon,
provider, SSH, credentials or private data is used.

This closes internal creator-lock/probe/journal assembly, not production controller
transport. Current authorized guard integration, account-wide physical admission,
continuous reference enforcement, reviewed live Linux image/mount acceptance,
unknown-ID/incident recovery and retention, physical storage quota, actual frozen source-root
binding and authenticated file delivery remain outstanding. Snapshot inventory
checks cannot prevent a trusted daemon administrator from racing volume names or
adding/removing references between observations. The exited-and-removed probe is
not a persistent source capability. No live P0 gate is certified.

Source forks can replace runners, guards or private assembly and thereby invalidate
managed isolation guarantees. Unsupported live platforms retain mandatory refusal;
there is no host, direct local/SSH launch or automatic uncontained fallback.

[Recorded probe recovery](workspace-probe-recovery.md) now reconciles interrupted known-ID helpers under creator ownership; unknown create replies remain unresolved.
