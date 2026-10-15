# Security and cleanup benchmark — 13 October 2026

Measured wall-clock time in this development session, including tool execution,
test runs, documentation and corrections. This is not a general delivery guarantee.
No work was delegated; no commits or pushes were created.

| Stage | UTC start | UTC checkpoint | Elapsed |
| --- | --- | --- | --- |
| Initial context disclosure implementation, tests and documentation | 18:31:16 | 18:38:56 | 7m 40s |
| Atomic operations, retry receipts, revisions, and remaining disclosure corrections | 18:38:56 | 19:05:59 | 27m 03s |
| Total for both security changes through full verification | 18:31:16 | 19:05:59 | **34m 43s** |
| `main.go` structural cleanup, documentation and verification | 19:05:59 | 19:24:24 | **18m 25s** |

The first checkpoint was implementation progress, not independently completed
release verification. Disclosure corrections and its new UI/session tests continued
in the second stage; the combined **34m 43s** is the reliable security benchmark.
The cleanup includes repairing an extraction error involving UTF-8 offsets and an
incorrectly selected nested switch. That correction time is intentionally included.
The README rewrite followed these checkpoints and is not included in them.

Security verification: `make test` **2.37s**, `make check` **6.85s**.
Post-cleanup verification: `make test` **2.13s**, `make check` **6.52s**.
All database interactions were mocked; no live database data or provider agents
were used. These measurements exclude fresh deployment/provider interoperability
and do not establish that all context-disclosure or distributed-execution risks
have been eliminated. See the feature guides for exact boundaries.

The structural result is `main.go` **1,461 → 184 lines** against the committed base.
Immediately before extraction it was 1,633 lines, including the security additions.
The code moved to focused storage, projection, identity, handler and validation
files; it was not removed from the application.

Final application check: **2026-10-13 19:49:32 UTC**. Total elapsed from the first
clock reading: **1h 18m 16s**, including both security changes, structural cleanup,
README rewrite, final review corrections and local deployment. Later corrections
covered legacy private-memory visibility and retaining the displayed form revision
across background updates; they are included in this overall total, not retroactively
hidden inside the earlier checkpoints.

Latest checks: `make test` **0.34s** and `make check` **2.34s** (Go results cached;
JavaScript checks ran). The preceding uncached Go run took **2.22s**, with full
checks at **6.32s**. The rebuilt controller serves the application and Help with
HTTP 200; unauthenticated state access returns 401. Worker idle was verified before
restart. No live model interaction or database test fixtures were introduced.

## Security-first implementation checkpoint — 13 October 2026

Workflow outcome atomicity and fail-closed persistence: started **17:58:11 EDT**,
verified **18:00:18 EDT**, elapsed **2m 07s**. Includes source inspection,
implementation, mocked regression tests, documentation and required checks.
`make test`: **1.44s**; `make check`: **6.75s** (includes race/static/UI checks).
An earlier successful test run took 1.98s; the initial sandbox-only attempt could
not access the existing Go build cache and did not run the server tests.

No live database/provider tests or controller deployment were performed. This is
one focused security correction, not completion of the security/refactor/plugin
roadmap. The prior 15x planning factor remains a hypothesis for comparable slices;
this checkpoint has no independent pre-implementation estimate.

## Atomic task dispatch — 13 October 2026

Started **21:14:42 EDT**, final code verification **21:21:14 EDT**: **6m 32s**.
Includes implementation, a corrected synthetic fixture, mocked transaction/retry/
permission/result tests, UI tests, documentation and full checks. No separate
pre-implementation estimate was recorded for this slice.

Final `make test`: **2.05s**; full `make check`: **6.37s**, including Go race/static
and JavaScript checks. Earlier successful runs: test 2.27s/2.10s, full check
6.74s/6.39s. Tests did not connect to a database or launch real provider agents.
This is source verification only; the running controller was not replaced and
live worker/provider acceptance is still outstanding.

## Managed connection verification — 13 October 2026

