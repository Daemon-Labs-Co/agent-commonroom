# Workspaces: many customers in one installation

Design record; stages are implemented in order, each default-off. Stages 1 to 6 (storage, runtime, platform, identity, execution controls and the
verification suite) are implemented: see [events and persistence](persistence.md) and the runtime notes below. See ADRs
[0004](archive/adr/0004-many-workspaces-in-one-installation.md),
[0005](archive/adr/0005-workspace-chosen-by-host-name.md),
[0006](archive/adr/0006-passkeys-and-oidc-without-new-dependencies.md) and
[0007](archive/adr/0007-workspaces-created-by-signed-platform-api.md).

## Input

Goal: one installation serves many independent **workspaces**. People sign up or are
invited, each workspace keeps today's guarantees (permanent audit log, mandatory
container isolation, owner approval of agent work) and no workspace can read, write,
stop or slow another beyond its stated limits.

Configuration (operator process settings, read at startup):

| Setting | Default | Meaning |
| --- | --- | --- |
| `COMMONROOM_TENANCY` | `single` | `single`: one workspace named `default`, behaviour unchanged. `multi`: workspaces resolved from the host name. |
| `COMMONROOM_WORKSPACE_DOMAIN` | none | Base domain in `multi` mode; a workspace is served at `<slug>.<domain>`. Requires HTTPS origins and wildcard DNS/certificate. |
| `COMMONROOM_PLATFORM_KEYS_FILE` | none | Registry of Ed25519 public keys allowed to call `/platform/v1/`. |
| `COMMONROOM_SIGNUP` | `closed` | `open` enables the built-in rate-limited sign-up page. |
| `COMMONROOM_OIDC_*` | none | Per-installation OIDC provider settings for workspace sign-in. |

Current single-tenant assumptions this changes: one `events` table and hash chain, a
full replay of that table per request, one global lock, one `server` value holding
stop, incident, budget and run state, the constant owner actor, and one `.data`
folder for the owner key, sessions, agent credentials and local configuration.

## Output

A workspace is a complete, independent `server` instance created on first use and
cached. Requests are routed to it by host name. Refusals are explicit and
non-revealing: an unknown host returns 404 with no workspace information; a valid
session for another workspace's host is simply unauthenticated; a suspended
workspace returns 503 with a fixed message.

## Functionality

**Stage 1 - storage (migration 002, administrator-run).** `events` gains
`workspace_id` (existing rows become `default`), an index on
`(workspace_id, sequence)` and row-level security forced for the runtime role: rows
are visible and writable only when `workspace_id = current_setting('app.workspace_id',
true)`. A scoped database wrapper starts every transaction with
`set_config('app.workspace_id', $1, true)`; no setting means no rows and no inserts.
The hash chain, previous-hash lookup, count checks and integrity verification are
therefore per workspace with unchanged SQL; the append lock becomes per workspace.
A `workspaces` registry table (id, slug, name, status, limits, timestamps) is the
only cross-workspace table.

**Stage 2 - runtime.** The body of `main` that builds a `server`, its routes, owner
identity, sessions and integrity watcher becomes `newWorkspace(...)`, called once for
`default` in `single` mode or lazily per workspace in `multi` mode. Each instance has
its own mutex, run maps, stop and incident flags, request budgets, auth and data
folder `.data/workspaces/<id>/`. A top-level handler resolves host to workspace, adds
security headers and dispatches. Session cookies are host-only; the Origin check uses
the workspace's exact origin.

**Stage 3 - platform.** `/platform/v1/workspaces` (create, list, suspend, resume,
delete, set limits) with signed requests, replay refusal and a rate limit. Creating a
workspace returns its setup link. Limits per workspace: members, rooms, agents,
requests per minute, daily agent turns, concurrent containers and stored bytes;
reaching one is a visible refusal.

**Stage 4 - identity.** Owner-created invitations (single use, expiring, role and
rooms fixed); passkey registration and assertion; OIDC sign-in linked to an
invitation or an existing member; the key-file methods remain.

