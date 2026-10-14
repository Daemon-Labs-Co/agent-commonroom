# Security hardening and operational limits

This update strengthens required security boundaries; it does not make Agent
CommonRoom ready for untrusted public hosting or claim the security backlog is
closed. See [database setup and audit recovery](database-security.md).

## Identity, load and context

Authenticated API actors must be enabled humans. Agent API credentials are
unsupported and rejected, including read requests. Authorization and room access
still apply independently. Audit appends overwrite client-provided `performedBy`
with the authenticated actor ID. Human and agent messages have explicit identity
badges; typed agents named "You" cannot receive human styling. Legacy messages
may have inferred attribution rather than a recorded authenticated identity.
Display names never confer permissions.

Human API mutations have a per-actor token bucket: 20 requests of burst capacity,
one token replenished per second, and a maximum of 4,096 tracked actors. Idle
entries can be reclaimed. Exhaustion returns HTTP 429 and `Retry-After: 1`.
GET/HEAD requests do not consume this budget; login has [separate failure/proof controls](login-availability.md).
The owner's emergency agent-safety endpoint bypasses the mutation budget so load
cannot prevent a stop. All other permissions remain enforced. Budgets apply
across rooms/sessions within one controller, not across multiple controllers.

Chat prompts have a 128 KiB assembled context budget. Profile instructions,
reference history, scoped shared memory, recent room messages and the request
count toward it. Oversized broadcasts fail with HTTP 413 before any member is
launched or a message is appended. Reduce the request or reference context rather
than silently dropping it. Shared memory is explicitly identified as untrusted
reference material. Incident/emergency stops also block native conversation
listing/import before provider access. These controls are mandatory, with no
feature flag or off switch.

## Worker startup and emergency resume

With the supported remote worker enabled, controller startup checks the approved
worker's `idle-v1` state. Active or unknown state latches an application stop and
records the reason; it does not kill, adopt or replay unknown jobs. A failed stop
append also latches an integrity stop. The owner must inspect worker jobs and
explicitly resume after verified shutdown. A temporary SSH failure can therefore
require owner recovery even if no job is running; fail-closed is intentional.

Emergency resume requires a freshly valid audit chain, no integrity incident,
no active controller turns and a fresh shared-worker idle confirmation. Unknown
worker state leaves the stop in place. Application-wide stop cancels workflow
contexts directly even during a database read outage. It still does not guarantee
remote termination when SSH/cleanup fails; inspect background work and worker
jobs. These mandatory checks have no flag and do not replace durable job leases
or reconciliation after transport loss. Local Linux-worker startup reconciliation
has not been validated; the live deployment uses the approved Debian worker.

## Dependency checks

Updated pgx and Go text dependencies address the reachable advisories found by
Go's vulnerability scanner. `go.mod` requires Go 1.26 or later. The verification
used Go 1.27.2 and found no remaining reported reachable Go vulnerabilities at
scan time; this does not certify provider images, npm packages or OS components.
Run an official `govulncheck ./...` separately from the fast offline test suite,
and scan reviewed container images before deployment. Source secret scans use
redacted output and exclude private ignored state; absence of findings is not a
guarantee that every secret has been detected.

## Deployment and remaining work

Restarting is not sufficient to enable this update: provision a restricted
runtime database role first. Migrations are now explicit (`--migrate` with
`MIGRATION_DATABASE_URL`); startup does not perform schema DDL. Verify idle remote
workers before controller/worker replacement. Do not copy credentials into test
fixtures or disable container isolation to diagnose failures.

Remaining work includes independently protected audit anchoring, durable job
leases and orphan reconciliation after controller failure, confirmed cancellation
under transport loss, multi-controller quotas, spend/storage limits, disclosure
policies for untrusted references, and provider/container dependency audits.
Multiple companies and public remote deployment are not supported isolation
boundaries. Authentication, loopback restrictions, container isolation, audit
checks and retention cannot be bypassed by feature flags.

Automated coverage uses only synthetic records, fake provider callbacks and
mocked database interactions. It covers tampering, metadata changes, outage
cancellation, restart stop markers, explicit owner recovery and failure paths,
privilege refusals, context limits, rate budgets and message identity. Live
PostgreSQL role configuration and provider execution require separate deployment
verification; automated tests must never insert data in the real database.

## Private file reads

See [private file safety](private-file-security.md) for mandatory descriptor-based
validation of credentials, keys, sessions, worker configuration, job records and
lease files. These checks do not establish host-administrator resistance.

