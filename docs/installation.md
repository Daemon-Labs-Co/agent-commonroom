# Install the local controller

[Documentation home](README.md)

This guide gets the controller and human workspace running. Agent execution is a
separate setup step. Repository-bound agent development remains unavailable to
users until the isolated runtime passes production admission and live acceptance.

## 1. Install prerequisites

You need **Go 1.26+**, **PostgreSQL** (17 has been exercised), and a current browser
with JavaScript and Web Crypto. **Node.js and Python 3.8+** are needed for
development checks. Python checks use only the standard library; no Python
packages or provider credentials are required.
There is no frontend package install or bundling step.

On macOS with Homebrew:

```sh
brew install go postgresql@17 node
brew services start postgresql@17
export PATH="$(brew --prefix postgresql@17)/bin:$PATH"
git clone https://github.com/Daemon-Labs-Co/agent-commonroom.git
cd agent-commonroom
go mod download
```

If Go/PostgreSQL are already installed, skip their installation. On Debian or
other systems, use the [remote installation guide](deployment/debian.md).
Always run the server and migration from the repository root.

## 2. Prepare a dedicated database

CommonRoom requires **two separate database logins**:

| Login | Purpose |
| --- | --- |
| `commonroom_migrator` | Owns the dedicated database/schema and runs explicit migrations. |
| `commonroom_runtime` | Serves the application; can read and append events, but cannot rewrite or delete them. |

Complete the [one-time administrator setup](database-security.md#administrator-setup).
It includes role/database creation, authentication and the exact restricted grants.
Use an existing PostgreSQL installation if you have one; no separate server is
required. Do not reuse another application's database or give the runtime login
ownership, role memberships, DDL, UPDATE, DELETE or TRUNCATE privileges.

For local TCP connections, configure SCRAM passwords and save them in a private
[PostgreSQL password file](https://www.postgresql.org/docs/current/libpq-pgpass.html)
with mode `0600`. Alternatively, use properly configured same-host peer
authentication. Do not put passwords in Git, shell command arguments or a `trust`
rule. The URLs below assume the documented local authentication is ready.

## 3. Migrate the schema and apply runtime grants

```sh
export MIGRATION_DATABASE_URL='postgres://commonroom_migrator@localhost:5432/commonroom?sslmode=disable'
go run ./cmd/server --migrate
unset MIGRATION_DATABASE_URL
```

Apply the runtime grants from [database security setup](database-security.md#administrator-setup)
**after** the migration creates the event table and sequence. Startup refuses an
overprivileged runtime account. `sslmode=disable` is only for a same-machine
local database; use verified TLS for a database reached across a network.

## 4. Start the controller

```sh
export DATABASE_URL='postgres://commonroom_runtime@localhost:5432/commonroom?sslmode=disable'
go run ./cmd/server
```

Open **[http://127.0.0.1:8787/](http://127.0.0.1:8787/)**. Keep this terminal running.
The UI and human collaboration features do not require a provider login. Agent
turns require the worker setup in step 6.

For a reusable binary, run `make build`, then `./bin/commonroom` from the same
repository root. Stop the controller with Ctrl+C after stopping active work.

## 5. Create the owner account

1. Read the private setup code in `.data/owner-key` and enter it in the sign-in screen.
2. Download your owner PEM key, keep a secure backup, and complete key sign-in.
3. Invite other humans from Settings; there is no public signup.

Later sign-ins use your PEM key by paste or upload. Refreshing an authenticated
page does not require it again. Keys, sessions and local configuration in `.data/`
must stay private and out of Git. Account key recovery is not yet implemented.

## 6. Add a worker when you need agents

Follow [isolated worker and provider sign-in setup](connection-sign-in.md).
The worker needs Linux, rootless Docker with enforceable resource limits, an
approved digest-pinned agent image, and noninteractive SSH with verified host keys.
Provider tools and credentials live in isolated connection storage; installing a
CLI on the controller's PATH does not enable managed execution.

After configuring the worker, set these on the **controller**, then restart:

```sh
export COMMONROOM_REMOTE_WORKER=true
export COMMONROOM_REMOTE_WORKER_ALIAS=commonroom-worker
```

`commonroom-worker` is an example SSH alias; configure your own alias as described
in the [SSH guide](agent-ssh.md). In **Connections**:

1. Create a containerized CLI connection with `ssh://commonroom-worker` for Codex,
   or `ssh://commonroom-worker/claude` for Claude.
2. Choose **Sign in** and complete that connection's provider authentication.
3. Choose **Test connection**, then invite an agent into a room.

Install matching controller/worker binaries supporting lifecycle protocol v2.
See [worker verification and recovery](worker-jobs.md) for rollout and diagnostics.
A local Linux worker also uses the approved SSH transport. Managed agent turns
cannot run directly on macOS or through unrestricted SSH commands.

## Other deployment options

- **[Debian server](deployment/debian.md):** dedicated account, PostgreSQL,
  systemd, worker setup, private SSH access and backups.
- **[nginx with existing PostgreSQL](deployment/nginx-postgresql.md):** private
  reverse-proxy deployment and reuse of an existing database server.
- **[SSH connection details](agent-ssh.md):** aliases, executables and approved
  worker transport.

The controller binds numeric loopback addresses. Use the documented private
SSH tunnel/proxy setup; public internet exposure is unsupported. pgvector is an
optional DBA-managed prerequisite for future semantic memory, not needed to run
the current application. Fresh-server deployment remains an integration check
rather than a claim established by the mocked test suite.


## Optional room coordination

Room appointments are off by default. The owner enables `agentCoordination` in
Settings without restart, then selects a loaded agent in the room. Disabling
cancels local role-dependent turns and preserves appointments and history. Old
dependencies remain invalid after re-enable; ordinary independent work keeps its
existing permissions. Security checks and human stop remain active.
See [coordination and authority](governance-appointments.md) for scope, running-work
behavior and remaining governance implementation.


## Optional conversation task proposals

`taskProposals` defaults off. The owner enables it in Settings without restart,
then captures a room suggestion and reviews its source in Tasks. Approval creates
an unassigned Draft, never an execution. Off preserves all history and tasks and
refuses new captures/decisions; independent work retains its existing permissions.
See [task proposals](task-proposals.md) for usage, API, permissions and limits.


## Optional bounded delegation

`taskDelegation` defaults off and is owner-controlled in Settings, without restart.
Review controls appear in Tasks. Off invalidates old delegation authority while
preserving history and independent human jobs. Provider adapters still need to
emit the typed private supervisor request; this is not autonomous model delegation
out of the box. See [bounded delegation](task-delegations.md) for setup, lifecycle,
permissions, expiry/recovery and verification limits.


Optional room definitions: the owner can enable default-off `roomOutcomes` in Settings → Features → Workflows. No restart or data migration is required. Existing rooms can be defined gradually; [room outcomes](room-outcomes.md) documents server enforcement and limits.

The separate `roomOutcomeContext` setting defaults off. Enable it only when the owner wants room deliverables included in task-provider context; `roomOutcomes` must also be on. Existing provider allowlists and reference-export exclusions remain enforced, with no restart.

The default-off `roomOutcomeReview` setting enables human room-deliverable acceptance and reopening. It requires Room outcomes, with no restart. Existing history is preserved when disabled. See [room review](room-outcome-review.md) before enabling.
