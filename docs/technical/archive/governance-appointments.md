# Room coordination appointments

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


This increment implements durable initial appointments and explicit human-selected
replacement and explicit human-requested, role-dependent turns independently of
the model provider. It does not implement agent-issued grants, automated delegation,
task extraction, consensus or a new execution path. A coordinator record grants no extra permissions; existing human
API authentication, task permissions, isolation, quotas, audit and retention still
apply. The [full design](governance-model.md) contains the remaining requirements.

## Enable and use

The owner opens **Settings → Agent coordination** and enables room appointments.
The `agentCoordination` feature flag defaults to false, is owner-only, persists in
the audit journal and takes effect immediately without restart. It cannot be set
by an agent. While enabled, the first newly registered eligible agent in a room
is appointed in the same transaction as its registration. Eligible agents are
Idle/Working room members with a currently supported executable connection
contract; this does not verify a live provider, authorize development or imply an
approved production execution image. Existing rooms are not retroactively elected.

In the room, choose **Choose coordinating agent**, select a loaded member and
confirm. Existing rooms can use this to establish their first explicit appointment.
Each selection increments the room's governance revision. A new agent joining,
a reconnect or losing the current coordinator never silently transfers the role.
If the appointed member is removed, blocked/stopped or loses its connection, the
record remains but is unavailable. The owner selects an eligible replacement.

Disabling prevents new automatic and manual appointments, preserves all records
and marks the appointment unavailable. Re-enabling restores visibility and eligible
status without an election or revision change. Existing independent authorized
jobs continue under their own permissions and safety stops. Role-dependent turns are cancelled locally and lose admission, tool acknowledgments
and result-publication authority. Re-enabling never revives an old dependency; an
explicit fresh request is required.
Human stop/recovery remain available regardless of this flag.

## API and persistence contract

`GET /api/governance?room=engineering` requires a signed-in owner and returns
`enabled`, `appointment` and `available`. The appointment contains company, room,
participant and revision. An empty appointment has revision zero. Full appointment
maps are private controller projections, excluded from generic state responses.

`POST /api/governance?room=engineering` requires owner authentication, a CSRF token
and a unique `Idempotency-Key`. Example JSON body:

```json
{"participantId":"selected-room-agent","revision":1}
```

The expected revision must match. Unknown JSON fields, trailing values, negative
revisions, unsupported/foreign/absent targets and missing operation keys refuse.
Conflicts return 409 and require reload. An identical acknowledged operation replay
returns its original receipt without reassigning again; changed input under the
same key refuses. Replays acknowledge prior history, never resurrect old authority.
Feature off returns 403; emergency/integrity stops prevent new appointments.

Appointment and operation receipt commit atomically under the append-only audit
lock with expected event count. Another controller's concurrent mutation causes
conflict rather than overwriting a selection. Failure grants nothing, preserves
the prior appointment and latches integrity stop for a required persistence failure.
Journal replay rejects invalid scope, revision, target or performed-by identity.
Generic event commands cannot inject `governance.appointed` directly.

The durable appointment is metadata, not a bearer credential or human approval.
The server must not authorize a tool call because its participant is coordinator.
Explicit human coordination turns bind the governance revision and revalidate at
the existing managed execution boundaries described below. Agent-driven coordination,
delegation and full GOV-06/07/08/09/13 acceptance remain incomplete; no P0 production
execution gate is certified by this increment.

## Verification and limitations

Mocked HTTP tests exercise transfer, replay/restart, concurrent requests, automatic
first registration, unavailable membership, feature off, forged agent authority,
invalid input, audit failure and safety-stop refusal. Existing agents retain their
permissions. Tests launch no provider, Docker or database. Browser controls use
current revisions and idempotent requests; the server remains authoritative.
Live multi-controller/provider acceptance and all later governance grants/delegation
remain pending. See [release acceptance](p0-release-gate-acceptance.md).

Automatic registration, agent creation and history attachment use the same
appointment-batch failure policy: required persistence failure latches integrity
stop; optimistic concurrency conflict alone does not. Both the ordinary settings
control and the shared switch render an unset coordination flag as off.