Started **21:46:35 EDT**, final code verification **21:54:06 EDT**: **7m 31s**.
Includes inspection, implementation, a corrected non-comparable binding check,
mocked lifecycle/HTTP/cancellation/capacity regressions and documentation review.
Final full `make test`: **2.42s**; full `make check`: **6.58s** (race/static/UI).
Earlier successful test runs took 2.49s, 1.91s and 2.18s; the first full check took
6.68s. All provider/SSH/DB operations in tests were mocked.

No live controller/worker deployment or provider acceptance was performed. Update
both components before claiming the old verification launcher is disabled at
runtime. The capability reuses the existing default-on `connectionTesting` flag;
mandatory job journalling and shutdown verification have no independent off switch.

## Durable daily turn budgets — 13 October 2026

Started **21:57:00 EDT**, final integrated verification **22:29:20 EDT**:
**32m 20s elapsed**. This includes implementation, mocked tests, documentation,
Git dependency inspection, preserving the budget patch, fast-forwarding the feature
branch to the owner's existing connection-verification commit and resolving four
overlaps. Main was unchanged; no new commit or push was created. This is total
elapsed time, not a separately measured coding-only duration.

Final full `make test`: **2.06s**; full `make check`: **6.42s** (race/static/UI).
Earlier integrated runs took 2.36s / 6.72s. All DB, transport and provider interactions
were mocked. The final regression explicitly proves verification jobs share the
budget. No controller/worker deployment or live provider acceptance was performed.
Daily limits depend on an intact controller journal and a single-controller
installation; monetary/storage quotas and independent rollback protection remain
outstanding. See [execution budgets](execution-budgets.md).

## Private file and lease hardening — 13 October 2026

Started **22:38:59 EDT**, automated verification **23:17:07 EDT**:
**38m 08s elapsed through verification**. Includes source inspection, private-read/
lock implementation, temporary-file regressions, documentation and identifying
three required budget files omitted from the preceding owner commit. Those files
were preserved for inclusion in the next commit. This is elapsed wall time, not
a separately measured coding-only duration. A later user message resumed work
at 00:36 on 14 October; that intervening time is not included in this checkpoint.

Full `make test`: **2.56s**; full `make check`: **9.33s** (race/static/UI).
Earlier successful test runs took 2.53s and 2.12s. File fixtures are synthetic and
use temporary directories; no live credentials, database or provider were used.
Controller/worker upgrades and installation ownership compatibility remain
unverified. Parent directories and host administrators remain trusted.

## Login availability — 14 October 2026

Started **00:40:27 EDT**, final verification **00:45:42 EDT**: **5m 15s**.
Valid proofs survive failed-attempt response exhaustion; invalid member proofs are
consumed/bounded before DB replay; session renewal works at capacity and rolls
back on persistence failure. Full `make test`: **2.53s**; `make check`: **6.52s**.
Synthetic keys/session fixtures and mocked DB only; no live logins/deployment.
This does not close volumetric, challenge-pool or cryptographic CPU DoS concerns.

## Audit format barrier — 14 October 2026

Review started **00:47:27 EDT**, final verification **00:52:06 EDT**: **4m 39s**.
Protected-to-legacy format downgrades now fail integrity checks, cancel work and
block recovery while legitimate historical prefixes stay untouched. `make test`:
**2.46s**; full `make check`: **7.06s**. Synthetic in-memory event history only;
no live DB/provider use. This is not independent audit rollback protection.

## Storage deadlines — 14 October 2026

HTTP/storage review started **00:53:23 EDT**, storage implementation began
**00:55:48 EDT**, final verification **00:59:46 EDT**. Review-through-verification:
**6m 23s**; implementation-through-verification: **3m 58s**. Full `make test`:
**2.64s**; `make check`: **6.51s**. Mocked DB stalls and synthetic cancellation
contexts only. No live storage outage or physical worker shutdown was tested.

## Restart-safe failed emergency stops — 14 October 2026

Started **01:01:21 EDT**, final verification **01:05:47 EDT**: **4m 26s**.
Stop/resume audit failure now latches the private integrity marker and cancels all
contexts. Mocked room/application begin/insert/commit failures verify cancellation,
restart restoration and refusal of ordinary resume. `make test`: **2.54s**;
full `make check`: **6.93s**. Synthetic temporary markers only; no live DB/worker
outage or deployment. Unwritable marker storage remains an operator blocker.

