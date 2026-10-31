# Browser and documentation interface

## Input

Authenticated state/CSRF for application mutations; allowlisted Help document IDs and local illustration assets for public documentation.

## Output

Rendered workspace/Help pages, operation requests or escaped/sanitized errors.

## Functionality

Browser scripts currently share globals and order-sensitive render wrappers. Frontend controls present permission; backend checks enforce it. Help serves allowlisted documents/images and rejects arbitrary paths/symlinks; Markdown raw HTML and executable URL schemes are escaped/refused. User guides are grouped by functionality; active technical reference uses only the four contract sections.

## Properties and state changes

View selections, draft/dirty state and documented preferences may change. Documentation reorganization adds no product execution authority or feature toggle. Legacy document IDs remain mapped to retained references; current user guide IDs are separately available.

Source: [documentation.go](../../cmd/server/documentation.go) · [http_state.go](../../cmd/server/http_state.go).