**Stage 5 - execution.** Container names, credential volumes and workspaces are
namespaced by workspace in the worker; controller admission counts running containers
per workspace and installation-wide.

**Stage 6 - verification.** A cross-workspace attack suite runs every route with one
workspace's session against another's host and identifiers, checks the database refuses
cross-workspace reads and writes with the application filter removed, and an
adversarial review attempts escapes before release.

**Stage 2 as built.** `newWorkspaceRuntime` (cmd/server/workspace_runtime.go) builds a
workspace's complete `server`, routes, owner identity, sessions and integrity watcher;
`main` only reads process settings and starts a `workspaceHost`. In `single` mode the
host serves the one default workspace for every Host header, builds it at startup and
behaves as before (same key and session locations). In `multi` mode the host maps
`<id>.<COMMONROOM_WORKSPACE_DOMAIN>` to a workspace (exactly one lowercase DNS label,
reserved names refused), asks the registry (`workspaces` table, migration 003) whether it
exists and is active, trusts that answer for 5 seconds, builds the workspace server on
first use exactly once, remembers unknown workspaces for 30 seconds (bounded) so probing
cannot hammer the database, and answers unknown or deleted workspaces with a bare 404
and suspended or failing ones with a generic 503. Each workspace keeps its files in
`.data/workspaces/<id>/` and has its own setup code, session store, mutex, stop and
incident flags, request budgets and Origin (`https://<id>.<domain>`). By default hosted workspaces refuse all managed agent execution at admission and do not
expose provider sign-in, connection tests, job diagnostics and recovery (503); stage 5
describes the operator opt-in. Workspace disposal, extensions and conversation import
are refused for hosted workspaces whatever the operator chooses.

Known limits of stage 2: every built workspace runs its own 30-second integrity replay
and each request still replays that workspace's history (no snapshot cache yet), and
workspace servers are not evicted when idle.

**Stage 3 as built - platform API.** Served only on `platform.<domain>` and refused unless
`COMMONROOM_PLATFORM_KEYS_FILE` lists the caller's Ed25519 key (same registry format and
permission rules as extensions; the signing input is `commonroom-platform-v1`, so an
extension signature never works here). Headers: `X-Platform-Id`, `X-Platform-Timestamp`
(within 60 seconds), `X-Platform-Signature`; replayed signatures (compared by decoded bytes, so re-encoding one does not help) and more than 120
requests a minute are refused; every authentication failure returns the same 403.

| Request | Effect |
| --- | --- |
| `POST /platform/v1/workspaces` `{id, name, limits?}` | Create (201) or return the existing one with the same name (200); different name 409; deleted ids are never reused. Returns the setup link. |
| `GET /platform/v1/workspaces?limit=&after=` | Page of up to 100, ordered by id. |
| `GET /platform/v1/workspaces/{id}` | Status, name, limits, whether setup is pending (never a link). |
| `POST .../{id}/suspend`, `/resume`, `/delete` | Takes effect on the next request. Delete is terminal and keeps a registry tombstone. |
| `PUT .../{id}/limits` | Replaces limits live; unknown fields and out-of-range values refused. |
| `POST .../{id}/setup-link` | Replaces a pending workspace's setup code (old code and any session it created stop at once); 409 once claimed. |

Each action is also appended to the workspace's own history as `platform.workspace.action`,
so its owner can see what the operator did. The API returns no rooms, messages, members or
settings. Hosted setup codes expire after 24 hours (restart or reissue renews).

Limits (defaults members 5 besides the owner, rooms 20, agents 10, 600 API requests a
minute, 200 daily turns, 2 concurrent turns) refuse the next creation with 409, or exceed the request rate
with 429 and `Retry-After`; they never change another workspace. Operator tools:
`commonroom --platform-keygen` and `commonroom --platform-request`. Deleted workspaces'
history is removed only by the administrator-run `scripts/purge-workspace.sql`.

