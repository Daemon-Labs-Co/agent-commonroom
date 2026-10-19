# HTTP API overview

[Documentation index](README.md)

This describes the current internal browser API, not a versioned public SDK contract. Route implementations and tests remain authoritative. There is no API-key or bot service-account authentication yet.

## Authentication and requests

The server accepts configured localhost/loopback Host values and same-origin requests. The browser port must match the configured listening port, including over an SSH tunnel. The HTTP origin is local loopback; remote public hosting is unsupported.

Routine human login signs a server challenge locally with RSA-PSS/SHA-256 and sends actorId, challengeId, and a Base64 signature. Keys use RSA-3072; private keys do not travel during routine signature login. First-owner bootstrap login is a separate setup-code path. Invitations contain an individual actor ID and private key generated in the owner's browser.

Authenticated requests use the `commonroom_session` HttpOnly, SameSite=Strict cookie. GET /api/session returns the current session's CSRF token. Protected mutations send `X-CSRF-Token`. Challenges and login are pre-session exceptions; do not send a downloaded private key as an API login field.

Pass `?room=<room-id>` to room-sensitive state/events/agent operations. Omitting room defaults to `engineering`. Members must choose a room they can access; the UI starts them in their personal room. The owner can access all rooms. Successful JSON responses and plain-text error responses should be handled separately.

## Routes

| Method and route | Purpose | Access notes |
| --- | --- | --- |
| POST /api/challenge | Create a one-use, two-minute sign-in challenge. | Before login; bounded pending challenges. |
| POST /api/login | First-install setup-code or signed-key login. | Before login; account/signature checks and rate limiting. |
| POST /api/owner-key/download | Download pending owner PEM during setup. | Owner, session, CSRF, pending-setup checks. |
| POST /api/logout | Revoke the current session. | Session and CSRF. |
| GET /api/session | Session actor/type, CSRF token, setup status. | Signed-in enabled account. |
| GET /api/state | Replayed state filtered for room/account. | Room access required; public-key metadata removed. |
| GET /api/events | Latest 200 event metadata entries. | Owner; returns sequence, kind, time, hash, not full payloads. |
| POST /api/events | Validate and append a supported domain operation. | CSRF, room access, operation-specific permissions. |
| GET /api/integrity | Verify current chain and record a detected incident. | Owner; this GET can append an incident record. |
| POST /api/members | Create/edit an invited human account. | Owner and CSRF. |
| POST /api/agents | Create/attach/remove participants or request turns/history. | Currently owner and CSRF; room access. |
| POST /api/provider-auth | Inspect saved worker credentials, launch worker sign-in or sign out. | Owner and CSRF; registered approved Codex/Claude worker connection. |
| POST /api/connections/test | Run a fixed isolated readiness check. | Owner, CSRF and enabled connection testing. |
| POST /api/task-dispatch | Work, review or reconcile a task. | Owner, room/revision checks and taskDispatch. |
| GET/POST /api/workflow-runs | Inspect or control sequential runs. | Owner; mutations require CSRF and current revision. |
| POST /api/agent-safety | Stop or explicitly resume agent admission. | Owner and CSRF; core enforcement cannot be disabled. |
| POST /api/agent-recovery | Check, resume or reset an idle session. | Owner, CSRF, feature and verified worker quiescence. |
| GET /api/job-diagnostics | Read bounded job/worker observations. | Owner and jobDiagnostics. |
| POST /api/job-recovery | Cancel or reconcile a managed job. | Owner, CSRF and verified identity/cleanup. |
| GET /api/company/organization | Read reporting chart metadata. | Owner and companyOrgChart. |

## Agent session recovery

Owner-only, authenticated, CSRF-protected `POST /api/agent-recovery?room=ROOM`
accepts `{"action":"check|resume|fresh","participantId":"ID"}`. Default-on
`agentRecovery` is enforced server-side. The participant must belong to the room.
Check returns `{idle,recovered:false,detail}` and changes no history. Mutations
recheck shared-worker quiescence at action time, then atomically append one
`agent.recovered` event containing the resulting participant snapshot, action and
authenticated actor. Replay applies the Idle state and optional session reset.
There is no model launch, task replay or audit rewrite. Busy worker returns
`idle:false`; active controller work/stops/incidents/invalid resume return 409;
unknown worker state returns 503; disabled flag/non-owner/invalid CSRF return 403.

