# Product concepts and usage

[Documentation index](README.md)

## Company, humans, and rooms

An installation currently has one company. Rooms organize discussion, participants, tasks, evidence, and workflows. Create a shared room using the + beside Rooms. A room is not yet a complete filesystem or network security boundary.

The initial human is the owner. Other humans can only join through invitations issued under Settings → Company members. The owner assigns Developer or QA and grants shared-room access. Only the owner changes those roles and permissions. Members can edit their own display names. Display names are not identity: two people can have the same name and still be different accounts.

Each human and registered bot identity has a personal room. Humans can access their own room and assigned shared rooms; the owner can access all rooms. Owner reads of another person's personal room are recorded. Bot identities currently follow connection registrations and do not have independent API credentials.

Humans appear in the squad. The owner's room presence defaults to enabled and can be removed from shared rooms. Presence does not remove the owner's account permissions. Human seniority defaults to 11 and can be lowered through room settings. Higher numbers express greater decision authority; automated consensus and conflict arbitration are still planned.

## Connections versus participants

A connection describes an agent provider, transport, location, and endpoint. A participant is that connection's membership in a room, with its own name, title, instructions, seniority, challenge level, and native conversation session.

Register connections before inviting participants. One connection can support distinct room profiles; changing a room profile does not create a new provider account. Local means the computer running the server, including when that computer is a remote Debian machine.

See [SSH agent setup](agent-ssh.md) for concrete remote-host configuration and tests.

Working adapters are CLI Codex and Claude, locally or over SSH. HTTP, WebSocket, and MCP registrations are placeholders. Website addresses for ChatGPT or Claude are not adapter endpoints. Agent provider authentication is separate from human sign-in.

### Guided connection editor

Choose **Connections → Plug in agent** to open **New connection**. Give it a name, select the provider icon and connection method, then choose where it runs:

- **This computer:** the computer running the CommonRoom server. CLI examples are `codex` or `claude`; no port is needed.
- **Internal network:** a reachable machine on your LAN. CLI endpoints refer to an SSH alias configured for the server account.
- **Remote / internet:** a service outside the internal network. Remote CLI agents still use SSH aliases; URL registrations need the complete documented service URL.

Local OpenAI/Codex and Claude CLI connections automatically use `codex` or `claude`; no address field is shown. Remote CLI choices show an SSH endpoint, while HTTP, WebSocket and MCP show a service URL. Other local CLI providers and nonstandard saved commands retain an editable command field. Address drafts are remembered separately for each method/location (and CLI provider), so switching to a web method never carries an SSH alias into its URL. Returning to a previous choice restores its draft. **Use example** explicitly replaces the visible address with a placeholder to edit. Open **Setup instructions** or the **SSH setup guide** for detailed help. Do not include credentials in addresses. Saving registers a connection; invite its agent to a room separately.

Only Codex and Claude CLI adapters currently respond in rooms. Other provider selections choose icons, not new adapters. HTTP, WebSocket and MCP are marked **registration only**; selecting them does not enable agent execution. Examples are not live services.

Editing uses the same form and preserves saved values. Changing a method/address resets native sessions according to existing server rules; active work blocks those changes. History stays saved. The owner controls connections; other members retain their existing permission limits. The redesigned form changes presentation, not authorization, persistence, or execution capabilities.

The optional server flag `COMMONROOM_CONNECTION_EDITOR` defaults to `true`. Set it to `false` and restart the application to restore the previous editor. Only operators with access to the server environment can change it; there is no browser setting for this flag. The server then serves no replacement editor script, without disabling connections, deleting data, or interrupting existing work. The existing connection API is unchanged. Enabled/disabled serving, guidance, safe rendering, preservation of edited values, and connection validation are covered by automated tests; live provider interoperability still requires separate verification.

## Profiles and challenge levels

Engineering discipline, level, speciality, and assignment describe the participant's job. They influence prompt guidance; they do not grant operating-system permissions.

