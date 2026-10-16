# Security hardening and operational limits

These controls protect the supported private single-controller deployment. They
do not establish public-hosting readiness, tenant isolation or SOC 2 compliance.
See the [risk register](../SECURITY_REVIEW.md) for remaining risks and
[release readiness](release-readiness.md) for unfinished delivery gates.

## Identity and request admission

Authenticated API actors must be enabled humans. Agent API credentials are
unsupported, including read requests. Display names never confer permission;
new messages carry actor ID/type and cannot gain human styling by being named
“You”. Legacy messages may have inferred rather than authenticated attribution.

### No implicit owner identity

Protected handlers require explicit authenticated actor context. Missing/malformed
identity never becomes the owner; actor-less sessions must sign in again. Owner-key
download independently requires owner context and setup/CSRF checks. Route authors
must still wire authentication and action authorization: this defense does not
protect an otherwise unguarded handler automatically. Never derive owner context
from caller headers, display names or body fields.

Human mutations use a per-actor token bucket: burst 20, refill one token/second,
at most 4096 tracked actors. Exhaustion returns 429 with `Retry-After: 1`.
GET/HEAD do not consume it; login has [separate limits](login-availability.md).
The owner emergency-stop endpoint bypasses this load budget while retaining all
identity/permission checks. Budgets are per controller, not distributed quotas.

## Context and disclosure

Chat context is capped at 128 KiB across profiles, imported references, shared
memory, recent messages and request. Oversized broadcasts return 413 before any
message is appended or participant launches. Reduce context explicitly rather
than silently dropping it. References remain untrusted input, never authority.

[Disclosure policy](context-disclosure.md) filters provider destinations,
human-only memory and cross-room exports on the server. Incident/emergency stops
also block native history listing/import before provider access. These controls
do not eliminate prompt injection, detect every secret or recall provider data.

## Storage and audit integrity

Use explicit migration credentials with `--migrate`; runtime startup must use the
[restricted database role](database-security.md) and performs no schema DDL.
[Atomic operations](atomic-operations.md) commit supported business events and
retry receipts together. [Storage waits](storage-deadlines.md) cap cooperative
replay, transactions, integrity and history operations at five seconds while
preserving earlier caller deadlines.

### Audit metadata format barrier

New events protect ID, sequence and commit time within version-1 `_auditRecord`
metadata. Legacy records may appear only before protected records. A later
unprotected record fails verification even when its chain hash is consistent.
Invalid verification latches execution off; owner recovery requires a valid chain.
Historical bytes are not rewritten or retroactively attributed.

The chain does not detect a fully rewritten valid history or removal of an entire
protected suffix. Independent immutable storage and externally anchored heads/
sequences remain required against privileged database rollback. Preserve invalid
records for investigation; deleting them is not recovery.

## Stops and recovery

With the approved remote worker enabled, startup performs a fresh bounded
`idle-v1` check. Active, unknown or failed worker state latches an application stop;
it never adopts, kills or replays unknown work. Resume requires a valid audit
chain, no integrity incident, no local active turns and fresh worker quiescence.
A transient SSH failure can therefore require explicit owner recovery.

### Emergency-stop audit outages

Failure to record a stop or resume escalates to the integrity stop and cancels
all chat/workflow contexts, even for a room-scoped request. The private marker
survives restart when its directory is available. Ordinary Resume cannot clear
it; explicit owner integrity recovery still requires valid history and worker
shutdown evidence. No record is forged after failed persistence.

If marker persistence fails, the controller remains stopped in memory but cannot
guarantee restart persistence. Repair storage and verify shutdown before restart.
Context cancellation never proves physical termination: fenced exact-ID cleanup
and released creator ownership remain independently required. Unknown cleanup
stays unresolved. See [worker jobs](worker-jobs.md) and
[paired shutdown](worker-pair-shutdown.md).

## HTTP and gateway limits

The local HTTP server caps headers at 16 KiB plus net/http parser lookahead,
with five-second header reads, fifteen-second total reads, fifty-second writes
and fifteen-second idle keep-alive. Oversized headers receive 431 before handlers.
Login proofs belong in bounded bodies; do not enlarge headers for credentials.
These are bounded transport controls, not public volumetric-attack protection.

Provider gateways cap live clients at 16 before parsing and active CONNECT slots
at eight. Exact provider-host/public-address policy, bounded requests and
cancellation still apply. See [gateway admission](container-isolation.md#gateway-connection-admission).

## Execution and credential separation

Mandatory Linux rootless containers have no direct host/SSH model-turn fallback.
The legacy credential-bearing development path is blocked before credential
access. Internal Codex pair assembly now exists, with fresh source bootstrap,
private credential delivery, owned gateway, audited callbacks and fenced shutdown.
Public production admission and live acceptance remain incomplete.

The engine contains provider credentials; the separate tool role contains source
and receives no provider home, token, gateway or network access. Read-only mounting
alone never protects secrets from reads. Native engine shell/filesystem/plugin
paths must remain unavailable; prompt instructions or redaction do not replace
that boundary. See [provider/tool separation](provider-tool-separation.md) and
[engine runtime](engine-container-runtime.md).

### Codex chat command-tool policy

Legacy conversational Codex turns and connection reply verification pass
`-c features.shell_tool=false` with the read-only sandbox. Worker validation
requires that exact policy; Claude chat uses an empty tools allowlist. Login is
separate trusted administration. Update controller/worker together: older
command-enabled chat requests must be refused, not silently accepted.

This narrows chat command exposure; it is not proof that every provider-native
read path is unavailable. Effective tool inventory still needs live pinned-client
acceptance. Development uses the separate mandatory broker boundary and cannot
fall back to credential-bearing tools.

## Private state and supply chain

[Private-file checks](private-file-security.md) validate opened descriptors for
credentials, keys, configuration, sessions and journals. They do not resist the
trusted host administrator. Never copy actual credentials or user conversations
into fixtures. Keep ignored local state and backups private.

Run official `govulncheck ./...` separately from fast offline tests and scan
reviewed images/provider/OS dependencies before publication. `go.mod` declares the
minimum Go version. Past clean scans are not current certification. Immutable
image references and metadata checks do not prove source/binary provenance.
Run `make secrets` before sharing source; hooks and scanners can miss unknown
formats and do not replace revoking a leaked credential.

## Verification boundaries

Automated tests mock every database interaction and use synthetic messages,
credentials, local sockets and temporary directories. They cover authorization,
failed transactions, tampering, deadlines, cancellation, quotas and protocol
ordering without real provider agents or Docker workloads. Deployment role/TLS
checks, live resource enforcement, provider authentication and transport-loss
termination require separate acceptance evidence.

Required security controls have no off switch. Optional UI/editor flags cannot
bypass authentication, authorization, audit integrity, isolation or retention.
Multiple controllers and hostile enterprise tenants remain unsupported promises.
