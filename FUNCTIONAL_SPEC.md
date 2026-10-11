# Agent CommonRoom — Functional Specification

Version 2 · 9 October 2026 · UI workflow baseline

## Purpose and scope

A browser-based application where a human connects local or remote agents, invites them to rooms, assigns room-specific profiles, and coordinates work using shared memory. Initial use is single-user and local; ownership and access boundaries must accommodate future multi-user use. Remote agents participate through the same workflow as local agents.

The planned implementation uses Go, a TypeScript web interface, and PostgreSQL. These are technical direction, not a claim that the accompanying UI demo implements them.

## Connections and invitations

FR-01: Users can create, inspect, edit, test, disable, and remove agent connections. Connections contain adapter, protocol, endpoint/process configuration, credentials references, and capabilities; they do not contain room profiles.

FR-02: Support extensible adapters for Claude Code, Codex CLI/Jerry, Gemini, and this assistant where an accessible integration exists. Separate agent behavior from transport. Support local processes, HTTP, WebSocket, and MCP. An HTTP or MCP endpoint requires an explicit mapping to agent operations; protocol support alone does not ensure compatibility.

FR-03: When inviting an agent, users choose a new session or an existing conversation. Show whether the adapter supports live attachment, resumption, or transcript import. Do not claim these preserve the same state. Unsupported methods cannot be selected.

FR-04: Each room membership has an independent session. One connection can serve multiple memberships across rooms, or several memberships in one room, subject to adapter capacity. Do not concurrently attach the same live session to multiple memberships.

## Rooms, profiles, and directories

FR-05: Users can create, view, edit, archive, and reopen projects and rooms. Rooms have objectives, participants, tasks, memory, directory boundaries, and participation settings.

FR-06: Participant profiles support create/read/update/remove operations for name, role, instructions, permissions, seniority, acceptance authority, and kicker settings. Optional templates supply defaults, copied into membership. Profile revisions preserve historical attribution. Edits apply next turn by default; immediate application depends on interruption capability.

FR-07: Rooms define multiple allowed directories and read/write permissions. Profiles may narrow access. Effective access intersects room restrictions with underlying agent capabilities and permissions. Room membership never grants unavailable access. Working directory is separate from access scope.

FR-08: Directory expansion requires a human action. Narrowing access pauses affected work until restrictions are enforced. Existing context may contain previously acquired information; directory restrictions cannot erase it. Adapters unable to enforce boundaries must disclose the limitation and cannot be represented as restricted execution.

## Conversation and task workflow

FR-09: Humans and agents can send messages, reply, mention participants, create assignments, ask questions, and hand off work. Responses stream where supported. Keep tool activity distinguishable from conversational content. Directed participation is default; roundtable and bounded autonomous modes are optional.

FR-10: Independent work runs concurrently. Dependencies gate dependent work. Users can request individual or squad status while work runs. Label status derived from runner activity separately from a direct agent response.

FR-11: Tasks progress through Assigned, Working, Submitted, Reviewing, Changes requested, and Accepted. QA participants can return findings to the developer. Only an explicitly authorized participant or a human can accept work. Completion is not acceptance.

FR-12: Humans override application decisions and rules through recorded actions, except immutable-history protection. Application authority cannot create access or operations unavailable in the underlying system.

## Shared memory

FR-13: Agents are encouraged to write useful context freely within authorized scopes. Memory supports create, read, search, update, trash, and restore. Each entry has one scope: Room (default), Project, or Application. Broader writes must be explicit. Retrieval covers authorized applicable scopes and labels source and scope.

FR-14: Memory revisions preserve author, source references, timestamps, and prior values. Detect concurrent edits; never silently overwrite. Room-specific guidance takes precedence over broader defaults, but conflicting claims remain visible. File access does not automatically authorize publishing content to broader memory scopes.

FR-15: Conflicts open a discussion with the squad, including humans. Request human direction and wait for a configurable timeout; online presence alone is not a response. Human decisions prevail. Without a timely human response, the highest-ranked available agent decides (larger numeric decision-seniority values mean greater authority). Equally senior agents must reach consensus; no automatic tie-breaker. If no eligible agent exists or consensus fails, retain the unresolved conflict and escalate. A later human decision supersedes by new revision.

