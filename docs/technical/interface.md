# Browser and documentation interface

## Input

Authenticated state/CSRF for application mutations; allowlisted Help document IDs and local illustration assets for public documentation.

## Output

Rendered workspace/Help pages, operation requests or escaped/sanitized errors. `GET /instance-label.js` supplies the application process machine hostname from the operating system, independent of request Host, proxy headers and browser URL. The header and connection location card display this name as text; an unavailable hostname is labelled “Server name unavailable”. The response is not cached; HEAD returns no body and other methods return 405.

## Functionality

Stop/resume confirmations close after the server accepts the action, before
refreshing workspace state. Rejections remain visible as text in the dialog's
alert region; they do not clear stop state or imply acceptance. A refresh failure
after acceptance reports a distinct warning without reopening the confirmation.
Existing authentication, owner authority, stop scopes and shutdown verification
remain unchanged.

Browser scripts currently share globals and order-sensitive render wrappers. Frontend controls present permission; backend checks enforce it. Help serves allowlisted documents/images and rejects arbitrary paths/symlinks; Markdown raw HTML and executable URL schemes are escaped/refused. User guides are grouped by functionality; active technical reference uses only the four contract sections.

The guided connection editor starts new provider/method selects empty and required;
editing preserves saved selections, addresses and legacy remote location values.
Focus, click, input and change events select field-specific right-side guidance.
The setup overview appears initially or when all required field values are present;
filled values do not imply validity, provider sign-in or successful execution.

Connection editor tests retain the saved connection revision and provider identity
when protocol/address/provider are unchanged. Unsaved address/method/provider
changes receive readiness checks only and cannot borrow registered credentials
for final verification. Existing stop, flag and journal/shutdown guards remain.

## Properties and state changes

actor.photo_size.updated accepts size small or large for the authenticated enabled
human only; a submitted identity cannot target another account. It requires the
existing humanProfileImages flag (default enabled). The actor photoSize preference
persists across replay, restart and profile revisions. Absent/unknown presentation
values render at 30 pixels; large renders at 60 pixels. Disabling photos hides the
size effect and refuses preference changes without erasing the saved choice.

`actor.chat_style.updated` accepts a boolean `enabled`; `actor.agent_highlight.updated`
accepts `agentId` and a boolean `enabled`. Both apply only to the authenticated human:
a submitted identity is replaced with the session actor. They require the
`chatDisplayChoices` flag (default enabled, owner-only change); while it is off the
server refuses both events with 403, the interface ignores saved values and offers no
controls, and saved values are kept. Adding a highlight requires an existing agent
actor (404 otherwise) and at most 50 highlights per person (409 beyond that);
removing is always accepted. The actor properties `chatStyle` and `highlightedAgents`
persist across replay, restart and profile revisions, and the state response returns
them only to their owner. They are presentation only: `chat-display.js` adds CSS
classes and a text marker to rendered room messages, moving the viewer's own
messages to the right and tinting highlighted agents. Message records, order,
agent prompts, permissions and the organization chart are unaffected.

View selections, draft/dirty state and documented preferences may change. Documentation reorganization adds no product execution authority or feature toggle. Legacy document IDs remain mapped to retained references; current user guide IDs are separately available.

COMMONROOM_CONNECTION_EDITOR controls this presentation layer, defaults true and
is changed only through operator process configuration with a restart. Its existing
server gate returns an empty script when disabled, retaining the legacy editor.
Guidance creates no connections or provider calls. Authentication, owner permissions,
connection validation and mandatory isolation remain enforced independently.

Source: [documentation.go](../../cmd/server/documentation.go) · [http_state.go](../../cmd/server/http_state.go).

### Provider sign-in waiting indicator

- Input: pending connection-authentication check or login stream.
- Output: decorative alternating speech bubbles beside the existing status text.
- Functionality: hidden after completion, failure or dialog cancellation; visible
  while an authentication request is pending. Status text remains available to
  assistive technology; decorative shapes are hidden from it.
- Properties and state changes: reduced-motion preference disables animation.
  The existing connectionSignIn flag and owner permissions govern the parent
  operation; the indicator changes no execution or authorization state.

### Room agent invitation

- Input: participant.created with an existing connectionId, room identity and
  participant profile fields, through the authorized event endpoint.
- Output: a saved room participant and its connection association; no provider
  launch occurs while adding the participant.
- Functionality: Add agent opens the saved-connection invitation in both enabled
  and disabled agentInvitationLayout states. Normal server event authorization,
  stop and profile-validation rules remain enforced.
- Properties and state changes: the obsolete agents action=create returns409
  directing clients to Add agent, without creating a connection or invoking a
  host CLI. Sending messages continues through mandatory isolated execution.

### Squad list scrolling

- Input: roster scroll position, selected room and roomViewportFit (default on).
- Output: the same scroll offset after a same-room render, clamped by the browser
  if the list has shortened. New messages do not reset the roster position.
- Functionality: the desktop member list is the single Squad scroll container;
  heading and footer remain outside it. Room changes start with a fresh list.
- Properties and state changes: disabling roomViewportFit retains legacy layout
  and refresh behavior. No server state, permissions or execution changes.

### Background refresh and editing

- Input: two-second agent-status polling and current focus/form state.
- Output: updated state and rendered room only when editing is inactive.
- Functionality: background refresh checks focus before requesting and before
  applying responses, skips concurrent polls and discards stale results after
  navigation or explicit state changes. Failed requests release the polling slot.
- Properties and state changes: input/textarea/select, editable content and open
  forms defer background updates; subsequent polls resume after editing ends.
  Explicit send/save refreshes retain their behavior. This correction introduces
  no new feature flag and changes no server permissions or execution limits.

### Room human display name

- Input: projected actor name and optional room humanName override.
- Output: owner profile name when the room override is blank, You or Change Me;
  otherwise the explicit room override. Other humans use their actor names.
- Functionality: applies the same resolution to the Squad label and room profile
  form. Trims override whitespace without rewriting stored room/history data.
- Properties and state changes: Settings profile changes flow through inherited
  labels; deliberate room overrides remain. No execution or permission changes.
