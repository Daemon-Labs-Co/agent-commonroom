# Provider/tool credential separation design

**Status: runtime separation planned; protocol, injected dispatcher, real Auditor and real DockerRunner foundations implemented, not wired and not a security closure.** This is the
next core/free security boundary. Current development tools share a credential-
bearing container and can read its provider authentication. Private staging,
read-only binds, prompt instructions and output filtering do not prevent that.
Do not admit untrusted development workloads on the strength of those controls.

## Required architecture

One logical managed job owns three trusted relationships:

- A provider engine container holds its connection-specific authentication and
  provider session. It has no repository/work volume, host mount, Docker socket
  or model-accessible native shell/file/plugin/hook path. Only explicitly reviewed
  bridge tools are exposed. Unknown native tool availability must refuse dispatch.
- A separate development tool container holds the private workspace cloned from
  the approved staged snapshot. It receives no provider credential, provider home,
  OAuth environment, controller key or credential-export volume. It has no direct
  network, host mount or Docker socket. Arbitrary development commands remain
  inside this credential-free environment, not inside the provider engine.
- The trusted isolated worker supervises both containers and brokers typed tool
  requests. Only the supervisor talks to Docker; neither model nor tool payload
  supplies Docker options, host paths, images, mounts, environment or container IDs.

Both containers retain rootless/non-root/read-only-root/capability-drop/
no-new-privileges restrictions. No host execution, privileged mode or namespace
relaxation is a fallback. Give engine and tools distinct persistent volumes and
IPC endpoints. Keep a declared aggregate job resource budget: adding a second
container must not silently double unaccounted worker capacity.

This protects credentials injected by CommonRoom from development tools. It does
not guarantee an approved repository, imported transcript or human prompt contains
no secrets, or protect against compromise of the trusted provider client, worker
account, Docker daemon or host kernel. Review repository artifacts separately.

## Broker authority and request contract

The supervisor issues a random per-job capability bound to lifecycle version,
job ID, room, participant, connection, provider, repository digest and original
deadline. This capability authorizes only that job's declared tool set; it is not
an owner session, provider credential or worker-administration credential.

Start with read/write/patch/command operations on the fixed tool workspace. Use a
strict bounded schema, reject duplicate/unknown fields and trailing data, and
require unique bounded call IDs. Resolve job/container/workspace from trusted
supervisor state; never accept them from the model. File operations must reject
path traversal, symlink/mount escapes and special files. A command may manipulate
its own tool workspace, but cannot access engine volumes or the host.

Bind the bridge transport to the expected peer/job. Authenticate it independently
of tool names. Revoked/expired/mismatched capabilities refuse execution. Use a
bounded queue, action count, input/output limits, per-action deadline no later
than the job deadline and a fixed environment. Heartbeats never extend authority.
Results are untrusted data, not new capabilities or instructions to the broker.

Before a side effect, obtain a controller-issued acknowledgement of its durable
audit intent. Correlate intent/result with the job and call IDs. If the controller
connection or required persistence fails, do not execute the action or invent a
success result. Duplicate/in-progress calls must not execute again automatically;
an ambiguous prior side effect requires explicit recovery. These safety rules
are required regardless of any optional action-preview/delay UI.

## Provider adapter gates

These are design inferences from vendor interfaces, not verified launch recipes:

