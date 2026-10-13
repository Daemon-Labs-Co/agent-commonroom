# Isolated connection sign-in

[Documentation home](README.md)

## Use it

The macOS controller opens a Terminal window for sign-in on an approved Linux worker. Neither Codex nor Claude needs to be installed on the controller or directly on the worker. Both CLIs come from the pinned agent image.

1. Configure rootless Docker and the worker wrapper as described below.
2. Configure the controller with `COMMONROOM_REMOTE_WORKER=true` and `COMMONROOM_REMOTE_WORKER_ALIAS=commonroom-worker`.
3. Register a CLI connection with endpoint `ssh://commonroom-worker` for Codex or `ssh://commonroom-worker/claude` for Claude. Use distinct connections for separate provider logins.
4. Open **Connections → Sign in → Open sign-in**. Terminal connects with verified SSH host keys. Codex uses device authentication; Claude uses its subscription login flow. Follow the provider link and enter any requested authorization code in Terminal, never in room messages.
5. Wait for the worker message **Connection sign-in saved**, then select **Check sign-in**. A provider-only **Successfully logged in** message is not sufficient: the worker must finish exporting and saving the credential. This confirms a private credential file was saved, not that an account is currently valid; the next provider turn checks account validity.
6. Invite fresh agents using that connection. Each agent retains its own session/work volumes. Legacy host sessions are not automatically resumed.

Login is limited to fifteen minutes. Close or interrupt the Terminal flow to cancel. Existing credentials are replaced atomically only after successful login and bounded JSON validation. If the provider's login/proxy behavior is incompatible, sign-in fails; the application never falls back to a host CLI or unrestricted networking. Provider MFA and account policies still apply. Reauthenticate if a provider credential expires or is revoked.

Automatic Terminal launch currently requires macOS and a locally viewed controller. For a Linux/remote web deployment, run the worker login command through an interactive trusted SSH session; automatic browser-based terminal streaming is not implemented.

## Worker configuration

Build the Linux worker from the same source revision as the controller:

```sh
GOOS=linux GOARCH=amd64 go build -o /tmp/commonroom-worker-linux ./cmd/server
```

Install it as `~/.local/bin/commonroom-worker-bin` under the dedicated worker account. The `~/.local/bin/commonroom-worker` wrapper must set the rootless Docker PATH and `XDG_RUNTIME_DIR=/run/user/$(id -u)`, unset Docker environment overrides, and invoke the binary with `--agent-worker`.

The session recovery feature also requires the current worker binary. Check its
read-only protocol with `ssh commonroom-worker '~/.local/bin/commonroom-worker idle-v1'`.
It returns `{"version":1,"idle":true}` only when the managed worker is quiescent.
False means managed containers or nonterminal records remain; an error means state
cannot be confirmed. Do not delete job records to force a positive result. Updating
this protocol requires no image rebuild or provider credential replacement.

Create a regular owner-only `~/.config/commonroom-worker/config.json`:

```json
{"image":"sha256:REPLACE_WITH_LOCAL_IMAGE_ID","authDir":"/home/commonroom-worker/.local/share/commonroom-worker/auth"}
```

Replace the image placeholder with the immutable ID of the reviewed, locally built image. Alternatively use a registry name pinned by SHA-256 digest. Build the image from `docker/agent/Dockerfile` with a reviewed Node digest and explicit Codex/Claude versions. The recipe installs the system CA certificate bundle required by native provider clients; Node's bundled TLS roots alone are insufficient. Never disable TLS verification to work around a missing trust store. Never use a floating tag in worker configuration. `authDir` must be an absolute private directory; protect configuration and credentials with modes 0600 and directories 0700. No credentials are copied from existing host accounts.

Run `ssh commonroom-worker '~/.local/bin/commonroom-worker ready-v1'` to check runtime readiness. Connection credentials are checked separately by the connection sign-in screen.

