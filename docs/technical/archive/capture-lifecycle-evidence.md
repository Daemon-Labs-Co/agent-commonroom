# Independent capture lifecycle evidence

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


`workspaceCaptureLedgerEvidence` verifies worker-local evidence before the controller
acknowledges a capture lifecycle transition. It is constructed only inside the live
creator-ownership callback, for one authenticated worker, retained record and exact
request/image/object-volume intent. It implements metadata observation only: no
create, start, remove, byte read, refund or volume deletion is exposed.

## Configuration and trusted witness

Trusted worker setup calls `newWorkspaceCaptureLedgerEvidence` with the creator
scope, worker, intent and mandatory `captureLifecycleWitness`, then supplies it to
`newControllerCaptureLedger`. Scope lifetime, actual creator lease, retained source
identity and runner must be available. Requests and records are detached from caller
mutation. A missing scope/witness fails closed.

The witness must independently authenticate operation-specific current source/cleanup
permission, actual creator/worker/image/volume identity, account capacity/storage
accounting and physically enforced quotas. It must honor context and not reenter
controller/admission locks or the same control stream. No production or permissive
witness is supplied; Docker labels/counters do not establish physical storage quotas.

## Required evidence ordering

For every transition the observer checks the exact worker/job/execution/definition/
company/room binding and the controller candidate's intent/history. It rereads the
actual retained record and cancellation fence, then reads bounded canonical local
capture records. Controller events must exactly match the persisted local history
and its last transition. Local terminal `confirmed`/`recovered` markers are stripped
only for this comparison; they cannot fabricate a missing controller transition.

After the first independent witness succeeds:

- `begin` requires both retained volumes and no job/helper-name/volume references.
- `bound` requires the exact locally recorded helper in strictly validated created
  state, with both volume identities and complete reference/catalog checks.
- `finish`/`recovery`/`recovery-finish` inspect only the durable exact ID. A claimed
  absence must actually be observed. A present owned drifted helper grants cleanup
  eligibility only, never source reads or execution. Unknown empty IDs never become
  discovered/adopted names; an uncertain finish can be observed only with no refs.

The observer rereads the retained fences and local history after physical checks,
then requires the witness again. Changed records/history, revoked witnesses or an
expired scope reject acknowledgment. The controller separately reloads current
human/task/capacity state and applies its audit-revision fence before persistence.

## Read-only inspection APIs and limits

`CaptureContainerPlan.VerifyNoReferences` performs bounded two-pass retained-volume/
catalog/reference checks without discovering an ID. `InspectRecorded` inspects only
a separately durably known immutable ID and reports observed presence/absence.
Neither authorizes removal or storage release. Runtime/recovery retain their own
trusted journal requirements and exact-ID removal paths.

No optional product flag disables these mandatory safeguards. Existing task flags
remain server-enforced for new work; independently authorized cleanup may preserve
history while flags are off. Required file-task dispatch remains blocked.

Tests use actual synthetic local journal files and creator locks, with mocked
Docker and quota/account witnesses. They cover vacant begin, exact stopped binding,
known recovery, actual absence, false absence claims, mismatched/corrupt history,
quota refusal, witness revocation, extra refs, missing scope and expiry. Observers
attempt no lifecycle effects, and no real DB/provider/container/private fixture is
used. This is not live quota or image acceptance.

Remaining: deployed authenticated worker transport/guard, real physical quota and
account witness, creator/custody orchestration, volume provisioning, reviewed image/
latency/live acceptance, unknown-ID incident handling and authenticated artifact
delivery/human review. Source forks with dishonest witnesses/runners invalidate
isolation claims. Never enable dispatch with a no-op witness.
