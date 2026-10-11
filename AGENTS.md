# Agent CommonRoom development rules

- Work on a feature/fix/documentation branch. Never commit or push to main. The owner signs commits manually; do not create commits or push on their behalf.
- Preserve unrelated edits. Never use live messages or credentials as test fixtures.
- A feature is complete only with meaningful automated tests, full user and technical documentation, and a documented on/off feature flag. Follow the feature completion rule in CONTRIBUTING.md.
- Tests must be fast and parallel-safe. Mock every database interaction; do not connect to a database, create schemas, or insert test data. Use temporary directories for file fixtures. Do not launch real provider agents in automated tests.
- Feature flags must have server enforcement, permission checks, documented defaults, and tests for both states. They must not bypass authentication, authorization, audit integrity, or retention safeguards. Infrastructure/test changes do not need a product toggle for verification itself.
- Run make test during development and make check before completion. Report meaningful remaining coverage or integration gaps honestly.