## Background work telemetry

With `agentBackgroundWork` on (default), room-authorized participant state includes
up to 64 `background` observations: `turnId`, optional `jobId`, trusted stage/label,
room/participant IDs and `observedAt`. These are controller observation times, not
worker heartbeats or exact command timestamps. Flag off omits the property and
blocks telemetry collection; existing audit history remains. Only the owner may
toggle the flag. The internal `agent.background` event cannot be submitted through
public event ingestion. Stop all reuses authenticated owner-only, CSRF-protected
`POST /api/agent-safety`, independent of cosmetic safety button visibility.

## Execution capacity and long-running status

Agent chat dispatch returns HTTP 429 when its entire eligible target batch would
exceed eight active model turns overall or four in the requested room. That
rejection launches none of the batch and appends no chat message. Workflows use
the same admission check; overload becomes an explicit failed run for owner
inspection/resume. Limits are single-controller safeguards, not durable queues.

When default-on `agentStatusDetails` is enabled, Working participants may include
`statusDetails.elapsedSeconds` and `statusDetails.longRunning` (true at 120
seconds). Values derive from recorded dispatch time, not worker telemetry.
Disabled delivery omits the entire details object; room access protections apply.

## Event mutations

POST /api/events accepts an envelope:

```json
{"kind":"message.created","payload":{"text":"Discuss the proposed design."}}
```

For example, send it to `/api/events?room=engineering` with the cookie and CSRF header. The server assigns human identity, author, room, timestamp, and message ID. Posting a room message through this operation does not itself guarantee agent invocation; the UI uses agent actions for directed turns and broadcasts.

The accepted mutation families include company/profile, rooms, connections, participants, messages, memory revisions, tasks, evidence, and workflow revisions. They are not unrestricted event injection: unsupported kinds are rejected and each kind has validation rules. Responses normally return `{"committed":true}`; refresh state to obtain projected values and new revisions.

Non-owner mutations are currently limited to own-name changes, messages, Room-scope memory, and task edits excluding acceptance. Developers may also save workflows; QA members may not. Agent control, connection/profile administration, evidence creation, audit listing, integrity, and account administration currently require the owner. This conservative policy is incomplete and may change.

For an existing revisioned object, submit its current revision. Memory revisions increment on successful edits; stale revisions return 409. Other payload limits and revision rules depend on the handler—this overview is not a complete schema catalog.

## Agent actions

| Action | Main fields | Behavior |
| --- | --- | --- |
| create | action | Legacy participant metadata helper; checks a local ChatGPT-authenticated Codex login. It does not configure an isolated worker. |
| attach | threadId, mode, name, optional additionalContext/contextSource | Legacy host-history copy (`fork`) or resume; does not convert it into an isolated worker session. |
| list-conversations | optional search, cursor | List supported local saved Codex conversations. |
| read-context | threadId | Read reference context from a supported saved conversation. |
| send | participantId, text | Request one available supported participant's turn. |
| broadcast | text | Request available supported room participants; report skipped members. |
| remove | participantId | Remove room membership and request cancellation; retain history. |

send/broadcast return HTTP 202 with queued count and skipped reasons. This confirms dispatch acceptance, not successful completion or durable queuing. Later state includes the reply, or Blocked status and an agent.error event. No response-token streaming API exists.

`POST /api/provider-auth` accepts `{"action":"status|login|logout","connectionId":"ID"}` for a registered approved worker connection. Status returns `credentialPresent`, `recoveryRequired` and `provider`; uncertain credential custody forces `credentialPresent:false`. A saved file is not proof of valid provider authentication. See [credential recovery](worker-credential-recovery.md). Login returns `launched:true` after opening a macOS Terminal for the worker sign-in command. Other controllers use the documented manual command. Logout returns `signedOut:true` only after scoped local credential removal and completion audit. See [connection sign-in](connection-sign-in.md).

## Failure handling and limits

