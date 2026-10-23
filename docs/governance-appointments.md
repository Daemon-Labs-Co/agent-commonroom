# Room coordination appointments

This increment implements durable initial appointments and explicit human-selected
replacement independently of the model provider. It does not implement agent
coordination grants, automated delegation, task extraction, consensus or a new
execution path. A coordinator record grants no extra permissions; existing human
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
jobs continue under their own permissions and safety stops. No role-derived agent
execution grants exist in this increment, so no such grants can survive transfer.
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
Future coordination requests must bind the governance revision and revalidate at
the effect boundary before agent-driven coordination can be enabled. This increment
does not certify GOV-06/07/08/09/13 or the full P0 production execution gates.

## Verification and limitations

Mocked HTTP tests exercise transfer, replay/restart, concurrent requests, automatic
first registration, unavailable membership, feature off, forged agent authority,
invalid input, audit failure and safety-stop refusal. Existing agents retain their
permissions. Tests launch no provider, Docker or database. Browser controls use
current revisions and idempotent requests; the server remains authoritative.
Live multi-controller/provider acceptance and all later governance grants/delegation
remain pending. See [release acceptance](p0-release-gate-acceptance.md).
