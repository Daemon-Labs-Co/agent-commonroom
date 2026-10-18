# Release readiness

Reviewed against the source on 18 October 2026. This page describes current
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
| Storage | Hard aggregate 5,000,000,000-byte free-account enforcement across retained workspaces, staged bundles and artifacts. |
| Harness upgrades | Instantiation-time checks, compatibility review and approved immutable-image publication/selection. |
| Idle agents | Owner-configurable shutdown after inactivity, default five minutes, without interrupting active work or discarding source. |
| Commands and artifacts | Bounded isolated builds/tests and permission-checked, revision-bound observed evidence/artifact delivery. |
| Approved disposal | Exact human approval, volume identity checks, absence proof and correct storage-credit release. |
| Live acceptance | Reviewed image build, rootless enforcement, transport interruption, crash recovery and real provider round trips. |

The [remaining acceptance guide](p0-acceptance-tests.md) maps these gates to 74
prepared scenarios. Four credential-custody scenarios now have production-backed hermetic drivers;
70 scenarios remain pending and are skipped by normal tests. `make test-p0-ready`
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
