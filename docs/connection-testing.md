# Test an agent connection

[Documentation index](README.md) · [Sign in](connection-sign-in.md)

**Test connection** checks an approved isolated worker and, for an unchanged saved
connection, asks the provider for one fixed reply. A valid credential file or reachable
SSH host alone cannot produce **Connection working**.

## Run a test

1. Open **Connections** and select **Test connection** beside a saved connection.
   The owner can also test the current fields in **New connection** or **Edit connection**.
2. Read the result and suggested next action. Stages show running, passed, required,
   blocked or failed with text as well as color.
3. Open **View log** for sanitized check details. Close it to return to the result,
   or choose **Back to editing** to preserve the entered fields.
4. After a fix, choose **Test again**. Testing never saves or registers a connection.

The editor keeps its header/actions visible and puts longer details in bounded
scrolling overlays. Closing a running test aborts its request; uncertain execution
still requires independent worker cleanup confirmation.

## What each stage proves

| Stage | Checks |
| --- | --- |
| Approved worker | Endpoint matches the controller allowlist and remote transport is enabled. |
| Worker runtime | Verified SSH reaches matching private worker configuration, rootless Docker, limits and pinned image. |
| Saved credentials | This connection/provider has a supported private JSON credential file. |
| Provider reply | A fresh isolated turn returns `COMMONROOM_CONNECTION_OK` to a fixed prompt with tools disabled. |

No room history, repository, profile context or memory is sent. The provider reply
may consume account quota. Success requires the expected response **and durable
confirmed shutdown**; it is a point-in-time check, not a monitoring guarantee or
development-readiness certificate.

## Interpret results

| Result | Next action |
| --- | --- |
| Connection working | Invite a fresh participant; development has separate admission requirements. |
| Sign-in needed | Complete [isolated sign-in](connection-sign-in.md), then test again. |
| Connection blocked | Correct the named worker/transport/configuration stage; there is no host fallback. |
| Connection test failed | Check provider account access, expiry and the reviewed CLI/image compatibility. |
| Adapter unavailable | HTTP, WebSocket and MCP execution adapters are not implemented; no request is made to their URL. |
| Save and sign in to finish testing | Draft setup passed; save it and establish connection-scoped credentials. |

Local CLI records require an approved Linux worker endpoint; this diagnostic never
launches a host CLI. Other provider icons do not imply executable adapters.

## Testing before saving

New drafts and changed execution endpoints/methods receive setup-only checks. They
cannot read saved credentials or launch provider turns. An unchanged saved connection
can receive full verification. Editing only its name/icon does not change credential
authority. Field changes invalidate old results, including while a request is running.

The **Setup guide** opens separately and preserves the draft. Returning to the form,
closing help or viewing diagnostics does not save changes or start sign-in.

## Feature flag and permission

`connectionTesting` is enabled by default and owner-controlled in Settings. Off
hides controls and rejects the API. Owner authentication, CSRF, incident/stop checks,
registered identity and worker approval remain server-enforced. The optional
`COMMONROOM_CONNECTION_EDITOR` presentation flag defaults on and changes layout
only; it cannot change test authority or isolation.

## API and implementation

`POST /api/connections/test` accepts a registered `connectionId` and optionally a
`draft` with `name`, `protocol`, `endpoint`, `provider` and `location`. Strict JSON
is limited to 4 KiB; unknown fields and trailing objects are rejected. Missing
records return 404, flag/permission denial 403 and incident/stop conflict 409.

Results contain `status`, `checks`, `checkedAt` and `providerResponseTested`. The
last field is true only for a valid correlated provider outcome. Expected diagnostic
failures return safe status/check detail; raw provider errors are withheld.

The request uses a forty-second diagnostic budget, including bounded SSH probes.
Provider execution has a maximum twenty-second deadline, followed by bounded
cleanup that can outlast the turn. Nothing is automatically replayed after an
uncertain partial execution.

## Live test progress

For `Accept: application/x-ndjson`, the server emits sanitized `stage`, `result`
and `error` messages. Progress follows actual stage boundaries. A terminal result
requires audit recording; interrupted streams cannot claim success. Responses set
`Cache-Control: no-store` and `X-Accel-Buffering: no`; a reverse proxy must allow
streaming. JSON clients remain supported.

`connection_diagnostics.go` owns admission and worker probes;
`web/connection-testing.js` owns presentation. `connection.tested` records actor,
connection and safe overall status without credentials or provider output.

## Controller-owned verification lifecycle

Readiness and credential-presence probes use `ready-v1` and `auth-status-v1`.
Provider verification uses the journalled `run-v2` lifecycle. The old `verify-v1`
launch command is rejected; its name survives only as an internal stage label.
Upgrade controller and worker together.

The controller records a private synthetic verification participant/job before
execution and reserves capacity. It is neither a human identity nor a real squad
member. Connection edits, deletion and sign-out are blocked while verification is
active. Its session is not attached to a room participant. Verification volume
storage is retained; automatic disposable-volume cleanup is not claimed.

Browser disconnect/emergency stop cancels execution and initiates fencing. Unknown
shutdown latches an application stop and remains in job diagnostics for owner
reconciliation. Restart checks include verification jobs. Hiding `jobDiagnostics`
does not disable its journal or cleanup requirements.

### Chat command-tool restriction

Codex probes require the fixed tool-disabled chat prefix, including
`features.shell_tool=false`; read-only sandbox alone is insufficient. Claude uses
an empty tool allowlist. The worker rejects altered policies with no fallback.
See [Codex chat policy](security-hardening.md#codex-chat-command-tool-policy).

## Verification limits

Tests mock database events, readiness/auth probes, provider outcomes and lifecycle
proofs. They cover scope, draft boundaries, permissions/flags, streaming, failure,
cancellation and unknown shutdown without real provider, SSH or database access.
UI tests check draft preservation and escaping. Real deployment timing, account
access and controller/worker compatibility need separate live acceptance.
