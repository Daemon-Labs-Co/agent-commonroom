# Release readiness

Reviewed against the source on 19 October 2026. This page describes current
capability and remaining gates; [FUNCTIONAL_SPEC.md](../FUNCTIONAL_SPEC.md)
defines intended behavior. Passing component tests does not mean a feature is
available through the application or verified against a real provider.

## Available foundations

Human accounts, room access, conversations, agent profiles, tasks, revisioned
memory, workflow editing and sequential workflow orchestration exist. Owner
controls, emergency stops, context disclosure, append-only audit and explicit
job recovery remain mandatory around agent work. Read the [product guide](concepts.md)
for usage and [security review](../SECURITY_REVIEW.md) for current limits.

Agent chat and sign-in have Codex/Claude paths through an approved Linux rootless
worker. Other provider labels or transport registrations do not establish an
executable adapter. A connection test is not proof of code-development capability.

## Internal execution path

The new Codex path connects a credential-free tool container to a separate provider
engine. It includes fresh pair/workspace creation, scoped callback authority,
private audit transport, bounded read/write tools, bootstrap, credential framing,
private gateways, crash-aware credential leases, private final-credential export
and persistence after durable pair fencing.

Combined synthetic tests exercise callback success, rejected capabilities, audit
failures, disconnects and cleanup. Four physical container slots are reserved
across the account journal; each engine/tool pair costs two. Unresolved allocations
retain their reservations. Source is retained unless separate disposal authority
is established.

These are internal components. RunCodex currently has test callers, not a public
production dispatch path. The legacy development barrier remains in force.
See [engine runtime](engine-container-runtime.md), [paired lifecycle](worker-pairs.md)
and [container quotas](container-quotas.md).

## P0 gates still open

| Gate | What is missing |
| --- | --- |
| Production dispatch | Authorized controller/worker assembly and live admission for repository-bound development. |
| Required providers | Reviewed isolated development paths and acceptance for Claude, Gemini, GitHub Copilot and llama.cpp; live Codex acceptance. |
| Credential renewal | Durable uncertainty blocking and explicit sign-in recovery exist. Private final export and fenced, connection-scoped persistence are implemented internally. Reviewed image, live refresh and recovery acceptance remain. |
| Storage | Durable 5,000,000,000-byte reservation accounting is implemented internally. Complete physical inventory, hard worker enforcement and allocation/disposal integration remain. |
| Harness upgrades | Signed operator-review admission and immutable Codex image selection now exist internally, with rollback, startup binding and exact staged repository-digest checks. An offline tool generates review keys and exclusively publishes signed review envelopes. Upstream discovery, real compatibility review, approved image publication/staging, production launch integration and live acceptance remain. See [release admission](harness-release-admission.md). |
| Idle agents | Internal worker guard now defaults to five minutes with an administrator environment fallback, busy/queued-work protection and retained source. Owner settings and immutable allocation snapshots now exist. Internal pre-cleanup controller acknowledgments and room notices now exist. The private audit-link acknowledgment path is implemented internally. Production launch/policy assembly and live acceptance remain. |
| Commands and artifacts | Bounded isolated builds/tests and permission-checked, revision-bound observed evidence/artifact delivery. |
| Approved disposal | Exact internal request/receipt volume binding now exists. Owner review, one-minute delay, expiry/revocation and one-shot claims are implemented off by default. Production storage allocation binding, artifact integration and live approval acceptance remain. |
| Live acceptance | Reviewed image build, rootless enforcement, transport interruption, crash recovery and real provider round trips. |

The [remaining acceptance guide](p0-acceptance-tests.md) maps these gates to 74
prepared scenarios. Four credential-custody, seven storage-accounting, nine disposal, one idle-permission, six idle lifecycle protection/acknowledgment, three command admission refusal, one command outcome failure, one command output-limit, one command timeout and four harness-review refusal scenarios plus one active-agent upgrade stability scenario now
have production-backed hermetic drivers; 36 scenarios remain pending and are skipped by normal tests. `make test-p0-ready`
still fails for those missing drivers. Neither a catalog nor a mock is live acceptance.

## After P0

P1 includes governing-agent appointment/replacement, bounded delegation,
structured disagreement/dependencies, comprehensive decision/action attribution
and safe automatic recovery. Existing workflows and manual reconciliation do not
complete those features. The [backlog](planned-features.md) describes follow-up work.

Secret exposure protection and the managed key/value editor/export are requested
follow-up requirements (FR-25–28), not delivered capabilities. MCP/skills product
execution is not implemented; the UI must label it coming soon. Paid consensus, multiple-human authority, Fulginn,
Veriscribe and enterprise compliance work remain separate. No SOC 2 compliance
claim is made.

## Deployment boundary

Use one controller with a dedicated restricted PostgreSQL runtime role. Supported
installation is local/private, using numeric loopback and approved SSH transport.
Public hosting, shared multi-controller ownership and hostile multi-tenancy are
not ready. Linux rootless isolation, authentication, authorization, audit integrity
and retention have no owner, paid or feature-flag bypass.

For a release decision, run the required automated checks and record separate
live acceptance evidence using synthetic repositories and dedicated accounts.
Do not reuse old timing logs or a previous binary's metadata check as proof of
the current deployment.

## Internal bounded-command increment

[Isolated Go commands](isolated-commands.md) now provide two opt-in, explicitly
granted terminal command profiles, bounded trusted output capture and exact tool
container absence before receipt release. Default sessions remain file-only.
This is synthetic-test evidence, not live command/image acceptance. Public
production dispatch, current-revision artifact attestation, other provider paths
and hard aggregate storage enforcement remain blocked; no P0 gate is closed solely
by this increment.

## Internal storage-accounting increment

[Storage accounting](storage-accounting.md) now persists account-wide capacity
reservations with checked arithmetic and optimistic journal revisions. Unknown
inventory blocks reservations, retained data stays charged after container
shutdown, and there is no owner/paid/feature-flag override. This does not yet
physically constrain Docker volume writes. Complete inventory, a reviewed physical
quota backend, production allocation integration and live disposal remain open.

## Initial live tool evidence

A dedicated Linux rootless worker now has [tool-only live evidence](live-tool-acceptance.md):
the minimal tool image builds, a cold synthetic Go test succeeds under the fixed
limits, kernel cgroup values are observed, and fresh diagnostic containers/volumes
are removed with absence checks. No provider, controller audit transport or
production dispatch was exercised; the broader live acceptance gate stays open.
The cold test took about 107 seconds of the 120-second command allowance.

## Disposal binding hardening

The [internal disposal contract](workspace-disposal-binding.md) now requires the
complete reviewed volume identity in both request and successful receipt. Stale
bindings are rejected before effects and rechecked under creator ownership. Crash
discovery retains data until the complete identity can be reviewed separately.
The owner approval flow now uses this contract with exact prebound storage-credit
release. Production allocation binding and live approval acceptance remain open,
so the disposal P0 gate is not complete.

## Owner disposal approval increment

[Owner-reviewed workspace disposal](workspace-disposal.md) now provides an
off-by-default proposal/approval/revocation flow. It enforces a full review minute,
one-hour expiry, fresh stopped-worker identity, exact confirmation and a durable
one-shot claim. Verified outcome, container-slot release and exact prebound
workspace credit release are atomic. Unbound charges remain. Unknown outcomes
stay blocked. All nine disposal catalog drivers pass, including removal/absence failures,
stale approval, retained-source cancellation and audit-history protection.
Production allocation and live deployment acceptance remain pending.
