# Restricted audit database access

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


The controller must use a dedicated runtime login that cannot rewrite history or
remove its protections. This is mandatory, has no feature flag, and applies to
existing installations as well as new ones. Stop dispatch and verify remote
workers are idle before changing credentials or restarting. Back up the existing
database and private key/session files first. Never rewrite historical records
as part of this migration.

## Administrator setup

Use a dedicated CommonRoom database. The migration account owns its schema,
table, sequence and trigger function; the runtime account owns none of them and
has no role memberships. Do not reuse an existing application database. The
following names are examples; adapt them consistently. Run administrator SQL
through a DBA-approved connection, never through the web application:

```sql
CREATE ROLE commonroom_migrator LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE
  NOREPLICATION NOBYPASSRLS;
CREATE ROLE commonroom_runtime LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE
  NOREPLICATION NOBYPASSRLS;
CREATE DATABASE commonroom OWNER commonroom_migrator;
```

If objects already exist, retain their ownership and use the existing owner for
migration instead of recreating roles or the database. Verify that the proposed
runtime role owns no objects and has no memberships. Configure authentication
before connecting: prefer same-host Unix-socket peer authentication with matching
OS/runtime role names; for TCP use a DBA-configured SCRAM password and verified
TLS where the connection crosses hosts. Do not enable `trust`. Use psql's
interactive `\password` command to set a password without placing it in shell
history or SQL source. Store credentials in an approved private credential store
or PostgreSQL password file (mode 0600); never in Git, connection endpoints or
command-line arguments.

From the repository root, use the migration owner's approved URL and credentials:

```sh
# The URL contains no password; use peer authentication or a private passfile.
export MIGRATION_DATABASE_URL='postgres://commonroom_migrator@localhost:5432/commonroom?sslmode=disable'
go run ./cmd/server --migrate
unset MIGRATION_DATABASE_URL
```

The local `sslmode=disable` example is not suitable across hosts. The migration
command never falls back to runtime credentials. It uses `migrations/001.sql`;
review migration changes before running them. For upgrades retain the existing
owner rather than transferring unrelated objects.

After migration, connect to the dedicated database as its administrator and
apply the grants (use the actual runtime role name):

```sql
REVOKE ALL ON DATABASE commonroom FROM PUBLIC;
GRANT CONNECT ON DATABASE commonroom TO commonroom_runtime;
REVOKE CREATE ON SCHEMA public FROM PUBLIC;
GRANT USAGE ON SCHEMA public TO commonroom_runtime;
REVOKE ALL ON TABLE public.events FROM commonroom_runtime;
GRANT SELECT, INSERT ON TABLE public.events TO commonroom_runtime;
REVOKE ALL ON SEQUENCE public.events_sequence_seq FROM commonroom_runtime;
GRANT USAGE ON SEQUENCE public.events_sequence_seq TO commonroom_runtime;
```

Coordinate these revocations with the DBA if the database has unexpected users;
this recipe assumes a dedicated database. Do not grant UPDATE, DELETE, TRUNCATE,
TRIGGER, sequence UPDATE, schema CREATE, ownership, superuser, role membership,
replication, or bypass-RLS privileges to the runtime account. Avoid inherited
PUBLIC grants that restore these capabilities. Restrict network and `pg_hba.conf`
access to the intended service account and administrator channels. A narrow SQL
role does not compensate for a `trust` rule that lets callers impersonate owners.

Serve using `DATABASE_URL` for the runtime account. Startup pins
`search_path=pg_catalog,public` and validates effective privileges, including
inherited/public privileges. Failure prevents startup; there is no owner-role
fallback. Keep migration credentials out of the service environment and out of
its accessible files. For Debian peer authentication, the service role can be
named `commonroom` instead of `commonroom_runtime`, matching its Linux account.

## Integrity and recovery

Audit verification runs before serving and every 30 seconds, and before managed
execution. A verification failure cancels controller turns/workflows and latches
execution off. A private `integrity-stop` marker beside `OWNER_KEY_FILE` preserves
a stop across restart even if the database could not record the incident. Keep
this directory writable only by the service account; do not delete the marker
manually to resume agents. A marker persistence error is logged and requires
operator repair before restart.

An owner uses **Review integrity incident** to provide a reason. Recovery requires
a freshly valid chain, no active controller turn, confirmed shared-worker idle,
and a successful `incident.recovered` append. It clears operational stop state
only, never modifies history or restarts tasks. If corruption remains, preserve
evidence and investigate with the DBA; the UI cannot repair checksums. Storage
restoration or an optional feature toggle never resumes execution by itself.

New records include server-generated ID, sequence and UTC timestamp in their
checksummed `_auditRecord` payload. The verifier checks those values against row
columns and reports protected/legacy counts. Existing records remain unchanged;
legacy metadata has weaker protection. The authenticated human's ID overrides
client-supplied `performedBy`; agents cannot authenticate with human APIs.

## Limits and source-level bypass risks

These protections defend against ordinary application SQL writes and accidental
corruption. Checksums are unkeyed. A database/host administrator can recompute
chains, remove suffixes, replace backups, delete the marker or modify source.
Independent append-only anchoring and an externally controlled immutable archive
are still required for administrator-resistant rollback/tamper detection. A
same-account local archive would not provide that guarantee.

A source fork **CAN** bypass startup checks or change the writer; doing so is a
bad idea because compromised runtime credentials could then erase evidence.
There is deliberately no UI/environment flag to disable these boundaries. Tests
mock catalog queries, writes, metadata and failures; they never connect to
PostgreSQL. Real database grant/HBA configuration needs administrator verification
in each deployment before it can be considered protected.

## Bounded startup checks

Serving startup allows ten seconds for runtime pool creation, connection probe
and least-privilege verification together. Earlier caller cancellation still
applies. Unreachable storage, stalled privilege queries or unknown/broad privileges
refuse startup; no agent execution begins. Connection/pool failures use generic
messages instead of printing driver errors that may contain credentials or
untrusted backend detail. Inspect protected database administration logs for
connection diagnostics and restart after repair. These mandatory checks have no
feature flag and do not weaken role restrictions.

Mocked tests stall connection, query and row iteration with a short caller
deadline, and verify least-privilege acceptance, broad-role rejection and synthetic
secret omission. No live DB outage has been tested. Cooperative driver cancellation
is required; this is not an external watchdog for a backend that ignores context.
The separate operator-run migration command still needs its own operational timeout.