## HTTP header limit and credential-boundary review — 14 October 2026

Review checkpoint **01:10:29 EDT**, final verification **01:12:03 EDT**:
**1m 34s** (including the credential-boundary review in this slice). Full
`make test`: **2.36s**; `make check`: **6.51s**. Mandatory 16 KiB header budget
rejects oversized input before handlers; isolated ephemeral loopback tests use
no database/provider. No live deployment. Credential/tool co-location remains
explicitly OPEN and requires a separate provider/tool execution architecture.

## Private controller configuration — 14 October 2026

Started **01:12:33 EDT**, final verification **01:14:04 EDT**: **1m 31s**.
Local/development configuration reads now validate the opened inode, bound size
and reject unsafe files. Full `make test`: **2.67s**; `make check`: **6.56s**.
Synthetic temporary configuration only, with no real secrets or live deployment.

## Session/stop directory durability — 14 October 2026

Started **01:15:11 EDT**, verified **01:16:38 EDT**: **1m 27s**.
Initial full tests **2.39s**; final full `make test` **1.98s**, `make check`
**6.85s**. Synthetic sync-failure injection/temp files only. Real power-loss,
remote filesystem and live worker/storage outage testing remains outstanding.

## Local controller startup fencing — 14 October 2026

Started **01:17:12 EDT**, verified **01:18:23 EDT**: **1m 11s**.
Required private kernel lease rejects overlapping controllers using the same
data directory. Full `make test` **2.38s**, `make check` **6.98s**. Temporary-file
lease tests only; no real DB/server/provider startup. Distributed coordination
and old controller versions remain outside this local lease.

## Bounded credential-safe DB startup — 14 October 2026

Started **01:19:44 EDT**, verified **01:20:43 EDT**: **59s**.
Runtime startup shares a ten-second deadline and generic connection errors. Full
`make test` **2.44s**, `make check` **6.54s**. Mocked ping/query/row stalls and
privilege checks only; no real DB or provider.

## Provider transition-address denial — 14 October 2026

Review checkpoint **01:22:16 EDT**, verified **01:23:31 EDT**: **1m 15s**.
Provider gateway rejects transition/translation and special-use IP destinations
while retaining native public IPv4/IPv6 and exact-host checks. Full `make test`
**2.04s**, `make check` **6.90s**. Synthetic addresses only: no DNS/provider
requests or live routing tests. NAT64-only networks are intentionally unsupported.

## Codex chat command-tool restriction — 14 October 2026

Review resumed **01:41:16 EDT**, implementation checkpoint **01:42:25 EDT**,
verified **01:43:45 EDT**. Review-through-verification **2m 29s**; implementation
through verification **1m 20s**. Full `make test` **2.41s**, `make check` **6.77s**.
Synthetic worker-prefix enforcement only; no live provider/image tool verification.
Development credential/tool co-location remains OPEN.

## Provider gateway shutdown — 14 October 2026

Started **01:46:00 EDT**, verified **01:46:59 EDT**: **59s**.
Gateway closure cancels pending DNS/dial work and refuses new work; pre-tunnel
HTTP limits are explicit. Initial `make test` failed in **2.55s** because the
macOS default fixture socket path exceeded the platform limit; a short unique
temporary fixture fixed that. Final full `make test` **1.81s**, `make check`
**6.65s**. Fake resolver/private Unix socket only, with no DB/provider/DNS access.
Physical worker termination and credential-free development tools remain open.

## Development snapshot substitution — 14 October 2026

Review started **01:56:14 EDT**, final verification **02:00:47 EDT**: **4m 33s**.
Development launches now mount a checksum-verified private per-turn copy, with
bounded nonblocking/no-follow source reads and normalized read-only permissions.
Initial full `make test` **2.32s**, initial `make check` **6.86s**; after permission
normalization, final full `make test` **2.32s**, `make check` **6.82s**. Synthetic
files only; no Docker/provider/DB access. Live mount/disk-outage and crash-temp
cleanup verification remains outstanding; credential/tool separation is OPEN.