## Kicker and recovery

FR-16: Show connection and work states: Disconnected, Idle, Queued, Working, Unresponsive, Blocked, Paused, and Cancelling as applicable. Provide manual kick, pause, resume, and cancel.

FR-17: Configure acknowledgement/activity timeouts, task deadlines, retry limits, and escalation. Distinguish busy agents from stalled agents. Wake, steer, reconnect, cancel, or resume only when supported. Serialize operations per session and avoid duplicate active tasks. Reconcile uncertain results before retrying potentially completed work.

FR-18: Log kicker failures and recovery attempts in searchable operational logs and immutable audit events. Remote cancellation may be unconfirmed; show this explicitly rather than claiming disconnection stopped work.

## Immutable history and retention

FR-19: Accepted chat messages, streamed events, task changes, profile changes, memory changes, and administrative actions create append-only records. No actor can edit or overwrite committed records during protection. Corrections, withdrawals, trash, restore, and resolution append events. Archiving/removing entities cannot cascade-delete history.

FR-20: Each canonical record has a cryptographic checksum, predecessor checksum, unique identifier, author, room/context, commitment timestamp, and sequence. Verify against independently protected checkpoints to detect complete-chain rewriting or truncation. Authorized retention deletion must preserve verification of retained history.

FR-21: A record is committed only when acknowledged by the protected authoritative write area. Pending UI state must be visibly distinct. PostgreSQL permissions alone do not satisfy administrator-resistant immutability. Storage-enforced protection is required; hashing detects alteration, not physical destruction. The technical design must establish the exact enforceable guarantees.

FR-22: Retention is a rolling window from commitment with a minimum of 30 days, and an option of Forever. Each record becomes eligible independently. Trash hides records from normal views and permits restoration without modifying originals. Trashing neither resets nor shortens retention. Permanent purge requires an explicit human action and can affect only eligible trashed records. Forever records cannot be purged. Agents and jobs cannot authorize purges.

FR-23: Every purge leaves a protected audit receipt recording authorization, affected range/identifiers, timestamp, and verification evidence. Define backup/archive deletion behavior and checkpoint retention in technical requirements; do not promise permanent erasure while recoverable copies remain.

## Integrity incident mode

FR-24: Integrity failure means evidence that committed history differs from its protected accepted state: checksum or chain mismatch; unexpected missing/inserted/duplicated/reordered records; checkpoint disagreement; premature deletion; or confirmed unauthorized mutation. Legitimate revisions, disputed facts, and authorized purges are not failures.

FR-25: A confirmed failure triggers application-wide incident mode. Disconnect all agents, halt dispatch, attempt supported cancellation, block agent authentication/reconnection/invitations, and permit authenticated human access only. Humans inspect findings and logs with uncertain records labelled, choose recovery, and explicitly re-invite agents. Never automatically repair originals. Record recovery in an independently trustworthy audit path.

FR-26: Verification unavailable is distinct from confirmed failure. Fail closed for agent work until verification succeeds, without claiming tampering. Failed pre-commit writes are operational errors. Scheduled and read/export verification must have defined coverage and latency in technical design.

## Access and persistence

FR-27: Authenticate remote participants and scope credentials to permitted memberships and memory. Agents cannot impersonate humans or claim human purge/override authority. Preserve application state and reconcile incomplete work after restart.

## Web interface and interaction requirements

The approved demo establishes the interface baseline: a persistent application header, project/room navigation, a primary work area, card-based content, and clear status indicators. Preserve this organization while replacing simulated operations with real, authorized workflows. The interface must remain usable on narrow screens, with labelled inputs, keyboard-operable controls, readable statuses, and confirmation dialogs that explain consequential actions.

FR-28: Navigation exposes Room, Connections, Shared memory, History & retention, and Settings. The current project and room remain visible. Project and room selection and management must be added to the demo's single-room navigation. Switching views must not discard unsaved work without notice or stop background work. The header identifies deployment and signed-in human; production removes the demo label and never implies simulated data is live.

