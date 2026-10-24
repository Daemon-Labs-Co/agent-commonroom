# Conversation task proposals

Task proposals separate a conversation suggestion, human review, draft task,
dispatch, observed outcome and acceptance. A message saying “approved” is still
message content. This increment implements **owner-mediated capture and review**;
it does not parse conversations automatically or expose an agent-authenticated
proposal/delegation channel. The governing agent has no special approval power.

## Enable and use

The owner opens **Settings → Task proposals** and enables `taskProposals`.
The flag defaults off, is persisted through the owner-only feature settings
endpoint, and takes effect immediately without restart. Both the ordinary control
and shared settings switch show an unset flag as off.

In **Tasks → Conversation proposals**, choose **Capture a suggestion**, select a
message from the current room, and enter a suggested title and optional criteria.
The proposal stores the original message ID and recorded author/actor fields,
alongside the owner who captured it. The title/criteria are your interpretation;
they do not replace the source. Empty criteria can capture an ambiguous suggestion,
but cannot approve it into a task until reviewed acceptance criteria are supplied.

Choose **Review and create draft** to inspect the original source, revise the
reviewed title/criteria and confirm. This records a human decision and creates
one unassigned **Draft** task atomically. It does not assign an agent, authorize
resources, start a provider, accept work or bypass execution restrictions.
Assignment, changing the draft status and explicit dispatch remain separate
human actions through existing task controls. Draft tasks cannot be dispatched.
A reviewed task retains its proposal ID through later ordinary task edits.

Choose **Reject proposal** to record a terminal rejection while preserving the
suggestion and original message. A proposal receives only one decision. To
reconsider a rejected suggestion, capture a new proposal explicitly. Multiple
intentional proposals may refer to one message; retries of the same operation do
not create duplicates.

Turning the flag off refuses new captures and decisions. It preserves proposals,
original messages, approved tasks and audit history; retained proposals remain
readable to the owner in their room. It does not cancel or change existing tasks
or independent jobs. Existing task/dispatch feature controls and all authentication,
permissions, audit, retention, quotas and mandatory isolation remain enforced.
Re-enabling resumes capture/review of still-pending proposals without an election,
new approval or automatic dispatch. Human stop and recovery stay available.

## Permissions and source safety

Only an authenticated enabled owner can capture, list or review proposals. Server
checks company and room access, CSRF for writes, current feature state, audit
integrity and safety stops. Foreign-room source/proposal IDs refuse without
returning their content. Source identity comes from the stored message, never a
client-supplied author, actor, governor label or claimed approval. Where an older
message lacks actor metadata, those fields remain unrecorded; the application
does not invent attribution. A captured source's author is not the approving human.

The original source is displayed as escaped, untrusted text. Summarized or revised
criteria remain separate from that original. A missing source blocks review rather
than silently approving an unverifiable summary. This feature adds no retention
exception, history rewrite, agent login or credential access. It does not prove
that the model's suggestion is truthful or that the human review is correct.

## API contract

`GET /api/task-proposals?room=engineering` returns `enabled` and `proposals`.
Each entry includes `proposal`, `source` and `sourceAvailable`, filtered to the
requested company/room and sorted by proposal ID. The full proposal map is a private
controller projection, excluded from generic `/api/state` serialization.

All POST actions require the owner session, CSRF token and a unique
`Idempotency-Key`. Unknown/duplicate fields, trailing JSON and action-inapplicable
fields refuse. The request body is limited to 8192 bytes, title to 200 UTF-8 bytes
and criteria to 4000 UTF-8 bytes. Required strings cannot be whitespace-only.

Capture:

```json
{"action":"capture","sourceMessageId":"message-id","title":"Suggested task","criteria":""}
```

Approve:

```json
{"action":"approve","proposalId":"proposal-id","revision":1,"title":"Reviewed task","criteria":"Explicit acceptance criteria"}
```

Reject:

```json
{"action":"reject","proposalId":"proposal-id","revision":1}
```

Responses contain the resulting proposal. Initial status/revision are `Proposed`/1.
Approval or rejection produces `Approved` or `Rejected`/2, with `reviewedBy`;
approval also links `taskId`. Reusing a key with the same acknowledged input returns
the same response without another decision/task. Changed input under that key
conflicts. Another decision after a terminal decision or a stale revision returns
409. Off returns 403 even for historical mutation retries; use GET to inspect
retained history. A competing disable/stop/journal update can conflict; reload and
review before a fresh attempt. Never infer success from silence or an uncertain
response.

`task.proposal.captured` journals immutable source metadata and capture identity.
`task.proposal.approved` contains the expected proposal revision and the complete
reviewed Draft task; projection applies the decision and draft together from that
one atomic fact. `task.proposal.rejected` records a terminal decision without a
task. Operation receipts commit in the same transaction under the expected event
count. Required persistence failure rolls back and latches integrity stop;
optimistic conflicts alone do not. Journal replay validates owner/feature state,
source/scope, revision and task shape; task revisions cannot strip or forge the
proposal link. Generic event commands cannot inject proposal decisions.

## Verification and remaining work

Mocked tests exercise actual authenticated capture/decision/read/dispatch/settings
endpoints, retries, competing controllers, disable races, source scoping,
authorship forgery, audit rollback, rejection, preserved task provenance, default
off and feature toggling. UI tests cover source escaping, explicit draft wording,
room changes and request fields. No database, provider or Docker is launched.

This is part of GOV-08, not completion of the full governance security model or
a P0 release gate. Automatic agent-originated proposal submission, bounded
delegation, resource approvals, dependency handling, independent objections and
live multi-controller/provider/browser acceptance remain separate work. See
[governance design](governance-model.md) and [release acceptance](p0-release-gate-acceptance.md).

Personal-room proposals use the same capture/review contract. During historical
replay the controller first derives personal rooms from the owner and the actors
or connections already recorded at that point, then validates proposal scope.
Capture/decision replay does not depend on earlier profile or Help preference
changes. Missing actors/connections and ones introduced only later cannot make
an invalid historical personal-room proposal valid. Existing room access checks
and source scoping remain enforced.

Automatic personal-room labels follow the current actor name and refresh at
final replay, even if an earlier proposal caused the room to be materialized.
Changing the room name in Room settings saves a custom name that survives refresh,
restart and later actor renames. Saving other settings without changing an
automatic name keeps it automatic; later settings saves preserve a custom name.
Connection-derived agent display names follow the current connection name.
Renames change those display labels without changing room IDs, personal-room
ownership, saved room settings, actor permissions or recorded source authorship.

The controller reconstructs this naming distinction from `room.updated` history;
it does not trust a client-supplied naming-mode flag. Personal rooms are derived
from actors/connections already recorded before applying their historical updates.
Only automatic labels refresh. Custom names, room identity, access and settings
remain intact whether the room update precedes or follows proposal capture.

A personal-room settings save must also match the name shown when the dialog
opened. Actor renames can update automatic labels without changing the room's
saved revision. `/api/state` therefore includes a `nameVersion` display-state
precondition for personal rooms; include it unchanged in `room.updated` alongside
the normal revision. It is not an authentication or authorization credential.
A missing or outdated value returns 409 with a reload instruction and writes
nothing. Refresh and reopen settings before saving. Unchanged custom labels keep
the same name precondition through actor renames; ordinary room revision checks
still protect other settings. Existing historical room events remain replayable.
