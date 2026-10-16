# Mandatory agent container execution

[Documentation index](README.md) · [SSH worker setup](agent-ssh.md)

## Current support

Managed turns require a Linux worker with local rootless Docker. CommonRoom can
run on a Mac or a private Debian controller; the execution worker remains Linux.
There is no direct host/SSH provider launch, Docker Desktop execution path,
uncontained fallback or isolation off switch.

| Capability | Current status |
| --- | --- |
| Isolated Codex/Claude chat and connection testing | Implemented through the approved SSH worker. |
| Provider sign-in | Isolated interactive login and connection-scoped credential storage. |
| Repository development | Public dispatch blocked; replacement separated Codex pipeline exists internally. |
| Other provider execution | Required adapters remain unfinished or unverified; icons are not support claims. |
| Hard storage quota | Required, not yet implemented. |

See [development workspaces](development-workspaces.md) before attempting repository
work. Historical single-container development checks do not validate the replacement.

## Set up a Linux worker

1. Provision a dedicated Linux account and a local [rootless Docker daemon](https://docs.docker.com/engine/security/rootless/).
   Readiness requires seccomp and enforceable memory/swap, CPU and PID limits.
   Remote Docker contexts and environment overrides are rejected.
2. Build the reviewed chat/login image from `docker/agent/Dockerfile`, supplying a
   digest-pinned Node base and explicit Codex/Claude versions. Review dependencies
   and licenses. Pull the final immutable image into the worker daemon; launches
   do not pull automatically. An example build is:

   ```sh
   docker build -f docker/agent/Dockerfile \
     --build-arg NODE_IMAGE='node:24-bookworm-slim@sha256:REVIEWED_DIGEST' \
     --build-arg CODEX_VERSION='REVIEWED_VERSION' \
     --build-arg CLAUDE_VERSION='REVIEWED_VERSION' \
     -t commonroom-agent:reviewed .
   docker image inspect commonroom-agent:reviewed --format '{{.Id}}'
   ```

   Replace every placeholder. A local build tag is not an acceptable worker image
   setting: configure its immutable ID or registry digest. No release image is
   established by this example.
3. Follow [worker configuration and sign-in](connection-sign-in.md#worker-configuration)
   to install the matching binary, wrapper and private configuration.
4. Approve the alias on the controller, register a connection, sign in, then run
   [Test connection](connection-testing.md). Invite fresh participants afterward.

The daemon, volumes and Unix provider sockets live on the worker. Controller paths
are not remote Docker mounts. Existing desktop conversations are not attached by
registering a connection; approved history import is a separate operation.

## Enforced launch policy

Chat/login execution uses non-root UID/GID 1000, read-only root, dropped
capabilities, no privilege escalation, default seccomp, an init process and
`--network=none`. It mounts neither the Docker socket nor the host home/checkout.
A turn is bounded to three minutes, 1 CPU, 1 GiB memory including swap and 128 PIDs;
captured stdout/stderr are each bounded to 4 MiB and Docker logging is disabled.

The replacement development pair divides its aggregate budget between roles and
uses fresh workspace/container identities. See [paired execution](worker-pairs.md)
and [container quotas](container-quotas.md). Four physical managed containers is
the current account/worker ceiling; unresolved reservations remain counted.
Persistent storage is not yet protected by the required aggregate 5 GB quota.

Chat session volumes can contain provider credentials and session data. The provider
necessarily receives its own credential; this is why model-controlled development
tools must run separately. The internal replacement engine receives credentials
through private stdin, not a host credential-file mount. See
[engine bootstrap](codex-engine-bootstrap.md) and [private files](private-file-security.md).

## Provider-only network gateway

Containers reach approved provider HTTPS endpoints through a per-turn Unix-socket
CONNECT gateway bridged to loopback. It is not the Docker control socket.

| Provider | Exact permitted hosts, port 443 only |
| --- | --- |
| Codex | `api.openai.com`, `chatgpt.com`, `auth.openai.com` |
| Claude | `api.anthropic.com`, `console.anthropic.com`, `claude.ai`, `claude.com`, `platform.claude.com` |

Every DNS answer is checked and a validated address is dialed directly. Private and
special-use addresses are refused. Arbitrary downloads and package registries are
blocked. TLS remains end-to-end. Permitted provider APIs can still receive context;
this boundary does not provide content-loss prevention.

## Native public provider address policy

IPv4 special-use/reserved destinations are denied. IPv6 permits ordinary native
global-unicast addresses in `2000::/3` while excluding special assignments,
translation/transition, documentation and retired blocks. IPv4-mapped IPv6 is
checked as IPv4. NAT64-only connectivity may fail; use ordinary public connectivity
rather than weakening address checks. Resolver/host-routing compromise remains a
trusted-host risk. Allocation changes require policy review.

## Gateway shutdown and pre-tunnel limits

Closing the gateway cancels DNS/dial operations, closes tunnels and stops its
listener. New work is refused after close. Cooperative resolver/dial cancellation
is required; gateway closure alone is not physical container shutdown proof.

| Limit | Value |
| --- | --- |
| Live client sockets, including pre-header clients | 16 |
| Concurrent CONNECT slots | 8 |
| Header size / timeout | 8 KiB / 5 seconds |
| HTTP read, write and idle timeouts | 15 seconds |
| Established tunnel deadline | 3 minutes, with separate byte limits |

### Gateway connection admission

Excess sockets close before HTTP parsing or DNS and may report reset/EOF. Admission
slots are held through connection cleanup and released once. These mandatory limits
have no product toggle. They do not establish cross-job bandwidth/spend quotas or
volumetric denial-of-service protection. See [gateway ownership](engine-container-runtime.md)
for the internal engine's job-bound socket checks.

## Cancellation and recovery

Managed jobs record intent before dispatch and keep private worker records and
leases. Cancellation latches the job, fences creation, removes only matching
container identities and observes absence. An unreachable daemon or lost SSH
connection is uncertainty, never proof of termination. Startup and owner-driven
reconciliation inspect unresolved jobs; they do not replay them.

Use [worker diagnostics](worker-jobs.md) for exact recovery steps. Named session
volumes persist, and normal pair shutdown retains source. Do not delete worker
records, locks or volumes to make readiness pass. Explicit approved workspace
disposal remains unfinished.

## Remote worker requirements

SSH is authenticated transport to the versioned worker protocol, with a dedicated
identity, verified host keys and bounded payloads. It does not grant arbitrary
shell/Docker argument execution. The controller binds authorization, connection,
room, participant, job and deadline before dispatch. The worker validates those
bindings and independently enforces isolation and durable lifecycle limits.

No public Docker TCP endpoint, agent-accessible SSH key or controller database/
owner credential belongs in a container. Run one controller per event store;
local startup locking does not support a distributed controller deployment.

### Provider credential location contract

The chat/login image fixes Codex file storage at `/home/agent/.codex/auth.json`
and Claude at `/home/agent/.claude/.credentials.json`. Login and export validate
bounded regular files, then atomically publish private connection credentials.
The separated Codex engine has a different private-home contract; do not mix its
image or bootstrap protocol with the chat/login image.

### Private credential export boundary

A separate network-disabled exporter reads only the known login credential file
from a read-only volume. Output is bounded and never diagnostic content. Publication
must sync before success. See [sign-in storage](connection-sign-in.md#security-and-storage).

### Development snapshot substitution protection

[Snapshot preparation](development-workspaces.md#per-turn-verified-snapshot-staging)
and [workspace bootstrap](workspace-bootstrap.md) define checked snapshot handling.
Neither permits the retired co-located provider/tool development path.

### Credential staging permissions

Legacy chat staging uses an exclusive file in a worker-owned private directory.
The replacement engine instead consumes a private stdin frame. See
[private credential staging](private-file-security.md#exclusive-per-turn-credential-staging).

## Bypass and verification limits

Source changes **CAN** remove these boundaries, **BUT** expose host files,
credentials and network authority and invalidate the supported contract. No
supported feature flag may do so. The host, daemon, deployment account and reviewed
images remain trusted; containers do not eliminate runtime vulnerabilities.

Tests use mocked Docker/database boundaries and synthetic temporary files/socket
peers. Live rootless containment, image compatibility, provider access and complete
descendant termination require separate deployment evidence. Synthetic success is
not production acceptance. Follow the [P0 acceptance plan](p0-acceptance-tests.md).
