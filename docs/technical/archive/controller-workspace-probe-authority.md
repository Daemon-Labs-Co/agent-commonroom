# Controller authority for retained workspace probes

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


The controller now assembles the retained workspace probe guard used behind the
private authority channel. Before acknowledging metadata inspection it checks the
current human, worker/job binding and task dependencies, obtains independent
physical evidence, and commits a durable audit record. Helper lifecycle events
use the existing account-wide capacity ledger. This is an internal integration;
it does not enable file outputs or provide deployed worker authentication.

## Usage and configuration

Trusted authenticated session setup constructs
`newControllerWorkspaceProbeAuthority(ctx, server, worker, record, intent, evidence)`.
The context must identify the original authenticated human. The worker alias must
be independently authenticated and approved, and the retained record and fresh
intent must come from trusted setup, not the agent or a decoded peer frame. Supply
the resulting guard to `serveProbeAuthority` for that same record and intent.
Missing, typed-nil or malformed dependencies refuse. There is no environment-only
shortcut, public HTTP constructor or default physical evidence implementation.

The supplied `controllerWorkspaceProbeEvidence` must independently verify the
approved worker, actual creator ownership, reviewed helper image, exact retained
volume and complete reference inventory on every authorization/observation. Its
lifecycle validator must independently prove container identity and absence before
release. It must not call back into controller locks. The controller guard is not
itself proof of these physical facts. Deployed authentication, remote invocation
and physical evidence remain required before production use.

## Authority and durable ordering

The human principal, worker, exact retained record and probe intent are pinned to
one session. Caller contexts provide cancellation, not an opportunity to replace
the principal. Each operation rechecks enabled human room access and owner/original
initiator authority. Observation/authorization also require a retained unconfirmed
job with matching provider, membership, execution/repository and deadline, current
task/definition, delegation and coordination dependencies, and no launch/integrity
stop. A canceled session cannot revive with a fresh request context. Recovery needs
its own fresh authenticated authority; this guard adds no recovery privilege.

Observation and authorization commit `workspace.probe.authorized` before success.
The record contains bounded identity metadata, operation, exact intent and a digest
of the retained record, not files or credentials. Its projection checks historical
identity at the event position; audit records grant no reusable capability. Current
clock-sensitive task/delegation checks happen on the live operation rather than
making historical authorization unreadable after expiration.

After physical validation the controller reloads state, requires the same revision
and appends with expected-event-count fencing. Revocation during physical checks
refuses. Ordinary concurrent changes refuse without declaring an integrity incident;
actual persistence failure latches the existing integrity stop. Helper lifecycle
acknowledgments additionally pin the session worker inside the ledger's fenced
snapshot, preventing a changed job worker from reusing an old authenticated session.
Only the existing independently verified absent lifecycle outcome releases a helper
slot. Retained workspace/storage charges remain separate.

## Flags, permissions and limits

This mandatory authorization/audit infrastructure has no optional security toggle
and no public UI/API. Optional task controls never bypass its human authorization,
physical validation or audit rules. Switching task definitions off still revokes
structured task authority through the existing dependency checks; legacy metadata
probes do not depend on optional diagnostics or task-dispatch display controls.
Public event injection is refused. Generic audit summaries omit probe identity.

Source forks can substitute dishonest evidence, connectors or guards and invalidate
these guarantees. No managed agent host execution or uncontained fallback exists.
The guard produces metadata authorization only: no frozen source root, bytes,
automatic acceptance or repository publication authority.

## Verification

Automated tests mock every database and Docker/provider interaction and use
synthetic temporary retained jobs and private streams. They cover durable audit and
replay, actual probe/journal/channel/controller-ledger assembly, confirmed helper
slot release, callback mutation, missing/agent/disabled/unrelated principals, room
access, session cancellation and revocation during evidence, changed worker/record,
stale task, both optional feature states, audit failure and stale revisions.

These tests do not certify a deployed worker, live Linux image, provider turn,
physical quota or file export. Remaining MVP work includes production session and
physical evidence integration, source freezing/capture, storage enforcement and
authenticated deliverable delivery/review.

A [worker-local physical implementation](workspace-probe-physical-evidence.md)
now performs these checks under actual creator ownership. It is not a deployed
remote attestation/evidence transport; authenticated cross-process routing remains
required when the controller and worker are on different hosts.
