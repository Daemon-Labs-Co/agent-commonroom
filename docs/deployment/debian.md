# Install Agent CommonRoom on a Debian server

This guide runs Agent CommonRoom privately on Debian and reaches it from your computer through an SSH tunnel. It does not enable public HTTP/HTTPS hosting or remove the application's loopback-only request checks. The commands are intended for a fresh installation; adapt them rather than replacing an existing installation or database.

The guide has been checked against the current application configuration and official dependency documentation, but has not yet been exercised end to end on a fresh Debian server.

## Deployment layout

- Dedicated Linux service account: `commonroom`.
- Source, executable, migrations, and web files: `/opt/agent-commonroom`.
- Private setup/key files: `/opt/agent-commonroom/.data`.
- Service account home and optional CLI credentials: `/var/lib/commonroom`.
- Local PostgreSQL database and owner role: `commonroom`.
- Application listener: `127.0.0.1:8787` on the Debian server.
- Browser access: `http://127.0.0.1:8787` on your computer, forwarded through SSH.

Keep the application port private. You need SSH access to the server and an administrative account with `sudo` for installation. Agent CommonRoom remains an early prototype; see [security limits](../../SECURITY.md).

## 1. Install dependencies on Debian

Connect to your server using its existing SSH alias or hostname:

```sh
ssh your-server-alias
```

On Debian:

```sh
sudo apt update
sudo apt install git ca-certificates golang-go postgresql postgresql-client
sudo systemctl enable --now postgresql
go version
psql --version
```

Check the Go version meets `go.mod` (currently at least 1.23). Debian's Go package version varies by release. If it is too old, follow the [official Go Linux installation instructions](https://go.dev/doc/install), choose the archive matching your server architecture, and verify its published checksum. If using the official `/usr/local/go` installation, add it to this terminal's PATH before building:

```sh
export PATH="/usr/local/go/bin:$PATH"
go version
```