## Legacy shutdown-proof barrier — 14 October 2026

Review checkpoint **02:26:17 EDT**, implementation **02:28:15 EDT**, verified
**02:30:04 EDT**: **3m 47s** review-through-verification, **1m 49s** implementation
through verification. Initial full test **2.45s** failed on an old fixture assumption
that legacy cancelled text proves shutdown; the regression now checks refusal
and current confirmed acceptance. Final full `make test` **2.12s**, `make check`
**6.89s**. Synthetic records only; no live DB/Docker/provider/legacy-process
inspection. Manual legacy maintenance is required; no automatic rewrite/replay.

## Fail-closed identity defaults — 14 October 2026

Started **02:41:16 EDT**, verified **02:44:41 EDT**: **3m 25s**.
Missing context/actor-less sessions no longer acquire owner identity; key download
has an explicit owner check. Initial full `make test` **2.64s** failed on a wrong
method in the new refusal table and an old actor-less owner test fixture. Corrected
fixture run **2.70s**; final owner-download-inclusive full test **2.23s**; full
`make check` **6.95s**. Synthetic keys/sessions and mocked storage only. No live
account/provider access or deployment; ordinary routes still require middleware.

## Exclusive per-turn credential staging — 14 October 2026

Started **02:56:20 EDT**, verified **02:59:27 EDT**: **3m 07s**.
Credential staging now uses exclusive creation, private-parent validation and
explicit mapped-UID readability. Full `make test` **2.59s**, `make check` **7.02s**.
Synthetic JSON/temp-file fixtures only; no live secrets, DB, Docker or provider.
Mapped-UID/crash cleanup integration and provider/tool credential separation
remain open.

## Durable local provider credential revocation — 14 October 2026

Started **03:11:17 EDT**, verified **03:14:01 EDT**: **2m 44s**.
Sign-out removal/retries and login publication now require directory durability
acknowledgement; unsupported providers cannot select a default credential. Full
`make test` **2.36s**, `make check` **6.70s**. Synthetic credentials and injected
sync errors only; no real DB/Docker/provider/token revocation. Physical power-loss
and historical volume cleanup guarantees remain unverified/unimplemented.

## Bounded history storage reads — 14 October 2026

Started **03:26:36 EDT**, verified **03:30:43 EDT**: **4m 07s**, including execution
tool waiting. Full `make test` **2.34s**, `make check` **6.68s**. Settings/query/row
reads now share the mandatory deadline and reject partial timeout responses.
Synthetic mocked rows only; no real DB/provider or live contention test.
Total-storage quotas, volumetric protection and provider/tool separation remain
open.

## Provider/tool separation design — 14 October 2026

Review checkpoint **03:42:01 EDT**, verification **03:46:49 EDT**: **4m 48s**.
Documented the planned two-container broker, scoped authority, durable action
acknowledgements, pair cleanup, fresh-volume migration and provider evidence
gates; added it to in-app Help. This is DESIGN ONLY, not a runtime fix. Full
`make test` **2.33s**, `make check` **6.75s** validate the catalog/existing suite.
No provider agents, live credentials, DB or deployment accessed. Pinned-provider
integration verification requires a separate authorized synthetic test environment.
The credential/tool disclosure finding remains OPEN.

## Scoped tool-broker protocol foundation — 14 October 2026

Review checkpoint **03:56:14 EDT**, verification **04:01:41 EDT**: **5m 27s**.
Implemented unused, bounded typed request decoding, scoped random capability
admission, expiry/revocation, replay refusal and per-job action limits. Full
`make test` **0.99s**, `make check` **4.10s** (existing Go package results cached;
new broker tests executed, full check includes race/static/UI suites).
Synthetic in-memory fixtures only; no DB, provider, container or live credentials.
This foundation has no execution endpoint or feature toggle and does not fix the
OPEN shared-container credential exposure. Transport/audit/queue/output/paired
runtime integration remains to be implemented and verified.

## Injected tool-broker dispatcher safety — 14 October 2026

