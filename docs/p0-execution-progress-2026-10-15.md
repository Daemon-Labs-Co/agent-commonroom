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
