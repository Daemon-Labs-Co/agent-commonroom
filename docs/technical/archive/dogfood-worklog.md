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

## Restore points

Each restore point is a git tag on a local commit plus a git bundle under
`/home/claude/backups/`. To return to a point: `git checkout <tag>` in the repo, or
`git clone <bundle> <directory>` to rebuild it elsewhere. Bundles contain full history.

| Point | Repository | Tag | Notes |
| --- | --- | --- | --- |
| R0 | AgentChat | `restore/20261105-r0` | Design and protocol docs, before any window work. |
| R0 | AgentRuntime | `restore/20261105-r0` | Skeleton with `ready-v1`; live check passed. |

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
