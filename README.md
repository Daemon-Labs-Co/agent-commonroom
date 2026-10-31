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

The [reviewed harness release admission](docs/harness-release-admission.md) can pin new internal Codex allocations to a signed, operator-reviewed image. It is disabled by default. Engine preflight always requires the exact pinned repository digest in staged-image metadata; upstream discovery, image publication and live upgrade acceptance remain P0 work.

Administrators can use the [offline harness review tool](docs/harness-release-admission.md#offline-review-publication-tool) to generate review keys and publish signed release manifests without placing private signing keys on workers.

Optional [room coordination](docs/governance-appointments.md) is off by default.
Owners may select a coordinator and send explicitly role-dependent requests.
Replacement or disabling invalidates those turns; ordinary human work keeps its
existing permissions. Selection never grants extra access or bypasses isolation.

Optional [conversation task proposals](docs/task-proposals.md) let the owner
capture a suggestion and review its original source before creating a draft.
Off by default; approval never starts execution or grants agent permissions.

[Bounded delegation](docs/task-delegations.md) adds authenticated worker assignment
proposals and human review without inherited extra permissions or automatic
launch. Off by default; automatic provider emission remains unimplemented.


Agents participate in assigned work rooms; unused automatic agent personal rooms
are omitted by default. Existing history and work remain available. The owner
can restore the legacy catalog with `automaticAgentRooms`; see the
[agent-room compatibility contract](docs/concepts.md#automatic-agent-room-compatibility).

Shared rooms can carry a human-defined deliverable and acceptance criteria. Enable the default-off `roomOutcomes` setting to require both for new shared rooms; existing rooms and saved history remain usable. See [room outcomes](docs/room-outcomes.md).

Owners can separately opt in to `roomOutcomeContext` to include a room deliverable as bounded reference in task work/review prompts. It defaults off and respects provider/reference disclosure restrictions; see [task-provider context](docs/room-outcomes.md#optional-task-provider-context).

Owners can enable default-off `roomOutcomeReview` to record human acceptance or reopening of a shared-room deliverable. Acceptance becomes stale when the room or its work changes. See [room deliverable review](docs/room-outcome-review.md).

Structured tasks use a seven-part definition with finite input/output contracts, restrictive limits and a failure policy. Enable default-off `taskDefinitions` to use the editor, per-execution value bindings, immutable returned-value records and paginated saved definition history. File/bundle execution refuses pending verified artifact storage. The internal [custody journal](docs/artifact-custody-journal.md) now persists reservations and receipts but still requires trusted source, physical quota and delivery integrations. The internal [source capture foundation](docs/workspace-source-capture.md) now binds actual bounded workspace bytes; production frozen-volume authority remains required. Internal [retained-volume admission](docs/retained-workspace-observation.md) now holds the creator lock and verifies container absence before observation. See [task definitions](docs/task-definitions.md).

See [accepted task inputs](docs/task-value-inputs.md) for verified same-room value dependencies.

- [Task criterion review](docs/task-acceptance.md): explicit human decisions against exact task outputs and criteria.

- [Recorded task progress](docs/task-progress.md): scoped contract, attempt, input and human-review status without claiming live verification.

The internal [read-only workspace mount probe](docs/readonly-workspace-mount-probe.md) verifies an isolated Linux mount and exact helper cleanup; production export integration remains outstanding.

The internal [durable probe journal](docs/workspace-probe-journal.md) retains lifecycle uncertainty across restarts and counts unresolved probes during worker admission; production controller assembly remains required.

The internal [guarded workspace probe](docs/guarded-workspace-probe.md) now connects creator ownership, exact helper admission and durable cleanup into one operation; production controller transport and file export remain outstanding.

Internal [recorded probe recovery](docs/workspace-probe-recovery.md) reconciles known helper IDs with fresh cleanup authority while preserving retained workspaces and unknown-ID uncertainty.

The internal [account probe ledger](docs/account-probe-capacity.md) now durably counts unresolved workspace helpers alongside managed jobs across workers; live controller transport and physical evidence remain required.

Workspace file-delivery infrastructure includes a bounded private
[probe authority channel](docs/workspace-probe-transport.md). This does not yet
enable file/bundle task outputs; production authority, frozen-source capture,
physical quotas and authenticated delivery remain required.
The [worker-local physical validator](docs/workspace-probe-physical-evidence.md)
now checks container/volume observations behind probe authorization. Deployed
remote evidence routing and actual file delivery remain incomplete.
[Isolated source capture](docs/isolated-source-capture.md) now has a fixed-root,
live-authority lease implementation. The reviewed helper/runtime, physical quotas
and authenticated output delivery still need integration before file tasks work.
[Source-to-custody transfer](docs/source-custody-transfer.md) now joins capture and
private object persistence, preserving uncertain charges on partial failures.
It still requires real quota/runtime/controller assembly before file tasks work.
The [capture helper](docs/capture-helper-channel.md) now has a bounded private
supervisor protocol and separate scratch-image build definition. Runtime/image
acceptance, physical quotas and public output delivery remain incomplete.

Internal [capture-container admission](docs/capture-container-admission.md) now
validates the fixed read-only source/private custody mounts and reviewed helper
policy. Supervised launch, physical quotas and end-to-end file delivery remain
incomplete; this infrastructure does not enable required file-task execution.
The internal [capture supervisor](docs/capture-supervision.md) now orders durable
creation/binding, private protocol execution and exact-ID cleanup before releasing
metadata. Its daemon/stream and physical quota integration remains unfinished.
The [capture worker stream](docs/capture-worker-stream.md) implements the fixed
rootless helper start and private stdin/stdout adapter. It remains unwired to task
dispatch pending physical quota, durable authority and live image acceptance.
[Interrupted capture recovery](docs/capture-recovery.md) can reconcile an authorized
recorded helper ID without restarting capture or deleting retained objects/volumes.
Production journal/creator integration and unknown-ID incident handling remain open.
[Contract-derived capture selection](docs/contract-capture-selection.md) now binds
finite capture members to required task file/bundle contracts, with explicit optional
outputs. This metadata binding and synthetic byte validation do not enable dispatch.
[Controller capture authority](docs/controller-capture-authority.md) now binds source
observations to the current human, task, worker and exact output contract, recording
permission before acknowledgment. Mandatory deployed physical evidence remains open.
[Worker capture evidence](docs/worker-capture-evidence.md) now independently inspects
exact running helpers and retained source/custody identities under creator ownership.
The deployed durable binding/physical-quota witness is still mandatory and missing.
[Durable worker capture journals](docs/worker-capture-journal.md) now retain exact IDs
and unconfirmed helper charges across interruption. Worker admission counts captures;
controller/account transport and physical quota enforcement are still required.
The [controller capture ledger](docs/controller-capture-ledger.md) now records helper
lifecycle and includes captures in account-wide role/probe admission. Deployed
physical/quota evidence and authenticated worker transition assembly remain open.
[Capture lifecycle evidence](docs/capture-lifecycle-evidence.md) now matches local
journal history and independently observes helper state/absence before controller
acknowledgment. Real quota/account witnesses and full dispatch assembly remain open.

The internal [capture authority transport](docs/capture-authority-transport.md)
connects local lifecycle journals to separately authenticated controller guards over
an owned private stream. Uncertain acknowledgments retain charges and stop authority.
It does not establish deployment authentication or provide physical storage quotas;
required file-task dispatch and artifact delivery remain blocked pending composition
and live acceptance.

[Pinned controller capture sessions](docs/controller-capture-session.md) now adapt
private lifecycle streams to current permissions and the durable ledger. Session
identities cannot be selected by message fields, and stopped sessions cannot revive.
Deployment authentication and independently enforced physical quota witnesses remain
required; this internal adapter does not enable file-task dispatch.

The internal [capture operation observer](docs/capture-operation-evidence.md)
checks retained history and exact container/volume references before and after
lifecycle operations. Independent physical quota/account witnesses remain mandatory;
metadata inspection alone does not enable file-task execution or prove live readiness.

The internal [sealed-object read channel](docs/sealed-object-read-channel.md) transfers
one receipt-pinned object over an already authenticated private stream. Receiving
supervisors recheck actual bytes and require durable release authority before returning
them. It supplies no public download route or deployed isolated reader; those and
physical quota/transport assembly still block end-to-end artifact delivery.

[Pinned artifact read sessions](docs/controller-artifact-read-session.md) connect the
private byte channel to controller permissions and durable release records. They
recheck access after evidence callbacks and permanently revoke stopped sessions.
Actual isolated reader deployment, physical quota/account witnesses and public task
artifact delivery remain incomplete; this adds no public download endpoint.

The [read-only source journal relay](docs/sealed-read-journal-relay.md) lets a fixed
isolated reader obtain controller permission and durable source release acknowledgment
without database credentials. Its phase order and exact receipt are enforced on both
peers; deployed reader image/mount/lifecycle and physical quota assembly remain open.

The [fixed isolated object reader](docs/isolated-object-reader.md) now has a credential-free
entrypoint, read-only kernel mount lease and separate image recipe. It rechecks the actual
view around controller authorization and sealed reads. Static Linux compilation passed;
container build/live acceptance, production reader lifecycle assembly and physical
quota enforcement remain incomplete. File-task dispatch remains blocked.

[Object-reader admission](docs/object-reader-admission.md) now verifies exact helper
image/volume identities, strict read-only policy and independent reference inventories.
It observes metadata only; production lifecycle/account admission, actual quotas and
live acceptance still require assembly before task artifact delivery can be enabled.

Supervised capture keeps its attached process stream open until successful exit and
exited-container verification; protocol failure or cancellation still interrupts it.
See the [capture lifecycle contract](docs/capture-supervision.md). This ordering fix
does not enable file dispatch or replace live container acceptance.

[Task repository references](docs/task-repository-links.md) provide a default-off,
API-only list of GitHub issue/branch/commit/PR links per shared-room task. Enabled
humans with room access may read; only the owner may edit. References are not verified
deliverables or repository execution/push authority. The guide is available in Help.

[Supervised isolated reads and recovery](docs/object-reader-supervision.md) now
withhold artifact bytes until successful helper exit, exact-ID cleanup, durable
lifecycle finish and a final permission acknowledgment. Known-ID recovery performs
cleanup only. Production authentication/quota assembly, public delivery and live
acceptance remain incomplete; file-task dispatch stays blocked.

The [durable worker reader journal](docs/worker-object-read-journal.md) records
reader ownership before launch and retains capacity until cleanup is durably
acknowledged. Readers and capture helpers share conservative admission accounting.
This initial adapter requires a retained workspace; authenticated controller transport,
physical quota witnesses and public artifact delivery remain incomplete.

The [controller reader ledger](docs/controller-object-read-ledger.md) now records
account-wide reader capacity and exact-ID cleanup alongside probes/captures.
It requires independent trusted evidence; no deployed transport or artifact
delivery capability is enabled by this infrastructure.

The [reader authority bridge](docs/reader-authority-bridge.md) connects pinned human/worker
sessions, bounded private control frames, local/daemon observations and the isolated
worker runtime. Deployment authentication, real witnesses, verified starter deployment and public delivery
remain incomplete; its assembled tests use synthetic private channels and mocked daemons.

Reader deployment now has [private worker profiles and offline preflight](docs/reader-worker-profiles.md), including fixed endpoint selection, certificate loading and service shutdown ownership. This does not enable production file-task dispatch or certify physical storage/live acceptance.

Managed worker deployment includes [private reader profiles](docs/reader-worker-profiles.md)
with offline certificate/policy validation, and an operator-only
[physical storage preflight](docs/worker-storage-preflight.md) for approved Linux
XFS roots. These checks do not start agents or certify production dispatch:
complete storage inventory, custody/admission integration and live acceptance
remain required. Both refuse missing security evidence without a host fallback.

Use `./bin/commonroom --help` to list operator commands without starting services.
Misspelled commands and extra migration arguments refuse before controller or
database startup; see [explicit startup commands](docs/operations.md#explicit-startup-commands).

Internal [controller manifest assembly](docs/artifact-manifests.md#controller-selected-sealed-manifest-assembly)
now selects sealed file/bundle receipts from the audit-backed custody records,
verifies actual bytes and rechecks current access after all members resolve.
Production file dispatch, isolated custody guards, public export and human
acceptance integration remain required; a manifest is not task completion.

Verified manifests can now be recorded once per execution/output in an internal,
audit-backed immutable registry, with replay checks and no automatic overwrite.
This preserves verification metadata; it does not attach outputs, authorize
future downloads or accept tasks. See the [manifest registry contract](docs/artifact-manifests.md#immutable-controller-registry).

Internal recorded-member reads now pin the reviewed manifest digest, revalidate
all bundle members against actual custody bytes and recheck current export/access
before returning a selected member. Public downloads and production physical
custody assembly remain unfinished; see [recorded-member verification](docs/artifact-manifests.md#private-recorded-member-revalidation).

The internal [owned capture assembly](docs/capture-supervision.md#retained-worker-ownership-assembly)
connects the supervised helper to its retained creator lease and durable worker
journal, joining capture cleanup before releasing ownership. Independent production
authority/quota setup and public task deliverables remain unfinished.

Retained-worker teardown also joins isolated reads and reader recovery before
creator unlock, including failed-operation authority revocation. Reader setup
requires its inspection runner to own the starter, and owner expiry withholds
bytes even after an acknowledged internal read. See the
[reader authority bridge](docs/reader-authority-bridge.md).

The [retained capture recovery assembly](docs/capture-recovery.md#retained-worker-assembly)
reconciles only durably known helper IDs under fresh cleanup authority and creator
ownership. Interrupted cleanup can be explicitly resumed without duplicate history;
it cannot restart capture, publish outputs or release artifact storage charges.

Capture control now has a continuously watched authority lifetime: idle connection
loss cancels capture/recovery and withholds success. The private client joins its
bounded read pump on stop; old acknowledgments cannot revive a closed session.

[Private capture sessions](docs/capture-authentication.md) now bind mutual TLS
identity to the exact worker, retained record and finite capture intent before
controller authorization. TLS 1.2/1.3 are supported with protocol separation from
artifact readers. Internal fixed-endpoint dial/one-shot serving adapters consume
their connections, bound setup time and refuse uncertain setup without retries.
A [private capture service](docs/capture-worker-profiles.md) now loads capture-only
company/worker/role policy and joins setup, serving calls and retained client pumps
on shutdown. `--check-capture-profiles ABSOLUTE_PATH` validates that private policy
and its identities offline before controller startup, without starting services.
Deployment endpoint provisioning and physical custody/quota wiring remain
unfinished; this installs no public capture or file-delivery endpoint.
