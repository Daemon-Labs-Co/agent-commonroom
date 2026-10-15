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

The next concrete integration unit is a worker-owned paired lifecycle under the
existing membership/job leases:

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
