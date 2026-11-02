# Identity, settings and access

## Input

Authenticated human session, CSRF token for mutations, room/account scope and owner-managed feature settings. Agent credentials cannot act as human API sessions.

## Output

Authorized state views, accepted mutations or explicit authentication/permission failures.

## Functionality

Startup initializes an owner key only after checking database history. Empty
history permits RSA-3072 generation and an actor public-key record committed with
an empty-history compare-and-swap. Established history without the original public
key file refuses automatic replacement; pending setup with other human accounts
also refuses. A registered database public key must match the local key. Private
PEM bytes and the bootstrap code never enter event records.

GET /api/setup-status exposes only a no-store setupRequired boolean, under normal
Host/Origin checks. It grants no session or setup authority. First-owner setup
requires the existing private bootstrap code, followed by private-key download and
proof of possession. Returning-user PEM failures do not initiate registration.

Enforce actor/room checks on the server. Feature flags retain their individual defaults and disabled-state behavior. Provider credentials remain outside model tools. Paid organization/multi-user/billing placement does not make security checks optional.

## Properties and state changes

Session CSRF/expiry and feature settings may change. Immutable audit records remain retained. Isolation/authentication/retention cannot be disabled by product flags.

Source: [auth.go](../../cmd/server/auth.go) · [event_mutations_identity.go](../../cmd/server/event_mutations_identity.go) · [http_state.go](../../cmd/server/http_state.go).
