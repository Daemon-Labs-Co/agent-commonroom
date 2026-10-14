# Development guide

[Documentation index](README.md)

## Get running

Use the main README for dependencies, database setup, owner sign-in, and local startup. Runtime needs Go, PostgreSQL, and a browser; Node.js is needed for JavaScript checks and Python 3.8+ for offline engine build-policy tests (standard library only). Gitleaks is needed for secret checks. Provider CLIs and SSH are optional until working on those integrations.

## Branch and commit policy

Do development on a feature, fix, or documentation branch. Do not push changes directly to main. Submit a pull request for review and merge through the repository's agreed process. The project owner signs their own commits; tooling must not create commits on their behalf.

Keep changes focused. A useful description explains the trigger/problem, resulting behavior, verification, and remaining limits. Preserve unrelated local changes. Never use live history or production credentials as fixtures.

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

`make test` runs fast Go tests; `make check` includes race detection, Go vet, browser-script syntax checks, and help-reader security tests. `make coverage` writes an ignored `coverage/go.out` report. Go tests run in parallel with a 30-second timeout; `TEST_PARALLEL` defaults to 8. Use `go test -count=1 -shuffle=on ./...` to check test-order independence.

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

New features need accompanying automated tests, full usage/configuration/technical documentation, and a documented on/off feature flag with server enforcement and tests in both states. See [Feature completion rule](../CONTRIBUTING.md#feature-completion-rule). Existing unflagged features remain a follow-up; this policy does not imply a feature-flag control panel has been implemented.

## Documentation conventions and next passes

This foundation does not add blanket source comments or a generated technical API reference. The planned next passes should follow the language actually used:

1. Go: document package responsibilities, exported structs/interfaces/functions, receiver methods, and important internal boundaries. Describe errors, side effects, concurrency, persistence, and security requirements where meaningful.
2. JavaScript: use JSDoc for callable entry points and shared state; document dependencies between scripts and wrappers. Do not invent classes where the code uses functions and globals.
3. Technical file reference: for each file record purpose/scope, dependencies, entry points, relevant methods/functions, variables/properties and their lifetimes, side effects, tests, and known limitations.
4. SQL/config/build assets: document their operational effects, ownership/permissions, and safe change procedure.

Prefer comments explaining contracts and decisions over restating obvious code. Keep source references and guides current with implementation changes. CONTRIBUTING.md and SECURITY_REVIEW.md provide additional constraints and backlog context.
