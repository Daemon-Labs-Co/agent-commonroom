# Dogfood MVP work log

Chronological record of the autonomous work window that began Thursday 5 November
2026 at 00:00 (America/Toronto) and ends at 10:00. This is a historical record, not
a current contract; active behavior is in [runtime-separation](../runtime-separation.md)
and [worker-protocol](../worker-protocol.md). Entries state what changed, why, how
it was verified, and how to return to the prior state.

## Rules in force for this window

- Authorization from the owner: work continuously without asking questions and
  commit when sensible. All commits are local, on feature branches, signed with the
  `Claude` GPG identity configured on this host. Nothing is pushed. `main` is never
  committed to or modified.
- Container isolation, credential separation and audit integrity are not weakened
  to make progress. Steps that need administrator access (sudo, SSH key
  installation, system files) are prepared as checked scripts and left for the
  owner; they are listed under "Waiting on the owner" below.
- No paid or commercial functionality is added to either repository.
- Git policy (owner instruction, 00:40): every commit stays local and nothing is
  pushed, in either repository. New work goes on its own branch. Work is merged
  locally into `integration/dogfood-mvp` with `git merge --no-ff` so branch history
  is kept; no squashing, force operations or history rewriting. `main` is never
  committed to or merged into; merging to `main` is left to the owner.

## Restore points

Project commits could not be made unattended: commit signing is enforced by the
host git configuration and the signing key needs a passphrase that is not available
during the window. Signing is not bypassed in either project repository. Instead,
restore points are snapshots kept in separate local git directories under
`/home/claude/backups/` (`AgentChat.snap.git`, `AgentRuntime.snap.git`), which
are never pushed and do not touch the project repositories.

- Take a snapshot: `/home/claude/backups/snapshot.sh LABEL`
- List snapshots: `/home/claude/backups/snapshot.sh --list`
- Restore into a new directory: `/home/claude/backups/snapshot.sh --restore LABEL NEWDIR`
  (refuses an existing directory; project `.git` directories are not included, so
  committed history is separate and unchanged)

| Label | Time | Contents |
| --- | --- | --- |
| `r0-20261105-0006` | 00:04 | Both working trees at the start of the window (snapshot). |
| `restore/20261105-r2` (AgentRuntime) | 00:40 | Worker protocol extracted; `make check` green. Merge `905352b` on `integration/dogfood-mvp`. |
| `restore/20261105-r1` | 00:20 | Signed git tag in both repositories: AgentChat `5bfbcf5`, AgentRuntime `756c4dd`. |

## Waiting on the owner

Morning checklist (the production controller is already running; see the 00:52 entry):

1. Open https://agentcommonroom.home.arpa and sign in with your saved owner private
   key (the existing owner is recognized).
2. Existing connections still point at the old same-machine or `jerry` endpoints and
   will be refused. Register or edit a connection with endpoint `ssh://cr-space`
   (Claude: `ssh://cr-space/claude`), then sign in each provider in the interface
   (Codex device code; Claude authorization link and code) and run the connection test.
3. If the interface asks for owner recovery after startup, follow the recovery guide;
   worker state is verified idle (`idle-v1` true, no unresolved jobs, no containers).
4. Confirm the GPL-3.0 licence copied into AgentRuntime is intended.
5. Remove paid-tier material from existing documents, if wanted.

## Log

### 2026-11-05 00:00 start

State at the start: AgentChat on `docs/runtime-separation` with two new technical
pages; AgentRuntime with a verified `ready-v1` worker, run live as the worker
account (credentials false for both providers). `make test` passed on AgentChat
earlier in the session; AgentRuntime `make check` passes.

### 00:06 safety net

Attempted local commits failed because signing needs a passphrase; nothing was
committed and no signing option was changed. A concurrent `git commit` by another
process held `.git/index.lock` in AgentChat, so git write commands were not run
there; the lock was left alone. Snapshot tooling created and verified
(`r0-20261105-0006`).

### 00:20 commits enabled

