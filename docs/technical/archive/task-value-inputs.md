# Accepted task results as inputs

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


A task can use an actual returned value from another accepted task in the same
shared room. The controller copies the stored value and records its provenance.
A manual value bearing an execution-like label is still a manual value.
This capability grants no additional tool, provider or resource permissions.

## Configure and use

The owner enables **Settings → Features → Workflows → Accepted task inputs**
(`taskValueDependencies`, default **off**). **Task definitions** must also be on.
Both settings are enforced by the server and need no restart. With grouped
Settings disabled, use the corresponding cards directly. Members may read
eligible results in shared rooms they can access; only the owner can change
these settings or dispatch structured work.

1. Define and run a source task with a finite value output contract.
2. Inspect the actual returned value. A successful schema check is not proof
   of correctness; the human owner separately accepts the task.
3. Define the next task's named input with a compatible value schema and limit.
4. Choose **Run work**, then select an accepted task result under **Input source**.
   The preview shows the exact stored JSON and hash. **More accepted results**
   loads another page. Choose **Provide a new value** to enter manual data instead.
5. Dispatch. The server resolves the reference again; a stale picker never
   authorizes stale work. Optional inputs must explicitly be selected for use.
6. Inspect **Execution records and returned values** for copied inputs, trusted
   source identity/revision/hash and current dependency status.

For example, task A returns `{"summary":"Two synthetic findings"}` and the owner
accepts it. When [Task criterion review](task-acceptance.md) is enabled, A must
have a current explicit review before it can be selected. Task B expects a text input named `source`. Choose A's `summary`.
B receives that exact text; its immutable attempt records A's work execution,
output name, accepted task revision, definition digest and output hash.
Agent review preserves the original work output identity; a review response does
not replace it. A changed definition or a newer unsuccessful attempt cannot
silently fall back to an older successful result.

## Staleness, failure and limits

The source must still be accepted, with the same task revision, definition,
current successful work/review chain and returned value. Any source task edit,
even an unrelated assignment change, conservatively invalidates existing
bindings to that revision. Reopening, revising, rerunning or changing an upstream
source also invalidates downstream bindings transitively. A human must explicitly
dispatch again against current inputs; no automatic rerun or acceptance occurs.

Validation occurs at admission, managed-job registration/submission, each broker
tool acknowledgment and result publication. Stale inputs cannot authorize a new
broker effect or publish successful output. Source changes do not promise instant
physical termination across controllers; existing deadlines, cancellation,
mandatory cleanup and uncertainty handling remain. History retains old bytes
and provenance even when inputs become stale. Existing human task acceptance is
retained as history, but stale structured outputs cannot support a new acceptance
or make a room ready for acceptance.

Only **same-company, same-shared-room, different-task value outputs** qualify.
Personal rooms, foreign rooms, unaccepted tasks, undeclared outputs, incompatible
schemas, oversized values and missing/stale execution references refuse.
No URI, host path, file/bundle claim or external download can satisfy a dependency.
Dependencies must be acyclic; validation permits at most 16 upstream levels
and 128 visited executions across a traversal, counting shared subgraphs on each
visit. Work exceeding these bounds refuses rather than skipping validation.
Existing limits still apply: 16 named inputs, 16 KiB per value, 32 KiB aggregate
canonical input bytes, and the task's narrower limits where configured.

Switching **Accepted task inputs** off preserves all records and manual bindings,
blocks new references and revokes active dependent authority. It cancels locally
tracked dependent turns; manual structured turns retain their ordinary authority.
A durable admission generation prevents off/on from reviving an interrupted
allocation. Completed results can be used again after re-enabling only when
their entire source lineage is still current and the owner explicitly dispatches.
Changing the separate criterion-review policy also generation-fences active
dependent work; old authority cannot revive through a policy off/on cycle.
Switching **Task definitions** off additionally blocks all new structured work.

Values are retained in shared-room history and disclosed to the selected provider
when dispatched under existing destination policy. Do not supply secrets.
Acceptance and provenance establish human decisions and content identity; neither
proves that a value is true, safe to execute or meets an observed external check.
Returned text is reference data, never executable authority. No database
migration, new credential access or exception to container isolation is added.

## API and replay contract

`POST /api/task-dispatch?room=<room>` supports an input reference:

```json
{
  "taskId":"dependent-task",
  "revision":3,
  "mode":"work",
  "inputs":[{
    "name":"source",
    "source":{"executionId":"source-work-execution","outputName":"summary"}
  }]
}
```

A reference cannot include `value`, `valueJson` or `sourceVersion`. The server
resolves actual stored bytes and creates a `provenance` object on the immutable
bound input; clients cannot submit that object. The server-derived source version
is `<execution-id>/<output-name>`. Manual bindings retain their existing shape
and cannot manufacture trusted provenance with a matching label. Review and
delegated execution reuse their original bound input values and provenance.

`GET /api/task-input-sources?room=<room>&taskId=<target>&input=<name>` requires
room access and both flags. It returns up to 25 compatible current `sources`
and `nextCursor`; pass `after=<nextCursor>` for the next page. Each source has
`taskId`, `taskTitle`, `taskRevision`, `executionId`, `outputName`, `valueJson`
and `digest`. A stale/missing cursor returns 409 and requires discovery again.
Other missing resources return 404; disabled discovery returns 403. POST refuses.
Discovery checks each named contract independently of other required bindings;
all required bindings are checked together during actual dispatch.

`GET /api/task-executions` additionally includes `inputsCurrent` per history row.
This is computed when read and does not mutate the immutable attempt. `false`
means a dependency is stale, disabled or revoked; old values remain visible to
authorized room readers. The private global execution map never becomes an
unscoped state response. Personal-room execution-history audit rules still apply;
personal rooms are excluded from dependency discovery and resolution entirely.

Replay validates the exact provenance and source state at admission and before a
successful terminal outcome. Terminal failures preserve the original snapshot
without claiming output. History cannot be edited to retarget an old reference.
A durable flag generation is private control state, not a user grant.

## Verification limits

Automated coverage uses only mocked database/provider/worker boundaries and
synthetic fixtures. It exercises actual dispatch and broker acknowledgment,
source reopening, review identity, transitive/cyclic/excessive lineage, schema
compatibility, flags, permissions, pagination, exact previews and retained
history. No live provider, database, worker or browser acceptance is claimed.
Immutable file/bundle custody and independently observed artifact checks remain
unimplemented integration work. This feature does not certify a P0 release gate.
