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

Morning checklist, in order (all administrator steps are scripts that check first,
ask before replacing anything, keep backups and support `--dry-run`):

1. In AgentRuntime, run `deploy/host-ssh-setup.sh --dry-run`, then without
   `--dry-run`. It authorizes a controller key for the gate only, pins the host key
   and adds the SSH alias, then checks `ready-v1` and a refused plain command.
2. Production database (system PostgreSQL, port 5433): create the database, the
   migration owner and the restricted runtime role with the templates in
   `deployment/postgres/`, then migrate. Not scripted yet.
3. Start the controller with `COMMONROOM_REMOTE_WORKER=true`,
   `COMMONROOM_REMOTE_WORKER_ALIAS=cr-space`, `COMMONROOM_SERVER_WORKER=false`.
4. Sign in the agent connections (Codex, Claude) through the interface; the worker
   currently reports no provider credentials.

- Forced-command SSH access so the controller account can reach the worker account.
- Confirmation that the GPL-3.0 licence copied into AgentRuntime is intended.
- Removing paid-tier material from existing documents, if wanted.

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