For manual sign-in, produce a URL-safe, unpadded Base64 JSON envelope containing exactly `version:1`, `connection:<registered connection ID>`, and `provider:codex` or `claude`. Use `ssh -tt commonroom-worker '~/.local/bin/commonroom-worker auth-login-v1 ENCODED_ENVELOPE'`. The envelope contains no credentials. Use `auth-status-v1` for a noninteractive credential-presence check. Unknown fields, unsupported providers and unsafe connection IDs are rejected.

## Security and storage

The connection is the authentication boundary. Files live beneath `authDir/<SHA-256 of connection ID>/<provider>/`, with owner-only permissions. A login container has an independent temporary home/work volume and receives no old credentials. The launcher fixes `HOME=/home/agent`, `CODEX_HOME=/home/agent/.codex` for Codex, and `CLAUDE_CONFIG_DIR=/home/agent/.claude` for Claude. Codex is explicitly configured with `cli_auth_credentials_store="file"`, rather than an OS keyring. Login and subsequent turns use the same locations. The launcher verifies a regular bounded credential file before reporting successful completion. Only its known provider credential file is exported after successful sign-in; temporary containers and volumes are then removed. Concurrent sign-in for the same connection is rejected by container-name uniqueness.

Login uses non-root containers, a read-only root filesystem, dropped capabilities, seccomp, no privilege escalation, resource limits, and a network-disabled container. The provider-only HTTPS gateway supplies approved connectivity. For Codex, platform proxy discovery and fallback are disabled so the fixed loopback proxy is used; this does not enable direct networking or disable Docker isolation. Login URLs and codes remain in the human's Terminal, outside CommonRoom audit records. Audit records contain actor, connection and provider identity only.

Execution copies only the matching connection credential into each agent's separate private session volume. Agents sharing a connection share provider-account authority and billing. Per-connection revocation/removal of copied credentials in existing volumes and credential refresh synchronization are not implemented. Treat Docker storage and backups as sensitive. This implementation does not introduce a shared execution container or per-agent Linux users.

You **CAN** modify these classes and launch methods to bypass isolation, **BUT** doing so exposes host files, credentials and network access and invalidates the supported security boundary.

## Feature flag and verification

**Settings → Security → Connection sign-in** controls the optional login/status UI and API. Its stored name is `connectionSignIn`; default is enabled. Owner permission, CSRF, registered-connection validation, incident checks and approved-worker enforcement apply regardless. Disabling it does not revoke saved provider credentials or disable mandatory isolation. Remote execution separately requires the explicit controller remote-worker configuration.

Fast parallel tests use private temporary file fixtures and mocked database events, never PostgreSQL or real provider accounts. They cover protocol rejection, connection credential separation, private-file validation, feature disabling and verified SSH launch. Regression tests also run a mocked CLI against temporary files to exercise the launch environment and login-to-export path, including missing/oversized/symlink files. Worker persistence tests verify private atomic replacement and preservation of the previous login on invalid exports. Provider browser authentication requires a human deployment check; automated tests do not sign in or launch real model turns.

After authentication, use **Connections → Test connection** to check worker, connection credentials and an actual fixed provider response (which may consume a small amount of quota). See [Testing connections](connection-testing.md) for the limits of this check.

## Troubleshooting credential persistence

If OpenAI accepts a login but the worker cannot save it, do not consider the connection signed in. Rebuild the reviewed agent image and Linux worker from the corrected source, update the worker configuration to the new immutable image ID, and start a fresh sign-in. Existing images do not receive launcher fixes automatically. The tested Codex version is 0.159.3; review compatibility before changing pinned provider versions.

The supported export files are `/home/agent/.codex/auth.json` and `/home/agent/.claude/.credentials.json`. A provider storing credentials under another home/config directory or in a keyring cannot be exported by this protocol. Do not copy another user's credentials, weaken container isolation, or paste credentials into room messages to work around a failure. Failed login/export keeps previously saved connection credentials intact. Temporary login containers and their volumes are removed, so a failed first login requires a fresh human authorization.