### Room screen

FR-29: The room header shows its name, project, objective, and participation mode. Provide Invite agent and Request squad status actions. The main area shows the conversation and task workflow; the companion area shows the squad and directory boundary. Users can inspect the profile and session state of each participant without leaving the room.

FR-30: Conversation entries show author, timestamps, reply context, delivery/commitment state, and relevant task association. The composer sends messages and supports participant mentions and reply targets. Pending, committed, failed, and streamed output are distinguishable. A failed send preserves draft content and offers a safe retry. Trash hides content from the normal conversation, not from the protected source record. Profile changes do not change the recorded author identity of older entries.

FR-31: Each task card shows its current state, developer/owner, reviewer, dependencies, and acceptance authority. Actions are available according to state and authority: submit work, start review, request changes with findings, resume development, and accept. The production interface must not permit arbitrary state jumps offered by demo simulation buttons. Human overrides remain explicit and audited. Agent operations awaiting acknowledgement display pending state; a click alone cannot mark work complete.

FR-32: Squad cards show participant name, role, seniority, connection health, work state, and acceptance authority. Provide Profile, Kick, Request status, Pause/Resume, Cancel, and Remove where applicable. Kick displays the selected recovery operation and its outcome, including unsupported capabilities and unconfirmed cancellation. Removing (kicking out) a participant immediately revokes room participation, cancels supported active local work, prevents further delivery or late replies, and appends a removal event. Its history and reusable connection remain. This is separate from the watchdog wakeup action. Removed sessions cannot be silently reactivated by late runner events.

### Invitation and profile dialogs

FR-33: Invite agent selects an already registered connection, a supported session method, and room-specific profile fields. Existing-session selection identifies the actual conversation and prevents duplicate live attachment. Display capability limitations and directory enforcement before activation. If no connections are available, offer a route to plug one in. Re-invitation after an incident requires human action and renewed access checks.

FR-34: Profile editing exposes name, role, instructions, seniority, acceptance authority, narrowed permissions, and kicker overrides. Save appends a profile revision and indicates when it takes effect. Cancel leaves the profile unchanged. Validate required fields and numeric values. Equal seniority is permitted and explicitly means consensus for fallback decisions. A template may populate these fields without binding subsequent edits to other participants.

### Connections screen

FR-35: Show registered connections in a table/list with name, adapter, transport, health, and capabilities. Plug in agent opens a configuration form for name, adapter, transport, endpoint or executable, connection direction where relevant, and authentication. Never expose stored secrets in lists, logs, or profile forms. Provide real Test, Edit, Disable, and Remove operations. Test reports actual outcomes, supported methods, and actionable errors. Disabled or unavailable connections cannot start new work; removal must not erase historical references.

### Shared memory screen

FR-36: Provide All, Room, Project, and Application filters, plus search and Write memory. Entries show title, content, scope, provenance, author, timestamps, revision, and dispute state. Write defaults to Room. Changing to a broader scope makes its visibility clear. Editing creates a revision; users can inspect prior versions, trash, and restore. Show retrieval/loading/empty/error states without replacing missing data with fabricated content.

FR-37: A disputed entry exposes the competing revisions and a linked squad discussion, the human response deadline, and responsible senior agents. Resolve as human requires written direction and records the resolution as a revision. Expiry transitions to agent decision or equal-rank consensus; it does not silently resolve the dispute. Show deadlock and escalation. Demo conflict-injection buttons are replaced by conflicts detected during actual writes; production never invents competing revisions.

### History and retention screen

FR-38: Show integrity state, configured retention, protected-until timestamps, trash state, and purge eligibility for individual records. Provide history search/filtering, trash, restore, record details, and human-only purge. Details expose checksum, predecessor reference, sequence, verification result/time, and source references. The 30-day minimum uses elapsed protection from commitment; additional supported calendar presets must have precisely defined calculation rules. Forever cannot expire.