Started **04:11:34 EDT**, final verification **04:16:21 EDT**: **4m 47s**.
Added dependency-injected audit-before-execution, outcome acknowledgement,
shared-job admission, revocation/deadline cancellation and bounded output.
Initial full `make test` **1.08s**; after cancellation evidence tests **0.75s**;
full `make check` **4.10s** (including new Go race tests). Final documentation
verification: full `make test` **0.17s**, `make check` **2.28s**, Go results cached.
All dependencies use synthetic in-memory mocks; no DB, provider, Docker or live
credentials. No execution endpoint is wired and credential separation remains
OPEN. Cancellation is cooperative, not verified container termination; concrete
controller persistence, transport and credential-free execution remain missing.

## Provider gateway pre-header admission — 14 October 2026

Review checkpoint **04:26:45 EDT**, verification **04:29:42 EDT**: **2m 57s**.
Added a fixed 16-client gateway listener cap before HTTP header parsing, while
preserving 8 CONNECT slots and existing cancellation/deadlines. Full `make test`
**2.34s**, `make check` **6.88s** (including race/static/UI suites).
Tests use synthetic pipe listeners and ephemeral local Unix sockets, with no
provider/DNS/DB/credential access. No deployment was performed. This fixes
unbounded pre-header HTTP connection state in this gateway; volumetric accept
CPU/socket pressure, public-server admission and provider/tool separation remain
open. The cap is mandatory free/core infrastructure, with no off switch.

## Trusted tool-container launch policy foundation — 14 October 2026

Review checkpoint **04:42:30 EDT**, verification **04:47:08 EDT**: **4m 38s**.
Added a trusted-authority-bound, immutable Docker argument plan with separate
job workspace, no credential/host/socket mounts, disabled network and declared
half-job resource allocation. Expiry/revocation prevents new plan generation.
Full `make test` initially **1.07s**, after authority binding **0.61s**; full
`make check` **4.43s**, including new race tests. Existing Go packages were cached.
Only synthetic scopes/argument checks used; no Docker/provider/DB/credentials or
image build. This is unused launch-policy infrastructure, not an executor or
credential-isolation closure. Image provenance, fresh volume initialization,
atomic fenced launch, pair budgeting/lifecycle and runtime evidence remain open.

## Tool-container inspection gate foundation — 14 October 2026

Review checkpoint **04:58:27 EDT**, verification **05:06:09 EDT**: **7m 42s**,
including primary Docker schema research. Added bounded container/volume metadata
verification for scoped identities, environment, selected containment/resource
policy and redirected-volume options; aligned launch namespace/health/environment
policy. Full `make test` initially **1.07s**, final **0.75s**; full `make check`
**4.27s**, including new race tests (existing Go packages cached).
Synthetic inspection fixtures only; no Docker/provider/image build/DB/credentials.
This verifier is unwired and requires trusted-worker/daemon evidence. It cannot
prove image contents, fresh volume history, kernel enforcement, atomic launch,
disk quotas, pair shutdown or actual credential separation; those remain OPEN.
No deployment or optional safety-off flag was introduced.

## Mock-driven container executor adapter — 14 October 2026

Review checkpoint **05:13:17 EDT**, final verification **05:21:57 EDT**: **8m 40s**.
Added inspected immutable-ID helper routing, bounded typed input/output, shared
revocation/deadline context and independent 10-second failed-work cleanup with
successful exact-ID absence proof. Full `make test` initially **1.05s**, after
revocation/audit integration **0.97s**, final **0.64s**; full `make check` **4.23s**,
final **3.96s**, including new race/static/UI checks (existing packages cached).
All Docker interactions are injected mocks; no daemon/provider/DB/credentials
or image build. No real runner/helper/production wiring exists. The adapter
cannot alone establish pair shutdown, fresh workspace/disk quotas or provider
native-tool restrictions; credential separation remains OPEN. No deployment,
commits or safety-off flag introduced.

## Typed broker pre-allocation budgets — 14 October 2026

