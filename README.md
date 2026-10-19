# Agent CommonRoom

## Introduction

Bring humans and different AI agents into one workspace: discuss ideas, challenge
proposals, assign work and review the evidence behind decisions. A connection can
supply several agents with different roles, so a staff engineer, reviewer and
specialist can contribute distinct perspectives while the human keeps authority.
The record should explain who asked for what, what was decided and what happened.

CommonRoom is an open-source AI harness from **Daemon Labs**, licensed under
[GPL-3.0-only](LICENSE). Use it when a single assistant conversation is too narrow
and you want collaboration with accountable handoffs. Run the controller locally
or on your own private Linux server and use the browser interface from your laptop.
A hosted service and paid enterprise capabilities are planned, not current offers.

**Current stage: early prototype.** Human collaboration, tasks, memory and workflow
foundations exist. Isolated Codex/Claude chat paths require an approved Linux worker.
The new Codex development components have synthetic tests but no production
repository-development dispatch or live provider acceptance. Repository-bound
agent development remains blocked. Public internet and hostile multi-tenant
hosting are unsupported. Check [release readiness](docs/release-readiness.md)
before choosing this for a deployment.

## Quick start guide

Get the human workspace running first; add an isolated worker when you need agents.
You need a trusted source checkout, **Go 1.26+**, PostgreSQL and a current browser.
Node and Python are needed only for development checks. The scripts build the
controller; database roles and authentication are an explicit one-time setup.