FR-39: Retention changes explain their effect before saving and cannot shorten a previously enforced protection lock. Trashed records remain protected for their original window. Purge remains unavailable for untrashed, protected, or Forever records. Human confirmation lists exactly what will be purged and distinguishes reversible trash from permanent deletion. Successful purge shows an audit receipt; failure must not falsely indicate deletion. Authorized purge evidence must allow integrity checks of remaining history.

FR-40: Audit events are searchable by time, room, actor, task, and event type. Show administrative, profile, memory, task, recovery, and purge actions. Provide separately labelled operational error logs with correlation to audit records. Protect original records when presenting annotations or derived views. Demo fixture insertion and simulated verification badges are excluded from production.

### Settings and room directory controls

FR-41: Settings exposes acknowledgement timeout, activity timeout, task deadline, retry limit, human decision timeout, and consensus timeout, with scope and inheritance clearly shown. Validate settings, save them persistently, and append an audit event. Values displayed by the demo are examples, not adopted production defaults. Participation limits and retention policy management must also be available here or linked to the relevant screen.

FR-42: Directory editing provides structured rows for path and read/write permission, with add/remove controls and access validation. Show a before/after summary distinguishing expansion from narrowing. Saving a narrowing pauses affected execution and displays pending enforcement until adapters confirm. Do not claim resumed or restricted execution merely because configuration was saved. Remote paths are labelled as belonging to the relevant remote workspace, not assumed to be local paths.

### Incident mode and recovery

FR-43: Integrity failure shows a persistent application-wide banner and directs humans to findings and logs. Disable agent invocation, invitations, connection tests that invoke agents, and reconnection; enforce those restrictions beyond the interface. Human inspection and permissible administrative actions remain available. Indicate unresolved remote cancellation separately. Show an unverified integrity state separately from confirmed failure.

FR-44: Human recovery requires an incident review and an explicit recorded decision; clearing a banner does not constitute recovery. Present integrity evidence and remaining restrictions. Only after the recovery conditions are satisfied may the human authorize reopening agent access. Existing agents remain disconnected until explicitly re-invited. Never rewrite failed records to obtain a healthy badge. Simulation/fault-injection controls exist only in isolated test/demo environments.

### Persistence and feedback

FR-45: A page refresh or browser disconnect preserves committed conversations, profiles, settings, memory, task state, and audit events. Work and watchdogs run independently of browser presence. Reconnecting the browser reconciles current state and avoids duplicated messages/actions. Dialogs and notifications report pending, successful, rejected, and failed operations accurately. The initial single-user identity must be enforced server-side, not inferred from button labels.

FR-46: Agent invitations and room profiles offer Additional context, optional and empty by default. Humans may paste notes or load a saved Codex conversation for text preview, select/edit excerpts, and explicitly save the material to that participant. Store context and source references in profile revisions and supply them on agent turns. Imported source conversations remain separate from room history. Loading a preview alone does not share it with the agent. Indicate size limits and excluded attachments/tool output; do not claim a text excerpt clones full runtime state.

## UI acceptance walkthrough

1. Register and test a connection; verify failure details and unavailable capability states. Invite it with a room profile and a supported new/existing session choice.
2. Edit a profile, cancel another edit, and confirm revisions and next-turn application. Use the same connection in another room without sharing a live session.
3. Send a message, observe pending/committed state, request squad status, kick a participant, and verify acknowledgement/error feedback.
4. Submit work, have QA request changes with findings, resubmit, and accept using appropriate authority. Reject unauthorized agent acceptance.
5. Write Room memory, filter/search it, create a broader-scoped entry, and inspect revisions. Exercise a real concurrent-write dispute, human direction, timeout, consensus, and deadlock.
6. Trash and restore a record. Reject early/Forever purge; confirm eligible human purge and inspect its receipt and retained-history verification.
7. Narrow directories and verify pause until enforcement confirmation. Simulate enforcement failure without displaying success.
8. Trigger integrity failure in a test environment: confirm every agent route is blocked, review logs as human, authorize recovery, and explicitly re-invite participants.
9. Refresh and reconnect the browser; confirm persistence and no duplicate dispatch. Complete the main flows using keyboard navigation and a narrow viewport.

## Acceptance scenarios

