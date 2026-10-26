# Controller capture observation authority

The private `controllerCaptureAuthority` implements the capture helper's source
observation guard. It binds an authenticated human session, worker identity, exact
retained execution record and contract-derived request before any observation can
be acknowledged. It is not a model-facing endpoint, lifecycle capacity admission,
object seal, public export capability or task acceptance decision.

## Trusted construction and configuration

Internal session setup supplies the authenticated context and worker, retained
record, request and mandatory independent `controllerCaptureEvidence`. The record
must describe reconciled engine/tool roles and their retained source volume. The
request must match the job, repository and exact volume name/nonce/creation time.
Members and nested record state are detached from caller mutation. Peer/helper
frames may not choose the human or worker identity.

The evidence implementation must independently verify creator ownership, the exact
reviewed running helper, its source/object volume identities, real mount proof,
all other writers fenced and physical storage accounting. It must respect context,
avoid server-lock or same-stream reentry and use independently authenticated worker
metadata. No permissive/default evidence implementation or deployed peer assembly
is supplied. A source fork that lies through this interface defeats isolation.

## Current authorization and durable acknowledgment

Each `AuthorizeFrozenSource` call verifies the exact pinned binding/challenge and
live session, then loads current controller state. It requires an enabled human
with room access and owner/initiator authority; the exact worker/provider/job,
retained execution/source and original unexpired task binding; the exact execution
record/definition and complete canonical contract member set; current task,
delegation/coordination and feature-flag policy; and no incident, launch block or
integrity fault. Missing current definition or a stale/replaced attempt refuses.
A canceled session cannot be revived using a new operation context or actor ID.

After independent physical evidence succeeds, the controller reloads and rechecks
all policy. A changed audit revision produces a conflict. It appends the private
`workspace.capture.authorized` event with expected-revision fencing before returning
success. Persistence uncertainty trips integrity stop; ordinary conflicts do not.
Cancellation detected after persistence still withholds the acknowledgment; stored
history is retained.

The event contains finite request/mount metadata and a digest of the retained
record, not source bytes or credentials. It grants no reusable authority or quota
release. Public event injection is refused. Audit summaries expose no selected
source contents. Replay checks historical identity/contract/proof shape without
reapplying today's deadline/flags/physical state; a past observation is never a
fresh capability. Current authorization separately applies those live conditions.

## Flags, tests and deployment limits

There is no switch for bypassing this mandatory observation boundary. Existing
server-side task/delegation/coordination flags remain enforced by current job checks;
switching them off cannot grant source access. Historical records remain replayable.
The managed required-file dispatch barrier remains in place.

Mocked tests cover durable grant/replay, pinned principal and detached caller state,
physical denial, wrong binding/proof/worker, missing/agent actor, canceled lifetime,
feature-off refusal, stale audit revisions, failed persistence/integrity stop,
current-vs-historical expiry/incident/definition/member checks and public injection.
All DB/worker/provider boundaries are mocked with synthetic fixtures. This does not
prove actual physical quotas, source isolation, trusted images or live readiness.

Remaining: deployed independent capture evidence, durable lifecycle/account/creator
and custody composition, physically enforced source/object quotas and volume setup,
reviewed image/live acceptance, recovery incident handling and authenticated delivery
with human review. Do not enable dispatch using synthetic evidence or no-op guards.
Real daemon/inspection and durable-journal latency within the bounded control-channel
deadlines has not been measured; synthetic tests do not establish live throughput.

[Worker-local capture evidence](worker-capture-evidence.md) supplies the creator-scoped
retained-record and exact Docker inspection adapter. Its independent durable binding,
physical-quota and private-proof witness remains required and has no default.