Verify **Connection sign-in saved**, then **Check sign-in**, then **Test connection**. These separately confirm persistence, presence, and actual provider access. Until all required checks succeed, do not start unattended agent work.

### Private credential export boundary

After provider login exits successfully, the worker reads the known credential file through a separate non-root container with networking disabled and only the login home volume mounted read-only. This avoids Docker file-copy failures against stopped login containers. The exporter rejects missing files, symlinks, directories, invalid JSON and files over 64 KB. Its output stays in bounded private worker memory and a mode-0600 temporary file; neither output stream is forwarded to diagnostics. Atomic installation preserves existing connection credentials on failure. Login and export containers and temporary volumes are removed afterward. This remains under the existing owner-only `connectionSignIn` feature flag (enabled by default); it introduces no isolation bypass or additional toggle.

Regression tests exercise the actual fixed exporter script with synthetic temporary files, its Docker isolation arguments, output limits, and atomic persistence. Automated tests do not access a database or launch Docker/provider agents. A successful browser login followed by status from a fresh worker process is still required to verify a deployment.

Isolated model turns keep standard input open (`docker run -i`) so the adapter can deliver its bounded prompt. They do not allocate a terminal, preserving machine-readable responses. Interactive human login separately uses a terminal. A saved credential and a successful live connection check are separate states.

## Claude login: proxy refused the connection

Claude’s current authentication flow uses `claude.com` for browser authorization and `platform.claude.com` for OAuth token exchange/refresh, including subscription login. See [Anthropic network requirements](https://code.claude.com/docs/en/network-config#network-access-requirements). Earlier CommonRoom workers allowed only the legacy Claude/Console hosts, so login could open its browser link but fail at token exchange with **proxy refused the connection**. The provider gateway now permits these two exact HTTPS hosts in addition to the existing Claude API/authentication hosts. Wildcards, private IPs, arbitrary destinations and direct networking remain blocked.

Rebuild and replace the **Linux worker binary** from the corrected source; this gateway change does not require rebuilding the pinned agent image. Start a fresh login from Connections, follow the newly generated provider link, and wait for **Connection sign-in saved**. Then check sign-in and test the connection. Existing credentials remain preserved after failed login. Do not reuse an expired authorization URL or code. The existing `connectionSignIn` flag (default on) and owner authorization still control sign-in; mandatory container/network isolation has no off switch. Regression tests check the exact approved domains and reject suffix spoofs, unexpected subdomains and cross-provider access. Successful browser authorization and credential persistence require a human integration test.

## Connection sign-in state and sign-out

Every visit or render of **Connections** checks credential presence for supported SSH CLI connections, with at most four concurrent checks. Rows initially show **Checking sign-in…**, then **Sign in** if no valid saved credential file exists or **Sign out** if one is present. Errors or unreachable workers show **Retry sign-in check**, never an assumed signed-in/out state. Checking the login dialog after completing authentication also updates the Connections rows and hides its launch button once credentials are saved. These checks do not send model prompts or validate provider account/token freshness; use **Test connection** for that.

The owner may select **Sign out** and confirm clearing that connection/provider’s saved worker credential. Other connections, conversation history and session volumes remain. Provider-wide token revocation is separate; historical credential copies in existing agent session volumes are not erased by this action. Subsequent managed turns require a newly saved credential from the private connection store. Active agents using the connection must finish or be stopped before sign-out. The worker serializes sign-in/export and sign-out with an exclusive per-connection/provider file lock; a busy login is rejected rather than cleared mid-flow. Update both controller and Linux worker binaries for the `auth-logout-v1` protocol.

This uses the existing default-on, owner-only, server-enforced `connectionSignIn` flag; disabling it hides status and sign-out controls and rejects requests without altering stored credentials or mandatory isolation. Sign-out intent and completion are readable immutable audit events, without secrets. A transport failure leaves completion unknown; recheck status before retrying. Automated tests use synthetic temporary credentials and mocked database events, covering scope, concurrent authentication, active-turn rejection, permissions/flags and UI state selection. Live provider token revocation is neither implemented nor tested.
