# Reviewed engine-container runtime

The internal Codex operation connects fresh container creation, private credential
provisioning, provider networking and callback execution. Synthetic tests exercise
the combined path. **Public production dispatch remains gated**; reviewed release
images and live refresh/transport acceptance are still required.

This is mandatory isolation infrastructure. There is no off switch, default image,
paid exemption or uncontained fallback.

## Trusted inputs

Trusted worker code supplies the registered scope, reviewed immutable image,
connection credential lease, owned gateway and approved local rootless Docker
runner. Models and public requests cannot supply paths, runners or Docker options.
The engine and audited session must match the complete scope: job, participant,
connection, provider, room, repository and normalized deadline.

The worker account, host administrator and daemon remain trusted. Reported Docker
metadata is not image provenance or protection against a hostile host. Source
changes that substitute a host runner, Docker socket or unreviewed image violate
the mandatory boundary.

### Private provider gateway

`newWorkerEngineGateway` requires the original worker job context and full Codex
scope. It creates a fresh random directory under an existing worker-owned 0700
staging root, then creates its socket there. The trusted root cannot be a symlink
or public directory. Path-length errors fail without a shared-directory fallback.

The socket is 0666 so the container's subordinate rootless UID can connect through
its sole bind mount. Its unmounted host parent remains 0700. Exact provider-host,
public-address, connection and request limits still apply.

Construction and engine operations recheck scope, original lifetime, directory
identity and socket inode/type/owner. Cancellation closes networking. Cleanup
unlinks only the recorded socket and removes only its empty owned directory;
replacements, permission changes and unexpected files are preserved and reported
as unconfirmed cleanup. Gateway closure is not container absence proof.

### Connection credential custody

`acquireWorkerEngineCredential` holds the same nonblocking connection/provider
lock as sign-in and sign-out. While holding it, the worker checks bounded durable
job records: the current full-scope job must have no role creation intent, and
prior jobs on that connection/provider must have settled roles. Expiry alone is
insufficient. Invalid or missing records refuse acquisition. This prevents a lost
process lock after a crash from authorizing credential reuse beside an old engine.

Credentials come from a bounded private single-link regular file opened with
no-follow, descriptor-bound checks. Only the reviewed subscription schema is
accepted: exact fields, no duplicates, `auth_mode: chatgpt`, nonempty access token
and no non-null API key. Unknown fields require compatibility review; failures
return generic errors without publishing credential bytes.

The lease provisions one authentication frame before callback initialization:
`COMMONROOM-AUTH-V1\n`, a four-byte unsigned big-endian length and 2–65536 JSON
bytes. Partial writes consume the attempt and are never replayed. The bootstrap
reads exactly that frame and creates an exclusive UID-1000 0600 file in private
tmpfs. **No host credential file is mounted.** See the
[bootstrap contract](codex-engine-bootstrap.md) for process configuration.

Cancellation retains the connection lock until matching durable proof settles
both roles as absent or never created. Retained source alone does not retain the
credential lease once both roles settle. After valid turn completion, a fixed private helper requests the final credential
from the inspected engine bootstrap. The worker stages strict subscription bytes,
then reads matching durable absence evidence before atomic persistence and marker
clearance. No controller, model or audit sink receives these bytes. Release clears
owned buffers and preserves persistence errors across retries; it does not
guarantee erasure of every runtime copy. See [custody](engine-credential-custody.md).
Process recovery still relies on the journal and uncertainty marker.

### Owned worker operation

`workerPairSupervisor.RunCodex` accepts an already admitted job, its original
authenticated ToolSession, private capability, approved source-bundle reader and
trusted configuration. It is not a public endpoint or a substitute for controller
authentication, permissions, feature flags or quota admission.

Validation refusal leaves resources with the caller. After acceptance, the
one-shot operation owns the session and bundle and performs:

1. Credential lease acquisition and private gateway creation.
2. Assembly of the original tool runtime and reviewed engine.
3. Fresh pair startup, workspace bootstrap and one callback turn.
4. Fenced shutdown on success, failure, cancellation or panic.

