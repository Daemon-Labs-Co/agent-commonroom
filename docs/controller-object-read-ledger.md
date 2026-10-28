# Controller artifact-reader lifecycle ledger

This private controller ledger durably accounts for isolated artifact-reader
helpers across an account. It prevents helper reads from escaping the same
container budget used by managed jobs, workspace probes and output captures.
It does not deliver files, execute a task, accept a deliverable or certify a
physical cleanup observation.

## User and operator behavior

There is no new button, public API or configuration setting. Existing task,
artifact, room and owner permissions remain in force. An interrupted reader can
keep capacity reserved until trusted infrastructure verifies its exact recorded
container has been removed. Do not clear the reservation by deleting worker files
or editing audit history. Reader cleanup never refunds retained object storage.

For example, `begin -> bound -> finish` records one reader and its exact container
ID. A finish with independently verified absence releases its controller helper
slot. A finish without that evidence keeps the charge. Known-ID recovery records
`recovery -> recovery-finish`; repeated attempts at the same recovery checkpoint
must authenticate again without appending duplicate events. Lost/unknown IDs
remain unresolved; name discovery and adopting another container are unsupported.

Worker confirmation remains a separate durable step: controller acknowledgment
alone cannot free the worker's slot or return bytes. A task is complete only through
the existing human acceptance workflow, not because its reader exited.

## Trusted integration contract

`newControllerObjectReadLedger` requires a server and non-nil
`controllerObjectReadEvidence`. Its private `Acknowledge(ctx, intent, event)` method
requires an authenticated human principal supplied by trusted middleware/transport.
Missing identity never becomes the owner. Worker/model metadata cannot authenticate
a principal, worker, quota, sealed receipt, Docker volume or cleanup observation.

The evidence adapter must independently verify current operation-specific export
or cleanup permission, approved worker/image, actual creator ownership, private
local journal persistence, exact custody volume, physical quota and container
observations for each transition, including retries. It must honor cancellation
and must not reenter server or worker admission locks. There is deliberately no
permissive default adapter or network authentication implementation.

Begin and bound transitions require the exact recorded sealed receipt and its
unreleased storage reservation. Historical job, task execution, declared output,
company, room and worker identities must match. Completed tasks and disabled task
editing do not erase a retained artifact's meaning. Current incident/integrity stops
refuse all acknowledgments; launch stops additionally refuse begin/bound. Cleanup
requires current enabled human room access and either the original reader principal
or the owner, plus independent cleanup evidence. It grants no byte-release rights.

The ledger reloads authority after evidence validation and compares the global
audit revision. A concurrent audit change refuses the operation; a failed audit
write trips the controller integrity latch. Exact latest-transition retries repeat
current authority and evidence but do not append. Earlier checkpoints cannot
reopen settled reads. Duplicate events in replay are invalid audit history.

## Storage, replay and capacity

The internal event kind is `artifact.reader.acknowledged`. It is not accepted by
the public event endpoint. `State.ObjectReads` is excluded from public JSON and
stores immutable request/image/volume metadata and up to five controller lifecycle
events per reader, keyed by job/challenge. Local worker-only `confirmed` and
`recovered` stages cannot be injected into this ledger. Histories are bounded to
1024 reader records; exhaustion refuses new readers, with no automatic deletion
or retention bypass. Existing histories with no reader events need no migration.

Replay validates transitions, recorded source identity, room authority and sealed
receipt linkage at admission. All helpers count toward the existing account
container limit; malformed reader accounting refuses admission rather than
silently omitting a charge. Unresolved same-job readers and output captures exclude
one another in both directions. Concurrent controllers are fenced by atomic audit
revision checks, not just an in-process mutex.

## Feature flags and security boundary

This ledger is mandatory isolation/accounting infrastructure, with no feature flag,
off switch or restart setting. Task editing flags do not authorize reads or disable
cleanup accounting. Existing release policy still requires independently verified
current export permission. No host read, local/SSH provider launch, storage refund,
public download or optional isolation path is introduced. Source-level replacement
of evidence or persistence can invalidate the security contract.

## Verification and limitations

Synthetic tests cover durable replay, normal and uncertain finish, known recovery,
fresh retry authorization, wrong/unknown IDs, shared helper quota, capture exclusion,
concurrent controllers, permission and receipt substitution refusals, editing-off
retained reads, stale/revoked authority, audit failure, detached callback history,
private state and public event refusal. Database/provider/Docker boundaries are
mocked; fixtures use temporary synthetic files and never private data/credentials.

Still incomplete: authenticated worker/controller transport and session assembly,
independent physical quota/custody/export witnesses, worker/runtime orchestration,
object-only ownership after workspace disposal, unknown-ID incident reconciliation,
reviewed deployment/live acceptance and public artifact delivery. Required file-task
dispatch remains blocked. Passing synthetic checks certifies none of those live
release gates.

See [worker reader journal](worker-object-read-journal.md),
[reader supervision](object-reader-supervision.md),
[controller read sessions](controller-artifact-read-session.md) and
[mandatory isolation](architecture.md#mandatory-container-isolation-design-contract).
