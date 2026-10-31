# Identity, settings and access

## Input

Authenticated human session, CSRF token for mutations, room/account scope and owner-managed feature settings. Agent credentials cannot act as human API sessions.

## Output

Authorized state views, accepted mutations or explicit authentication/permission failures.

## Functionality

Enforce actor/room checks on the server. Feature flags retain their individual defaults and disabled-state behavior. Provider credentials remain outside model tools. Paid organization/multi-user/billing placement does not make security checks optional.

## Properties and state changes

Session CSRF/expiry and feature settings may change. Immutable audit records remain retained. Isolation/authentication/retention cannot be disabled by product flags.

Source: [auth.go](../../cmd/server/auth.go) · [event_mutations_identity.go](../../cmd/server/event_mutations_identity.go) · [http_state.go](../../cmd/server/http_state.go).