| Environment | Setup script | Walkthrough |
| --- | --- | --- |
| macOS | `bash scripts/setup-macos.sh` | [Local setup](docs/quick-start.md) |
| Linux | `bash scripts/setup-linux.sh` | [Local or server setup](docs/quick-start.md) |
| Windows / WSL2 | `powershell -File scripts/setup-windows.ps1 -LinuxRepoPath /home/YOU/agent-commonroom` | [Windows prerequisites](docs/quick-start.md#1-choose-your-platform) |
| Linux + nginx + Docker + existing/separate PostgreSQL | `bash scripts/setup-hybrid.sh` | [Hybrid walkthrough](docs/quick-start.md#4-hybrid-linux-controller-nginx-docker-and-separate-postgresql) |

After [creating the dedicated database and separate roles](docs/database-security.md#administrator-setup),
use a private passfile or approved peer authentication. These local examples contain
no passwords; a remote database needs verified TLS instead of `sslmode=disable`.

```sh
export MIGRATION_DATABASE_URL='postgres://commonroom_migrator@localhost:5432/commonroom?sslmode=disable'
bash scripts/setup.sh --migrate
unset MIGRATION_DATABASE_URL
# Apply the runtime grants from the database guide before starting.
export DATABASE_URL='postgres://commonroom_runtime@localhost:5432/commonroom?sslmode=disable'
bash scripts/setup.sh --start
```

Open **http://127.0.0.1:8787/**, enter the private setup code from `.data/owner-key`,
and create your owner account. Save the downloaded PEM sign-in key securely.
Create a room and start a conversation; no provider account is needed for human
collaboration. The scripts do not load `.env` files or alter existing system services.
See [quick start](docs/quick-start.md) for prerequisites, script modes and troubleshooting.

### Verify your installation

With the controller running, open a second terminal as the same service account:

```sh
python3 scripts/check-installation.py
# On the Linux host, also check nginx and the current Docker identity:
python3 scripts/check-installation.py --hybrid
```

Requires Python 3.8+. Each failed stage includes repair guidance; skipped checks
are reported as incomplete. Export the runtime `DATABASE_URL` in that terminal
for the read-only database check. For another port/tunnel, pass
`--url http://127.0.0.1:YOUR_PORT`. See [diagnostic details](docs/installation-check.md).

## Installation process

Choose each component separately, then compile, migrate and start. This keeps
setup failures understandable and avoids changing an existing database or web server.

| Step | Options and instructions |
| --- | --- |
| 1. Get the source | [Checkout](docs/installation-options.md#source-checkout) |
| 2. Install Go | [mise](docs/installation-options.md#go-through-mise) or [official installer / packages](docs/installation-options.md#go-through-the-official-installer-or-os-package-manager) |
| 3. Prepare PostgreSQL | [Native/existing service](docs/installation-options.md#postgresql-native-or-existing-service) or [local Docker database](docs/installation-options.md#postgresql-optional-local-docker-service) |
| 4. Compile and initialize | [Build, migration, restricted grants and startup](docs/installation-options.md#compile-migrate-grant-serve) |
| 5. Serve the interface | [Go directly](docs/installation-options.md#web-server-go-directly) or [private nginx proxy](docs/installation-options.md#web-server-nginx) |
| 6. Add agents | [Worker and provider sign-in](docs/connection-sign-in.md) |

**Go is the web server.** Node is used for checks; it does not replace the Go API.
The [installation guide](docs/installation.md) explains each command. The
[Debian service guide](docs/deployment/debian.md) covers ongoing operation.
Your [hybrid topology](docs/installation-options.md#hybrid-deployment) keeps nginx
and Docker on Linux, with PostgreSQL on its existing server and the browser on
your laptop. It currently uses a private SSH tunnel, not public-domain hosting.

## Design philosophy

**Give agents room to contribute; keep authority explicit.** Personas shape guidance,
while permissions determine actions. Job titles and organization charts do not grant
access. Free-edition disagreement requires human direction; automated consensus is
a planned paid capability.

**Keep a record you can inspect.** Decisions, task changes and evidence belong in
append-only history with attribution. Model assertions and human-supplied evidence
are distinct from independently observed execution. Audit protections have limits
against host/database administrators; see [security](SECURITY.md).

**Separate conversations, execution and credentials.** The browser handles interaction,
the Go controller authenticates and coordinates, PostgreSQL retains events, and an
approved Linux worker enforces container execution boundaries. Engine credentials
stay separate from tool workspaces. There is no uncontained host fallback or
isolation-disable switch. See [architecture](docs/architecture.md).

**Make changes understandable.** Domain types, HTTP coordination, execution contracts
and browser components have separate responsibilities. Small interfaces and focused
guides help contributors find the boundary they need without learning everything
at once. See [server structure](docs/server-structure.md).

## More thorough documentation

| You want to… | Read |
| --- | --- |
| Learn rooms, agents, tasks and memory | [Product guide](docs/concepts.md) |
| Configure or troubleshoot the system | [Operations](docs/operations.md) |
| Understand current gaps | [Release readiness](docs/release-readiness.md) |
| Read intended behavior | [Functional specification](FUNCTIONAL_SPEC.md) |
| Review security limitations | [Current security review](SECURITY_REVIEW.md) |
| Find any technical guide | [Documentation index](docs/README.md) |
| Review the previous README | [Timestamped copy](README_20261016_115423.md) |

Help is also available inside the application and at `/help/` on your installation.
Provider icons retain [their license notice](web/icons/LICENSE.lobe-icons); names
and logos identify integrations and do not imply endorsement.

## Every file, type and method

The [repository reference](docs/reference/README.md) inventories source and support
files, grouped by directory. Each file has a purpose and a source link. Go types,
fields, interfaces, functions and receiver methods have declarations, available
contract comments and direct-call references. JavaScript named declarations are
indexed alongside source links; callbacks and dynamic methods remain in source.
Go does not use classes in the Java/C# sense.

This exhaustive source map lives outside the README to keep installation readable.
Missing prose contracts are marked explicitly. Private runtime data, credentials,
Git internals, separate worktrees and generated artifacts are explained by category
without exposing their contents. Regenerate the reference after source changes:

```sh
go run ./tools/source-reference
```

## Extending functionality

Add a provider, workflow, tool or interface when your team needs a capability the
current harness lacks. Start at the matching boundary rather than threading a
special case through the whole application. The [extension guide](docs/extending.md)
maps each kind of change to its code, contracts and tests. MCP, skills and a runtime
plugin loader are future work; developer-assistant skills are not application plugins.

```sh
make test
make check
make build
```

Follow [CONTRIBUTING](CONTRIBUTING.md) for branch work, synthetic fixtures,
documentation and enforced feature flags. Tests mock database interactions and
never launch real provider agents. Mandatory isolation and audit safeguards cannot
be disabled. P0 acceptance scenarios remain pending; green component tests do not
establish live readiness. The owner signs commits manually.

Copyright © 2026 Agent CommonRoom contributors.

Internal worker development: [bounded isolated Go commands](docs/isolated-commands.md)
are opt-in and require a separate operation grant. Production repository dispatch
and live runtime acceptance remain gated.

[Owner-reviewed workspace disposal](docs/workspace-disposal.md) is off by default.
It requires a one-minute review and explicit confirmation. Verified removal releases
only an exactly bound workspace reservation; legacy unbound data stays charged.
Physical quotas and live production acceptance remain pending.

The [internal idle shutdown guard](docs/worker-idle-shutdown.md) protects unused
paired workers, retaining source. Owners can set the idle interval and on/off policy in Settings for new allocations.
Controller-bound idle stops now require audited acknowledgment and can show fixed room notices.
Idle acknowledgments use the scoped private audit link. Production launch/policy
assembly and live acceptance remain pending.

The [reviewed harness release admission](docs/harness-release-admission.md) can pin new internal Codex allocations to a signed, operator-reviewed image. It is disabled by default; upstream discovery, image publication and live upgrade acceptance remain P0 work.

Administrators can use the [offline harness review tool](docs/harness-release-admission.md#offline-review-publication-tool) to generate review keys and publish signed release manifests without placing private signing keys on workers.
