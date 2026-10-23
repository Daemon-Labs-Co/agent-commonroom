# P0 batch checkpoint

Updated: 22 October 2026 (America/Toronto).

## Current batch

- Branch: `codex/p0-release-gates`.
- Base: verified clean merged main `a2112b6a233f08774cffaeb5bc88827a3e38b429`.
- Published head: `d3d10743335791fffd0112c12f7e45a145e4294c`. PR #53 open;
  Greptile succeeded with valid P1 comment `4355424030` on that head.
  Local fix requires full daemon capability evidence before preparation/start,
  including explicitly true MemoryLimit, SwapLimit, PidsLimit and CpuCfsQuota.
  Refusal tests cover false, missing and malformed values for each field.
  No GitHub discussion was posted; current-head rereview follows publication.
- PR #52 merged the exact-role cleanup fix as `465b7f7`; Greptile succeeded
  on that reviewed head. No local fix was carried forward.
- User requested consolidation of remaining P0 work into the next commit.
- State: owner authorized sequential implementation of all nine gates on this
  branch for one eventual commit. Gate G09 platform/build foundation is in progress;
  no gate is certified and this is not a publication checkpoint.
- Added docs/p0-release-gate-acceptance.md with setup, failure/permission controls,
  exact evidence, teardown, pass/BLOCKED rules and all 74 catalog mappings.
- Added a standalone Codex provider regression through signed allocation and the owned
  isolated runtime with synthetic Docker/provider boundaries.
- Acceptance snapshot: 42 bound, 32 pending, 74 total.
- make test passed in 9.01 seconds; make check passed in 18.69 seconds.
  Elapsed approximately 9 minutes including repository and gate inspection.
  Live provider/daemon verification remains
  separate; no production readiness claim is made.
- Owner identified dedicated Codex SSH alias `jerry` and authorized necessary
  image pulls. Read-only checks confirmed Linux, rootless Docker, seccomp and
  cgroup v2/systemd. The recorded Go/Node/Rust digest-pinned build inputs were
  pulled successfully. No provider was launched or credential file read.
- The existing tool image is the previously documented local acceptance artifact.
  No worker executable was found on PATH or at the standard local-bin location.
  A reviewed bridge-only engine build, worker setup and live acceptance remain.
  Other dedicated provider accounts/images have not been identified. Never put
  credentials in this checkpoint or test fixtures.
- Open implementation gates remain production dispatch, other provider adapters,
  artifacts/revision evidence, storage enforcement/integration and idle policy
  assembly. Do not replace these gaps with canned observations.
- Specification validation: nine gate sections, all 74 unique scenario mappings
  and local links verified. make test passed in 8.59 seconds; make check passed
  in 15.82 seconds. Elapsed approximately 6 minutes for specification preparation.
- G09 foundation: added a reproducible bridge-only build launcher with rootless
  admission, source/policy pins, bounded resources and reduced environment.
  Seven mocked tests cover successful confinement, missing rootless evidence,
  unsafe paths/configuration, pin/preparation failure and compiler failure.
  Build infrastructure has no product toggle and does not approve engine images.
- Live build on `jerry`: directory `/var/tmp/commonroom-engine-build.oBvGZdku`,
  container `commonroom-reviewed-engine-build-oBvGZdku`, log `build.log` there.
  Upstream commit `092d3acd6bec3e3a14bdc7e7a2810ab628ab759d`; archive SHA-256
  `79c8ea8547e1dd9969f63ef6b76a4fb0d1a024ef36c0c627fbcd2ffb658d9312`.
  Source/policy/lock pins passed. The actual initial command uses 4 CPUs,
  8 GiB memory and 512 PIDs, all capabilities dropped, no-new-privileges,
  public source only and no credential/socket mounts. Future launcher also
  bounds compiler jobs and additional swap. No provider has been launched.
  All three bridge-router tests passed. The subsequent release build failed
  with a confirmed container OOM while compiling codex-tui; no release binary
  success is claimed. Preserve the failed container and log.
  A one-job retry uses the same pinned image/source and 8-GiB memory bound,
  with no additional swap allowance. Its public cache is copied to resume-cache;
  retry container is `commonroom-reviewed-engine-build-oBvGZdku-j1`, log
  `release-j1.log`. Verify cache-copy/start state before creating another retry.
  Reinspect exact container/log before retrying; do not start a competing build.
