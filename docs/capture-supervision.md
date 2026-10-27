# Supervised private capture lifecycle

`toolbroker.CaptureRuntime` connects the fixed capture-container plan to the private
capture protocol. It is internal mandatory isolation infrastructure with no off
switch, host launcher or direct SSH fallback. It does not enable file-task dispatch
or public delivery. The [worker stream adapter](capture-worker-stream.md) supplies a fixed Docker
starter; production authority/quota assembly remains incomplete.

## Configuration and use

Trusted worker code constructs a `CaptureContainerPlan`, then calls
`NewCaptureRuntime(plan, runner, starter, lifecycle, authority, custodyJournal)`.
Every dependency is mandatory, including typed-nil rejection. `Run(ctx)` is one-shot.
For example, the existing retained task workspace and a separately provisioned private
custody volume can be described by the plan; the runtime then supervises only the
reviewed fixed helper. Models cannot supply commands, paths, Docker options or pipes.

The runner and starter must target the same approved rootless daemon. The starter
may start only the supplied immutable ID and must return exclusively owned, non-TTY
helper stdin/stdout; it may not create/replace containers or borrow provider streams.
Its process `Wait` must observe actual successful exit and join pumps; `Close` must
promptly interrupt I/O and join pumps, including on failure. A returned stream remains owned even when start also returns an error.
The runtime closes it before container cleanup. It uses
`ServeSupervisedCaptureChannel`, which leaves a successful attached stream open
through `Wait` and both exited-container admission passes. Closing an attached
command cancels it, so protocol completion alone must not close that process.
Protocol errors/cancellation still close promptly; the runtime owns final closure
and joining on all outcomes. The standalone `ServeCaptureChannel` retains its
owned-stream close-on-return contract for callers without a supervised process.

The lifecycle is:

1. `BeginCapture` authenticates current human/task/export authority, holds exclusive
   creator ownership, fences all writers, verifies the reviewed image, establishes
   physical source/object quotas and durably reserves creation capacity/intent.
2. Refuse an existing helper name; attempt fixed creation once. Unknown or lost
   creation replies retain uncertainty. Never discover/adopt/remove by name.
3. Check exact created-container policy and reference inventories, then durably
   `BindCapture` its exact ID before starting.
4. Start the private stream and recheck the running container. Before every protocol
   authority acknowledgment, independently repeat running physical admission and
   invoke the supplied current-authority guard. Reserve/seal operations retain the
   private protocol's durable custody ordering. Callbacks must not reenter this stream.
5. After the complete capture report, wait for actual successful helper exit and
   independently verify the exited container's policy/inventories.
6. Close the stream; with a fresh bounded cleanup context, remove only the exact
   owned ID. Repeat absence/reference/volume identity/catalog checks twice. Finally
   `FinishCapture` durably records absence or uncertainty and rechecks current authority.

The operation has a35-second outer deadline; protocol sessions and individual
physical inspections impose their own shorter bounds. Cleanup has a separate
five-second deadline so cancellation can still stop the exact owned helper. The
caller must retain creator ownership and capacity until durable reconciliation.

## Failure and permission contract

A failure after accepted creation intent produces a zero report and an unconfirmed
custody error, including authority denial, lost create reply, missing binding,
stream/protocol failure, unsuccessful exit, ownership/policy drift, returning
references, replaced volumes, cleanup failure and final audit/authority denial.
There is no automatic retry, refund, adoption or volume deletion. Partial objects,
individual seals and charges remain retained. A failed begin performs no creation;
its trusted journal must retain any uncertain persistence outcome itself.

Ownership inspection used for stopping admits only the same ID/name/image/labels;
it deliberately permits stopping an owned policy-drifted process. It does not
permit starting or exporting from that process. Unknown IDs are never guessed.
An absent container alone is insufficient: both retained volumes and their creation
identities/catalogs must still match, with no attached job/name/volume references.

## Verification limits

Tests use mocked Docker, starter, quota/audit and authority boundaries, synthetic
metadata/receipts and an in-memory private protocol stream. They prove lifecycle
ordering (including wait/exited verification before close), cancellation/error closure
and report suppression, not actual storage bytes, kernel quotas, physical
shutdown, image provenance or live Linux readiness. Actual byte transfer is covered
separately by the synthetic source/custody tests.

Still required: reviewed image build/acceptance, deployed daemon/private-stream
composition, durable creator/account integration, physically enforced quotas,
object-volume initialization/ownership, crash recovery, content/format validation,
authenticated delivery and human review. A dishonest source-fork dependency defeats
isolation claims. Never substitute permissive guards to enable production tasks.

The [interrupted recovery contract](capture-recovery.md) provides explicit exact-ID
reconciliation with separate durable cleanup authorization. Normal creation intent
now passes its reviewed image explicitly to `BeginCapture` for durable binding.
