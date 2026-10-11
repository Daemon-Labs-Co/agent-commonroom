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
| POST /api/provider-auth | Check or launch local Claude authentication. | Owner and CSRF; registered local Claude only. |

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
| create | action | Create a fresh ChatGPT-authenticated Codex participant/connection; requires valid local login. |
| attach | threadId, mode, name, optional additionalContext/contextSource | Copy (`fork`) or resume a saved local Codex conversation. |
| list-conversations | optional search, cursor | List supported local saved Codex conversations. |
| read-context | threadId | Read reference context from a supported saved conversation. |
| send | participantId, text | Request one available supported participant's turn. |
| broadcast | text | Request available supported room participants; report skipped members. |
| remove | participantId | Remove room membership and request cancellation; retain history. |

send/broadcast return HTTP 202 with queued count and skipped reasons. This confirms dispatch acceptance, not successful completion or durable queuing. Later state includes the reply, or Blocked status and an agent.error event. No response-token streaming API exists.

provider-auth accepts `{"action":"status","connectionId":"<id>"}` or action login. Status returns loggedIn. Login returns launched and opens the official local Claude flow on macOS; launched is not logged-in confirmation. Other platforms do not support that login launcher.

## Failure handling and limits

Expect 400 for malformed/unsupported operations, 401 for missing/expired sessions, 403 for access/token/host checks, 404 for missing objects where the handler uses it, 409 for conflicts or unavailable participants, 429 for login/challenge limits, and 5xx for storage/provider failures. Exact statuses vary by handler. Bodies are usually plain text rather than a stable structured error schema.

There is no API versioning, OpenAPI schema, idempotency-key support, durable workflow execution API, or complete per-actor capability model. Do not retry arbitrary mutations automatically: an interrupted multi-record operation may already have applied part of its work. Read state and review history first.

### Applying workflow templates

`POST /api/events?room=<id>` accepts top-level `templateId` for a `workflow.revised` request. The templates flag must be enabled, the ID must be known, and the caller must be an owner or authorized Developer with room access. A new payload without `id` creates a workflow; the response includes `workflowId`. Ordinary manual saves remain available when the catalog flag is off. See [template usage and configuration](concepts.md#workflow-templates).

### Emergency agent safety

`POST /api/agent-safety?room=<id>` requires an authenticated owner session and `X-CSRF-Token`. The body is `{ "action": "stop", "scope": "application", "reason": "optional explanation" }`. Actions are `stop`/`resume`; scopes are `application`/`room`. Reasons are limited to 500 characters. Scope `room` requires an existing room; application stop does not depend on a database read.

Success returns `action`, `scope`, `recorded: true`, and `remoteTerminationConfirmed: false`. A resume conflicting with an integrity incident or global stop returns 409. Audit failure returns 503 and leaves local stop latches active. Event kinds `agents.stop` and `agents.resume` preserve the decision. State exposes `agentsStopped` and `agentRoomStopped`; internal stopped-room IDs are not included in state JSON. The backend remains active even when its optional UI flag is disabled. See [cutoff limitations](concepts.md#emergency-agent-cutoff).

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

Only `humanProfileImages` and `agentSafetyControls` are supported, and an explicit boolean is required. Unsupported names/missing values return 400; nonowners return 403. `/api/state` includes effective `featureFlags`. Photo default is true. The emergency flag overrides `COMMONROOM_AGENT_SAFETY_UI`; hiding it never resumes dispatch. `/agent-safety.js` serves no content when disabled, and returns 503 if its saved settings cannot be loaded. Protected emergency APIs remain enabled. Environment-only flags continue to follow the configuration table in the README.

## Connection readiness

See [Testing connections](connection-testing.md) for `POST /api/connections/test`, its response and owner/CSRF/feature enforcement. Checks send only a fixed, tool-disabled provider diagnostic prompt, without room history.

## Company organization chart

`GET /api/company/organization` requires an authenticated owner and enabled `companyOrgChart`; returns `company`, sanitized `actors`, room-profile metadata and `organization: {revision, nodes}`. No private keys, connection endpoints, conversation text or agent context/instructions are returned. Invited humans receive 403 and company graph nodes are removed from their `/api/state` projection. Responses are not cached; unsupported methods return 405.

Save through `POST /api/events` with `kind: "company.organization.updated"` and `payload: {revision, nodes: [{actorId, reportsTo, x, y}]}`. A blank `reportsTo` means no manager. The supplied revision must match the current revision (409 otherwise); successful append increments it. Owner authentication and CSRF are required. Disabled features return 403; unknown/duplicate actors, missing manager nodes, self-reporting, cycles, nonfinite/out-of-bounds coordinates and graphs over 500 nodes return 400. This uses the existing serialized event mutation path and existing 64 KiB request limit, not a new database table or external service. Reporting relationships cannot change account roles, room permissions or work acceptance.

Connection tests optionally accept `Accept: application/x-ndjson`. Each newline-delimited message is `{ "type": "stage", "check": { "name": "Worker runtime", "status": "running", "detail": "…" } }`, `{ "type": "result", "result": <existing test response> }`, or `{ "type": "error", "message": "safe error" }`. Authorization failures use normal HTTP errors before streaming. Clients must require a terminal result; an HTTP 200 or passed intermediate stage is insufficient. See [live connection progress](connection-testing.md#live-test-progress).

`POST /api/provider-auth` also accepts `{ "action": "logout", "connectionId": "REGISTERED_ID" }`. Owner/CSRF, connection-sign-in feature flag, worker allowlist and incident checks apply. Active turns using that connection return 409. After audited intent, the worker clears only its scoped saved credential, returning `{ "signedOut": true }`; completion is audited before success. This does not revoke provider tokens or purge agent session volumes.
