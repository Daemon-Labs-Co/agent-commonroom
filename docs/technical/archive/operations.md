# Operations guide

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


[Documentation index](documentation-index.md) · [Installation](installation.md) · [Release readiness](release-readiness.md)

## Supported installations

Run privately on numeric loopback, either on your computer or on a Debian controller
accessed through SSH forwarding. Browser and server ports must match the Host/Origin
checks. Public internet/reverse-proxy hosting remains unsupported. Managed execution
requires an approved [Linux/rootless-Docker worker](container-isolation.md).

## Configuration and assets

| Item | Behavior |
| --- | --- |
| `DATABASE_URL` | Runtime restricted-role connection; overrides local config/development default. |
| `LISTEN_ADDR` | Defaults to `127.0.0.1:8787`; numeric loopback only. |
| `OWNER_KEY_FILE` | Defaults to `.data/owner-key`; companion RSA files use suffixes. |
| `.data/local-config.json` | Optional private controller configuration; strict schema/permissions. |
| `web/`, `docs/`, root documentation and `migrations/` | Deploy beside executable; run from repository root. |
| `.env.example` | Reference only; the server does not automatically load `.env`. |

Run reviewed migrations explicitly with `--migrate` and a separate owner account.
Serving validates restricted runtime privileges and never applies migration DDL.
See [database security](database-security.md). Keep migration credentials inaccessible
to the service.

## Sign-in and keys

First setup creates a code and RSA key pair. Download and protect the owner private
key; successful signature login removes the server's pending private material and
setup code. Keep the public-key file. General key recovery/rotation and encrypted-PEM
import are not implemented.

Human sessions expire after eight hours and persist across restart by default.
Account enablement/access is checked on requests. Only the owner administers invited
humans and their access. Keep private keys backed up outside the repository; deleting
a file does not prove physical erasure from storage or backups.

### Private human session store

`human-sessions.json` lives beside `OWNER_KEY_FILE`. Keep it private and never share
it as diagnostics. Unsafe/corrupt files block startup. To force fresh login, stop
the application, remove only this operational session file and restart. Audit history
and RSA keys remain separate.

`COMMONROOM_PERSISTENT_SESSIONS=false` selects memory-only sessions and clears the
saved store; it does not weaken login proofs, authorization or session limits.
Only one controller may use the store. See [session renewal](login-availability.md).

## Installation diagnostics

Run `python3 scripts/check-installation.py` from the checkout as the controller
service account, with the same runtime configuration in that shell. Use `--hybrid`
on a Linux host with nginx/Docker, or `--url` for a different numeric-loopback port.
See [checks and repair guidance](installation-check.md). It never repairs state,
applies migrations or launches agents; incomplete checks remain visible.

## Agent health

A saved connection or an Idle profile is not evidence of provider availability.
Use [Test connection](connection-testing.md) for a fixed isolated provider reply.
For failures, inspect the approved alias, controller service-account SSH identity,
worker readiness, saved credential presence and safe job diagnostics.

Turns have bounded deadlines, but elapsed time or cancelled SSH is never shutdown
proof. [Durable jobs](worker-jobs.md) preserve unresolved intent across restart and
require reconciliation. Do not replay uncertain work or delete records to make
readiness pass. Public development execution is still [blocked](development-workspaces.md).

## Integrity and history

The history view presents readable immutable-event descriptions with verification
details. Use its available filtering/pagination for investigation; do not confuse
a displayed subset with a complete independent archive. Message text, instructions,
endpoints and raw provider errors are excluded from readable audit descriptions.

