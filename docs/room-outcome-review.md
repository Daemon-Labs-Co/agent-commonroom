# Human acceptance of a room deliverable

Room outcome review records the owner's judgment that the shared room has
achieved its defined result. It is a review of a particular saved snapshot,
not a model decision or an execution capability.

## Configure and review

The owner can enable **Settings → Features → Workflows → Room outcome review**.
`roomOutcomeReview` defaults **off**, also requires `roomOutcomes`, and needs no
restart. With tab grouping disabled, use the Room outcome review card directly
in Settings. Authorized members can read review status and notes; only the owner
can accept or reopen a deliverable.

1. Define the shared room's outcome and acceptance criteria in Room settings.
2. Finish and human-accept its tasks using the existing task review process.
3. In the Room or Tasks view, choose **Review and accept deliverable**.
4. Inspect the current definition, task snapshot and saved evidence shown in
   the dialog. Each task includes its own acceptance criteria; evidence is labeled
   with its task title/ID or as room-level evidence. Historical associations
   missing from the current task set are labeled explicitly. Room names and
   snapshot text render as text, including names containing angle brackets.
   Independently verify any artifacts or claims; the application
   does not run acceptance tests or inspect external links for you.
5. Record the evidence reviewed and your acceptance rationale, then save.

The room needs at least one task. Every task in that room must be **Accepted**,
with no pending task dispatch. These checks make a review available; they do not
prove that the deliverable works. The owner must make that judgment. A nonempty
review note is required, limited to 4,000 Unicode characters. A saved evidence
record is useful but not mandatory: the note can describe evidence assessed
outside the application. Personal rooms cannot receive deliverable decisions.

An accepted room shows **Accepted** only while its snapshot still matches. Any
room revision, task change/addition, or saved evidence change shows **Needs
review**. This conservatively includes a room name, description or presence edit;
a later reversion of the definition does not revive acceptance because the room
revision has advanced. Choose **Reopen deliverable** with a reason to record the
need for further work, or accept a fresh snapshot once the requirements are met.
Reopening does not launch an agent or change task assignments/statuses.

Changing either feature flag off preserves review history and makes decision
controls unavailable. Existing acceptance remains readable, with staleness still
computed from current work. Reenabling restores decisions; it does not generate
one automatically. Repeated submission with the same operation key returns the
original committed decision. If work changes while a review dialog is open, save
returns 409; reload and inspect the new snapshot before deciding.

## API and durable record

`GET /api/room-outcome-review?room=<id>` requires authenticated room access. It
returns `enabled`, `ready`, `status`, the latest `review`, current `definition`,
room-scoped `tasks` and `evidence`, and an opaque `snapshot` digest. The dialog
uses this response rather than an older cached definition. Review notes are
visible to authorized readers of the room; do not put secrets in them.

Owner-only `POST` to the same endpoint requires CSRF and an `Idempotency-Key`.
Its exact JSON fields are:

```json
{"action":"accept","revision":0,"snapshot":"<digest from GET>","note":"Evidence checked and acceptance rationale"}
```

Use `action: "reopen"` to reopen an existing acceptance, including a stale one.
`revision` is the latest review revision, initially zero. A strict bounded parser
rejects additional/duplicate fields, invalid actions and missing values. The
server selects company, room, human reviewer and review time. Client claims of
reviewer identity are not accepted. Generic `/api/events` writes cannot inject
this event.

Each successful decision appends `room.outcome.reviewed` and its operation
receipt atomically. The record contains company/room, monotonic review revision,
Accepted/Reopened status, snapshot digest, note, owner identity and UTC time.
Existing audit history retains earlier decisions; the UI displays the latest.
Replay validates human provenance and the historical transition. The private
review map is not exposed through unrelated room state.

The digest binds company/room, room revision, outcome/criteria, and sorted full
room task/evidence records. It grants no authority. Review revision and digest
must match; transaction-level event-count fencing rejects concurrent changes,
including another controller disabling the flag. Persistence failure cannot
report acceptance and trips the existing integrity stop. Existing safety/audit
stops must be investigated before a new decision can be recorded.

## Boundaries and verification

Acceptance does not expand permissions, dispatch agents, skip audit checks,
prove worker shutdown, certify provider behavior, dispose of data or bypass
retention. It does not automatically validate artifacts or map each task to a
specific criterion. Those remain human review responsibilities. Existing
mandatory isolation and all resource authorization remain independent.

Tests use mocked storage and synthetic tasks/evidence. They cover acceptance,
reopening, restart, idempotency, stale snapshots, concurrent flag changes,
missing/incomplete/pending tasks, personal rooms, disabled flags, permissions,
CSRF, malformed inputs, persistence failures, escaped UI and stale-room guards.
No live browser, provider, worker or database acceptance is claimed.
