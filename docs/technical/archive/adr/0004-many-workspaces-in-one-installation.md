# Host many workspaces in one installation, isolated by construction and by the database

Status: accepted

CommonRoom must serve many customers as a hosted product. The owner chose one shared
installation (one process, one database) over one process per customer. Each
**workspace** therefore gets its own server instance, event stream, hash chain,
keys, sessions, files and limits, and PostgreSQL row-level security on `events`
(scoped by a per-transaction `app.workspace_id`) makes a forgotten filter in
application code return nothing instead of another workspace's history. An
installation with `COMMONROOM_TENANCY=single` (the default) is the same code with
one workspace named `default`, so existing installs and the free self-hosted
edition do not change.

## Considered Options

- **One process per customer.** Strongest blast-radius isolation and no code change,
  but a database, process and containers per customer. Not chosen; it stays possible
  because a workspace is the unit that can run alone.
- **Shared tables with `WHERE workspace_id` only in application code.** One missed
  clause leaks history. Rejected: the audit log is the product's trust anchor.
- **Schema or database per workspace.** Migrations and connection pools multiply with
  customers. Rejected for the first release.

## Consequences

- The global `events_sequence_seq` is shared, so a workspace can infer coarse activity
  volume from gaps in its own sequence numbers. Accepted; revisit with per-workspace
  sequences if it matters.
- A noisy workspace can starve others, so every workspace has request, turn,
  container and storage limits, and an integrity incident or emergency stop affects
  only its workspace. The platform operator keeps a separate stop for everything.
- The database migration (`migrations/002.sql`) is a privileged step the administrator
  runs; the restricted runtime role cannot apply it.
- Execution isolation (containers, credentials, quotas) is a separate boundary in the
  worker and must be namespaced by workspace before a hosted launch.
