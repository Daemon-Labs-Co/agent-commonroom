# Room outcomes and acceptance criteria

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
not require a bulk migration. The room settings form asks for both fields when
adding its definition. Personal rooms are exempt. Only the owner can create or
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
agent permissions, change existing dispatch eligibility, or add outcome text to
provider prompts. Those execution and completion integrations are separate work.
Changing the room definition does not rewrite approved task scopes or cancel
running work. Human acceptance and existing task/delegation authorization remain
in force. Definitions are not secrets or authority credentials; room access
controls govern who can read them. Mandatory container isolation has no off
switch.

Automated verification uses mocked storage and synthetic data. It covers flag
states, owner/member permissions, required fields and limits, stale saves,
legacy/personal rooms, omitted-field preservation, replay/restart and escaped UI
rendering. It does not establish live provider or browser acceptance.
