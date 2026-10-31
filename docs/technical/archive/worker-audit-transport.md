# Private worker audit transport

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


Managed tools must receive durable controller acknowledgement before execution
and before results are released. The transport uses a dedicated SSH process,
separate from provider output, diagnostics, prompts and model callback pipes.
This is mandatory security infrastructure without a bypass flag.

## Current status

Internal transport primitives, supervisor session attachment and the constrained
`audit-link-v1` worker action are implemented. Internal Codex execution uses the
audited session and fences both containers on failure. Public dispatch still must
coordinate listener creation and the dedicated controller connection; it remains
gated.
Synthetic tests exercise real local Unix sockets, framed audit exchanges,
job-scope checks, cancellation and a harmless dedicated child-process substitute.
No live SSH, provider account, Docker workload or database has been exercised.

## Controller contract

`serveRemoteToolAudit` accepts only `COMMONROOM_REMOTE_WORKER=true` and the
configured `COMMONROOM_REMOTE_WORKER_ALIAS`, matching the connection in the
original managed-job registration. Existing initiating-human, company,
participant, repository, deadline and permission checks remain mandatory and are
repeated by the auditor for each phase. Remote transport defaults disabled;
configuration is an operator responsibility and requires controller restart.
Disabling transport prevents new links without deleting history or authorizing
alternate host execution.

The separate SSH process uses batch authentication and strict existing host-key
verification. It invokes only `commonroom-worker audit-link-v1` with a bounded,
canonical base64-encoded nonsecret scope. No provider tokens, commands, prompts,
source contents or capabilities are in that argument. Stderr is discarded rather
than copied into audit or model output. The link owns and closes both pipes,
cancels and reaps its process; cancellation is not container shutdown evidence.

## Supervisor and worker contract

The admitted version-4 supervisor calls `OpenAuditListener` with its exact scope
while holding its creator lease. The private jobs directory must deny group and
other access, and cannot itself be a symlink. The socket is mode 0600 and its
parent is never mounted into either model or tool container. Existing socket
paths are refused, never adopted or removed speculatively.

Before binding, a private one-shot marker is exclusively created and synced,
including the parent directory. The marker is retained after socket closure,
failed binding or a crash. An uncertain link cannot be reopened for that job.
Operators must reconcile container and workspace state and create a new job;
never delete a marker to resume an uncertain action.

The SSH worker action checks canonical scope encoding, exact durable version-4
record binding, a live job lease, private directory and bounded handshake. The
supervisor accepts exactly one connection and compares a digest of its complete
scope. The digest is a binding check, not a secret or authentication mechanism:
SSH authenticates the approved worker account and filesystem isolation excludes
other users. The worker account, same-UID processes, host administrator and Docker
daemon remain trusted. Never expose the socket, forward it publicly or give
containers access to the supervisor account.

The original supervisor lifetime and job deadline continue to bind an accepted
link even if a caller supplies a new context. Failed or wrong handshakes consume
the listener. Cancellation closes sockets and owned SSH pipes and joins relay
copies. EOF, uncertainty and framing failures fail closed without replay.
Framed nonce, ordering, digest acknowledgements and bounded-message enforcement
remain in `internal/toolbroker`; transport does not replace those checks.

## Required production integration and acceptance

Public dispatch must order listener creation before the controller's attach,
keep the original supervisor lease throughout execution, and start no tools until
registration and authenticated attachment are established. Failure to attach,
expiry, cancellation, controller loss and pending-outcome ambiguity must fence
both role containers and retain unresolved reservations until absence is proven.
Do not recover by mixing audit frames into model stdout or self-acknowledging.

Live acceptance still needs approved SSH account/host-key configuration,
controller loss, SSH termination, worker crash, strict registration mismatch,
paired container fencing and recovery using synthetic repositories and dedicated
provider accounts. Mocked persistence and local sockets do not establish live
provider or remote-host compatibility.

## Session attachment and disconnect supervision

`OpenAuditListener` now records ownership on its admitted supervisor.
`AcceptToolSession` accepts only that exact listener and the original live
version-4 scope. A same-scope socket created elsewhere is insufficient. It waits
for the private transport handshake outside the supervisor mutex so cancellation
can interrupt attachment, then builds the real shared tool session with the
approved runner/image/root policy. The capability is returned privately once;
it is never written to a worker record, stdout or an audit event.

A single read pump watches the accepted link. It forwards bounded chunks through
an unbuffered pipe to the audit client; the audit client remains the sole frame
parser. This avoids competing reads for acknowledgements and detects EOF even
while no tool call is running. There is no unbounded prefetch or reconnect.
Transport EOF, parent cancellation, failed session construction or explicit link
closure cancels the supervisor's original lifetime. Startup/runtime calls using
fresh contexts cannot revive it. Pipe and socket closure joins the read pump.

Cancellation revokes tool access and interrupts in-flight launch operations. It
does not itself stop already running Docker containers, release the creator
lease, prove shutdown, release account reservations or authorize workspace
removal. The internal `RunCodex` operation performs fenced shutdown of both roles after
the session ends. Public dispatch must retain that behavior. Production admission
remains blocked pending dispatch integration, reviewed images and live acceptance.

Additional tests exercise framed intent/outcome acknowledgement without frame
loss, idle EOF, blocked forwarding cancellation, foreign-listener refusal, invalid
image construction and real ToolContainerRuntime startup cancellation using a
mocked blocked Docker operation. Disconnect preserves the pending workspace
journal and held creator lease and prevents engine creation. These tests establish
local lifecycle behavior; live SSH and container shutdown are still unverified.

See [paired turn shutdown](worker-pair-shutdown.md) for unconditional
reconciliation, active-turn lease fencing and workspace retention.

## Idle-stop acknowledgments on the same link

The [terminal idle-stop extension](tool-audit-callbacks.md#terminal-idle-stop-extension)
uses the existing one-shot `audit-link-v1` SSH attachment and private socket. It
adds no endpoint, forwarded socket, container mount, SSH permission or reconnect
path. Model stdout and callback tools cannot produce this control frame.

Tests now combine the real Unix-socket handshake/held creator lease, owned pipe
relay, watched session, controller journal and mocked Docker fencing. They cover
committed acknowledgment, audit failure, lost acknowledgment, wrong policy
revision and unsupported controller. They use no actual SSH session, provider,
Docker daemon or database; current-host socket tests are not live remote acceptance.