1. Register one CLI and one HTTP connection; invite both using distinct profiles. Invite the same connection into another room with separate context.
2. Choose supported existing-session attachment without silently discarding context; reject unsupported attachment modes.
3. Assign concurrent tasks, request status, submit developer work, invite QA, request changes, and accept through an authorized participant.
4. Create room memory, explicitly create broader memory, detect concurrent edits, consult a human, then exercise timeout and equal-seniority consensus paths.
5. Kick an idle/stalled participant without launching a duplicate task. Log unsuccessful recovery.
6. Enforce multiple room directories. Pause work when access narrows and demonstrate no new permissions from invitation.
7. Trash and restore a protected message. Reject early purge, allow human purge of expired trashed history, and preserve retained-chain verification. Reject purge of Forever history.
8. Alter a committed record in an isolated test system: verification detects it, every agent is disconnected, and only humans can recover. Separately simulate unavailable verification.

## Remaining technical discovery

- Actual integration and existing-session capabilities of Jerry and this assistant.
- Concrete adapter/transport contracts, authentication, cancellation, and remote directory mapping/enforcement.
- Consensus/deadline defaults, adapter capacity, and resource limits.
- Immutable storage, canonical serialization, checkpoint topology, commitment protocol, retention choices above 30 days, backups, and purge verification.
- Local installation and future multi-user migration.

## UI demo boundary

The demo is browser-local simulated state. It does not call models, enforce directory permissions, authenticate agents, provide PostgreSQL, or implement immutable storage. Its incident and purge controls illustrate the intended workflow only.

## Implemented room and squad expansion
Room creation and room name/description editing are available from the sidebar. Existing unscoped records project into Engineering room without modifying stored events. Settings sits at the sidebar bottom. Room views provide Conversation, Tasks, and Memory tabs. Connections remain global; sessions and profiles are independent per invitation.

Tasks store acceptance criteria, owner, reviewer, status and optimistic revision. Explicit dispatch and review requests send task context to an agent; humans update the task lifecycle. Evidence records are append-only, checksum-covered events with Changes, Checks, Artifacts and Review categories. Human-submitted content is explicitly labelled and never counted as verified execution. Automatic evidence collection and workflow orchestration remain future work.

Profiles include editable personal names, engineering discipline, level, speciality, room assignment and challenge 0–4, default Constructive. Titles and behavioural challenge do not confer permissions or decision authority. Personal names are generated from a mixed name pool. Connection provider metadata selects local OpenAI, Claude, Meta, DeepSeek or Gemini SVG icons.

Human room membership: the single local human appears in the squad with an editable personal name and decision seniority, default 11, range 1–11. Agent seniority is 1–10. Human demotion defers routine ranked decisions; explicit human overrides remain available except immutable history. Automated arbitration remains pending.

Human membership is optional per room and defaults to present. Leave removes the human from the squad without deleting their profile or history; Join restores membership with the saved rank and name. Room membership does not revoke the local owner’s administrative access. Automatic human-absence arbitration remains pending.

Company: one local company owns the human profile, agent connections, room participants and rooms. Settings supports editing its name and description through audited events. Existing records are projected into the default company without rewriting history. Room profiles remain independent. Multiple-company switching and tenant isolation are not implemented.

Invitation-only humans: owner creates distinct accounts (duplicate display names permitted), assigns Developer/QA roles, enables/disables accounts and controls shared-room memberships. Browser generates each invited human’s RSA-3072 credential and downloads a JSON invitation key file containing the account ID and private key; server stores only the public key. No public signup exists. Members can edit their own name but cannot change roles, memberships, connections or company settings. Developer may save workflows in accessible rooms; QA may read workflows. Task acceptance, agent control, shared-scope changes and evidence creation remain owner-only in this first permission set.
Every identity has a deterministic private personal room. Owner may administratively access any personal room; state reads in another identity’s personal room append an access audit event. Disabling a human blocks subsequent authenticated requests and login without deleting their personal room or history. Bot organizational identities presently map to registered connections; room profiles and session IDs remain separate. Bot API credentials and room-move workflows remain pending.
