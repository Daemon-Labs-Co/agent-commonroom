# Agent CommonRoom red-team review — 10 October 2026

## Provider/tool separation: real Auditor and DockerRunner — 14 October 2026

Two of the mock-driven foundation's three "no real implementation" gaps are
closed in source: `cmd/server.storedToolAuditor` (a real `toolbroker.Auditor`
persisting intent/outcome through the existing append-only audit chain, as
new `tool.call.intent`/`tool.call.outcome` events replayed into
`room.State.ToolCalls`) and `cmd/server.dockerCLIRunner` (a real
`toolbroker.DockerRunner` shelling out to the local `docker` CLI with literal
argv, no shell, and the same local-only/resource-enforcing context check used
by the existing single-container agent path, extracted into
`verifiedLocalDockerContext` and shared rather than duplicated). A new
`cmd/server.newToolDispatcher` proves these assemble with the existing
`Dispatcher`/`ContainerExecutor`/`ToolContainerPlan` into a working whole.
Mocked/precondition-only tests pass; no live Docker or database was used.
See [provider/tool separation](docs/provider-tool-separation.md#real-auditor-and-dockerrunner-first-concrete-implementations).

This closes no finding by itself. `newToolDispatcher` has no caller outside
its own test — no live agent turn dispatches a real tool call through it, and
the transport question of how a provider's in-container tool call would be
intercepted and routed here remains a separate, un-started design decision.
Volume bootstrap, tool-container lifecycle pairing with its owning agent job,
and the real tool image build/install are also still unimplemented. A new
default-off `toolSeparation` feature flag was added as a rollout
marker/kill-switch ahead of that first real caller; it currently gates nothing
observable. **Shared-container credential disclosure (provider and tool
processes sharing one credential-bearing container) remains OPEN.**

## Provider/tool separation implementation plan — 14 October 2026

Scoped capability admission, an injected audited dispatcher, a trusted tool-container
launch plan and a bounded inspection verifier exist as unused `internal/toolbroker`
infrastructure with synthetic adversarial tests. The verifier refuses mismatched
mounts, redirected named-volume options, unexpected environment and weakened
selected runtime policy. It consumes only future trusted-daemon evidence, not
model/plugin claims. A mock-driven container executor adapter now verifies policy and routes a fixed
helper by immutable ID, with bounded cleanup/absence checks on failure. No real
Docker runner, installed image helper, concrete audit store, production endpoint or provider
bridge is wired. Credential exposure
remains OPEN; these components do not prove transport, image/kernel enforcement,
volume freshness, filesystem safety or pair shutdown.

Typed broker requests/actions now have pre-copy/hash/serialization budgets too,
including aggregate argument sizes and canonical non-allocating hex validation.
Broker/helper input and typed actions also reject malformed Unicode before lossy
JSON replacement can change paths/content; synthetic regression tests pass.
Synthetic early-refusal/allocation tests cover non-JSON callers. This hardens the
unused foundation; it does not change production provider execution.

The executor adapter independently enforces the copied declared-tool policy and
trusted scope/lifetime before any Docker interaction, in addition to dispatcher
admission. Synthetic tests reject undeclared command/write/patch operations while
preserving permitted reads. It also rejects malformed/inconsistent helper receipts
before output/success, and refuses command/patch operations without an implemented
helper contract. Consistent receipts are not independent filesystem attestation.
It remains an unused internal adapter, not an audit
or authentication API and not a deployed separation guarantee.

Unused `internal/toolworkspace` now provides bounded descriptor-rooted reads and
atomic replacement writes, refusing unsafe inodes and unconfirmed durability.
Synthetic temporary-file tests cover traversal, escaping links, FIFO/hard-link
refusal and pinned-parent replacement. Rooted paths are not container/mount
isolation, and no helper/RPC/production path is wired; credential exposure remains
OPEN. Runtime enforcement and concurrent-code/disk/crash concerns stay separate.

`cmd/tool-helper` now defines a strict bounded read/write subprocess protocol for
the future tool image. It rejects alternate roots/arguments, malformed/duplicate
fields, commands and patches; binary reads use bounded base64 responses and write
response failures are unconfirmed outcomes. Synthetic tests match the broker's
wire format. A fixed Linux-only shim now launches it with no shell/arguments and an explicit
credential-free child environment; mocked process tests pass. A tool-only image recipe now copies both with a narrow context and pinned-base
guard; no actual image build/install or real runner invocation has been verified; this does not
change production execution or close credential exposure.

The remaining credential/tool exposure now has an explicit two-container broker
design with scoped capability, audit acknowledgement, pair shutdown proof, fresh
volume migration and pinned-provider evidence gates. Vendor interface references
are distinguished from unverified adapter assumptions. Runtime integration is still PLANNED; no
production broker or security closure is claimed. It is a free/core mandatory boundary, not
a premium entitlement or isolation-off switch.
See [implementation/acceptance plan](docs/provider-tool-separation.md).

## Bounded history storage reads — 14 October 2026

History-setting lookup and event queries/row scans now use the mandatory storage
deadline, shared across a history request. Explicit context checks prevent partial
responses even when a timed-out row iterator reports no driver error. Mocked
settings/query/partial-row stalls cover paged and bounded legacy responses.
This narrows DB-pool/handler availability risk; live performance, volumetric
limits and durable total-storage quotas remain open.
See [history deadlines](docs/storage-deadlines.md#history-reads).

## Durable local provider credential revocation — 14 October 2026

Worker sign-out requires directory sync after removal, including absent-file
retries; sync errors do not acknowledge success. Login publication also syncs
its directory. Unknown providers cannot default to deleting Codex credentials.
Synthetic failure/retry tests pass without real credentials/providers. This is
local storage revocation only, not provider-wide token invalidation, historical
volume cleanup or verified physical crash durability.
See [credential durability](docs/connection-sign-in.md#durable-local-credential-publication-and-sign-out).

## Exclusive credential staging — 14 October 2026

Per-turn credentials are now exclusively created in a validated private parent,
without following/overwriting an existing destination. Permissions are explicitly
normalized for the mapped container UID, independent of host umask. Synthetic
fixtures verify alias/non-overwrite/input/parent/mode boundaries; no real secrets
were accessed. Mapped-UID runtime behavior and crash cleanup remain unverified.
This does not close credential visibility to tools in the same container.
See [staging boundary](docs/private-file-security.md#exclusive-per-turn-credential-staging).

## Fail-closed identity defaults — 14 October 2026

Missing request context and actor-less sessions no longer inherit owner identity.
Owner-only handlers refuse unidentified requests; key download has its own owner
check alongside CSRF/middleware. Synthetic tests cover direct-handler refusal,
raw/hashed session rejection and explicit owner-key lifecycle. Route registration
must still use authentication/authorization middleware; no universal claim is
made for handlers that omit their own permission checks.
See [identity contract](docs/security-hardening.md#no-implicit-owner-identity).

## Legacy shutdown-proof barrier — 14 October 2026

Legacy terminal worker statuses no longer count as verified shutdown. Only
current-format terminal records with explicit confirmation are settled; legacy
records remain unchanged and visible as unresolved, blocking admission/idle
recovery. Synthetic matrices and temporary-record tests cover these paths.
An operator must independently review/archive old operational records during
maintenance; no automatic proof, rewrite or task replay is supplied.
See [upgrade guidance](docs/worker-jobs.md#legacy-terminal-records-after-upgrade).

## Development snapshot substitution — 14 October 2026

The launcher now mounts a private checked per-turn copy rather than an original
host source path that may change after hashing. Source descriptor checks use
no-follow/nonblocking opens and bounded actual reads. Synthetic replacement,
FIFO/symlink/write-permission/size and failure-cleanup tests pass without Docker
or providers. Live mounting, crash temporary cleanup and worker-account compromise
remain outside this verification. Credential/tool co-location remains OPEN.
See [staging contract](docs/development-workspaces.md#per-turn-verified-snapshot-staging).

## Provider gateway pre-header admission — 14 October 2026

A fixed 16-client listener limit now covers accepted sockets before headers,
DNS or CONNECT slot admission. Excess sockets close immediately and slots are
released once through concurrent/double cleanup. Existing 8-tunnel and timeout
limits remain. Synthetic pipe and ephemeral local Unix tests exercise the bound
without DNS/providers/DB. This narrows gateway memory/connection exhaustion;
volumetric CPU pressure, global public-server admission and credential/tool
separation remain OPEN. It is a required free/core limit, not an optional toggle.
See [gateway admission](docs/container-isolation.md#gateway-connection-admission).

## Provider gateway cancellation — 14 October 2026

Gateway closure now cancels pending DNS/dial work and rejects new requests before
resolution, in addition to closing listener/tunnels. Pre-tunnel read/write/idle
timeouts are explicit. A synthetic blocking resolver verifies cancellation with
no external requests. Physical worker shutdown and development credential
separation remain separate requirements. See [gateway shutdown](docs/container-isolation.md#gateway-shutdown-and-pre-tunnel-limits).

## Codex chat command-tool restriction — 14 October 2026

Chat and reply verification now explicitly disable Codex command tools; read-only
alone permitted tool reads. Worker exact-prefix validation rejects removing or
overriding the restriction for fresh/resumed turns. Synthetic argument tests
pass; deployed CLI tool availability remains unverified. Development credential/
tool co-location remains OPEN and needs a separate worker/broker boundary.
See [chat policy](docs/security-hardening.md#codex-chat-command-tool-policy).

## Provider transition-address denial — 14 October 2026

Provider egress now rejects IPv6 translation/transition and other special-use
routes, plus reserved IPv4 space that generic global-unicast classification can
admit. Mapped IPv4 is classified as IPv4; native public IPv4/IPv6 remains allowed.
Exact provider hosts and validation/pinned dialing are preserved. Synthetic
address regressions only; no DNS/provider access. NAT64-only connectivity is
intentionally unsupported. See [address policy](docs/container-isolation.md#native-public-provider-address-policy).

## Bounded credential-safe DB startup — 14 October 2026

Runtime pool/connection/privilege startup shares a ten-second deadline and does
not print raw connection errors. Mocked connection/query/row stalls and
least-privilege tests verify refusal without live DB access. Unknown privileges
still fail closed. Cooperative cancellation and operator migration timeout remain
requirements. See [DB startup checks](docs/database-security.md#bounded-startup-checks).

## Local controller startup fencing — 14 October 2026

The supported single-controller deployment now holds a private persistent-inode
kernel lease before DB/identity/worker startup. Concurrent startup in the same
data directory fails closed. Separate directories/hosts and old binaries remain
outside this lease; distributed controller coordination is still OPEN. Do not
unlink an active lease or launch two installations against one DB. Synthetic
lease tests verify exclusion/release/inode stability without real services.
See [startup lease](docs/operations.md#single-controller-startup-lease).

## Session/stop directory durability — 14 October 2026

Session replacement/removal and emergency-stop marker creation/removal now sync
the parent directory. Errors remain failures; recovery relatches and retries
marker creation when removal durability fails. Synthetic failure-injection tests
verify complete replacement before sync, propagated errors and temporary cleanup.
Filesystem/hardware crash guarantees and unavailable marker storage remain
operator/integration concerns. See [durability limits](docs/private-file-security.md#crash-durability-of-sessions-and-emergency-stops).

## Controller configuration descriptor checks — 14 October 2026

Private local and development configuration now use bounded descriptor validation
including ownership, link count and file type. This removes the development
loader's check/reopen race and the local loader's unchecked read. Invalid local
JSON is rejected without echoing credential-containing input. Synthetic fixtures
cover unsafe file types/links/permissions/size and strict decoding. No live
configuration or credential was read. See [private-file configuration](docs/private-file-security.md#controller-configuration).

## HTTP header parsing and credential-tool exposure — 14 October 2026

HTTP headers now have a mandatory 16 KiB budget (plus Go parser lookahead), with
explicit bounded idle/header/read/write durations. Synthetic loopback requests
verify HTTP 431 before handler invocation and subsequent availability.

**OPEN: provider credential disclosure from development tools.** Current provider
and tool processes share a credential-bearing container/UID. Read-only credentials
can still be read and emitted through model output. Separate provider execution
from credential-free tool workers via a narrow broker; do not claim prompt rules
or output filtering fix this. Source review only; no credential reads or live
provider tests. See [details](docs/security-hardening.md#open-credentials-shared-with-development-tools).

## Durable emergency-stop audit failure — 14 October 2026

Failed stop/resume audit persistence now latches the private integrity marker and
cancels all contexts, including when the request originally targeted a room.
Restart restoration and explicit owner integrity recovery apply; ordinary Resume
cannot clear the failed-audit stop. Mocked begin/insert/commit failures test marker
persistence and cancellation without live data. An unwritable marker directory
still prevents a restart-persistence guarantee and requires operator repair.
See [emergency-stop outages](docs/security-hardening.md#emergency-stop-audit-outages).

## Storage deadline hardening — 14 October 2026

State replay, append transactions and integrity verification now cap cooperative
storage work at five seconds. Mocked query/row/begin/insert stalls roll back, and
a controller-lock regression verifies emergency stop can proceed after a stalled
read times out. Integrity timeout cancels contexts and latches execution off.
This narrows storage-driven cancellation/availability risk; arbitrary locks,
non-cooperative backends and physical worker termination remain separate concerns.
See [bounded storage waits](docs/storage-deadlines.md).

## Audit format downgrade barrier — 14 October 2026

Verification now rejects unprotected records after the first metadata-protected
record, even when hashes and links are consistent. Legacy prefixes remain
unchanged. Mocked tests verify cancellation and refusal of owner recovery on a
synthetic downgrade. This is mandatory local integrity enforcement, not independent
rollback protection: stripping an entire protected suffix or rewriting all records
still needs an externally anchored head/sequence. See [audit format barrier](docs/security-hardening.md#audit-metadata-format-barrier).

## Authentication availability hardening — 14 October 2026

Valid owner/invited-human proofs remain usable after the shared failed-attempt
response budget is exhausted; invalid attempts still receive throttled responses.
Unknown/expired or malformed member challenges cannot force repeated DB replay.
Session renewal works at capacity and restores the prior session on persistence
failure. Tests cover RSA/owner/member proofs, replay, strict decoding and raw/hashed
session rollback. This does not claim volumetric/challenge-pool/cryptographic CPU
DoS resistance. See [login availability](docs/login-availability.md).

## Private file/lease substitution hardening — 13 October 2026

Security-critical file reads and persistent lock opens now validate the opened
inode with no-follow/nonblocking flags, ownership, private permissions, link count
and bounded size/reads. Synthetic tests cover link substitution, FIFO rejection,
aliased locks and descriptor stability. This narrows worker/credential state-file
substitution risks; trusted parent directories, same-account/host administrators
and independent content/rollback integrity remain outside the guarantee. No live
credential or deployment verification was performed. See [private file safety](docs/private-file-security.md).

## Connection verification lifecycle — 13 October 2026

Provider connection checks now use controller journal allocation, managed run-v2,
emergency cancellation, verified shutdown and restart reconciliation. Legacy
worker verify-v1 model launch is rejected. Verification counts toward admission
limits and prevents connection edit/delete/sign-out during its active reservation.
Mocks cover allocation/cancellation/shutdown failure and HTTP integration. This
closes the identified controller-journalling gap for connection verification in
source; live worker/controller upgrade remains unverified. Human login operations,
legacy worker records, independent audit rollback protection and distributed
ownership are separate remaining issues. See [connection lifecycle](docs/connection-testing.md#controller-owned-verification-lifecycle).

## Durable launch-volume budget — 13 October 2026

Controller admission now has a mandatory application-wide UTC-day allocation
budget, default 200, configurable from 1 to 100000. The job journal supplies usage;
restart and optional diagnostic flags do not reset it. Registration rechecks usage
under the controller lock and appends intent with an expected event count. Owner
diagnostics expose safe usage/reset details. Mocked tests cover restart, final-slot
races and rollback. This narrows SEC-10: monetary/token/storage budgets, legacy
pre-journal work, independent rollback protection and distributed controllers
remain outside this guarantee. See [execution budgets](docs/execution-budgets.md).

## Task dispatch transaction hardening — 13 October 2026

Task send/review now uses an owner-only server command rather than separate browser
agent/task writes. Intent and outcomes are atomic, retries use durable receipts,
and pending task records cannot be edited or replayed until result or verified
worker-quiescence reconciliation. Task dispatch is optional/default-on; disabling
it preserves history and does not bypass mandatory execution safeguards. This
narrows the browser dispatch gap in SEC-13; distributed controllers, independent
audit anchoring and exactly-once external effects remain unsupported. Tests mock
all DB/provider interactions. See [task dispatch](docs/task-dispatch.md).

## Workflow outcome transaction hardening — 13 October 2026

Background workflow success/failure outcomes now commit run and participant state
atomically. Gate/completion persistence errors are handled; state-read, intent and
outcome storage failures latch the mandatory integrity stop rather than silently
ending the runner. Mocked regression tests cover partial-write rollback and prevent
next-node dispatch after failed outcome persistence. This narrows SEC-13; browser
initiated task dispatch, distributed ownership, independent rollback protection
and external action reconciliation remain open. No live provider/DB verification
is implied. See [workflow persistence](docs/atomic-operations.md#workflow-outcome-persistence).

## Disclosure and atomic business operations — 13 October 2026

Server-side prompt assembly checks room/provider disclosure policies; foreign-room
messages, unsupported cross-room Project memory, human-only notes and unapproved
Application exports are excluded. Owner-managed source/provider grants and native
session resets preserve application history. The editor flag cannot bypass checks.
Reference source IDs and trust labels are included; prompt injection prevention,
provider-side deletion, automatic secret classification and repository-content
redaction are not claimed. See [disclosure controls](docs/context-disclosure.md).

Compound agent/workflow intent and result writes use one checksummed transaction.
User event and workflow mutations support durable, actor/room-bound retry receipts
and transactional stale-state checks. Company, room, connection and participant
edits now require revisions. Single-controller support remains explicit; distributed
workflow ownership and transactional external provider effects remain unfinished.
See [atomic operations](docs/atomic-operations.md). Verification is mocked; no live
provider or PostgreSQL data is used to validate these changes.

## Free worker lifecycle hardening — 13 October 2026

Implemented on the durable-worker-jobs feature branch: mandatory controller intent
journal, worker protocol-v2 replay deadlines, persistent job/member file leases,
worker-level global/room admission, pre-launch cancellation tombstones, verified
supervisor/container shutdown, explicit owner reconciliation and bounded debug UI.
Unconfirmed shutdown or intent blocks replacement/resume; startup checks both
journal and worker. Optional Free `jobDiagnostics` defaults on and cannot disable
execution safeguards. Tests use mocked database/transport and synthetic temporary
supervisors; provider interoperability is not implied by those tests. An explicit Debian
rootless smoke verified synthetic container shutdown in 2.45 seconds. Matching
worker/controller binaries were installed after verified idle; legacy completed
records remained readable. No provider model was launched for validation.

This mitigates ordinary managed-turn crash/cancellation and duplicate SSH job
execution (SEC-06/09/10). It does NOT complete HTTP idempotency, atomic business
updates or multi-controller workflow ownership (SEC-13/ARCH-01), connection-check
controller journaling, legacy-job adoption or host-administrator resistance.
See [contracts, deployment and remaining gaps](docs/worker-jobs.md).

## Mandatory hardening — 13 October 2026

Implemented (source and mocked tests): explicit migration/runtime database separation,
startup effective-privilege rejection, pinned search path, startup/30-second audit
verification, cancellation on verification/storage failure, persistent stop marker,
owner-only verified recovery, startup worker quiescence checks and verified emergency resume, checksummed new-record ID/sequence/time metadata,
authenticated audit attribution, human-only API identity, unambiguous message
styling, per-actor mutation budgets and pre-dispatch chat context limits.
pgx/text dependencies were updated; govulncheck reported no remaining reachable
Go advisories with Go 1.27.2. A redacted source-only Gitleaks scan found no secrets.
Neither scan covers all provider image/OS dependencies or proves absence of secrets.

The local controller and Debian worker were updated after verified worker idle.
Local PostgreSQL runtime grants and CommonRoom-only SCRAM/peer rules were applied;
passwordless runtime access was rejected, and the new controller started with the
restricted account. No audit records were modified for verification and no
provider model was launched. Other installations using owner credentials must follow [restricted database setup](docs/database-security.md)
before the new controller can start. Role/HBA verification is deployment-specific; live provider execution under
this update and fresh-server installation remain unverified. Read [operational limits](docs/security-hardening.md).

Still open: independent audit anchoring/rollback protection (SEC-05), legacy
metadata protection, durable job reconciliation and confirmed cancellation under
transport failure (SEC-06/09), transactional/multi-controller coordination,
spend/storage limits and disclosure controls (SEC-10/11/12/13). Do not mark these
closed on the basis of checksums, in-memory budgets or successful mocked tests.

## Incremental execution hardening — 12 October 2026

Session recovery follow-up: owner-only check/resume/fresh controls require approved
worker quiescence, block active controller work/stops/incidents, and atomically
audit the resulting session snapshot. Unknown worker state and nonterminal job
records fail closed. This protects the new recovery path; ordinary dispatch,
connection reassignment and workflow resume still need durable per-job reconciliation
to establish universal restart safety. The whole-worker idle check assumes one
controller and may block on stale records; no automatic force override is offered.

Worker telemetry follow-up: background operational observations are now bounded,
job/provider-validated over approved SSH, and protected by an owner-controlled
server flag. Raw commands/outputs are excluded. The expandable view can invoke
the existing mandatory owner emergency stop independently of cosmetic safety
button visibility. This improves SEC-09 observability; it does not resolve orphan
reconciliation, durable cancellation enforcement or malicious-worker trust.

The original findings below are historical evidence, not a fresh finding that
every cited behavior still exists. Authentication, authorization, isolation and
bounded-output work are described in later remediation notes.

SEC-10 now additionally has mandatory single-controller admission limits shared
by chat and workflows: eight active model turns overall and four per room. Chat
rejects an over-capacity batch before any member is dispatched. Workflow overload
fails explicitly for owner recovery. Mocked tests cover global/room bounds,
removed active jobs, chat rejection without writes and workflow rejection without
provider calls. This does **not** close SEC-10: spend/rate/context quotas, durable
leases, multi-controller bounds and global-lock latency still require work.

Status diagnostics preserve deadline causes and report elapsed dispatch age;
they do not verify worker health, progress or cancellation. No live provider
long-running scenario has been replayed for this change.

Next security priorities:

1. Durable job identity/reconciliation and authenticated worker lifecycle events,
   including confirmed cancellation and fail-closed handling of orphaned jobs
   (SEC-06/09, ARCH-01). Never extend deadlines solely on heartbeats.
2. Independently controlled audit anchoring, rollback detection and least-privilege
   database migration/runtime roles (SEC-05). Checksums alone are insufficient.
3. Remaining multi-actor capability/attribution, context-disclosure and transactional
   command protections (SEC-01/03/11/12/13). Re-test each finding against current
   code rather than treating the historical list as an up-to-date verdict.

Scope: current Go/Postgres/JavaScript implementation, source inspection and non-destructive local HTTP probes. No agent launches, private payload disclosure, live database tampering, or production exploitation. This is a prioritized developer backlog, not a guarantee that all vulnerabilities have been found.

Threat model: multiple cooperative and malicious humans, compromised agents, malicious model output and shared context, compromised SSH destinations, hostile browser sites, other local processes, and administrators able to alter storage. The present application supports a single trusted local owner; that trust assumption does not satisfy this threat model. Priority denotes repair order, not a formal CVSS score. The original review did not perform a dependency audit; the 13 October update below records the later Go scan.

## P0 — required before admitting untrusted actors or remote users

### SEC-01: No authenticated actor identity or authorization
Evidence: cmd/server/main.go:87 exposes the sole CSRF token without login. State and events GET handlers also require no authentication. Events and agents POST handlers only compare this token. Actor attribution is hard-coded as human/You. Read-only probes confirmed /api/session, /api/state and /api/events return HTTP 200 without credentials; payloads and token values were not printed.
Impact: any process that reaches the listener can obtain the token, read conversations, impersonate the owner, change connections/profiles, accept tasks and invoke agents with the owner's provider accounts and SSH credentials. Agents cannot be safely given API access under this design.
Fix: authenticate sessions; assign immutable actor IDs and human/agent/service types; enforce permissions on every resource/action, with narrowly scoped agent credentials. Separate CSRF protection from authentication. Attribute actions from authenticated identity, never request text or display names. Agent actors must not edit human identities, company ownership, permissions, connection endpoints or incident recovery.
Acceptance: anonymous reads/writes rejected; agents cannot perform human-only operations; two human sessions have independent revocable credentials; spoofed actor fields have no effect.

### SEC-02: Host/origin protection missing; configurable cleartext network exposure
Evidence: cmd/server/main.go:68–81 defaults to loopback but permits arbitrary LISTEN_ADDR and serves plain HTTP; no Host or Origin validation. Read-only probes using Host: untrusted.example returned 200 on session, state and events. Current CSP and absence of permissive CORS are useful, but do not establish authentication or block DNS rebinding.
Impact: foreign Host acceptance enables a DNS-rebinding attack path, subject to browser networking protections; binding to a network interface exposes the unauthenticated API and token over cleartext.
Fix: exact Host allowlist, Origin/Fetch Metadata validation for browser actions, authenticated sessions, secure cookie policy, and TLS for remote access. Reject non-loopback deployment unless secure deployment configuration is explicitly enabled. Configure WebSocket origin checks before implementing that transport.
Acceptance: unexpected Host and Origin rejected, normal localhost works, remote endpoints require authentication and TLS. Do not claim ordinary cross-origin JavaScript can currently read responses without a rebinding or other same-origin bypass.

### SEC-03: Room/company scoping is a view filter, not an access boundary
Evidence: cmd/server/rooms.go:requestedRoom/scoped; all callers choose any existing room query. Company membership is assigned to the singleton company during replay in main.go; no actor-to-company/room authorization exists. Conversation list/read actions expose saved Codex conversations outside Agent CommonRoom without resource-specific authorization.
Impact: a future room-limited user/agent can request other rooms, shared memory or unrelated native session history. Multiple companies cannot be isolated by adding a selector alone.
Fix: bind actor credentials to company and room memberships; authorize resource IDs independently of query filters; partition company data and shared memory. Keep external conversation import human-only with explicit approved sources. Distinguish administrator access from active room participation.
Acceptance: malicious room/company/session ID substitutions return denial; no cross-tenant reads, writes or enumeration.

### SEC-04: Chat-only instructions do not enforce Codex directory/read isolation
Evidence: cmd/server/agents.go:24 runs Codex read-only with approvals never; no per-room filesystem allowlist is enforced. Prompts include lower-trust conversations, shared memory and imported context. Claude explicitly disables tools, which is stronger than the Codex path here. Native sessions/processes inherit the OS account's environment and credentials.
Impact: malicious context may persuade a Codex agent to read accessible files or attempt network/tool actions; read-only does not mean no reads or no disclosure. SSH dispatch brings the remote account's access into scope. Exact provider sandbox behavior is platform-dependent and was not penetration-tested.
Fix: enforce tool capabilities outside the model. For chat-only mode disable tools using a supported provider mechanism or isolate the process with no sensitive readable mounts, restricted network and sanitized environment. For coding mode use dedicated accounts/containers and explicit room workspaces, secret isolation and egress policy. Treat prompts as data, never authorization.
Acceptance: adversarial context cannot read outside the allowed workspace, obtain credentials or contact unauthorized destinations.

## P1 — significant current weaknesses

### SEC-05: Audit history is not administrator-resistant; rollback and metadata tampering escape verification
Evidence: migrations/001.sql prevents UPDATE/DELETE/TRUNCATE via a trigger, but the application runs DDL at startup and can replace that trigger. Checksums cover only previous hash, kind and payload; event ID, sequence and committed_at are excluded. Verification has no independently anchored expected count/head. An offline reproduction showed truncating a valid suffix leaves a prefix that passes the present algorithm.
Impact: sufficiently privileged database actors can disable triggers, rewrite/re-hash history or restore an older valid snapshot without detection. Altered timestamps/IDs are not detected by this checksum. This fails the stated absolute immutability requirement.
Fix: separate migration owner from append-only application role; include canonical event identity, sequence and timestamp in protected content. Independently anchor sequence/head and use an externally controlled append-only/WORM archive with protected backups and retention policy. Detect rollback against the independent anchor. Hashes alone do not prove an event was truthful.
Acceptance: metadata changes, prefix rollback and chain recomputation are detected; application role cannot alter history or protections. Test on an isolated database, never tamper with the live log.

### SEC-06: Integrity incident does not disconnect active agents
Evidence: main.go:747 integrity appends incident.opened on mismatch but does not call s.cancels. runAgent suppresses a successful reply when Incident is true, but running processes may continue. No startup/periodic verification is evident; mutation paths do not universally fail closed on integrity loss. Conversation list/read runs before the agent incident guard.
Impact: a compromised audit store does not trigger the promised complete agent shutdown. Work and possible external activity continue until completion/timeout; history subprocesses still run in incident mode.
Fix: centralized incident transition under synchronization: freeze agent capabilities, cancel local process groups and remote jobs, revoke credentials, block all agent paths, and verify on startup and periodically. Require authenticated human recovery and explicit reinvitation. Decide which human repair writes remain permissible and audit them independently.
Acceptance: a test integrity failure stops every active job, blocks dispatch/history tools and requires authenticated recovery. Failed verification must remain fail-closed even if incident persistence fails.

### SEC-07: Stored HTML injection through workflow IDs
Evidence: workflows.go:23/38 accepts arbitrary nonempty node/edge IDs. web/workflows.js:25/30 interpolates these values unescaped into HTML/SVG attributes using innerHTML. Names/text mostly use esc, but IDs bypass it.
Impact: a malicious API writer can persist quotes/markup in IDs and inject elements or links into the editor, spoof its UI or disrupt it. CSP blocks many script payloads; arbitrary JavaScript execution has not been demonstrated and must not be assumed. A same-origin HTML injection is still a security defect.
Fix: generate IDs on the server or validate a strict bounded identifier format; build DOM/SVG with createElement/setAttribute, or consistently escape every attribute. Reject unsafe workflow IDs at ingestion. Preserve restrictive CSP.
Acceptance: quote/markup payloads are rejected or rendered only as inert values, including node IDs, edge IDs and referenced IDs.

### SEC-08: Connection changes can reuse the wrong native session
Evidence: main.go participant.updated accepts a different ConnectionID while retaining b.ThreadID and b.Status. Connection endpoint edits reset sessions during replay, but changing a participant's connection does not apply equivalent protection. Updates are not rejected while that participant is active.
Impact: a session identifier can be carried to a different provider/account/host, with potential unintended resume, context disclosure or mismatched attribution. An in-flight turn uses the old captured endpoint while later profile state points at the new connection.
Fix: reject active connection reassignment; clear or explicitly migrate native session bindings when a participant changes connection. Bind session identity to provider, endpoint/account and membership, not just a string. Audit migration and validate ownership.
Acceptance: reassignment while working denied; reassignment while idle cannot silently resume the prior connection's session.

### SEC-09: SSH destinations are not an application-owned allowlist; cancellation is incomplete
Evidence: model.go SupportsAgent accepts any syntactically valid SSH alias. agents.go:475/507 uses the owner's SSH configuration and credentials. Arguments are correctly shell-quoted and aliases cannot begin with an option; no shell-injection claim is made. CommandContext kills the local command, with no explicit remote job cancellation protocol or local process-group management.
Impact: callers with connection-edit access can dispatch to another reachable SSH destination; prompts may be sent to a malicious or unintended host. Timeouts/removal do not reliably prove remote descendants stopped.
Fix: human-admin-approved destination registry, dedicated least-privilege SSH credentials and pinned/managed host keys; track remote job IDs and implement acknowledged cancellation. Kill local process groups and report unconfirmed remote termination honestly.
Acceptance: unapproved hosts rejected; cancellation prevents late jobs and confirms remote shutdown or reports failure.

### SEC-10: Unbounded resource consumption and global lock blocking
Evidence: agents.go holds s.mu during history RPCs (up to 20 seconds) and provider status checks; all mutation paths share it. Broadcast launches one goroutine/process per eligible participant with no global/per-actor concurrency quota. Claude cmd.Output and stderr buffers are unbounded; Codex accumulated text/stderr are unbounded despite a per-line scanner limit. State replays every event and dispatch verifies the whole log.
Impact: malicious or merely busy actors can exhaust memory, processes, provider spend and database/disk capacity; slow history requests block unrelated room mutations. Append-only storage grows without quotas/pagination.
Fix: bounded job queue, per-actor/company quotas, cost limits, output byte caps, rate limiting, context-size budgets, paginated history and checkpoints validated against audit anchors. Move external calls outside the global lock with explicit state revalidation.
Acceptance: burst/slow-provider/large-output tests have bounded memory and work; one actor cannot monopolize all rooms. Do not load-test the user's live provider accounts.

### SEC-11: Display names can impersonate humans/agents; audit attribution lacks stable identities
Evidence: profile names allow duplicates and the literal You. Assistant messages store Author as display name, not participant ID. Browser code interprets author===You as human and otherwise guesses provider from names. Existing mutable profiles can affect displayed historical branding.
Impact: an agent named You can appear to be the human; duplicate names can confuse directed messaging and review attribution. Evidence/provenance cannot reliably identify the responsible actor from names alone.
Fix: store actor ID/type, membership ID and provider snapshot on every message/event; render human/agent badges from immutable actor type. Validate unique room aliases used for addressing; permit display-name duplicates only when IDs are explicit. Keep historic attribution immutable.
Acceptance: agent named You is visibly an agent; duplicate aliases cannot misroute messages; renaming never changes past attribution.

### SEC-12: Stored context and memory have no trust labels or disclosure policy
Evidence: agents.go:389 concatenates profile, imported conversations, shared memory and recent room messages into one prompt. Application/Project memory is automatically visible across rooms; changing scope can broaden visibility. API currently has no agent-vs-human permissions or explicit export policy.
Impact: malicious agents/humans can plant instructions, launder unsupported claims through shared context or broaden data exposure; material may be sent to a remote/provider account without resource-level approval. Current chat-only capability restrictions reduce but do not eliminate disclosure risk.
Fix: retain source actor, scope and trust metadata; mark lower-trust context clearly and enforce tool/data policies outside prompts. Require authorized scope changes and explicit sensitive-context export permissions, with minimization/redaction. Keep evidence verification independent of agent claims.
Acceptance: lower-trust text cannot authorize tools or scope expansion; denied information is excluded from prompts before provider dispatch.

## P2 — multi-actor correctness and deployment preparation

### SEC-13: Last-writer-wins updates and non-atomic multi-event operations
Evidence: tasks, memory and workflows have revision checks; company/human/room/connection/participant updates do not. Attach creates connection, participant and attachment audit using separate transactions. Broadcast persists/starts targets incrementally. Server mutex protects one process only.
Impact: concurrent users or server instances overwrite changes; interrupted multi-event operations leave partial state. Duplicate dispatch across instances is possible because active ownership lives in memory.
Fix: optimistic revisions for all mutable aggregates, transactionally committed operation bundles, idempotency keys, database-backed job ownership/leases and revalidation before dispatch.
Acceptance: simultaneous stale edits return conflict; retrying attach/send does not duplicate side effects; multiple instances cannot run the same job twice.

### SEC-14: Missing secure deployment and supply-chain verification
Evidence: default DB URL uses sslmode=disable; startup executes schema-owner DDL. No application-level encryption/redaction/secret-store integration is visible. Provider executables are resolved through PATH or remote home paths; inherited configuration/environment is not tightly controlled. Dependency/CLI vulnerability scans were not run during this review.
Impact: backups/log/context may retain sensitive material; a network database needs transport protection; executable/configuration compromise inherits the application's access. These are deployment-dependent risks, not demonstrated package vulnerabilities.
Fix: separate DB credentials/roles and require verified TLS for network databases; protect storage/backups/credentials and document retention; validate executable provenance/version and sanitize environments. Add govulncheck, dependency/CLI update review, secret scanning and security regression checks to CI.
Acceptance: documented deployment passes least-privilege checks, denies insecure remote DB configuration and has repeatable dependency scanning with reviewed results.

## Current positive controls
Loopback default; SQL parameters rather than string SQL interpolation; 64 KiB request caps; restrictive CSP and nosniff; escaped ordinary UI text; syntactically restricted SSH hosts and correctly quoted remote arguments; Claude tools disabled; task/memory/workflow revision checks; removed-participant checks suppress late replies. Keep these controls. None replaces authenticated identities or independent immutable storage.

## Recommended implementation order
1. Authenticated actor identities, authorization, Host/Origin safeguards and isolated agent capabilities.
2. Audit role separation/external anchors and complete fail-closed incident shutdown.
3. Safe workflow DOM/IDs, stable message identities and connection/session ownership.
4. Resource quotas, process lifecycle control and transactional multi-actor updates.
5. Deployment/secret controls and continuous adversarial regression tests.

Workflow runner, cross-room transfer, arbitrary HTTP/MCP dispatch and multiple-company tenancy are not implemented. Before enabling them, add authorization at each handoff, SSRF/egress protection, authenticated protocol peers, signed provenance and bounded execution/approval rules. These are future design gates, not claims of exploitable current HTTP adapters.

# Principal-engineer architectural review — 10 October 2026

This section extends the security backlog above. It is a source-based architectural assessment, not a refactor or execution-engine implementation. Priorities describe dependencies for the next milestones. Security fixes SEC-01–04 remain the first gate for untrusted actors.

Verdict: a useful prototype with the right major ingredients (Go, PostgreSQL, explicit room profiles, saved graphs and evidence), but not yet a resilient multi-actor orchestration platform. Keep Go/PostgreSQL and build a modular monolith with a durable worker subsystem. Microservices or a wholesale rewrite are not justified by the current scale.

## ARCH-01 — P0: Durable execution is missing
Evidence: agents.go stores active work and cancellation in server maps and starts background goroutines. The saved Working status survives restart; active ownership and cancellation do not. There is no startup reconciliation, durable dispatch queue, lease or terminal job record.
Consequence: a crash can strand participants as Working, lose results or leave remote work continuing; a subsequent process may dispatch overlapping work. Three-minute contexts are not durable deadlines.
Fix: durable Job/Attempt records with queued/running/completed/failed/cancelled states, operation IDs, leases, heartbeats and startup reconciliation. Commit dispatch intent with its audit event, then claim jobs in a worker. Make provider callbacks and result commits idempotent. Report uncertain remote completion explicitly; do not promise exactly-once external execution without provider support.
Acceptance: restart during each dispatch/result stage recovers a visible, consistent outcome without silently repeating non-idempotent work. See SEC-09/10/13.

## ARCH-02 — P0: Identity, agent definition and room membership are conflated
Evidence: model.go Room contains HumanName/HumanRank/HumanPresent, State contains another HumanName/HumanTitle, Participant mixes room role, session and connection identity, and CompanyID is assigned uniformly during replay. No durable company-level agent/person roster exists separately from connections and room participants.
Consequence: a second human or company requires special-case expansion; one bot's organizational identity cannot be reliably tracked across room profiles. Transport registration is being used as a substitute for an agent identity.
Fix: explicit Company, Actor (human or agent), CompanyMembership, AgentDefinition, Connection, RoomMembership/Profile and SessionBinding. Rank/title/presence overrides belong to room membership; default organizational title belongs to the actor/company membership. Removing room membership should retain organizational identity. Use foreign keys and actor IDs in messages/audit records.
Acceptance: two humans and one bot in two rooms retain distinct identities, independent profiles and correct provenance. See SEC-01/03/11.

## ARCH-03 — P1: HTTP, domain rules, event replay and execution share too much responsibility
Evidence: cmd/server/main.go contains startup DDL, HTTP routing, command validation, event persistence, projection replay and integrity handling. agents.go mixes protocol decoding, history import, prompt assembly, dispatch and job lifecycle. Most code remains package main.
Consequence: domain invariants are duplicated or bypassed by alternative paths (generic participant creation versus attach/create), and authorization will be difficult to apply consistently.
Fix: modules for identity/company, rooms, agents/sessions, tasks/workflows, memory, evidence and audit. Thin HTTP handlers call application services; services apply policies and invoke repositories/adapters. Keep infrastructure dependencies behind small interfaces. Avoid splitting services into separate deployments until operational demand exists.
Acceptance: the same domain command behaves identically through HTTP, CLI or a future MCP endpoint and can be tested without an HTTP server/provider process.

## ARCH-04 — P1: The event store doubles as an unversioned, expensive read database
Evidence: load scans all events for every state request and many commands, repeatedly linearly searches aggregates, and discards json.Unmarshal errors. Event records have no explicit schema version. Dispatch separately verifies the whole log.
Consequence: replay costs grow with lifetime activity (some aggregate lookups compound that cost); schema changes or malformed events can silently produce incorrect state. Full-history responses are unsuitable for long-lived rooms.
Fix: typed/versioned event envelopes, explicit validation and deterministic reducers; fail closed on undecodable state. Maintain transactional read projections with indexes and paginate messages/evidence. Add replay checkpoints validated against audit anchors, and controlled event upcasters preserving original records. Keep audit authority separate from efficient current-state queries.
Acceptance: an old event corpus replays deterministically across upgrades; invalid events fail visibly; large-history query cost depends chiefly on page size. See SEC-05/10.

## ARCH-05 — P1: Task status is an enum, not an enforced lifecycle
Evidence: main.go task.revised validates that Status is one of seven strings, but accepts arbitrary transitions. features.js offers status editing and dispatches an agent before separately writing Working/Reviewing. Agent completion appends a chat reply without a correlated task attempt/result transition.
Consequence: tasks can be marked Accepted without submitted work/review; a dispatch can succeed while its task update fails; chat replies cannot reliably identify which attempt they satisfy. Human override and ordinary transition are indistinguishable.
Fix: explicit commands (assign/start/submit/request-review/request-changes/resume/accept/override) with transition rules, authorization and reasons. Link job/attempt IDs to task revisions, evidence and findings. Commit scheduling intent plus task transition atomically; handle results through idempotent commands. Keep human overrides explicit and audited.
Acceptance: invalid ordinary transitions are rejected; QA changes return to development; dispatch failure leaves an honest task state. See SEC-01/13.

## ARCH-06 — P1: Adapter abstraction does not match extensibility requirements
Evidence: AgentAdapter exposes only Capabilities while execution uses a separate sessionAdapter.Run. Dispatch selects Codex versus Claude through endpoint strings; supported-provider logic is duplicated in Go and JavaScript. HTTP/MCP/WebSocket are saved labels without executable adapters.
Consequence: adding a protocol requires edits across multiple switches and UI rules; capability support is static rather than tied to installed provider/version/session availability.
Fix: adapter registry with typed configuration and validated discovery; explicit start/resume/send/status/cancel/history capabilities and lifecycle events. Separate transport (local process/SSH/HTTP/MCP) from provider protocol where useful. Return effective capabilities and health from the backend. Define conformance tests for each adapter, including unsupported operations and cancellation semantics.
Acceptance: a fake/new adapter registers without changing unrelated room/UI code; capability failures are reported consistently. See SEC-09/14.

## ARCH-07 — P1: Workflow drafts are not executable specifications yet
Evidence: workflow nodes include instructions, retry/timeout/evidence fields but Run is disabled. Decision nodes have no structured condition schema; parallel ports have no join semantics; no workflow run/attempt/approval models or immutable definition pinning exist.
Consequence: enabling Run directly from saved graphs would introduce ambiguous branching, uncontrolled loops and confusing human approval behavior, especially in rooms without humans.
Fix: separate draft validation from publish/executable validation. Specify typed inputs/outputs, decision expressions, joins, retry/error paths, bounded loops and human-absence policy. Runs must pin a workflow revision and retain step attempts, approvals and resumable state. Cross-room handoffs need explicit authorization, context contracts, destination workflow entry and correlated results; preserve task identity and immutable origin/handoff events.
Acceptance: a branching review loop survives restart, terminates at its configured bound and never silently skips required human approval. Do not enable Run before these contracts and ARCH-01 exist.

## ARCH-08 — P1: Frontend composition is fragile and asynchronous state is global
Evidence: app.js defines render/form; features.js repeatedly wraps them; workflows.js wraps render again. Mutable globals include roomId/state/view/workflow. Entire sections are replaced through innerHTML. API response handling renders global state; room switches have no request cancellation/version guard. An unsaved workflow is discarded on switching to another room's workflow through the roomId reset.
Consequence: load order controls behavior, new squad rows broke index-based provider rendering, slow responses can render the wrong room's state, and navigation can lose drafts/focus/scroll. Accessibility and interaction defects become difficult to isolate.
Fix: TypeScript modules with explicit components/services/store and keyed member identities. Capture room/context per request; cancel or ignore stale responses. Separate server state from per-room drafts and interaction state. Add dirty-navigation protection and persistent draft recovery. A framework is optional; global monkey-patching is the problem.
Acceptance: rapid room switches with reordered responses never cross-render data; drafts survive intended navigation; human/agent rows do not rely on array-position assumptions.

## ARCH-09 — P1: Shared memory scope lacks a real project model and provenance contract
Evidence: scoped makes both Project and Application memory visible across every room; no Project entity exists. Memory revisions hold title/body/scope but not structured author/source/decision status. Conflicts return 409, with no durable dispute or consensus workflow.
Consequence: the three advertised scopes do not yet have three distinct boundaries, and consensus/seniority cannot be implemented reliably from a rejected write alone.
Fix: explicit Project ownership and room association, typed memory revisions with author/source/confidence/provenance, durable conflict records and competing proposals. Record human deadlines, ranked decisions, equal-rank consensus and unresolved escalation as domain events. Use storage-enforced scope permissions.
Acceptance: Project memory stays within its project; both competing writes remain available for resolution; equal-rank disagreement never silently chooses a winner. See SEC-03/12.

## ARCH-10 — P2: Evidence is text, not a reproducible verification record
Evidence: Evidence supports kind/title/body/source/time, and Source is correctly labelled Human-provided. There are no run IDs, artifact digests, test invocation/environment or structured reviewer findings.
Consequence: reviewers cannot reproduce claimed checks or associate results with an exact code/task revision. The UI appropriately avoids claiming runner verification; preserve that honesty.
Fix: distinct human statements, runner observations and reviewer decisions; link each to task revision and attempt. Store artifact manifests/digests, code revision, check command/environment, exit/result and timestamps. Keep blobs outside the event payload with controlled access and retention; preserve provenance independently.
Acceptance: reviewers can identify exactly what was checked, by whom and against which revision; an agent's unsupported assertion never becomes a verified pass count.

## ARCH-11 — P2: Operations, migrations and observability are underdeveloped
Evidence: startup executes unversioned DDL and recreates a trigger; no graceful shutdown orchestration, job drain, readiness endpoint, structured correlation or durable job metrics is visible. Agent completion ignores several append errors. Native history fork embeds a developer-specific absolute directory.
Consequence: upgrades and shutdowns can strand work; failed result writes look like missing responses; fresh installations depend on workstation details.
Fix: numbered migrations run independently with migration credentials; configurable workspace/adapter paths; signal-aware graceful shutdown and lease handoff; readiness/health checks. Propagate/persist failed result commits and expose operator-visible job errors. Add structured IDs and metrics with sensitive-content redaction.
Acceptance: a clean installation works outside the original home path; interrupted writes surface failures; shutdown leaves recoverable job state. See SEC-05/14.

## ARCH-12 — P1: Tests and documentation do not define reliable product contracts
Evidence: current server tests cover checksums, room projection, challenge guidance and graph validation. No durable lifecycle/auth/concurrent command integration suite exists. README's early Not implemented paragraph contradicts later adapter/room descriptions; DEVELOPMENT_STATUS likewise contains superseded statements. Manual walkthroughs have been valuable but are not repeatable failure tests.
Consequence: regressions in the main orchestration paths escape checks, and developers cannot reliably distinguish active capabilities from planned guarantees.
Fix: reusable fake adapters, isolated PostgreSQL integration fixtures, deterministic clock/failure injection, restart/cancellation/conflict/idempotency tests, and focused UI tests for room switching/drafts/accessibility. Publish one current capability matrix and architecture decision records; move milestone history out of normative setup instructions. Add security regressions from SEC-01–14 as fixes land.
Acceptance: CI exercises failed provider runs, interrupted commits, concurrent actors and restart recovery without live accounts; docs match effective capabilities.

## Proposed target structure and sequence

Target: browser → HTTP/API layer → authenticated application services/policies → PostgreSQL domain state, read projections and audit/outbox → durable workers → adapter registry → isolated provider sessions. Independent protected audit anchoring remains outside the application's administrative control.

1. Identity/company/membership model and security boundary (ARCH-02/03; SEC-01–04).
2. Durable jobs, transactional commands, event versions and read projections (ARCH-01/04/05).
3. Adapter registry and explicit lifecycle contracts (ARCH-06).
4. Frontend modules, race-safe state and draft recovery (ARCH-08).
5. Executable workflow publication/runs, conflict resolution and cross-room handoffs (ARCH-07/09).
6. Structured evidence, operational readiness and continuously updated tests/docs (ARCH-10–12; start testing in step 1, not only at the end).

This sequence is incremental: preserve existing history and UI behavior through adapters/projections and migrate one domain at a time. Do not let additional visual features substitute for reliable execution and authorization contracts.

## Remediation progress — owner authentication boundary

SEC-01 partially addressed: every API except login now requires an expiring HttpOnly/SameSite=Strict owner session; per-session CSRF, logout revocation, bounded sessions and login rate limits are implemented. The randomly generated persistent owner access key is stored in `.data/owner-key` with owner-only permissions, never returned by HTTP. Startup restarts revoke sessions. Multi-actor identities, scoped authorization and immutable actor attribution remain open; do not expose owner credentials to agents. Same-OS-account compromise is outside this local credential boundary.

SEC-02 local boundary addressed: exact localhost Host allowlist, foreign Origin/cross-site browser rejection and numeric-loopback-only binding. Remote access is deliberately unavailable until authenticated TLS deployment is implemented. Cookies are not Secure because this deployment is local HTTP; remote deployment must require Secure cookies and TLS.

Verification: automated tests cover anonymous API denial, wrong credentials, login, CSRF denial, logout, expiry, Host/Origin rejection, rate limiting, key-file permissions/persistence and remote-bind refusal. Go race tests passed. No agents launched for verification.

RSA provisioning update: RSA-3072 PKCS#8 PEM download is authenticated and CSRF-protected during setup. Browser paste/file selection imports a non-extractable signing key and sends only an RSA-PSS/SHA-256 signature (32-byte salt) of a random, one-use, two-minute challenge. After verified key login, bootstrap credentials and the pending server private-key file are removed; public-key verification persists across restart. Challenges are bounded and login attempts remain rate-limited. Private PEM downloads are unencrypted and need protected storage/backups. File removal does not guarantee physical erasure. Recovery, key rotation and encrypted-PEM UX remain open. RSA is not claimed to be inherently stronger than the previous high-entropy secret. Tests cover provisioning, protected download, valid/invalid signatures, replay, expiry and public-only restart.

Invitation-only identity milestone: individual human RSA accounts with stable IDs, owner-controlled Developer/QA roles and shared-room access; no public signup. Server checks room authorization and owner-only operations, blocks disabled actors on each request, and grants each human/bot a private personal room. Owner personal-room reads are audited. Messages now capture actor ID/type for new human and assistant records; legacy history is not rewritten. Integration tests use a disposable schema and cover distinct same-name users, signed invitation login, access isolation, self-promotion denial, account disabling and owner access receipts. This is a first permission policy, not complete SEC-01/03 remediation: agent credentials, capability-scoped operations, project boundaries and all-event actor attribution remain pending.

## Execution isolation decision — 11 October 2026

SEC-04 remains open. Mandatory container isolation is now the documented execution contract, with no supported disable switch or direct-launch fallback. Adapter/launch comments explicitly identify current uncontained paths and explain source-fork bypass risks. This documentation does not implement confinement or close the finding. See docs/architecture.md#mandatory-container-isolation-design-contract.

## Mandatory worker implementation — 11 October 2026

Managed Codex/Claude turns now use Linux/rootless Docker with no direct fallback, non-root/read-only/capability restrictions, private membership volumes, provider-only CONNECT egress and bounded resources/output/cleanup. Direct SSH and non-Linux dispatch are blocked. Legacy host-session resume is blocked; client profile payloads cannot set execution provenance, and connection reassignment clears old bindings and is denied while active. Mocked regression checks pass. Live Docker/provider behavior, configured host mounts, remote workers, crash reconciliation and independent incident shutdown remain unverified or unfinished. SEC-04/08/09/10 are mitigated in these paths, not universally closed. See docs/container-isolation.md.


## Legacy development admission barrier — 14 October 2026

Updated controller dispatch, worker request validation and local preparation now
refuse the known credential-bearing native development path before credentials or
Docker are accessed. No owner/flag/paid bypass is supported. Status identifies the
missing separated bridge. Tests cover refusal and retained chat validation. This
closes new legacy development admission in updated source, not all provider native
capabilities or a working replacement. Existing/deployed old worker binaries/jobs,
pinned Linux bridge/native-tool verification and production separation remain OPEN.

## Reviewed engine preparation — 14 October 2026

Dedicated Codex build preparation now verifies the replacement bridge policy as
well as upstream source before editing the trusted build checkout. Altered policy
and unexpected dependency lockfiles fail without changing the source. The release
lockfile requires a pinned local-workspace version normalization only; external
versions/checksums remain unchanged. Eight offline synthetic tests cover drift,
symlink refusal, preservation of rejected input and successful one-time preparation.
This is mandatory free/core build policy, with no opt-out or paid bypass.

The isolated Linux compile/test verification is separate from CommonRoom's fast
checks. No provider login/model turn is performed by these tests. Production
engine artifact verification, broker audit transport and paired supervision remain
OPEN; source preparation does not make the credentials-separated runtime usable.
See docker/engine/README.md and docs/provider-tool-separation.md.

Pinned Linux/amd64 compilation and three synthetic router tests now pass. A
development CLI binary also built and passed network-disabled version/schema
metadata checks, with no provider turn. This verifies the selected factory's
construction and packaging only; it does not close production SEC-04 separation.


## Paired lifecycle component — 15 October 2026

Version-3 worker journals now retain both engine/tool identities, creation intents
and absence observations. Pair cancellation has an explicit versioned command,
refuses legacy acknowledgements, verifies role/job/repository bindings, cleans by
immutable IDs and requires released supervisor ownership before settlement.
Partial cleanup, name replacement, unknown job containers and unavailable inventory
remain unresolved. Tests use synthetic temporary journals and mocked Docker;
no live runtime verification was performed. See [paired lifecycle](docs/worker-pairs.md).
This closes no claim of complete credential-separated execution: actual launch,
bootstrap, aggregate runtime verification, authenticated audit transport,
callback-only provider drivers and artifact delivery remain open. The mandatory
legacy development admission barrier stays enforced.

Controller version-3 registration/recovery now retains immutable execution and
repository binding, requires real initiating-human/company provenance and current
repository policy, rejects legacy snapshot/deadline substitution and uses the
paired cancellation protocol. Automatic lifecycle evidence retains the initiating
human. The legacy launcher refuses version-3 registrations. Owner recovery is
mock-tested with diagnostic flags off/on, mismatches and failed shutdown; no live
paired execution is enabled or claimed.

### Architecture continuation — 15 October 2026

Paired lifecycle now owns durable creation-before-start ordering and ID-bound
aggregate reported-resource checks. Managed job lifecycle concentrates controller
submission and recovery recording. Shared immutable-container shutdown mechanics
do not merge tool absence with pair/lease proof. These mandatory safeguards have
no off switch. The launch adapter remains internal without a production adapter;
this does not close P0 credential separation or authorize development admission.
Live rootless runtime/provider enforcement remains unverified. See
[execution progress](docs/p0-execution-progress-2026-10-15.md).

### Constrained tool-side launch continuation — 15 October 2026

The internal tool runtime derives fixed Docker-create arguments, refuses named
container adoption and missing/unsafe prepared volumes, verifies stopped policy,
rechecks before immutable-ID startup and verifies running policy afterward.
Original job lifetime and no-replay state survive fresh call contexts. There is
no provider, host-launch or optional-security fallback. This does not establish
fresh volume contents, bootstrap provenance, engine driver safety or production
pair integration. Those P0 gates and live verification remain open; development
admission is still blocked. See [runtime contract](docs/tool-container-runtime.md).

### Workspace bootstrap continuation — 15 October 2026

Streamed bootstrap now checks the approved bundle hash before controlled Git,
requires an empty isolated workspace, bounds accepted tree size/file types and
suppresses hooks/configuration/checkout transforms. Job-bound receipts and fresh
runtime inspection gate file tools; pair startup requires bootstrap before the
engine. Partial or uncertain results cannot be replayed. No optional flag bypass
exists. Fresh-volume allocation/cleanup, real Git/image/rootless verification,
production adapter assembly and authenticated audit/engine integration remain open
P0 work; development admission stays blocked. See [bootstrap contract](docs/workspace-bootstrap.md).
