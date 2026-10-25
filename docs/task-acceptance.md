# Task criterion review

Task criterion review records the human owner's judgment against a task's exact
returned values. Every human criterion needs a passed/failed decision and a
rationale. Schema criteria are assessed by the controller from the stored value;
a model's statement that a check passed never becomes that assessment.

## Configure and use

Enable **Settings → Features → Workflows → Task criterion review** as the owner.
`taskAcceptanceReview` defaults **off**, also requires **Task definitions**, is
enforced on the server and needs no restart or migration. Use its Settings card
directly when grouping is disabled. This feature covers structured tasks in
shared rooms; legacy tasks retain their existing owner acceptance controls.
Personal rooms are excluded from this additional execution-history read path.

1. Define the task's seven parts and explicit criteria. Each criterion names an
   output and uses either `human` or `schema` assessment.
2. Run work and inspect the structurally validated returned values. An optional
   agent review remains advice; it does not make the human decision.
3. Choose **Review task criteria** on the task. The dialog shows the current
   definition, exact bound inputs, exact returned values, immutable output IDs
   and hashes, and every criterion.
4. For each human criterion choose **Passed** or **Failed** and explain what you
   checked. No human criterion starts as passed. Read the server's structural
   results for schema criteria; you cannot override those through the form.
5. Choose **accept** or **changes** and provide an overall rationale. Acceptance
   requires all criteria passed. A changes request may record additional owner
   concerns even if all individual criteria passed.
6. Save. The review and task-state update commit together. Acceptance makes the
   task Accepted; a changes request makes it Changes requested. Neither launches
   an agent or retries work. Dispatch is still a separate authorized action.

For example, a summary task has a human criterion for fidelity to its source and
one schema criterion for a text output. The owner compares the displayed input
and output, records the fidelity decision, and sees the controller's text-schema
assessment. A syntactically valid string does not establish fidelity or truth.
An output omitted under an optional output contract fails a schema criterion
that explicitly requires assessing that output; absence is never inferred as a
successful check.

Members can view authorized shared-room reviews, but cannot record decisions.
History is paginated ten records at a time. Each historical review includes its
own original work attempt and definition, so later work does not replace the
values the owner actually reviewed. Text and exact JSON previews are escaped;
large integers are not rounded by the browser. Notes are retained room history:
do not enter secrets or credentials.

## Stale work and switching off

A decision binds the task revision, definition, current work/review chain,
immutable inputs/outputs and relevant policy generations. Changed work, source
lineage, assignments or policy invalidate an open dialog. The owner must reload
and assess the current snapshot; resubmitting stale decisions refuses. A stored
review authorizes acceptance only for the exact post-decision task revision.
Editing an accepted structured task through the task editor explicitly reopens
it as Changes requested when criterion review is on. The editor does not offer
Accepted as a substitute for the criterion form.

Turning the feature on preserves existing task state and review history. However,
previously accepted structured tasks need a current explicit review before they
can serve as accepted value inputs or make a room ready for acceptance while
this feature is enabled. Review an already Accepted task to establish that
current decision; no synthetic criterion judgments are generated during upgrade.
Upstream tasks should be reviewed before their dependent tasks.

Off preserves all review history and returns to ordinary owner acceptance with
current structurally valid outputs. It never allows a model or member to accept
work, bypass permissions, omit audit persistence or revive a worker. Disabling also revokes active turns that depend on accepted
task values; manual structured turns retain their ordinary authority. Off/on
changes the durable review generation, so an old open dialog cannot become
current again. Re-enabling requires fresh review for current acceptance. Turning
Task definitions off also disables new criterion decisions. Manual execution
without value dependencies is not invalidated merely by cycling the separate
Accepted task inputs setting.

Room acceptance remains a separate human decision about the entire deliverable.
A stale/unreviewed structured task cannot make that room ready while criterion
review is enabled. Task acceptance does not grant tool, provider, resource or
workspace disposal authority and cannot bypass mandatory container isolation.
Independently observed build/test/artifact checks remain distinct evidence; an
owner note is a human claim, not manufactured runtime observation.

## Technical interface

