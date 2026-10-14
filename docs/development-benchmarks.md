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