| Level | Guidance |
| --- | --- |
| 0 — Supportive | Help progress and identify clear errors or risks. |
| 1 — Constructive | Question unclear assumptions and propose alternatives. |
| 2 — Critical | Examine tradeoffs and request evidence. |
| 3 — Adversarial | Stress-test proposals and look for failure cases. |
| 4 — Red team | Challenge claims and seek weaknesses within authorized scope. |

Challenge the work, not the people. Authority and access must not be inferred from a persuasive prompt or impressive title.

## Messages and context

The composer defaults to Room only. Select an agent or mention it to request a turn. Select Everyone or mention `@room` to request replies from supported available participants. Busy or unsupported agents may be skipped; a broadcast is not proof every member replied.

An agent turn includes profile guidance, optional reference context, available shared memory, and the latest 12 previously stored room messages, plus the new request. Its native session also retains provider-side conversation continuity. This does not promise every historical room message is included in every prompt.

Existing supported local Codex conversations can be copied into independent sessions or resumed. Resuming changes the original conversation; copying keeps future turns separate. Importing optional reference text is not a live attachment to a browser chat or interactive terminal.

Remove an agent through its squad menu to stop its participation. Application history is retained. Cancellation of remote processes is not yet guaranteed.

## Shared memory

Memory entries have a title, body, scope, and revision. Editing a stale revision returns a conflict requiring a reload; previous revisions stay in event history.

- Room: available within the room; use this as the normal starting scope.
- Project: currently shared across the installation's single project.
- Application: shared across the installation.

Separate projects and complete project permissions do not exist yet. Non-owner memory writes are currently restricted to Room scope. Automatic squad discussion, human decision timeouts, and senior-agent resolution of conflicts remain planned.

## Tasks, evidence, and workflows

Tasks have an assignee, optional reviewer, criteria, status, and revision. Evidence entries record human-provided Changes, Checks, Artifacts, or Review. They are assertions supplied by a person, not automatically verified tool results. Automatic QA transitions, evidence collection, and file uploads are pending.

Workflows belong to rooms. Add nodes, connect outputs to inputs, edit properties, save versions, and Validate. Validation reports structural/assignment problems; Save accepts incomplete drafts. Run is disabled. Retry/timeouts, approvals, evidence requirements, and room handoffs stored in the graph do not execute yet.

Human overrides, acceptance authority, and immutable history remain product requirements with incomplete enforcement. Consult the security review before treating the prototype as a production policy system.

## Workflow templates

Open **Workflow templates** in the main menu (or **Browse templates** inside the editor). The catalog offers:

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

Choose **Apply to room**, select a room you can access, and confirm **Apply template**. This creates a new saved workflow and opens that room’s editor. Reapplying creates another copy; it does not overwrite an existing workflow. An unsaved editor draft requires confirmation before it is left. Each copy has fresh node/edge IDs, detailed instructions, an evidence step, an independent review with a changes loop, and human acceptance. The default revision limit is three iterations; node timeouts default to 300 seconds with no retries.

Assign actual room agents to task and review nodes, adapt their instructions and evidence requirements, then Validate and Save. Templates describe suggested roles; they do not invite agents, assign authority, execute tests, or automatically gather evidence. Validation initially reports unassigned agents. **Run is still disabled**: the runner, enforced human approvals, evidence gates, retries, and automatic review loops have not been implemented. Use the graphs as collaboration plans until execution is available.

Owners and Developers may apply templates to rooms they can access. QA members may browse the catalog but cannot apply templates. Room authorization and workflow validation remain server-enforced. Templates cannot override access, immutable history, or other security rules. Shared catalog availability does not expose another room’s workflows.

### Configuration and technical contract

