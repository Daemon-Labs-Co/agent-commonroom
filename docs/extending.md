# Extending Agent CommonRoom

[README](../README.md) · [Architecture](architecture.md) · [Source reference](reference/README.md)

Extend CommonRoom when you need a provider, organization workflow, tool or interface
that the current application does not support. Start with the functional spec and
release-readiness guide: intended behavior, saved configuration and executable
capability are different things. There is no public runtime plugin loader yet.
MCP, skills and paid integrations remain planned; bundled `.agents/skills` guide
contributors and are not installed agent capabilities.

## Find the right seam

| Change | Start here | Contract to preserve |
| --- | --- | --- |
| Room data or task behavior | `internal/room/`, `cmd/server/task_dispatch.go` and related server modules | Revisions, permissions, append-only events and attribution. |
| API or configuration | `cmd/server/`, [API](api.md), [server structure](server-structure.md) | Authenticate first, authorize server-side, strict input and durable audit. |
| Provider engine | `internal/codexbridge/`, `internal/toolbroker/`, `docker/engine/` | Credential/tool separation, reviewed adapter protocol and isolated execution. |
| Tools and workspaces | `internal/toolbroker/`, `internal/toolworkspace/`, `internal/workspacebootstrap/` | Bounded capabilities, path containment, quotas, audit and absence proof. |
| Browser UI | `web/`, matching `tests/*.test.cjs` | Permission-aware controls, accessibility, safe rendering and contextual Help. |
| Database changes | `migrations/`, storage code in `cmd/server/` | Explicit owner migration, restricted runtime grants and preserved history. |
| Documentation and Help | `docs/`, `cmd/server/documentation.go`, `web/help.js` | Allowlisted guides, current links and accurate readiness claims. |
| Future memory/MCP integrations | [Planned work](planned-features.md), [functional spec](../FUNCTIONAL_SPEC.md) | Attribution, disclosure, lifecycle and edition boundaries before implementation. |

Use the [file and API reference](reference/README.md) to locate exact declarations.
Go has structs/interfaces and receiver methods rather than classes. The reference
shows signatures, available contracts and direct calls, helping you trace changes
without reading a giant README. For architectural responsibility and rationale,
use [server structure](server-structure.md) and [design contracts](architecture.md).

## Add a provider without bypassing isolation

1. Define what the adapter can actually execute: authentication, conversation,
   development commands, cancellation and artifacts. A provider label is insufficient.
2. Define a versioned, bounded engine protocol and credential lifecycle. Implement
   the isolated adapter and approved immutable image rather than a host CLI launch.
3. Register capabilities only after worker admission, policy and permissions checks.
   Preserve emergency stops, audit delivery, credential separation and quotas.
4. Add mocked/hermetic success and failure tests, including disconnect, cancellation,
   audit failure, expired authority and cleanup uncertainty. Never launch real agents
   or connect to PostgreSQL in automated tests.
5. Record live acceptance separately before advertising a working provider. Update
   readiness, connection guidance and the source reference.

The [mandatory isolation contract](architecture.md#mandatory-container-isolation-design-contract)
has no feature flag or host fallback. A source fork can undermine it; extending the
adapter is not permission to remove it.

## Ship a feature

Work on a feature/fix/documentation branch and preserve unrelated changes. Follow
[CONTRIBUTING](../CONTRIBUTING.md): implement meaningful tests, user and technical
documentation, and a documented on/off flag with server enforcement, permission
checks, defaults and tests for both states. Infrastructure verification itself
needs no product toggle. Flags cannot disable authentication, authorization,
audit integrity, retention or mandatory isolation.

```sh
make test
make check
make build
go run ./tools/source-reference
```

Add test cases near the affected contract; keep fixtures synthetic and parallel-safe.
Update the source reference after adding or renaming files. The owner signs commits
manually; provide a complete proposed commit message rather than creating a commit.
