# Canonical task definitions and execution records

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


A task definition has exactly seven parts:

1. **Task:** the specific work to perform.
2. **Expected input contract:** named, typed inputs needed for that work.
3. **Process:** a prescribed procedure or bounded objective, with instructions
   and descriptive constraints.
4. **Desired output contract:** named objects or values that must be returned.
5. **Acceptance criteria:** conditions for assessing those outputs.
6. **Execution constraints:** permitted tool operations and bounded execution.
7. **Failure contract:** what happens when input is missing or execution fails.

The definition is versioned. Actual input values, attempts, results and human
acceptance are separate. A **deliverable** is an immutable returned object that
satisfies its output contract; an agent saying “done” is not one. Structural
validity, independently verified checks and human acceptance are different facts.
See the [canonical vocabulary](../../../GLOSSARY.md) and [decision](adr/0002-task-contracts-and-executions.md).

## Configure and use

The owner enables **Settings → Features → Workflows → Task definitions**.
`taskDefinitions` defaults **off**, is enforced on the server and needs no restart.
With Settings grouping disabled, use the Task definitions card directly.

When enabled, the owner creates or edits tasks using the seven-part editor. Name
inputs and outputs with short identifiers such as `source` or `summary`; those
names connect input bindings, output properties and criteria. Declare no inputs
with an empty input list when none are needed. At least one output is required.
A process can prescribe instructions or let an agent choose steps within a
bounded objective. Descriptive process instructions do not mechanically prove
that the steps were followed and never grant tools or resources.

For example: a task reads a supplied text value, summarizes it, and returns
`{"summary":"..."}`. The expected input is `source` under `text.v1`; the desired
output is `summary` under `text.v1`; a human criterion checks fidelity to the
source. Before dispatch, the owner supplies the actual source value and source
version label. Those bindings belong to this execution, not the definition.

The assignee and reviewer remain separate task assignments. Only the owner may
change a structured definition or its scope. Members can read the definition and
apply their existing permitted task-state changes; they cannot accept work or
revise the contract. An owner can accept a structured task only after its latest
work attempt has structurally valid outputs, optionally followed by a successful
agent review of that same work. An agent review is not human acceptance. The
separate default-off [Task criterion review](task-acceptance.md) capability adds
explicit per-criterion owner decisions tied to those exact returned values.

### Finite data types

- **`value`:** a bounded JSON value under a registered schema.
- **`file`:** a declared file name, media type and positive byte limit.
- **`bundle`:** an explicit bounded member manifest with safe relative file
  names, media types and per-member limits. Total member limits cannot exceed
  the bundle limit; duplicate/overlapping or traversal/absolute names refuse.

Registered value schemas are `text.v1`, `integer.v1`, `number.v1`, `boolean.v1`,
`object.v1` and `array.v1`. Integers use signed 64-bit decimal integer syntax;
numbers must be finite. Object contracts declare one to 32 named primitive
properties with required/optional status and forbid undeclared properties.
Array contracts declare a primitive item schema and a maximum of 1–256 items.
Arbitrary schemas, recursive shapes and unrestricted objects/arrays are refused.
All values have a byte limit no greater than 16 KiB. A value can be an empty text,
empty array or an object omitting optional properties where its schema allows.
Required means a named input/output must be present, not that every value is
nonempty. JSON `null` does not satisfy these schemas.

The current executable path supports **value inputs and value outputs**.
File/bundle declarations are valid definitions, but required file/bundle inputs
or outputs block dispatch before effects until independently verified immutable
artifact storage is connected. A host path or model claim cannot satisfy them.
Optional file/bundle inputs may be omitted; a supplied unverified object claim
still refuses. No native filesystem, URI download or provider fallback is added.

### Actual inputs, outputs and reviews

Manual bindings carry a name, actual JSON value and human-provided source-version
label. The label is not independently verified provenance. With the separate
default-off [Accepted task inputs](task-value-inputs.md) capability, the owner can
instead select an accepted task’s verified returned value in the same shared room. The stored content
hash identifies the exact immutable value copy. Do not enter credentials or
secrets: values are retained in the room's history and sent to the selected
provider when the owner dispatches work. Non-text JSON inputs are transmitted
as JSON text so the browser cannot round large integers before validation.

