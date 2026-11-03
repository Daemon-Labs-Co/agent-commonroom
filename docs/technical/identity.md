# Identity, settings and access

## Input

Authenticated human session, CSRF token for mutations, room/account scope and owner-managed feature settings. Agent credentials cannot act as human API sessions.

Provider sign-in accepts owner-authenticated, CSRF-protected
`POST /api/provider-auth` with `action=login` and a registered CLI connection ID.
The existing `connectionSignIn` flag defaults enabled; disabling it rejects new
requests without deleting credentials. Active/unresolved connection work,
integrity incidents and emergency stops block login.

## Output

Authorized state views, accepted mutations or explicit authentication/permission failures.

Codex login returns no-store NDJSON `starting`, `instructions`, `waiting`,
`complete` or `error` events. Instructions contain only the fixed
`https://auth.openai.com/codex/device` URL and a validated one-time device code.
Raw worker stdout/stderr, passwords and tokens are never returned or audited.
Claude login returns an `authorization` event containing an allowlisted HTTPS
`claude.com/cai/oauth/authorize` URL (or a supported legacy
`/oauth/authorize` endpoint) and a random pending-login ID. The owner submits
`action=login-code`, the same connection ID, login ID and the provider’s code.
Codes contain 8–2048 permitted characters, are delivered once to the pending
CLI’s stdin, and are never echoed, audited or saved as configuration.

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

Codex uses a noninteractive container without Docker TTY/stdin attachment;
Claude retains interactive stdin/TTY inside the isolated container; an anonymous
OS pipe forwards the owner’s submitted code to the unmodified CLI. Pending login
state is transient, bound to connection/worker/provider, and removed on completion,
failure or cancellation. Duplicate starts, expired IDs, cross-connection codes,
replay, non-owner requests, missing CSRF and disabled sign-in are rejected.
No second code submission or sign-out can race an active login. Verification instructions
must arrive within 45 seconds or the request is cancelled with an explicit error.
Codex authentication starts the constrained isolated worker's existing
`auth-login-v1` device-auth path rather than a platform Terminal application.
Only a successful worker exit reports completion, after credential persistence.
Output is bounded to 16 KiB; the response is bounded to 15 minutes and sends
15-second heartbeats. Its write deadline is 16 minutes; normal HTTP deadlines
remain unchanged. Disconnect, dialog close, cancellation and timeout terminate
the worker request, with SIGTERM and a 30-second process wait allowance for
container/volume cleanup. Unreachable remote workers can leave cleanup unconfirmed.
The owner then checks saved credentials separately from actual provider access.
Successful dialogs hide Start/Check sign-in, show Done, and ignore stale status
responses that would otherwise overwrite completion. Claude password entry and
approval remain on the provider website; no custom token exchange is implemented.

## Properties and state changes

Session CSRF/expiry and feature settings may change. Immutable audit records remain retained. Isolation/authentication/retention cannot be disabled by product flags.

Source: [auth.go](../../cmd/server/auth.go) · [event_mutations_identity.go](../../cmd/server/event_mutations_identity.go) · [http_state.go](../../cmd/server/http_state.go).

### Provider terminal output filtering

- Input: bounded isolated CLI output, including ANSI styling and OSC terminal
  hyperlinks terminated by BEL or ST.
- Output: validated visible provider authorization instructions only.
- Functionality: removes terminal hyperlink controls and hidden targets before
  extracting URLs; handles controls split across output chunks.
- Properties and state changes: existing output bounds, startup deadline, URL
  validation and credential privacy requirements remain enforced.