**Stage 4 as built - identity.** Passkeys, single sign-on and invitation links per workspace,
specified in [identity](identity.md) and the sign-in [user guide](../user/sign-in.md). Hosted
workspaces start with passkeys and invitations on. One installation-wide OpenID Connect
provider is supported; it must allow a redirect URI per workspace host (wildcard redirect URIs,
or a provider that permits them), which not every provider does.

**Stage 5 as built - execution controls.** Hosted agent execution is off unless the operator
sets `COMMONROOM_HOSTED_EXECUTION=true`; startup then requires an isolated worker and refuses
to start while repository workspaces, isolated commands or worker development are enabled,
because those read the host's disk through process-wide settings: hosted agents are chat-only.
Running turns are counted per workspace and across the installation
(`COMMONROOM_HOSTED_MAX_TURNS`, default 8, at most 256) by one shared pool; a workspace is also
held to its `concurrentTurns` limit (default 2, 1-16). A refusal says only that this workspace
or the installation is at capacity and names no other workspace. Provider credentials stay in
per-connection folders keyed by server-assigned 128-bit connection ids, so two workspaces cannot
share one; this was not re-pathed. The pool check and the later start are not one atomic step, so
a brief overshoot is possible: the limits protect resources and are not a security boundary.
**Not verified live:** nothing in this stage has run against the real worker, containers or a
provider; verify on a staging installation before enabling it. Per-workspace quotas inside the
worker (and the worker protocol) were not changed.

**Stage 6 as built - verification.** `workspace_isolation_suite_test.go` builds two hosted workspaces,
each holding recognisable data (a room, a message, a member with a passkey, an invitation), and
attacks every route a workspace serves (`routeMux` records them, so a new route is attacked
automatically) with the other workspace's owner and member sessions under both cookie names, five
HTTP methods and the other workspace's identifiers, then checks that none of its content appears and
that no refused request wrote to history. A second test uses a valid owner of one workspace with
the other's room, invitation, passkey and member identifiers; a third shows one identity can belong
to two workspaces as independent accounts; a fourth fails when a new route answers an anonymous
request with success. Boundary breaks (every host served by one workspace, shared sessions, shared
credential lookup) were each confirmed to fail the suite. Hardening found by reviewing the design as
an attacker: hosted cookies use the `__Host-` prefix (Secure, Path=/, no Domain) so a sibling
workspace cannot plant a cookie for another ("cookie tossing"); labels with hyphens in the third and
fourth positions (`xn--` look-alike names) are refused; cross-site and foreign-origin requests to the
public identity endpoints are refused; forwarding headers never choose the workspace.

The database policy is verified on a real PostgreSQL server by `scripts/check-workspace-database.sh`
(excluded from `make check`, which must not touch a database): migration idempotence, the isolation
proof for a restricted role, per-workspace hash chains, the upgrade of existing history and the
registry.

Known limits: every built workspace runs a 30-second integrity replay and each request replays its
history (no snapshot cache) and workspaces are never evicted, so cost grows with active workspaces;
a request-rate limit is per workspace, not per client address; the platform API's replay memory is
in process, so a replay inside its 60-second window is possible across a restart; a dedicated
external review (for example the red-team review) has not been run.

## Properties and state changes

- Defaults: `single` tenancy, `closed` sign-up, no platform keys. Without `multi` no new
  route is registered and behaviour is byte-for-byte today's.
- Persistent state: `events.workspace_id`, the `workspaces` registry, per-workspace
  files. All workspace history stays append-only; deleting a workspace is a platform
  action that leaves a tombstone, never an owner action.
- Isolation properties: database refuses cross-workspace access even if application code
  forgets a filter; a session, CSRF token and setup link are bound to one host; an
  incident, stop or budget exhaustion in one workspace never changes another.
- Not covered by these stages: billing and plans, email delivery, custom domains,
  per-workspace encryption keys, and moving a workspace between installations.
