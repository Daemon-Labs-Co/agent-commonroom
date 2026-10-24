# Product guide

[Documentation home](README.md)

Use rooms to keep a goal, its participants, context and evidence together. Start
with a small authorized task and make its review criteria visible before asking
for agent work. [Release readiness](release-readiness.md) distinguishes the current
prototype from the intended agent-development harness.

## Company, humans, and rooms

The [organization chart](organization.md) records reporting relationships. Titles,
reporting lines and retained seniority fields do not grant account permissions.

An installation currently has one company. Rooms organize discussion, participants, tasks, evidence, and workflows. Create a shared room using the + beside Rooms. A room is not yet a complete filesystem or network security boundary.

The initial human is the owner. Other humans can only join through invitations issued under Settings → Company members. The owner assigns Developer or QA and grants shared-room access. Only the owner changes those roles and permissions. Members can edit their own display names. Display names are not identity: two people can have the same name and still be different accounts.

Each human has a personal room. Registering an agent no longer adds an unused personal room to the room catalog by default. Existing agent rooms referenced by saved history or work remain available. Humans can access their own room and assigned shared rooms; the owner can access all rooms. Owner reads of another person's personal room are recorded. Bot identities currently follow connection registrations and do not have independent API credentials.

Humans appear in the squad. The owner's room presence defaults to enabled and can be removed from shared rooms. Presence does not remove the owner's account permissions. Legacy seniority fields remain for compatibility; they are not authority. Structured disagreement and governing-agent delegation remain unfinished.

## Connections versus participants

A connection describes an agent provider, transport, location, and endpoint. A participant is that connection's membership in a room, with its own name, title, instructions, seniority, challenge level, and native conversation session.

Register connections before inviting participants. One connection can support distinct room profiles; changing a room profile does not create a new provider account. A location label describes configuration; it does not enable native host execution.

See [SSH agent setup](agent-ssh.md) for concrete remote-host configuration and tests.

Current chat adapters are Codex and Claude through the approved isolated worker. HTTP, WebSocket, and MCP registrations are placeholders. Website addresses for ChatGPT or Claude are not adapter endpoints. Agent provider authentication is separate from human sign-in.

### Guided connection editor

Choose **Connections → Plug in agent** to open **New connection**. Give it a name, select the provider icon and connection method, then choose where it runs:

- **This computer:** a UI location label; saving a local command does not bypass the required isolated worker.
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

Saved-conversation metadata and optional reference text can provide context where supported. Host conversation IDs do not automatically become isolated worker sessions. Importing text does not attach to a browser chat or terminal.

Remove an agent through its squad menu to stop its participation. Application history is retained. Confirm worker shutdown through job diagnostics; a cancellation request alone is not proof.

## Shared memory

Memory entries have a title, body, scope, and revision. Editing a stale revision returns a conflict requiring a reload; previous revisions stay in event history.

- Room: available within the room; use this as the normal starting scope.
- Project: currently shared across the installation's single project.
- Application: shared across the installation.

Separate projects and complete project permissions do not exist yet. Non-owner memory writes are currently restricted to Room scope. Automated memory discussion is unfinished. Free disagreements require human resolution; voting and escalation timeouts belong to later paid capabilities.

## Tasks, evidence, and workflows

Tasks have an assignee, optional reviewer, criteria, status, and revision. Evidence entries record human-provided Changes, Checks, Artifacts, or Review. They are assertions supplied by a person, not automatically verified tool results. Automatic QA transitions, evidence collection, and file uploads are pending.

Workflows belong to rooms. Add nodes, connect outputs to inputs, edit properties, save versions, and Validate. Validation reports structural/assignment problems; Save accepts incomplete drafts. Run is available when the owner enables `workflowExecution`. Supported sequential nodes execute with bounded timeouts, explicit review outcomes and human gates. Parallel nodes, retries and cross-room handoffs are rejected or unavailable. See [workflow execution](workflow-execution.md).

Human overrides, acceptance authority, and immutable history remain product requirements with incomplete enforcement. Consult the security review before treating the prototype as a production policy system.


## More workspace guides

| You want to… | Read |
| --- | --- |
| Apply a process template | [Workflow templates](workflow-templates.md) |
| Interpret status, stop or recover work | [Agent activity](agent-activity.md) |
| Find recorded decisions | [History](history.md) |
| Edit reporting relationships | [Organization](organization.md) |
| Adjust profiles, layout and display preferences | [Interface preferences](interface.md) |



### Automatic agent-room compatibility

Agents participate in shared rooms assigned by the owner; they do not need a
personal room to run. `automaticAgentRooms` defaults **off**, is enforced in the
server room catalog, and can be changed only by the owner through Settings → Features →
Agent rooms. Changes apply on the next state load without a restart. Switching
on restores the legacy automatic room catalog; switching off omits only unused
implicit agent rooms. It never deletes events, moves tasks, cancels work or
changes authentication, room permissions, retention or isolation.

The controller still derives legacy agent rooms internally while replaying old
events, so historical scope checks remain valid. After replay, it omits an
implicit agent room only if no saved event references its identifier. References
are retained conservatively across nested audit payloads, settings and removed
work; an incidental matching string can therefore keep a room visible. Explicit
shared rooms and human personal rooms are unaffected. Existing used agent rooms
are retained as legacy spaces; consolidating them into outcome-based shared rooms
requires a separate migration. This change does not introduce direct messages or
require a deliverable field on rooms yet.