Heartbeat wake **05:26:08 EDT**, documentation complete **05:39:01 EDT**:
**12m 53s** end-to-end, including pre-tool design review. Recorded implementation
checkpoint **05:32:05 EDT**. Rejected oversized typed payloads/capabilities before
copy/hash/serialization; retained escaped JSON wire caps and bounded config inputs.
Initial full tests **1.06s** failed because allocation profiling cannot run in a
parallel test; **0.67s** then exposed the remaining allocating hex decoder.
After canonical byte validation: **0.60s** passed. Final full `make test` **0.63s**,
`make check` **4.20s**, including race/static/UI checks (existing packages cached).
Only the allocation profiler is scheduler-serialized; behavior tests stay parallel.
Synthetic in-memory fixtures and mocked dependencies only; no DB/Docker/provider
or credentials. This hardens unused broker infrastructure, not deployed execution;
provider/tool credential separation remains OPEN. No safety-off flag or deployment.

## Executor declared-tool authorization — 14 October 2026

Heartbeat wake **05:42:08 EDT**, verification **05:45:30 EDT**: **3m 22s**.
Added independent executor permission enforcement for the copied tool set,
trusted scope, issuance/deadline and revocation before any Docker interaction.
Full `make test` **1.03s**, `make check` **4.37s**, including new race/static/UI
checks (existing Go packages cached). Synthetic scopes and Docker mocks only;
no real DB/provider/daemon/credentials. This closes an internal adapter policy
bypass in the unused foundation, not production credential exposure. Audited
transport, real tool helper/runner and paired supervisor remain OPEN. Required
free/core safety check has no off switch or deployment change.

## Rooted workspace file primitives — 14 October 2026

Heartbeat wake **05:57:08 EDT**, final verification **06:11:10 EDT**:
**14m 02s** end-to-end, including rooted-I/O design review. Added bounded owned
regular-file reads, atomic new-inode writes, pinned-parent rename/sync/cleanup and
unconfirmed-outcome refusal. Full `make test` initially **1.04s**, after pinned
parent checks **0.64s**; full `make check` **4.34s**, including race/static/UI suites.
Linux/amd64 test cross-compilation **0.70s** passed; binaries were not executed.
Fixtures are private temporary files/links/FIFO and synthetic canaries with injected
sync/parent replacement failures; no DB/Docker/provider/live data or credentials.
Local standard-library documentation was read; unrelated module-metadata lookups
were network-restricted but did not block tests. No image helper, RPC or production
execution was wired. Mount/kernel/Linux runtime, arbitrary concurrent-code/disk/crash
controls and provider/tool credential separation remain OPEN. No safety-off flag,
commit or deployment was introduced.


## Bounded tool file helper — 14 October 2026

Active implementation began **07:31:20 EDT**, final verification/documentation
**07:35:32 EDT**: **4m 12s**. Earlier queued heartbeats did not produce implementation
work in this slice and are not counted as execution. Full `make test` initially
**1.14s**, after budget regressions **0.80s**; full `make check` initially **4.25s**,
final **3.69s**, including race/static/UI suites (unchanged Go packages cached).
Added strict broker-compatible read/write schema, fixed Linux workspace entry,
bounded binary responses and post-write delivery ambiguity handling. Synthetic
private temp files/links and input/output failures only; no DB/provider/Docker,
live credentials, image build or deployment. The image shim/real runner, original
job supervision and production credential separation remain OPEN. This is mandatory
free/core safety infrastructure, without a bypass feature flag.


## Fixed tool launcher — 14 October 2026

Active implementation **07:36 EDT** through verification **07:37:39 EDT**, about
**1m 39s** (start rounded to minute). Full `make test` **0.38s** (Go packages cached);
full `make check` **2.52s**, including five new mocked process-launch tests plus
race/static/UI suites. Fixed executable/cwd, cleared child environment, generic
launch errors and signal/exit behavior tested without starting any real helper,
provider, container or DB. Image installation, original-job/pair supervision and
production credential separation remain OPEN. Required free/core infrastructure
has no bypass flag; no deployment performed.


## Tool-only image recipe — 14 October 2026