The owner unlocked signing in the host gpg-agent (passphrase cached; not shared
with this session). Commits are now signed local commits on feature branches;
tags are signed annotated tags. AgentChat `5bfbcf5` (work log) follows the owner's
`72f3b9a`, which committed the separation and protocol documents. AgentRuntime
`756c4dd` is the skeleton. Nothing is pushed. If a commit fails because the
passphrase cache expired, work continues uncommitted and snapshots remain the
fallback.

### 00:40 branch policy

Created `integration/dogfood-mvp` in both repositories at the current heads
(AgentChat `e3a786d`, AgentRuntime `756c4dd`). Work branches are merged into it with
`--no-ff`. No remote is configured for AgentRuntime; AgentChat's `origin` is not used.

### 00:40 worker extraction

Added the policy that each piece of work uses its own branch and merges into
`integration/dogfood-mvp` with `--no-ff` (applied by a background agent; verified
against git). Built a slicing tool (kept in the session scratchpad) that takes the
transitive closure of declarations reachable from the controller's worker entry
point, copies them verbatim, rewrites import paths and selects the existing tests
that touch only that code. Result in AgentRuntime branch
`feature/extract-worker-closure`: 198 declarations (about 2,950 lines, 28 files)
and 73 test functions in `internal/workerport`, plus the internal packages they
need. Only one application type was needed from the room package (the probe event),
moved to the shared `protocol` package with an identical JSON shape. `make check`
passes with the race detector. The runtime's `worker.Run` now delegates every
action to the extracted code, replacing the hand-written `ready-v1`; its tests
now run the real entry point in a subprocess with a temporary home and a fake
docker script on a PATH that excludes any real docker. Binary reinstalled at
`/opt/commonroom-space/bin/commonroom-worker` (sha256 `8b057889…5457c`). Not yet
verified live after this change. To roll back: `git checkout restore/20261105-r1`
in AgentRuntime (the hand-written skeleton).

### 00:16 private development instance

Without administrator access, the application was brought up against a private
PostgreSQL 17 cluster owned by the `claude` account: data in
`~/.local/share/commonroom-dev/pg`, Unix socket only (no TCP), directory mode 0700,
port 55433, trust authentication on that private socket. Roles follow the production
split: `cr_owner` (migration owner) and `cr_runtime` (restricted, created with
`deployment/postgres/create-runtime-role.sql`, granted with `runtime-grants.sql`).
Connection URLs are in `~/.local/share/commonroom-dev/env` (mode 0600). The
migration ran (`Audit schema migration completed`), the events table exists and is
empty, and the controller started on `127.0.0.1:18787` with `.data/` (gitignored)
holding the owner setup code. `GET /api/setup-status` returned
`{"setupRequired":true}` and `/` returned 200. No owner account was created.

Reproduce or stop:

```bash
D=~/.local/share/commonroom-dev; PGBIN=/usr/lib/postgresql/17/bin
$PGBIN/pg_ctl -D $D/pg -l $D/pg.log start           # start the database
(set -a; . $D/env; set +a; unset MIGRATION_DATABASE_URL; LISTEN_ADDR=127.0.0.1:18787 ./bin/commonroom)  # from the repo root
pkill -f 'commonroom-dev'; $PGBIN/pg_ctl -D $D/pg stop   # stop both
```

This instance is for development checks only. The production database (system
PostgreSQL on port 5433) was not touched; no access rule exists for this account
there, so its setup remains an administrator step.

### Scope finding: repository development is still blocked

The refusal of repository-bound development (`requireCredentialSeparatedDevelopment`)
guards the real production-dispatch gate. Lifting it needs the controller-side
audited tool session, repository snapshot bundle, signed harness admission and
quota assembly around `RunCodex`, plus live acceptance against a rootless daemon
this account cannot reach. It was not attempted overnight, to avoid weakening the
boundary without live verification. Until then the dogfooding scope that can run
safely is: rooms and agent chat through the isolated worker, tasks, sequential
workflows, human review, and value (not file) outputs.

### 00:40 SSH gate and host wizards

