# Agent CommonRoom

A local workspace where humans and AI agents collaborate in rooms, with individual profiles, persistent conversations, shared memory, tasks, evidence, and a visual workflow editor.

Agent CommonRoom is an early development prototype built with Go, PostgreSQL, and browser JavaScript. It supports Codex and Claude CLI agents through a mandatory rootless Docker worker on Linux, reachable over approved SSH. The workflow editor saves and validates graphs; sequential execution is available behind the default-off `workflowExecution` flag. See the guides for supported nodes and remaining limits.

For a blocked or stale agent, the owner can select **Recover session**, **Check worker**, then **Resume session** or **Start fresh session**. Recovery requires the shared worker to be idle, preserves room history, and never automatically replays a task. The default-on Security flag `agentRecovery` controls this feature. Install the current worker binary for its read-only `idle-v1` protocol. See [safe session recovery](docs/concepts.md#safe-session-recovery) for usage, permissions and limitations.

## Documentation

Read these guides inside the app using **Help** in the header or **Help & documentation** beside Settings. Contextual Help links are available on feature screens and forms. The standalone reader is at `/help/`.

Browse the [documentation index](docs/README.md) for product concepts, architecture, the HTTP API, development, operations, and Debian deployment. Detailed source comments and per-file technical references will be added in subsequent passes.

## Screenshots

These screenshots use fictional people, conversations, and connections in a separate read-only preview of the current interface. They contain no live workspace data and do not imply workflow execution is available.

### Room conversation and squad

![Agent CommonRoom room conversation with a human and two agent profiles](docs/screenshots/room.jpg)

### Visual workflow editor

![Agent CommonRoom workflow canvas with node library and properties panel](docs/screenshots/workflow.jpg)

Edit, validate and save workflows. Enable the optional sequential runner to dispatch isolated agent steps, independent reviews and explicit human gates. See [workflow execution](docs/workflow-execution.md) for limits and recovery.

### Agent connections

![Agent CommonRoom connection management with local Codex and Claude examples](docs/screenshots/connections.jpg)

## What works today

- Create rooms and invite agents with room-specific names, engineering titles, seniority, instructions, and challenge levels.
- Register, edit, and delete connections independently of room memberships.
- Send room messages, direct messages, and `@room` broadcasts. Native agent session IDs persist across follow-up turns.
- Copy or resume supported existing local Codex conversations, and optionally supply additional context when inviting agents.
- Store shared memory with Room, Project, or Application scope and revision checks. Project scope currently means the one shared project, not separate project permissions.
- Assign tasks and reviewers, record evidence, and design workflows using draggable nodes, connections, decisions, review steps, and room handoffs.
- Invite individual human accounts with Developer or QA roles and access to selected rooms. Each human and registered bot identity has a personal room; owner access to another person's room is audited.
- Sign in with an individual RSA key. Launch local Claude sign-in from the Connections screen on macOS.
- Preserve application events in a PostgreSQL append-only log with a checksum chain and manual integrity inspection.

## Requirements

- Go 1.26 or later, as specified in `go.mod`.
- A running PostgreSQL server; local development has been exercised with PostgreSQL 17.
- A current browser supporting Web Crypto, with JavaScript enabled.
- Optional: authenticated `codex` and/or `claude` executables on the Agent CommonRoom server's `PATH`.
- Optional: SSH aliases, noninteractive SSH authentication, and the appropriate CLI installed on a remote agent machine.

The server, migration, and web assets must be run from the repository root. There is no frontend package installation or bundling step.

## Remote server installation

For nginx with an existing database, see [nginx and existing PostgreSQL](docs/deployment/nginx-postgresql.md). pgvector is documented as an optional DBA-managed prerequisite; semantic memory is not implemented.

To run Agent CommonRoom on your Debian server, follow the [Debian remote-install guide](docs/deployment/debian.md). It covers dependencies, a dedicated service account, PostgreSQL, systemd startup, owner setup, agent authentication, SSH-tunnel browser access, backups, and troubleshooting.

The supported approach keeps the app on server loopback and forwards it over SSH. Public HTTP/HTTPS hosting is not implemented. When installed remotely, local agents run on the server rather than your laptop. The guide is source-checked documentation; a fresh Debian deployment has not yet been validated end to end.

## Installation

### 1. Get the source

Clone this repository from its published hosting page, or choose Download ZIP and extract it. Open a terminal in the extracted repository root—the directory containing `go.mod`, `web/`, and `migrations/`.

### 2. Install Go and PostgreSQL

Install [Go](https://go.dev/doc/install) and [PostgreSQL](https://www.postgresql.org/download/) for your operating system. Go must meet the version requirement above. Start PostgreSQL using your installation's service manager.

For a fresh macOS setup with Homebrew already installed:

```sh
brew install go postgresql@17
brew services start postgresql@17
export PATH="$(brew --prefix postgresql@17)/bin:$PATH"
```

PostgreSQL 17 is keg-only in Homebrew, so the PATH step makes `createdb` and other database commands available in this terminal. See the [Homebrew PostgreSQL 17 instructions](https://formulae.brew.sh/formula/postgresql@17). Skip installation or service-start steps if your dependencies are already installed/running. The service command also enables PostgreSQL at login.

Verify the tools and database server:

```sh
go version
psql --version
pg_isready -h localhost -p 5432
```

The shell commands below use macOS/Linux syntax. Windows users should use their corresponding PostgreSQL service controls and shell environment-variable syntax; native Windows deployment has not been validated.

### 3. Create the database and start Agent CommonRoom

Create a dedicated local database using an account allowed to create databases:

```sh
createdb commonroom
go mod download
export MIGRATION_DATABASE_URL='postgres://localhost:5432/commonroom?sslmode=disable'
go run ./cmd/server --migrate
unset MIGRATION_DATABASE_URL
# Complete the restricted runtime-role setup linked below before serving.
export DATABASE_URL='postgres://commonroom_runtime@localhost:5432/commonroom?sslmode=disable'
go run ./cmd/server
```

The example assumes PostgreSQL accepts your local account. Adjust the connection string for your database installation; never commit credentials. `sslmode=disable` here is for a local development database. Run migrations explicitly with administrator credentials, then serve with a separate restricted runtime login. Follow the [database security setup](docs/database-security.md) to create that login and configure authentication. Startup refuses database owners, role memberships, DDL privileges, or permissions to alter audit rows.

Open [Agent CommonRoom](http://127.0.0.1:8787). The server listens on numeric loopback addresses only. Use remote access only through the documented private SSH tunnel or private nginx setup; public hosting is unsupported.

### 4. Complete first owner sign-in

1. Open the generated `.data/owner-key` file locally and paste its setup code into the sign-in screen. Keep this code private.
2. Download the owner private key when prompted. Store it securely with a protected backup.
3. Choose sign-in with a saved key and upload or paste the downloaded PKCS#8 PEM file.

The browser signs a one-use, expiring challenge with RSA-PSS/SHA-256. Routine key login sends the signature, not the private key. After successful key login, the server removes its pending private-key copy and setup code and retains the public key. Filesystem removal does not guarantee erasure from disks or backups. Downloaded keys are unencrypted; encrypted PEM import, key recovery, and rotation are not implemented.

Sessions expire after eight hours or a server restart. Do not delete the retained public-key file: it is needed to recognize your saved key.

### 5. Create a room and invite people or agents

Use the **+** beside Rooms to create a shared room. Use Connections → Plug in agent to register a provider, then Invite agent inside your room to assign its profile. Agent CLI installation and authentication are separate prerequisites; see Connect agents below.

### Invite people

Under Settings → Company members, invite a human, assign Developer or QA, and select their shared-room access. An individual JSON key credential downloads in the owner's browser. Deliver it privately to its intended recipient; they upload or paste it at sign-in. Humans cannot register themselves. Only the owner can change roles/access or disable accounts. Developer and QA permissions are currently limited and do not implement a complete organizational permission system.

## Connect agents

Create a connection first, then invite it into a room with its own profile.

| Agent | Protocol | Endpoint | Authentication |
| --- | --- | --- | --- |
| Local Codex | CLI | `codex` | Authenticate the local CLI separately |
| Local Claude | CLI | `claude` | Connections → Claude sign-in on macOS, or authenticate the CLI separately |
| SSH Codex | CLI | `ssh://your-ssh-alias` | Authenticate Codex on the remote machine |
| SSH Claude | CLI | `ssh://your-ssh-alias/claude` | Authenticate Claude on the remote machine |

For complete alias, hostname, user, port, key, host-verification, and connection-test examples, follow [SSH agent setup](docs/agent-ssh.md).

SSH uses configured aliases and noninteractive authentication. Remote executables are currently expected at `~/.local/bin/codex` or `~/.local/bin/claude`; arbitrary remote executable paths are not supported. Never place passwords, tokens, or private keys in endpoints.

Local means the machine running the Agent CommonRoom server. LAN means a reachable internal-network machine. Remote/Internet means an external service or machine. These labels describe a connection's location; they do not enable unfinished adapters.

HTTP, WebSocket, and MCP connections can be registered, but dispatch through those protocols is not implemented. ChatGPT and Claude website URLs are not conversational agent API endpoints. The ChatGPT-labelled participant uses a Codex CLI conversation, not a ChatGPT website session.

### Claude sign-in from Agent CommonRoom

For a registered local `claude` connection, the owner can choose Claude sign-in → Open Claude sign-in. On macOS this opens the official CLI login in Terminal/browser. Finish the flow and choose Check sign-in. Agent CommonRoom does not receive your password or authentication codes. All local Claude connections share the local CLI account. Remote sign-in must be completed on the remote machine.

### Conversations and safety boundaries

Messages default to the room. Select an agent or use its mention to request a response; `@room` requests attention from the supported room agents. Merely posting to the room does not guarantee an agent turn. Additional room history and memory are supplied when an agent is invoked, while its native session preserves its own prior turns.

Managed turns require the [Linux/rootless-Docker worker](docs/container-isolation.md), locally on supported Linux or through the approved remote worker. There is no host-execution bypass. Chat connections remain read-only; administrator-approved [development workspaces](docs/development-workspaces.md) enable private writable repository snapshots for selected rooms. Existing host sessions cannot automatically resume inside the worker. General directory permissions, automatic patch export and shared workspace synchronization remain incomplete.

## Tasks, evidence, and workflows

Tasks support assignment, reviewer selection, lifecycle edits, explicit dispatch, and human-entered evidence. Evidence is organized as Changes, Checks, Artifacts, and Review. Check counts are not inferred, and automatic evidence collection and artifact uploads are not implemented.

In a room's Workflows tab, add and connect nodes, edit their properties, arrange the graph, and save it. Validate checks assignments, connectivity, and reachability; incomplete drafts can still be saved. Delete removes a selected node, and Undo restores edits. The optional `workflowExecution` flag enables Run for saved sequential graphs. The runner records attempts, routes explicit review outcomes, enforces iteration limits and pauses for evidence and human approval. Parallel work, automatic retries, cross-room handoffs and automatic patch/test-evidence export remain unavailable. See [usage and recovery](docs/workflow-execution.md).

The company organization chart supplies reporting-authority guidance: managers outrank reports, peers need consensus and humans retain the final override. Workflow human acceptance remains explicitly owner-controlled. Automatic conflict arbitration and escalation timers remain planned.

## Configuration

| Setting | Default | Purpose |
| --- | --- | --- |
| `DATABASE_URL` | Local config, then `postgres://localhost:5432/agentroom?sslmode=disable` | PostgreSQL connection |
| `COMMONROOM_AGENT_SAFETY_UI` | `true` | Initial emergency controls visibility; owner Settings override persists; backend safety remains enforced when hidden |
| `COMMONROOM_WORKFLOW_TEMPLATES` | `true` | Main-menu template catalog; set `false` and restart to restore the previous template option |
| `COMMONROOM_CONNECTION_EDITOR` | `true` | Guided connection form; set `false` and restart to restore the previous form |
| `LISTEN_ADDR` | `127.0.0.1:8787` | Numeric loopback bind address |
| `OWNER_KEY_FILE` | `.data/owner-key` | Initial setup-code path; public/private companion paths use suffixes |

If `DATABASE_URL` is unset, `.data/local-config.json` can supply `{"databaseUrl":"postgres://localhost:5432/commonroom?sslmode=disable"}`. The server does not automatically load `.env`; `.env.example` documents values to export yourself. `.data/` contains private local configuration and authentication material and must stay out of version control.

To build a local executable:

```sh
go build -o bin/commonroom ./cmd/server
./bin/commonroom
```

Run that executable from the repository root. Stop the server with Ctrl+C.

## Installation troubleshooting

- **PostgreSQL unavailable / connection refused:** check the service is running and that `DATABASE_URL` uses its actual host and port. Run `pg_isready` before starting the app.
- **Role does not exist / permission denied:** use a database login configured for your PostgreSQL installation. Database creation and explicit migrations require administrator permissions; ask the database administrator if using a managed installation.
- **Database already exists:** omit `createdb` and reuse the intended database. Do not drop a live database as a setup shortcut.
- **Missing web assets or migration:** start the server from the repository root, including when using the built executable.
- **Address already in use:** stop the older Agent CommonRoom process or choose another loopback port, for example `export LISTEN_ADDR='127.0.0.1:8788'`, then use that port in your browser.
- **CLI not installed or signed out:** make `codex` or `claude` available on the server's PATH and complete its official authentication. The Connections screen supports local Claude sign-in launch on macOS.
- **Key no longer recognized:** ensure you are using the same database and `OWNER_KEY_FILE` public-key companion as the original installation. Key recovery is not implemented; retain secure backups.

## Development and checks

```sh
make check
make build
```

`make secrets` runs the redacted Gitleaks secret scan; setup and the pre-commit hook are documented in [CONTRIBUTING.md](CONTRIBUTING.md).

`make test` runs the fast Go suite. `make check` adds race detection, Go vet, JavaScript syntax checks, and help-reader security tests. `make coverage` reports statement coverage into the ignored `coverage/` directory. Tests run in parallel (default 8; override with `TEST_PARALLEL=4`) and have a 30-second timeout.

Tests use an in-memory database mock: **no PostgreSQL connection, schema creation, or database writes**, even if database environment variables are set. Provider command tests inspect arguments without launching agents or SSH. File fixtures use temporary directories. PostgreSQL, provider credentials, and running agents are not required for the suite. Mocks do not validate actual PostgreSQL migrations or provider interoperability.

## Code map

| Path | Responsibility |
| --- | --- |
| `cmd/server/` | HTTP server, authentication, membership/access rules, agent adapters, history, workflows, and tests |
| `internal/room/` | Domain models |
| `web/` | Browser UI, styles, and provider icons |
| `migrations/` | PostgreSQL event schema and protections |
| `demo/` | Original standalone UI prototype |
| `FUNCTIONAL_SPEC.md` | Product requirements, including planned behavior |
| `SECURITY_REVIEW.md` | Security findings and architecture backlog |
| `DEVELOPMENT_STATUS.md` | Historical development notes; this README is the current capability overview |

## Current limits and roadmap

Agent CommonRoom is not hardened for hostile multi-user or internet deployment. Major remaining work includes durable workflow execution, complete per-actor and project permissions, enforceable agent workspaces, additional transport adapters, bounded resource controls, watchdogs, automatic QA/evidence, key recovery, and retention/purge workflows.

The local event trigger and checksum chain do **not** provide absolute immutability against a database administrator. Checksums currently omit some record metadata, and there is no independent rollback anchor or external immutable archive. Integrity failure handling does not yet guarantee cancellation of every active remote/local process. Preserve database backups and key material securely; do not rewrite historical events to reset a live installation. See the [security review](SECURITY_REVIEW.md) for details.

## Contributing and licensing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines and [SECURITY.md](SECURITY.md) for vulnerability handling. Bundled provider icons retain their upstream MIT notice in `web/icons/LICENSE.lobe-icons`. Provider names and logos identify supported integrations and do not imply endorsement. Agent CommonRoom is licensed under [GNU GPL version 3 only](LICENSE) (`GPL-3.0-only`). Copyright © 2026 Agent CommonRoom contributors. Distributed derivative software must retain GPL terms and provide corresponding source as required by the license. GPLv3 does not impose a source-sharing requirement merely for operating modified software as a network service; that is a distinction from AGPL.

The [guided connection editor](docs/concepts.md#guided-connection-editor) explains provider, method, and location choices with concise contextual help.

Browse **Workflow templates** in the main menu and apply an [idea, development, UI, bug, security, documentation, disaster-planning, or cooperative adversarial template](docs/concepts.md#workflow-templates) to a room. These create saved plans; automatic execution is not yet enabled.

Owners can [stop agent dispatch for a room or the application](docs/concepts.md#emergency-agent-cutoff), request cancellation, and explicitly resume. Remote/child-process termination is not guaranteed.

### Grouped feature flags and human photos

Owners manage flags in **Settings → Feature flags**. **Security** contains emergency controls; **Cosmetic** contains human profile photos (enabled by default). **Connections** and **Workflows** identify the existing environment-configured connection editor and template flags; those require a restart.

Edit your human profile in Settings or the squad, then select **Change profile photo**. Upload an optional JPEG/PNG, preview it and save; select **Remove my photo** to return to initials. Photos appear in the squad and beside your human messages across accessible rooms. See [profile photos and flags](docs/concepts.md#human-profile-photos-and-feature-flags) for limits and retention.

### Mandatory agent worker setup

Before sending an agent turn, follow the [container execution setup](docs/container-isolation.md): configure a local Linux rootless Docker daemon with enforceable limits, a reviewed digest-pinned provider image and dedicated provider credential files. Missing prerequisites block agent execution; the browser UI remains usable. No isolation-disable feature flag exists.

See the [remote worker requirements](docs/container-isolation.md#remote-worker-requirements) for keeping CommonRoom on one machine while running isolated agent containers on Debian. Remote-worker transport is planned; current direct SSH dispatch remains blocked.

See [isolated connection sign-in](docs/connection-sign-in.md) for remote Codex/Claude authentication, connection credential storage and the Security feature flag.

[Testing connections](docs/connection-testing.md): per-connection setup and provider-response checks and troubleshooting from the Connections screen.

[Help window preference](docs/help-display.md): choose same-window or separate-window documentation per human profile.

Squad names now open profiles; human Leave appears inside the eligible room-profile overlay. Agent invitations use spacious Agent, Role and Context tabs. Standard question-mark help controls sit beside the overlay close X or the relevant field. These have owner-controlled cosmetic flags under Settings → Feature flags → Cosmetic. See [squad profiles, invitations and flag behavior](docs/concepts.md#squad-profiles-and-invitations).

Human sessions now survive refreshes and server restarts for their existing eight-hour lifetime, using private server-side storage rather than saved PEM keys. Set `COMMONROOM_PERSISTENT_SESSIONS=false` to disable restart persistence; restart required. See [human login persistence](docs/concepts.md#human-login-persistence). Feature controls are grouped into Security, Cosmetic, Connections and Workflows tabs under Settings → Feature flags.

History date tabs separate Today, Yesterday, recent weeks and previous months; search filters the selected period. See [history date tabs](docs/concepts.md#history-date-tabs).

Settings is organized into Profile, Company, Features, Retention, Help and About tabs according to the current human’s permissions. Retention policy is under Settings; the sidebar History page focuses on activity. This does not enable purge or alter the minimum 30-day retention requirement.

## Planned feature work

Page content now anchors at the top-left, and long agent/human profile overlays use tabs. Both are owner-controlled Cosmetic features; see [layout guidance](docs/concepts.md#page-alignment-and-compact-profile-overlays).

The visual organization chart, automatic tasks from discussions and automatic/human-managed development branch creation are separate upcoming features. See [planned feature work](docs/planned-features.md). Human profiles now show read-only account and visible room details; see [profile details](docs/concepts.md#human-account-and-room-details).

The owner can manage a visual company organization chart under Settings → Company. Double-click an agent to edit its connection, or a human to edit their profile; an equivalent Edit button is available in the properties panel. Reporting relationships replace numbered decision seniority while the `companyOrgChart` feature is enabled. See [organization and authority](docs/concepts.md#company-organization-and-decision-authority) for usage, permissions, compatibility and current arbitration limitations.

For agents that need to edit code and run tests, configure [ready-to-work development workspaces](docs/development-workspaces.md). Chat/provider connectivity alone does not establish a writable repository or a functioning command sandbox.
