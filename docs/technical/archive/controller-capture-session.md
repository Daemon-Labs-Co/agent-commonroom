# Pinned controller capture sessions

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


The internal capture session adapter connects the private lifecycle stream's
journal guard to current controller policy and the durable capture ledger. It is
mandatory security infrastructure, not a user-facing feature or an isolation toggle.
It introduces no public route or network authentication mechanism.

## Setup and identity

Trusted private session setup supplies the authenticated human context, approved
worker alias, exact retained creator record, reviewed finite capture intent and
mandatory independent evidence implementation to `newControllerCaptureSession`.
The adapter copies its metadata and pins the principal and lifetime. Peer requests
cannot choose another human, worker, room, source volume, output selection, helper
image or object volume. Contexts supplied to later calls cannot replace the original
principal. Cancellation or an explicit stop permanently ends this session; recovery
requires fresh authenticated setup.

The evidence implementation combines `ValidateCaptureOperation` with the existing
controller ledger evidence interface. The former must independently verify current
operation-specific source/cleanup permission, actual creator ownership, physical
quotas, storage/account accounting and approved image/volume policy. The latter must
also prove exact persisted local history and actual physical transition/absence.
There is no permissive/default implementation. Use the creator-scoped lifecycle
observer for its supported acknowledgment checks; it does not supply physical quota
provisioning or the operation-specific witness by itself.

## Calls and persistence

`AuthorizeCapture` validates the exact retained record/intent/event, loads current
policy under the server lock, checks the pinned worker/job, and derives the only
valid next or latest acknowledged transition. It requires independent operation
evidence, then reloads policy and checks the audit revision. An older historical
transition cannot become fresh operation authority. This precheck does not reserve
capacity or grant file access. Only a separate successful durable lifecycle
acknowledgment permits the supervisor to advance its lifecycle.

`AcknowledgeCapture` validates the exact received local history inside the ledger's
locked snapshot. Truncated/substituted histories are refused before physical evidence.
The adapter pins worker/creator/intent throughout the evidence callback. The ledger
reauthorizes the human, reloads policy/capacity and appends with audit revision fencing
before acknowledgment. Identical current history can be retried under current
permission/evidence; this is not automatic transport replay. Persistence failures
trip integrity; audit conflicts remain conflicts and do not fabricate acknowledgments.

Begin and bound checks honor current task-definition/delegation/coordination policy,
launch restrictions and deadline. Separately authorized cleanup may proceed with
product flags off; disabling flags never grants new execution or source access.
History and uncertain reservations remain retained. Explicit session stop cancels
its own lifetime and trips controller integrity; clearing a global fault cannot
revive the old session.

Callbacks run inside the controller lock and may coincide with the worker's
admission lease. They must not reenter either lock or the waiting control stream.
Copies isolate record/intent/history callbacks; authoritative state is reloaded
before a successful decision. This adapter has no container launch, host execution,
source read, storage refund or authentication fallback.

## Verification and limits

Tests use mocked database and independent evidence, with synthetic task/job records.
They cover authenticated principal pinning, full durable lifecycle/retry, wrong
worker/record/intent, truncated history, missing or agent principal, physical denial,
flags-off denial and cleanup, stale audit snapshots, persistence faults, callback
mutation, canceled lifetime and permanent stop. They contact no real provider,
container or database and certify no production acceptance gate.

Deployed authenticated stream establishment and concrete operation/quota/account
witness composition remain incomplete. Actual source/object quota provisioning,
end-to-end worker/custody/artifact delivery and human review, image/latency/live
acceptance and unknown-ID incident handling remain required. File-task dispatch is
still blocked. Source forks that substitute evidence, principal setup or transport
can invalidate this contract; decoded metadata is never authentication.

See [private stream](capture-authority-transport.md),
[controller lifecycle ledger](controller-capture-ledger.md) and
[independent lifecycle evidence](capture-lifecycle-evidence.md).

[Creator-scoped operation evidence](capture-operation-evidence.md) implements the
retained-history and read-only physical observer for this interface, while retaining
mandatory independent operation and lifecycle quota/account witnesses.
