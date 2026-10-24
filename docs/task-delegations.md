# Bounded task delegation

This increment implements authenticated assignment proposals, human-reviewed
reassignment, and enforcement of the inherited task/repository/deadline boundary.
It does not grant a child extra tools, credentials or a longer lease. It permits
one handoff only; further subdelegation refuses. General delegation trees and
new resource approvals remain future work.

## Availability and setup

`taskDelegation` defaults off. The owner enables it in **Settings → Bounded task
delegation**, without restart. It is server-enforced, owner-only and audited.
The Tasks view lists retained proposals and provides explicit review/rejection.
Disabling refuses new proposals, approvals and delegated execution, invalidates
old proposals across re-enable, and cancels locally tracked delegated turns.
It preserves history and independent human work. Cleanup and human recovery
remain available under their existing rules, with fenced shutdown proof required.

**Provider integration remains incomplete.** Current model adapters do not emit
the new typed delegation requests automatically. No model prose or generic HTTP
request is interpreted as one. The operation snapshot is supervisor evidence,
not a permission request from the model. The authenticated supervisor API, controller
handling, review UI and execution revalidation are implemented and tested with
synthetic transports. A provider adapter must explicitly forward the typed
request through its credential-separated supervisor, and its live path must be
verified before autonomous delegation is advertised. Mandatory development
execution barriers and outstanding P0 release gates remain intact.

## Human workflow

1. Sign in with the owner key and open a room. Define a task, scope and acceptance
   criteria; assign it and explicitly dispatch through the approved worker path.
2. An admitted paired worker for that task can propose a different eligible room
   participant. Its identity comes from its authenticated job, not request fields.
3. Inspect **Tasks → Delegation proposals**. Review the exact task, repository,
   source, proposed delegate and deadline. A proposal changes no assignment.
4. The original worker must have verified shutdown and a recorded task outcome
   before approval. Pending or uncertain execution is never hidden by reassignment.
5. Approve to reassign that same task, or reject to retain the prior assignment.
   Approval never starts the delegate. The owner separately dispatches the task.
6. The delegate must satisfy its own connection/disclosure/worker permissions.
   Only the original repository and task content are permitted, with no extension
   past the original parent's deadline. The owner reviews outcomes separately;
   agents do not accept work automatically.

The parent job has a short bounded lease. A handoff can expire while waiting for
human review; that is a refusal, not automatic renewal. Changed task scope,
participant/connection revisions, removed membership, revoked coordinator role,
feature off/on, expired authority or safety stops refuse affected work. To create
new independent work, the owner must explicitly review and create a new task;
do not strip delegation history or assume a proposal authorizes broader resources.
Changing delegated title/criteria/assignee/reviewer via generic edits refuses.

## Identity, scope and persistence

The private supervisor/controller audit channel supports version-3 `delegation`
frames with `requestId`, `taskId`, `taskRevision`, `delegateId` and a
supervisor-owned operation snapshot. `ToolSession.ProposeDelegation` overwrites
that snapshot from its actual parent Authority; model-supplied permission claims
are ignored. A custom supervisor that forges this snapshot violates the trusted
worker contract. Child intent acknowledgment intersects the recorded parent
operation mask with the child broker's own policy; a read-only parent cannot
delegate write, patch or command permission. The registered
`toolbroker.Scope` supplies job, room, participant, connection, provider,
repository and deadline. The engine request cannot supply a claimed actor, approvals,
credentials, launch settings, capabilities or different resources. The channel
must remain authenticated and separate from model stdout; a trusted source fork
can violate this contract and must not be treated as an equivalent implementation.
`ToolSession.ProposeDelegation` shares the supervisor's active-work slot; requests
cannot overlap an in-flight tool acknowledgment. Unsupported/mixed/duplicate-key
wire frames close the channel. Existing version-1 tools and version-2 idle events
retain their strict stage contracts; there is no fallback to a public agent login.

The controller revalidates current human provenance and worker scope before
persistence. The parent must be a submitted paired task job with exact pending
human-dispatched task identity and revision. A deterministic job/request ID
makes acknowledged proposal retries idempotent; changed retry input refuses.
At most the normal tool-call budget of proposals is admitted per parent job.
Source and delegate actor/connection/participant revisions are recorded,
and disabled agent identities refuse further delegation effects, and ordinary
repository, context-disclosure, human access, quota and audit checks remain required.

Managed jobs journal immutable task/dispatch/delegation bindings. Delegated
admission/submission, broker intent/outcome, native-session persistence and final
result publication revalidate the chain. Off/on uses a durable generation, so
old proposals cannot revive. A coordinator-derived parent also depends on its
current appointment revision, without making the delegate a coordinator.
Already-acknowledged in-flight effects are not claimed to be undone. Cross-controller
revocation is enforced at subsequent authorization boundaries; immediate remote
process termination is not assumed. Normal deadlines/reconciliation remain.

`task.delegation.proposed` records the authenticated source job and bounded target.
`task.delegation.approved` atomically updates the decision and task assignment;
`task.delegation.rejected` records a terminal refusal without reassignment. Replay
validates these scopes and preserves provenance through later task lifecycle
updates. No generic event endpoint can manufacture these events. Failed required
persistence latches integrity stop; optimistic conflicts return refusal without
falsely declaring corruption. Unknown/uncertain outcomes are not retried as effects.

## Owner review API

`GET /api/task-delegations?room=engineering` requires an enabled authenticated
owner and room access. It returns `enabled` and the room/company-filtered
`delegations` list, including retained history when the feature is off.
The full map is private, excluded from generic state serialization.

`POST /api/task-delegations?room=engineering` requires the owner session, CSRF token
and a unique `Idempotency-Key`. It accepts only an explicit review:

```json
{"action":"approve","id":"proposal-id","revision":1}
```

Use `reject` for refusal. No HTTP `propose` action exists. Unknown/duplicate fields,
trailing values and missing revisions refuse. Stale decisions, conflicting journal
updates, unreconciled parent work or changed scope return 409 without partial
assignment. Acknowledged retries return the same receipt without another revision
or dispatch. Off returns 403. Rejection may record an expired proposal's disposition;
approval never renews it. The repository policy must still match at approval and
later dispatch. Credentials never appear in these records.

## Verification limits

All tests mock database, worker and provider boundaries. Tests cover private
stream identity, duplicate receipt, malformed frames, actual human review and
dispatch endpoints, parent shutdown ordering, durable task/job scope, child tool
acknowledgments, no automatic dispatch, feature-off and off/on invalidation,
role/membership/connection changes, forged browser identity, audit rollback,
competing disable, explicit rejection, scope expansion and subdelegation refusal.
UI tests verify escaped scope display and separate review/dispatch requests.

No real provider, Docker, database or live multi-controller/browser deployment
was exercised. Provider emission, reviewed production assembly and full GOV-07
live acceptance remain outstanding. No P0 release gate is certified by synthetic
acceptance. See [governance design](governance-model.md), [task proposals](task-proposals.md)
and [P0 acceptance](p0-release-gate-acceptance.md).