Expect 400 for malformed/unsupported operations, 401 for missing/expired sessions, 403 for access/token/host checks, 404 for missing objects where the handler uses it, 409 for conflicts or unavailable participants, 429 for login/challenge limits, and 5xx for storage/provider failures. Exact statuses vary by handler. Bodies are usually plain text rather than a stable structured error schema.

There is no versioned public SDK or complete OpenAPI schema. Selected compound operations support `Idempotency-Key` and revision checks; see [atomic operations](atomic-operations.md) for supported routes and exact replay rules. A receipt proves prior admission, not current completion. Do not retry an uncertain provider effect merely because its HTTP response was lost.

### Applying workflow templates

`POST /api/events?room=<id>` accepts top-level `templateId` for a `workflow.revised` request. The templates flag must be enabled, the ID must be known, and the caller must be an owner or authorized Developer with room access. A new payload without `id` creates a workflow; the response includes `workflowId`. Ordinary manual saves remain available when the catalog flag is off. See [template usage and configuration](workflow-templates.md).

### Emergency agent safety

`POST /api/agent-safety?room=<id>` requires an authenticated owner session and `X-CSRF-Token`. The body is `{ "action": "stop", "scope": "application", "reason": "optional explanation" }`. Actions are `stop`/`resume`; scopes are `application`/`room`. Reasons are limited to 500 characters. Scope `room` requires an existing room; application stop does not depend on a database read.