Valid proofs are no longer denied solely because earlier failed attempts exhausted
the shared response budget. Session replacement is capacity-aware and rolls back
on persistence failure. See [login availability](login-availability.md) for limits.

## Audit metadata format barrier

Legacy records may appear only before metadata-protected records in the retained
chain. Once a version-1 `_auditRecord` is seen, every subsequent record must retain
valid metadata. A later unprotected record fails verification even when its hash
and previous-hash link are otherwise consistent. The existing integrity stop
cancels active jobs and refuses owner recovery until the chain is valid; there is
no off switch and no automatic history repair or rewrite.

Historical prefix bytes are unchanged. `GET /api/integrity` continues reporting
protected and legacy counts; invalid results latch execution off. New writes
already carry version-1 metadata. Installations should review unexpected downgrades
as integrity failures, retain the records and investigate rather than delete them.

This barrier does not detect removal of the entire protected suffix or a fully
rewritten/all-legacy chain. Independent sequence/head anchoring and an externally
protected archive remain required for rollback resistance. This does not make
legacy identity/time metadata retroactively protected.

Mocked regressions use hash-consistent synthetic downgrades, verify cancellation
and denied recovery, and preserve an unchanged legitimate legacy prefix. Fixture
seeding mirrors the format transition after protected writes; `seedLegacy` is
reserved for explicit compatibility/adversarial tests. No real database history
is edited or inserted by tests.

## Bounded storage waits

State replay, event transactions and integrity verification cap cooperative
storage waits at five seconds. This prevents unbounded database waits while
holding the controller lock. See [storage deadlines](storage-deadlines.md) for
emergency-stop behavior and limits.

## Emergency-stop audit outages

If recording a stop/resume operation fails, the controller now latches the existing
integrity stop marker and cancels all agent/workflow contexts. This applies even
when the original request targeted one room: loss of required audit persistence
escalates to application-wide execution safety. Storage restoration and ordinary
Resume cannot clear it. The marker survives restart on a correctly configured,
writable private controller directory; explicit owner integrity recovery still
requires a valid chain and verified worker shutdown.

Marker persistence failure remains an operational blocker: the controller logs a
fixed warning and stays stopped in memory, but cannot guarantee restart persistence
on unavailable storage. Do not restart until repaired and shutdown is verified.
No stop record is forged after a failed database transaction, and no history is
rewritten. These mandatory safeguards have no off switch. Mocked begin/insert/
commit failures with temporary markers cover both room/application stops, job
cancellation, restart restoration and refusal of ordinary resume. Real DB/worker
outage verification is not claimed.

## HTTP transport limits

The local HTTP listener caps request headers at 16 KiB, with net/http's parser
lookahead allowance. Oversized headers receive HTTP 431 before authentication or
application handlers run. Header reads have a five-second timeout; total reads
have fifteen seconds; writes have fifty seconds; idle keep-alive connections have
an explicit fifteen-second timeout. These required limits have no feature flag.
Do not enlarge them as a workaround for oversized credentials: login proofs
belong in the bounded request body. Authentication, authorization and body limits
remain independent. A real ephemeral loopback-listener test verifies rejection,
no handler invocation and successful requests after rejection. It uses no DB or
provider. These limits do not solve volumetric attacks, handler CPU exhaustion or
all long-running handler cancellation; public hosting is still unsupported.

## Open: credentials shared with development tools

Managed development currently mounts provider authentication into the same
container/UID as the provider's development tools. Read-only mounting protects
against modification, not reading. A malicious repository or compromised model
can request a tool read and return credentials through normal model output.
Connection-specific storage and provider-only network egress reduce scope but do
not close this disclosure channel. Do not treat untrusted repositories as safe.

