# Development guide

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


[Documentation index](documentation-index.md)

## Get running

Use [installation](installation.md) for dependencies, database setup, owner sign-in and local startup. Runtime needs Go, PostgreSQL, and a browser; Node.js is needed for JavaScript checks and Python 3.8+ for offline engine build-policy tests (standard library only). Gitleaks is needed for secret checks. SSH is needed for approved remote-worker integration. Provider CLIs run inside reviewed images; installing them on the controller is not required.

## Branch and commit policy

Do development on a feature, fix, or documentation branch. Do not push changes directly to main. Submit a pull request for review and merge through the repository's agreed process. The project owner signs their own commits; tooling must not create commits on their behalf.

Keep changes focused. A useful description explains the trigger/problem, resulting behavior, verification, and remaining limits. Preserve unrelated local changes. Never use live history or production credentials as fixtures.

For ongoing P0 work, use the [batch and review workflow](p0-batch-workflow.md).
Related scenarios share a publication checkpoint; scheduled follow-ups read PR
feedback and resume after merge. Mark retains signing, publishing and merging.

## Common commands

```sh
make run
```

Run the server from the repository root. In a separate terminal, use the appropriate checks:

```sh
make check
make secrets
make build
```

`make test` runs fast Go tests. `make check` adds race detection, Go vet, Python build-policy tests and JavaScript/bootstrap/UI/Help tests. `make coverage` writes an ignored `coverage/go.out` report. Go tests run in parallel with a 90-second aggregate timeout per package; `TEST_PARALLEL` defaults to 8. The server package has more than 650 test functions, including durable temporary-file tests; its previous 30-second aggregate limit expired despite passing race partitions. Individual test and production-operation deadlines remain independent and are not extended by this ceiling. The dedicated P0 acceptance targets retain their 30-second bounds. Use `go test -count=1 -shuffle=on ./...` to check test-order independence.

All database interactions in tests are mocked in memory through `eventDatabase`. Tests never connect to PostgreSQL, create schemas, or insert database records; database environment settings are ignored. The mock covers query/scan/stream errors, transaction failures, rollback, serialization of concurrent audit appends, state projection, and HTTP authorization/validation. The invitation flow now runs on every test invocation rather than skipping behind an integration setting.

Tests inspect provider command arguments without executing CLIs or SSH. Each test owns its mock/authentication state and uses temporary directories for file fixtures. Production still uses the PostgreSQL pool through the same interface. The suite does not prove PostgreSQL canonicalization, migrations, cross-process advisory locking, or live provider compatibility; those need a separately designed deployment verification process.

`make build` writes ignored `bin/commonroom`. `make secrets` runs redacted Gitleaks scanning separately. The staged-content pre-commit hook is described in CONTRIBUTING.md.

## Where changes belong

- Domain shapes and capability declarations: internal/room/model.go.
- HTTP validation, access, event projection, adapters: cmd/server/.
- Persistence schema: migrations/; review effects on existing immutable history.
- Browser behavior: web/; preserve script loading order and shared-state expectations.
- User-facing or technical behavior changes: README and relevant docs/ guides.

For every mutation, enforce authorization and validation on the server. Do not rely on a hidden button, title, prompt, challenge level, or seniority value to authorize an action. Native agent context is reference material, not tool permission.

## Testing expectations

Choose tests that verify meaningful boundaries or behavior. Authentication, actor/room permissions, revision conflicts, parsing, adapter invocation, and integrity changes need focused coverage. Use fabricated credentials and isolated fixtures. Review UI changes in the browser without leaking private messages or keys in screenshots.

Do not describe a syntax check as end-to-end validation. Document unverified database/provider integration, unsupported transports, disabled workflow execution, and remote cancellation limits when relevant.

## Definition of complete

New features need accompanying automated tests, full usage/configuration/technical documentation, and a documented on/off feature flag with server enforcement and tests in both states. See [Feature completion rule](../../../CONTRIBUTING.md#feature-completion-rule). Mandatory isolation and infrastructure/test verification are exempt from optional product toggles. Feature flags must never weaken authentication, authorization, audit integrity, retention or the execution boundary.

## Documentation conventions

Keep guides focused on one reader task and separate usage from implementation
contracts. Explain current support before showing setup steps; identify pending
acceptance instead of presenting a historical deployment as proof of current support.

Document Go and JavaScript boundaries where contracts matter: errors, side effects,
concurrency, persistence, authorization and ownership. Prefer explanations of decisions
over comments that restate code. Keep SQL/config/build ownership and safe-change
procedures alongside the relevant operational guide.

For the current release gates and pending acceptance cases, see
[release readiness](release-readiness.md) and [P0 acceptance tests](p0-acceptance-tests.md).
