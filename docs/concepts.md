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
