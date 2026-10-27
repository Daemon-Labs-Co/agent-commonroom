# Lock-scoped workspace probe physical evidence

The worker-local physical validator now connects retained workspace observation to
controller probe authorization. It inspects actual Docker metadata, the private
creator record and the durable local probe journal. Controller lifecycle frames
alone cannot prove container absence or release account capacity.

This is internal mandatory security infrastructure. It does not expose a public
operation, read source files, enable required file outputs or certify a frozen
source root. There is no optional isolation switch or host execution fallback.

## Trusted assembly

Inside the real creator-ownership callback, construct
`newWorkspaceProbePhysicalEvidence(lease, authenticatedWorker, reviewedImage)` and
supply it to `newControllerWorkspaceProbeAuthority`. Both controller and worker
must be bound to the same retained record and fresh intent. The probe authority
factory now runs under actual creator ownership before the first observation;
its temporary forwarding guard refuses all authorization until setup completes.
Existing ordinary observation and final zero-reference checks remain enforced.

Only a live ownership callback with an actual worker lease can construct this
validator. It receives immutable copies of the record, a callback-scoped lifetime,
the trusted Docker runner and private worker-state directory. Ownership callback
exit cancels that lifetime before the creator lock is released, including failure
and panic paths. A retained handle cannot be revived by passing a new context.
The validator does not reenter the authorization guard or observation mutex.

Trusted configuration must supply the approved authenticated worker, a runner
that targets the approved rootless daemon on every call, its verified Docker root,
and an independently reviewed immutable helper image. An image requested in a
frame is not its own review approval. Missing ownership or changed image refuses.
This implementation is local to the worker; deployed remote authentication and
routing of independently verified evidence to the controller remain integrations.
It is not a remote attestation protocol or a claim against a malicious daemon
administrator, who can race Docker observations.

## Physical checks

Every operation rereads the exact retained creator record, role absence flags,
volume identity and persistent cancellation latch before and after physical calls.
It checks current controller job/company/worker binding, caller cancellation and
ownership lifetime. Source data is never opened.

Ordinary observations require zero engine/tool name references and exact retained
volume shape/labels/nonce/creation time. When an acknowledged probe is in progress,
only its fully verified stopped read-only helper may appear. Before its ID is
bound, a candidate discovered by its exact fresh name is inspected only: no ID is
adopted into history, launched or removed by name. The exact-ID bound acknowledgment
still requires the worker's durable journal, full helper admission and controller
persistence before execution. Previously bound IDs must match; settled reservations
cannot admit a returning helper.

The existing bounded broker checks inspect the helper image, exact labels and
command, UID, read-only root/volume, no credentials/host binds/network/capabilities,
resource limits, volume identity and full job/name/volume inventories repeatedly.
This validator cannot execute a helper: its inspection-only journal refuses every
create/start lifecycle operation.

Lifecycle validation also requires the exact last event in the actual local
journal to match the controller acknowledgment:

- Begin requires no references before helper creation.
- Bound verifies the exact recorded stopped helper and complete inventory.
- An absent finish requires that exact ID to be absent and the retained workspace
  to have zero references. A frame's `absent` boolean is insufficient.
- Uncertain known-ID cleanup uses owned recovery inspection and never grants a
  release. Unknown-ID outcomes may be recorded only with verified zero references;
  an actual unknown attached helper remains blocked and charged.
- Recovery stages inspect only the durably known ID; this validator does not grant
  the separate fresh cleanup authority needed to initiate recovery.

## Configuration, flags and permissions

There is no default-on/off product capability, UI or HTTP command. Controller
human/room/task authorization, audit integrity, account capacity and independent
local journaling remain mandatory in every task feature-flag state. No environment
variable replaces ownership, image review or physical validation. Disabling a
product feature never deletes retained data or refunds uncertain capacity.

A source fork can fabricate ownership handles or replace runners/guards and then
loses these isolation/evidence claims. The supported path constructs ownership
only through the actual worker lease and preserves its final checks.

## Verification and remaining work

Tests connect the real physical validator, controller guard/ledger, private channel,
worker journal and exact-ID cleanup using synthetic temporary jobs and mocked
Docker/database/provider boundaries. Fixtures assert that actual controller begin
and bound records exist before create/start. Cases cover healthy slot release,
extra/returning references, writable helper, volume replacement, forged mount proof,
cleanup failure, lost create reply, wrong image/volume/worker, false exact-ID absence,
unrecorded lifecycle events and expired/fabricated ownership.

These checks do not launch Docker/providers or certify a live Linux image. Remaining
work includes deployed authenticated peer/evidence routing, continuous writer fencing,
source-root freezing and capture, physical storage enforcement, authenticated file
delivery and human review of returned repository changes. Required file/bundle
execution remains blocked until that complete path is verified.

The [isolated source lease](isolated-source-capture.md) now supplies fixed-root
kernel checks plus mandatory live supervisor authorization inside a running capture
helper. The metadata probe itself still returns no source root or lasting lease;
the reviewed capture runtime and quota/custody/delivery assembly remain required.
