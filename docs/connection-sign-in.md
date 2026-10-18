# Isolated connection sign-in

[Documentation index](README.md) · [SSH setup](agent-ssh.md) · [Connection testing](connection-testing.md)

Sign-in stores provider credentials on the approved Linux worker, scoped to one
connection and provider. The controller and worker host do not need native provider
CLIs: the reviewed chat/login image supplies them. Provider authentication is separate
from the CommonRoom owner's RSA sign-in.

## Use it

1. Complete the worker configuration below and approve its SSH alias on the controller.
2. Register `ssh://commonroom-worker` for Codex or
   `ssh://commonroom-worker/claude` for Claude.
3. On a locally viewed macOS controller, choose **Connections → Sign in → Open sign-in**.
   Terminal opens a verified SSH connection. Follow the provider's authorization
   link and enter requested codes in Terminal, never in a room message.
4. Wait for **Connection sign-in saved**. A provider's own success message does not
   prove that the worker exported and persisted its credential.
5. Select **Check sign-in**, then **Test connection**. These check saved-file presence
   and a real fixed provider reply respectively; the latter may consume quota.

Login is bounded to fifteen minutes. Interrupt the Terminal flow to cancel. Failed
validation preserves the previous credential; a storage error after publication can
be ambiguous, as explained below. There is no host CLI or unrestricted-network fallback.

Automatic Terminal launch requires macOS and a locally viewed controller. On a
Linux/remote web deployment, use the manual trusted SSH operation below. Browser
terminal streaming is not implemented. Development execution remains separately
[blocked pending acceptance](development-workspaces.md).

## Worker configuration

### Install a matching binary

Build the worker from the same reviewed revision as the controller:

```sh
GOOS=linux GOARCH=amd64 go build -o /tmp/commonroom-worker-linux ./cmd/server
```

Choose the architecture matching the worker. Install as
`~/.local/bin/commonroom-worker-bin` under its dedicated account. The wrapper at
`~/.local/bin/commonroom-worker` must set the rootless Docker PATH and
`XDG_RUNTIME_DIR=/run/user/$(id -u)`, unset Docker environment overrides and invoke
the binary with `--agent-worker`. The wrapper is trusted administrator configuration,
not a model-selected executable.

### Configure image and private storage

Create an owner-only regular `~/.config/commonroom-worker/config.json`:

```json
{"image":"sha256:REVIEWED_LOCAL_IMAGE_ID","authDir":"/home/commonroom-worker/.local/share/commonroom-worker/auth"}
```