`COMMONROOM_WORKFLOW_TEMPLATES` defaults to `true`. An operator can set it to `false` and restart to hide the catalog and restore the previous single-template editor. Existing saved workflows, history, room membership, and ordinary workflow editing are retained. The flag is managed through the server environment, not a browser control; invalid boolean values prevent startup. The server serves no catalog script when disabled and rejects explicit template-application requests.

The catalog and graph factory are in `web/workflow-templates.js`. The server gates the script and validates known template IDs. Applying calls `POST /api/events?room=<id>` with `kind: "workflow.revised"`, top-level `templateId`, and a new workflow payload. The API returns `workflowId` so the client opens the exact created workflow, rather than choosing whichever workflow was saved most recently by another actor. Room scoping, permissions, revision rules, and append-only event storage remain unchanged. Subsequent edits use the ordinary workflow API without template metadata.

Automated tests cover graph structure, valid output ports, review loops, room-agent assignments, fresh IDs, independent copies, delivery guidance, new-workflow persistence, room access, role restrictions, unknown IDs, and both flag states. All persistence tests use an in-memory mock. Live workflow execution is neither tested nor enabled.

### Documenter and disaster planning

**Documenter** starts by defining the audience and scope. Agents audit the existing guides against source, write missing user/technical documentation and appropriate source comments, check links/examples, gather evidence, and request independent accuracy/usability review before human acceptance. Its output is documented behavior with links and a verification record, not a claim that untested setup instructions work.

**Disaster planning** identifies critical systems, acceptable downtime/data loss, dependencies, failure scenarios and responsible people. Agents define recovery time objectives (how quickly service should return) and recovery point objectives (how much data loss is acceptable), then prepare backup/restore procedures, escalation responsibilities, rollback criteria and a recovery runbook. The rehearsal step plans a tabletop exercise or separately authorized isolated verification. It does not authorize production restores, deletion, failover, or changes to immutable history. Evidence must distinguish completed checks from untested assumptions. An independent reviewer challenges readiness before the human accepts residual risks and follow-up priorities.

Both use the same **Apply to room** flow, three-iteration default, fresh-copy behavior, unassigned agent slots, room permissions and `COMMONROOM_WORKFLOW_TEMPLATES` flag described above. Neither template executes automatically or schedules rehearsals; Run remains disabled. Graph/copy tests and mock-only API tests cover both new template IDs, including disabled-flag rejection.

### Cooperative adversarial review

This template uses cooperation to define the common goal and adversarial review to stress-test the proposed work. Its sequence is **Human goal → Builder proposal → Challenger critique → Reconciliation → Evidence → Independent review → Human decision → Finish**. The review’s changes output returns to the builder; the default maximum revision count is three, as with other templates.

Assign different agents to the Builder and Challenger steps, ideally with another independent reviewer. Constructive challenge level 1 is suggested for the builder; critical level 2 or adversarial level 3 is suggested for the challenger. These are prompt/profile suggestions, not enforced permissions or automatic profile changes. The reconciliation step may go to the builder or a neutral facilitator. A human decides unresolved tradeoffs; agents should record disagreement instead of inventing consensus.

Criticisms must identify evidence, risks or concrete alternatives. The evidence step records the original proposal, critique, revised outcome, accepted/rejected criticisms and unresolved issues. Critique targets the work, never the participants. Titles, challenge levels and persuasive arguments confer no extra authority or tool access.

Apply it through **Workflow templates → Apply to room**. It uses the existing catalog flag and room permissions, creates a fresh saved copy, and leaves agents unassigned. Tests cover its connected graph, review loop, fresh copies, builder/challenger/reconciliation sequence, guidance, application and disabled-flag rejection through an in-memory database mock. Execution, multi-round debate, enforcing distinct agents, consensus and escalation are not automated; Run remains disabled.

## Emergency agent cutoff

The owner can choose **Stop all agents** in the header or **Stop room agents** in a room. Confirm the scope and optionally record a reason. New agent dispatch and session creation/invitation in the stopped scope are blocked; managed turns receive cancellation requests. Human access, saved messages, memory, evidence and workflows are preserved. Per-agent **Remove** remains available and also requests cancellation.