- Foundation make test passed in 0.70 seconds (cached); make check passed in
  6.78 seconds after the final resource-limit refinement. Live build elapsed
  time remains open until completion; it is separate from hermetic check timing.
- Review fix checks: make test 0.76 seconds; make check 6.54 seconds, both passed.
  Elapsed approximately 5 minutes including live failure diagnosis; retry build
  time remains open. Launcher now limits compilation to one job.
- Last notification: published PR verified; capability review fix ready locally.
  Three router tests passed but release build OOM failed; bounded retry underway.
  No release gate is certified. Await owner publication of the grouped fix.
- Mark retains signing, publication and merge control; nothing staged by agent.

## Previous batch evidence

The prior storage batch and exact-ID review fix passed make test and make check
before owner publication. PR #52 now contains that fix. Its synthetic cleanup
proof does not establish live daemon or descendant termination guarantees.

## Governance design branch

- Owner requested preservation of the Codex work and a separate branch for the
  provider-independent governance design.
- Active branch: `codex/provider-independent-governance`, created clean from
  `c3c76e95c45a788bade224e24750ee607919c4b6` after PR #53 merged.
- `codex/p0-release-gates` retains the committed P0/Codex build work. No work
  was discarded, stashed or committed by the agent during branch separation.
- Added accepted architecture decision, domain glossary and governance design
  with 13 pending adversarial acceptance scenarios. Runtime implementation has
  not changed; no governance case or P0 release gate is certified.
- Continue on the active branch only after checking actual Git state; do not
  mix unrelated Codex review fixes into governance changes or overwrite edits.
- Governance design checks passed: make test 8.37 seconds; make check 15.95
  seconds. Document links verified. Elapsed approximately 5 minutes.
- Last notification: separate design branch prepared; no staging/commit/push,
  runtime enforcement remains to implement and verify.

## Governance authority runtime increment

- Active branch: `codex/governance-authority`, from verified PR #54 merge
  `44f192333bfe0c05679972e6dd12df9f08205dab`. Prior branches retained.
- Implemented durable room appointments and owner-only explicit transfers,
  monotonic expected revisions, atomic audit/receipts, replay validation and
  idempotent selection. First newly registered eligible room agent is appointed
  atomically while enabled; reconnect/removal never silently replaces it.
- `agentCoordination` defaults off; server-enforced owner settings and room
  selection UI added. Off preserves history and independent work; no extra
  execution privileges or agent coordination grants are minted by selection.
- Mocked tests cover restart/replay, concurrency, forged agent identity, scope,
  malformed/duplicate input, audit failure, stops, first selection and feature off.
- User/technical Help: docs/governance-appointments.md. Full grant/delegation,
  task-proposal and independent escalation acceptance remain pending; no complete
  new security-model or live production-readiness claim is made.
- Last notification: appointment authority foundation implemented and undergoing
  final checks. Nothing staged/committed/pushed by the agent.
- Runtime checks passed: make test 8.92 seconds; final make check 5.82
  seconds (cached Go packages). Four browser-control tests passed; no live
  browser/provider or multi-controller deployment was exercised. Active elapsed
  work time was not separately measured.
- Publication checkpoint: owner appointment foundation ready, with full grant/
  delegation implementation explicitly pending. Owner signs/pushes/merges.

## Governance PR #55 review

- Published head `dc6d4d4ff7f5783ae0daa34056e3ab5e95adadd9`; PR #55 open.
- Valid P1 `4356133587`: automatic appointment persistence failures now latch
  integrity stop in registration, create and attach batches; optimistic conflicts
  do not. Tests exercise transaction begin/insert/commit failure and later refusal.
- Valid P2 `4356133592`: shared feature switches now treat agentCoordination as
  default off. Regression checks its initial display and first enable click.
- Both findings assessed against source and grouped into one local publication
  checkpoint on the existing branch. No GitHub discussion posted.
- Checks passed: make test 9.07 seconds; make check 18.19 seconds. Active
  elapsed work time was not separately measured. Tests use mocked boundaries;
  no live browser/provider or deployment acceptance was performed.
- Last notification: both review fixes verified and ready with one commit message;
  stay quiet awaiting owner publication and current-head rereview.