Active work **07:38:35 EDT** through verification/documentation **07:41:38 EDT**:
**3m 03s**. Full `make test` initially **0.32s**, final **0.19s** (Go cached); full
`make check` initially **2.54s**, final **2.52s** including four new image-contract
checks and actual offline pin-guard tests with synthetic image names. Minimal-source
Linux cross-build in a private temporary directory, empty module cache and downloads
disabled: **1.94s**, binary not executed. No Docker image built/deployed, no real
Docker-ignore engine, provider, credential or DB interaction. Official Docker
build-context documentation was checked for Dockerfile-specific ignore precedence.
Base provenance/pinned Linux image runtime, real runner and provider/tool separation
remain OPEN. No mandatory safety bypass/product activation flag was introduced.


## Strict tool input encoding — 14 October 2026

Active implementation **07:44:23 EDT** through final verification **07:45:41 EDT**:
**1m 18s**, with documentation afterward. Full `make test` first **1.59s** and
`make check` **2.11s** failed a test assertion tied to escaped versus literal U+FFFD
marshal representation; the check now asserts semantic replacement independently
of Go version. Final full `make test` **0.53s**, full `make check` **4.11s**, including
race/static/UI and encoding/malformed input regressions (unchanged packages cached).
Updated minimal-source offline Linux cross-build **0.42s**, binary not executed.
Synthetic malformed text, scopes and private temp files only; no DB/provider/Docker
or credentials. Required free/core input validation has no off switch. Actual tool
runtime/transport and production credential separation remain OPEN.


## Helper result admission — 14 October 2026

Heartbeat/active start **07:48:15 EDT**, final verification **07:51:00 EDT**:
**2m 45s**, with documentation afterward. Full `make test` initial/final **1.04s**;
full `make check` **4.57s**, including race/static/UI, strict result consistency,
false-success/fencing mocks and unsupported declared-operation refusal. Unchanged
Go packages cached; synthetic file payloads/scopes and daemon/process mocks only.
No real DB/provider/Docker/image or credential interaction. Consistent result
metadata is not filesystem attestation. Actual runner/engine-pair supervision and
production credential separation remain OPEN. Required free/core result checks
have no bypass flag or deployment change.


## Legacy credential-bearing development admission — 14 October 2026

Task/research start **09:28:22 EDT**, final verification **09:39:49 EDT**:
**11m 27s**, including official interface research and local CLI metadata/schema
inspection. First full `make test` **2.32s** failed a superseded test that expected
legacy development acceptance; it now verifies mandatory refusal. Passing runs
**2.41s** and final **2.37s**; `make check` **6.93s** and final **7.22s**. Synthetic
requests, mock DB and metadata-only CLI inspection; no provider turns, credential
contents, live DB or deployment. The replacement bridge is not implemented by this
slice and old deployed workers/jobs are unchanged. Required free/core safety has
no flag or owner/paid bypass.

## Reviewed Codex build preparation — 14 October 2026

Recorded window **10:15:58–10:37:29 EDT: 21m 31s** through successful Linux router
tests. Initial repository/worker inspection before the first timed checkpoint is
excluded; documentation/staging afterward is excluded. This is a tracked window,
not a claim that the whole credential-separation implementation is complete.

Full `make test`: initial **0.37s**, final **0.40s** (Go results cached). Full
`make check`: initial **3.21s**, final **2.91s** including eight new offline Python
tests (**0.004s**). DB interactions mocked; no live provider turns/credentials used.

The successful isolated Linux library verification ran **3m 45s** including startup;
Cargo reported **3m 39s** compilation, with **3/3 router tests passing in <0.01s**.
It reused dependencies compiled in earlier attempts. The recorded window includes
the initial builder download, refusal of stale local workspace lock versions,
offline comparison/pinned normalization, correcting macOS archive metadata that
the migration compiler mistook for SQL, and narrowing the cancelled full-target
build to library tests. No migration was run and no database was created/accessed.
Unused upstream native-handler warnings remain visible. Release binary, OAuth,
paired execution/audit and production integration remain unverified/unimplemented.

Through verification-binary metadata checks, the recorded window extends to
**10:52:45 EDT: 36m 47s** from the same 10:15:58 checkpoint. The unoptimized CLI
build itself took **7m 52s** including container startup (Cargo **7m 49s**), reusing
the dependency cache. The resulting binary passed version/schema-generation
commands with network disabled, read-only root and an empty temporary home. These
commands start no model turn. The development binary is not a release/deployment;
OAuth and production paired execution remain unverified.


