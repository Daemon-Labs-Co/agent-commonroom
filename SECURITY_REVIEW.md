# Agent CommonRoom red-team review — 10 October 2026

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
