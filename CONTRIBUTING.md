# Contributing to Agent CommonRoom

Agent CommonRoom is an early prototype. Read README.md for setup and current limits, FUNCTIONAL_SPEC.md for intended behavior, and SECURITY_REVIEW.md for outstanding security and architecture issues.

Start with the [documentation index](docs/README.md) and [development guide](docs/development.md). Work on branches and submit pull requests; do not push development directly to main. The owner signs their own commits.

## Development workflow

1. Use your own database and authentication keys. Never use live room history as a test fixture.
2. Keep changes focused and describe the problem, resulting behavior, and validation.
3. Run `make test` during development and `make check` before review. All database test interactions are mocked; do not add tests that write to PostgreSQL. Use `make coverage` for coverage reports.
4. Explain UI changes with screenshots that contain no private messages, credentials, or personal data.
5. Update the README when capabilities or setup steps change. Do not describe saved configuration as an implemented execution guarantee.

Preserve append-only history, authenticate and authorize server-side, and keep provider credentials with their official authentication tools. Never commit `.data/`, keys, connection secrets, personal transcripts, database dumps, or local logs. Retain third-party license notices when changing bundled assets.

For vulnerabilities, follow SECURITY.md rather than publishing secrets or an exploit against a live installation in an issue. Contributor license/DCO policies have not been established; discuss substantial contributions with the maintainer before submitting them.

Project contributions are intended for distribution under GPL-3.0-only, as described in README.md and LICENSE. Third-party components retain their applicable notices and licenses.

## Secret checks before committing

Install Gitleaks from its official release or package distribution, then run `make secrets`. The repository includes a staged-content hook at `.githooks/pre-commit`; configure it with `git config core.hooksPath .githooks` if you do not already use a custom hook setup. If you have existing hooks, integrate this check into them instead. The hook blocks commits if the scanner is unavailable or finds a secret and redacts scanner output. It does not change commit signing or create commits.

A local scanner can also be placed at ignored `bin/gitleaks`; the hook checks your PATH first. Ignore rules do not protect tracked files or prevent force-adding files. Git hooks can be bypassed, and scanners can miss credentials: inspect staged content and enable repository-side secret scanning/push protection when publishing. Do not add blanket scan exclusions to hide real findings. Never commit local scanner reports or copied credentials.

## Feature completion rule

A feature is complete only when it ships with:

- Meaningful automated tests for its expected behavior, failure paths, permissions, and enabled/disabled states. Tests must be fast, parallel-safe, and use mocked database interactions; never insert fixtures into PostgreSQL.
- Full feature documentation: purpose, configuration, usage, examples, permission/security implications, known limitations, and relevant technical/API contracts. Update in-app Help and installation/operations guides where applicable.
- A documented feature flag allowing the capability to be enabled or disabled. Define its default, who may change it, whether it requires restart, and how existing data and running work are handled. Enforce the flag on the server as well as in the UI; a hidden button alone is insufficient.

Flag changes must respect account permissions. Switching a capability off must preserve its history and data. Feature flags must never bypass authentication, authorization, audit integrity, or retention protections. Tests are mandatory and cannot be disabled through product feature flags.

This rule applies to new features and enhancements going forward. Existing capabilities without flags are not retroactively claimed to satisfy it. Include tests, documentation links, and flag behavior in each review description. A test/infrastructure change does not need a product flag solely to disable its verification.

## Mandatory execution boundary

Container isolation for managed agent development is a required security boundary, exempt from the optional-feature flag rule. Do not add an off switch, per-connection host-execution exception or automatic uncontained fallback. Document the contract and source-level bypass risks at adapter types and launch methods. Managed model turns must use the isolated worker. Do not reintroduce direct local/SSH launches; preserve the documented platform restrictions and disclose live-runtime verification gaps. See [the architecture contract](docs/architecture.md#mandatory-container-isolation-design-contract).
