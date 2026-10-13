# Security hardening and operational limits

This update strengthens required security boundaries; it does not make Agent
CommonRoom ready for untrusted public hosting or claim the security backlog is
closed. See [database setup and audit recovery](database-security.md).

## Identity, load and context

Authenticated API actors must be enabled humans. Agent API credentials are
unsupported and rejected, including read requests. Authorization and room access
still apply independently. Audit appends overwrite client-provided `performedBy`
with the authenticated actor ID. Human and agent messages have explicit identity
badges; typed agents named "You" cannot receive human styling. Legacy messages
may have inferred attribution rather than a recorded authenticated identity.
Display names never confer permissions.

Human API mutations have a per-actor token bucket: 20 requests of burst capacity,
one token replenished per second, and a maximum of 4,096 tracked actors. Idle
entries can be reclaimed. Exhaustion returns HTTP 429 and `Retry-After: 1`.
GET/HEAD requests do not consume this budget; login already has separate controls.
The owner's emergency agent-safety endpoint bypasses the mutation budget so load
cannot prevent a stop. All other permissions remain enforced. Budgets apply
across rooms/sessions within one controller, not across multiple controllers.

Chat prompts have a 128 KiB assembled context budget. Profile instructions,
reference history, scoped shared memory, recent room messages and the request
count toward it. Oversized broadcasts fail with HTTP 413 before any member is
launched or a message is appended. Reduce the request or reference context rather
than silently dropping it. Shared memory is explicitly identified as untrusted
reference material. Incident/emergency stops also block native conversation
listing/import before provider access. These controls are mandatory, with no
feature flag or off switch.

## Worker startup and emergency resume

With the supported remote worker enabled, controller startup checks the approved
worker's `idle-v1` state. Active or unknown state latches an application stop and
records the reason; it does not kill, adopt or replay unknown jobs. A failed stop
append also latches an integrity stop. The owner must inspect worker jobs and
explicitly resume after verified shutdown. A temporary SSH failure can therefore
require owner recovery even if no job is running; fail-closed is intentional.

Emergency resume requires a freshly valid audit chain, no integrity incident,
no active controller turns and a fresh shared-worker idle confirmation. Unknown
worker state leaves the stop in place. Application-wide stop cancels workflow
contexts directly even during a database read outage. It still does not guarantee
remote termination when SSH/cleanup fails; inspect background work and worker
jobs. These mandatory checks have no flag and do not replace durable job leases
or reconciliation after transport loss. Local Linux-worker startup reconciliation
has not been validated; the live deployment uses the approved Debian worker.

## Dependency checks

Updated pgx and Go text dependencies address the reachable advisories found by
Go's vulnerability scanner. `go.mod` requires Go 1.26 or later. The verification
used Go 1.27.2 and found no remaining reported reachable Go vulnerabilities at
scan time; this does not certify provider images, npm packages or OS components.
Run an official `govulncheck ./...` separately from the fast offline test suite,
and scan reviewed container images before deployment. Source secret scans use
redacted output and exclude private ignored state; absence of findings is not a
guarantee that every secret has been detected.

## Deployment and remaining work

Restarting is not sufficient to enable this update: provision a restricted
runtime database role first. Migrations are now explicit (`--migrate` with
`MIGRATION_DATABASE_URL`); startup does not perform schema DDL. Verify idle remote
workers before controller/worker replacement. Do not copy credentials into test
fixtures or disable container isolation to diagnose failures.

Remaining work includes independently protected audit anchoring, durable job
leases and orphan reconciliation after controller failure, confirmed cancellation
under transport loss, multi-controller quotas, spend/storage limits, disclosure
policies for untrusted references, and provider/container dependency audits.
Multiple companies and public remote deployment are not supported isolation
boundaries. Authentication, loopback restrictions, container isolation, audit
checks and retention cannot be bypassed by feature flags.

Automated coverage uses only synthetic records, fake provider callbacks and
mocked database interactions. It covers tampering, metadata changes, outage
cancellation, restart stop markers, explicit owner recovery and failure paths,
privilege refusals, context limits, rate budgets and message identity. Live
PostgreSQL role configuration and provider execution require separate deployment
verification; automated tests must never insert data in the real database.