A work response must be exactly one JSON object keyed by desired output names.
Markdown fences, undeclared outputs, missing required values, wrong types,
unknown object properties, invalid encodings or oversized data fail validation.
Values are canonically encoded before hashing and storage; both supplied and
canonical sizes are bounded. Canonical JSON and hashes survive replay without
browser numeric precision loss. `valueJson` preserves exact canonical text for
inspection. An output's identity is `<execution-id>/<output-name>`.

The immutable definition and bound inputs are recorded with the admitted attempt
before a provider turn is queued. Admission is not proof that a worker actually
started. Final values are stored only with the durable terminal task outcome.
A failed outcome commit withholds success and preserves pending work for explicit
reconciliation. No automatic retry or manufactured delivery occurs.

**Execution records and returned values** on a task opens paginated history
(ten records per page). It shows the definition revision, actual bindings,
returned values and hashes. The `validated` status means structural output
validation, not verified facts, passing tests or human acceptance. `running`
means an admitted attempt awaiting a durable result, not independently observed
worker activity. A separate review uses the original work's bound inputs and
actual returned values, including when a different participant reviews it.
Replacement review inputs refuse. Review findings are model-provided text;
they do not become a work deliverable or auto-accept the task.

A delegated structured task inherits its original parent's exact input values
and versions. A replacement input set must match; human reassignment does not
expand the approved work's input scope. The existing deadline/repository/operation
intersection and no-subdelegation restrictions still apply.

## Bounds, authority and failure

Definitions use schema version 1 and monotonic definition revisions. A semantic
contract edit increments that revision; ordinary task-state changes do not.
The task revision still protects all saves. Pending or delegated definitions
cannot change. To revise an accepted definition, first move the task to an
unaccepted state. Older clients omitting a definition preserve it; explicit
removal/downgrade is refused. Criteria displayed on a structured task are derived
from its canonical criterion list rather than a second editable text field.

Execution limits are restrictive intersections, never grants:

- operations are a subset of `read`, `write`, `patch`, `command`, possibly empty;
- only the assigned approved task workspace is in scope; existing repository,
  room, participant and tool authorization still apply;
- duration is 1–180 seconds, also bounded by any parent delegation deadline;
- returned response is 1–16,384 bytes;
- at most 1–64 tool requests, also bounded by the broker's own limit;
- human owner alone can accept completion.

Existing worker CPU/memory/PID/account quotas, mandatory container isolation,
provider disclosure rules, audit acknowledgments and recovery constraints remain.
Task text cannot authorize an external effect or defeat any of those checks.
Explicit task inputs are part of the owner-dispatched task request; excluding
imported reference material does not silently remove required task inputs.
The provider must still be allowed by the room's destination policy.

The version-1 failure contract is deliberately finite: missing inputs **block**;
failed execution is **recorded and stopped**; retry is **human-only**. Contract
violations withhold success and do not create returned outputs. Successful
schema validation never auto-accepts work or certifies observed checks.

Switching the feature off preserves definitions, inputs, outputs and history.
It blocks new structured dispatch, cancels locally tracked structured turns and
revokes their effect/result authority. Off/on cannot revive an old allocation:
admission generations are durably fenced. Other controllers deny stale authority
at use; immediate physical termination across controllers is not promised.
Worker cleanup and uncertainty handling remain mandatory. Legacy independent
work keeps its existing permissions. Switching on requires a definition before
legacy tasks can be dispatched; proposed legacy drafts can be defined gradually.
No automatic conversion invents input/output contracts for old work.

## Technical interface

`task.revised` accepts an optional `definition` object. Owner authentication,
CSRF, task revisions and audit integrity checks remain. Canonical definition
JSON rejects unknown/duplicate fields, malformed Unicode, trailing data and
excessive nesting. The definition is at most 48 KiB, with at most 16 inputs,
eight outputs, 32 criteria, 16 descriptive process constraints and 32 bundle
members per contract. Text lengths and aggregate input/context limits are
checked before allocation/dispatch. No database migration is required; current
and historical definitions live in append-only task events.

A minimal definition is:

```json
{
  "schemaVersion": 1,
  "revision": 1,
  "task": "Summarize the supplied source",
  "expectedInput": [
    {"name":"source","type":"value","required":true,"schemaId":"text.v1","maxBytes":4096}
  ],
  "process": {"mode":"objective","instructions":"Summarize faithfully","constraints":[]},
  "desiredOutput": [
    {"name":"summary","type":"value","required":true,"schemaId":"text.v1","maxBytes":4096}
  ],
  "acceptanceCriteria": [
    {"id":"faithful","description":"Faithful to source","method":"human","output":"summary"}
  ],
  "executionConstraints": {
    "allowedOperations":[],"maxDurationSeconds":60,"maxResponseBytes":8192,
    "maxToolCalls":1,"completionAuthority":"owner","resourceScope":"assigned-task-workspace"
  },
  "failureContract": {"missingInput":"block","executionFailure":"record-and-stop","retry":"human-only"}
}
```

