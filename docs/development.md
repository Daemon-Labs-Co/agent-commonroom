# Development guide

[Documentation index](README.md)

## Get running

Use the main README for dependencies, database setup, owner sign-in, and local startup. Runtime needs Go, PostgreSQL, and a browser; Node.js is needed for JavaScript syntax checks. Gitleaks is needed for secret checks. Provider CLIs and SSH are optional until working on those integrations.

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

make check runs Go race tests, Go vet, and syntax checks for the application and help-reader browser scripts. make build writes ignored bin/commonroom. make secrets runs redacted Gitleaks scanning; it is separate from make check. The staged-content pre-commit hook is described in CONTRIBUTING.md; custom existing hooks must be preserved.

Database integration tests require a separate disposable database:

```sh
createdb commonroom_test
export COMMONROOM_TEST_DB='postgres://localhost:5432/commonroom_test?sslmode=disable'
go test -race ./...
```

Without COMMONROOM_TEST_DB, the relevant integration test skips. Current member integration tests create/remove isolated schemas; do not point this setting at a live database.

## Where changes belong

- Domain shapes and capability declarations: internal/room/model.go.
- HTTP validation, access, event projection, adapters: cmd/server/.
- Persistence schema: migrations/; review effects on existing immutable history.
- Browser behavior: web/; preserve script loading order and shared-state expectations.
- User-facing or technical behavior changes: README and relevant docs/ guides.

For every mutation, enforce authorization and validation on the server. Do not rely on a hidden button, title, prompt, challenge level, or seniority value to authorize an action. Native agent context is reference material, not tool permission.

## Testing expectations

Choose tests that verify meaningful boundaries or behavior. Authentication, actor/room permissions, revision conflicts, parsing, adapter invocation, and integrity changes need focused coverage. Use fabricated credentials and isolated fixtures. Review UI changes in the browser without leaking private messages or keys in screenshots.

Do not describe a syntax check as end-to-end validation. Document skipped integration tests, unsupported transports, disabled workflow execution, and remote cancellation limits when relevant.

## Documentation conventions and next passes

This foundation does not add blanket source comments or a generated technical API reference. The planned next passes should follow the language actually used:

1. Go: document package responsibilities, exported structs/interfaces/functions, receiver methods, and important internal boundaries. Describe errors, side effects, concurrency, persistence, and security requirements where meaningful.
2. JavaScript: use JSDoc for callable entry points and shared state; document dependencies between scripts and wrappers. Do not invent classes where the code uses functions and globals.
3. Technical file reference: for each file record purpose/scope, dependencies, entry points, relevant methods/functions, variables/properties and their lifetimes, side effects, tests, and known limitations.
4. SQL/config/build assets: document their operational effects, ownership/permissions, and safe change procedure.

Prefer comments explaining contracts and decisions over restating obvious code. Keep source references and guides current with implementation changes. CONTRIBUTING.md and SECURITY_REVIEW.md provide additional constraints and backlog context.