- [Codex app-server documentation](https://learn.chatgpt.com/docs/app-server)
  describes client-executed dynamic-tool callbacks as experimental. A transport
  prototype must pin that protocol, reject unexpected native tools/requests and
  verify resume cannot restore broader authority. Do not infer that disabling
  command tools also disables every native filesystem capability.
- [Claude CLI documentation](https://code.claude.com/docs/en/cli-reference)
  distinguishes built-in tool selection from MCP tool configuration. A broker
  adapter must prove the exact pinned flag combination exposes only its reviewed
  bridge, with no discovered settings/hooks/plugins or direct built-in tools.
  Do not assume safe-mode plus explicit MCP configuration works without testing.
- [OpenAI sandbox security guidance](https://developers.openai.com/api/docs/guides/agents-api/environments/security)
  recommends keeping real credentials outside tool environments. Its API proxy
  examples are not proof that these CLI subscriptions support token substitution.

Retain the existing CLI-based accounts; do not silently migrate customers to API
billing or another provider. If the pinned provider cannot establish the required
interface, its separated development mode is unavailable, with an explicit
reason. Do not return to credential-bearing native tool execution as a fallback.
The updated controller and worker now refuse the legacy native development path;
this closes admission to the known unsafe shared-container mode, but does not
install a working replacement or terminate existing old-worker jobs.

## Lifecycle, cancellation and migration

Introduce a new explicit worker/session execution version for the pair. Journal
both container identities and every supervisor state transition. Admission must
account for complete logical jobs and all managed containers, including unknown
orphans. A terminal engine label alone cannot settle the job.

Stop/recovery must fence future calls, cancel pending bridge work, stop both
containers and all descendants, verify absence and wait for the supervisor lease
to release before setting shutdownConfirmed. Unknown cleanup leaves execution
stopped. Startup may diagnose old work, but must not adopt/replay it automatically.

Use fresh engine and tool volumes. A previous credential-bearing writable home
or tool workspace may contain altered configuration or copied credentials; do
not reuse it as a clean boundary or claim a regex scan makes it safe. Preserve
room history, existing journals and selected consented context through the
existing disclosure policy. Review/export code artifacts independently. Keep
old operational records/volumes for operator-reviewed retention and cleanup.

## Delivery and evidence gates

1. Implement the strict scoped protocol and capability verifier with synthetic
   tests for forgery, expiry, cross-job/room/connection/digest mismatch, replay,
   queue/output bounds and cancellation. This alone does not close the finding.
2. Add credential-free tool worker and broker execution with mocked Docker
   lifecycle. Test mount/env/volume separation and refusal of client launch options.
3. Add provider bridge adapters and new pair lifecycle; test dropped callbacks,
   failure before/after side effects, restart, both-container cleanup and stale
   session/tool-list refusal. All DB interactions stay mocked.
4. Verify pinned providers and rootless runtime in a dedicated synthetic
   integration environment with synthetic canary credentials, never the user's
   live provider secrets. Demonstrate hostile file/shell/proc/symlink requests
   cannot recover the engine canary, including resumed sessions. Record actual
   effective tool inventory, image versions and termination evidence. Automated
   suites must not launch real provider agents; pinned-provider verification is
   a separately authorized manual integration exercise using dedicated test
   accounts/environment, not existing user credentials.
5. Only then claim separation for the tested adapter/runtime and remove the old
   development policy. Unknown/unsupported adapters remain unavailable.

The development capability remains off by default and operator-controlled via
its existing setting. Once implemented, this boundary is mandatory whenever that
capability is enabled; there must be no security-off switch or paid entitlement
for basic credential isolation. An experimental adapter flag may hide an adapter,
but cannot enable the old unsafe path as its fallback. No new product flag or
working broker is introduced by this design-only change.

## Scoped protocol foundation (not an execution endpoint)

`internal/toolbroker` is unused admission infrastructure. There is no configuration,
UI, network endpoint, worker integration or product toggle for it. Existing launch
behavior is unchanged. Infrastructure verification is exempt from optional feature
flags; the eventual credential boundary is mandatory and free/core.

- `Scope` carries protocol version 1, a 32-hex job ID, bounded room/participant/
  connection IDs, provider, 64-hex repository digest and the original deadline.
- `New` validates trusted scope, copies the allowed operation set and issues a
  random 256-bit capability, retaining its SHA-256 hash. The supervisor must use
  a trusted clock and deliver the returned token only through protected transport.
  Never serialize requests into diagnostics or chat: they contain the capability.
- `Decode` caps messages at 64 KiB, rejects unknown or duplicate fields (including
  nested fields and incorrect case), null values, trailing JSON and excessive
  nesting. It accepts typed read/write/patch/command payloads, not launch options.
- `Reserve` uses a mutex, constant-time hash comparison, exact scope matching,
  issuance/expiry checks, copied tool policy and a 100-call limit. Each call ID
  is consumed once even if future execution fails. Replays cannot retry side effects.
- `Revoke` prevents future reservations. It does not cancel any running process.

A reservation is **not** execution authorization. The production dispatcher still
needs authenticated peer binding, concrete durable controller persistence, safe
workspace descriptor handling and both-container shutdown proof. The injected
dispatcher below tests ordering, cancellation and bounds using mocks. Lexical path validation does not prevent symlink/mount escapes.
No concrete filesystem or command executor exists in this package. Source changes could
bypass these checks; downstream execution must enforce the complete contract.

Synthetic tests cover scope forgery, expiry/revocation, malformed input, traversal,
operation/input bounds, policy copying, concurrent replay and action exhaustion.
They do not verify a real provider, container, transport or audit integration.

## Injected dispatcher safety contract (not connected to agents)

`NewDispatcher` requires explicit `Authority`, `Auditor` and `Executor`
dependencies; no automatic executor or persistence fallback exists. `Dispatch`
reserves the call ID and copies command arguments, binds work to the earlier of
parent/original job deadlines, and shares one active slot across every dispatcher
using that authority. Concurrent calls receive a busy refusal, not a hidden queue.
Busy calls are not reserved; admitted calls are never retried automatically.

`Intent` binds trusted scope/call ID/operation to the SHA-256 action payload
without carrying the capability. `Auditor.RecordIntent` MUST acknowledge durable
controller persistence before `Executor.Execute` can be invoked. Returning nil
without persistence violates the implementation contract; mock tests prove call
ordering, not real storage durability. Intent failures revoke the whole job.
Cancellation between intent and dispatch attempts a `not_executed` outcome.

`Executor` receives typed `Action`, trusted scope, cancellation context and a
bounded writer. It must run only in the future credential-free tool container,
honor cancellation and stop all writers before returning. The writer buffers at
most 16 KiB, rejects oversized writes, cancels work and refuses writes after
completion. Overflow revokes authority. Revocation also cancels cooperative
active work. Context cancellation does not prove a rogue process has stopped;
supervisor termination and verified container absence remain essential.

`Outcome` records status, output length and digest without raw output. A result
is returned only after `RecordOutcome` acknowledgement and final cancellation
checks. Failed execution, canceled work, overflow and missing persistence return
no partial output or underlying diagnostic text. Missing outcome acknowledgement
revokes the job and leaves an unresolved intent: operator recovery must not replay
an ambiguous side effect. Cancellation/deadline may prevent outcome persistence;
that is a failure, not evidence of successful completion. Tool output remains
untrusted data even when the dispatcher succeeds.

There is no user setup, network endpoint, product flag or deployed execution path
for this internal foundation. It is free/core safety infrastructure, exempt from
optional toggles. Tests inject only in-memory audit/executor mocks and cover
intent-before-execution, outcome failure, replay, shared-slot admission,
cancellation before/during execution, expiry, output bounds, closed writers and
command argument snapshots. No database, Docker or provider process is launched.
Credential/tool separation remains OPEN until concrete adapters, lifecycle,
transport and runtime isolation pass the documented evidence gates.

## Tool-container launch policy foundation

`NewToolContainerPlan` accepts the already minted trusted `Authority` and a
privately configured immutable tool-image reference. It rejects unknown scope,
unpinned/malformed image references, pre-issuance/expired or revoked authority.
It accepts no model-selected job ID, container name, volume, host path, mount,
network or extra Docker flags. `RunArgs` checks original lifetime/revocation again
and returns a new argument slice; editing that slice cannot change stored policy.
Planning is not an atomic launch: the future supervisor must fence/revalidate at
execution and confirm termination after cancellation.

Container and workspace identities are derived from the scoped job:
`commonroom-tool-<jobId>` and `commonroom-tool-work-<jobId>`. The only mount is
this job's named workspace volume at `/workspace`; no provider home, credential,
proxy socket, Docker socket or host snapshot bind is included. It has no network,
a read-only root, UID/GID 1000, dropped capabilities, no-new-privileges, no Docker
logs and explicit proxy/Node-options resets. Scratch `/tmp` is limited/noexec.
The fixed Node entrypoint merely holds the container alive; it exposes no tool
API and establishes neither readiness nor authorization to execute commands.

The tools reserve **512 MiB memory/swap, 0.5 CPU, 64 processes, 512 file descriptors
and 64 MiB scratch**. The future paired engine must fit within the other half of
the existing logical-job budget, or undergo an explicit aggregate admission
redesign; adding a second container cannot silently double limits. Capacity and
real compilation/provider behavior are not benchmarked by argument tests.

`RemoveArgs` can target only that job's tool container. It neither executes cleanup
nor proves absence, releases pair leases, deletes volumes or revokes credentials.
The future supervisor must create fresh volumes, initialize the approved snapshot
through a credential-free bootstrap, refuse pre-existing/unresolved identities,
validate the tool-only image's contents/provenance and check readiness. An immutable
image digest alone does not prove its contents are credential-free or safe.
There is no default image, build, Docker invocation, runtime flag, provider bridge
or production wiring in this foundation. Never manually adopt these containers
into the legacy worker lifecycle, which does not understand paired identities.

Synthetic tests verify fixed containment arguments, exactly one separate named
workspace mount, absence of credential/host/socket mounts, immutable returned
policy, derived cleanup/job identities, resource allocation, image/input refusal,
expiry and revocation. Docker and provider agents are never launched. This is
required free/core infrastructure, exempt from optional toggles. Current agent
execution is unchanged; the credential-disclosure finding remains OPEN.

## Tool-container inspection gate foundation

`VerifyInspection` validates bounded single-object container and volume metadata
fetched directly from the approved rootless Docker daemon by the trusted worker.
Neither model requests nor plugin responses may supply this evidence or the
`dockerRoot` argument. The latter must come from approved daemon configuration.
The metadata shape follows the [Docker Engine API inspection reference](https://docs.docker.com/reference/api/engine/version/v1.51/).
This is an unwired verifier, not an authenticated transport or attestation service.

The gate checks scoped names/labels, approved pinned image reference, resolved
image-ID syntax (and exact ID when configured by image ID), running/unpaused/
non-restarting state, no restart count, fixed UID/workdir/entrypoint/command,
disabled healthchecks and exact environment. Inherited unknown environment is
refused; launch policy now explicitly resets Node image version fields, PATH
and proxy/Node options. Image contents/provenance and trusted daemon resolution
of manifest versus image digests remain separate requirements.

It checks private IPC/cgroup namespaces, no host PID/user namespace override,
read-only root, init, no privilege/capability additions, required security options,
fixed CPU/memory/swap/process/file limits, no devices, inherited volumes, bind
paths or published ports, disabled logs and the declared scratch tmpfs. It accepts
only one expected writable named workspace mount, plus at most the declared
`/tmp` tmpfs when the Docker version reports it separately.

The volume must use the plain local driver/scope, carry exact job labels, have no
mount options and point to `<approvedDockerRoot>/volumes/<jobVolume>/_data`.
A matching volume name with bind/device options is refused: names alone do not
prove storage isolation. `VolumeCreateArgs` generates fixed local-volume creation
with job labels and no driver options. It does not check absence, create a volume,
initialize a snapshot or prove volume history. Those future fenced-supervisor
steps must reject existing/unresolved volumes rather than adopt them.

Each inspection object is limited to 64 KiB, with a bounded nesting depth,
duplicate/case-alias rejection and no trailing JSON. Additional descriptive Docker
fields are ignored for compatibility; mandatory security values remain checked.
Malformed/missing required values, mismatched policy and revoked/expired authority
return a generic refusal without returning raw metadata or environment values.
The verifier checks selected reported policy, not every Docker field or actual
kernel enforcement. Exact supported CLI/API representations must be recorded in
separate synthetic runtime verification before deployment; unknown representations
fail closed rather than weakening the checks.

Tests mutate identities, mounts, volume options, environment, entrypoint,
namespaces, limits, logging, privilege/capabilities and malformed JSON. They also
check generated launch/environment/label policy matches inspection expectations.
No real Docker daemon, image build, provider agent, credential or database is
used. There is no runtime endpoint or optional flag for this free/core foundation.
Fresh-volume evidence, snapshot bootstrap, tool execution, disk quotas, verified
pair shutdown and provider bridge restrictions remain unimplemented/unverified.
The current shared-container credential exposure remains OPEN.

## Mock-driven container executor adapter

`NewContainerExecutor` binds a trusted plan, explicit `DockerRunner` and approved
canonical daemon data root. It implements the broker `Executor` interface but is
not installed in production. No real runner, image helper, bootstrap, network
endpoint or provider bridge is supplied. It must be used through the audited
`Dispatcher`; it is not an authentication or audit API for direct caller use.

`Execute` requires exact trusted scope, valid typed action, an existing parent
deadline no later than the original job deadline and live authority. It copies
command arguments and sends bounded action JSON on stdin to a fixed helper,
never as Docker launch flags. It first captures bounded container and volume
inspection from the injected runner and applies `VerifyInspection`. Failed or
unsafe inspection fences authority and refuses dispatch; it does not kill an
unverified container by name, which could belong to a conflicting job. The future
supervisor must quarantine and diagnose unresolved identities before admission.

After inspection, calls target the full immutable container ID rather than the
reusable name. The only execution command is a non-root Docker exec of
`/usr/local/bin/node /opt/commonroom/tool-call.cjs` in `/workspace/repository`.
The helper does not exist yet in the shipped image; this adapter is not ready to
launch tools. The runner must select only the approved local rootless daemon,
use literal argv (no shell), sanitize its host environment, stop on writer errors,
honor cancellation and avoid implicit context/host/provider fallback. These are
trusted implementation contracts, not guarantees made by an interface alone.

Inspection output is capped at 64 KiB and tool output at 16 KiB. Inspection and
actions share original deadline/parent cancellation and authority revocation.
Only successful tool output is forwarded to the dispatcher's private bounded
buffer; the dispatcher still requires durable outcome acknowledgement before
returning it to a bridge. No underlying Docker stderr/error or raw inspection
metadata is returned as diagnostics. Tool output is untrusted data.

Failed/canceled/oversized/ambiguous execution revokes authority and uses an
independent **10-second cleanup deadline**, because cleanup must still run after
execution expires. It queries exact full-ID inventory, removes only that ID when
present, then requires a successful empty inventory query. A daemon error,
unexpected ID, failed removal or remaining container produces
`ErrShutdownUnconfirmed`; failed inspection is never interpreted as absence.
A confirmed removal does not authorize retrying the original action. It proves
only the tool-side report, not engine shutdown, supervisor-lease release or kernel
containment. The future paired supervisor must also fence and verify the engine,
persist unresolved failures and release leases only after complete pair proof.

`Authority.executionContext` centralizes cooperative deadline/revocation binding
for both dispatcher and container adapter. Cancel on return prevents watcher
leaks; it cannot stop a runner that ignores context. Already spawned/background
processes and successful-call descendants remain under the future whole-job
supervisor until verified pair cleanup. Disk quotas also remain a separate gate.

Synthetic tests mock every Docker call: fixed helper/ID routing, scope/context
refusal, policy/oversized/unavailable inspection, output limits, confirmed and
unconfirmed cleanup, already-absent containers, wrong-ID refusal, independent
cleanup deadlines, running revocation and audit-before-adapter integration.
No daemon, provider, credential or DB is accessed. This mandatory free/core
infrastructure has no isolation-off feature flag. Current model turns are unchanged
and shared-container credential disclosure remains OPEN.

## Typed input allocation budgets

The JSON wire decoder is not the only entry point: trusted adapters may construct
`Request` or `Action` values directly. These values now receive pre-allocation
checks too. `validAction` counts total payload string bytes (including all command
arguments), caps content at 64 KiB, paths at 4096 bytes, arguments at 8192 bytes
and command count at 128. Oversized typed input is refused before argument copying,
capability hashing, audit calls, Docker interaction or JSON serialization.

Canonical hex identifiers/capabilities check exact length and ASCII hex bytes
without decoding/allocating an arbitrary-sized buffer. Image references are capped
at 327 bytes and approved daemon roots at 4096 bytes before expensive processing.
`Reserve` validates before hashing; `Dispatch` validates before copying commands.
The serialized 64 KiB cap remains enforced because JSON escaping and scope metadata
can make encoded messages larger than their original strings. Rejected input does
not consume a call ID. These budgets limit work done by this layer; they cannot
undo memory already allocated by an upstream caller or replace bounded transport.

Synthetic tests cover oversized content/capabilities, aggregate argument sizes,
escaped JSON expansion, early refusal before audit/executor/Docker dependencies,
image/root bounds and retrying a corrected rejected call. One zero-allocation
measurement runs serially because Go's allocation profiler changes scheduler
settings; behavior tests remain parallel and all fixtures are local. No product
flag can disable these mandatory free/core safeguards. Production separation
and runtime/provider integration remain OPEN.

## Executor declared-tool permission gate

`ContainerExecutor.Execute` now independently checks the authority's copied
allowed-operation set before serialization, inspection or any Docker call.
`Authority.permitsAction` also requires exact trusted scope, valid bounded input,
original issuance/deadline and non-revoked authority under its mutex. A read-only
job cannot call write, patch or command simply because an internal caller invoked
the adapter directly. Mutating the caller's original operation slice does not
expand the copied policy. Refusal preserves valid admission for permitted tools.

This is defense in depth, not a new public execution API: the adapter still lacks
capability authentication and durable audit acknowledgement by itself. External
requests must go through the authenticated, audited dispatcher and future fenced
supervisor. No optional flag or paid entitlement can bypass this required free/core
permission check. Tests cover undeclared operations before all Docker interaction,
allowed reads, copied-policy protection, cross-connection scope, expiry/issuance
and revocation using synthetic fixtures/mocks only. Production wiring and actual
credential separation remain OPEN.

## Rooted workspace file primitives (not wired to a helper)

`internal/toolworkspace` adds read/write building blocks for the future tool-only
container. `New` borrows an already opened `os.Root` from trusted setup; it accepts
no root path in model input and does not authenticate or authorize an action.
The caller owns/closes the root after all operations and must keep write data
unchanged until return. Passing a host worktree root to managed model work CAN
bypass isolation, BUT exposes host files and invalidates the required container
contract. No supported host fallback or security-off feature flag exists.

`Read` accepts a clean relative path of at most 4096 bytes, uses final-component
no-follow/nonblocking flags and validates the actual descriptor as an owned,
single-link regular file without group/world write bits. Directories, symlinks,
FIFOs, hard links, missing/unsafe files and oversize data are refused. Stat size
and actual reads are capped at **8 KiB**; larger-file/range APIs are not implemented.
Returned bytes are untrusted content, not instructions or authorization.

`Write` caps content at **64 KiB** and requires existing parents. It pins a parent
root/directory descriptor, validates the destination, exclusively creates a random
0600 temporary entry, fills it, applies descriptor-based permission bits and
syncs the file before rename. Existing inodes are never truncated: replacement
cannot overwrite another hard-link target's content. New files use 0644; existing
0777 permission bits (including ordinary executable bits) are retained without
set-ID bits, ownership migration, ACLs or xattrs. Unsafe existing entries are
refused. Rename, cleanup and directory sync share the pinned parent.

After rename it verifies the resulting inode, syncs the parent and checks the
requested parent/name still identify the written entry. Post-rename uncertainty
or directory-sync failure returns `ErrWriteUnconfirmed`, not success or automatic
retry. A competing writer may alter data after any check; this is not an immutable
snapshot/CAS or lock against arbitrary concurrent code. Crashes can leave temporary
files; whole-job supervision and reviewed workspace cleanup remain required.

Go's rooted operations constrain path/symlink traversal, but do not prohibit bind
mounts, special filesystem boundaries, `/proc` or devices. Descriptor validation
and the separate verified volume/no-extra-mount/device/privilege policy remain
essential. A root follows its directory inode across renames. Parent/final identity
checks prevent acknowledging a detached requested name, not every possible future
namespace mutation. Local disk I/O/sync is not forcibly interruptible by context;
cancellation is checked around steps and the supervisor must enforce job cleanup.

The package supports Linux and Darwin for synthetic verification; managed execution
still requires the supported Linux rootless worker. Tests use private temporary
roots, a synthetic outside canary, links/FIFO, size limits, atomic inode replacement,
mode preservation, cancellation, injected directory-sync failure and parent
replacement during sync. They never use live repository data, messages, credentials,
DB, Docker or provider agents. No image helper, filesystem RPC or production path
is wired; mount/kernel/pinned-Linux runtime verification and credential separation
remain OPEN. Patch, command execution and directory-creation APIs are not provided
by these file primitives.


## Bounded file helper protocol (not installed or wired)

`cmd/tool-helper` implements one read/write action per invocation for the future
credential-free tool image. Production does not invoke it. No connection setting,
UI action or operator command activates this foundation today. A shim and tool-image recipe now exist in source; an actual reviewed image,
real Docker runner, durable broker audit implementation and paired supervisor are
still required before managed use. There is no optional safety-off flag: this is
free/core mandatory isolation infrastructure. Existing development opt-in controls
remain unchanged; they do not enable this helper.

The eventual fixed executable takes **no arguments**, runs only on Linux and opens
`/workspace/repository` through a fixed `/workspace` descriptor root. It refuses a
symlink repository root and verifies the opened directory identity. Input cannot
select a root, executable or launch options. Source changes CAN redirect it to a
host root, BUT expose host files and invalidate the mandatory container contract.
It must only be invoked after authenticated broker admission, durable intent and
trusted container/mount inspection. The helper itself provides no authentication,
authorization, audit persistence or attestation of its container.

Input is a single UTF-8 JSON object, at most **64 KiB** including JSON encoding,
with exactly these four case-sensitive broker `Action` fields:

```json
{"Operation":"read","Path":"notes.txt","Content":"","Command":null}
```

`Operation` is `read` or `write`; `Path` is a clean relative workspace path.
`Content` is a string, empty for reads, text to replace for writes. `Command` must
be null or an empty array. Missing, unknown, duplicate (including escaped aliases),
incorrect-case or incorrectly typed fields and trailing JSON are refused. Commands,
patches and directory creation are deliberately unavailable; no shell fallback
exists. The wire cap can reduce usable content below the primitive's 64 KiB write
cap, especially when JSON escaping expands it. Parents must already exist.

The rooted file safety rules above remain in force. A read is capped at **8 KiB**
and returns `operation`, byte count, SHA-256, `encoding: "base64"` and base64 `data`
(empty files may omit `data`). Base64 supports binary files and bounds JSON expansion.
A confirmed write returns only operation/count/digest metadata, never a second
read of a concurrently mutable destination. The complete JSON response plus newline
is capped at **16 KiB**. Returned file content is untrusted data, not instructions.

Exit 0 means the response was delivered. Exit 2 refuses the operation/output with
a generic message; exit 3 reports an unconfirmed write and warns against automatic
retry. Response delivery/cancellation after a write is ambiguous even if the file
was replaced: the trusted broker must fence and reconcile rather than repeat it.
Partial response delivery is never success. SIGINT/SIGTERM and a three-minute local
ceiling supply cooperative cancellation; the broker's original job deadline and
whole-container termination proof are authoritative. Blocked stdin or disk I/O is
not forcibly interrupted by a Go context. The supervisor must bound and stop the
process. File sync/namespace/kernel limitations above still apply.

Automated tests use only private temporary roots and synthetic byte data, match
actual broker serialization, refuse malformed paths/schema/oversize streams and
files, check outside-link refusal and distinguish post-write response failure.
They do not launch the CLI on the host, Docker, provider agents or a database.
Darwin supports in-process synthetic verification; production remains Linux-only.
Image installation, pinned Linux execution, mount/kernel controls and actual
provider credential separation remain OPEN.


## Fixed tool launcher (not installed in a worker image)

`docker/tools/tool-call.cjs` is the source shim expected by the container executor.
It only runs on Linux with no extra CLI arguments and launches the fixed
`/opt/commonroom/tool-helper` with no arguments, `shell: false`, a fixed workspace
cwd and inherited stdin/stdout/stderr. It replaces the entire child environment
with the ten reviewed tool-environment values; provider credentials, proxy settings,
Node preload options and other inherited variables are never copied. Root selection
and executable selection are unavailable in the payload and CLI. Errors use generic
messages and preserve helper refusal/unconfirmed-write exit codes.

SIGINT/SIGTERM are forwarded to the helper; signal handlers are removed on completion.
Signal forwarding is not descendant/container shutdown proof. The verified executor
and future pair supervisor must bound output/time, fence failures and prove absence.
The shim must only be invoked in the verified credential-free tool container after
broker admission and durable audit intent. Import-time dependency injection exists
for synthetic tests, not model configuration. Source changes CAN redirect it to host
execution, BUT invalidate mandatory isolation; there is no supported fallback/off
switch. This is required free/core infrastructure, not an optional paid feature.

The new launcher tests mock every process launch and cover Linux/argument rejection,
fixed executable/cwd/environment, generic startup errors, failure codes and signals.
`make check` executes them. A tool-only image recipe now copies these files, but no image has been built
or installed, no real runner invokes them and no pinned Linux runtime is verified. The connection UI
and production dispatch remain unchanged; credential separation is still OPEN.


## Tool-only image recipe (unbuilt, not deployed)

`docker/tools/Dockerfile` builds only the file helper and launcher. It takes required
operator-reviewed immutable `GO_IMAGE` and `NODE_IMAGE` references ending in
`@sha256:` plus 64 lowercase hex digits. The actual shell guard refuses mutable tags,
missing values and malformed digests. The Go base must be Linux with Go at least
1.26; the Node base must be Linux with `/usr/local/bin/node`. Select supported,
reviewed bases with no baked credentials/provider tools, then scan/sign the result.
Pinning is not provenance: the recipe cannot prove a supplied base is clean. Base
selection happens before its build-stage validation; this guard is not a defense
against a malicious build operator or compromised builder.

The Go step disables module/toolchain downloads, CGO and network access; it copies
only `go.mod`, helper source, rooted workspace source and strict JSON encoding checks. The final stage copies
only the built binary and shim, uses UID/GID 1000 and fixed paths, and installs no
provider clients or package dependencies. No auth file, provider home, host snapshot,
SSH mount or secret build argument is required. The recipe does not itself enforce
runtime mounts/network/limits; the inspected container policy remains mandatory.

The Dockerfile-specific `docker/tools/Dockerfile.dockerignore` excludes all files,
then permits only those source leaves and necessary directories, re-excluding each
subtree before leaf exceptions. The root ignore file is unchanged. Docker documents
that a Dockerfile-specific ignore file takes precedence over the root file; see
[official build-context documentation](https://docs.docker.com/build/concepts/context/#filename-and-location).
Do not add broad context copies/exceptions or use a live working data directory as
build material. Docker still transfers the recipe/ignore metadata as needed.

For a separately approved synthetic Linux integration environment, after setting
both values to reviewed real digest references, build from the repository root:

```sh
docker build --network=none -f docker/tools/Dockerfile \
  --build-arg GO_IMAGE="$COMMONROOM_TOOL_GO_IMAGE" \
  --build-arg NODE_IMAGE="$COMMONROOM_TOOL_NODE_IMAGE" \
  -t commonroom-tools:review .
```

The local review tag is a build label, not an approved runtime reference. Runtime
configuration must use the reviewed resulting immutable image ID/digest and pass
inspection. Build network restrictions apply to RUN steps, not registry base-image
resolution; preload reviewed images/use a protected builder as appropriate. Never
pass provider credentials to this build. This recipe is not connected to installation,
connection setup or dispatch, and there is no product switch to activate it today.
It is mandatory free/core infrastructure with no supported isolation bypass.

Checks enforce the narrow copy/context contract and execute only the pin guard
against synthetic strings. A minimal-source Linux cross-build with an
empty module cache/downloads disabled passed; its binary was not executed. No actual
Docker image build, Docker ignore-engine evaluation, image scan/signature, live
mount/kernel/pinned Linux execution or provider turn was performed. Those gates,
provider bridge, volume/bootstrap freshness, audited runtime and pair shutdown remain
OPEN; production provider credential exposure is not resolved by this recipe.


## Strict tool input encoding

Broker wire admission and file-helper input reject malformed UTF-8 and unpaired
escaped UTF-16 surrogates before Go JSON decoding can silently replace them.
Typed internal actions independently require valid UTF-8 in paths/content/command
arguments before copying or marshaling. Explicit replacement characters, ordinary
Unicode, valid surrogate pairs and literal escaped backslash text remain valid.
This prevents admitted text from changing silently into a different path/content.
It does not normalize Unicode, resolve filename equivalence or authorize content.

`internal/strictjson.ValidEncoding` is a no-allocation encoding check, not a complete
JSON grammar/schema parser. Callers first impose their wire byte budget and still
run their strict schema/duplicate/trailing checks afterward. It must not replace
those checks or authentication/audit. The tool image's narrow source context includes
only the extra encoding source file needed by the helper.

Synthetic tests demonstrate Go's lossy default marshal/decode behavior, reject
malformed wire and typed actions, preserve valid/literal Unicode, verify zero
allocation and exercise helper refusal before filesystem effects. Full tests/checks
pass. This required free/core validation has no bypass switch. Runtime transport,
pinned image execution and provider/tool credential separation remain OPEN.


## Helper result admission before success

The unused `ContainerExecutor` now supports only the helper's implemented `read`
and `write` operations. Even if a trusted authority declares `command` or `patch`,
this adapter refuses them before daemon interaction; future implementations must
supply reviewed operation/result contracts rather than accept arbitrary output.

A zero process exit does not by itself acknowledge a successful tool action.
Before publishing output, the executor validates a single bounded UTF-8 JSON result
with exact case-sensitive fields, no duplicates/nulls/trailing data and canonical
SHA-256. Required `operation`, integer `bytes` and `sha256` must match the action.
Read results require `encoding: "base64"`, canonical base64 bytes at most 8 KiB,
matching byte count and SHA-256. Empty read data may omit `data`. Write results
must omit encoding/data and match the original admitted content length/digest.
This is a bounded consistency receipt, not independent filesystem attestation:
a compromised image/helper could forge consistent metadata without performing a
write. Reviewed image provenance, inspected mounts, trusted helper and supervision
remain essential. Concurrent namespace/disk and crash limitations remain OPEN.

Malformed or inconsistent output never reaches the caller as success. It fences
authority and enters the existing immutable-container cleanup/absence-proof path,
without replaying an ambiguous action. A shutdown failure remains unconfirmed;
the future pair supervisor must additionally stop and verify the provider engine.
Dispatcher success auditing only follows accepted executor output. Audit outcome
output hashes refer to the entire returned JSON; the result's SHA-256 refers to
file bytes. Neither hash proves independent audit anchoring or persisted filesystem
state after subsequent concurrent work.

Synthetic tests cover false-success output, incorrect operations/counts/digests,
unknown/duplicate/escaped keys, malformed encodings/base64, oversized reads,
write-content binding and refusal before Docker for unsupported declared tools.
Process/container interactions are mocked; invalid-output tests prove caller output
is withheld and the job fenced/cleanup queried. Full tests and checks pass. This
required free/core verification has no bypass flag. Real image/tool runner,
provider bridge, durable audit transport and production credential separation
remain OPEN, with no UI/connection behavior or deployment change.


## Legacy development admission barrier

`requireCredentialSeparatedDevelopment` rejects nonempty development repository
bindings in controller execution, worker request validation and local preparation,
before credential reads or Docker setup. There is no flag, owner exemption or paid
entitlement for the old path. Signing in again, using read-only credential mounts
or disabling inner namespaces cannot satisfy this boundary. The user-facing agent
status identifies the bridge requirement rather than an authentication failure.

This is temporary feature unavailability while the replacement is implemented,
not a completed provider bridge. Existing off-by-default development operator
flags/registries and stored snapshots remain intact; enabling them cannot restore
native tool execution. Chat/login retain their policies, without a claim that this
barrier proves all chat native capabilities. Existing jobs/old binaries are not
killed or modified; operators must stop and verify workers before deploying updates.
Automated tests use synthetic requests and prove rejection before preparation.

Local metadata inspection found Codex CLI `0.162.0-alpha.17.2`; generated app-server
schemas include experimental dynamic tools but no guaranteed disable-all-native-tools
field. Claude metadata reported `2.1.287`; vendor docs distinguish built-in tool
selection, MCP configuration and restricted/customization modes. Neither local
metadata inspection nor published flag descriptions substitutes for pinned Linux
runtime/native-tool verification. No provider agent was launched and no credential
contents were read. A reviewed client build or explicitly authorized API-backed
engine is required where the stock CLI cannot prove the boundary; do not silently
change existing subscription accounts to separately billed API usage.

## Real Auditor and DockerRunner (first concrete implementations)

Two of the three remaining "no real implementation" gaps in the mock-driven
foundation above are closed in source: `cmd/server.storedToolAuditor` is the
first real `toolbroker.Auditor`, and `cmd/server.dockerCLIRunner` is the first
real `toolbroker.DockerRunner`. Neither has a production caller; `cmd/server.
newToolDispatcher` proves they assemble into a working `Dispatcher` and is
itself only called by its own test.

`storedToolAuditor.RecordIntent`/`RecordOutcome` persist through the existing
append-only audit chain (`event_store.go`'s `s.append`), as `tool.call.intent`/
`tool.call.outcome` events, inheriting that path's hash-chain, actor
attribution and transaction guarantees rather than adding a parallel storage
mechanism. `state_projection.go` replays both kinds into `room.State.ToolCalls`
(unexported from `/api/state` for now; add a projection consumer only when one
needs to display pending/settled calls). An outcome event with no matching
intent fails replay closed, matching this package's existing philosophy.
`RecordIntent` validates with `toolbroker.ValidAuditIntent` at the real
admission instant; `RecordOutcome` deliberately does not re-run that
deadline-in-future check, since a call that ran right up to (or past) its
deadline still produces a legitimate outcome — re-validating "now" there would
incorrectly reject it, exactly as `ValidAuditIntent`'s own doc comment warns.
Replaying historical intents reuses `ValidAuditIntent` against the intent's
own recorded deadline instant rather than wall-clock "now", for the same
reason: replay happens long after any 3-minute admission window closed.

`dockerCLIRunner.Run` shells out to the local `docker` CLI with the exact
arguments `ContainerExecutor` already built — it parses no shell, accepts no
caller-chosen executable, and never inherits the ambient environment (same
three-variable allowlist as the existing single-container agent path). It
re-verifies the local-only, resource-enforcing Docker context
(`verifiedLocalDockerContext`, extracted from `isolation.go` so this check is
defined once and shared by both the existing agent-container path and this
new tool-container path) before *every* call, not once per session, since one
tool action drives several `Run` invocations (inspect, inspect, exec, and on
failure ps/rm/ps) and each is a point where a tampered or remote context must
still be caught.

Automated tests mock the database (`storedToolAuditor`) or check only
precondition refusal before any subprocess runs (`dockerCLIRunner`); per this
project's established convention for every other Docker-touching path, live
Docker behavior is not mocked with a fake `docker` shim and stays a documented
manual verification step. `newToolDispatcher`'s test confirms the assembly
compiles and constructs correctly without touching Docker, since
`NewContainerExecutor` only validates inputs at construction time.

**Still open, unchanged by this update:** no caller dispatches a real tool
call during a live agent turn; the transport question of how a provider's
in-container tool call would get intercepted and routed through this broker
(instead of running directly in the provider's own container, as today)
remains an explicit, separate, un-started design decision. Volume bootstrap
from an approved repository snapshot, pairing a tool container's lifecycle
with its owning agent job, and the real tool image build/install are also
still unimplemented. Shared-container credential disclosure remains OPEN.

## Dedicated Codex client build

The [reviewed engine build instructions](../docker/engine/README.md) describe the
pinned upstream source, callback-only router patch, source/policy integrity checks
and required Linux verification. This retains CLI subscription accounts. It is not
yet a production bridge: binary verification, controlled engine settings, durable
audit callbacks and paired supervision remain required before development can run.
No feature flag can bypass these prerequisites.