A proper remediation must separate credential-bearing provider execution from
credential-free tool execution, with an authenticated, capability-limited tool
broker and isolation for both. Built-in tools that can inspect the provider's
filesystem must be disabled or independently confined. Prompt rules, output
redaction and a second read-only mount do not establish that boundary. This is
OPEN, not a completed fix. No credentials were read during this source review.
See [OpenAI sandbox security](https://developers.openai.com/api/docs/guides/agents-api/environments/security)
for the external-credential broker model; that API example is guidance, not proof
that the current CLI adapter already supports this separation.

## Codex chat command-tool policy

Fresh and resumed Codex chat turns, workflow discussion turns without an approved
development repository, and connection reply verification explicitly pass
`-c features.shell_tool=false` in addition to the read-only provider sandbox.
Read-only permits reads, so it was insufficient to prevent command-based reads
of mounted provider credentials. Claude chat retains its empty tools allowlist.
This mandatory command-tool restriction has no product flag or user override.
Worker validation accepts only the exact policy; removing, changing or appending
a command-tool override is rejected before launch. Login remains a separate
administrative policy. Existing transcripts are preserved.

Update controller and worker binaries together: the new worker rejects the old
command-enabled chat prefix rather than silently accepting it. Use the trusted
pinned provider image supporting the documented Codex setting. Mocked policy
tests verify fresh/resumed requests and tampered options, with no real provider
launch. Effective tool availability in the deployed pinned CLI has not been
verified live. This narrows command-tool exposure in chat; it does not establish
a complete provider/filesystem separation or fix development credentials shared
with Bash/Read/etc. Development remains the open architecture finding above.

Reference: [Codex configuration reference](https://learn.chatgpt.com/docs/config-file/config-reference)
(`features.shell_tool` controls command execution tools).

## No implicit owner identity

Protected request identity must be supplied by authenticated actor middleware.
A missing or malformed context identity resolves to no actor, never the owner.
Sessions with an empty actor ID are rejected and removed from runtime session
maps; raw and hashed-cookie lookup paths behave identically. Valid owner/invited
human sessions retain their explicit identity. These mandatory checks have no
feature flag, and the emergency-stop budget bypass still requires a valid owner
session. An actor-less legacy/malformed session must sign in again rather than
being promoted through compatibility fallback. Persisted session-store validation
continues to reject malformed identity records at startup.

The owner-key download handler independently requires explicit owner context,
in addition to its CSRF/setup checks and the outer middleware. Administrative
handlers reject absent identity before DB/provider access. Route authors must
still wire both authentication and actor authorization: this defense does not
turn an unprotected ordinary handler into a protected route automatically. Never
construct owner context from caller-provided headers or body fields.

Synthetic tests exercise missing/malformed context, owner-only handlers with nil
DB/provider dependencies, actor-less raw/hashed sessions and the normal synthetic
owner-key lifecycle. No live account keys, sessions or providers were accessed.

The [provider/tool separation design](provider-tool-separation.md) defines the required free/core two-container broker, lifecycle migration and evidence gates. It is a plan, not a completed boundary.


### Broker dispatcher foundation

The unused `internal/toolbroker` dispatcher now enforces mock-tested audit-before-
execution ordering, shared single-job admission, original deadline/cancellation,
16 KiB output limits and no automatic replay. Audit failure fences authority.
This is internal infrastructure with no toggle or deployed execution endpoint;
mandatory security boundaries remain free/core. Concrete durable persistence,
credential-free execution, transport peer authentication and verified pair cleanup
are still missing. See [dispatcher contract](provider-tool-separation.md#injected-dispatcher-safety-contract-not-connected-to-agents).

### Pre-header provider gateway connection cap

Provider gateways now admit at most 16 live clients before HTTP parsing, in
addition to the existing 8 active CONNECT slots. Idle/pre-header clients cannot
create unbounded HTTP connection state in this gateway. Slots release once on
close, including shutdown and double cleanup. This is mandatory free/core
infrastructure with no off switch. Synthetic/local Unix tests pass without
provider/DNS/DB access; volumetric CPU/socket pressure and credential separation
remain open. See [gateway admission](container-isolation.md#gateway-connection-admission).


### Tool-container launch policy foundation

Unused broker infrastructure now derives a separate tool workspace/container
from trusted job authority, with no credential, host or socket mounts; network
is disabled and resource limits are fixed. Expired/revoked plans are refused.
Synthetic argument tests prove the policy construction, not image contents,
volume freshness, atomic launch, actual Docker containment or provider separation.
No runtime is wired or deployed; the existing credential exposure remains OPEN.
See [tool plan contract](provider-tool-separation.md#tool-container-launch-policy-foundation).


### Tool runtime inspection foundation

The unused tool-plan verifier now refuses unexpected mounts/environment/privilege,
resource/namespace policy changes and local volumes redirected by bind/device
options. Evidence must be fetched by the trusted worker from the approved rootless
daemon; model/plugin-supplied metadata is not authority. Tests use synthetic
inspection objects only. No Docker/provider is launched or production path wired;
actual kernel enforcement, freshness and credential separation remain OPEN.
See [inspection gate](provider-tool-separation.md#tool-container-inspection-gate-foundation).


### Mock-driven container executor adapter

The unused adapter now routes audited typed actions to a fixed helper through a
verified immutable tool-container ID. Cancellation/failure fences authority and
requires successful exact-ID absence checks; daemon errors cannot count as
shutdown proof. All Docker interactions are mocked in tests. No real runner,
helper or pair lifecycle is wired; provider credential separation remains OPEN.
See [adapter contract](provider-tool-separation.md#mock-driven-container-executor-adapter).


### Typed broker allocation budgets

Typed requests/actions now reject oversized strings and aggregate command input
before copying, hashing or serializing; canonical hex checks do not allocate from
arbitrary input. Wire-size limits still apply after JSON escaping. These required
free/core limits have no off switch and do not replace upstream transport budgets.
Synthetic zero-allocation/early-refusal tests pass; broker runtime wiring and
provider credential separation remain OPEN. See [typed budgets](provider-tool-separation.md#typed-input-allocation-budgets).


### Executor operation permission gate

The unused container adapter now independently refuses actions outside its
trusted job's copied tool set, even when called directly by internal code.
Scope, issuance/expiry and revocation must also match before Docker interaction.
This mandatory free/core check cannot replace dispatcher authentication/auditing
or fenced runtime supervision. Synthetic tests cover refusal and permitted reads;
production separation remains OPEN. See [executor permission gate](provider-tool-separation.md#executor-declared-tool-permission-gate).


### Workspace file primitive foundation

Unused tool-only file primitives now bound reads/writes, validate rooted descriptors
and atomically replace entries without truncating an existing inode. Parent sync
or identity uncertainty cannot report successful completion. Synthetic temp-file
and injected-failure tests pass. They do not replace authenticated audited broker
admission, verified container/mount isolation or whole-job supervision. No helper
or runtime path is wired; credential separation remains OPEN. See [workspace contract](provider-tool-separation.md#rooted-workspace-file-primitives-not-wired-to-a-helper).


### Bounded tool file helper foundation

An unused Linux-only helper now accepts strict read/write actions for the future
credential-free tool image, with input/output budgets, rooted file checks and
unconfirmed write responses on delivery failure. Commands, patches and root
selection are refused. Synthetic tests pass; no image, real runner or production
path installs/invokes it. Required free/core safety has no bypass flag; provider
credential separation remains OPEN. See [helper protocol](provider-tool-separation.md#bounded-file-helper-protocol-not-installed-or-wired).


### Fixed tool launcher source

The future tool-image launcher uses a fixed helper/cwd, no shell/caller arguments
and an explicit credential-free child environment, with generic startup failures.
Mocked process tests are included in full checks. No worker image installs it and
no production execution is changed. Container/pair shutdown proof and provider
credential separation remain OPEN; this mandatory free/core boundary has no off
switch. See [launcher contract](provider-tool-separation.md#fixed-tool-launcher-not-installed-in-a-worker-image).


### Tool-only image recipe

The unbuilt tool image has a minimal source context, required immutable Go/Node
base pins, offline Go compilation and no provider/auth setup. Configuration/pin
checks and a minimal-source Linux cross-build pass. No Docker image, ignore-engine,
base provenance, mount/kernel or live provider verification has occurred; production
credential separation remains OPEN. This mandatory free/core boundary has no off
switch. See [recipe and verification gates](provider-tool-separation.md#tool-only-image-recipe-unbuilt-not-deployed).


### Strict tool input encoding

The unused broker/helper refuse invalid UTF-8 and unpaired JSON surrogate escapes,
with separate typed-action checks, before JSON replacement can silently alter file
paths/content. Valid Unicode/literal escape tests and zero-allocation verification
pass. Byte/schema/auth/audit checks still apply; required free/core safety has no
bypass flag and production credential separation remains OPEN. See [encoding contract](provider-tool-separation.md#strict-tool-input-encoding).


### Helper result admission

The unused executor now requires a strict read/write receipt with matching
operation, bounded byte count and digest before success; reads also validate
canonical base64 content. Malformed results fence/clean up instead of publishing
success. Unsupported command/patch operations are refused before Docker. Synthetic
mock tests pass. A consistent receipt is not filesystem attestation; image trust,
paired shutdown and production credential separation remain OPEN. Mandatory
free/core checks have no bypass flag. See [result contract](provider-tool-separation.md#helper-result-admission-before-success).
