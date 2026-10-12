# Workflow execution

The optional sequential runner turns a saved room workflow into actual isolated
agent calls, review feedback and recorded human decisions. It is an early delivery
step, not a complete automatic software release pipeline.

## Enable and use

1. As the owner, open **Settings → Features → Workflows → Execution**. The
   `workflowExecution` flag defaults **off**. Enabling/disabling takes effect on
   the server immediately; the settings page reloads without a server restart.
2. Test the assigned connections and provision [development workspaces](development-workspaces.md)
   for any step that needs to edit files. Execution does not grant tools or access.
3. In a room, create or apply a template, assign implementation agents and an
   independent reviewer, then save. Existing drafts with 300-second timeouts must
   be edited to 180 or less. New nodes/templates default to 180 seconds.
4. Start with **Start → Agent task → Review → Evidence → Human approval → Finish**.
   Connect Review's **changes** port back to Agent task and **passed** to Evidence.
   All finish paths must have human approval after the last agent/review step.
5. Click **Run**, describe the approved task and submit. The run pins the saved
   graph revision; editing the graph later does not modify that run.
6. Watch **Execution** for running/failed/awaiting-human/completed states and open
   **View outputs** for attempts, text hashes, reviewer verdicts and evidence.
7. At Evidence, supply artifact references and the actual checks you inspected.
   At Human approval, accept with a rationale or **Request changes** to return to
   the latest implementation step. The independent review executes again.

Example input: “Add a bounded search filter, including empty-input and failure-path
tests, user/technical documentation and an owner-controlled server feature flag.
Do not commit, push or deploy.” An agent's “tests passed” statement is not enough
for acceptance; inspect its patch and actual results separately.

## Supported nodes and limits

| Node | Execution |
| --- | --- |
| Start | Records owner-started execution and advances |
| Agent task | Calls the assigned current room agent through the isolated worker |
| Review | Calls a separate agent; explicit JSON `outcome` selects passed/changes |
| Evidence | Pauses for owner-supplied evidence references/rationale |
| Human approval | Pauses for owner acceptance; human rejection returns work |
| Decision | Pauses for an owner yes/no decision and rationale |
| Wait | Cancellable delay; 0–180 seconds |
| Finish | Completes only after a validated human-approval path |
| Parallel | Rejected; concurrency/join semantics are not implemented |

Reviews must return only `{"outcome":"passed","summary":"Reasoning"}` or
`{"outcome":"changes","summary":"Required changes"}`. Malformed, empty or
ambiguous responses fail the run; the runner never guesses from prose.

Automatic retries are rejected: set Retries to zero. Agent timeout zero means
180 seconds; explicit timeouts must be at most 180 seconds. Maximum iterations is
the visit limit **per node**, including explicit recovery attempts. A run has a
500-completed-step bound, 32 KiB maximum step output, 16,000-byte task/evidence
inputs and 128 KiB maximum accumulated prompt. Exceeding a limit fails visibly
rather than silently dropping prior context. One unfinished run per room is
allowed. An agent already busy with chat cannot be simultaneously dispatched.

Templates remain editable plans and may need different agent assignments or
shorter timeouts. Cross-room transfers, scheduled runs, automatic branch/patch
integration, parallel QA/security joins and deployments are not supported.

## Permissions and safety

Only the authenticated owner may start, inspect or control execution. Every POST
requires the normal CSRF/Host/Origin protections. The owner can accept work without
being listed as a room participant. Other accounts' state excludes run outputs.

Every dispatch and gate checks audit integrity and emergency-stop state. Current
membership and connection capability are rechecked before dispatch. Agent removal,
stop, cancellation or a disabled flag prevents late results from being recorded as
successful outputs. The feature cannot weaken mandatory rootless Docker isolation,
administrator-owned repository allowlists or provider credential boundaries.

Disabling execution cancels active drivers, blocks start/resume/gate operations
and preserves recorded runs. Cancel and inspection remain available. Paused gates
stay paused until execution is enabled again. No provider call is initiated just
by opening the editor or saving a graph.

## Recovery and evidence limitations

Work proceeds without the browser. A controller restart marks orphaned running
records interrupted when the owner lists runs. **Resume** explicitly repeats the
current step; inspect its private workspace first because it may already have
modified files. Completed steps are retained. Failed/exhausted workflows can be
cancelled and replaced with a revised graph. Automatic retries are deliberately
absent because model/tool side effects cannot be made exactly-once by event logs.

Use only one controller per database. There are no distributed leases or automatic
orphan-container reconciliation. Audit write failures stop progress; investigate
before recovery. Existing emergency worker termination is best-effort with bounded
confirmation, not a claim that provider-side sessions are revoked.

**Evidence is not yet automatically verified.** Step output hashes identify the
captured response text, not the Git revision or patch. Review agents receive prior
outputs and human-supplied references, not another agent's private filesystem.
They are instructed to request changes when necessary code/evidence is unavailable.
Manual patch collection and independent checks remain necessary. Do not represent
a completed run as a security certification or automatically merge its work.

## Technical contract

- `cmd/server/workflow_runner.go`: runnable-graph validation, commands, safety
  checks, per-run goroutine, gate routing and recovery. `workflowLive` tracks live
  cancellations; `active`/`cancels` coordinate with chat and emergency stop.
- `cmd/server/agents.go`: `executeAgentTurn` shares the mandatory isolated adapter
  path with chat. A source fork **CAN** replace this with a host launcher, **BUT**
  that exposes host files/credentials and invalidates the isolation contract.
- `internal/room/model.go`: `WorkflowRun` pins graph/input/status/current node,
  revision, creator, time, visit counts and completed attempts. `WorkflowStep`
  stores node/agent IDs, response text/SHA256, human evidence, selected port and time.
- `workflow.run.updated` is append-only internal state; generic `/api/events`
  callers cannot forge run records. No database migration or new table is needed.
- `GET/POST /api/workflow-runs?room=ROOM`: owner commands with revision checks.
  See [API examples](api.md#workflow-execution).
- `web/workflows.js`: owner Run button, status/output panels, gates/recovery and
  non-overlapping polling. Unsaved graphs/open overlays are not refreshed.

Tests mock every database and provider interaction. They cover complete review
rework/acceptance, both flag states, cancellation, emergency stop, membership
removal, failures, invalid verdicts, limits, graph safety, recovery and auth/revision
boundaries. They do not constitute a live provider workflow certification.

For a manual browser check with real HTTP handlers and synthetic data, run
`COMMONROOM_BROWSER_WORKFLOW_TEST=1 go test ./cmd/server -run '^TestWorkflowBrowserHarness$' -count=1 -timeout 4m -v`.
The localhost-only fixture lasts three minutes on port 8794, uses the mockServer
bootstrap key documented in test source, and keeps every event in memory. It
cannot dispatch real providers. Normal fast suites skip this manual wait.

The same synthetic harness can show blocked-status tooltips with
`COMMONROOM_BROWSER_AGENT_STATUS_TEST=1` instead; that mode lasts 60 seconds and
seeds only in-memory activity/error records. No real providers or database are used.
