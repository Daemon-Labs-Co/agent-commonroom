# Provider/tool credential separation

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


Development tools must never receive CommonRoom's provider credentials. The
supported design separates a credential-bearing provider engine from a
credential-free repository/tool container, joined by an authenticated audited
broker. This is mandatory free/core security, with no disable switch.

## Current implementation and admission

The internal Codex path now assembles the fresh pair, source bootstrap, scoped
audit transport, private credential stream, provider gateway and callback driver.
`workerPairSupervisor.RunCodex` owns this lifecycle after admission. Synthetic
protocol peers and mocked Docker metadata exercise successful callbacks and
failure cleanup. They do not prove real images, subscription authentication,
rootless enforcement or production deployment.

Public production dispatch remains gated. `requireCredentialSeparatedDevelopment`
refuses the legacy native-development path before credential access or Docker
preparation. Operator opt-in cannot bypass that refusal. Existing chat and trusted
sign-in/history administration are separate policies. Do not read the presence
of internal components as permission to enable development workloads.

Required remaining gates include reviewed images and live acceptance, production
entry-point wiring/recovery, live refreshed-token acceptance and other required
provider adapters. Commands, patches, trusted artifacts and approved source
volume disposal are unfinished. Read/write callbacks are the implemented tool
surface. See [P0 acceptance](p0-acceptance-tests.md).

## Trust boundaries

| Component | Holds | Must not receive |
| --- | --- | --- |
| Provider engine | Connection-scoped subscription credentials and one ephemeral provider session | Repository/tool volume, host home, Docker/audit socket or model-accessible native shell/file/plugin/hook path |
| Tool container | Fresh approved source snapshot and private writable workspace | Provider credentials/home, OAuth environment, controller keys, host mounts or direct networking |
| Trusted worker | Docker supervision, private credential/gateway custody and scoped broker authority | Model-supplied launch options, image choices, host paths or container IDs |
| Controller | Human authorization, account reservations and durable audit acknowledgements | Unauthenticated worker claims as authority |

Both roles use non-root execution, rootless Docker, read-only root filesystems,
capability drop and no-new-privileges. Only the trusted worker talks to Docker.
The engine has a single owned provider-socket mount; credentials arrive through
bounded private stdin and are created inside its private home. The tool role has
no engine volume or provider socket. Neither role receives host credentials or
the application database key. See [engine runtime](engine-container-runtime.md)
and [tool image recipe](../../../docker/tools/Dockerfile).

This protects credentials supplied by CommonRoom from development tools. It does
not prove that prompts, imported history or repository files contain no secrets,
nor protect against a compromised trusted client, worker account, daemon or host
kernel. Review source/artifact disclosure separately.

## Broker authority and request contract

The supervisor issues a random per-job capability tied to the protocol version,
job, room, participant, connection, provider, repository digest and original
deadline. Its hash is retained; the token is not an owner session, audit value or
provider key and must never enter chat, diagnostics or model arguments.

`internal/toolbroker` validates exact scope, declared operations, expiry and
revocation. It consumes each admitted call ID once, caps authority at 100 calls,
and shares one active slot across dispatchers using that authority. Busy calls
are refused rather than queued. Original cancellation and deadline remain binding
when a caller supplies a fresh context; heartbeats cannot extend authority.

Strict JSON admission caps the wire object at 64 KiB and rejects unknown/duplicate
fields, incorrect case, nulls where forbidden, trailing values and excess nesting.
Typed calls also bound aggregate strings before copying or hashing: 64 KiB
content, 4096-byte paths, 8192-byte arguments and 128 command arguments. These are
validation ceilings, not implemented command support. Encoding validation rejects
malformed UTF-8 and unpaired escaped surrogates before lossy JSON replacement.

The executor independently enforces the copied operation policy and current
container/volume policy. A valid capability or reservation alone is insufficient:
authenticated transport, controller authorization and durable audit still apply.
Neither tool names nor model-provided scope can grant authority.

## Audit ordering and failures

Before a helper action, `Auditor.RecordIntent` must acknowledge durable controller
persistence. The intent binds trusted scope, call ID, operation and payload digest
without storing the capability. After execution, the outcome binds status, output
length and digest. Return output only after outcome acknowledgement and final
cancellation checks. See [scoped audit callbacks](tool-audit-callbacks.md) and
[private audit transport](worker-audit-transport.md).

Intent failure executes nothing and revokes authority. Missing outcome evidence,
cancellation, oversized output, malformed receipts or uncertain process results
withhold output, fence the job and require reconciliation. Never replay an
ambiguous side effect automatically. An acknowledgement is a durability contract;
returning success from a no-op auditor violates it. Local chain integrity does not
replace independent rollback anchors.

## Workspace file operations

The fixed Linux `cmd/tool-helper` accepts one strict read/write action, no command
arguments and no caller-selected root. It opens `/workspace/repository` through a
trusted rooted descriptor. The launcher supplies a fixed executable, working
directory and credential-free environment; no shell fallback exists.

