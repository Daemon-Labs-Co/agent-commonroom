# Durable worker capture journal

`workerCaptureJournal` implements capture lifecycle and known-ID recovery journals
under the actual worker creator-ownership lifetime. It records finite intent, exact
helper ID and cleanup history in private worker files before controller acknowledgment.
No model/provider interface, host execution fallback or optional bypass is added.

## Configuration and durable ordering

Inside the creator callback, trusted setup calls `newWorkerCaptureJournal` with the
exact request, reviewed image, retained object-volume identity and mandatory
`workerCaptureJournalGuard`. The callback lifetime/lease must remain held through
cleanup. Inputs must match the retained source/job/repository creation identity.
No permissive controller/account/physical-quota guard is supplied.

The guard must authenticate current operation-specific source or cleanup permission,
actual creator/image/volume/quota evidence and account-wide capacity. Its acknowledgment
must durably mirror the exact local transition. Callbacks run under shared worker
admission ownership and must not reenter that lock. Requests, nested records and
event slices are detached before callbacks to prevent accidental mutation.

Transitions serialize with `.admission.lease`. `begin` is persisted and synced before
its acknowledgment permits creation; `bound` records the exact immutable ID before
start. `finish` retains actual absence or uncertainty before final acknowledgment.
Only successful current acknowledgment followed by a durable `confirmed` marker
releases the helper slot. Lost create replies retain an empty ID and remain charged.
Names are never adopted. The journal does not read source bytes or delete volumes.

A known bound ID may enter explicit `recovery`, then a verified `recovery-finish` and
acknowledged durable `recovered` marker. Unknown IDs, replayed transitions, substituted
requests/images/object volumes and unsupported histories are refused. There is no
automatic retry or invented completion record. Unconfirmed recovery remains charged.

## Worker admission and failure behavior

Existing worker/probe admission now also counts every unconfirmed capture. The shared
four-container worker ceiling remains conservative, including retained creator roles.
It does not replace account-wide multi-worker admission. A second unsettled capture
for the same job is refused even when slots would otherwise fit.

Any ambiguous local persistence or controller acknowledgment writes the existing
worker integrity-stop marker and independently revokes live authority. Marker failure
cannot authorize continued work. Ordinary permission denial does not manufacture an
integrity incident; already persisted intent/uncertainty and charges remain retained.
A final physical-absence observation alone never releases capacity.

Records use bounded canonical JSON, exact derived names, private owned regular files
and no links. Reads reject malformed/noncanonical/unsafe records. Such records also
block new worker/probe admission. Historical records are retained; no cleanup API
silently deletes them. Directory/record-count limits remain bounded. Metadata-only
validation uses `toolbroker.ValidCaptureMetadata`; it invents no observed daemon root.

## Flags, tests and remaining work

There is no toggle for disabling durable isolation/accounting safeguards. Existing
product flags and permissions must still be enforced by the guard/controller. The
required-file dispatch barrier remains unchanged.

Tests persist actual synthetic private files and use real synthetic creator locks,
with mocked controller/quota guards. They cover durable-before-ack ordering, normal
confirmation, known recovery, unknown retention, capacity counting, denied authority,
second-capture refusal, failed local/remote persistence, failed confirmation, replay,
identity substitutions and corrupt/noncanonical/symlink/unsafe-mode records. No real
database, Docker daemon, provider or private fixture is used.

Still required: deployed durable controller capture/account ledger and authenticated
transport; physical quota/private-proof witness and volume provisioning; reviewed
image/latency/live acceptance; unknown-ID incident resolution; and authenticated
artifact delivery/human review. Local durable files and mocked acknowledgments are
not live P0 acceptance, proof of quota enforcement or a complete file-task workflow.
