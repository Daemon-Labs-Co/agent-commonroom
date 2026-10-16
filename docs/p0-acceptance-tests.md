# Remaining P0 acceptance tests

Prepared 16 October 2026 from FUNCTIONAL_SPEC.md and the current execution work.
This is a test preparation inventory, not proof of implemented functionality.
It covers remaining P0 execution work; P1 collaboration and the explicitly later
FR-25–28 secrets-editor/detection feature are outside this suite's release scope.

## What is ready

`cmd/server/testdata/p0_remaining_acceptance.json` contains 74 scenarios across
nine areas: production admission, provider adapters, refreshed credentials,
aggregate storage, harness upgrades, idle shutdown, commands, artifacts and
approved disposal. Each scenario specifies requirements, synthetic setup, an
action, expected decision, ordered security effects, forbidden effects and key
observations. The test-only vocabulary does not prescribe public API field names.
The future-adapter scenario tests a capability interface, not a requirement to
ship arbitrary-model plugin execution in v0.1.

`cmd/server/p0_acceptance_pending_test.go` loads and checks the catalog and provides
runnable assertions. It checks exact effect ordering, required observations and
forbidden host execution, credential auditing, history mutation, container reuse
and silent API billing fallback. Its own regression tests reject missing evidence,
wrong decisions, reordered audit/execution, unsafe effects and stale revisions.
Catalog checks reject duplicate cases, absent requirement references, contradictory
expectations and missing positive/refusal coverage.

## Surface gap snapshot

This is a filename/call-site/guard inspection of the current checkout on
16 October, not a deep code/security review or deployment verification. The older
15 October assessment records an earlier snapshot and is not current completion
evidence. No completion percentage is inferred from searches or passing tests.

| Area | Surface evidence | Remaining gap |
| --- | --- | --- |
| Production execution | `worker_codex_run.go` implements RunCodex; observed callers are tests. `credential_boundary.go` still rejects repository-bound execution. | Connect authorized production dispatch and controller/worker audit startup without removing the barrier prematurely. |
| Provider scope | Codex has the new synthetic path; `agents.go` has Claude/Codex adapters. Gemini appears in UI registration. | Reviewed isolated Claude/Gemini/Copilot/llama.cpp development paths are not demonstrated; UI labels are not execution support. |
| Credentials | Private framing, connection locks and journal checks exist. | Persist/serialize refreshed engine credentials and verify actual provider authentication. |
| Quotas | `container_quota.go` and worker admission reserve four physical slots. | No aggregate 5 GB storage admission/accounting path identified. Storage-deadline helpers are not storage quotas. |
| Harness upgrades | Reviewed immutable image inputs exist. | No instantiation-time upgrade catalog/check/compatibility/publication flow identified. |
| Idle shutdown | `worker_idle.go` probes whether work is idle. | No owner-configurable default-five-minute agent idle timer identified. A probe is not a shutdown policy. |
| Commands/artifacts | `worker_codex_turn.go` admits only read/write callbacks. | Bounded builds/tests and trusted revision-bound artifact delivery remain unfinished. |
| Disposal | Paired shutdown retains workspace data; low-level removal supports a discard input. | A public, exact-approval-bound disposal workflow and verified storage credit release are not demonstrated. |
| Live acceptance | Combined synthetic worker tests exist. | No live image/rootless/SSH/provider acceptance established by this scan. |

The table identifies visible seams and missing evidence, not proof that every
similarly named feature is absent. Full extension execution, paid consensus and
enterprise features remain deferred; they are not added to the P0 readiness gate.

All 74 behavioral scenarios currently lack production-backed adapters. They are
explicitly **pending/skipped**, not passing acceptance tests. The existing concrete
runtime/audit/quota/bridge regression tests remain separate and runnable. No stub
adapter copies expected output or pretends to verify an absent implementation.

## Running and completing them

