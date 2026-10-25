# Interface and personal preferences

[Documentation home](README.md) · [Product guide](concepts.md)

Choose profile, room and display behavior without changing account permissions.
With the default tabbed Settings, open **Settings → Features** for owner controls.
The untabbed fallback calls the same section **Feature flags**. Personal Help
preferences live under Settings → Help.

## Human profile photos and feature flags

In Settings, choose **Edit profile → Change profile photo**. The same option is available from your own squad profile. Upload a JPEG or PNG up to 5 MB, inspect the preview, and save. The browser reduces it to a maximum of 160 pixels per side; the server validates bounded raster data and re-encodes it as JPEG without original metadata. SVG, GIF and remote image URLs are rejected. Initials remain the fallback when no photo is set.

The photo belongs to the human account, rather than a room or role. It appears in each room's squad and beside messages identified by that human's account ID. Earlier messages display the current profile photo; their stored text, authors and immutable records are not rewritten. Legacy messages without an account identity retain initials unless explicitly identified as the original “You” owner. Members can change only their own image; they cannot change their roles or another account's photo. Owner access edits preserve the person's photo.

Photo revisions are append-only audit events. Removing a photo clears its current display; earlier image records remain. Audit purge is not implemented. Do not upload images you need immediately erased from all historical records.

Owners see grouped flags under **Settings → Features**:

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

**Invite agent** opens a wider overlay with **Agent**, **Role**, and **Context** tabs. One section is shown at a time so the primary action stays easy to find. Role fields use two columns on larger screens. Starting context is optional. Choose an existing connection, edit the suggested name, and set the room role. Engineering title and challenge settings are expanded for easier access. Notes and conversation excerpts remain optional. Select **Invite agent** to save or use Cancel/the top-right X to dismiss.

At narrow screen widths fields stack into one column. Arrow keys, Home and End navigate the section tabs. Overlays aim to fit normal desktop windows; very small viewports or enlarged text retain overflow access rather than hiding controls. New agents cannot accept work by default. Work acceptance has its own permission setting; legacy seniority values do not grant authority. When no connection is registered, plug one in first.

## Cosmetic feature flags

The owner can toggle these in **Settings → Features → Cosmetic**:

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

## Feature-category tabs

Feature category tabs supply the category context: a card inside Security is titled "Connection sign-in", rather than "Security · Connection sign-in". The same applies to Cosmetic, Connections and Workflows, and repeated category-only subheadings are omitted. This presentation is part of the existing `featureFlagTabs` flag; disabling tabs retains the category prefixes in the legacy list. Controls, flag names and permissions are unchanged.

**Settings → Features** brings Security, Cosmetic, Connections and Workflows into one tabbed section. Click a tab or use Left/Right, Home and End while focused on a tab. The selected category stays selected while settings rerenders. Controls retain their existing actions and permissions; unrelated profile/company settings remain outside this section. The owner-only `featureFlagTabs` cosmetic flag defaults on and can be changed in the Cosmetic tab. Turning it off restores the previous grouped cards after reloading. The server gates `/feature-flag-tabs.js`; toggling never disables authentication or safety enforcement.

Environment-configured features still require a server restart and are labelled accordingly.

Field explanations are placed to the right of their field in the standardized-help presentation. Question-mark field help remains beside the label. The `standardHelpIcons` flag also controls this presentation; it changes no saved values or validation. With the tabbed invitation, changing sections preserves entered values. Required-field validation opens the relevant tab so errors are reachable.

## Human login persistence

After signing in with your PEM, refreshing the page retains the authenticated session. Server restarts also retain unexpired sessions when `COMMONROOM_PERSISTENT_SESSIONS=true` (the default). You do not need to provide the PEM again until the existing eight-hour session expires or you sign out. The PEM is never placed in browser storage or this server session file. Cookie protections, origin checks, CSRF and actor permissions remain enforced.

The server stores only opaque cookie-token hashes, actor identities, CSRF values and fixed expiry times in the owner-key directory's private `human-sessions.json` file (normally `.data/human-sessions.json`). Protect this directory and its backups. The file has mode 0600, bounded content, atomic replacement and binding to the owner public key. Missing files start with no sessions; unsafe permissions, symlinks or invalid data fail startup closed. Bootstrap/setup sessions are not restored. Changing the owner key invalidates saved sessions. This file is not served by the web server or included in immutable room history.

