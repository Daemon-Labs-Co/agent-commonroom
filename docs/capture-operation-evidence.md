# Creator-scoped capture operation evidence

The capture session evidence adapter checks actual retained local history and
read-only container metadata before and after each lifecycle operation. It joins
the controller's operation guard to the existing persisted-transition observer.
It is mandatory security infrastructure and has no optional bypass flag.

## Assembly

Construct `newWorkspaceCaptureSessionEvidence` from the actual active creator
observation lease, authenticated worker alias, reviewed finite capture intent and
mandatory `captureSessionWitness`. Pass the result to the
[pinned controller session](controller-capture-session.md). Both source and object
volume identities and the helper image must match the retained creator scope.
An arbitrary job record or newly acquired name-based handle is insufficient.

The witness must independently authenticate current operation-specific source or
cleanup authority, approved worker/creator/image/volumes, account capacity and
storage accounting, and real physically enforced quotas. Its lifecycle method
also verifies the persisted transition. It cannot infer these from peer frames,
container labels or mode bits. No permissive witness or deployed physical quota
provisioning is supplied. Callbacks must not reacquire controller/admission locks or
call over the waiting private stream.

## Observation contract

For operation prechecks, the adapter rereads the exact retained creator record and
cancellation fences. It reads canonical private capture records and the integrity
stop latch. Before a begin record exists, only begin can be checked. Otherwise the
requested event must be the exact latest controller transition or a valid next
transition extending retained history. Settled local records cannot authorize a
new lifecycle. Unknown helper IDs cannot become known through a recovery request.

Independent witness permission is required before physical inspection. Begin
requires no references to either retained volume, job or helper name. Bound requires
strict admission of the exact created helper. Finish and recovery inspect only the
known durable ID; claimed absence must actually be observed. Unknown uncertain
finish can check only that references remain absent, without adopting any ID.
Inspection is shared with the lifecycle observer and performs no create, start,
remove, byte-read or quota-refund effects.

After inspection, the adapter rereads retained fences, integrity latch and complete
local history. Any change rejects the operation. Independent permission/quota
witness and creator lifetime must still hold at the end. Detached callback copies
cannot rewrite the pinned creator record or output selection. The controller session
separately reloads current permissions, flags and audit revision. This observer is
physical evidence, not a replacement for those checks or durable acknowledgment.

Acknowledgments use the embedded
[independent lifecycle evidence](capture-lifecycle-evidence.md), which requires the
exact already-persisted local candidate. Operation checks may occur before local
persistence; they do not pretend a proposed record has been persisted. A successful
precheck reserves no capacity, grants no source bytes and does not prove completion.
Only the ledger's successful durable acknowledgment permits lifecycle advancement.

## Verification and limits

Tests use actual temporary synthetic creator leases and canonical private journals,
with mocked Docker and independent witnesses. They cover begin/bound/finish/recovery
before and after persistence, independent denial and revocation, extra references,
false absence, corrupt/changed local history, changed creator record, foreign worker,
expired scope, integrity stop, old transition reuse and unknown-ID adoption. No real
DB, Docker daemon, provider, credentials or private messages are used.

Deployed authentication, concrete operation-specific quota/account witnesses and
physical source/object quota provisioning remain incomplete. End-to-end worker/
custody/artifact delivery with human review, rootless image/latency/live acceptance
and unknown-ID incident handling remain required. Required file-task dispatch is
still blocked; synthetic evidence tests certify no live P0 release gate. Source
forks can substitute the trusted witness/runner or fabricate creator scope and
invalidate isolation guarantees; there is no supported host fallback.