`GET /api/task-acceptance?room=<room>&taskId=<task>` requires room access and a
structured task in a shared room. It remains readable with flags off. It returns
`enabled`, `ready`, `snapshot`, current `task`, `workExecution`, read-only
`schemaAssessments`, latest `review`, up to ten `reviews`, `reviewedExecutions`
keyed by those records' work IDs, `nextCursor` and criterion-review `status`.
Use `before=<nextCursor>` for older reviews; the returned current work is labelled
separately from the original attempts attached to historical reviews. Missing
resources/personal rooms return 404; invalid cursors return 400.

`POST /api/task-acceptance?room=<room>` requires authenticated owner access, CSRF,
a bounded strict JSON body and an `Idempotency-Key`:

```json
{
  "taskId":"task",
  "taskRevision":7,
  "reviewRevision":2,
  "snapshot":"<current 64-character snapshot hash>",
  "action":"accept",
  "note":"Compared the declared source and returned summary",
  "decisions":[{
    "criterionId":"faithful",
    "decision":"passed",
    "note":"Each stated finding matches the supplied source"
  }]
}
```

`decisions` contains every human criterion exactly once and no schema criteria.
Use an explicit empty list when every criterion is server-assessed. Each human
rationale is nonempty and at most 1,000 Unicode code points; overall rationale
is nonempty and at most 2,000. There are at most 32 criteria and the request is
at most 64 KiB. Unknown/duplicate fields, malformed Unicode, excessive nesting,
unknown criteria, duplicate decisions and spoofed schema assessments refuse.
`action` is `accept` or `changes`. A mismatched task/review revision, snapshot,
source lineage or policy returns 409; disabled decisions return 403.

The controller appends `task.acceptance.reviewed`, `task.revised` and the operation
receipt in one audit transaction with an event-count fence. A persistence failure
commits neither decision nor status, withholds success and latches integrity stop;
a concurrent state change returns 409 without forcing a review through. Exact
idempotent replay acknowledges the original operation without another review.
The browser generates one operation key for each review save, including both
acceptance and changes requests, and retains it when refreshing an expired CSRF
token and retrying that same request. The task ID is part of the fingerprint, so
a key cannot replay a different task.
Generic event requests cannot inject review records or accept structured tasks
in place of criterion review while the feature is enabled.

Review records bind company/room/task, monotonic review revision, pre-decision task
revision, current dispatch and actual work execution IDs, content snapshot,
criterion methods/decisions/output hashes/rationales, action, owner and time.
Replay rechecks authority, current output/source identity, criterion coverage,
server-computed structural results and transitions. Acceptance's own status and
revision change do not invalidate its content snapshot: the exact pre/post task
revisions are fenced separately. See [the decision](adr/0003-human-task-acceptance.md).
The global review map stays private; no unscoped state response includes it.

## Verification limits

Tests use synthetic fixtures and mocked database/provider/worker boundaries.
They exercise atomic acceptance, changes requests, idempotency, restart/replay,
criterion coverage, structural failures, stale versions/policy, permission and
privacy refusals, pagination with original work, rollback/integrity stop and
multi-controller commit conflicts. UI checks cover escaped values, exact numbers,
explicit choices, read-only history and stale room/view responses. No live
browser screenshot, database, provider or worker acceptance is claimed. Verified
file/bundle custody and independently observed runtime checks remain integration
gaps; this capability does not certify a P0 release gate.

The optional [recorded task progress](task-progress.md) view highlights structured tasks requiring current criterion review while preserving their historical saved state.


Acceptance metadata and current lineage are checked separately. Matching a saved
review snapshot does not mean its upstream sources are current. Source selection
checks recorded decision metadata without starting another dependency traversal;
acceptance readiness, input resolution and output eligibility then share one
bounded lineage walk (16 upstream levels, 128 visited executions, including
repeated DAG branches). Cycles, stale source revisions and over-budget graphs
refuse. Criterion-review on/off never disables dependency integrity bounds.
Tests include healthy maximum-depth chains and synthetic corrupt cyclic/deep/wide
records; those defensive fixtures are not production acceptance evidence.
