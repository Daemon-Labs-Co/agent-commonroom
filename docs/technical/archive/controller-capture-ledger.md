# Controller capture lifecycle ledger

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


`controllerCaptureLedger` durably acknowledges worker capture lifecycle transitions
and accounts for their helper containers across the entire account. It complements
the private worker journal, current source-observation authority and physical
inspection. It is not a model event endpoint, file deliverable or quota provisioner.

## Trusted setup and transition flow

Internal authenticated controller code constructs `newControllerCaptureLedger`
with mandatory independent `controllerCaptureLedgerEvidence`, then calls
`Acknowledge(ctx, exactIntent, transition)`. The principal must come from trusted
middleware/private transport; a peer payload may not select a human or worker.
The evidence implementation must authenticate the actual worker and creator,
local persisted transition, reviewed image/volume and physical observations. No
permissive evidence or deployed transition transport is supplied.

Intent includes the complete contract-derived request, reviewed image and retained
object-volume creation identity. It binds the job, company, room, worker, task
execution and exact definition/member set. Current enabled human room permission
and owner/initiator authority are required for every transition.

New creation/binding additionally requires current task/delegation/coordination/
feature-flag policy, an unexpired original deadline, and no launch block. Cleanup
can proceed while product flags/launches are disabled, with fresh human and physical
cleanup authority. It grants no source access or new execution. Incidents and
integrity faults remain fail-closed.

The controller validates history/capacity, independently verifies evidence, reloads
and rechecks current state, then appends `workspace.capture.acknowledged` with expected
audit revision. A changed revision conflicts; persistence uncertainty trips integrity
stop. A valid retried acknowledgment is idempotent only after current authorization
and evidence succeed. Duplicate events in the audit log itself are rejected on replay.
Callback request/event slices are detached from pinned input.

## Capacity and recovery

Private `WorkspaceCaptures` state counts each unconfirmed helper once. Ordinary role
allocation, metadata-probe admission and capture admission all count both helper
families against the same four-container free-account ceiling, across workers and
rooms. Retained/unconfirmed role allocations remain charged conservatively. Malformed
history, changed execution binding or a forged release bit blocks admission.

Only independently verified, durably acknowledged helper absence in `finish` or
`recovery-finish` releases its controller helper slot. That releases neither source/
object storage quota nor retained bytes. Unknown creation IDs remain charged and
cannot be adopted by name. Explicit known-ID recovery requires matching history and
fresh cleanup authority. Worker-only `confirmed`/`recovered` markers are not accepted
as controller events; the local journal records them after acknowledgment.

History projection does not perform new physical checks or turn recorded transitions
into reusable execution capabilities. Current deadline/source policy remains a
separate live check. Public event injection is refused, and audit summaries contain
no source bytes or credentials.

## Flags, verification and limitations

No optional flag disables capacity, authentication, audit or isolation safeguards.
Existing task flags remain server-enforced for new work; disabling them preserves
history and permits only independently authorized cleanup. Required file-task
dispatch remains blocked until the full production assembly is ready.

Mocked DB/evidence tests cover durable lifecycle/replay, combined helper/role capacity,
idempotent retries, unknown retention, known recovery, physical/actor/definition/member/
flag denials, persistence/integrity failure, CAS conflicts, flags-off cleanup, malformed
release/history, callback mutation and public injection. No real database, container,
provider, private data or quota observation is used.

Remaining: deployed authenticated worker transition guard/transport, independent
local/physical lifecycle evidence, physically enforced source/object quotas and volume
provisioning, end-to-end creator/custody assembly, reviewed image/latency/live acceptance,
unknown-ID incident handling and authenticated artifact delivery/human review. The
ledger and synthetic acknowledgments do not certify live readiness.

The [worker lifecycle observer](capture-lifecycle-evidence.md) supplies actual private
journal and read-only Docker metadata checks. Its deployed quota/account/authority
witness and authenticated worker transport remain required.