Set `COMMONROOM_PERSISTENT_SESSIONS=false` and restart to use memory-only sessions. Disabling removes the saved session store so re-enabling cannot resurrect earlier sessions. Authentication remains required in both states; this flag does not weaken verification or extend expiry. The Security tab describes this deployment flag; it cannot be changed by an invited human or an agent through the application. Logout saves revocation before reporting success; storage failure reports an error so you can retry.

This private-file store supports one server process and local/private deployment, not shared sessions across multiple instances. Automated temporary-file tests cover reload/restart, expiry, identity changes, durable logout, failures and both flag states. Real browser cookie restoration and host-reboot behavior remain deployment checks.

## Public and private room groups

The sidebar groups shared company rooms under **Public rooms** and personal rooms under **Private rooms**. These labels describe the existing room types, not internet exposure or a new access policy. Only rooms already visible to the signed-in human are listed; room permissions and the owner's audited personal-room access remain unchanged. New room creation remains with Public rooms; each invited human's personal room appears under Private rooms. Arbitrary shared/private visibility switching is not introduced.

The owner-only `roomSections` Cosmetic flag defaults on. Saving reloads the browser without a server restart; disabling restores the single Rooms list. The server gates `/room-sections.js`. Grouping moves existing navigation controls and preserves their handlers, selected state and creation permissions. Tests cover personal/shared classification, movement, retained handlers, disabled fallback and server flag permissions without live database interactions.

Settings content is anchored to the top-left of the main area. Its heading and tab bar keep a stable position when sections have different amounts of content or the viewport changes; this basic layout correction applies in both settings-layout flag states.

Click either Public rooms or Private rooms heading to collapse/expand that list independently. Both start expanded. The browser remembers your choice per human account on that browser; it does not change membership, permissions or the selected room. The shared-room creation button remains accessible while Public rooms is collapsed. Heading buttons support keyboard activation and expose expanded/collapsed state to assistive technology. Unavailable or malformed browser storage falls back to expanded lists. This behavior is part of the existing `roomSections` flag.

Settings help is anchored at the far right of the section tab bar, separate from the tabs, retaining its existing destination and same/separate-window preference.

Room section headings use larger, bold text in the main button text color with cobalt disclosure arrows. This makes Public rooms and Private rooms distinguishable from individual room links while preserving keyboard activation and collapsed-state indicators. This refinement uses the existing `roomSections` flag.

## Human account and room details

Settings → Profile and human squad-name overlays show read-only company, assigned role, account status, account ID, personal room and shared room access. A room profile also shows the current room, presence and any legacy rank presentation. For example, an invited Developer sees their own personal room and visible assigned shared rooms. No new bio, contact email or timezone is collected. Only rooms already in the authorized state are displayed; keys, provider credentials and connection endpoints are never included. The owner still controls assigned roles and access; showing a profile does not grant editing permission.

The owner-controlled Cosmetic flag `humanProfileDetails` defaults on. Save it under Settings → Features and reload; no restart is required. Turning it off removes the extra read-only details while keeping the existing profile editing and membership controls. The server gates `/human-profile-details.js`; the module projects existing authorized state without adding API mutations or changing authorization. Tests cover escaping, room filtering, key omission, legacy rank presentation, existing actions, both delivery states and owner-only flag changes, with mocked database interactions.

## Page alignment and compact profile overlays

Application pages start at the top-left of their content area regardless of how much content they contain. This includes conversation, connections, tasks, templates, workflows, memory, history, Settings and the sign-in page. Existing content width limits, responsive padding and independent chat/squad scrolling remain in place. The owner-controlled Cosmetic flag `topAlignedViews` defaults on, is enforced by server delivery of `/viewport-alignment.js`, and applies after saving/reloading without a restart. Turning it off restores the previous page positioning, while the earlier Settings anchoring fix remains active.

Agent profile editing now uses **Agent**, **Role**, **Behaviour** and **Context** tabs. Existing title/challenge fields appear in a two-column Role panel; instructions and work-acceptance permission appear under Behaviour; legacy rank controls may appear when organization-chart presentation is disabled. Context remains optional. Human profiles use **Profile** and **Account & rooms** tabs when read-only account information is enabled. Short forms without extra information stay as a single panel. Save, Cancel and room Leave actions remain outside the changing panels.