Replace the image and path placeholders. Use a reviewed immutable image built from
`docker/agent/Dockerfile`, with digest-pinned base and explicit provider versions.
See [image prerequisites](container-isolation.md#set-up-a-linux-worker). Never
substitute the separate development-engine image or a floating tag.

Keep directories mode 0700 and configuration/credentials mode 0600, owned by the
worker account. `authDir` must be absolute. Do not copy another account's provider
home or host credentials. The image includes system CA certificates; never disable
TLS verification to work around trust-store failures.

### Check readiness

```sh
ssh commonroom-worker '~/.local/bin/commonroom-worker ready-v1'
ssh commonroom-worker '~/.local/bin/commonroom-worker idle-v1'
```

Readiness checks runtime configuration; idle returns `{"version":1,"idle":true}`
only when managed state is quiescent. False or an error is not proof of shutdown.
Do not delete records to force success. Credential presence is checked separately.

### Manual login protocol

Under an interactive trusted SSH session, invoke `auth-login-v1` with a URL-safe,
unpadded Base64 JSON envelope containing exactly `version:1`,
`connection:<registered ID>` and `provider:codex` or `claude`:

```text
ssh -tt commonroom-worker '~/.local/bin/commonroom-worker auth-login-v1 ENCODED_ENVELOPE'
```

The envelope contains no secrets. Use the exact registered connection ID;
`auth-status-v1` uses the same scope for a noninteractive presence check. Unknown
fields, unsafe IDs and unsupported providers are rejected. Do not paste raw
credential JSON into command lines or issue reports.

## Security and storage

Credentials live below `authDir/<SHA-256 of connection ID>/<provider>/`. Connections
sharing one provider account share its authority and billing, while keeping separate
stored files. Per-connection locks serialize login/export, sign-out and engine use.
Busy or unresolved work blocks reuse rather than silently sharing credentials.
A durable recovery marker also blocks credentials whose token custody is uncertain.
See [credential recovery](worker-credential-recovery.md) for the repair sequence;
container cleanup alone does not make a potentially stale token safe.

Login uses a fresh temporary home/work volume, non-root execution, read-only root,
dropped capabilities, resource limits and network-disabled containers. The fixed
provider HTTPS gateway supplies approved connectivity. Authorization URLs/codes stay
in the human Terminal; audit records contain safe actor/connection/provider metadata.

### Private credential export boundary

The image fixes Codex storage at `/home/agent/.codex/auth.json` with file-based
credential storage and Claude at `/home/agent/.claude/.credentials.json`. After login,
a separate non-root, network-disabled exporter mounts only the login volume read-only
and reads the known file. It rejects missing, symlinked, nonregular, oversized or
invalid JSON files. Export is bounded to 64 KiB and never forwarded as diagnostics.
Temporary login/export containers and volumes are then removed.

Chat turns copy the matching credential into their private session storage. The
internal separated Codex engine receives a bounded private stdin frame instead of
a host credential mount. The internal separated engine now exports final credentials privately and
persists them after matching durable pair fencing. Initial sign-in still does not
establish live refresh compatibility. The worker marks custody uncertain before
sending engine credentials; invalid/unavailable exports retain that marker and
require explicit sign-in recovery. See the [custody protocol](engine-credential-custody.md).

You **CAN** alter source launchers to bypass these rules, **BUT** doing so exposes
host files, credentials and network authority and invalidates mandatory isolation.

## Feature flag and verification

**Settings → Security → Connection sign-in** controls `connectionSignIn`, enabled
by default. The server enforces owner permission, CSRF, registered connection,
incident and approved-worker checks. Off rejects login/status/sign-out requests;
it neither revokes stored credentials nor weakens mandatory isolation.

Tests use synthetic temporary credentials, mocked processes and database events.
They cover both flag states, scope, private-file rules, export, persistence and
verified SSH arguments. They do not authenticate a real provider. Complete browser
login and a fresh-process status/provider check when accepting a deployment.

## Troubleshooting credential persistence

Follow this order: **Connection sign-in saved → Check sign-in → Test connection**.
If export fails, verify the reviewed image's exact home/config paths and matching
worker version. Keyring-only or changed provider storage is not supported by this
file export protocol. Rebuild, review and repin an incompatible image; existing
images do not acquire source fixes automatically.

Do not weaken permissions, copy another user's credential or reuse expired login
codes. A failed first login requires fresh human authorization. Preserve raw
credentials outside logs and room messages.

## Claude login: proxy refused the connection

The gateway permits `claude.com` and `platform.claude.com` alongside the exact
Claude API/auth hosts listed in [network policy](container-isolation.md#provider-only-network-gateway).
An older worker missing those hosts can fail during token exchange even if the
browser opens. Update the reviewed worker binary, start a fresh login and wait for
worker persistence confirmation. Unknown destinations still fail closed. Successful
browser authorization and persistence require live human verification.

## Connection sign-in state and sign-out

Connections checks at most four credential-presence queries concurrently. It shows
**Checking sign-in…**, then **Sign in** or **Sign out**. An unreachable worker shows
**Retry sign-in check**, never an assumed authenticated state. Presence does not
validate token freshness; use **Test connection**.

The owner can confirm **Sign out** to remove the selected connection/provider's
saved file. Active work must finish or be stopped and reconciled first. This does
not revoke the provider token, erase old session-volume copies or remove history.
After transport failure, completion is unknown: check status before retrying.
Sign-out intent/completion is audited without credential contents.

## Durable local credential publication and sign-out

Publication writes and syncs a complete private temporary file, atomically renames
it, then syncs the parent directory before acknowledging success. Invalid exports
preserve the prior login. A post-rename sync failure may leave the replacement
visible with unconfirmed durability; an error does not prove nothing changed.

Sign-out syncs the directory after unlinking. Retrying an already-absent file still
syncs it before acknowledgment. Repair storage before retrying. These mandatory
controls have no off switch and do not promise physical erasure, provider-wide
revocation or hardware power-loss guarantees. Tests inject sync failures; live
filesystem/hardware durability remains deployment-specific.
