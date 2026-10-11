# Connect Codex or Claude agents over SSH

> Managed agent execution now requires the [Linux/rootless-Docker worker](container-isolation.md). Direct SSH model dispatch is blocked until a remote isolated worker is implemented. Earlier SSH endpoint setup in this guide does not bypass that restriction.


[Documentation index](README.md) · [Connection concepts](concepts.md#connections-versus-participants)

An endpoint such as `ssh://debian-codex` is not a full remote executable URL. The part after `ssh://` names a host entry in the SSH configuration of the account running Agent CommonRoom. That entry supplies the real hostname, remote username, SSH port, and key.

## 1. Decide which account makes the connection

- Running CommonRoom from your laptop terminal: SSH uses your laptop account's configuration and credentials.
- Running CommonRoom as the Debian service: SSH uses `/var/lib/commonroom/.ssh/config` and credentials belonging to the `commonroom` service account.

The remote account can be different for each agent. Authenticate Codex/Claude under the remote account that will actually run it; a provider login under another user is not automatically shared.

## 2. Configure host aliases

Put these example entries in `~/.ssh/config` **for the account running CommonRoom**:

```sshconfig
Host debian-codex
    HostName agent-server.example.com
    User codex-agent
    Port 22
    IdentityFile ~/.ssh/id_ed25519_commonroom
    IdentitiesOnly yes
    BatchMode yes
    StrictHostKeyChecking yes

Host debian-claude
    HostName agent-server.example.com
    User claude-agent
    Port 22
    IdentityFile ~/.ssh/id_ed25519_commonroom
    IdentitiesOnly yes
    BatchMode yes
    StrictHostKeyChecking yes
```

Replace `agent-server.example.com` with your server's hostname/IP, `codex-agent` and `claude-agent` with its real Linux login names, and Port with its SSH port. The alias names are arbitrary; they do not have to match usernames or DNS names. You may use one account for both providers by setting the same User, but that shares its filesystem/provider credentials.

The sample domain and usernames are placeholders, not working accounts. If your existing alias is `jerry`, use `ssh://jerry` for Codex. If it is `claude`, use `ssh://claude/claude` for Claude. Those aliases must exist in the runtime account's SSH configuration.

For the dedicated Debian service, prepare the private SSH directory and edit its configuration administratively:

```sh
sudo install -d -o commonroom -g commonroom -m 0700 /var/lib/commonroom/.ssh
sudo -H -u commonroom touch /var/lib/commonroom/.ssh/config
sudo chmod 0600 /var/lib/commonroom/.ssh/config
sudoedit /var/lib/commonroom/.ssh/config
```

If you already have a config, add entries without replacing existing ones. Keep config/key ownership with the runtime account.

## 3. Provision a dedicated SSH key

Use an existing dedicated key if you already have one. Otherwise, for the Debian service account, create a new key at the configured path—only if that file does not already exist:

```sh
sudo -H -u commonroom ssh-keygen -t ed25519 -f /var/lib/commonroom/.ssh/id_ed25519_commonroom -C commonroom-agent-access
```

Choose how to handle the key's passphrase. CommonRoom runs SSH with BatchMode and cannot answer password or passphrase prompts. A passphrase-protected key needs an explicitly managed SSH agent made available to the service; the current service guide does not configure one. If using an unattended unencrypted key, protect it with owner-only permissions and limit the remote account's access. Never overwrite an existing key or copy your CommonRoom owner RSA key into SSH.

Install **only the public key** in the intended remote user's authorized_keys through your normal administrative process. On Debian with ssh-copy-id available, an example interactive setup command is:

```sh
sudo -H -u commonroom ssh-copy-id -i /var/lib/commonroom/.ssh/id_ed25519_commonroom.pub codex-agent@agent-server.example.com
```

Supply the real username/hostname, and use `-p` if its SSH port differs from 22. Repeat for a separate Claude account when needed. Key installation grants access to that account; keep the remote account scoped to intended work. Private keys remain on the CommonRoom host, outside Git.

## 4. Verify the server's host key

Confirm the server SSH host fingerprint through a trusted administrator or server console. Establish the known_hosts entry interactively before using the application. For a configured alias with StrictHostKeyChecking=yes, a one-time interactive command can allow a verified first connection:

```sh
sudo -H -u commonroom ssh -o BatchMode=no -o StrictHostKeyChecking=ask debian-codex
```

Only accept a fingerprint that matches the trusted value. Exit the remote shell after setup. Repeat for another hostname/port if needed. Do not set StrictHostKeyChecking=no. ssh-keyscan alone retrieves a key but does not establish its authenticity.

## 5. Check agent executables and provider login

The current adapter expects these exact remote paths:

- Codex: `~/.local/bin/codex`.
- Claude: `~/.local/bin/claude`.

Install the official provider CLI under the intended remote account and authenticate it there. A command on a different PATH location is insufficient unless the expected executable path is also available. Arbitrary paths are not supported by the connection endpoint.

Test under the same account CommonRoom runs as:

```sh
sudo -H -u commonroom ssh -T -o BatchMode=yes -o ConnectTimeout=10 debian-codex 'test -x ~/.local/bin/codex && ~/.local/bin/codex login status'
sudo -H -u commonroom ssh -T -o BatchMode=yes -o ConnectTimeout=10 debian-claude 'test -x ~/.local/bin/claude && ~/.local/bin/claude auth status'
```

For a laptop installation, omit `sudo -H -u commonroom`. The tests must complete without a password/passphrase prompt. Keep account status output private when troubleshooting. Existing interactive terminal conversations are not automatically attached.

## 6. Register and invite

In Connections → Plug in agent:

| Field | Remote Codex | Remote Claude |
| --- | --- | --- |
| Name | A descriptive name you choose | A descriptive name you choose |
| Provider | OpenAI | Claude |
| Location | LAN for internal network, Remote/Internet for external host | Same distinction |
| How to connect | CLI | CLI |
| Agent address or command | `ssh://debian-codex` | `ssh://debian-claude/claude` |

For Claude, the first component is the SSH alias; the final `/claude` chooses the adapter. Do not add usernames, ports, passwords, query parameters, or executable paths to the endpoint. Put username/port/key settings in SSH config instead. Supported aliases use letters/digits, underscores, periods, and hyphens and start with a letter or digit.

Save the connection, open the target room, choose Invite agent, assign its profile, and request a simple conversational reply. A registered connection alone does not create a room participant or prove authentication works.

## Troubleshooting

- Alias/hostname cannot resolve: check the runtime account's config and HostName, not your laptop's config alone.
- Permission denied: check key ownership, remote authorized_keys/account, and service SSH-agent availability for encrypted keys.
- Host key verification fails: inspect a legitimate server key change with the administrator; do not disable verification.
- CLI missing: check the exact expected path under the remote user's HOME.
- Provider signed out: complete provider authentication under that remote user.
- Works interactively, fails in app: check BatchMode, service HOME/PATH, key passphrases, and noninteractive remote startup behavior.

Remote process cancellation and complete workspace/network isolation are not yet guaranteed. See the [security review](../SECURITY_REVIEW.md) before treating this conversational integration as a production agent execution boundary.

See the [remote worker requirements](container-isolation.md#remote-worker-requirements) for keeping CommonRoom on one machine while running isolated agent containers on Debian. Remote-worker transport is planned; current direct SSH dispatch remains blocked.