Arrow keys, Home and End navigate tabs, and browser validation reveals the panel containing an invalid required field before correction.

The owner-controlled Cosmetic flag `compactProfileOverlays` defaults on. The server gates `/overlay-layout.js`; switching it off and reloading restores the existing profile form layout without losing field values or profile data. It does not change who may edit roles, access or work acceptance. This module moves existing controls rather than copying them, preserving handlers and FormData, and works independently of invitation and squad-name presentation flags. No restart is needed. Automated tests verify control preservation, validation, keyboard navigation, flag combinations and permission checks with mocked database interactions.

Tabs reduce normal desktop scrolling in profile overlays; small viewports, increased text size, long room lists or expanded conversation results may still require scrolling to keep information accessible. Other long overlays, such as member access lists and evidence, have not yet received a full layout redesign. Modal dialogs retain their dialog placement; the page alignment flag applies to page content.

## Room message order

In **Settings → Features → Cosmetic**, the owner can enable **Newest room messages at top** (`newestMessagesFirst`). It defaults **off**: rooms keep older messages above newer messages, with new messages at the bottom. When enabled, newest messages appear at the top. Receiving or sending a message scrolls to the selected edge; ordinary refreshes preserve the current reading position. Switching rooms opens at the newest edge. The composer stays below the conversation in either mode.

The setting applies across this company installation, not separately per human/room. It takes effect after saving and automatic page reload; no server restart is required. Disabling restores bottom-first placement without deleting or rewriting any records. Server-side owner permissions protect flag changes, and module delivery is gated off by default (`/room-message-order.js` returns 204 when disabled). A stale loaded module also checks the current flag before changing display order.

Technically, `arrangeNewestRoomMessages` moves existing rendered article nodes only after provider icons and human photos are applied. `state.messages`, API chronological order, provider prompts, timestamps, audit checksums and retention stay unchanged. The wrapper remembers the room/newest-message identity so background refreshes preserve scroll position. The normal send handler and new-message scrolling choose the same edge. Automated tests cover both flag states, default-off delivery, owner/non-owner permissions, nonmutation, avatar identity, incoming messages, refresh position, room changes and empty/non-room views, using synthetic fixtures and mocked database interactions.

## Room viewport and feature switches

Two owner-controlled Cosmetic flags default **on**: `roomViewportFit` (**Room fits visible window**) and `featureFlagSwitches` (**Feature on/off switches**). Their modules are gated by the server and also check current state. Turning either off restores the preceding presentation, without changing room access, records, retention or security policy. Save/reload is automatic; no server restart is required.

Room fitting keeps the desktop page header, room controls and message composer inside the visible window. At widths of 600 pixels or more, the room uses a bounded dynamic viewport height: messages and an overflowing squad scroll within their panels. Entering a room resets document scroll to the top; background renders preserve inner conversation reading position. At narrower widths, normal document scrolling remains available. Workflow/task views and Settings retain their own layouts.

Feature flags appear as rounded on/off switches in Settings → Features, with labels, visible On/Off state, keyboard-operable native buttons and `role="switch"`/`aria-checked`. Cobalt indicates on; gray indicates off. Saving temporarily disables that flag's switch to prevent duplicate requests. Errors restore its preceding state and show the reason; successful saves reload and return to the same feature category. Only the owner can save flag changes, through the existing authenticated event API. Existing emergency stops remain enforced when their controls are hidden. Deployment-only settings that require configuration/restart remain informational rather than presenting switches that cannot change them.

Technical modules: `room-viewport.js` applies/removes the scoped body class only for an active room conversation; CSS uses flex/grid with `min-height:0` and preserves a nonshrinking composer. `feature-switches.js` replaces identified flag controls after settings/category layout, manages per-flag pending saves, and preserves the category through a short-lived session-storage UI preference consumed by `features.js`. Labels use text nodes. Reduced-motion preferences disable knob animation. Tests cover both feature flags and server delivery/permissions with mocked database interactions; synthetic UI tests cover defaults, accessible attributes, duplicate-click prevention, save rollback, unauthorized callers, view transitions and disabled fallbacks.