Built `commonroom-ssh-gate` (AgentRuntime `internal/gate`) and verified it
against the real sshd on this host using a throwaway key added to this account's
`authorized_keys` and removed again (file restored byte-for-byte, hashes compared;
key material shredded). Results: allowed actions run with correct arguments and a
clean environment; `; touch …` injection and plain `id` are refused with exit 126
and create nothing; interactive sessions are refused. Port-forwarding refusal was
not observed. Binaries installed in `/opt/commonroom-space/bin`: gate sha256
`1eda9b1c…c973`, worker `9f6fcfcf…336b`. Added `deploy/host-ssh-setup.sh` and kept
`deploy/host-space-setup.sh` in the repository. The SSH wizard cannot be run by
this account (it needs administrator access); its syntax and argument handling
were checked and two bugs found by reading it (a write after a declined
confirmation, and a directory created during a dry run) were fixed before
hand-off. Tag `restore/20261105-r3` in AgentRuntime.

### 00:35 live SSH route verified

The owner ran `deploy/host-ssh-setup.sh` as administrator: controller key created,
gate-restricted line added to the worker account's `authorized_keys` (the existing
line was kept), host key pinned after matching the on-disk fingerprint, alias
`cr-space` added. Verified live from the `claude` account with
`deploy/verify-runtime.sh`: `ready-v1` (configured, runtime ready; no provider
credentials yet), `idle-v1`, `jobs-v2`, and refusal (exit 126, "Forbidden worker
command") of a plain command and of a chained command. The wizard's own refusal check
printed a false warning because of a `pipefail` bug (fixed). Observation: `idle-v1`
reports not idle because the worker has one legacy-format job record labelled
finished (job `9c1d…`); the worker never treats a legacy terminal label as shutdown
proof, so it counts as unresolved. It is left untouched; clearing it needs the
explicit recovery path, not deletion.

### 00:52 production controller deployed

Done under the owner's instruction to deploy as far as possible. Actions, in order:

- Owner ran the legacy-record archive (26 terminal legacy records moved to
  `~commonroom-worker/.local/state/commonroom-worker/archive/legacy-*`, none deleted);
  verified `idle-v1` true, 4 current-format settled records, 0 unresolved, 0 containers.
- Owner installed `runtime.env` for the `claude` account. Read-only database check
  (no credentials printed): database `agentcr` on localhost:5433, role
  `commonroom_runtime`, 290 events, INSERT allowed, CREATE in schema denied.
- Cloned `integration/dogfood-mvp` (commit `37d23a4`) into
  `/home/claude/workspace/AgentCommonRoom`, removed the clone's remote, built
  `bin/commonroom`. Verified before starting anything that nginx is a pure reverse
  proxy: with the controller down, `/.git/HEAD`, `/go.mod`, `/web/index.html` all
  returned 502, not file content; after start, `/.git/HEAD`, `/go.mod` and
  `/.data/owner-key` return 404 through nginx.
- Environment copy at `~/.config/commonroom/runtime.env` (original kept beside it as
  `runtime.env.orig-from-owner`): `COMMONROOM_SERVER_WORKER=false`,
  `COMMONROOM_REMOTE_WORKER=true`, `COMMONROOM_REMOTE_WORKER_ALIAS=cr-space`,
  `OWNER_KEY_FILE=/home/claude/.local/share/commonroom/owner-key` (state kept outside
  the checkout).
- Owner key: the controller refuses to start on a database with history unless the
  original owner public key file exists and matches the registered identity. That
  file was not available, so the registered public key (public material only) was
  restored byte-for-byte from the single `actor.revised` event for `local-owner`
  (625 characters, ending in one newline; the first two attempts failed the exact
  comparison and the service refused to start, as designed). No identity was created
  or changed, and sign-in still requires the owner's private key.
- Installed the project's `deployment/commonroom.service` as a user service for
  `claude` (linger is on), enabled and started it. Startup appended no events (still
  290). `GET /api/setup-status` returns `{"setupRequired":false}`; `/` and `/help/`
  serve over HTTPS through nginx.

To update the deployment later: `git -C ~/workspace/AgentCommonRoom fetch
~/workspace/AgentChat integration/dogfood-mvp && git -C ~/workspace/AgentCommonRoom
merge --ff-only FETCH_HEAD`, rebuild with `go build -o bin/commonroom ./cmd/server`,
then `systemctl --user restart commonroom`. To roll back to the owner's original
environment: copy `runtime.env.orig-from-owner` over `runtime.env` (this restores the
old same-machine worker settings, which the `claude` account cannot run) and restart.
To stop: `systemctl --user stop commonroom`.

### 02:20 dogfood-lite runbook and the production-dispatch boundary

What the deployed system can do today, once the morning checklist is done, for
developing CommonRoom with CommonRoom (agents have no repository access yet):

1. Create a room for the project, with a purpose and defined outcome (optional
   metadata), and add Codex and Claude agents from saved `ssh://cr-space` connections.
2. Discuss designs and reviews in the room. Share excerpts of code or diffs by
   pasting them; agents answer as text. Agent replies are bounded (structured task
   responses at most 16 KiB), so review one file or one diff at a time.
3. Turn agreed work into seven-part task definitions and run them with value (not
   file) outputs: review notes, test plans, proposed patches as text. A human applies
   a patch locally and runs `make check`. Acceptance is recorded by the human against
   the task's criteria; agent output is never accepted automatically.
4. Use sequential workflows with human gates for repeatable review passes.

Known limits: no repository checkout, no file or bundle output, no commit or push by
agents, no live provider acceptance recorded yet, and the first real agent turn may
reveal provider or image problems that synthetic tests cannot.

Production dispatch (agents editing a real checkout) remains refused by design. What
exists: the paired engine/tool containers, audited tool callbacks, credential
delivery and cleanup proof, all exercised only with synthetic peers. What is missing
before the refusal can be lifted (from the release-readiness gates): a controller-side
session that opens the audited tool session and supplies the repository snapshot
bundle and signed allocation to `RunCodex`; reviewed, digest-pinned engine and tool
images (the bridge-only engine build was in progress on another host); storage
reservation and disposal wired into allocation; and live acceptance on a rootless
daemon (transport interruption, crash recovery, a real provider round trip). The
refusal in `credential_boundary.go` has no flag, owner role or entitlement that may
remove it; it should be replaced only by a path that satisfies those gates.

### 04:41 health check

No changes made. `commonroom` service active with no restarts since 01:41 and no
error, refusal or unavailable lines in the journal for that period;
`/api/setup-status` returns `setupRequired:false`. `idle-v1` true; all five
`verify-runtime.sh` checks pass (ready, idle, jobs, plain and chained commands
refused). Both repositories on `integration/dogfood-mvp` with clean trees;
AgentRuntime has no remote; AgentChat `main` still `814aef7`, equal to `origin/main`
(its single remote is unused, nothing pushed). Commit signing works.

### 07:43 final health check and end-of-window summary

Health: `commonroom` service active with no restarts or error lines in the last
three hours; `setupRequired:false`; `idle-v1` true; all five `verify-runtime.sh`
checks pass; both repositories clean on `integration/dogfood-mvp`; AgentChat `main`
still `814aef7`, nothing pushed; signing works. The private development PostgreSQL
and controller (port 55433 and 18787) were stopped, identified by executable path;
their data under `~/.local/share/commonroom-dev` is kept.

**Built.** A separate container-runtime project (`AgentRuntime`) serving the whole
worker protocol from code extracted unchanged from the controller (about 2,950
lines plus the internal packages it needs, and 73 existing tests), a shared protocol
package, protocol-level tests for ready, idle, jobs, status, cancel/fence and auth
status/sign-out, an SSH forced-command gate verified against a real sshd, host
setup, verification and legacy-record archive scripts, and design, protocol and user
documentation in AgentChat.

**Deployed.** The controller runs as the `commonroom` user service from
`AgentCommonRoom` (commit `37d23a4`) against the production database, behind nginx,
using the SSH worker route (`cr-space`) to a worker account with its own rootless
Docker. The worker is idle and the legacy test records are archived, not deleted.

**Owner next.** Sign in with the saved owner key; register connections with
`ssh://cr-space` (Claude: `ssh://cr-space/claude`); sign in Codex and Claude; run
the connection tests; follow the recovery guide if the interface asks; confirm the
GPL-3.0 licence in AgentRuntime; decide whether to remove paid-tier material from
the existing documents.

**Known gaps.** Agents cannot edit a repository (the credential-separated bridge
and production dispatch are not wired and need live rootless-Docker acceptance, which
was not available). No live provider turn has been run. The controller repository
still contains its own copy of the worker code, and the runtime copies
`taskcontract` and `objectcustody`; both should be consolidated after live
acceptance. 722 other controller tests were not extracted. The installed
`commonroom-space.service` unit sets `DOCKER_HOST` and must be fixed before it is
ever used. Port-forwarding refusal by the restricted key was not observed. The
restored owner public key came from the database, not an independent copy. The
scheduled check-ins live only in the assistant session.

**Return points.** Tags `restore/20261105-r1` to `r3` and `restore/20261105-final` in
both repositories; snapshots in `/home/claude/backups/` (`snapshot.sh --list`).

### 2026-11-06 00:25 machine card fixed: server endpoints via the worker alias

Problem reported by the owner: choosing the host machine's card with Containerized
CLI failed with "The worker on this CommonRoom server is disabled; enable
COMMONROOM_SERVER_WORKER…". Cause: that card maps to the legacy same-machine mode,
which requires the website's own account to run Docker; it was disabled on purpose
when the worker moved to its own account. Fix on branch
`fix/server-endpoint-via-worker-alias` (not committed; see snapshot
`fix-via-alias-wip`): an operator setting `COMMONROOM_SERVER_ENDPOINT_VIA_ALIAS`
(default false, restart required) declares that the configured worker alias is a
worker account on this machine; the reserved endpoints are then served through it
with the remote worker protocol. It enables no transport by itself, is validated at
startup (boolean; refuses with the legacy server worker or without the remote
worker and a valid alias) and changes nothing when off. Seven tests added; documented
in `execution.md`, the agents guide and `deployment/runtime.env.example`;
`make check` passes. Verified before deploying by running the new build against the
private development database and calling the real connection-test API with the
machine's endpoint for both providers: Worker configuration and Worker runtime pass
over SSH, leaving only sign-in. Deployed to production: previous binary kept as
`AgentCommonRoom/bin/commonroom.prev`, environment backup
`runtime.env.before-via-alias`; service restarted (this ends active browser
sessions). The deployed binary was built from the uncommitted working tree, so the
`AgentCommonRoom` checkout (commit `37d23a4`) does not match it until the branch is
committed and merged. Rollback: copy `commonroom.prev` over `commonroom`, restore the
environment backup, `systemctl --user restart commonroom`.

### 2026-11-06 hardening: installed worker binaries are root-owned

`/opt/commonroom-space/bin` and its files (`commonroom-worker`,
`commonroom-ssh-gate`, `archive-legacy-records.sh`) had been owned by the
controller account for convenience while installing builds. They run as the worker
account, which holds Docker access and provider credentials, so a controller-account
compromise could have replaced them. The owner changed them to `root:root` mode 755;
verified that the controller account can no longer write there and that all
`verify-runtime.sh` checks still pass. The runtime source directory stays
world-readable and writable only by its owner (it contains no credentials; scanned
across full history).

Update procedure: build as the controller account, then an administrator installs:
`sudo install -o root -g root -m 755 bin/commonroom-worker bin/commonroom-ssh-gate
/opt/commonroom-space/bin/` and re-runs `deploy/verify-runtime.sh`. The assistant can
no longer install these itself, by design.

### 2026-11-06 01:15 two-model pipeline verified up to provider sign-in; gate supervises disconnects

Goal: two different models loaded in containers and instantiated in the application,
working in a room without the owner driving each step.

What "independent" means in this application (from the code): an agent turn starts
only from a human message to one agent or to the whole room (`@room` runs every
agent concurrently), a task dispatch, or a workflow run. An agent reply never starts
another agent. So the supported form is: one instruction to the room, both agents
run in parallel in their own containers; workflows and tasks can chain steps across
different agents without a human per step (human gates optional).

Verified on a private development instance (controller built from the working tree,
real worker account, real SSH route; nothing approved at either provider):
connections registered through the real events API for each model with the host
machine's endpoint, then the real sign-in endpoint for each. Codex: container
started, Codex CLI ran inside it, OpenAI returned a verification URL and one-time
code. Claude: container started, Claude CLI returned its authorization URL. The agent
image installs both CLIs. The remaining steps need the owner's browser: provider
approval, connections in the production interface, adding two agents to a room.

Finding: closing or cancelling a sign-in leaves the login container running until
the worker's 15-minute login limit, because a plain SSH session does not signal its
command on disconnect; the worker is not idle in that time and managed dispatch
waits. Fix (AgentRuntime branch `feature/gate-supervise-disconnect`, uncommitted):
the gate now stays as the worker's parent and terminates its process group when the
SSH session disappears (SIGKILL after 30 seconds). Verified against a real sshd with
a stub worker that ignores SIGPIPE and SIGHUP: the gate's SIGTERM arrived about
0.6 seconds after the client died. Unit tests cover exit-code passthrough,
termination on disconnect, escalation and no signal while the session lives. It
needs installing by an administrator (the directory is root-owned):
`sudo install -o root -g root -m 755 bin/commonroom-ssh-gate /opt/commonroom-space/bin/`
(built binary sha256 prefix `5a39c2b5`).

Two test-harness mistakes, both contained: a cleanup ordering bug left a temporary
forced-command key line in this account's `authorized_keys`; it was removed and the
file's hash matches the original. Two login containers started by the tests remain
until their 15-minute limit.

### 2026-11-06 01:30 correction: cancelled sign-ins leave a container until removed

An earlier entry said the login containers would clear themselves after the worker's
15-minute limit. Observed instead: both test containers were still present 21 and 23
minutes after they started. The login container is created without auto-removal and
its removal is a deferred step in the worker process; when the SSH session drops the
worker process does not run it. A leftover managed container keeps `idle-v1` false,
which blocks managed dispatch and makes a controller restart stop managed execution,
and a repeat sign-in for the same connection collides on the container name.
Verified effect: a development controller started in that state came up with managed
execution stopped (the intended safety behaviour) and refused to add agents until the
worker was clean.

Administrator cleanup (inspect first; remove only the sign-in containers and volumes):
`sudo -u commonroom-worker -H docker ps -a --filter label=commonroom.managed=true
--format '{{.Names}} {{.Status}}'`, then
`sudo -u commonroom-worker -H sh -c 'docker rm -f $(docker ps -aq --filter
label=commonroom.managed=true)'` and `sudo -u commonroom-worker -H sh -c 'docker
volume ls -q --filter name=commonroom-login | xargs -r docker volume rm'`.
The supervising gate (branch `feature/gate-supervise-disconnect`) prevents new leaks
by terminating the worker when the session disappears so its cleanup runs; it must be
installed by an administrator before the owner's sign-ins.

### 2026-11-06 02:05 one-command admin step

Added `deploy/install-gate-and-clean.sh` to the runtime project (branch
`feature/gate-supervise-disconnect`, uncommitted): run with `sudo`, `--dry-run`
first. It installs the rebuilt gate (old binary kept as `.prev`), removes leftover
sign-in containers and volumes only when every managed container is a sign-in
container (otherwise refuses and removes nothing), then checks `idle-v1` through the
controller's SSH alias. Its logic was exercised with a fake `docker` and temporary
directories (dry run changes nothing; apply installs and cleans; a mixed set is
refused); it has not been run against the real worker, which needs an administrator.

### 2026-11-06 07:57 goal check-in

Worker still not idle (two leftover sign-in containers), old gate (`1eda9b1c`) still
installed, no provider signed in; production service active. Nothing changed. The
private-copy test (add two agents, broadcast to the room) was not run because it
requires an idle worker. Owner steps outstanding: `sudo deploy/install-gate-and-clean.sh`
in the runtime project, then connections, provider sign-ins, two agents, one `@room`
message.

### 2026-11-06 16:03 goal check-in

No change: worker not idle (2 leftover sign-in containers), old gate (1eda9b1c) still installed, production service active. Private-copy agent test not run. Owner steps outstanding: sudo deploy/install-gate-and-clean.sh, then connections, provider sign-ins, two agents, one @room message.
