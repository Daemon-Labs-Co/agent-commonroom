# Mandatory agent container execution

[Documentation index](README.md)

## Current support

Managed Codex and Claude turns now use a mandatory rootless Docker worker on Linux. There is no enable/disable flag and no direct-process fallback. Direct macOS/Windows execution remains blocked. Approved SSH worker transport is implemented; see [connection sign-in](connection-sign-in.md). CommonRoom's UI and database can still be used without an execution worker.

The default worker uses private Docker volumes and chat-only provider policies. Operators can explicitly configure room-scoped [development workspaces](development-workspaces.md): a reviewed, checksummed Git bundle is mounted read-only and cloned into a writable private feature branch before the provider starts. No host working directory is mounted writable. Room directory-selection UI, task-specific worktrees, patch export and merge automation remain follow-up work.

## Set up a Linux worker

1. Install [rootless Docker](https://docs.docker.com/engine/security/rootless/) on the Linux account running CommonRoom. Select that local rootless context. The daemon must report rootless mode, seccomp, and support memory/swap, PID and CPU limits. Do not set `DOCKER_HOST` or `DOCKER_CONTEXT`; remote Docker endpoints are rejected.
2. Build and publish a trusted agent image from `docker/agent/Dockerfile`. Supply a reviewed Node base image by digest and explicit installed-provider versions:

   ```sh
   docker build -f docker/agent/Dockerfile \
     --build-arg NODE_IMAGE='node:24-bookworm-slim@sha256:REPLACE_WITH_REVIEWED_DIGEST' \
     --build-arg CODEX_VERSION='REPLACE_WITH_REVIEWED_VERSION' \
     --build-arg CLAUDE_VERSION='REPLACE_WITH_REVIEWED_VERSION' \
     -t YOUR_REGISTRY/commonroom-agent:reviewed .
   docker push YOUR_REGISTRY/commonroom-agent:reviewed
   ```

   These placeholders must be replaced; no provider versions or image digest have been verified by this change. Resolve the published digest, pull that exact image into the local rootless daemon, then configure `COMMONROOM_AGENT_IMAGE=YOUR_REGISTRY/commonroom-agent@sha256:ACTUAL_64_HEX_DIGEST`. Launches never pull automatically. Registry publication and installation are human/operator steps, not agent capabilities. The build context allowlist excludes application secrets and local data.
3. Export `COMMONROOM_AGENT_AUTH_DIR` to a private credential directory (default `.data/agent-auth`). Provide only `codex/auth.json` and/or `claude/.credentials.json` from the provider's authorized authentication setup. Each must be a regular JSON file, no larger than 64 KB, with owner-only permissions (`chmod 600`). Do not copy the entire provider home, SSH files or CommonRoom keys. Provider authentication may use keychains or different mechanisms; if the required file is unavailable, this worker cannot use that account yet. The current host-side Claude sign-in action does not automatically provision container credentials.
4. Restart CommonRoom with those environment values. Invite a **fresh** local Codex/Claude participant. Existing host-session bindings cannot automatically resume: they are explicitly blocked to avoid silently dropping context or crossing storage boundaries. Preserve old conversation history and import selected context into the new participant instead.

The worker must run on the same Linux machine as the selected Docker daemon. Unix socket mounts are not currently supported through Docker Desktop's VM or remote Docker contexts. Follow [Debian installation](deployment/debian.md) to run CommonRoom there and reach its browser UI through the documented SSH tunnel.

## Enforced launch policy

- No external container network (`--network=none`), no host networking, exposed ports, Docker socket or host home-directory mount.
- Non-root UID/GID 1000, read-only root filesystem, all capabilities dropped, no privilege escalation, Docker's default seccomp, and an init process.
- 1 CPU, 1 GB memory including swap, 128 processes and a three-minute turn deadline. Temporary storage is bounded. Persistent volume disk quotas are not implemented.
- Membership/room/connection/endpoint-specific home and work volumes, named from a hash of those identifiers. Provider auth is copied into a per-turn protected host temporary directory; only that file is mounted read-only. The bootstrap copies it into the agent's private home. A provider credential is intentionally accessible to that provider process; containers do not hide necessary credentials from the agent using them.
- At most 4 MB each of captured stdout/stderr. Docker log capture is disabled for these containers.
- A per-turn Unix-socket HTTPS CONNECT gateway supplies provider access. The socket is not the Docker control socket. The bootstrap bridges it to container loopback. Exact hosts only: Codex uses `api.openai.com`, `chatgpt.com`, `auth.openai.com`; Claude uses `api.anthropic.com`, `console.anthropic.com`, `claude.ai`, `claude.com`, `platform.claude.com`. Only port 443 is accepted; private/reserved DNS results are rejected, and a validated IP is dialed directly. Other downloads, external tools and dependency registries are blocked. Unsupported provider proxy behavior fails connectivity rather than restoring unrestricted networking.
- The gateway limits concurrent tunnels, time and bytes. TLS remains end-to-end. Context can still be sent to permitted providers, including through misuse of their APIs; this is an egress boundary, not content-loss prevention.

## Cancellation and recovery

After a turn ends or its context is cancelled, the gateway is closed and Docker receives a bounded forced container-removal request. Unlike killing the client alone, this targets the container's processes. If removal fails, the server prints the container name and reports that cleanup could not be confirmed. Independently check the daemon's containers carrying `commonroom.managed=true`. Daemon outages, worker crashes and create/cancel races remain recovery risks; startup reconciliation is not yet implemented. Do not infer confirmed termination from the cancellation request alone.

Named volumes persist after turns and can contain credentials, sessions and work. Protect the Docker storage and its backups. Volume lifecycle, export and secure deletion are not automated. Host-side human-authorized history import and provider setup remain trusted administrative operations, not isolated model turns.

## Bypass and verification limits

You **CAN** modify the source to restore direct host execution, **BUT** that exposes account files, credentials and network access, removes OS-enforced room confinement and weakens cancellation. The supported application provides no switch for this. Containers also depend on a trusted runtime, reviewed image and worker host; runtime vulnerabilities or a malicious administrator remain outside this containment guarantee.

Automated tests inspect launch policy, pinned-image requirements, unsupported endpoints, daemon prerequisites, egress denial, private addresses, output bounds and legacy-session blocking. All database interactions are mocked. No real provider is used. Live rootless-Docker confinement, allowed provider connectivity and descendant termination still require isolated deployment verification; Docker was unavailable on the development machine when this implementation was prepared. SEC-04 is mitigated for the new launch path, not declared fully closed.

## Remote worker requirements

**Required deployment:** CommonRoom may run on a Mac or another controller machine while all managed agent containers run on an external Linux worker, such as Debian. Container isolation is mandatory regardless of location. SSH is authenticated transport to the worker, not permission to invoke Codex or Claude directly on the remote host. Docker does not need to be installed on the controller.

**Current status:** constrained SSH worker transport and connection-specific sign-in are implemented. Direct SSH model launches remain blocked. Worker leases, restart reconciliation and a full registration lifecycle remain follow-up work. See [connection sign-in](connection-sign-in.md) for the current deployment procedure.

### Observed Debian prerequisites — 11 October 2026

Read-only checks through the existing `jerry` and `claude` SSH aliases found Linux and `/usr/bin/docker` for both accounts. Both configured Docker contexts pointed at `unix:///var/run/docker.sock`; both accounts received permission denied when contacting that daemon. Installing the Docker CLI alone does not establish worker readiness. No accounts, permissions, containers or provider sessions were changed by these checks.

Configure an appropriate rootless worker account and verify daemon access under the exact identity running the worker. Jerry and Claude can use separate workers/accounts, or a dedicated CommonRoom worker can manage distinct private provider sessions; provider credentials must be deliberately provisioned, never inferred or copied wholesale from another user's home. Merely adding the accounts to the system Docker group is not the recommended fix: access to a privileged Docker daemon broadens host authority and does not satisfy the rootless worker contract.

### Required boundaries

| Area | Requirement |
|---|---|
| Controller | Keep the web UI, authorization and orchestration separate from the execution host. Send authenticated, authorized requests; do not forward owner keys or database credentials to agents. |
| Transport | Start with SSH using a dedicated worker identity and verified host keys. No public Docker TCP endpoint, unauthenticated worker API or agent-accessible SSH credentials. |
| Worker entry point | Use a constrained, versioned worker protocol with validated operations and bounded payloads. Requests cannot supply arbitrary shell commands, Docker arguments, host mounts or environment variables. |
| Readiness | Verify Linux, reachable rootless Docker, seccomp, enforceable resource limits, reviewed digest-pinned image, provider credentials and gateway prerequisites before admitting a job. An unreachable/unready worker blocks execution. |
| Execution | The remote worker creates containers and enforces the same user, capability, filesystem, network and resource restrictions as the local worker. No direct-execution fallback. |
| Credentials and gateway | Provider credential files and Unix gateway sockets are created on the worker host and mounted locally there. They are not paths on the Mac mapped into a remote daemon. Keep provider credentials out of transport logs and audit payloads. |
| Storage | Keep private session/work volumes scoped to company, room, agent membership, connection and endpoint. Validate any future approved repository mounts on the worker; do not accept arbitrary controller-supplied host paths. |
| Job identity | Use durable job/attempt IDs, idempotent submission and worker ownership/leases. Reconnection must not silently duplicate work or resume a session belonging to another worker/account. |
| Results | Bound and correlate status/output to the authenticated worker and original attempt. Persist meaningful audit events without recording secrets. |
| Cancellation | Send cancellation to the worker, terminate the execution container and its gateway there, and return an acknowledgement with the observed outcome. Closing SSH alone does not prove container termination. |
| Failure and recovery | On connection loss show status as unknown, preserve job identity, and apply worker-side deadlines/lease expiry. Reconcile running containers after worker/controller restart; do not silently resubmit uncertain work. |

### Completion criteria

Remote support is complete only after mocked protocol tests cover identity/scope denial, invalid launch requests, readiness failures, duplicate submissions, disconnect/reconnect, output bounds and cancellation outcomes. Database interactions remain mocked; no test data may be inserted into PostgreSQL. A separate controlled deployment check must verify Linux/rootless-Docker containment, provider connectivity, denied filesystem/network access and descendant shutdown using harmless fixtures before relying on the worker for live development. No supported feature flag may disable isolation or restore direct execution.

### Provider credential location contract

`docker/agent/entrypoint.cjs` supplies a fixed provider launch environment for login and execution. Codex uses `CODEX_HOME=/home/agent/.codex` and file-based credential storage; Claude uses `CLAUDE_CONFIG_DIR=/home/agent/.claude`. A successful login must create a regular credential file at the worker's expected export path. The image explicitly installs the entrypoint with readable file permissions and a traversable parent directory for its non-root user. It also installs the system CA certificate bundle used by native provider clients, without disabling TLS verification. `persistWorkerCredential` validates the exported JSON and atomically installs it with mode 0600, preserving the old login when export validation fails.

Codex uses the existing loopback proxy environment with platform proxy discovery/fallback disabled. The mandatory Docker/network gateway boundary is unchanged. Regression tests mock provider processes and use temporary files; no database or live-provider calls are made. Rebuild and repin the image after changing this contract; a source-only controller update cannot repair an older worker image.

### Private credential export boundary

After provider login exits successfully, the worker reads the known credential file through a separate non-root container with networking disabled and only the login home volume mounted read-only. This avoids Docker file-copy failures against stopped login containers. The exporter rejects missing files, symlinks, directories, invalid JSON and files over 64 KB. Its output stays in bounded private worker memory and a mode-0600 temporary file; neither output stream is forwarded to diagnostics. Atomic installation preserves existing connection credentials on failure. Login and export containers and temporary volumes are removed afterward. This remains under the existing owner-only `connectionSignIn` feature flag (enabled by default); it introduces no isolation bypass or additional toggle.

Regression tests exercise the actual fixed exporter script with synthetic temporary files, its Docker isolation arguments, output limits, and atomic persistence. Automated tests do not access a database or launch Docker/provider agents. A successful browser login followed by status from a fresh worker process is still required to verify a deployment.

Isolated model turns keep standard input open (`docker run -i`) so the adapter can deliver its bounded prompt. They do not allocate a terminal, preserving machine-readable responses. Interactive human login separately uses a terminal. A saved credential and a successful live connection check are separate states.

## Native public provider address policy

The mandatory CONNECT gateway rejects special-use/reserved IPv4 destinations
and permits only ordinary native IPv6 global-unicast addresses in `2000::/3`,
excluding protocol assignments, transition mechanisms, documentation and retired
blocks. NAT64/translation, 6to4, Teredo and other special-use routes are denied
even if a platform classifies them as global unicast. IPv4-mapped IPv6 is checked
as IPv4. Exact provider hostname/443 checks, validation of every resolved answer
and dialing the validated address directly remain required. No optional flag or
uncontained network fallback exists.

This conservative policy may reject a provider on a NAT64-only network. Configure
ordinary native public IPv4/IPv6 connectivity; do not relax the denial to reach
private addresses. Future allocation changes need an explicit policy review.
Synthetic address tests perform no DNS queries or provider connections. Trusted
resolver/host routing compromise remains outside this address classifier; this
does not fix the separate provider credential/tool co-location issue.

Policy references: [IANA IPv6 special-purpose registry](https://www.iana.org/assignments/iana-ipv6-special-registry),
[IANA IPv6 address space](https://www.iana.org/assignments/ipv6-address-space), and
[IANA IPv4 special-purpose registry](https://www.iana.org/assignments/iana-ipv4-special-registry).

## Gateway shutdown and pre-tunnel limits

Each provider gateway owns a cancellable context derived from its turn. Closing
the gateway cancels in-flight DNS resolution and dialing, closes established
tunnels and stops the listener. Requests reaching a closed gateway are refused
before resolution. The pre-tunnel HTTP server bounds headers to 8 KiB with a
five-second header timeout and fifteen-second read/write/idle timeouts. Hijacked
provider tunnels retain their separate three-minute deadline and byte limits.
These mandatory limits have no product toggle and do not relax hostname/IP policy.

Synthetic tests use a temporary Unix socket and injected blocking resolver; they
verify prompt cancellation and refusal of new work without any DNS/provider/DB
access. Cooperative resolver/dial cancellation is required. This narrows leftover
network work during shutdown; it does not prove physical container termination
or fix readable provider credentials in development tools. Worker shutdown proof
and integrity recovery requirements remain unchanged.

### Development snapshot substitution protection

Development launches mount a private per-turn checksum-verified bundle copy.
They do not mount the original operator source path after a prior hash check.
Descriptor checks and bounded reads also reject special files and growing-input
overreads. See [snapshot staging and capacity requirements](development-workspaces.md#per-turn-verified-snapshot-staging).
These mandatory checks have no isolation bypass; development credential/tool
co-location is still an open separate issue.

### Credential staging permissions

Per-turn credential staging requires exclusive file creation in a worker-owned
0700 directory and explicitly normalizes file mode 0444 for the mapped container
UID. Existing/aliased targets are refused. See [staging contract and limits](private-file-security.md#exclusive-per-turn-credential-staging).
The read-only bind does not prevent development tools from reading credentials;
credential/tool separation remains open.

### Gateway connection admission

Each provider gateway accepts at most **16 live client connections**, including
clients that have not sent HTTP headers yet. Its existing **8 CONNECT slots**
continue to bound DNS/dial/tunnel work. Excess accepted sockets are immediately
closed before HTTP parsing or DNS; callers may see a connection reset/EOF rather
than an HTTP error. A slot remains held for the full connection/tunnel lifetime
and is released once when cleanup closes it. Header/read/write/idle deadlines
and shutdown cancellation remain independently enforced.

This is a required free/core availability safeguard, with no optional feature
flag, paid entitlement or supported limit-disable option. It takes effect when
the updated worker binary is deployed; this source change does not deploy it.
Do not loosen it to work around agent loops. Investigate repeated reconnects and
provider/client behavior instead. This bounds live HTTP connection state, not
socket backlog, accept-loop CPU, bandwidth, cross-job quotas or provider spend.

Technical contract: `gatewayListener.Accept` admits into a fixed buffered slot
channel before handing a client to `net/http`, closing excess clients. Its
`gatewayClient.Close` uses `sync.Once` to release admission after underlying
connection cleanup, including concurrent/double closes. `newProviderGateway`
uses this wrapper for both its listener record and HTTP serving. A source fork
CAN remove the cap, BUT exposes the worker to untrusted pre-header connections.
Synthetic pipe-listener and ephemeral local Unix-socket tests cover saturation,
release, concurrent close, pre-header admission and shutdown without contacting
DNS/providers, reading credentials or using a database. Pinned-provider/runtime
integration and volumetric protection remain unverified.
