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
| `restore/20261105-r1` | 00:20 | Signed git tag in both repositories: AgentChat `5bfbcf5`, AgentRuntime `756c4dd`. |

## Waiting on the owner

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