Success returns `action`, `scope`, `recorded: true`, and `remoteTerminationConfirmed: false`. A resume conflicting with an integrity incident or global stop returns 409. Audit failure returns 503, cancels managed work and attempts to persist the private integrity-stop marker. Owner integrity recovery is required; an unwritable marker is an operator fault. Event kinds `agents.stop` and `agents.resume` preserve the decision. State exposes `agentsStopped` and `agentRoomStopped`; internal stopped-room IDs are not included in state JSON. The backend remains active even when its optional UI flag is disabled. See [cutoff limitations](agent-activity.md#emergency-agent-cutoff).

### Profile photos and saved feature flags

Authenticated, CSRF-protected `POST /api/events` supports:

```json
{"kind":"actor.image.updated","payload":{"image":"data:image/jpeg;base64,..."}}
```

An empty `image` removes the current photo. The server binds the event's account ID to the authenticated human, ignoring submitted IDs and role fields. Input accepts JPEG/PNG data URLs only, at most 32 KB decoded and 256 × 256 pixels; accepted data is re-encoded as JPEG at quality 80, limited again to 32 KB. The normal 64 KB event request limit also applies. Missing/invalid image returns 400; disabled uploads return 403. Clearing one's photo remains allowed with the flag off. This uses immutable event storage, not a public upload directory. `actors[].image` holds the current normalized data URL. CSP allows same-origin and data images, not arbitrary external image servers.

Owner-only saved flags use:

```json
{"kind":"feature.flag.updated","payload":{"name":"humanProfileImages","enabled":false}}
```

An explicit boolean and a recognized server allowlisted flag name are required. The allowlist includes workflow, security, connection and presentation features; use the relevant guide for defaults and disabled behavior. Unsupported names/missing values return 400; nonowners return 403. `/api/state` includes effective `featureFlags`. Photo default is true. The emergency flag overrides `COMMONROOM_AGENT_SAFETY_UI`; hiding it never resumes dispatch. `/agent-safety.js` serves no content when disabled, and returns 503 if its saved settings cannot be loaded. Protected emergency APIs remain enabled. Environment-only controls require restart; see [operations](operations.md) and the relevant feature guide.

## Connection readiness

See [Testing connections](connection-testing.md) for `POST /api/connections/test`, its response and owner/CSRF/feature enforcement. Checks send only a fixed, tool-disabled provider diagnostic prompt, without room history.

## Company organization chart

`GET /api/company/organization` requires an authenticated owner and enabled `companyOrgChart`; returns `company`, sanitized `actors`, room-profile metadata and `organization: {revision, nodes}`. No private keys, connection endpoints, conversation text or agent context/instructions are returned. Invited humans receive 403 and company graph nodes are removed from their `/api/state` projection. Responses are not cached; unsupported methods return 405.

Save through `POST /api/events` with `kind: "company.organization.updated"` and `payload: {revision, nodes: [{actorId, reportsTo, x, y}]}`. A blank `reportsTo` means no manager. The supplied revision must match the current revision (409 otherwise); successful append increments it. Owner authentication and CSRF are required. Disabled features return 403; unknown/duplicate actors, missing manager nodes, self-reporting, cycles, nonfinite/out-of-bounds coordinates and graphs over 500 nodes return 400. This uses the existing serialized event mutation path and existing 64 KiB request limit, not a new database table or external service. Reporting relationships cannot change account roles, room permissions or work acceptance.

Connection tests optionally accept `Accept: application/x-ndjson`. Each newline-delimited message is `{ "type": "stage", "check": { "name": "Worker runtime", "status": "running", "detail": "…" } }`, `{ "type": "result", "result": <existing test response> }`, or `{ "type": "error", "message": "safe error" }`. Authorization failures use normal HTTP errors before streaming. Clients must require a terminal result; an HTTP 200 or passed intermediate stage is insufficient. See [live connection progress](connection-testing.md#live-test-progress).

`POST /api/provider-auth` also accepts `{ "action": "logout", "connectionId": "REGISTERED_ID" }`. Owner/CSRF, connection-sign-in feature flag, worker allowlist and incident checks apply. Active turns using that connection return 409. After audited intent, the worker clears only its scoped saved credential, returning `{ "signedOut": true }`; completion is audited before success. This does not revoke provider tokens or purge agent session volumes.

### Paged event history

Owner-authenticated `GET /api/events?page=true` returns `{events,nextBefore}` with at most 100 summaries. Pass the previous `nextBefore` as `before` for the next older page. Optional `from`/`until` timestamps use RFC3339 and an inclusive/exclusive date window; malformed cursors/ranges return 400. The server uses parameterized keyset queries and a fixed 101-record lookahead. Raw payloads remain private. The legacy unpaged endpoint still returns its latest-200 array. Reads never alter checksums, history, retention or authorization.

Cosmetic flag events also accept owner-only `roomViewportFit` and `featureFlagSwitches`, both defaulting on. Their presentation modules are gated through `/room-viewport.js` and `/feature-switches.js`; disabled modules return 204 with no-store caching. Disabling these flags does not relax room permissions or security enforcement. See [room viewport and feature switches](interface.md#room-viewport-and-feature-switches) for usage and defaults.

## Workflow execution

Owner-only `GET /api/workflow-runs?room=ROOM` lists room run snapshots, even while execution is disabled. A previously running snapshot without a live driver is recorded as interrupted. `POST` also requires the session CSRF token and trusted Host/Origin.

- Start: `{ "action":"start", "workflowId":"ID", "revision":1, "input":"Approved task" }` returns HTTP 202 and the pinned run.
- Advance evidence/approval: `{ "action":"advance", "runId":"ID", "revision":N, "port":"next", "evidence":"Artifact references and decision rationale" }`.
- Decision: same advance contract with `port` equal to `yes` or `no`.
- Human rework: `{ "action":"reject", "runId":"ID", "revision":N, "evidence":"Requested changes" }` at an approval gate returns to the latest implementation node.
- Recover: `{ "action":"resume", "runId":"ID", "revision":N }`; only failed/interrupted runs can recover, within visit limits, after inspecting possible partial side effects.
- Stop: `{ "action":"cancel", "runId":"ID", "revision":N }`; remains available with the feature disabled.

All mutations require the current run/graph revision. Foreign-room IDs return 404, stale/busy/unsafe state 409, unsupported graphs 400, disabled execution/non-owner/invalid CSRF 403. `workflow.run.updated` events cannot be supplied through the generic events endpoint. The workflowExecution flag is changed through owner-only feature.flag.updated; absent means off. See [complete run schema, lifecycle, bounds and security contract](workflow-execution.md).

## Agent status explanations

With owner-controlled `agentStatusDetails` enabled (default on), room-scoped state
participants contain `statusDetails: {activity, reason, next}`. These fields are
server-derived safe summaries, not editable profile properties. Disabled state
omits them. Internal `agent.activity` events identify generic chat activity or a
workflow/step title; `agent.error` events supply the recorded cause. Both are
rejected through the public generic events endpoint. Ordinary profile edits retain
current diagnostic history; a new non-blocked session clears the prior error.

`cmd/server/agent_status.go` classifies failures using allowlisted wording, bounds
activity labels and prioritizes incident/emergency stops. `web/agent-status.js`
uses text nodes, keyboard focus and an unclipped fixed-position tooltip. Mock-only
Go tests and synthetic DOM tests cover legacy fallbacks, both flag states, stale
errors, room scope, safe rendering and keyboard/pointer cleanup.

## Human attention requests

With `humanAttention` enabled (default on), agent message state includes
`humanRequest: "explicit" | "suggested"` and, after acknowledgment,
`attentionAcknowledgedBy: ACTOR_ID`. Rooms include `pendingHumanRequests` when
nonzero. Projection occurs before existing authorized-room filtering; no hidden
room messages or counts are returned. Original messages are never edited.

Authenticated room-authorized humans POST through `/api/events`:
`{"kind":"message.attention.acknowledged","payload":{"messageId":"ID"}}`.
The server supplies actor/room identity, validates an agent request in that room,
and makes repeated acknowledgment idempotent. Missing/foreign/human messages
return 404; agent accounts, disabled feature or invalid CSRF return 403. Generic
human message creation clears any forged attention fields and agent identity.
Only the owner can change the feature flag.

`cmd/server/human_attention.go` provides default-on flag handling, explicit protocol
and conservative phrase classification, immutable acknowledgment projection and
room counts. `web/human-attention.js` renders text-only badges/banner/counts, moves
to a request and submits acknowledgment. It runs before message-order/viewport
wrappers so existing scroll restoration accounts for its layout. Tests mock all
database operations and use synthetic DOM fixtures for detection, idempotency,
identity/room boundaries, both flag states, safe display and error handling.

## Refined History presentation

When `historyRefinement` is enabled (default on), owner-only GET `/api/events`
responses add `presentation` to each event: `category`, `tone` (info/warning/error/
success), `routine`, and optional recorded `actor`, `actorId`, `subject`, `roomId`,
`outcome`, safe `explanation` and `next`. This is an allowlisted projection; raw
payloads are never returned. Original sequence/kind/at/hash/summary fields and
100-record keyset paging remain unchanged. Disabled responses omit presentation.

`cmd/server/history_presentation.go` classifies events and reads only the latest
flag record through a parameterized, one-record query. It does not load the full
event stream to enrich a page. Flag read failures fail the response rather than
assuming enablement. `web/history-presentation.js` defines safe cards, daily groups,
filters and loaded-count statistics; `web/history.js` retains async paging, stale
response protection and per-visit filter state. Tests use mock queries and synthetic
UI events for both flag states, permissions, privacy, filtering, day grouping,
noise visibility and pagination compatibility.

## Mandatory security controls

See [security hardening](security-hardening.md) and [restricted database setup](database-security.md) for required migration/runtime role separation, integrity shutdown/recovery, identity and request/context budgets. These boundaries have no off switch.

`POST /api/integrity` requires owner authentication, CSRF and `{ "reason": "..." }` (nonempty, at most 500 characters). Recovery verifies the chain and worker idle, appends approval, and returns `{ "recovered": true, "jobsRestarted": false }`. Failures leave execution stopped. `GET /api/integrity` reports `valid`, `records`, `metadataProtectedRecords`, `legacyMetadataRecords` and `executionBlocked`; legacy records are never silently rewritten. Mutation rate limits return 429 with `Retry-After: 1`; oversized assembled chat context returns 413 before dispatch.

`POST /api/agent-safety` resume now requires a freshly valid audit chain, no active controller turns and verified shared-worker idle. A 409/503 leaves the stop unchanged. Startup with active/unverified remote worker state records an application stop rather than admitting new work.

## Free durable job safety

See [worker lifecycle, shutdown verification and diagnostics](worker-jobs.md) for mandatory protocol-v2 execution, durable intent, crash reconciliation, owner diagnostics and the `jobDiagnostics` flag (default on). Shutdown uncertainty never counts as success; no task is replayed automatically.

- `GET /api/job-diagnostics`: owner-only and flag-gated; current room by default, or `scope=application` for all tracked jobs. Optional `participantId` filters a member. Returns bounded safe job/worker observations and verification status; no raw output or credentials.
- `POST /api/job-recovery`: owner+CSRF, `{ "action": "cancel|reconcile", "jobId": "...", "reason": "..." }`. Optional `scope=application` addresses an archived-room job. Unknown or mismatched bindings, active replacement turns and unverified shutdown are refused. Successful responses have `stopped:true` and `jobsRestarted:false`. Recovery remains enforced/available independently of the diagnostics flag.

## Disclosure and operation safety

See [context disclosure contracts](context-disclosure.md) for room/provider and memory export controls, and [atomic operations](atomic-operations.md) for required edit revisions, `Idempotency-Key`, response replay and single-controller limits.

## Atomic task dispatch

See [Task dispatch](task-dispatch.md#api-and-technical-contract) for the owner-only
`POST /api/task-dispatch` work/review/reconcile contract, retry receipts, permission
checks and default-on `taskDispatch` flag. The browser no longer sends an agent
message and task-status mutation as separate operations.

Connection-test provider turns use controller-journalled run-v2 and require durable
shutdown proof before reporting Working. See [verification lifecycle](connection-testing.md#controller-owned-verification-lifecycle).

Owner job diagnostics also includes application-wide `turnBudget` usage and reset
time. Admission rejects exhausted daily launch budgets; mandatory registration
rechecks journal usage atomically before launch. See [execution budgets](execution-budgets.md).


### Paired worker recovery infrastructure

The existing owner-only job diagnostics/recovery routes can represent explicitly
registered version-3 historical pairs and version-4 workspace-bearing jobs. Their durable execution/repository binding
cannot change across lifecycle updates. Reconciliation requires matching worker
version, membership, repository, original deadline and released ownership.
The controller sends the explicit paired cancellation protocol and records failed
shutdown as unresolved; diagnostic flag changes never bypass these checks.
No paired launch API is enabled. Existing version-2 chat jobs keep their original
protocol; neither version can be silently substituted for the other. See
[paired worker lifecycle](worker-pairs.md) for the implementation boundary.

Version-4 recovery uses `cancel-pair-workspace-v1` and requires workspace-removal evidence as well as container/lease proof. HTTP cancellation/reconciliation does not grant workspace discard authority: a present workspace remains retained/unresolved until the separate owner-reviewed disposal workflow succeeds. Older protocols cannot settle these records. See [workspace volumes](workspace-volumes.md).

### Owner-reviewed workspace disposal

`GET/POST /api/workspace-disposal` is enabled-human-owner only. POST requires CSRF.
The `workspaceDisposal` flag defaults false and blocks new proposals/approvals
server-side. Pending revocation and history remain available when off. Actions are
`propose` (`jobId`), `approve` (`proposalId`, exact job-ID `confirm`) and `revoke`
(`proposalId`). The server enforces a full minute of review and one-hour expiry.
Claimed deletion is one-shot; unknown outcomes cannot be retried. The proposal
includes `storageReservationId` only for an exact prebound workspace charge.
Verified removal releases that charge atomically with the outcome and container
slots; unbound legacy charges and separate artifacts stay charged. See the [full contract](workspace-disposal.md).

### Owner idle shutdown policy

Owner-only `agents.idle_policy.updated` events accept exactly integer `minutes`
(1–60) and the current `revision` (initially 0). CSRF is required; stale forms
return 409 and successful saves increment the revision. `idleAgentShutdown` is
a default-on feature flag for new paired allocations. Existing jobs keep their
immutable policy snapshot. These settings do not enable production dispatch.
See [idle policy and worker handoff](worker-idle-shutdown.md#controller-and-api-contract).

### Idle-stop notice visibility

Owner-only `feature.flag.updated` supports default-on `idleStopNotices`. It
controls derived room notices immediately; audit recording and existing job
policy remain enforced when off. `worker.idle_stop.announced` is an internal
trusted callback record, rejected by the public events API. The record is an
intention, never container absence, source disposal or storage-credit proof.
See [announcement contract](worker-idle-shutdown.md#idle-stop-announcements).