## Squad action menus

Only one squad action menu stays open. Opening another closes the first; clicking
outside, pressing Escape, scrolling its surrounding area, resizing or navigation
closes it. Escape returns focus to the menu trigger; Arrow Down opens and focuses
the first enabled action. Menus open above the trigger when space below is limited
and remain inside the browser viewport. The scrollable squad no longer adds artificial
bottom padding for an open menu. Existing Message, Edit profile and Remove actions
and their permissions remain unchanged.

The owner-controlled Cosmetic flag `squadMenuLayout` (**Contained squad action
menus**) defaults on. Save/reload applies changes without restart; off restores
native menus and blocks `/squad-menus.js` delivery server-side. No agent execution,
audit, retention or access policy changes. Technical module: `web/squad-menus.js`;
synthetic tests cover exclusivity, edge positioning, dismissal, keyboard focus and
both states, with mocked server tests for flag authorization and script delivery.

## Stable conversation scrolling

With **Room fits visible window** (`roomViewportFit`, default on) enabled,
conversation scroll position is restored after room tabs, task cards, avatars and
message ordering finish rendering. Status-only refreshes keep the current message
edge when within 24 pixels of it, or retain the first visible message and its offset
when reading older material. New messages, room changes or display-order changes
still follow the configured newest-message edge. Both bottom-first and top-first
presentation are supported. Browser scroll anchoring is disabled only for this
explicitly managed message panel to prevent a second correction.

Disabling the existing owner-controlled flag restores the previous render-time
scroll behavior; no extra setting is required. The server continues gating
`/room-viewport.js`, and message history/order/prompts remain unchanged. Technical
helpers `captureRoomScroll`/`restoreRoomScroll` in `web/room-viewport.js` use stable
`data-message-id` attributes and visible-row offsets. Synthetic tests reproduce the
intermediate larger-panel clamp, verify older-message anchoring and both message
orders, and cover disabled behavior. Tests use no database or real messages.

Manual browser regression: `COMMONROOM_BROWSER_SCROLL_TEST=1 go test ./cmd/server
-run '^TestWorkflowBrowserHarness$' -count=1 -timeout 2m -v` serves a 60-second,
localhost-only fixture on port 8794. Its messages and periodic status refreshes
exist only in the in-memory mock; no provider runs and no PostgreSQL writes occur.

## Squad membership count

The badge beside **Squad** counts the humans and agents currently displayed in that room. Inviting or removing an agent, joining or leaving as a human, and switching rooms update it with the rendered membership. Background status and profile changes do not affect the total. Membership changes made elsewhere become visible when room state next refreshes; this fix does not introduce idle-room polling.

Technical contract: `trackSquadCount` in `web/features.js` counts direct `.participant` children of `.squad-members`, updates an accessible member-count label, and observes direct membership changes. Each render disconnects the previous observer, including when leaving the room view. It does not count nested menus or conversation authors. This is a correction to existing membership presentation, always active with no new feature flag or permission changes. Regression tests in `tests/squad-count.test.cjs` use synthetic DOM membership only.


Agents join work rooms rather than receiving an empty personal room automatically.
Legacy agent rooms containing saved history, settings or work remain in the room
list. Owners can restore unused automatic agent rooms with **Settings → Features → Agent
rooms → Automatic agent rooms** (default Off). See the [compatibility and
retention contract](concepts.md#automatic-agent-room-compatibility).


With **Settings → Features → Workflows → Room outcomes** enabled, New room and Room settings collect a defined outcome and acceptance criteria for shared rooms. Room and Tasks show the definition to authorized members. The owner controls edits; saved definitions remain visible when the setting is off. See [room outcomes](room-outcomes.md).


**Settings → Features → Workflows → Room deliverable context** is a separate, default-off owner opt-in for sending the current room definition to task providers. It also requires Room outcomes. Turning it off affects future dispatches; saved definitions remain visible. Provider disclosure restrictions still apply. See [task-provider context](room-outcomes.md#optional-task-provider-context).


**Settings → Features → Workflows → Room outcome review** enables owner acceptance and reopening in Room and Tasks. The dialog shows the current definition, task and evidence snapshot, and requires a human rationale. Changed work shows Needs review. Members can read authorized history; see [room review](room-outcome-review.md).
