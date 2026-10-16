# Reviewed engine-container runtime

The Codex engine role now has a real DockerRunner-backed create, verify, inspect
and start implementation. It is internal launch infrastructure, with no bypass
flag, default image or uncontained fallback. Production entry-point orchestration
and live acceptance remain incomplete; this is not a release-ready provider path.

## Trusted inputs

The worker supplies the original registered scope, reviewed immutable engine
image, approved provider-gateway socket path, connection credential lease and
local rootless Docker runner. `newWorkerCodexEngine` requires the original
worker job context and rejects a different job or role. Pair assembly additionally
requires equality of the complete engine/session scope, including room, provider,
participant, connection, repository and normalized deadline.

The generic runtime cannot prove private host-file custody or gateway ownership
from path spelling. The production worker must create a fresh private gateway
socket, verify its type and private ancestry, preserve credential leases and
supply its own approved provider gateway. Neither paths
nor runner/image options may come from a model or public request. Supplying a
Docker socket, host runner or unreviewed image through source changes violates
the mandatory boundary; no source-level type can make a hostile host trustworthy.

### Private provider gateway

The engine constructor requires a worker-owned gateway handle, rather than an
arbitrary socket path. `newWorkerEngineGateway` requires the original worker job
context and full Codex scope. It creates a fresh random directory under an
existing worker-owned mode-0700 staging root, then creates its own gateway socket
there. The root and ancestors are trusted deployment configuration; the root
cannot be a symlink or public directory. Unix socket path length errors fail
closed, with no shared-directory fallback.

The socket is mode 0666 so the engine's subordinate rootless UID can connect
through its sole bind mount. Its host parent stays mode 0700 and is not mounted,
so other host accounts cannot traverse it. This does not permit a public host
socket directory, credential-file mount or unrestricted network access. Existing
exact provider-host, public-address, connection and request limits still apply.
The worker account and host remain trusted.

Construction, engine create/verify/inspect/start and turn attachment recheck full
scope, original lifetime, private directory identity and socket inode/type/owner.
A replaced or closed gateway is refused. Cancellation closes the listener and
connections; a fresh caller context cannot revive it. Closing the gateway does
not prove container absence or release account reservations.

Cleanup disables automatic pathname unlinking, closes the owned listener, and
unlinks only the recorded socket inode inside the recorded private directory.
It removes the directory only when empty. Changed permissions/inodes or unknown
files return an unconfirmed-cleanup error and preserve unexpected content.
Production startup still needs to retain and report these handles and errors
through its complete recovery workflow. No live rootless socket bind was tested.

### Connection credential custody

`acquireWorkerEngineCredential` is internal preparation infrastructure. It takes
the same nonblocking connection/provider lock as sign-in and sign-out, then reads
a bounded private single-link regular file using descriptor-bound no-follow
checks. Only the reviewed subscription schema is admitted: exact field names,
no duplicate fields, `auth_mode: chatgpt`, a nonempty access token and no non-null
API key. Unknown schema fields require compatibility review. Failures return
generic errors and release the lock without publishing credential bytes.

The lease provisions exactly one private authentication frame into the inspected
engine attachment before callback-driver initialization. It requires the original
worker context and matching job; partial writes consume the attempt and are never
replayed. The frame is `COMMONROOM-AUTH-V1\n`, a four-byte unsigned big-endian
length, then 2–65536 credential JSON bytes. The bootstrap reads exactly that
length without buffering ahead, writes its own exclusive UID-1000 mode-0600 file
inside private tmpfs, and leaves subsequent stdin for the app-server. There is no
host credential bind mount or mapped-host-UID credential file to relax.

Cancellation deliberately retains the connection lock until matching durable
fencing evidence settles both roles (absent or never began creation). A retained
source volume alone does not retain the lease after the roles settle. Release
clears the owned raw buffer and is idempotent; this does not guarantee erasure of
all runtime copies. The worker Turn method provisions the frame before invoking
CodexTurn. Its surrounding Run still owns all cleanup.

Production admission is not yet wired. Locks disappear when the worker process
dies: production recovery must reconcile journaled engines before issuing a new
lease. A filesystem lock alone is not crash recovery or distributed custody.
Exporting and persisting refreshed tokens remain unimplemented. Provider-gateway
ownership and live private-stdin/Docker compatibility still need acceptance.
There is no credential-isolation bypass flag.

### Owned worker operation

`workerPairSupervisor.RunCodex` connects the existing pieces for an already
admitted worker job. Its caller supplies the original authenticated ToolSession,
private capability token, approved source-bundle reader and trusted worker config
(engine image, authentication root, private gateway root, prompt/model). This is
an internal operation, not a public endpoint or a substitute for controller
authentication, authorization, feature flags or account quota admission.

Validation refusal retains caller ownership. Once accepted, the one-shot operation
owns the session and bundle, acquires credentials, creates the private gateway,
assembles the original tool runtime and reviewed engine, runs one callback turn,
and performs fenced shutdown on all outcomes. The original supervisor context
bounds resource admission; the request context separately bounds execution.
Gateway errors, source-reader close errors, turn errors and shutdown errors remain
errors. Repeated bundle Close calls preserve the first close error.

Credential acquisition scans bounded worker records while holding the connection
lock. The current full-scope record must exist and have no role creation intent.
Prior jobs on the same connection/provider must have settled roles (or full legacy
shutdown proof); expiry alone does not release them. Invalid/missing journal data
refuses acquisition. A lost process lock cannot turn a surviving engine into
permission for another engine to use the connection. Journal ownership still relies
on the existing worker admission/creator leases; source changes that launch before
credential admission violate this ordering.

The supervisor retains an unresolved credential lease until a later Shutdown
obtains matching role proof. Once both roles settle, it releases the lease even
when source retention keeps the overall job unresolved. Process death still needs
the journal/recovery checks above; memory-held locks do not survive a crash.

