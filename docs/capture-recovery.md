# Interrupted capture recovery

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
