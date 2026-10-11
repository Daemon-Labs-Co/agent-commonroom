# Isolated connection sign-in

[Documentation home](README.md)

## Use it

The macOS controller opens a Terminal window for sign-in on an approved Linux worker. Neither Codex nor Claude needs to be installed on the controller or directly on the worker. Both CLIs come from the pinned agent image.

1. Configure rootless Docker and the worker wrapper as described below.
2. Configure the controller with `COMMONROOM_REMOTE_WORKER=true` and `COMMONROOM_REMOTE_WORKER_ALIAS=commonroom-worker`.
3. Register a CLI connection with endpoint `ssh://commonroom-worker` for Codex or `ssh://commonroom-worker/claude` for Claude. Use distinct connections for separate provider logins.
4. Open **Connections → Sign in → Open sign-in**. Terminal connects with verified SSH host keys. Codex uses device authentication; Claude uses its subscription login flow. Follow the provider link and enter any requested authorization code in Terminal, never in room messages.
5. When Terminal reports success, select **Check sign-in**. This confirms a private credential file was saved, not that an account is currently valid; the next provider turn checks account validity.
6. Invite fresh agents using that connection. Each agent retains its own session/work volumes. Legacy host sessions are not automatically resumed.

Login is limited to fifteen minutes. Close or interrupt the Terminal flow to cancel. Existing credentials are replaced atomically only after successful login and bounded JSON validation. If the provider's login/proxy behavior is incompatible, sign-in fails; the application never falls back to a host CLI or unrestricted networking. Provider MFA and account policies still apply. Reauthenticate if a provider credential expires or is revoked.

Automatic Terminal launch currently requires macOS and a locally viewed controller. For a Linux/remote web deployment, run the worker login command through an interactive trusted SSH session; automatic browser-based terminal streaming is not implemented.

## Worker configuration

Build the Linux worker from the same source revision as the controller:

```sh
GOOS=linux GOARCH=amd64 go build -o /tmp/commonroom-worker-linux ./cmd/server
```

Install it as `~/.local/bin/commonroom-worker-bin` under the dedicated worker account. The `~/.local/bin/commonroom-worker` wrapper must set the rootless Docker PATH and `XDG_RUNTIME_DIR=/run/user/$(id -u)`, unset Docker environment overrides, and invoke the binary with `--agent-worker`.

Create a regular owner-only `~/.config/commonroom-worker/config.json`:

```json
{"image":"sha256:REPLACE_WITH_LOCAL_IMAGE_ID","authDir":"/home/commonroom-worker/.local/share/commonroom-worker/auth"}
```

Replace the image placeholder with the immutable ID of the reviewed, locally built image. Alternatively use a registry name pinned by SHA-256 digest. Build the image from `docker/agent/Dockerfile` with a reviewed Node digest and explicit Codex/Claude versions. Never use a floating tag in worker configuration. `authDir` must be an absolute private directory; protect configuration and credentials with modes 0600 and directories 0700. No credentials are copied from existing host accounts.

Run `ssh commonroom-worker '~/.local/bin/commonroom-worker ready-v1'` to check runtime readiness. Connection credentials are checked separately by the connection sign-in screen.

For manual sign-in, produce a URL-safe, unpadded Base64 JSON envelope containing exactly `version:1`, `connection:<registered connection ID>`, and `provider:codex` or `claude`. Use `ssh -tt commonroom-worker '~/.local/bin/commonroom-worker auth-login-v1 ENCODED_ENVELOPE'`. The envelope contains no credentials. Use `auth-status-v1` for a noninteractive credential-presence check. Unknown fields, unsupported providers and unsafe connection IDs are rejected.

## Security and storage

The connection is the authentication boundary. Files live beneath `authDir/<SHA-256 of connection ID>/<provider>/`, with owner-only permissions. A login container has an independent temporary home/work volume and receives no old credentials. Only its known provider credential file is exported after successful sign-in; temporary containers and volumes are then removed. Concurrent sign-in for the same connection is rejected by container-name uniqueness.

Login uses non-root containers, a read-only root filesystem, dropped capabilities, seccomp, no privilege escalation, resource limits, and a network-disabled container. The provider-only HTTPS gateway supplies approved connectivity. Login URLs and codes remain in the human's Terminal, outside CommonRoom audit records. Audit records contain actor, connection and provider identity only.

Execution copies only the matching connection credential into each agent's separate private session volume. Agents sharing a connection share provider-account authority and billing. Per-connection revocation/removal of copied credentials in existing volumes and credential refresh synchronization are not implemented. Treat Docker storage and backups as sensitive. This implementation does not introduce a shared execution container or per-agent Linux users.

You **CAN** modify these classes and launch methods to bypass isolation, **BUT** doing so exposes host files, credentials and network access and invalidates the supported security boundary.

## Feature flag and verification

**Settings → Security → Connection sign-in** controls the optional login/status UI and API. Its stored name is `connectionSignIn`; default is enabled. Owner permission, CSRF, registered-connection validation, incident checks and approved-worker enforcement apply regardless. Disabling it does not revoke saved provider credentials or disable mandatory isolation. Remote execution separately requires the explicit controller remote-worker configuration.

Fast parallel tests use private temporary file fixtures and mocked database events, never PostgreSQL or real provider accounts. They cover protocol rejection, connection credential separation, private-file validation, feature disabling and verified SSH launch. Provider browser authentication requires a human deployment check; automated tests do not sign in or launch real model turns.

After authentication, use **Connections → Test connection** to check worker, connection credentials and an actual fixed provider response (which may consume a small amount of quota). See [Testing connections](connection-testing.md) for the limits of this check.