Tests cover owned preparation failures, validation ownership, journal crash
survivors and partial cleanup/recovery. The combined synthetic callback test now
also drives RunCodex through real session/runtime/bridge code against mocked
Docker metadata and local protocol peers: fresh volume/pair creation, bootstrap,
private credential frame, callback dispatch, framed audit intent/outcome and
independent role cleanup. It checks forbidden callbacks, audit intent/outcome
failures and engine disconnects. A failed audit never returns success text; an
intent failure executes no helper. The successful model turn still returns an
overall lifecycle error while its source workspace remains retained; tests do
not fabricate workspace disposal to obtain a green lifecycle result.

Mocked reported metadata and synthetic protocol responses do not prove image
contents, real Docker resource enforcement, OAuth or provider compatibility.
Those live production checks remain acceptance gaps. Public entry-point
wiring, refreshed-token persistence, reviewed release images, storage quotas,
commands/artifacts and the other P0 requirements remain unfinished. Managed
development admission stays gated; this operation does not enable it.

## Before creation

The runtime consumes one creation attempt and requires successful fixed-name
inventory with no existing container. It never adopts, starts or resumes a used
container. A failure or lost reply requires fenced recovery and a new job rather
than automatic replay.

Image inspection must report a valid immutable image ID and the expected
UID/workdir/entrypoint, with no inherited command, implicit volumes, OnBuild hooks
or unknown environment keys. Known environment values are all overridden by the
fixed create policy. For example, inherited LD_PRELOAD is refused before create;
NODE_OPTIONS is explicitly emptied before Node can start. The runtime records
the resolved image ID and requires that exact ID in later container inspection.
Image metadata is not binary provenance, signature, scan or compatibility proof.

## Container contract

Creation is stopped and one-shot, without automatic removal or image pulls.
The container has:

- UID/GID 1000, read-only root, private IPC/cgroup namespaces and init.
- No network, host PID/user namespace, ports, devices, extra capabilities or
  inherited host mounts; no-new-privileges and all capabilities dropped.
- 512 MiB memory and swap ceiling, half a CPU, 64 PIDs and 512 file descriptors.
- Bounded no-exec scratch tmpfs and no Docker log driver.
- Exactly one read-only bind mount: the private provider socket.
  No credential file, source workspace, tool volume, Docker socket or audit socket
  mount is permitted.
- Fixed Node entrypoint and engine bootstrap script, fixed environment, disabled
  healthcheck, open stdin without TTY, no restart policy and no automatic removal.

Full reported policy is checked immediately after create, before supervisor bind,
before immutable-ID start and after start. Identity, image, labels, exact mount
policy, environment, running state and resources must match. Inspect calls also
revalidate policy. Matching names/labels cannot substitute a different container
ID. An uncertain start consumes the attempt and cannot be retried.

The paired supervisor still journals intent before create, binds both identities
and verifies aggregate resources before starting either role. The tool bootstrap
receipt remains required before engine start. This engine starts its fixed
app-server process. Its one-shot stdio attachment connects to the callback driver;
production orchestration and an actual provider turn remain unverified.

## Protocol attachment

`EngineContainerRuntime.Attach` requires the original running immutable ID and
rechecks the full container policy before and after opening the stream. It uses
the same runner as creation and inspection. An attempted attachment consumes its
one-shot admission even if opening fails; there is no reconnect or uncertain-turn
replay. A policy change closes the newly opened stream before returning it.

The local runner revalidates the rootless daemon and invokes only `docker attach`
with signal proxying and detach keys disabled. It inherits only the Docker
selection/platform environment used by the tool runner, discards stderr, and
exposes protocol stdin/stdout. Closing cancels the attachment, closes pipes and
waits for its process to exit. The stdout pipe remains readable after process
exit so a final buffered frame is not lost. The callback driver retains its own
frame, stream, output, call and deadline limits; this attachment does not parse or
log provider output.

The trusted worker calls `workerCodexEngine.Turn` inside the supervisor's `Run`
callback, after pair startup. It checks the original session/scope, obtains the
journaled engine ID and passes its owned stream to `CodexTurn`. Failure cancels
the supervisor; the enclosing Run still performs fenced shutdown. Calling this
method alone does not replace Run/Shutdown. No public launch endpoint, provider
fallback or container-isolation switch is introduced.

## Failure and verification

The original job lifetime bounds every call even with a fresh caller context.
Both the original cancellation context and its deadline-bounded child are checked;
the interval before cancellation reaches a child cannot admit a new operation.
Tool workspace readiness and execution enforce the same original-lifetime rule.
Cancellation and runtime Close provide no container absence evidence. The outer
paired Run/Shutdown path must remove and observe both immutable role identities,
retain unresolved workspaces and keep account reservations until proof succeeds.
The runtime does not automatically discard source or credentials.

Tests use mocked image/container inspection and runner calls. They cover fresh
stopped creation, immutable start, repeated operation refusal, existing-container
refusal, image preflight injection/implicit volumes, scope mismatch, unsafe
mounts/environment/network/resources, changed policy before/after start, uncertain
start and cancelled-lifetime refusal. All database interactions remain mocked.
No live Docker image, provider account, gateway or rootless resource enforcement
has been verified. Private gateway staging, token refresh custody,
production orchestration and all remaining P0 acceptance gates
must be completed before managed development is enabled.

Attachment tests use synthetic local pipes, a test subprocess and mocked Docker
metadata. They cover duplex transfer, blocked-read cancellation, final-frame
preservation, idempotent close/process reaping, wrong identity, policy changes
before/after attachment and refusal to retry an uncertain connection. They do
not establish real Docker attach compatibility or subscription authentication.