The distribution PostgreSQL package is the starting point. Local development has used PostgreSQL 17; if you need that version specifically, use the [official PostgreSQL Debian repository instructions](https://www.postgresql.org/download/linux/debian/). Installing Node.js is optional for frontend development checks; it is not needed to run the web UI.

## 2. Create a dedicated service account and database

On a fresh server where this account and database do not already exist:

```sh
sudo useradd --system --user-group --create-home --home-dir /var/lib/commonroom --shell /usr/sbin/nologin commonroom
sudo install -d -o commonroom -g commonroom -m 0750 /opt/agent-commonroom
sudo -u postgres createuser commonroom
sudo -u postgres createdb --owner=commonroom commonroom
sudo -H -u commonroom psql -d commonroom -c 'SELECT current_user;'
```

The PostgreSQL role owns only its dedicated database; it is not a PostgreSQL superuser. This guide uses local Unix-socket peer authentication, matching the Linux account to the PostgreSQL role, so there is no database password to put into Git or an environment file. It assumes Debian's local peer-authentication configuration. If the check fails, inspect PostgreSQL's `pg_hba.conf` with your administrator rather than enabling `trust` authentication.

Startup applies the application's migration and creates event protections, so this database owner has DDL permissions. A separate migration owner and more restricted runtime role are still on the security backlog. Do not grant access to unrelated databases.

## 3. Download the source and build

Use the company's repository URL. If the repository requires authentication, configure access for the service account before cloning; do not put an access token in the URL. For a controlled release, check out an explicitly reviewed release tag or commit before building.

```sh
sudo -H -u commonroom git clone https://github.com/Daemon-Labs-Co/agent-commonroom.git /opt/agent-commonroom
sudo -H -u commonroom env PATH="$PATH" sh <<'BUILD'
set -eu
cd /opt/agent-commonroom
go mod download
go build -o bin/commonroom ./cmd/server
BUILD
sudo install -d -o commonroom -g commonroom -m 0700 /opt/agent-commonroom/.data
```

Retain `web/`, `docs/`, the root documentation files, `migrations/`, and the license notices alongside the executable: the executable does not embed these assets. Run it from the repository root.

## 4. Configure the service

Create the local environment file:

```sh
sudo install -d -m 0750 /etc/agent-commonroom
sudo tee /etc/agent-commonroom/environment >/dev/null <<'CONFIG'
DATABASE_URL="postgres://commonroom@/commonroom?host=/var/run/postgresql&sslmode=disable"
LISTEN_ADDR=127.0.0.1:8787
OWNER_KEY_FILE=/opt/agent-commonroom/.data/owner-key
CONFIG
sudo chmod 0600 /etc/agent-commonroom/environment
```

`sslmode=disable` here applies to a local Unix socket, not an unencrypted connection across a network. This is a systemd environment file, not a shell script; do not source it. The systemd service manager reads it before starting the process.

Create `/etc/systemd/system/agent-commonroom.service`:

```sh
sudo tee /etc/systemd/system/agent-commonroom.service >/dev/null <<'UNIT'
[Unit]
Description=Agent CommonRoom collaboration workspace
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=commonroom
Group=commonroom
WorkingDirectory=/opt/agent-commonroom
Environment=HOME=/var/lib/commonroom
Environment=PATH=/var/lib/commonroom/.local/bin:/usr/local/bin:/usr/bin:/bin
EnvironmentFile=/etc/agent-commonroom/environment
ExecStart=/opt/agent-commonroom/bin/commonroom
Restart=on-failure
RestartSec=5
UMask=0077
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ReadWritePaths=/opt/agent-commonroom/.data /var/lib/commonroom

[Install]
WantedBy=multi-user.target
UNIT
sudo systemd-analyze verify /etc/systemd/system/agent-commonroom.service
sudo systemctl daemon-reload
sudo systemctl enable --now agent-commonroom
sudo systemctl status agent-commonroom --no-pager
```

The service has explicit writable locations for its local state and CLI credentials. It does not give agents a complete security sandbox or enforce room workspace policies. If an agent or later feature needs additional filesystem access, review that requirement rather than disabling protections indiscriminately. CLI programs must be available on the service PATH, which is separate from your interactive shell's PATH.

## 5. Open an SSH tunnel from your computer

Run this on your own computer, not inside the server's SSH shell:

```sh
ssh -N -o ExitOnForwardFailure=yes -L 127.0.0.1:8787:127.0.0.1:8787 your-server-alias
```

Leave the tunnel running and open [Agent CommonRoom](http://127.0.0.1:8787) in your local browser. SSH encrypts the server connection; the browser connects to local loopback. Do not forward the port onto `0.0.0.0`, open the application port in the server firewall, or add a public reverse proxy.

**The local forwarded port must equal `LISTEN_ADDR`'s server port.** The application checks the browser's Host and Origin against that port. If your laptop already runs Agent CommonRoom on 8787, either stop that local instance or change the server configuration to another loopback port, such as 8788, restart the service, and forward **8788 to 8788**:

```sh
ssh -N -o ExitOnForwardFailure=yes -L 127.0.0.1:8788:127.0.0.1:8788 your-server-alias
```

For that configuration, open `http://127.0.0.1:8788`. Forwarding local 8788 to server 8787 will fail the current request checks.

## 6. Finish owner setup in your local browser

For a first installation, read the generated setup code privately on the server:

```sh
sudo -H -u commonroom cat /opt/agent-commonroom/.data/owner-key
```

Paste it into the browser through the tunnel, download your owner private key, store it securely, and sign in with the downloaded key. Do not paste the setup code or private key into Git issues, chat rooms, logs, or commits.

After successful RSA key sign-in, the server removes the setup code and pending private-key copy; the public-key file remains. Later sign-ins use your saved key. Restarting the service invalidates browser sessions but does not invalidate that saved key. Key recovery and rotation are not implemented.

## 7. Configure agent connections for the server

Use the [step-by-step SSH agent guide](../agent-ssh.md) for host aliases, remote usernames/ports, keys, host verification, and runtime-account connection tests.

When Agent CommonRoom runs on Debian, **Local means the Debian server**, not your laptop.

- A local `codex` or `claude` connection runs as the `commonroom` Linux account and uses that account's provider credentials.
- Install provider CLIs using their official instructions and ensure the executable is on the systemd service's PATH. Install/login under the intended account, not as root.
- macOS Terminal-based Claude sign-in launch is unavailable on Debian. Authenticate the CLI through an interactive server terminal, for example `sudo -H -u commonroom /var/lib/commonroom/.local/bin/claude auth login --claudeai`, if installed at that path. Follow the CLI's displayed authentication instructions. Provider login is separate from the Agent CommonRoom owner key.
- To use an existing agent account or another machine over SSH, use a LAN/Remote CLI endpoint such as `ssh://agent-host` for Codex or `ssh://agent-host/claude` for Claude. Configure the SSH alias, verified host keys, and noninteractive authentication for **the service account**; your laptop's SSH config does not automatically transfer to it.
- Remote executables are currently expected at `~/.local/bin/codex` or `~/.local/bin/claude` on the agent host. Arbitrary executable paths are not supported.
- Native interactive sessions are not attached merely by installing a connection. Supported saved-session import/resume and additional context are separate invitation options.

Connecting multiple browsers through SSH does not turn this into a hardened internet service. Fine-grained multi-actor policy and process containment remain incomplete.

## Operations, updates, and backups

Useful server commands:

```sh
sudo journalctl -u agent-commonroom -n 100 --no-pager
sudo systemctl restart agent-commonroom
sudo systemctl stop agent-commonroom
```

Do not share raw logs without reviewing them for private data. For an update, finish/cancel active agent work, stop the service, back up the database and key/state files, deploy a reviewed commit/release, build again, and restart. Startup migration changes need review before deployment. A stopped web service is not currently a guarantee that every remote agent process has terminated; inspect active remote jobs separately.

For a new backup, create a uniquely named private directory outside the repository and store a database dump there:

```sh
commonroom_backup_dir="/var/backups/agent-commonroom/$(date -u +%Y%m%dT%H%M%SZ)"
sudo install -d -m 0700 "$commonroom_backup_dir"
sudo sh -c 'umask 077; sudo -H -u commonroom pg_dump -Fc -d commonroom > "$1/database.dump"' sh "$commonroom_backup_dir"
sudo cp -a /opt/agent-commonroom/.data "$commonroom_backup_dir/state"
```

For a coherent maintenance backup, first stop Agent CommonRoom as described above. Protect these backups as private data and keep an encrypted off-server copy. Back up the service account's provider credentials separately if required. The human's downloaded private key must have its own protected backup; the server removes its private-key copy after setup. Protect the environment file and document deployed commit/version too. Do not erase the live event log to reset an installation.

This guide does not implement backup automation, restoration tooling, external immutable/WORM archives, absolute tamper resistance, or retention/purge enforcement. Test restoration in a separate environment.

## Troubleshooting

| Symptom | Check |
| --- | --- |
| PostgreSQL peer authentication fails | Linux account, DB role, socket directory, and `pg_hba.conf`; do not change authentication to `trust`. |
| Migration or web assets missing | Service `WorkingDirectory` and deployed `migrations/`/`web/` directories. |
| Permission denied writing state | `.data` ownership/mode and systemd `ReadWritePaths`. |
| Browser gets an untrusted-host/request error | Browser uses localhost/127.0.0.1 and tunnel local port matches server `LISTEN_ADDR`. |
| Local tunnel address already in use | Stop the conflicting local service or change both server and tunnel ports together. |
| CLI works in your shell but not in the app | Service account, HOME, PATH, provider login, and executable permissions. |
| SSH agent connection fails | Service-account SSH alias/host key, BatchMode authentication, and expected remote executable. |
| Owner setup code missing | Setup may already be complete; use the saved private key and retain the public-key file. |

## Reference documentation

- [Go installation](https://go.dev/doc/install)
- [PostgreSQL Debian packages](https://www.postgresql.org/download/linux/debian/)
- [PostgreSQL peer authentication](https://www.postgresql.org/docs/current/auth-peer.html)
- [OpenSSH port forwarding](https://man.openbsd.org/ssh.1)
- [systemd service units](https://www.freedesktop.org/software/systemd/man/latest/systemd.service.html)
- [systemd execution options](https://www.freedesktop.org/software/systemd/man/latest/systemd.exec.html)
