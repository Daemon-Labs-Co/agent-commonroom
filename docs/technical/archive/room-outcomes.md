# Room outcomes and acceptance criteria

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


A shared work room should explain what the team is trying to deliver and how a
human will judge the result. Its **Defined outcome** states the result; its
**Acceptance criteria** describe the evidence or conditions that would satisfy
it. Tasks are the steps toward that result, and agents participate in the room.

For example, an outcome could be “Publish the first installation package.” Its
criteria could be “A clean-machine installation succeeds, the documented example
runs, and the owner reviews the package and release notes.” Avoid vague goals
such as “Work on the project.”

## Enable and use

Only the owner can enable **Settings → Features → Workflows → Room outcomes**.
The `roomOutcomes` flag defaults **off** and changes apply on the next state load
without restarting. With Settings/feature tabs disabled, use the Room outcomes
card directly in Settings.

When enabled:

1. Create a shared room with its name, defined outcome and acceptance criteria.
2. For an existing room, use **Room settings** to define or revise those fields.
3. Read the deliverable in the Room or Tasks view while discussing, assigning
   tasks and assessing evidence.

Existing rooms remain usable without an outcome, so enabling the feature does
not require a bulk migration. The room settings form allows both fields to stay blank on an undefined
existing room, so a name or description edit does not require migration. Once
you add a definition, both fields must be filled; new rooms also require both. Personal rooms are exempt. Only the owner can create or
edit a room definition; authorized members can read it. Agent connections and
models cannot revise it through the room event API.

Switching off preserves all definitions and history. Saved definitions remain
visible, but edits to them and new nonempty definitions are refused by the
server. Ordinary room edits preserve saved fields even if an older client omits
them. Turning the feature back on restores editing and new-room requirements.

## API and persistence

`room.created` and `room.updated` accept `outcome` and `acceptanceCriteria` string
fields in their room payload. They are included in authorized room state and
persist in the existing append-only event log. Use the current room `revision`
for updates; a stale revision returns 409. Omitted fields on updates preserve
saved values. Whitespace is trimmed. The outcome limit is 2,000 Unicode
characters and the criteria limit is 8,000.

With the feature enabled, new shared rooms need both nonempty fields (400 if
missing, blank or oversized). Existing legacy rooms can receive unrelated updates
without defining an outcome. A defined outcome cannot be cleared through a later
update while enabled. Personal-room definition changes return 400. Changes while
disabled return 403; resubmitting unchanged values is allowed. Ordinary owner,
authentication, CSRF, event integrity and optimistic concurrency checks remain.
The owner sets the flag using `feature.flag.updated` with `name: "roomOutcomes"`.

## Meaning and limits

This increment records and displays human intent. It does not automatically link
tasks to individual criteria, mark a room complete, evaluate evidence, expand
agent permissions, change existing dispatch eligibility, or expand tool access. Optional task-provider context is described below; completion
and automated evidence assessment remain separate work. Optional [human acceptance](room-outcome-review.md) can record the owner’s judgment about a saved snapshot.
Changing the room definition does not rewrite approved task scopes or cancel
running work. Human acceptance and existing task/delegation authorization remain
in force. Definitions are not secrets or authority credentials; room access
controls govern who can read them. Mandatory container isolation has no off
switch.

Automated verification uses mocked storage and synthetic data. It covers flag
states, owner/member permissions, required fields and limits, stale saves,
legacy/personal rooms, omitted-field preservation, replay/restart and escaped UI
rendering. It does not establish live provider or browser acceptance.


## Optional task-provider context

The owner can separately enable **Settings → Features → Workflows → Room
deliverable context** (`roomOutcomeContext`, default **off**). This is an explicit
export opt-in: dispatched task work and review may send the current shared-room
outcome and acceptance criteria to the assigned provider. `roomOutcomes` must
also be enabled. No restart is needed. Agents and invited members cannot change
either setting.

The controller includes only the dispatched room's definition, labeled as
untrusted reference with its room ID and room revision. It does not replace the
task's own acceptance criteria. The room's allowed-provider policy remains
mandatory, and **Exclude references** also excludes this room definition.
Personal rooms and undefined legacy rooms contribute no definition. Ordinary
chat and workflow turns are unchanged by this setting.

The complete prompt still has the 128 KiB limit. Oversized historical definitions
or a combined-context overflow refuse dispatch before writing its intent or
starting a worker. Metadata is bounded before serialization. Definitions are
encoded as JSON reference content, so text cannot break its structural wrapper;
the model still receives untrusted text, and actual authority remains enforced
by the controller and tool broker.

A queued dispatch uses the snapshot built at that dispatch. Later room edits or
switching either flag off affect future dispatches; they do not change already
queued prompts, recall provider context, cancel active work, accept a task, or
rewrite delegated task scopes. Retrying an acknowledged dispatch with the same
idempotency key does not create a new turn or refresh its snapshot. Shared-room
access, approved repository boundaries, audit acknowledgment and mandatory
container isolation remain unchanged.

Tests exercise work/review dispatch, idempotent replay, room separation, both
flags, provider denial, reference exclusion, personal-room exclusion, historical
and combined-size refusal, and owner-only configuration using synthetic data and
mocked storage/turn boundaries. No live provider or worker acceptance is implied.