Handler-level regression tests exercise agent creation and history attachment
with instance-scoped mocked login/history dependencies. They verify successful
registration and initial appointment, disabled coordination, and transaction
commit failure: all registration events roll back, and a required appointment
failure latches the integrity stop and refuses a subsequent selection. These
checks do not launch a provider or prove live provider authentication.

## Ask the current coordinator

With coordination enabled, the owner can choose **Ask current coordinator** in a
room. The confirmation shows the exact coordinator identity and appointment
revision. Enter the request and submit. This is an owner-authorized bounded turn,
not a model-issued approval, task assignment or delegation. Existing worker setup,
execution barriers, permissions and quotas still apply; selecting an agent does
not make an otherwise unsupported runtime executable. Ordinary room sends and
human-authorized tasks remain independent of the coordination role.

Replacing the coordinator or disabling coordination cancels locally queued/active
role-dependent turns. Cancellation uses existing worker cleanup and reconciliation;
only fenced evidence can confirm shutdown. A stopped participant or unconfirmed
job may require normal human recovery before another turn. Separate controllers
may not receive immediate local cancellation: current-state checks refuse stale
admission/tool acknowledgments and withhold stale results, while existing worker
deadlines and reconciliation still apply. This is not proof of instantaneous
remote process termination or interruption of an effect already acknowledged.

### Role-dependent send contract

`POST /api/agents?room=engineering` accepts this owner-only extension, with the
ordinary authenticated session, CSRF token and mandatory unique `Idempotency-Key`:

```json
{"action":"send","participantId":"selected-room-agent","text":"Review the bounded plan","coordinationRevision":2}
```

The positive revision must match the current eligible appointment and exact
participant in that company/room. Broadcast and foreign/stale targets refuse;
coordination off refuses. Omitting `coordinationRevision` keeps an ordinary
independent human request. No agent API identity can use this endpoint to acquire
human permission. Acknowledged retries return their historical receipt, even
after replacement, without starting another turn or reviving authority.

The controller captures the appointment and durable disable generation, copies
only that validated dependency into the detached turn, and journals coordinator
ID, governance revision and disable generation in its normal managed job binding.
The job retains its initiating human, company, room, participant/connection,
worker/provider, deadline and applicable repository/harness policies. It gains no
resource access. Initial journal projection validates the role dependency; later
job events cannot remove or rewrite it. Each persisted disable advances the
private generation, so off/on cannot restore old work after restart.

Admission, submission, broker intent/outcome acknowledgments, native-session
updates and final result publication recheck current authority. A rejected outcome
remains unresolved rather than being presented as completed. Lifecycle cleanup
metadata may still be recorded after revocation; it cannot authorize execution.
A required persistence failure invokes the existing integrity stop, which may
also cancel independent work because audit integrity is a global safeguard.

Mocked tests exercise the actual send/transfer/flag endpoints, durable job
registration/submission, broker acknowledgment and restart projection, including
independent-work controls, queued/active revocation, off/on, forged scope and
binding tampering. Browser control tests verify the displayed identity/revision
and submitted binding. No live provider, Docker, database, browser deployment or
multi-controller runtime was exercised. Bounded delegation and agent-originated
task proposals remain later increments.

### Concurrent controller writes

Result/task completion, native-session updates and the managed `submitted`
transition carry the event count from the authority check into their append
transaction. A competing transfer, disable or stop makes that write conflict.
The controller reloads and revalidates before retrying metadata, up to three
attempts; it never retries the provider turn or external effects. Unrelated
journal changes can retry successfully. Continued conflicts withhold the result
or return a conflict without falsely declaring audit corruption. Completion may
remain unresolved and require inspection; no shutdown or success is inferred.
Actual persistence failures retain the existing integrity-stop behavior.

Deterministic mocked two-controller regressions commit transfer, disable and stop
between validation and transaction start for each path. Healthy unrelated writes,
independent results and bounded repeated-conflict controls remain covered. These
are transaction-boundary regressions, not live multi-controller acceptance.