- `make test` runs the normal suite, catalog/assertion tests and pending markers.
- `make test-p0-plan` displays individual pending cases with requirement references.
- `make test-p0-ready` fails for every missing adapter or behavioral mismatch.
- `make check` runs the repository's normal race/static/JavaScript/Help checks.

`COMMONROOM_REQUIRE_P0_ACCEPTANCE=1` is only a test-runner strictness setting, set
by the readiness target. It changes no product behavior, permissions or feature
flags, requires no restart and cannot weaken isolation/audit/retention. Leaving it
unset makes unbound scenarios visibly skipped in verbose output. A green normal
suite with skipped scenarios is not a release-readiness claim. The strict target
is intentionally failing today and is not added to `make check` until the pending
work is actually implemented.

As each feature lands, add a driver keyed by scenario ID to `pendingP0Drivers`.
Drivers call real application entry points using mocked database/provider/Docker
boundaries, synthetic credentials, temporary files and fake clocks. Observe actual
results, effects and stored events. Never construct observations by copying Want,
filter away unexpected effects, or register a canned passing driver. The suite
compares all normalized security-relevant effects, including denied/failed attempts.
Use real authorization/feature validation in enabled and disabled scenarios.
Do not create new product flags merely to satisfy this test vocabulary; use the
feature's documented server-enforced control. Mandatory boundaries have no off
switch. Each adapter owns its fixtures and cleanup; no global clock or live data.

A driver can serve multiple related cases, but each case must drive its own
isolated fixture. Do not launch real provider agents, Docker workloads, SSH or
connect to a database in this automated suite. New implementations still require
full feature documentation and the normal feature-completion checks.

## Live acceptance is separate and still pending

These manual deployment checks require dedicated synthetic repositories/accounts
and a controlled Linux rootless-Docker worker. Do not turn them into ordinary
`go test` drivers. Never copy credentials or private transcripts into evidence.
Record reviewed image/binary digests, provider/harness version, job/revision IDs,
redacted observations and teardown proof. Synthetic acceptance cannot replace them.

| Check | Procedure | Required evidence |
| --- | --- | --- |
| Provider compatibility | For each required provider, instantiate a fresh agent with supported user authentication, run one authorized edit/check callback, then cancel/reconcile. | Version/digest, billing mode, exact revision, observed check and both-role absence. Include Codex, Claude, Gemini, Copilot and llama.cpp; unsupported capabilities must refuse safely. |
| Credential refresh | Use an expiring dedicated subscription session; exercise refresh, concurrent access and a failed persistence attempt. | Private atomic custody, connection isolation, recovery state and no values in logs/events/artifacts. |
| Rootless enforcement | Verify actual daemon/cgroup limits, read-only root, no direct network/host credentials and engine/tool separation; exercise resource ceilings. | Runtime metadata plus observed denial/limits, not configuration spelling alone. |
| Transport interruption | Drop the dedicated audit SSH link during an active synthetic command, and separately interrupt the engine stream. | No subsequent unauthorized tool effect or released unacknowledged result; fenced reconciliation and retained ambiguity where necessary. |
| Crash and freshness | Kill the worker between journal/create/bind/start milestones; restart and request work. | Durable refusal until reconciliation, no adoption/restart of old containers and no lost reservations. |
| Account capacity | Concurrently allocate across rooms/workers at four containers and 5,000,000,000 bytes, including retained data and exports. | One winner at the final slot/byte, bounded allocation and durable account accounting across restart. |
| Upgrades and idle | Publish a reviewed compatible harness update and an incompatible candidate; advance real deployment idle settings around active/queued work. | New image only for a fresh agent, incompatible update blocked, owner-only configuration, active work preserved and normal cleanup at idle expiry. |
| Artifacts and disposal | Collect revision-bound check/artifact evidence, retain source, then issue exact authorized disposal. Inject remove/absence failures. | Permission and size checks, explicit current approval, storage released only after verified absence, immutable audit preserved. |

This list is a deployment acceptance plan. No live check is claimed complete here.
