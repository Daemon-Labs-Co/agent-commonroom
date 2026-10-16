# Security policy

Agent CommonRoom currently supports local development only. It is not hardened for internet exposure or hostile multi-user deployments. Known findings and required fixes are recorded in SECURITY_REVIEW.md. There is no supported stable release series yet.

## Reporting a vulnerability

Do not include private keys, invitation files, personal conversations, or real credentials in public issues. A private reporting channel has not yet been configured. Before publishing this repository, the maintainer should enable GitHub private vulnerability reporting or provide a private security contact here.

Once a private channel is available, include the affected version, reproduction steps against an isolated test environment, impact, and any proposed mitigation. Do not test other users' installations without their permission. Until then, request a private reporting channel without publicly disclosing exploitation details.

## Protect local data

Keep `.data/`, downloaded owner/member keys, provider credentials, local configuration, and database backups private. Automated tests must use synthetic credentials and mocked database interactions; never connect them to a database or launch real provider agents. Do not expose the server through a tunnel or proxy; loopback-only binding is a development boundary, not production authentication or transport security.

PostgreSQL append-only triggers and the local checksum chain do not resist a privileged database administrator. New records protect identity, sequence and commit-time metadata, but external immutable storage and rollback anchors remain pending. Incident cancellation is implemented; verified physical shutdown under failures, approved disposal and complete secure retention/purge still need acceptance.

## Repository secret checks

Run `make secrets` before sharing source. A staged-content Gitleaks pre-commit hook is provided in `.githooks/`; installation and custom-hook handling are documented in CONTRIBUTING.md. The ignore rules cover local state, private keys, invitations, credential stores, and common backup formats while retaining source migrations and `.env.example`. These are layered checks, not a guarantee: hooks can be bypassed and unknown secret formats can evade scanners. Use host-side secret scanning/push protection too. If a real credential enters published history, revoke/rotate it; removing it from a later commit is insufficient.
