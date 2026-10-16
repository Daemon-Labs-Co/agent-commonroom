# P0 execution progress — 15 October 2026

[Functional specification](../FUNCTIONAL_SPEC.md) is the authoritative
owner-confirmed product baseline. The [gap assessment](v0.1-gap-assessment-2026-10-15.md)
records the earlier `405c89e` snapshot; this page records subsequent changes.
P1 implementation is paused until all P0 gates are satisfied.

## P0-1: build and verification baseline — verified

Reconciled provider policy arguments and the isolated launch signature. The
explicit separation input must match reviewed policy arguments and cannot enable
an unapproved development workspace. Production chat adapters do not silently
activate the unimplemented bridge. Both development-policy variants still hit
the mandatory admission barrier before credential/runtime preparation.

Fixed a whitespace-sensitive test to decode authenticated performer metadata
rather than compare JSON formatting. Full `make test` passed in 2.44s;
`make check` passed in 7.66s. Unaffected Go packages were cached. These tests
launch no providers and use no database connections.

## P0-2: one hardened scoped audit path — component verified

Integrated the hardened staged audit source, including fresh per-link nonces,
strict metadata decoding, immutable durable job initiator/company binding,
revalidation on each acknowledgement, unresolved-call fencing and scoped
controller stream assembly. Chat/task background contexts preserve only trusted
initiator identity; workflows use their recorded initiator. Missing identity
never becomes owner identity.

Removed `storedToolAuditor` and its unscoped write assembly. Historical
`tool.call.*` replay remains readable and does not grant current scoped authority.
Worker `newToolDispatcher` requires a fenced matching job and an explicit audit
channel. Its lifetime is bound to worker cancellation/deadline; the supervisor
must call the returned idempotent cleanup. Cancellation revokes authority and
closes blocked I/O even when dispatch is given another context.

Removed the obsolete `toolSeparation` product flag setter and helper. Stored
legacy flag data is not rewritten and grants no execution authority. Isolation
and durable auditing have no disable switch. See [the current audit
contract](tool-audit-callbacks.md) for behavior, metadata, limits and source-level
bypass risks. These changes do not establish authenticated production transport.

After initial integration, full `make test` passed in 2.62s and `make check` in
7.38s; unaffected packages were cached. A subsequent review identified and fixed
the constructor-lifetime cancellation gap with a blocked-I/O regression test.
Final full verification after that fix: `make test` **2.20s** and
`make check` **7.28s**, both passed. Server tests ran; unchanged Go packages
were cached. The complete race/vet/Python/JavaScript check target completed.
Whitespace and documentation link-target checks also passed.

## P0-3: complete paired runtime — open; development remains unavailable

The production execution path is not wired. Do not remove
`requireCredentialSeparatedDevelopment`, deploy a stock provider as the reviewed
engine, reuse old credential-bearing workspaces, or treat the constructor/tests
above as completed runtime separation.

The worker-owned paired journal/cancellation slice is now implemented under the
existing membership/job leases; see [the current contract](worker-pairs.md).
Version-3 records and explicit pair cancellation reject protocol downgrades, and
partial creation/shutdown stays unresolved until both-role inventory and released
supervisor ownership are verified. Controller paired registration now records
version/execution/repository and original human/company provenance, and owner
recovery uses the matching paired cancellation contract. Snapshot mismatches or
partial shutdown cannot settle those jobs. No paired launch endpoint is enabled.

Remaining integration requirements (journal/cancellation portions of items 1/6
are now component-tested):

1. Introduce explicitly versioned paired launch/cancel/status records. Persist
   both immutable container identities and partial-creation state before engine
   execution can begin. Existing single-container settled state is not pair proof.
2. Reserve aggregate resources; bootstrap the approved snapshot into only the
   credential-free tool volume. Refuse existing/unreconciled identities/volumes.
3. Authenticate a supervisor audit substream over the approved SSH relationship,
   separated from model stdout; bind it to original managed-job registration.
4. Drive the pinned callback-only engine interface, verify thread/turn/tool call
   correlation and reject native/customization tool paths. The current Codex
   adapter parses `codex exec`; app-server schema files alone do not implement this.
5. Add bounded isolated command execution needed for build/tests. Current helper
   implements read/write only; unsupported patch/command operations remain refused.
6. Fence calls on failure, stop both exact IDs and prove both absent plus released
   supervisor ownership. Reconcile crash/ack loss without replaying uncertain work.
7. Produce reviewed engine/tool image artifacts and dedicated synthetic rootless
   runtime/provider verification. No existing personal sessions or credentials may
   be repurposed as fixtures. Compilation and metadata checks are not OAuth,
   callback or native-route-refusal verification.

