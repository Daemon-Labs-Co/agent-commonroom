# Identity, settings and access

## Input

Authenticated human session, CSRF token for mutations, room/account scope and owner-managed feature settings. Agent credentials cannot act as human API sessions.

Provider sign-in accepts owner-authenticated, CSRF-protected
`POST /api/provider-auth` with `action=login` and a registered CLI connection ID.
The existing `connectionSignIn` flag defaults enabled; disabling it rejects new
requests without deleting credentials. Active/unresolved connection work,
integrity incidents and emergency stops block login.

Sign-in methods (ADR 0006), all per workspace and all ending in the same session cookie
as key-file sign-in:

| Request | Auth | Effect |
| --- | --- | --- |
| `GET /api/auth/methods` | none | `passkeys`, `oidc`, `oidcConfigured`, `oidcName`, `invitations` (configuration only). |
| `POST /api/webauthn/login/options` / `login/verify` | none | Passkey sign-in with a discoverable credential. |
| `POST /api/webauthn/register/options` / `register/verify`, `POST /api/webauthn/remove` | session, CSRF | Add or remove your own passkey (at most 10; never your last sign-in method). |
| `POST /api/invitations` (create), `/api/invitations/revoke` | owner, CSRF | Single-use link, 1-168 hours (default 72), role Developer or QA, shared rooms only, at most 50 open. |
| `POST /api/invitations/lookup`, `/passkey/options`, `/passkey/complete`, `GET /api/invitations/oidc/start` | none | The invitee names themselves and registers a passkey or signs in with single sign-on. |
| `GET /api/oidc/login`, `GET /api/oidc/link` (session), `GET /api/oidc/callback` | none / session | Single sign-on sign-in and linking; the callback alone accepts a cross-site navigation. |

Flags `passkeySignIn`, `oidcSignIn` and `memberInvitations` default on for hosted workspaces
and off otherwise; an explicit owner setting always wins. Operator settings for single
sign-on: `COMMONROOM_OIDC_ISSUER` (https, DNS host), `COMMONROOM_OIDC_CLIENT_ID`,
`COMMONROOM_OIDC_CLIENT_SECRET` or `COMMONROOM_OIDC_CLIENT_SECRET_FILE` (private file) and
optional `COMMONROOM_OIDC_NAME`; the provider must allow the redirect
`<workspace origin>/api/oidc/callback` for every workspace host.

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

GET /api/setup-status exposes only no-store `setupRequired` and `setupCodeExpired`
booleans, under normal Host/Origin checks. It grants no session or setup authority.
First-owner setup requires the private bootstrap code, followed by private-key
download and proof of possession. Returning-user PEM failures do not initiate
registration.

The bootstrap code is the generated private file named by `OWNER_KEY_FILE`
(default `.data/owner-key`, no expiry) unless the operator sets
`COMMONROOM_SETUP_CODE` to exactly 64 lowercase hexadecimal characters. A supplied
code is never written to a file, expires `COMMONROOM_SETUP_CODE_TTL_HOURS` hours
after process start (whole hours, 1 to 168, default 24) and is refused at sign-in
once expired. Invalid values stop startup with a message that does not echo them.
Either code is cleared at startup once the owner private key has been claimed, so
it cannot act as a standing owner password. `commonroom --setup-link` takes no arguments and prints
`<origin>/#setup=<code>` using `COMMONROOM_PUBLIC_ORIGIN` or `LISTEN_ADDR`; it
refuses when no setup is pending. It takes `OWNER_KEY_FILE`, `COMMONROOM_PUBLIC_ORIGIN`,
`LISTEN_ADDR` and `COMMONROOM_SETUP_CODE` from the process environment, else from
the service settings file (`COMMONROOM_ENV_FILE`, default
`~/.config/commonroom/runtime.env`), which must be private, owned by the caller and
at most 64 KiB; no other setting in that file is read or returned. The browser reads the code from the URL fragment,
fills the field without submitting, and removes the fragment from the address bar;
fragments are never sent to the server or proxies. Both variables are operator
process configuration read at startup; unset keeps the previous behavior. They do
not bypass authentication, the empty-history check or audit integrity.

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

Passkey and single sign-on verification uses only the standard library (`internal/webauthn`,
`internal/oidc`). Passkeys: `none` attestation only, ES256 or RS256 (RSA 2048-4096 bits),
user presence and verification required, exact origin and relying-party id (the workspace
host), single-use challenges bound to one purpose that expire after 5 minutes, a signature
counter that may not go backwards, and cross-origin or embedded use refused. CBOR input is
bounded (definite lengths, depth 8, 64 items, no tags or floats). Single sign-on: authorization
code with PKCE and nonce; ID tokens must be RS256 or ES256 with matching issuer, audience,
authorized party, expiry (60 seconds of skew), not-before and nonce; unknown keys trigger at most
one key-set refresh a minute; redirects from the provider are never followed; identities link by
issuer and subject, never email, and one identity belongs to one account. Each flow is bound to
its browser by a Lax cookie scoped to `/api/oidc`, and its state is single use and expires after
10 minutes. Invitation secrets are 256-bit random values stored only as a SHA-256 hash; every
refusal is the same 404; failures share the key-file sign-in budget (5 a minute); redemption
creates the member, records the credential or link and spends the invitation in one atomic batch.
A new session replaces any session the browser presented. Unused invitations count towards the
workspace member limit.

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
