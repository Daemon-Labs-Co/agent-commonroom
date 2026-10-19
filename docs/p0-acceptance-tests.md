# Remaining P0 acceptance tests

Prepared 16 October 2026 from FUNCTIONAL_SPEC.md and the current execution work.
This is a test preparation inventory, not proof of implemented functionality.
It covers P0 execution work, including completed hermetic custody scenarios; P1 collaboration and the explicitly later
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

## Current implementation status

The four credential scenarios (`refresh-atomic`, `refresh-concurrent`,
`refresh-persist-fails`, `refresh-foreign-connection`) now call real worker lease,
provisioning, validation and persistence methods with temporary credentials and a
pre-write filesystem-failure dependency. Observations come from actual files,
locks, errors and release state. Separate combined tests cover the private export
and callback/fencing path. These drivers do not launch providers or establish live
refresh compatibility.

Six storage scenarios now drive real controller reservation methods, transactional
journal appends and replay against the in-memory database adapter: exact boundary,
one byte over, concurrent last byte, retained categories, negative reservation and
no owner/paid/diagnostic bypass. They verify accounting, not physical disk quotas
or a production launch path. Sixty-four scenarios remain unbound.

Use [release readiness](release-readiness.md) for the current gap matrix. The
scenario catalog is a future acceptance contract, not a second implementation
status report.

## Running and completing them

- `make test` runs the normal suite, catalog/assertion tests and pending markers.
- `make test-p0-plan` displays individual pending cases with requirement references.
- `make test-p0-ready` fails for every missing adapter or behavioral mismatch.
- `make check` runs the repository's normal race/static/JavaScript/Help checks.

`COMMONROOM_REQUIRE_P0_ACCEPTANCE=1` is only a test-runner strictness setting, set
by the readiness target. It changes no product behavior, permissions or feature
flags, requires no restart and cannot weaken isolation/audit/retention. Leaving it
unset makes the 64 unbound scenarios visibly skipped in verbose output. A green normal
suite with skipped scenarios is not a release-readiness claim. The strict target
is intentionally failing for the remaining unbound scenarios and is not added to `make check` until the pending
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

Initial [tool-only live evidence](live-tool-acceptance.md) covers an image build,
synthetic command, kernel limit observations and diagnostic cleanup. It does not
complete any of the broader deployment checks in this table.
