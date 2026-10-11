# Product concepts and usage

[Documentation index](README.md)

## Company, humans, and rooms

With `companyOrgChart` enabled (the default), [company reporting relationships](#company-organization-and-decision-authority) replace numbered decision seniority. Older seniority descriptions refer to the retained legacy mode used when this feature is disabled.

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

## Squad profiles and invitations

Click a person's or agent's **name in the Squad** to open their profile. Name controls support keyboard focus and activation. Your human row no longer needs separate Edit and Leave buttons. **Leave room** appears in your own room-profile overlay when you are the owner in a shared room. Leaving removes your presence from the squad; it preserves your account, access and chat history. Use **Join room** to return. Personal rooms cannot be left. Invited humans retain owner-assigned access; they cannot use this presentation change to revoke or change it.

You can edit your own allowed profile fields. The owner can open another invited human's membership editor; other viewers see a read-only human profile. Clicking an agent name opens its existing profile editor; saving still requires the existing server permissions. Names, roles and room access retain their existing authorization rules.

**Invite agent** opens a wider overlay with **Agent**, **Role**, and **Context** tabs. One section is shown at a time so the primary action stays easy to find. Role fields use two columns on larger screens. Starting context is optional. Choose an existing connection, edit the suggested name, and set the room role. Engineering title and challenge settings are expanded for easier access. Notes and conversation excerpts remain optional. Select **Invite agent** to save or use Cancel/the top-right X to dismiss. At narrow screen widths fields stack into one column. Arrow keys, Home and End navigate the section tabs. Overlays aim to fit normal desktop windows; very small viewports or enlarged text retain overflow access rather than hiding controls. New agents retain seniority 1 and cannot accept work until their profile is changed. When no connection is registered, plug one in first.

### Cosmetic feature flags

The owner can toggle these in **Settings → Feature flags → Cosmetic**:

| Flag | Default | Disabled behavior |
| --- | --- | --- |
| `squadProfileActions` | On | Restores the previous squad Edit/Leave controls. |
| `agentInvitationLayout` | On | Restores the previous compact invitation form. |
| `standardHelpIcons` | On | Restores the previous contextual-help text buttons. |
| `featureFlagTabs` | On | Restores the previous separate feature cards. |
| `historyDateGroups` | On | Restores the original ungrouped activity list. |

Saving reloads the browser; no server restart is required. Finish or cancel unsaved forms before changing settings. Flags affect presentation only and never clear profiles, room memberships, photos or history. Only the owner may change flags; API authorization and immutable-history protections apply in both states.

The server gates `/squad-profiles.js` and `/agent-invitation.js` against persisted flags, returning HTTP 204 when disabled, with no-store caching. The modules also check the visible state before applying layouts. Existing event endpoints perform all edits and leave actions; no new privilege or agent execution path is introduced. Regression tests use synthetic UI fixtures and mocked database interactions to cover flag delivery, owner permissions, name activation, leave eligibility and invocation, read-only profiles, invitation escaping, and disabled fallbacks. No live database or provider is used in automated tests.

Contextual help uses the same question-mark icon with a descriptive tooltip and accessible label. Overlay help sits beside the close X; field-specific help sits beside its label. These controls retain the existing help-window preference and destinations. Navigation keeps its readable Help & documentation label. The `standardHelpIcons` flag uses the same owner-only, server-gated cosmetic delivery contract at `/help-icons.js`; disabled presentation never disables access to documentation.

### Feature-category tabs

Feature category tabs supply the category context: a card inside Security is titled "Connection sign-in", rather than "Security · Connection sign-in". The same applies to Cosmetic, Connections and Workflows, and repeated category-only subheadings are omitted. This presentation is part of the existing `featureFlagTabs` flag; disabling tabs retains the category prefixes in the legacy list. Controls, flag names and permissions are unchanged.

**Settings → Feature flags** brings Security, Cosmetic, Connections and Workflows into one tabbed section. Click a tab or use Left/Right, Home and End while focused on a tab. The selected category stays selected while settings rerenders. Controls retain their existing actions and permissions; unrelated profile/company settings remain outside this section. The owner-only `featureFlagTabs` cosmetic flag defaults on and can be changed in the Cosmetic tab. Turning it off restores the previous grouped cards after reloading. The server gates `/feature-flag-tabs.js`; toggling never disables authentication or safety enforcement. Environment-configured features still require a server restart and are labelled accordingly.

Field explanations are placed to the right of their field in the standardized-help presentation. Question-mark field help remains beside the label. The `standardHelpIcons` flag also controls this presentation; it changes no saved values or validation. With the tabbed invitation, changing sections preserves entered values. Required-field validation opens the relevant tab so errors are reachable.

### Human login persistence

After signing in with your PEM, refreshing the page retains the authenticated session. Server restarts also retain unexpired sessions when `COMMONROOM_PERSISTENT_SESSIONS=true` (the default). You do not need to provide the PEM again until the existing eight-hour session expires or you sign out. The PEM is never placed in browser storage or this server session file. Cookie protections, origin checks, CSRF and actor permissions remain enforced.

The server stores only opaque cookie-token hashes, actor identities, CSRF values and fixed expiry times in the owner-key directory's private `human-sessions.json` file (normally `.data/human-sessions.json`). Protect this directory and its backups. The file has mode 0600, bounded content, atomic replacement and binding to the owner public key. Missing files start with no sessions; unsafe permissions, symlinks or invalid data fail startup closed. Bootstrap/setup sessions are not restored. Changing the owner key invalidates saved sessions. This file is not served by the web server or included in immutable room history.

Set `COMMONROOM_PERSISTENT_SESSIONS=false` and restart to use memory-only sessions. Disabling removes the saved session store so re-enabling cannot resurrect earlier sessions. Authentication remains required in both states; this flag does not weaken verification or extend expiry. The Security tab describes this deployment flag; it cannot be changed by an invited human or an agent through the application. Logout saves revocation before reporting success; storage failure reports an error so you can retry. This private-file store supports one server process and local/private deployment, not shared sessions across multiple instances. Automated temporary-file tests cover reload/restart, expiry, identity changes, durable logout, failures and both flag states. Real browser cookie restoration and host-reboot behavior remain deployment checks.

### History date tabs

History & retention separates activity into tabs across the top: **All activity**, **Today**, **Yesterday**, **Earlier this week**, **Last week**, and month/year tabs for older records. Empty date buckets are omitted. Calendar dates use your browser's local timezone; weeks start Monday. Future timestamps and invalid dates receive explicit Future dates/Unknown date tabs rather than misleading labels. Search filters the selected period, keeping date navigation available when no result matches. Left/Right, Home and End navigate tabs; the selected period is retained during rerenders and falls back to All activity if its bucket disappears.

This is a read-only projection of the existing latest-200-record history endpoint. It does not rewrite timestamps, descriptions, checksums or immutable events, and it is not full historical pagination. The owner-only Cosmetic flag `historyDateGroups` defaults on, applies after saving/reloading without restart, and restores the original ungrouped activity list when off. The server gates `/history-date-groups.js`; record access, retention and integrity enforcement remain unchanged. Tests cover local date boundaries, daylight-saving changes, Monday/week transitions, month/year rollover, search, empty periods, tab isolation, and both delivery states with mocked database interactions.

### Settings sections and retention

The Settings page uses **Profile**, **Company**, **Features**, **Retention**, **Help**, and **About** tabs. Tabs appear only when there is content the current human is permitted to see. The Features section contains its own Security/Cosmetic/Connections/Workflows tabs. Your own allowed profile fields and help-window preference remain editable; invited humans do not gain company, membership, feature-management or retention permissions. Existing button actions are preserved when controls move between sections. The selected settings section is retained during rerenders; Arrow keys, Home and End provide keyboard navigation.

The sidebar now calls the activity page **History**. The existing retention-policy notice moves to **Settings → Retention**. The minimum remains 30 days; disposal and purge are still disabled pending storage enforcement and protected receipts. This layout change does not implement deletion or weaken immutable history. The owner-only `settingsTabs` cosmetic flag defaults on; saving reloads the browser without requiring a restart. Disabled presentation restores the original Settings cards and History & retention layout. The server gates `/settings-tabs.js`, while all underlying authorization and retention protections remain enforced in both states.

### Public and private room groups

The sidebar groups shared company rooms under **Public rooms** and personal rooms under **Private rooms**. These labels describe the existing room types, not internet exposure or a new access policy. Only rooms already visible to the signed-in human are listed; room permissions and the owner's audited personal-room access remain unchanged. New room creation remains with Public rooms; each invited human's personal room appears under Private rooms. Arbitrary shared/private visibility switching is not introduced.

The owner-only `roomSections` Cosmetic flag defaults on. Saving reloads the browser without a server restart; disabling restores the single Rooms list. The server gates `/room-sections.js`. Grouping moves existing navigation controls and preserves their handlers, selected state and creation permissions. Tests cover personal/shared classification, movement, retained handlers, disabled fallback and server flag permissions without live database interactions.

Settings content is anchored to the top-left of the main area. Its heading and tab bar keep a stable position when sections have different amounts of content or the viewport changes; this basic layout correction applies in both settings-layout flag states.

Click either Public rooms or Private rooms heading to collapse/expand that list independently. Both start expanded. The browser remembers your choice per human account on that browser; it does not change membership, permissions or the selected room. The shared-room creation button remains accessible while Public rooms is collapsed. Heading buttons support keyboard activation and expose expanded/collapsed state to assistive technology. Unavailable or malformed browser storage falls back to expanded lists. This behavior is part of the existing `roomSections` flag.

Settings help is anchored at the far right of the section tab bar, separate from the tabs, retaining its existing destination and same/separate-window preference.

Room section headings use larger, bold text in the main button text color with cobalt disclosure arrows. This makes Public rooms and Private rooms distinguishable from individual room links while preserving keyboard activation and collapsed-state indicators. This refinement uses the existing `roomSections` flag.

### Human account and room details

Settings → Profile and human squad-name overlays show read-only company, assigned role, account status, account ID, personal room and shared room access. A room profile also shows the current room, presence and decision seniority. For example, an invited Developer sees their own personal room and visible assigned shared rooms. No new bio, contact email or timezone is collected. Only rooms already in the authorized state are displayed; keys, provider credentials and connection endpoints are never included. The owner still controls assigned roles and access; showing a profile does not grant editing permission.

The owner-controlled Cosmetic flag `humanProfileDetails` defaults on. Save it under Settings → Features and reload; no restart is required. Turning it off removes the extra read-only details while keeping the existing profile editing and membership controls. The server gates `/human-profile-details.js`; the module projects existing authorized state without adding API mutations or changing authorization. Tests cover escaping, room filtering, key omission, human seniority, existing actions, both delivery states and owner-only flag changes, with mocked database interactions.

### Page alignment and compact profile overlays

Application pages start at the top-left of their content area regardless of how much content they contain. This includes conversation, connections, tasks, templates, workflows, memory, history, Settings and the sign-in page. Existing content width limits, responsive padding and independent chat/squad scrolling remain in place. The owner-controlled Cosmetic flag `topAlignedViews` defaults on, is enforced by server delivery of `/viewport-alignment.js`, and applies after saving/reloading without a restart. Turning it off restores the previous page positioning, while the earlier Settings anchoring fix remains active.

Agent profile editing now uses **Agent**, **Role**, **Behaviour** and **Context** tabs. Existing title/challenge fields appear in a two-column Role panel; instructions, seniority and work-acceptance permission appear under Behaviour. Context remains optional. Human profiles use **Profile** and **Account & rooms** tabs when read-only account information is enabled. Short forms without extra information stay as a single panel. Save, Cancel and room Leave actions remain outside the changing panels. Arrow keys, Home and End navigate tabs, and browser validation reveals the panel containing an invalid required field before correction.

The owner-controlled Cosmetic flag `compactProfileOverlays` defaults on. The server gates `/overlay-layout.js`; switching it off and reloading restores the existing profile form layout without losing field values or profile data. It does not change who may edit roles, access or work acceptance. This module moves existing controls rather than copying them, preserving handlers and FormData, and works independently of invitation and squad-name presentation flags. No restart is needed. Automated tests verify control preservation, validation, keyboard navigation, flag combinations and permission checks with mocked database interactions.

Tabs reduce normal desktop scrolling in profile overlays; small viewports, increased text size, long room lists or expanded conversation results may still require scrolling to keep information accessible. Other long overlays, such as member access lists and evidence, have not yet received a full layout redesign. Modal dialogs retain their dialog placement; the page alignment flag applies to page content.

### Company organization and decision authority

Open **Settings → Company → Organization chart** as the owner. The workflow-style screen lists all registered human and agent company identities, including disabled accounts, and shows reporting connectors on a dotted canvas. Select an identity in the roster, canvas or list view to see its properties. Agents show their room-specific names and roles separately: one registered connection has one company identity even when several room profiles use it. A removed identity is omitted from the live projection; references to a removed manager become unassigned until the owner saves a new chart. Company identity metadata is visible only to the owner through this editor; no conversation, instructions, credentials, keys or endpoints are exposed by its API.

Double-click an identity on the canvas, roster or list to edit it: agents open **Edit connection**, your owner identity opens **Your profile**, and other humans open the existing member profile/access editor. The properties panel also has **Edit connection** or **Edit profile** for keyboard access. These owner-only shortcuts use the existing authorized editors; they do not grant new permissions. The chart stays open behind the editor, retaining unsaved reporting lines and positions. After the editor closes, identity names and room-profile details reload; chart changes still require **Save chart**. Missing or removed connections show an explanatory message. This interaction uses the same default-on, server-gated `companyOrgChart` flag described below.

Use **Reports to** to choose a manager. Drag nodes, or focus a canvas node and use arrow keys (Shift for a larger step). Use zoom, canvas scrolling, **Auto-arrange** or the keyboard-accessible **List view**. Save explicitly; closing with an unsaved draft asks whether to discard it. Changes are appended to immutable history. Concurrent edits return a revision conflict; close/discard and reopen to load the current chart before reapplying changes. Only the owner may save reporting relationships. Up to 500 chart identities and bounded finite positions are supported.

Reporting hierarchy replaces numbered decision seniority while this feature is enabled. A manager outranks direct and indirect reports. Same-level peers, unrelated reporting branches and multiple room profiles of one company identity do not outrank one another; disagreement requires consensus. An absent manager cannot decide a room conflict. Explicit human directions retain the final override, but do not bypass authorization, work-acceptance permission, mandatory container isolation or immutable history. Job titles and adversarialness remain independent. For example, a Staff agent reporting to a Principal agent defers to that Principal, while an unrelated QA lead must reach agreement rather than winning through a number.

The implementation provides ancestry comparison and available-decision-peer calculation, supplies room-scoped authority guidance to managed agent turns, and removes numbered seniority controls/displays. Legacy rank fields remain in saved records and hidden form inputs for compatibility; they have no role in the hierarchy comparison. Existing history is not rewritten and ranks are not automatically converted into reporting relationships. New identities start without a manager. Automatic conflict arbitration, consensus tracking and escalation timers are still unimplemented; guidance is not a security enforcement mechanism or proof that models will obey it.

The owner-controlled `companyOrgChart` flag appears under Settings → Features → Cosmetic and defaults on. Disabling and reloading hides the editor, restores legacy seniority presentation, blocks organization reads/writes on the server and omits hierarchy guidance from subsequent agent turns. Existing chart records remain preserved; active turns are not recalled. No restart is required for flag changes. Tests use mocked databases and synthetic fixtures for persistence/projection, revisions, cycle and position validation, owner/CSRF authentication boundaries, enabled/disabled behavior, secret omission, room-scoped guidance and authority comparisons. Browser checks use a synthetic chart; real authenticated saving and live provider behavior remain manual integration checks.