The original supervisor context bounds admission; the request context also bounds
execution. Gateway, bundle-close, turn and shutdown errors remain visible. Repeated
bundle closure preserves the first error. Unresolved credential leases stay with
the supervisor until later shutdown obtains matching role proof.

## Before creation

Creation is one-shot and requires successful fixed-name inventory showing no
existing container. Used containers are never adopted or resumed. A failure or
lost reply requires fenced recovery and a new job, rather than replay.

Image inspection must report an immutable image ID and expected user, workdir and
entrypoint, with no inherited command, implicit volumes, OnBuild hooks or unknown
environment keys. Known values are overridden by fixed create policy;
`LD_PRELOAD` is refused and `NODE_OPTIONS` is emptied. Later inspection must report
the recorded image ID. These checks do not attest signatures, scans or binary
compatibility.

## Container contract

The runtime creates a stopped container without image pulls or automatic removal:

- UID/GID 1000, read-only root, private IPC/cgroup namespaces and init.
- No network, host PID/user namespace, ports, devices or extra capabilities;
  all capabilities dropped and no-new-privileges enabled.
- 512 MiB memory/swap ceiling, half a CPU, 64 PIDs and 512 file descriptors.
- Bounded no-exec scratch tmpfs and no Docker log driver.
- Exactly one read-only bind: the owned provider socket. No credentials,
  source/tool volume, Docker socket or audit socket mount.
- Fixed Node entrypoint/environment, disabled healthcheck, open stdin without
  TTY, no restart policy and no automatic removal.

Full identity, image, labels, mounts, environment, state and resource policy are
checked after creation, before binding/start, after start and during inspection.
Matching names cannot substitute a different immutable ID. An uncertain start
consumes the attempt.

The supervisor journals both IDs and verifies aggregate resources before either
role starts. It starts the idle tool and verifies bootstrap before engine startup.
See [pair assembly](worker-pair-assembly.md) for ordering.

## Protocol attachment

`EngineContainerRuntime.Attach` requires the original running immutable ID and
checks full policy before and after opening the stream. Attachment is one-shot,
including failed opens; there is no reconnect or uncertain-turn replay.

The local runner revalidates the rootless daemon and runs only `docker attach`,
with signal proxying and detach keys disabled. It inherits the restricted Docker
selection/platform environment, discards stderr and exposes protocol stdin/stdout.
Close cancels, closes pipes and reaps the process. Buffered stdout remains readable
after process exit so the final frame is not lost.

Inside the supervisor's Run callback, `workerCodexEngine.Turn` checks the original
session/scope, obtains the journaled ID, provisions credentials and passes the
stream to `CodexTurn`. The [callback driver](codex-callback-driver.md) enforces
protocol limits. Failure cancels the supervisor; outer Run/Shutdown still owns
cleanup. Calling Turn alone does not replace that lifecycle.

## Failure and verification

Every call checks both the original cancellation context and its deadline-bounded
child. A fresh caller context cannot exploit cancellation propagation delay to
revive engine or tool access. Close and cancellation do not prove container
absence: fenced shutdown must inspect both immutable roles, retain unresolved
workspaces and preserve reservations until durable proof succeeds.

Automated coverage uses mocked Docker, synthetic credentials, local protocol peers
and temporary files. It covers image/policy drift, unsafe mounts/resources,
one-shot operations, original-lifetime cancellation, private credential framing,
gateway replacement, stream closure and process reaping. Combined RunCodex tests
exercise bootstrap, callbacks, framed audit intent/outcome and independent role
cleanup, including forbidden callbacks, audit failure and engine disconnects.

A successful model turn still returns a lifecycle error while its workspace is
retained; tests do not fabricate disposal to force success. Mocked metadata does
not prove real resource enforcement, image contents, OAuth or Docker attachment.
Live rootless/SSH/provider checks, public dispatch, reviewed release images,
live refreshed-token custody, storage quotas, commands/artifacts and approved
disposal remain P0 requirements.
