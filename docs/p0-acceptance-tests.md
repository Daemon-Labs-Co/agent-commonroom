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
or a production launch path. Three disposal refusal scenarios now drive the authenticated owner HTTP boundary:
missing approval, unauthorized human and changed volume identity. They observe
journal/worker-call/ledger state. The approved-disposal driver now combines the owner endpoint, worker journal,
mocked Docker removal/absence, exact receipt validation and atomic storage-credit
release while checking retained history. The other five disposal scenarios now
exercise revoked/stale approval, removal failure, removal followed by unavailable
absence inventory, non-destructive cancellation and refused audit-history purge.
Failure drivers use the real worker journal and cleanup code with mocked Docker;
uncertainty retains capacity and refuses another attempt. Cancellation verifies
both roles are fenced while source remains. History protection uses the actual
owner event boundary with disposal switched off after approval. All nine disposal
scenarios have hermetic drivers; deployment acceptance remains open. Thirty-nine
scenarios remain unbound.

The idle permission driver now sends a real authenticated non-owner settings
request and observes unchanged audit/state. Three additional idle drivers call the real
supervisor with explicit clock ticks and mocked Docker: 299 seconds before default
expiry, active work after 600 seconds, and idle shutdown disabled after 600 seconds.
They observe unchanged durable worker journals, both running role containers and
retained source; unexpected shutdown or journal mutation fails the contract.
Two further drivers cover queued work arriving just before expiry and an owner
interval saved through the authenticated settings endpoint. The queued case ticks
at the original expiry boundary, then checks the completion-time reset and a full
new interval. The owner case passes the saved two-minute interval into a synthetic
admitted worker and observes a committed controller notice before pair cleanup;
source and history remain retained. Its catalog now requires `idle_announced` as
well as `fenced_pair`, so missing audit evidence cannot count as success. This is
a local controller callback test, not SSH/provider acceptance.

The cleanup-uncertainty driver uses an explicit one-minute interval within the
three-minute job lifetime. It forces engine removal to fail, observes successful
tool removal and retained source, then passes the actual unresolved worker record
through a mocked evidence-return transport into controller reconciliation. It
checks durable unconfirmed shutdown and the execution stop after controller
restart, refused replacement without new journal entries, retained physical
capacity, and refusal to revive or automatically retry the old worker. The catalog
records announcement, partial cleanup, unconfirmed shutdown and execution stop in
order. This exercises real cleanup/accounting code with mocked Docker, transport
and DB; it does not establish live remote failure recovery.

Default expiry remains pending: the current paired job has a three-minute hard
lifetime, shorter than the five-minute idle default. These refusal drivers do not
prove resident-pool behavior or relax that deadline. Other idle scenarios remain pending. Separate combined transport tests now exercise
the private socket handshake and terminal audit acknowledgment with mocked
controller storage and Docker cleanup; these do not count as live SSH acceptance.

Three command refusal drivers now exercise worker-session construction with the
feature disabled, broker dispatch without the command grant, and command dispatch
over the real framed audit stream into controller storage with an injected insert
failure. The first two require no Docker or audit transport calls. The failed
intent must roll back, latch the integrity stop and release no execution/result.
These tests use the fixed Go-test profile, not an arbitrary compiler or host
command. They establish admission refusal only; successful command execution,
timeouts, output bounds and artifact delivery remain separate pending cases.

The command-outcome failure case now runs real tool-session startup, bootstrap,
fixed Go-command capture, exact-ID removal/absence checks, framed audit and
controller persistence against mocked Docker/DB. A synthetic command observes its
committed human-bound intent before output is captured. Its outcome insert then
fails: output stays withheld, the intent remains pending, integrity shutdown
latches and a fresh call is refused by revoked authority. Source remains retained.
The catalog requires durable intent, command invocation and tool absence in that
order. A healthy control commits the outcome and releases a result, preventing a
blanket refusal from satisfying the test. No actual compiler, provider, Docker
daemon or database runs. This is not revision-attested artifact evidence or live
process-descendant containment acceptance.

Use [release readiness](release-readiness.md) for the current gap matrix. The
scenario catalog is a future acceptance contract, not a second implementation
status report.

Four harness-review refusal drivers now call actual allocation registration with
signed incompatible reviews, an untrusted signing key, a missing review file, or mismatched staged-image digest metadata.
The registration cases observe the reported review error and unchanged allocation/audit state; the digest case exercises actual engine preflight and observes refusal before container creation. These
prove prelaunch rejection only; successful fresh-pair upgrades, live image-content
verification remains pending. An active-agent driver now checks a newer deployment
review while a synthetic engine is running and observes unchanged inspection,
start/create/remove counters and durable allocation. Successful new-agent upgrade
and already-current fresh-pair catalog scenarios remain pending.

## Running and completing them

- `make test` runs the normal suite, catalog/assertion tests and pending markers.
- `make test-p0-plan` displays individual pending cases with requirement references.
- `make test-p0-ready` fails for every missing adapter or behavioral mismatch.
- `make check` runs the repository's normal race/static/JavaScript/Help checks.

`COMMONROOM_REQUIRE_P0_ACCEPTANCE=1` is only a test-runner strictness setting, set
by the readiness target. It changes no product behavior, permissions or feature
flags, requires no restart and cannot weaken isolation/audit/retention. Leaving it
unset makes the 39 unbound scenarios visibly skipped in verbose output. A green normal
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
