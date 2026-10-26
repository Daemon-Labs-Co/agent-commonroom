# Account-wide workspace probe capacity

The internal controller probe ledger reserves one physical-container slot for
each unresolved workspace mount helper. It shares the free edition's existing
four-container ceiling with managed jobs across rooms, workers and controller
processes. Ordinary managed-job admission now includes unresolved probe charges.
It does not enable file export or replace the worker-local journal and limits.

## Trusted controller usage

Trusted worker/controller transport constructs `newControllerProbeLedger(server,
evidence)` and calls `Acknowledge(ctx, intent, event)` from its lifecycle guard.
For example, an authenticated bridge delegates its begin acknowledgment with
`ledger.Acknowledge(ctx, recordedIntent, workerProbeEvent{Stage: "begin"})` before
allowing Docker creation. The exact managed job determines company, room and
worker; callers cannot choose another billing scope.

The context must carry the original authenticated human principal from trusted
middleware/transport. Missing identity never means owner, including cleanup with
a fresh timeout context. Enabled humans need current room access and must be the
job initiator or owner. Agents cannot assume human authority. Begin/bound also
require a retained version-4 job, current task/delegation/governance dependencies
and permission to launch under agent-stop policy. Cleanup acknowledgments may
run while launches are halted, but not through an integrity incident. Fresh
recovery permission remains the physical evidence guard's obligation.

`controllerProbeEvidence.ValidateProbe` is mandatory with no permissive default.
It must authenticate the approved worker and actual creator ownership, reviewed
image, exact volume identity and appropriate physical observations for each stage:
no references before creation, exact read-only helper before start, and independent
exact-ID absence before release. It must verify current cleanup permission for
recovery, never trust an `absent` boolean from a model or unauthenticated worker.
The callback only validates; it must not allocate containers, release capacity or
re-enter controller/admission locks. Production transport/evidence implementation
and connection to worker guards remain outstanding.

## Durable accounting and acknowledgment

`workspace.probe.acknowledged` records immutable company/room/worker/actor/intent
identity and one lifecycle stage. The reservation is private replay state and
never a source-root capability. Begin reserves one slot before its acknowledgment
returns. Unknown create replies, stopped/expired helpers and retained uncertain
outcomes keep that slot. Creator jobs remain conservatively charged independently;
a version-4 retained creator still occupies two role slots.

Normal begin, bound and finish histories, including known-ID recovery and
recovery-finish, are validated with the worker's finite transition rules. Local
confirmed/recovered stages cannot be injected into the controller ledger. History
is bounded to five acknowledged stages. Valid exact retries are reauthorized and
physically revalidated but do not append duplicate events or charge twice. The
same creator cannot reserve another challenge while a prior probe is unresolved.
Duplicate events in stored replay are corruption, not valid retry evidence.

Only an accepted physically verified absent finish/recovery-finish releases the
account probe slot. No timestamp, worker restart, job shutdown label, feature flag
or local file deletion releases it. Worker-local confirmation remains independent
and can conservatively retain a slot after the controller released its own charge.
Neither charge releases retained workspace/storage reservations or deletes data.

The controller loads current state, validates authority/capacity and physical
evidence, then reloads and requires the same event revision. Its append uses the
database's expected event count, so separate controllers cannot spend the same
last slot. Callback mutation or concurrent state changes cannot replace the
snapshot that was checked. Ordinary conflicts refuse without declaring an
integrity incident. Actual append/commit failure trips controller integrity stop.
A denied or failed acknowledgment must never authorize worker creation/start;
worker journal integrity handling remains mandatory.

## Availability, flags and verification

This mandatory accounting/isolation infrastructure has no on/off product flag,
HTTP command or UI setting. Existing task flags cannot bypass it. Public event
injection is rejected. Private intent/history is omitted from state responses;
audits expose a generic lifecycle summary rather than volume/nonce data.
Historical records are validated at their event positions, so later room/task or
actor changes do not invalidate legitimate earlier acknowledgments.

Tests use only mocked event persistence and synthetic physical validators. They
cover restart replay, cross-worker quota, concurrent/stale controllers, duplicate
acknowledgments, human room permissions and revocation, denied evidence, account
scope/identity mismatch, malformed history, cleanup during launch stops, private
state omission, public injection refusal and audit failure. They do not connect
to a database or launch Docker, providers or SSH, and use no credentials/private
fixtures.

Remaining work includes authenticated controller/worker transport and invocation,
production physical evidence/continuous reference enforcement, live Linux image
acceptance, unknown-ID/incident recovery, retention/compaction, physical storage
quota, frozen source-root authority and authenticated delivery. The durable
controller accounting portion is implemented; synthetic tests do not certify
live account-wide admission. Required file/bundle dispatch remains blocked and
no live P0 gate is certified. Source forks can falsify trusted evidence or bypass
accounting; such changes invalidate managed isolation claims. There is no host
or automatic uncontained execution fallback.