## Shared P0 integration — 15 October 2026

The shared development baseline at `405c89e` initially did not compile. After
launch-signature repair and correction of a formatting-sensitive audit test,
full `make test` passed in **2.44s** and full `make check` in **7.66s**.
Hardened audit integration and asynchronous initiator preservation then passed
in **2.62s / 7.38s**. Final runs after binding broker authority to worker
cancellation passed in **2.20s / 7.28s** respectively. Times are wall time around
each complete command, not estimates of total active authoring time. Unchanged
Go packages were cached; server tests ran. `make check` completed its full race,
vet, Python and JavaScript checks. Synthetic tests used mocked database and
executor interactions; no provider agents, databases or real Docker workloads
were launched. Production paired runtime remains unimplemented/unverified; see
[P0 progress](p0-execution-progress-2026-10-15.md).


## Paired lifecycle continuation — 15 October 2026

Version-3 worker/controller journalling and cancellation were developed with
synthetic temporary-file fixtures and mock Docker/database boundaries. Final
full `make test` passed in **3.04s**, full `make check` in **7.69s**. Server and
room-domain tests ran; unchanged Go packages were cached. Earlier complete
milestones passed in **2.29s / 7.17s** and **2.54s / 7.74s** as downgrade and
controller-recovery coverage were added. Times measure each complete command;
total active development time was not instrumented. No live provider, database
or Docker workload was launched. Paired launch and provider verification remain
open; see [the current component contract](worker-pairs.md).

### Architecture deepening and pair resource gate — 15 October 2026

Full command wall times measured with a monotonic clock: `make test` **2.92s**;
`make check` **6.36s**. Server tests ran; unchanged Go packages and the toolbroker
package tested immediately beforehand were cached in the final normal suite.
Full race/vet/Python/JavaScript verification passed. Tests cover lifecycle ordering,
cancellation, uncertain cleanup, mocked transport/audit failure and reported pair
resource rejection. No live Docker/provider/database verification was performed.

### Constrained tool launch — 15 October 2026

Monotonic full command wall times: `make test` **3.29s**; `make check` **8.66s**.
Server, tool-helper and toolbroker tests ran; unchanged room/strictjson/workspace
packages were cached in the normal suite. Full race/vet/Python/JavaScript checks
passed. Runtime scenarios use mocked Docker inspection/creation/start responses;
no live daemon/provider/database is exercised. Tests cover unsafe or missing
volume policy, adoption refusal, original cancellation/revocation, immutable IDs,
changed policy and refusal to replay uncertain startup.

### Credential-free workspace bootstrap — 15 October 2026

Monotonic full command wall times: `make test` **3.43s**; `make check` **8.31s**.
Server and broker normal tests ran; already-tested bootstrap/helpers and
unchanged domain/workspace packages were cached. Full race/vet/Python/JavaScript
verification passed. Tests use temporary files and mocked Git/Docker/database
interactions, including invalid snapshots/trees/receipts, output failure, blocked
input cancellation, readiness gating and engine-start ordering. No actual Git
clone, provider, Docker workload or database was used. These timings do not
measure image builds, Git checkout or live rootless runtime performance.

### Workspace volume lifecycle — continuous P0 run, 15 October 2026

Full command wall times measured monotonically: `make test` **3.27s**;
`make check` **8.48s**. Server tests ran; previously tested broker and unchanged
packages were cached in the normal suite. Full race/vet/Python/JavaScript checks
passed. Tests mock every Docker/database operation and use temporary journals and
local synthetic leases. No actual volume, provider or database was accessed.

### Shared tool session and timed-run closeout — 15 October 2026

Shared-session full suites passed: `make test` **3.16s**, `make check` **8.92s**.
Final closeout full suites passed: `make test` **0.69s**, `make check` **3.19s**.
All normal Go test packages were cached at closeout; full static/Python/JavaScript
checks still ran. Earlier uncached runs verified allocation/bootstrap, immutable
execution identity and framed durable audit with mocked Docker/database
interactions. No live provider, Git clone, Docker workload or database was used.