Use **Resume all agents** or **Resume room agents** explicitly to allow new requests. Resuming application dispatch does not clear separately stopped rooms, and resuming a room cannot bypass an application stop or integrity incident. Cancelled work is not restarted. A cancelled turn’s result is discarded even if dispatch has since resumed. Completed effects and already recorded messages cannot be undone.

A successfully recorded stop is replayed from event history on restart. Application stop cancels registered managed turns before recording its audit event and does not require a successful database state read. An audit failure returns an error but retains an in-memory stop latch. **That unrecorded latch does not survive restart**; investigate and successfully record the stop before restarting. Room stop requires state loading to identify its participants; use the application stop when database reads are unavailable. Stop requests serialize with other agent operations; the server must be running and reachable, and an operation holding the server lock can delay a stop.

Cancellation kills the directly managed local CLI/SSH command on a best-effort basis. It is not proof that remote processes, child processes, provider-side operations or tools already launched elsewhere have terminated. CLI agents have no independently authenticated CommonRoom credentials to revoke. This is a dispatch cutoff, not operating-system containment or continuous detection of policy violations. For a suspected rogue agent, verify shutdown on its host and isolate/revoke host or provider access separately where needed. Automatic policy monitoring and complete integrity-triggered process shutdown remain future work.

`COMMONROOM_AGENT_SAFETY_UI` defaults to `true` and sets the initial visibility. The owner can override it in **Settings → Feature flags → Security → Emergency agent controls**. The saved setting survives restarts; changing it reloads the page. **Backend stop enforcement and the protected API cannot be disabled by this UI flag**; an active stop stays active, and the emergency banner remains visible. Owner permissions and CSRF/Host/Origin protection remain mandatory. See the operations/API guides for access when controls are hidden.

Tests use mocked database interactions to verify scoped cancellation, persistent replay, explicit resume, role/CSRF restrictions, database-read outages, audit failures, flag states, and rejection of late cancelled results. No real provider processes or database fixtures are launched by these tests.

## Human profile photos and feature flags

In Settings, choose **Edit profile → Change profile photo**. The same option is available from your own squad profile. Upload a JPEG or PNG up to 5 MB, inspect the preview, and save. The browser reduces it to a maximum of 160 pixels per side; the server validates bounded raster data and re-encodes it as JPEG without original metadata. SVG, GIF and remote image URLs are rejected. Initials remain the fallback when no photo is set.

The photo belongs to the human account, rather than a room or role. It appears in each room's squad and beside messages identified by that human's account ID. Earlier messages display the current profile photo; their stored text, authors and immutable records are not rewritten. Legacy messages without an account identity retain initials unless explicitly identified as the original “You” owner. Members can change only their own image; they cannot change their roles or another account's photo. Owner access edits preserve the person's photo.

Photo revisions are append-only audit events. Removing a photo clears the current display, but earlier image records remain subject to the configured retention window and human-authorized purge. Do not upload images you need immediately erased from all historical records.

Owners see grouped flags under **Settings → Feature flags**:

| Group | Flag | Default | Management |
|---|---|---|---|
| Cosmetic | Human profile photos (`humanProfileImages`) | On | Settings; disabled blocks uploads and hides photos, but permits self-removal |
| Security | Emergency agent controls (`agentSafetyControls`) | On | Settings override; environment supplies initial default |
| Connections | Guided connection editor | On | `COMMONROOM_CONNECTION_EDITOR`; restart required |
| Workflows | Workflow templates | On | `COMMONROOM_WORKFLOW_TEMPLATES`; restart required |

Saved flags are immutable events and survive restarts. Re-enabling photos restores any retained current photos. Cosmetic settings grant no permissions. Emergency controls visibility never disables stop enforcement, clears an active stop, or bypasses an integrity incident.