`POST /api/task-dispatch?room=<id>` retains `taskId`, `revision`, `mode` and adds
optional `inputs`. Each input has `name`, `sourceVersion`, and either JSON `value`
or exact JSON text `valueJson`, never both. Requests are bounded at 64 KiB;
canonical actual inputs total at most 32 KiB. Review reuses the current valid
work's inputs and outputs and does not accept replacement bindings. Idempotent
replay returns the original attempt, not a new execution.

`task.execution.started` and `task.execution.finished` are controller-owned
append-only events. Generic `/api/events` cannot inject them. The record binds
company, room, task/revision, participant, human initiator, mode, definition and
digest, admission generation, bound input values/hashes, timestamps, optional
parent review/delegation identity, response digest and terminal state/outputs.
Replay validates identity, contract, provenance, value hashes and transitions;
terminal history cannot be mutated. Managed jobs additionally bind the definition
digest/generation at registration, submission and each tool acknowledgment.

`GET /api/task-executions?room=<id>&taskId=<id>` requires room access. It returns
up to ten `executions` and `nextCursor`; use `before=<nextCursor>` for older
records. History remains readable with the feature off. The private global map
is not serialized into unrelated room state. Owner reads of another person's
personal-room execution history require a durable access-audit record; audit
failure refuses disclosure. Other actors cannot read foreign rooms.

## Current verification limits

Tests mock every database, worker and provider interaction and use synthetic
fixtures. They cover validation, strict decoding, canonical values, precision,
versions, permissions, legacy preservation, actual dispatch/review, delegated
input immutability, broker restriction/budgets, deadline narrowing, revocation,
persistence failures, read scoping and pagination. They do not certify a live
provider, image, worker or database. Production development admission remains
blocked by its existing acceptance requirements. File/bundle delivery and
independently observed artifact checks remain separate integration work; this feature does not close a P0 release gate.

The optional [recorded task progress](task-progress.md) view separates saved task status from admitted attempts, validated values, current inputs and required human review. It never infers live worker activity.

## Saved definition revision history

The task card's **Definition revision history** shows complete saved contracts,
newest first, ten at a time. Use **Older definitions** to page back. Revisions
saved without any execution remain visible; status or assignee changes do not
invent another contract revision. The view shows each version's digest and all
seven parts. It does not restore old scope, run work or accept a result. Execution
history separately preserves the exact contract and actual bindings used by a run.

`GET /api/task-definition-history?room=<id>&taskId=<id>` returns `definitions`
and `nextCursor`. Each record contains company, room and task IDs, its complete
`definition`, `digest`, and `eventPosition` (replay order, not a timestamp).
Use `before=<nextCursor>` to retrieve older records; nonexistent cursors refuse.
Only acknowledged validated task events enter the projection. There are no new
write events, database migration or external provider calls.

This history is part of `taskDefinitions` (default off, owner-controlled, no
restart). Turning definitions off preserves read-only history, as it does for
execution records, while contract mutation and structured execution remain
blocked by the existing server rules. Both UI and API enforce the appropriate
read scope. Room members may read that room's retained versions; foreign-room
reads refuse. Owner reads of another person's personal-room history require a
durable access audit before disclosure. Audit failure returns no versions.
The private history map is omitted from global state serialization. Current
implementation replays retained events on each load; paging bounds the response,
not total replay memory or database query cost. Tests mock every database call.

Managed submission repeats the current task-role, pending-dispatch, definition
policy generation and bound-input checks after worker allocation. A cancelled or
expired execution context cannot submit work. Definition/dependency off-on
cycles, source revision changes and owner reconciliation cannot reuse an older
allocation. These checks are mandatory and apply even with diagnostic UI off;
cleanup acknowledgment remains allowed after cancellation/revocation. A refused
submission is not proof that any previously allocated worker was removed: the
ordinary verified shutdown/reconciliation contract still applies. Regression
tests use mocked managed lifecycle calls and database events, never live workers.
A submission consumes an allocated job once. Already submitted, shutdown-confirmed
or shutdown-unconfirmed jobs cannot be submitted again. Cleanup confirmations
remain idempotent; a late uncertainty report cannot undo confirmed shutdown.