Source seams: [worker command/records](../cmd/server/remote_worker.go),
[job leases/fencing](../cmd/server/worker_lifecycle.go),
[tool plan](../internal/toolbroker/container_plan.go),
[tool executor](../internal/toolbroker/container_executor.go),
[tool helper](../cmd/tool-helper/main.go), [engine pins](../docker/engine/README.md).
The design/acceptance gates remain in [provider/tool separation](provider-tool-separation.md).

No live provider, Docker deployment, credential inspection or database access was
performed for these changes. P0-3 remains a release blocker; P1 has not resumed.


## Paired lifecycle continuation verification

Implemented version-3 paired journals, creation intents and immutable IDs,
lease-fenced dual shutdown, partial-create discovery, explicit paired cancellation
and downgrade refusal. Added controller paired registration, immutable repository
and execution binding, matching owner recovery, original-repository audit checks
and initiating-human attribution for lifecycle updates. The legacy launch path
rejects paired registration.

Final full `make test`: **3.04s**. Final full `make check`: **7.69s**. Server and
room-domain tests ran; unchanged Go packages were cached. Full race/vet/Python/
JavaScript checks completed. New tests use synthetic temporary worker journals,
mocked container/database interactions and the real owner recovery handler.
Source review reported no actionable defect in these lifecycle paths.
No live containers/providers, deployment or database access were used.
P0-3 remains open for pair startup/bootstrap, authenticated callback transport,
callback-only provider drivers, reviewed artifacts and live verification.
P1 and secrets-manager implementation remain paused.

## Architecture deepening and aggregate resource gate

Implemented all three selected architecture recommendations:

- Paired lifecycle owns create/verify/durable binding/start ordering through one
  trusted internal launch adapter. Both IDs and resource observations precede
  either start. Cancellation, journal failure and uncertain creation refuse replay.
- Managed job lifecycle owns registration, submission, monitoring, bounded
  idempotent shutdown and durable recovery status. Endpoint approval, request
  authentication and independent worker verification remain enforced.
- Shared immutable-ID shutdown mechanics serve tool failure and paired recovery.
  Pair recovery still separately checks role/name/job inventory and creator leases;
  one tool's absence never settles a job.

Continued P0 with mandatory ID-bound aggregate resource inspection before start:
maximum 512 MiB, 0.5 CPU and 64 processes per role, unchanged total job ceiling,
positive limits and no extra swap. No product toggle or authority bypass exists.

Full `make test` passed in **2.92s**, `make check` in **6.36s**. Server tests ran;
unchanged packages and the already-tested toolbroker package were cached in the
final normal suite. Race, vet, Python and JavaScript checks passed. Tests use
mocked transports/Docker/database interactions and temporary journals only.
An initial sandboxed run could not access the Go cache; an intermediate run
caught a synthetic inspection fixture shape mismatch, corrected before final
verification. No live containers/providers, deployment or database access occurred.

