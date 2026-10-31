# Interrupted capture recovery

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


`toolbroker.CaptureRecovery` reconciles an interrupted capture helper from its
already durable exact-ID record. It is internal mandatory isolation/reconciliation
infrastructure, with no feature off switch or host/SSH fallback. It does not resume
capture, publish reports, release objects or delete volumes.

## Trusted configuration and use

An authenticated supervisor reconstructs the exact `CaptureContainerPlan` from the
durable request, reviewed image and retained source/custody volume identities. It
constructs `NewCaptureRecovery(plan, approvedRunner, recoveryJournal)` and calls
`Reconcile(ctx, recordedContainerID)` once. Empty IDs, names and invalid/canceled
inputs fail before any daemon invocation. A valid-looking ID is never sufficient
by itself: the recovery journal must match it to the actual interrupted record.

`BeginCaptureRecovery` must authenticate current cleanup permission and pin the
request, reviewed image, object-volume identity and exact bound ID against durable
history. It must acquire exclusive creator ownership and retain capacity/charges;
it does not authorize new creation, source export or object writes. The image is
passed explicitly for this comparison, as it now is for ordinary `BeginCapture`.
A failed/uncertain journal write must preserve its own integrity-stop obligations.
No permissive journal or production account/controller assembly is provided.

After authorization the reconciler checks physical evidence twice: exact presence
or absence by ID; ownership by ID/name/image/labels when present; matching retained
source/object volume nonces, timestamps, local driver and canonical mountpoints;
and exact job/name/volume-reference and volume-catalog inventories. A changed
second-pass presence, replaced volume or another attached container blocks cleanup.
Names are checked for conflicts, never used to discover or adopt a lost ID.

An owned known helper may be removed even if its running policy has drifted. That
narrow ownership check authorizes stopping only, never restarting or reading from
it. Removal uses the exact immutable ID. Already-absent helpers require the same
retained-volume and reference checks. The runtime and recovery path share these
physical inspection rules to prevent recovery policy diverging from normal cleanup.

## Durable result and uncertainty

`FinishCaptureRecovery` records actual observed absence or uncertainty with fresh
current authority before reconciliation succeeds. The operation has a ten-second
bound and a separate five-second final-journal deadline. Cancellation, failed
removal, returning references, identity drift and final persistence denial retain
an unconfirmed custody error. A successful physical removal followed by failed
persistence still requires reconciliation; no success is fabricated.

A nil result certifies only the observed helper absence plus durable acknowledgment.
It does not certify source validity, complete objects, physical quota release or a
publishable capture report. Stored partial bytes, individual seals, retained volumes
and storage charges remain. Unknown-ID creation failures require an incident path;
there is no name adoption, scan-and-delete, automatic retry/refund or volume cleanup.
A source fork with dishonest runner/journal evidence defeats isolation guarantees.

## Verification and deployment gaps

Synthetic tests mock all Docker and journal interactions. They cover present,
already absent and owned drifted helpers; exact removal; image/label/volume
replacement; extra references/catalog entries; returning references; cleanup/final
persistence failure; authorization/record mismatch before daemon; canceled and
missing IDs; typed-nil dependencies and one-shot use. No real database, provider,
container or private fixture is used.

Durable deployed recovery/creator/account integration, unknown-ID incident handling,
physical quotas, image/live acceptance and public file delivery remain incomplete.
Do not enable required file-task dispatch using synthetic acceptance as evidence.

## Retained worker assembly

Trusted setup can now call `recoverRetainedWorkerCapture(ctx, observation, intent,
recordedContainerID, freshGuard)` inside the retained worker ownership callback.
The helper constructs the fixed plan and actual private worker lifecycle journal
from that observation and durable intent, then invokes the cleanup-only reconciler.
For example, a helper whose prior record contains `begin`, the exact `bound` ID
and an uncertain `finish` can be reconciled after independently authenticating
current cleanup permission. It returns only an error, never capture metadata or
file bytes.

The guard MUST be freshly authenticated and enforce current human/task cleanup
permission, physical creator/quota/volume evidence and durable controller
acknowledgments. A previous capture authorization, valid-looking ID or successful
old exchange does not authorize this call. The operation receives no starter,
source authority or object journal; it cannot launch a helper, copy outputs, delete
volumes, publish a report or refund artifact storage.

`CaptureAuthorityLifetime()` must remain live throughout reconciliation. Losing
the control connection cancels recovery even between exchanges; an older positive
acknowledgment cannot grant continuing cleanup authority. A fresh cleanup session
must be independently authenticated after loss. Already retained records and
charges remain, and this cancellation is not evidence that the helper is absent.

The complete reconciliation is tracked under actual creator ownership, including
failure revocation. Callback teardown cancels and joins it before creator unlock.
Canceled/expired ownership refuses new work; final ownership loss withholds success.
Trusted callbacks must honor cancellation and may not reenter their waiting stream
or ownership teardown. Private control transport closure/join remains caller-owned.
There is no public endpoint, optional bypass flag or host execution fallback.

A failed removal can leave the exact `recovery` checkpoint already recorded.
An explicit later call with a NEW current guard may resume only that SAME checkpoint
and bound ID. It reauthorizes, persists and durably reacknowledges the existing
transition without appending duplicate logical history. A denied new guard performs
no daemon operations. Different/unknown IDs and terminal `recovered` history cannot
be resumed. There is no automatic retry or name discovery. Audit failure retains
the worker integrity stop; this helper cannot clear it to allow another attempt.

Synthetic worker-assembly tests cover present/already-absent helpers, failed removal
followed by denied and allowed explicit reconciliation, wrong/unknown IDs, current
permission denial, failed acknowledgment, cancellation, expired ownership and
terminal replay refusal. They check retained helper charges, no capture starts or
custody writes and no daemon effects before durable cleanup admission. Full deployed
account/physical evidence and live rootless acceptance remain outstanding.