Typed-task authority also binds the execution's company, task, room, attempt,
definition generation, delegation and participant to the current pending task.
Work requires the assigned worker in Working state; review requires the reviewer
in Reviewing state. Its task revision must be exactly the admitted revision plus
one. Registration, submission, broker checks and result handling share this
pending authority predicate, including current input lineage. A saved task status
alone grants no execution rights.

Successful `task.execution.finished` replay requires the matching completed task
transition and its exact next revision in the atomic publication batch. A result
cannot attach to another dispatch, changed role or a task still marked pending.
Revoked contexts withhold both returned values and the agent result message;
failed/stopped records remain history. Tests exercise both work and distinct
review roles, mismatched projected identities, incomplete transitions and mocked
turns whose task role/state changes before they return. These safeguards are
mandatory; feature flags cannot enable mismatched publication or bypass audit.

A pending task keeps its actual worker/reviewer connection reserved for
reconciliation after controller restart, including before a managed job exists.
A durable job additionally reserves its original connection until confirmed
shutdown. Connection edits/deletion, sign-out, verification and participant
connection reassignment cannot erase these bindings. Feature flags do not relax
this guard; reconcile pending task state and worker shutdown separately.

## Admitted agent configuration

New typed execution records carry `configurationVersion: 1`, participant revision
and configuration fingerprint, connection ID/revision and connection fingerprint.
The server captures these from the authorized current participant and registered
CLI connection; clients and models cannot supply replacement metadata. Execution
history shows these identifiers in **Admitted agent configuration**, without
copying endpoint configuration or instructions into that view.

Pending registration, submission, tools and result publication require the exact
recorded stable configuration. Profile edits (including name, instructions or
revision), reassignment and connection changes revoke that pending authority.
Successful result replay also checks configuration at its recorded completion.
A committed profile edit cancels its local typed turn promptly; another controller
still refuses its stale tools/result through durable state checks. Cancellation
is not proof of worker removal; the ordinary shutdown audit remains required.
Failed saves do not cancel a turn. This is mandatory under `taskDefinitions`,
not a separately bypassable switch.

Session thread, execution mode, status/telemetry, disclosure projection and derived
company/actor fields are excluded from the stable profile fingerprint. They may
change during a legitimate turn. Actual company/room/participant/connection scope
is separately enforced. Endpoint/provider/protocol and other connection fields
are included in the connection fingerprint, even if an invalid event omits a
revision increment. No fingerprint establishes isolated worker origin, image
approval, provider account identity or physical quota.

Version-zero records with no configuration fields remain readable/replayable as
retained earlier history. Pending attempts without configuration identity cannot
resume or publish a new result; reconcile and explicitly dispatch a new attempt.
Partial or unknown configuration formats refuse. Completed accepted values remain
history and may be reviewed/reused under the ordinary current task/value policies
after later agent settings changes; they do not acquire a new execution identity.
No database migration, live provider or startup retry is required. Tests use
mocked DB/turns and synthetic metadata, not production acceptance evidence.
Native session callback persistence also rechecks typed configuration authority;
a stale turn cannot attach its thread to a newer profile or connection. Healthy
session metadata changes remain excluded from the stable fingerprint.

## Context policy revocation during a typed attempt

Execution configuration also records the room's `contextPolicyEpoch`. Every
acknowledged context-export policy edit advances the durable replay generation,
including exclusion-only edits and deny/allow cycles. Pending typed registration,
submission, tool use, session persistence and result publication refuse an older
generation or a currently denied destination. A locally committed policy edit
cancels local typed turns in that room; failed writes do not. The policy-editor
feature flag controls editing, never enforcement of an existing saved policy.

Completed value history is retained. Re-enabling export cannot revive a cancelled
allocation: the owner must explicitly dispatch new work under the current policy.
Policy changes cannot recall context already sent to a provider. Physical worker
shutdown still requires independent verified cleanup. These checks do not create
provider deletion evidence or certify production isolation. Personal-room policies
are projected after deriving their room identity so an acknowledged policy cannot
silently disappear during event replay.

Automatic human evidence is a separate, bounded reference category, not a typed
input or returned deliverable. Reference exclusions suppress it for both work and
review. Explicit owner bindings and the seven-part task contract remain the
current request. An evidence note saying a check passed cannot become an observed
check or substitute for validated returned values. See
[context disclosure](context-disclosure.md).
