# Workflow templates

[Documentation home](README.md) · [Product guide](concepts.md)

Start with a reviewable process, assign real participants and adapt it to the task.

## Workflow templates

Open **Workflow templates** in the main menu (or **Browse templates** inside the editor). Repository-development templates remain plans while public development admission is blocked. Applying a template cannot unlock tools or override that barrier. The catalog offers:

| Template | Intended outcome |
| --- | --- |
| Idea assessment | A human idea reviewed for value, functionality, feasibility, and estimated effort |
| Feature development | Implementation with tests, documentation, a feature flag, evidence, and independent QA |
| UI refinement | Interaction design, implementation, browser checks, visual evidence, and human review |
| Bug investigation | Reproduction, root cause, fix, regression test, and verified resolution |
| Security review | Authorized threat review, prioritized findings, remediation evidence, and residual-risk decisions |
| Documenter | Documentation audit, complete guides/source comments, verified examples, and independent accuracy review |
| Disaster planning | Failure scenarios, recovery objectives, a recovery runbook, rehearsal planning, and readiness review |
| Cooperative adversarial review | Builder proposal, evidence-based challenge, reconciliation, independent review, and human resolution |

Choose **Apply to room**, select a room you can access, and confirm **Apply template**. This creates a new saved workflow and opens that room’s editor. Reapplying creates another copy; it does not overwrite an existing workflow. An unsaved editor draft requires confirmation before it is left. Each copy has fresh node/edge IDs, detailed instructions, an evidence step, an independent review with a changes loop, and human acceptance. The default revision limit is three iterations; node timeouts default to 180 seconds with no retries.

Assign actual room agents to task and review nodes, adapt their instructions and evidence requirements, then Validate and Save. Templates describe suggested roles; they do not invite agents, assign authority, execute tests, or automatically gather evidence. Validation initially reports unassigned agents. **Execution is off by default**. The sequential runner supports isolated agent steps, review loops and human evidence/approval gates. Assign independent implementation and review agents and use timeouts no greater than 180 seconds. Parallel nodes and automatic retries remain unsupported. See [workflow execution](workflow-execution.md).

Owners and Developers may apply templates to rooms they can access. QA members may browse the catalog but cannot apply templates. Room authorization and workflow validation remain server-enforced. Templates cannot override access, immutable history, or other security rules. Shared catalog availability does not expose another room’s workflows.

## Configuration and technical contract

`COMMONROOM_WORKFLOW_TEMPLATES` defaults to `true`. An operator can set it to `false` and restart to hide the catalog and restore the previous single-template editor. Existing saved workflows, history, room membership, and ordinary workflow editing are retained. The flag is managed through the server environment, not a browser control; invalid boolean values prevent startup. The server serves no catalog script when disabled and rejects explicit template-application requests.

The catalog and graph factory are in `web/workflow-templates.js`. The server gates the script and validates known template IDs. Applying calls `POST /api/events?room=<id>` with `kind: "workflow.revised"`, top-level `templateId`, and a new workflow payload. The API returns `workflowId` so the client opens the exact created workflow, rather than choosing whichever workflow was saved most recently by another actor. Room scoping, permissions, revision rules, and append-only event storage remain unchanged. Subsequent edits use the ordinary workflow API without template metadata.

## Documenter and disaster planning

**Documenter** starts by defining the audience and scope. Agents audit the existing guides against source, write missing user/technical documentation and appropriate source comments, check links/examples, gather evidence, and request independent accuracy/usability review before human acceptance. Its output is documented behavior with links and a verification record, not a claim that untested setup instructions work.

**Disaster planning** identifies critical systems, acceptable downtime/data loss, dependencies, failure scenarios and responsible people. Agents define recovery time objectives (how quickly service should return) and recovery point objectives (how much data loss is acceptable), then prepare backup/restore procedures, escalation responsibilities, rollback criteria and a recovery runbook. The rehearsal step plans a tabletop exercise or separately authorized isolated verification. It does not authorize production restores, deletion, failover, or changes to immutable history. Evidence must distinguish completed checks from untested assumptions. An independent reviewer challenges readiness before the human accepts residual risks and follow-up priorities.

Both use the same **Apply to room** flow, three-iteration default, fresh-copy behavior, unassigned agent slots, room permissions and `COMMONROOM_WORKFLOW_TEMPLATES` flag described above. Neither template schedules rehearsals. Optional sequential execution dispatches assigned steps; rehearsal instructions do not authorize production operations. Graph/copy tests and mock-only API tests cover both new template IDs, including disabled-flag rejection.

## Cooperative adversarial review

This template uses cooperation to define the common goal and adversarial review to stress-test the proposed work. Its sequence is **Human goal → Builder proposal → Challenger critique → Reconciliation → Evidence → Independent review → Human decision → Finish**. The review’s changes output returns to the builder; the default maximum revision count is three, as with other templates.

Assign different agents to the Builder and Challenger steps, ideally with another independent reviewer. Constructive challenge level 1 is suggested for the builder; critical level 2 or adversarial level 3 is suggested for the challenger. These are prompt/profile suggestions, not enforced permissions or automatic profile changes. The reconciliation step may go to the builder or a neutral facilitator. A human decides unresolved tradeoffs; agents should record disagreement instead of inventing consensus.

Criticisms must identify evidence, risks or concrete alternatives. The evidence step records the original proposal, critique, revised outcome, accepted/rejected criticisms and unresolved issues. Critique targets the work, never the participants. Titles, challenge levels and persuasive arguments confer no extra authority or tool access.

Apply it through **Workflow templates → Apply to room**. It uses the existing catalog flag and room permissions, creates a fresh saved copy, and leaves agents unassigned. Tests cover its connected graph, review loop, fresh copies, builder/challenger/reconciliation sequence, guidance, application and disabled-flag rejection through an in-memory database mock. The optional sequential runner can dispatch these steps and follow review feedback. Review agents must be distinct from implementation agents. Consensus and escalation are not automated.
