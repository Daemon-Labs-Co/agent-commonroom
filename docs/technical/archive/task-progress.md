# Recorded task progress

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


Task progress shows what the controller has recorded about a room's tasks:
contracts, admitted attempts, returned values, input dependencies and human
acceptance. It never polls providers or workers and does not turn a saved status
or model completion claim into independent verification.

## Configure and use

The owner enables **Settings → Features → Workflows → Task progress**.
`taskProgress` defaults **off**, is enforced by the server and needs no restart
or migration. Use the Task progress card directly when Settings grouping is off.
The view appears at the top of **Tasks** in shared rooms. Authorized members can
read it; only the owner gets edit, run and review actions. Personal rooms are
excluded, preserving their separate audited history access rules.

The view loads at most 25 tasks per page. **Next tasks** and **First tasks** change
the page; **Refresh progress** reads the current records again. The summary is
explicitly **this page**, alongside the total room task count. It is not a global
completion percentage. There is no background worker probe or automatic retry.

Each task shows its saved task state separately from its recorded progress:

| Progress | Meaning |
|---|---|
| Define task | Current policy requires a seven-part definition before dispatch. |
| Legacy task | No structured delivery contract is recorded. |
| Legacy work admitted | A legacy dispatch is pending; worker activity is not observed. |
| Legacy result recorded | A legacy message/result was recorded without typed output validation. |
| Not run | A definition is saved; actual inputs are bound at dispatch. |
| Work admitted | An immutable work attempt was admitted and awaits an outcome. |
| Agent review admitted | An agent review is pending; it cannot accept the task. |
| Reconcile pending work | A pending structured dispatch lacks its expected attempt record. |
| Inputs need attention | Bound source lineage is stale, disabled, revoked or exceeds validation limits. |
| New work required | The definition changed or no current valid work result is established. |
| Owner retry required | The latest attempt failed or stopped; retry is an explicit owner action. |
| Owner review required | Actual returned values passed structural validation; the owner still judges success. |
| Criterion review required | The separate criterion-review feature requires a current explicit owner decision. |
| Accepted by owner | The owner accepted current returned values under the active acceptance rules. |

For example, the saved task state can still say **Accepted** while progress says
**Criterion review required** after explicit review is enabled. A task can have
an admitted attempt while its input dependency later becomes unusable. Showing
these separate facts avoids claiming that status alone proves delivery.

Bound dependency rows identify the input, accepted source task/revision and
immutable output identity, with current/attention status computed at read time.
Individual edges may still be current when the combined graph exceeds a total
verification budget; the overall input state then correctly refuses authority.
Old returned values remain inspectable after a source change or failed review.

**Inspect attempts and values** opens the existing scoped execution viewer.
Owner **Edit task**, **Bind inputs and run** and **Review criteria** use their
ordinary task controls and server authorization. Progress does not predict live
provider/image availability or grant a dispatch. A pending attempt never gets a
run button. If work changes after a read, stale actions refuse and require a
refresh; the eventual server mutation repeats its own revision/authority checks.

## Flags, privacy and authority

Off hides this view and refuses its API; it preserves tasks, values, reviews and
history and changes no worker or acceptance authority. The other task feature
flags still independently control structured execution, accepted input binding
and criterion decisions. Legacy records remain visible as legacy facts without
inventing contracts or converting them to validated outputs.

The endpoint returns metadata, never raw input/output values, definitions,
human review notes, provider stderr, credentials or worker launch details.
Task titles are existing room metadata and are escaped in the view. Full values
and reviews remain in their separately authorized viewers. A read does not write
audit events, launch an agent, accept work, change a source or retry anything.
Existing authorization, audit integrity, retention protections, mandatory
container isolation and cleanup requirements remain unchanged.

## Technical interface

`GET /api/task-progress?room=<room>` requires authenticated room access, a shared
room and `taskProgress=true`. It returns:

- `tasks`: up to 25 metadata rows, ordered by task ID;
- `pageCounts`: counts for only the returned rows;
- `totalTasks`: the number of tasks in this room, excluding foreign rooms;
- `nextCursor`: pass as `after=<cursor>` for the next page;
- `structuredDispatchEnabled` and `criterionReviewEnabled`: feature policy,
  explicitly excluding provider/worker availability;
- `observations`: always `controller-recorded-state-only`.

A row includes task ID/title/revision/saved status and dispatch status,
definition revision, phase/detail, latest attempt mode/recorded status,
original work ID/current returned-value count, input state, captured dependency
identities/current status and latest criterion-review revision. Returned-value
count identifies stored structurally validated work; it does not mean tests
passed or facts were verified. An active or failed review still points to its
original work values when that record exists.

Disabled requests return 403; unavailable/foreign/personal rooms return 404;
a missing/stale cursor returns 409 and requires reloading. POST refuses with 405.
Only the requested page's dependency graphs are evaluated, using the existing
finite depth/node budgets. No derived progress is persisted into immutable tasks
or added to room acceptance snapshots. Whole-state replay remains the existing
read-scaling limitation; this view does not claim indexed projections.

## Verification limits

Automated tests use synthetic data and mocked DB/provider/worker boundaries.
They cover actual dispatch through pending/validated/accepted states, explicit
criterion requirements, stale inputs, failed reviews, legacy handling, immutable
derivation, bounded pagination, metadata-only privacy, flags, permissions,
escaping and stale UI responses/actions. No live browser screenshot or live
worker/provider/database verification is claimed. File/bundle custody and
independently observed runtime checks remain separate work; this view cannot
certify a P0 release gate.