The internal launch adapter has no production implementation. Fresh workspace
bootstrap, reviewed engine configuration/images, authenticated audit transport,
callback driver, bounded commands/artifacts and live isolated verification remain
open P0 requirements. Development admission remains blocked and P1 remains paused.
See [managed lifecycle](managed-job-lifecycle.md), [pair lifecycle](worker-pairs.md)
and [resource policy](provider-tool-separation.md#aggregate-pair-resources).

Final review also fixed Close retaining supervisor leases until in-flight create/start adapter work returns. Blocking create/start regression tests cover concurrent Close and eventual lease release. The full suites above were rerun after this fix.

## Constrained tool launch implementation

Continued P0 with `ToolContainerRuntime`, the concrete tool-side Docker
create/inspect/start implementation. The plan now has stopped-container create
arguments without automatic removal. Runtime creation refuses existing names,
unavailable inventory and missing/unsafe prepared volumes before issuing create.
Full immutable IDs and the entire stopped-container policy are checked; only a
fresh `created` state qualifies. Startup repeats policy checks before using the
immutable ID and verifies running evidence afterward. Creation/start uncertainty
is never retried, and failed creation does not grant later startup authority.
Original job cancellation and authority revocation cannot be revived with a new
call context. Created/running checks share the full inspection implementation.

This module is not yet assembled into a production pair adapter. It does not
initialize a volume or attest its contents: the future worker bootstrap must
establish a fresh credential-free volume and preserve provenance before use.
The controller admission barrier remains unchanged. Reviewed engine driver/image,
bootstrap, authenticated audit transport, full pair wiring, bounded commands,
artifacts and live runtime acceptance remain open P0 work. P1 remains paused.

Final full `make test`: **3.29s**. Final full `make check`: **8.66s**. Server,
tool-helper and toolbroker normal tests ran; room/strictjson/toolworkspace packages
were cached. Race/vet/Python/JavaScript checks passed. An initial focused test
exposed a historical-time fixture, corrected to live synthetic authority; a
revocation test then led to explicit synchronous cancelled-context refusal.
A sandboxed race attempt could not write Go cache; final suites completed with
cache access. No live provider, Docker workload, database or deployment was used.
See [the runtime contract](tool-container-runtime.md).

## Credential-free snapshot bootstrap

Implemented a streamed bootstrap helper inside the tool container. It requires an
empty workspace, checks the original bounded bundle's SHA-256 before controlled
Git, bounds source-tree metadata/files/bytes, rejects symlinks/submodules and
unsafe paths, suppresses hooks/inherited settings and checkout transformations,
and creates a verified job-specific feature branch. Raw source and Git diagnostics
are not persisted or returned as receipts. Both source and Git metadata must be
writable before success.

The tool runtime streams the worker-staged bundle by immutable container ID,
closes blocked input on cancellation, validates a strict job/digest/commit/branch
receipt and rechecks full runtime policy. Only that plan becomes tool-ready.
Unready plans cannot run file actions. Pair startup now starts the idle tool,
requires its bootstrap receipt and only then starts the engine. Missing/forged
receipts and uncertain bootstrap stop progression without replay. The tool image
recipe includes the helper and requires a reviewed base with fixed Git/Node paths;
the build context remains deny-all except explicit reviewed source.

Final full `make test` passed in **3.43s**; final full `make check` passed in
**8.31s**. Server and broker normal tests ran; previously tested bootstrap/helper and
unchanged domain/workspace packages were cached. Race, vet, Python and
JavaScript checks passed. All Git/Docker/database interactions are mocked;
filesystem fixtures use temporary directories. No live Git clone, provider,
Docker workload, database or deployment was used. Initial focused tests exposed
an updated plan initializer and test fixture edit; both were corrected before
final verification. New guides are linked from README/index and in-app Help.

This is internal implementation, not release closure. Fresh-volume allocation
and cleanup, production paired adapter/dispatcher assembly, authenticated audit
transport, reviewed engine/images, bounded commands/artifacts and real
Git/image/rootless verification remain P0 blockers. The older dispatcher assembly
creates an unready plan and cannot enable file tools until production assembly
shares the bootstrap-verified plan. Development admission stays blocked and P1
remains paused. See [bootstrap contract](workspace-bootstrap.md).

Final review bound workspace readiness to its original job lifetime. Fresh call contexts cannot revive tools after job cancellation, and cancellation during an action suppresses its success output. Both regressions pass in the final full suites above.

## Continuous P0 work: workspace volume lifecycle

Implemented fresh local-volume allocation with journal-before-create nonce intent,
daemon creation-time binding, fixed options and refusal of existing/unknown
identity. New starts now require lifecycle version 4 and execution
`separated-pair-workspace-v1`; historical version-3 pairs are recovery-only.
Version-4 cancellation/acknowledgement cannot be downgraded to container-only
proof. Controller registration, snapshots, recovery and scoped audit understand
the new immutable execution/repository binding.

Cleanup waits for creator ownership, validates volume identity twice, refuses
live references and proves name/job-workspace absence. Pending creation after a
lost reply can be discovered by its journalled nonce. Failed or partial cleanup
stays unresolved and counts against admission. Named-volume deletion still trusts
the daemon/administrator because Docker has no immutable volume deletion ID.

Workspace deletion is not implied by ordinary stop. The trusted worker's explicit
`discardWorkspace` authority defaults false and cannot be attached to older
protocols. The HTTP recovery path does not grant discard; present workspaces stay
retained/unresolved until a separate approved-discard or artifact-retention flow
is integrated. No owner/UI role grants an implicit destructive bypass.

Focused normal and race regressions passed for allocation, replacement/in-use
refusal, crash discovery, held leases, mixed versions and owner recovery with
optional diagnostics disabled. Full-suite timings are recorded below after final
verification. All Docker/database interactions are mocked and journals are in
temporary directories. No live volumes, credentials, provider or database were
accessed. Production assembly, reviewed engine/images, authenticated audit
transport, bounded commands/artifacts, approved discard workflow and live runtime
verification remain P0 work. See [workspace volumes](workspace-volumes.md).

Workspace-volume checkpoint: full `make test` **3.27s**, full `make check`
**8.48s**, both passed. Server normal tests ran; previously tested broker and
unchanged packages were cached. Full race/static/UI checks passed.

## Shared audited tool session and timed-run closeout

Implemented one tool-session constructor that shares authority, plan, runtime,
bootstrap readiness and dispatcher. Durable intent/outcome acknowledgement remains
mandatory. Runtime readiness now pins the immutable bootstrap container ID as
well as the original job lifetime; matching name/labels/settings on a replacement
cannot inherit execution. Audited integration tests exercise fresh allocation,
creation, bootstrap and file dispatch through real framed local audit channels
with mocked Docker operations. Intent failure prevents execution, outcome failure
suppresses output, and closing the session revokes authority and interrupts I/O.
The construction-failure path also closes its owned audit link.

Shared-session checkpoint full `make test`: **3.16s**; full `make check`:
**8.92s**, both passed. Final closeout verification: full `make test` **0.69s**;
full `make check` **3.19s**, both passed. All Go test packages were cached in the
final normal suite; complete JavaScript/Python/static checks still ran. No live
provider, Git clone, Docker workload, database or deployment was used.

The final continuation arrived after the requested 16:00 Toronto cutoff; the
clock read 23:17 UTC (19:17 Toronto). No further feature work was started in that
continuation; only final verification and reporting were performed. Work remains
on `codex/p0-worker-runtime-continuation`, based on merge ef04f00. Bootstrap source
files left untracked by the prior merge were preserved with this change set.

P0 is not complete. Remaining blockers include authenticated supervisor audit
transport, reviewed callback-only engine/production pair assembly, bounded command
execution, artifacts and observed-check delivery, approved workspace discard/
retention workflow, disk quotas and dedicated live Git/image/rootless/provider
verification. The development admission barrier remains enforced. P1 was not
started, and the user's condition for committing after clearing P0 was not met.
See [shared audited session](audited-tool-session.md) for the current contract.

## Owner scope additions and physical-container admission

The owner added Gemini and GitHub Copilot to P0, reviewed harness upgrades on
instantiation, configurable five-minute idle shutdown, fresh containers for every
execution, a hard free-account 5 GB storage ceiling and four-container ceiling.
These are recorded in FR-29–33. Paid quotas remain undecided. Harness upgrades
must preserve reviewed immutable images and fail closed on incompatibility.

Worker admission now reserves physical slots from durable unresolved records,
including creation intents with no containers yet. Each pair reserves two slots;
the worker refuses allocation beyond four. Regression coverage includes mixed
legacy/paired jobs, exact capacity, overflow and terminal labels without shutdown
proof. This is only a worker-local guard: aggregate account reservations across
workers, hard storage quotas, idle lifecycle and upgrade integration remain P0.
The development admission barrier remains in place. No provider execution,
container workload or database is used by the tests.

Verification: `make test` and `make check` both passed after permitting the test
runner's local sockets and Go build cache. Initial restricted runs could not bind
local test sockets/access cache; these were environment restrictions, not skipped
tests. Server tests ran uncached; unchanged packages reused cached results.
No live Docker/provider/database validation was performed. No commit was created.

## Private audit transport implementation

Added a dedicated host-key-verified SSH controller link and constrained
`audit-link-v1` worker relay. The original registered connection, initiating
human/company, version-4 job, repository and deadline bind the link. A supervisor
opens a private Unix socket with a durable one-shot marker; failed handshakes,
closure and crashes cannot reset the audit session. The socket is never a model
mount. Original lifetime cancellation closes accepted links even with fresh call
contexts. Relays close owned pipes and join their copies; the dedicated controller
process is cancelled and reaped. EOF or uncertainty is never automatically retried.

Focused normal and race tests passed using real local sockets, framed synthetic
audit acknowledgements, temporary journals, a synthetic child process and mocked
persistence. Production pair-start orchestration and live SSH/provider/container
acceptance are still pending; this checkpoint does not close that P0 blocker.
See [transport contract](worker-audit-transport.md).

Final transport verification: full `make test` and full `make check` passed.
Server normal tests ran in 2.471s; unchanged packages were cached. The full race,
vet, Python, JavaScript and Help checks passed. Added explicit accepted-link close
and unattached-listener deadline regressions. The complete P0 goal remains active;
no P1 work, deployment, commit or push was performed.

## Account-wide physical container reservations

Extended mandatory controller registration to reserve at most four physical
container slots across rooms and workers sharing the account event journal.
Paired jobs reserve two; single-container jobs reserve one. Expired or
shutdown-unconfirmed records retain reservations, and unattributed legacy jobs
count conservatively. Invalid or duplicate identity fails closed. Allocation is
committed with the loaded event count under the existing transactional audit lock,
so stale controllers cannot both acquire the last slot. The worker-local guard
uses the same ceiling. Neither optional diagnostics nor background-work flags
can turn this off; paid quotas remain undecided.

Focused tests cover exact/overflow capacity, mixed execution versions, distinct
workers, terminal proof, restart, disabled flags, concurrent registrations and
stale-controller compare-and-append rejection, with mocked database interactions.
Full make test passed (server 2.335s; unchanged packages cached). Storage quotas,
production pair assembly, providers/upgrades/idle, retention/artifacts and live
runtime acceptance remain open. See [container quotas](container-quotas.md).

Final account-quota verification: full `make check` passed, including the race
suite, vet, Python, JavaScript and in-app Help checks. No live database, provider,
SSH or Docker workload was used. No commit or push was created. P0 stays active.

## Pair launch assembly with the original audited runtime

Added a one-shot supervisor assembly method binding the exact private session
scope and original tool runtime to a separate reviewed engine-launch interface.
The adapter routes fresh volume preparation, role creation/inspection/start and
bootstrap through existing supervisor ordering. Unknown/foreign identities are
refused. Ambiguous create/start/bootstrap attempts cannot be replayed, and only
a successfully validated bootstrap permits engine startup. Supervisor cancellation
closes an unconsumed bundle once and revokes the shared session; shutdown proof
and quota release still require existing explicit reconciliation.

Tests cover synthetic full supervisor startup through the combined adapter,
failed/mismatched bootstrap, foreign role/repository/IDs, fresh-context cancellation,
uncertain-operation replay refusal, exact real ToolSession runtime binding,
one-shot supervisor assembly and accepted/refused bundle ownership. Current
jobs already end after work; there is no idle container pool to configure yet.
Production engine/provider assembly and the owner-configurable idle policy remain
open rather than adding a timer unrelated to actual work ownership. See
[worker pair assembly](worker-pair-assembly.md). No P1 work was started.

Final assembly verification: `make test` and `make check` passed. Server normal
tests ran in 2.448s and race tests in 3.801s; unchanged packages were cached.
The complete vet, Python, JavaScript and Help checks passed. Regression coverage
also rejects a separately constructed audited session with a different connection
and verifies that refused assembly preserves caller ownership. No live engine,
provider, Docker workload, SSH or database was used. No commit or push was made.
The P0 goal remains active and production development admission remains gated.

## Supervisor-owned audit session and idle disconnect detection

The supervisor now owns its audit listener and accepts a ToolSession only through
that exact listener and registered live scope. Attachment waits outside the
supervisor mutex, consumes the private handshake, and constructs the original
shared runtime/dispatcher with approved image/root/runner policy. A single read
pump forwards bounded chunks through an unbuffered pipe to the existing audit
client, preserving its exclusive frame parsing while observing idle EOF.

Disconnect, failed session construction, cancellation and explicit closure revoke
the original supervisor lifetime. The pump and owned I/O are joined on close.
No replay, automatic lease release, fabricated shutdown acknowledgement or
workspace discard is introduced. Production worker execution must still fence
both containers after the session ends; this is not live runtime acceptance.

Tests cover framed acknowledgements, idle link loss, blocked forwarding,
foreign-listener refusal, attachment cancellation, invalid image construction and
startup cancellation through the actual ToolContainerRuntime with mocked blocked
Docker I/O. Pending workspace intent and creator ownership survive disconnect;
engine creation is never reached. Full verification follows below. P0 remains
active, including provider engines, production runner/cleanup, commands/artifacts,
storage quotas, upgrade/idle policies and live verification. P1 remains paused.

Final audit-session verification: full `make test` passed (server 2.655s) and full
`make check` passed (server race tests 3.967s), including vet, Python, JavaScript
and in-app Help checks. Unchanged packages were cached. Only synthetic local
sockets, private temporary journals and mocked Docker/database boundaries were
used. No real SSH, provider, container workload, database, commit or push occurred.
The complete P0 goal stays active; cancellation is not claimed as verified live
container shutdown or full production provider integration.

## Unconditional paired-turn shutdown and active-work fencing

Added supervisor Shutdown using the existing tombstone, creator lease and
immutable-role removal/absence proof. It retains present workspaces and never
grants discard authority. Added Run to defer that reconciliation around startup
and trusted engine turn execution, preserving startup/turn/cleanup errors and
running cleanup even during panic propagation. Original audit loss remains an
error even before asynchronous callback cancellation reaches the turn context.

Close now retains creator ownership through the active model turn as well as
startup. Concurrent reconciliations on one supervisor are serialized. A focused
concurrency regression initially exposed overlapping cleanup attempts, which was
corrected before final verification. Another initial fixture used a fixed ID
instead of the requested role's ID; the synthetic fixture was corrected.

Tests exercise normal/error/panic return, startup failure, audit cancellation,
partial dual-role cleanup, retained workspaces, invalid cleanup contexts,
idempotent shutdown and concurrent shutdown while active work unwinds. All
Docker/database boundaries remain mocked. Production engine/provider runner,
approved discard/artifacts, quotas/upgrades/idle and live runtime acceptance
remain open P0 work. See [paired shutdown](worker-pair-shutdown.md).

Final paired-shutdown verification: full `make test` passed (server 3.074s) and
full `make check` passed, including race, vet, Python, JavaScript and in-app Help
checks. Unchanged packages were cached. No real Docker/provider/SSH/database
was used, and no commit or push was created. The complete P0 goal remains active;
production engine integration and live shutdown proof are still unverified.

## Bounded Codex app-server callback driver

Implemented actual initialize/thread-start/turn-start client flow using the local
reviewed rust-v0.162.1 schemas. One ephemeral thread exposes only commonroom_tool;
actual returned thread/turn identities bind callbacks. Native approval/execution,
unknown methods, wrong scope, duplicate RPC/call IDs, malformed framing and failed
handler/audit outcomes abort without replay. Bounded early notifications wait for
returned identity. Only final agent text is returned, never reasoning content or
claims of observed test evidence.

Supervisor CodexTurn now requires the original assembled session after running
pair startup. Trusted scope/capability/call ID are supplied outside model arguments
and the original Dispatcher enforces actual tool permissions and audit. No process
launch, unreviewed CLI fallback, live provider, image readiness or production entry
point is implied. Engine image/attach/gateway/configuration and dedicated live
binary protocol acceptance remain open alongside the rest of P0. See
[callback driver](codex-callback-driver.md). P1 remains paused.

Final callback-driver verification: full `make test` passed (server 3.271s,
callback package 0.322s); full `make check` passed (server race 4.440s, callback
race 1.521s), including vet, Python, JavaScript and Help checks. Unchanged packages
were cached. Added output/UTF-8, policy ambiguity, private argument duplicate-field
and premature-callback regressions. Only synthetic local streams and mocked
persistence/runtime boundaries were used. No provider, real container, SSH or
database was launched; no commit or push occurred. The full P0 goal remains active.

## Fixed credential-bearing engine process bootstrap

Added a reviewed-source engine entrypoint and deny-all-context image recipe.
Startup requires Linux UID 1000, a fresh exclusively created private home and a
private bounded mapped-UID subscription credential. API credentials/fallback,
extra arguments and existing config/home discovery are refused. Child environment
is explicit; hooks/plugins/browser/host skill discovery and system proxy fallback
are suppressed using options checked in the local pinned client source. The only
process is the fixed reviewed app-server executable, over inherited protocol stdio
and the loopback-to-approved-Unix-socket gateway. Errors are generic and credential
bytes are never printed. Signals forward and gateway resources close on child exit.

The recipe requires immutable source/base references and binary SHA-256, copies
upstream notices, and downloads/installs no CLI. This does not establish provenance
or a deployable image. Synthetic filesystem/process/gateway tests plus recipe
allowlist tests pass. One initial test selected the wrong trailing-argument slice;
that expectation was corrected. Make check includes these tests going forward.
Actual container create/inspection/stdio attachment, rootless credential UID mapping,
refreshed-token custody/concurrency and live binary/image/provider acceptance are
still required. There is no production launch enabled and P0 remains active.
See [engine bootstrap](codex-engine-bootstrap.md).

Final bootstrap verification: `make test` and `make check` passed. The final race
Go results were cached; synthetic engine Node tests, vet, Python, JavaScript and
Help checks ran. Provider stderr is discarded, with only generic wrapper errors.
No real image build, provider, credential, gateway connection, Docker workload,
SSH or database was used. No commit or push was made. The full P0 goal remains
active and no production engine/container compatibility is claimed.

## Reviewed engine-container creation and start policy

Added the DockerRunner-backed Codex engine role with original worker-job binding
and complete engine/session scope equality. One-shot creation refuses an existing
container and preflights image config for unknown environment keys, implicit
volumes/hooks and incorrect entrypoint/user/workdir. The resolved image ID is
retained and checked in every container inspection. The engine recipe now sets
an explicitly empty image Cmd rather than inheriting a base-image command.

Stopped and running inspection checks fixed environment, private namespaces,
resource ceilings, no network/privilege/devices/ports/restarts/auto-removal and
exactly two read-only private credential/gateway mounts, with no source/tool/audit
mount. Both creation and startup are immutable-ID, single-attempt operations.
Policy is rechecked before and after start; changed or unknown metadata fails.
Original lifetime cancellation cannot be bypassed using a fresh call context.

Tests use mocked Docker metadata/calls and cover preflight refusal, policy drift,
unsafe mounts/env/resources, reuse/retry refusal, cancellation and original/full
scope binding. Private host staging/type/ownership, gateway provenance, UID mapping,
credential refresh custody, stdio attachment and live rootless/provider/image
acceptance remain open. No public development endpoint or direct host fallback
was enabled. See [engine runtime](engine-container-runtime.md). P0 stays active.

Final engine-runtime verification: `make test` passed and full `make check`
passed (server race 4.577s, toolbroker race 1.409s; unchanged Go packages cached),
including vet, Python, engine bootstrap JavaScript and Help checks. An earlier
full check exposed a cancellation propagation window: parent cancellation could
be visible before the derived runtime context was cancelled. Runtime admission,
workspace readiness and execution now retain and check the original context as
well as the deadline child. Deterministic regressions simulate that ordering,
and the originally failing authenticated-link-loss race test passed 30 repeated
runs. All checks use synthetic data and mocked runtime/database boundaries; no
real provider, Docker workload, SSH or database was launched. No commit or push
was made. Stdio attachment and private staging remain the next integration work;
the complete P0 goal remains active.

## Engine protocol stream attachment

Added one-shot immutable-ID attachment with full running-policy checks before
and after stream opening. Attachment uses the creation/inspection runner; no
different daemon, restart or reconnect path is offered. The local rootless CLI
runner exposes stdin/stdout only, discards stderr, disables signal proxying and
detach keys, and closes/reaps the attachment process on Close. Original job
cancellation revokes stream I/O. A separately owned stdout pipe preserves the
last buffered protocol response after process exit.

The engine Turn method now connects the journaled engine to the original tool
session and existing callback driver. It must run inside supervisor.Run so all
outcomes receive fenced shutdown. Tests use mocked daemon metadata, local pipes
and a synthetic test subprocess; no provider is launched. Private staging,
credential custody/refresh, production orchestration and live Docker/provider
acceptance remain required before enabling development. See
[engine runtime](engine-container-runtime.md). P0 remains active.

Stream verification: `make test` and final `make check` passed, including race,
static, bootstrap and Help checks. Unchanged Go package results were cached.
The final check includes the original-session scope refusal and final-frame-after-
process-exit regressions. No real Docker/provider/SSH/database was used and no
commit or push was made. This verifies the stream boundaries with synthetic
fixtures, not production credential staging or a live provider turn.

## Worker connection credential lease

Added internal Codex credential preparation using the existing connection lock
shared with sign-in/sign-out. It accepts only a bounded private descriptor-bound
subscription snapshot with exact reviewed schema, duplicate/alias refusal and no
API billing fallback. Snapshot writes require the original live worker context
and matching job. Cancellation does not unlock a potentially running engine;
matching durable evidence must settle both roles before release, independently
of retained source volumes. Failure paths return generic errors and unlock.

Tests use temporary files and synthetic tokens to cover competing engines and
authentication operations, cancellation, foreign/partial fencing, idempotent
release and unsafe file/schema inputs. This is preparation infrastructure, not
production credential admission: mapped-UID staging, crash reconciliation before
reacquisition, refreshed-token persistence and actual provider acceptance remain
open. The filesystem lock is not claimed to survive a process crash. P0 remains
active; see [engine runtime](engine-container-runtime.md).

Credential-lease verification: `make test` passed (server 3.167s), and final
`make check` passed (server race 4.515s), including vet, Python, JavaScript and
Help checks; unchanged Go packages were cached. The final race run also covers
release requiring both created roles absent. No provider, Docker workload, SSH
or database was launched; no real credential was used, committed or exported.
No commit or push was made. Production credential staging remains a release gap.

## Private credential frame replaces host credential staging

Resolved the credential host-UID mapping requirement by removing the host
credential bind mount. The original connection lease now provisions one bounded
versioned frame into the inspected engine attachment before callback initialization.
The wrapper consumes exactly the header and declared payload, then writes its own
exclusive mode-0600 credential file as container UID 1000. It never reads ahead
into app-server protocol bytes. Invalid magic, bounds, short input or provisioning
failure refuses startup; a partial send is never replayed. Creation and inspection
now allow only the provider-socket bind, retaining all other isolation policy.

This supersedes the earlier two-mount/mapped-UID credential-file design, not the
mandatory container boundary. Matching newly reviewed immutable images are
required; old file-mount entrypoints receive no fallback. Source/tool/audit mounts
remain prohibited. Credential bytes never become model input or controller output.
The original connection lease remains held until matching role-fencing evidence.

`make test` passed (server 3.148s, toolbroker 0.561s). Final `make check` passed
(server race 4.622s, toolbroker race 1.422s), including Node short-frame/no-read-ahead
and Go partial-write/no-replay tests, vet, Python and Help checks. Unchanged Go
packages were cached. Only synthetic credentials, pipes and mocked runtimes were
used; no real image, provider, SSH or database was launched. Private gateway
ownership, crash reconciliation, refreshed-token custody, production assembly and
live Docker/private-stdin/provider acceptance remain open. No commits or pushes
were made. The full P0 objective remains active.

## Fresh worker-owned provider gateway

The worker engine constructor now requires an owned gateway handle for its full
job scope, rather than a supplied socket path. Creation uses a fresh private
directory and owned Unix socket; all engine launch/inspection/turn stages verify
the original lifetime, directory/socket identity, permissions and ownership.
The rootless-accessible socket is enclosed by a private unmounted host directory;
no credential file or broader network permission is introduced. Existing provider
destination/address/rate restrictions remain enforced.

Cancellation closes gateway network resources. Cleanup explicitly disables
automatic Unix pathname unlinking and removes only the recorded socket inode and
empty private directory. Replacement files, changed directories and unexpected
content are retained with cleanup errors. Tests use local Unix sockets and a
denied synthetic destination (no external DNS/dial), plus mocked engine runners.
Production assembly/recovery must still own and report this lifecycle, and live
rootless bind/provider acceptance remains open. No public development path was
enabled. P0 remains active; see [engine runtime](engine-container-runtime.md).

Gateway verification: `make test` passed (server 3.264s) and final `make check`
passed (server race 4.587s), including vet, Python, JavaScript and Help checks.
Unchanged Go packages were cached. Regressions cover scope/cancellation, unsafe
roots, replaced socket inodes, changed directory permissions, retained unknown
files, idempotent close and local rejection of an unapproved destination. No live
Docker/provider/SSH/database or real credentials were used. No commit or push
was made; the full P0 objective remains active.

## Owned Codex operation and crash-safe credential admission

Added internal RunCodex ownership across the original audited session, approved
source bundle, connection credential lease, private gateway, paired runtime,
callback driver and fenced shutdown. Validation refusal retains caller resources;
accepted work is one-shot and preparation failures reconcile before returning.
The original supervisor context bounds resource admission and the request context
bounds execution. Bundle close errors remain visible on repeated close calls.

Credential acquisition now scans the bounded durable journal under the shared
connection lock. The current scope must be fresh and earlier same-connection
role creation must be settled; elapsed deadlines and a lost process lock are
insufficient. Invalid/missing records refuse admission. Shutdown retains an
unresolved lease for later reconciliation, releasing it only with matching role
proof, independently of source-volume retention.

Tests cover missing credentials, unsafe gateway roots, invalid reviewed images,
caller ownership on refusal, cleanup after preparation, journal crash survivors,
expired unresolved jobs, partial role removal/retry and retained source volumes.
The first preparation test lacked the mocked daemon-root/workspace inventory
required for cleanup; it was corrected to supply explicit synthetic absence
evidence. No production proof was weakened. `make test` passed (server 3.298s),
and final `make check` passed (server race 4.643s), including the final bundle-close
and recovery tests, vet, Python, JavaScript and Help checks. Other Go packages
were cached. No real credentials, provider, Docker workload, SSH or database was
used; no commit or push was made.

This internal operation does not enable public managed development. A successful
combined synthetic callback round trip, live image/private-stream/provider
acceptance, refreshed-token persistence and public production admission remain
open alongside the rest of P0. See [engine runtime](engine-container-runtime.md).

## Combined synthetic Codex callback acceptance (16 October)

Added a complete mocked-daemon/local-peer run through the real owned RunCodex
operation, original ToolSession, tool/engine runtime inspection, source bootstrap,
private credential framing, callback bridge, dispatcher, framed intent/outcome
audit and fenced shutdown. The test verifies exactly two fresh creates/starts,
bootstrap before engine start, the immutable tool-helper ID and durable intent
before helper execution. Both role identities must be removed and their absence
recorded; credentials and gateway resources must be released.

The happy model turn yields its expected final text, while retained source keeps
the overall job unresolved. It never invents discard approval or deletes source
to force lifecycle success. Forbidden callbacks, audit intent failure, audit
outcome failure and engine disconnect return no success text. Failed intent runs
no helper; failed outcome withholds the observed result from the model. Completed
jobs cannot be resumed through the operation.

Initial fixture failures came from missing Docker template-format output and an
expectation of plaintext in a base64 helper response. The synthetic peer/daemon
were corrected without relaxing production checks. `make test` passed (server
3.435s), and final `make check` passed (server race 5.112s), covering all five final
scenarios plus vet, Python, JavaScript and Help checks. Other Go packages were
cached. No real credentials, image/provider, Docker workload, SSH or database was
used. This is combined synthetic acceptance, not live provider acceptance. No
commit or push was made. Live deployment, refreshed-token persistence, public
admission and the remaining P0 requirements stay open.
