# P0 batch checkpoint

Updated: 22 October 2026 (America/Toronto).

## Current batch

- Branch: `codex/p0-release-gates`.
- Base: verified clean merged main `a2112b6a233f08774cffaeb5bc88827a3e38b429`.
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
  Six mocked tests cover successful confinement, missing rootless evidence,
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
  Compilation is still running; router tests and binary success are NOT claimed.
  Reinspect exact container/log before retrying; do not start a competing build.
- Foundation make test passed in 0.70 seconds (cached); make check passed in
  6.78 seconds after the final resource-limit refinement. Live build elapsed
  time remains open until completion; it is separate from hermetic check timing.
- Last notification: sequential implementation started; G09 source compilation
  is active. No commit handoff or release-readiness claim.
- Mark retains signing, publication and merge control; nothing staged by agent.

## Previous batch evidence

The prior storage batch and exact-ID review fix passed make test and make check
before owner publication. PR #52 now contains that fix. Its synthetic cleanup
proof does not establish live daemon or descendant termination guarantees.
