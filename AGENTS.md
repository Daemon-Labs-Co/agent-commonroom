# Agent CommonRoom development rules

- Work on a feature/fix/documentation branch. Never commit or push to main. The owner signs commits manually; do not create commits or push on their behalf.
- Preserve unrelated edits. Never use live messages or credentials as test fixtures.
- Documentation changes must follow [the audience contract](docs/technical/documentation-policy.md): illustrated functionality-based user guides with small Bash helpers; technical reference limited to inputs, outputs, functionality and property/state changes. Archived records retain historical content.
- A feature is complete only with meaningful automated tests, full user and technical documentation, and a documented on/off feature flag. Follow the feature completion rule in CONTRIBUTING.md.
- Tests must be fast and parallel-safe. Mock every database interaction; do not connect to a database, create schemas, or insert test data. Use temporary directories for file fixtures. Do not launch real provider agents in automated tests.
- Feature flags must have server enforcement, permission checks, documented defaults, and tests for both states. They must not bypass authentication, authorization, audit integrity, or retention safeguards. Infrastructure/test changes do not need a product toggle for verification itself.
- Run make test during development and make check before completion. Report meaningful remaining coverage or integration gaps honestly.
- For ongoing P0 development, review feedback, or scheduled continuation, follow [the batch workflow](docs/technical/archive/p0-batch-workflow.md) and update [its checkpoint](docs/technical/archive/p0-batch-status.md).

## Mandatory execution boundary

Container isolation for managed agent development is a required security boundary, exempt from the optional-feature flag rule. Do not add an off switch, per-connection host-execution exception or automatic uncontained fallback. Document the contract and source-level bypass risks at adapter types and launch methods. Managed model turns must use the isolated worker. Do not reintroduce direct local/SSH launches; preserve the documented platform restrictions and disclose live-runtime verification gaps. See [the architecture contract](docs/technical/archive/architecture.md#mandatory-container-isolation-design-contract).