Integrity verification runs at startup, every thirty seconds and before managed
execution. Failure cancels controller work and latches execution off, including a
private durable marker when database incident recording is unavailable. Follow
[owner recovery requirements](database-security.md#integrity-and-recovery); do not
edit, delete or re-hash evidence to clear an incident.

Unkeyed checksums and restricted database privileges are not administrator-resistant
immutability. Independent external anchoring/WORM archives remain unfinished.
Retention/purge policy is not permission to schedule database deletion, and approved
workspace disposal remains pending.

## Backup and update procedure

1. Schedule maintenance. Finish/cancel work and confirm worker idle plus settled
   controller jobs; investigate uncertainty before replacement.
2. Stop the controller. Preserve private state, PostgreSQL history and deployed
   source/image identifiers. Back up provider storage separately and protect the
   human-held private-key backup.
3. Store encrypted off-machine copies outside the repository. Test restoration in a
   separate installation.
4. Build a reviewed revision and matching worker binaries/images. Run reviewed
   migrations explicitly with the migration identity and restricted grants intact.
5. Start one controller, sign in, check history/integrity and verify worker/provider
   readiness before resuming authorized work.

The [Debian guide](deployment/debian.md#operations-updates-and-backups) has concrete
service/backup commands. Do not restore an older event log as routine code rollback:
that loses authoritative history and may evade the current local chain checks.

## Emergency stop procedure

1. As owner, choose **Stop all agents** or **Stop room agents** and confirm a reason.
2. Check the stop banner/event, then inspect worker job shutdown evidence. A banner
   confirms a cancellation request, not every remote process's termination.
3. If the application/worker is unreachable or effects continue, use trusted host
   and network controls. Preserve records for reconciliation.
4. If audit/marker persistence fails, repair storage and record the stop durably
   before restart. Do not assume an unrecorded local latch survives failure.
5. Reconcile uncertainty and resume explicitly after investigation. Cancelled work
   does not automatically restart; room stops and integrity incidents remain separate.

`COMMONROOM_AGENT_SAFETY_UI` sets the initial presentation default; saved owner
settings override it. Hiding controls never disables backend stops. An authenticated
owner can use the protected [agent-safety API](api.md#emergency-agent-safety).
No stop control provides provider-wide token revocation or automatic rogue detection.

## Free durable job safety

[Worker diagnostics](worker-jobs.md) explains durable intent, protocol-v2 chat jobs,
crash reconciliation and owner recovery. Internal development pairs use a separate
versioned contract. `jobDiagnostics` defaults on and controls optional detailed UI;
the lifecycle journal and fencing remain mandatory when it is off.

## Single-controller startup lease

Run one controller per event store. Startup takes a nonblocking exclusive kernel
lease on `.data/controller.lock` before identity/database/worker initialization.
A second controller in that working directory refuses startup. The private empty
file remains after exit; the process releases its kernel lease.

**Never delete or replace the lock while a controller may be running.** Stop the
old service before replacement, including versions predating the lease. Another
working directory/host is not a supported workaround against the same database.
The lock does not implement distributed DB coordination or network-filesystem
leases. Migration and worker entry points do not take it; coordinate migrations
separately.

## Workflow execution and controller recovery

Workflow execution is optional and off by default. Enable only after checking
its [supported semantics](workflow-execution.md) and operational connections. Browser
closure does not stop its controller driver; controller restart does not silently
replay it. Inspect interrupted work before explicit Resume/Cancel. Resume can repeat
the current node and partial side effects; never use it to bypass an exhausted budget.

Disabling execution cancels live drivers and rejects further execution/gates while
preserving outputs/history. Cancel and authorized inspection remain available.
Emergency stop also cancels matching workflow drivers. Configuration never grants
public development admission or repository access by itself.

## Internal worker idle policy

The [paired-worker idle guard](worker-idle-shutdown.md) defaults to five minutes
and is configured in the trusted worker process environment using
`COMMONROOM_AGENT_IDLE_SHUTDOWN` and `COMMONROOM_AGENT_IDLE_MINUTES`. New pairs
snapshot the policy; disabling it never disables hard deadlines or ordinary
shutdown. The owner can configure new allocation policy in **Settings → Security → Idle agents**;
exact registered snapshots override the legacy environment fallback in reviewed
assembly. Controller-bound idle stops require a durable acknowledgment; default-on
`idleStopNotices` controls room notice visibility while preserving audit. Public
dispatch, production launch/policy assembly and live SSH/provider acceptance remain open.
Idle acknowledgments now use the existing private audit link; mixed peers that
lack its terminal extension refuse acknowledgment without fallback. Do not forward model/browser/SSH client values into this policy.

## Presentation and catalog flags

| Setting | Default | Effect |
| --- | --- | --- |
| `COMMONROOM_CONNECTION_EDITOR` | true | Guided editor layout; false restores basic presentation. |
| `COMMONROOM_WORKFLOW_TEMPLATES` | true | Catalog serving/explicit template application; saved workflows remain. |

Operator environment changes require restart; invalid booleans refuse startup.
Neither flag changes connection authorization, running-job safety or isolation.
Profile-photo flags control new uploads/display; historical immutable revisions
remain stored rather than erased by disabling a cosmetic feature.

## Troubleshooting and escalation

Start with [installation](installation.md), then [Debian](deployment/debian.md) or
[worker](agent-ssh.md) checks. Share the revision and redacted error summaries rather
than private transcripts, raw logs or credentials. Follow [security reporting](../../../SECURITY.md)
for vulnerabilities and [security hardening](security-hardening.md) for required
boundaries. Live acceptance and release gates are tracked in [release readiness](release-readiness.md).

## Optional bounded checks

The internal worker can explicitly admit one terminal Go build/test command when
its operator flag and separate operation grant both allow it. Defaults remain
file-only. See [isolated commands](isolated-commands.md) for configuration,
receipt semantics, container destruction and remaining release limits.

## Storage capacity

The internal [storage accounting](storage-accounting.md) ledger reserves a maximum
5,000,000,000 bytes across workspaces, bundles and artifacts. Missing complete
inventory refuses reservations, and container shutdown does not refund retained
data. No public storage-management operation or physical quota backend is enabled
yet. Do not fabricate inventory, edit the journal or delete retained volumes to
restore capacity. Production development stays gated pending physical enforcement
and exact approved disposal.

## Task repository reference configuration

`taskRepositoryLinks` is off by default. The owner may enable/disable this API-only
feature through an authenticated `feature.flag.updated` event, without restart or
migration. Disabling blocks reads and writes while retaining references/history.
See [task repository references](task-repository-links.md) for API requests, current
revision/CSRF/operation-key requirements and permission/URL limitations. It requires
no GitHub credentials or outbound integration and grants no repository authority.

## Isolated artifact-reader lifecycle

The [reader supervision and recovery contract](object-reader-supervision.md) describes
required production dependencies, bounded waits and known-ID cleanup. Do not wire it
to a permissive journal, model pipes or a host launcher. Unknown creation replies
retain capacity and require incident reconciliation; never adopt an ID by name or
refund storage from a missing helper alone. No product feature is enabled by this
infrastructure, and file-task dispatch remains blocked pending production assembly
and live acceptance.

## Reader lifecycle records

Keep `.object-read-<job>-<challenge>.record` files with the private worker state;
unknown/unfinished records retain container capacity. Never delete them to clear
an admission refusal or infer a storage refund from helper cleanup. The
[worker reader-journal guide](worker-object-read-journal.md) describes supported
exact-ID recovery, canceled-read metadata and the independent controller/account
services still required. Existing directories need no migration; missing records
mean no reader capacity was reserved, not evidence that an artifact was delivered.

The [controller reader ledger](controller-object-read-ledger.md) mirrors reader
transitions into the account audit chain. Missing/uncertain cleanup retains the
helper charge. Authenticated transport and independent physical evidence remain
required before production assembly; local metadata is not an absence witness.

The [reader authority bridge](reader-authority-bridge.md) describes private control
channel ownership, strict frame limits and scoped revocation. Its mutual-TLS
adapter requires provisioned trust roots, peer DNS identities and exact certificate
fingerprints at both endpoints before control frames can flow. The bound session
constructors also compare independently provisioned worker/record/intent bindings
and follow original authority lifetimes while idle. One-shot endpoint helpers
consume an operator-owned listener and dial one provisioned literal IP:port;
accept/dial waits are bounded, canceled connections close, and no retry or address
discovery is provided. A private-file loader now validates operator-provisioned certificates, keys and
trust bundles before constructing peer identities; keep every file owner-only and
outside model/container mounts. Reload requires a new session with an independently
approved peer pin. These helpers do not configure networking or issue keys.
Keep the creator
lease and original principal pinned; never expose a control socket to containers
or reconnect an uncertain link to bypass reconciliation. Actual transport
authentication, physical quota witnesses and reviewed starters remain deployment work.

Reader deployment now has [private worker profiles and offline preflight](reader-worker-profiles.md), including fixed endpoint selection, certificate loading and service shutdown ownership. This does not enable production file-task dispatch or certify physical storage/live acceptance.

A [read-only physical project-quota observer](physical-project-quota-observer.md) now checks pre-provisioned XFS enforcement on supported Linux x86-64 hosts. It does not provision quotas or replace complete inventory, account admission, rootless containment or live acceptance.

Operators can run the [worker storage preflight](worker-storage-preflight.md) against private independently approved bindings. It performs read-only observations before database startup and grants no allocation or live acceptance.

## Explicit startup commands

`commonroom --help` (or `-h`) prints supported commands and exits without taking
the controller lease, reading application configuration, migrating or connecting
to the database. With no arguments the binary starts the controller normally.
Unknown commands, misspelled preflight switches and unsupported operands exit2
with a sanitized usage diagnostic; they never silently fall through to startup.
For example, `--check-worker-storag` refuses instead of opening the application.

The private reader/storage preflights retain their documented0/1/2 status codes
and run before controller state. `--migrate` must appear alone: extra arguments
refuse before any migration work. The internal `--agent-worker` protocol keeps
its own bounded operand validation and mandatory isolation; listing it in help
is not permission to execute models directly. Arbitrary arguments/paths/credentials
are not echoed in unknown-command errors. There is no optional flag to bypass
this startup safety rule. Help/usage output does not certify deployment readiness.

Tests exercise actual executable entry-point ordering in temporary synthetic
processes with no inherited configuration/credentials, plus private profile
success and malformed command handling. They assert no controller state appears;
no database, provider, Docker or native quota query is used in these tests.
