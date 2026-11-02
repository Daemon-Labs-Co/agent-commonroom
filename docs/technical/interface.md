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

View selections, draft/dirty state and documented preferences may change. Documentation reorganization adds no product execution authority or feature toggle. Legacy document IDs remain mapped to retained references; current user guide IDs are separately available.

COMMONROOM_CONNECTION_EDITOR controls this presentation layer, defaults true and
is changed only through operator process configuration with a restart. Its existing
server gate returns an empty script when disabled, retaining the legacy editor.
Guidance creates no connections or provider calls. Authentication, owner permissions,
connection validation and mandatory isolation remain enforced independently.

Source: [documentation.go](../../cmd/server/documentation.go) · [http_state.go](../../cmd/server/http_state.go).