Reads are capped at 8 KiB and require owned single-link regular files without
unsafe write permissions. Traversal, symlinks, FIFOs, devices, hard links and unsafe
roots are refused. Writes cap content at 64 KiB, require existing parents, create
an exclusive temporary inode, sync it and atomically rename through the pinned
parent. Existing files are not truncated in place. New files use 0644; existing
ordinary permission bits are preserved without set-ID bits. ACL/xattr migration,
directory creation, range reads, patching and commands are not provided by this
file helper. Separately admitted [terminal Go commands](isolated-commands.md)
use trusted worker output capture and consume the inspected tool container.

Post-rename or parent-sync uncertainty returns an unconfirmed write, never success
or permission to retry. Namespace checks do not lock out concurrent writers;
rooted operations alone do not exclude bind mounts or special filesystems. Trusted
image/mount/device policy and whole-container supervision remain essential. Disk
I/O can be non-cooperative even when contexts are canceled.

### Helper result admission before success

A zero exit code is insufficient. The executor requires a bounded strict receipt
whose operation, byte count and canonical SHA-256 match the admitted action.
Reads return canonical base64 with at most 8 KiB of decoded bytes; writes return
only matching count/digest metadata. The complete result is bounded at 16 KiB.
Malformed receipts are withheld and fence execution. A compromised helper can
forge consistent receipts: these are consistency checks, not filesystem
attestation or independently observed test results.

## Provider adapter gates

The reviewed Codex driver pins the `rust-v0.162.1` app-server protocol, uses one
fresh read-only thread and exposes only `commonroom_tool` callbacks: read/write
by default, or the separately granted opt-in terminal Go command profiles.
It refuses native execution requests, mismatched identifiers and unknown protocol
outcomes. The engine disables host discovery, hooks, plugins and browser paths
through its fixed reviewed bootstrap. See [callback driver](codex-callback-driver.md),
[bootstrap](codex-engine-bootstrap.md) and [engine build](../../../docker/engine/README.md).

Claude, Gemini, Copilot and other required isolated adapters need their own pinned
interface and effective-tool verification. A saved provider/transport label does
not demonstrate an implemented bridge. Retain CLI subscription accounts; never
silently select API billing or another provider. Unsupported adapters remain
unavailable with an explicit reason. Disabling a command tool is not proof that
every native filesystem path is disabled.

## Lifecycle, cancellation and retention

Version-4 pairs bind a fresh plain local workspace volume and both immutable role
identities to journaled intent. Version-3 pair records are recovery-only; legacy
chat uses version 2. Names, labels or expired deadlines cannot authorize adoption
or restart of a used container/workspace. See [worker pairs](worker-pairs.md),
[workspace volumes](workspace-volumes.md) and [bootstrap](workspace-bootstrap.md).

Stop revokes future calls, closes transport and cancels active work. Fenced
shutdown waits for creator/turn ownership, removes both recorded roles and
verifies exact-ID absence. Daemon/SSH errors never count as absence. Source
workspaces are retained by default; ordinary shutdown and HTTP recovery do not
supply disposal approval. Uncertain evidence remains unresolved and reserves
capacity. See [paired shutdown](worker-pair-shutdown.md).

Credential leases serialize use with sign-in/sign-out. Cancellation retains the
lease until both roles settle; journal checks block reuse after a process crash.
The provider gateway checks full scope, original lifetime and socket identity,
and removes only its own socket/empty directory. Unexpected files or replacements
remain evidence, not cleanup targets. Private final-credential capture and fenced
persistence are implemented internally; [live custody acceptance](engine-credential-custody.md)
and complete production recovery remain outstanding.

## Aggregate pair resources

Each role is capped at 512 MiB memory including swap, 0.5 CPU and 64 processes;
the pair stays within 1 GiB, one CPU and 128 processes. Inspection refuses missing,
unlimited, ambiguous or oversized reported limits and shared identities. Four
managed physical containers per account are reserved before creation, so a pair
consumes two slots. See [container quotas](container-quotas.md).

Reported metadata does not prove kernel enforcement or reviewed binary contents.
Fresh launch, immutable IDs, full policy inspection and live rootless acceptance
must agree before claiming production isolation. These limits and credential
audit boundaries have no feature flag or paid exemption.

## Source-level bypass risk and verification

A source fork CAN replace the broker, runner or rooted workspace with host
execution, BUT then tools can reach host files, credentials and networks and
invalidate the supported isolation/audit assumptions. There is no supported
host/SSH direct-launch fallback. See the [mandatory architecture contract](architecture.md#mandatory-container-isolation-design-contract).

Automated suites use synthetic records, private temporary files, local protocol
peers and mocked DB/Docker interactions. Separately authorized live acceptance
must use dedicated synthetic accounts/canaries and record image/binary provenance,
effective tools, malicious read/shell/proc/symlink refusal, resource enforcement
and both-role termination. No existing user credentials or provider workload is
an automated fixture. A passing mock, image recipe or configuration flag is not
security closure.
