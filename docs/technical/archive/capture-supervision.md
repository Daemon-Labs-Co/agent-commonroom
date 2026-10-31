# Supervised private capture lifecycle

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


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

## Retained worker ownership assembly

Trusted worker setup can now call `newOwnedWorkerCapture(observation, request,
image, objects, lifecycleGuard, sourceAuthority, custodyJournal)` and synchronously
call its `Run(ctx)` inside `withWorkerWorkspaceOwnership`. This joins the fixed
container plan and supervised runtime to the real private worker lifecycle journal.
For example, the request selects the exact finite task members from the retained
workspace; the journal records the accepted creation intent and exact helper ID
before the runtime starts the private capture stream.

The observation's ONE runner must implement both Docker inspection/creation and
`CaptureStarter`. The constructor refuses a runner without that starter instead
of accepting a separately supplied process adapter. The standard Docker adapter
checks approved rootless selection for each invocation. This binding prevents
accidental assembly with separate daemon adapters; it does not authenticate a
dishonest injected runner. All three supplied authorities remain mandatory, with
nil and typed-nil refusal. Source authority and custody quota/export enforcement
must be real independent evidence, never synthetic callbacks in deployment.

The lifecycle guard must also supply `CaptureAuthorityLifetime()`: a live pinned
control-operation context, canceled on loss of authenticated authority even between
calls. Owned capture follows both that context and creator ownership. The
[private authority client](capture-authority-transport.md) observes idle EOF and
joins its sole bounded read pump on explicit stop; no successful old exchange can
replace current control authority. Failed captures revoke that scoped guard before
their tracked ownership operation finishes.

[Authenticated capture sessions](capture-authentication.md) bind that private
client/server link to approved certificate identities and the original controller
session. TLS/hello agreement never replaces current physical or account evidence.

Request members are copied into immutable plan/journal bindings. Construction does
not start a process or persist creation intent. `Run` is one-shot, uses the shorter
caller/owner lifetime and returns only the runtime's complete private metadata
report after its existing exit, exact-ID cleanup and final audit checks. A denied
source operation can still have confirmed helper cleanup: that settles the helper
slot only, never proves capture success or refunds artifact storage. Lost creation,
failed binding acknowledgment and uncertain cleanup retain their journal evidence.

Ownership teardown first refuses new admissions, cancels the owner lifetime and
joins admitted capture operations BEFORE releasing the creator lock. This applies
on callback failure or panic as well as normal return. It also prevents a retained
capture handle from starting after its callback ends. Operations must honor
cancellation and join their own streams; arbitrary blocked trusted callbacks cannot
be forcibly terminated by a Go context. Do not invoke ownership teardown from
inside a dependency on which that same operation is waiting.

The same tracked-operation boundary now covers isolated reads and known-ID reader
recovery. Cancellation occurs before waiting for the state mutex so a journal's
current-authority callback can observe it and release that mutex. See the
[reader bridge](reader-authority-bridge.md#private-local-and-physical-evidence) for
byte withholding, release revocation and cleanup retention behavior.

Synthetic assembly tests exercise the actual plan/runtime/protocol and worker
journal through mocked daemon/source/custody/controller boundaries. They cover
success, current authority denial, lost creation, failed acknowledgment, cleanup
failure, replay, expired/missing dependencies and creator-lock retention until a
canceled operation joins. No real provider, Docker or database is used.

This mandatory infrastructure has no optional security toggle or public route.
Existing task feature flags and human/account permissions must still be enforced
by the supplied current controller authorities in both flag states. The assembly
does not install them, provision custody volumes/quotas, expose file-task dispatch,
attach manifests to task results, or prove live independently retained bytes.
Production bootstrap, authenticated authority transports, complete inventory and
physical enforcement, output format checks and delivery/review remain unfinished.

Interrupted owned captures can now use the [retained worker recovery assembly](capture-recovery.md#retained-worker-assembly)
with fresh cleanup authority and an already durable exact ID. It joins reconciliation
before creator unlock and can explicitly resume the same interrupted recovery
checkpoint; it cannot restart capture, publish results or refund stored artifacts.
