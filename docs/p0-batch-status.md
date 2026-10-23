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

## Governance PR #56 handler regression review

- Current reviewed head: `2bfbca2f3890e089ed6b3f591e28d9fa2ee1b896`, branch
  `codex/governance-authority`; PR #56 open with successful Greptile check.
- Valid P2 inline comment `4360544342`: prior create/attach tests invoked the
  batch helper directly and could miss a handler reverting to ordinary append.
- Added authenticated handler-level create/attach tests with instance-scoped
  mocked login/history dependencies. Healthy enabled/disabled controls verify
  registration and appointment behavior. Failed commit verifies full rollback,
  required integrity-stop latching and subsequent appointment refusal.
- make test passed in 8.98 seconds; make check passed in 19.15 seconds.
  Initial sandboxed test attempt could not access the Go cache; required checks
  passed with cache/socket access. Active elapsed time approximately 6 minutes.
  All database/provider boundaries mocked; live authentication and deployment
  acceptance remain unverified. No execution or isolation contract changed.
- Last notification: handler regression fix ready for owner signing/publication
  on the same PR. Nothing staged, committed, pushed or posted by the agent.

## Revision-bound coordination turns

- Active branch: `codex/governance-grants`, created clean from verified PR #56
  merge `b9a8465b2a74b75ebd2c1d34e95597a5361239d3`. That merge contains exact
  reviewed head `e510a32bd6d4eb64f820ed6b9761bdafc781e2fa`; Greptile succeeded.
  Its prior handler-test finding was handled; no new review feedback pending.
- Implemented owner-requested coordination turns via the existing send handler,
  bound to exact company/room/coordinator/revision and a durable disable generation.
  Ordinary sends/tasks remain independent; no agent-issued grants, delegation,
  new task authorization or increased resource permissions are introduced.
- The binding survives the detached queue and managed job journal. Admission,
  submission, tool intent/outcome, native-session update and result publication
  revalidate. Journal replay refuses stripping/changing a job's dependency.
- Committed transfer/disable cancels local role-dependent turns only. Global audit
  integrity failure retains its existing broader safety stop. Off/on never revives
  old work. Cleanup and shutdown confirmation still require normal worker proof.
- Added owner room control, full user/API/security documentation and mocked
  handler, registration, broker, replay and browser-control regression coverage.
- make test passed in 8.86 seconds; make check passed in 18.41 seconds. Active
  elapsed wall time was not separately recorded. Initial test failures exposed
  synthetic key length and cancellation-test assumptions; both corrected.
- No real providers, database, Docker, deployed browser or multi-controller live
  runtime were exercised. Cross-controller cancellation is not immediate; stale
  effect acknowledgments/results refuse, with deadlines/reconciliation retained.
  In-flight effects already acknowledged are not claimed to be undone. Full
  GOV-06/07/08/09/10/13 and all live P0 release gates remain uncertified.
- Last notification: coherent revision-bound turn batch verified and ready for
  owner signing/publication, with one commit message. Nothing staged, committed,
  pushed or posted by the agent. Existing branches and adapter work preserved.

## PR #57 cross-controller authority review fix

- Reviewed exact head `2ea5223073d3d7bd43d5a04dc3f970daa5da5c5a` on
  `codex/governance-grants`, PR #57 open. Greptile check succeeded but valid P1
  inline comment `4361052481` identifies a validation-to-append race.
- Result/task completion, native-session callback persistence and managed
  submission now carry validated EventCount into the transaction. Conflicts
  reload and revalidate, with three bounded metadata-only attempts. Revoked
  results remain withheld; unrelated writes can retry. Ordinary conflicts do
  not falsely latch an integrity fault. Existing shutdown-proof semantics remain.
- Added deterministic mocked second-controller commits between validation and
  transaction start: transfer, disable, human stop, unrelated mutations and
  repeated conflicts across all three paths, plus independent-result control.
- make test passed in 9.38 seconds; make check passed in 20.08 seconds. Elapsed
  approximately six minutes, measured from 15:19:53 UTC. Initial test iteration
  corrected a missing import and used the real safety endpoint for stop requests.
- Database/provider boundaries remain mocked; no live providers, database,
  deployed browser or live multi-controller runtime exercised. Repeated conflicts
  can leave completion unresolved for inspection, never inferred as success.
- Last notification: P1 fix verified locally with one commit message; owner signs
  and publishes on the same PR before current-head rereview. Nothing staged,
  committed, pushed, merged or posted by the agent.

## Conversation task proposal batch

- Active branch: `codex/governance-task-proposals`, from verified PR #57 merge
  `880f20643ac9c5f796d375bdfc9ece8c2ed49d0c`. Merge tree matched exact reviewed
  head `aa68d55b462e50bdcaec48322b6e45d7755137e2`; Greptile passed with no new
  findings. Prior branches and Codex adapter work retained.
- Added owner-mediated capture/review of room message suggestions. Immutable
  source message/recorded authorship stays distinct from human capture identity,
  reviewed criteria and approval. Capture grants no authority or task execution.
- Atomic approval creates one unassigned Draft task. Rejection is terminal and
  creates none. Operation receipts deduplicate retries; expected revisions/event
  count prevent competing approvals or stale approval across disable/stop.
- Added controller-owned Task.ProposalID, preserved through normal edits and
  checked on replay. Draft dispatch refuses even with an assigned agent. Existing
  independent task permissions/execution barriers are unchanged.
- `taskProposals` defaults off, owner-only and server-enforced, no restart. Off
  refuses captures/decisions, preserves readable records and approved tasks, and
  does not cancel independent jobs. Settings/shared-switch defaults tested.
- UI presents original escaped source for review; in-app Help and user/API/setup
  docs describe source provenance, idempotency, flag behavior and limits.
- Tests exercise authenticated endpoints, scope/identity forgery, default/off,
  replay/restart, rollback/integrity failure, competing controllers, disable race,
  immutable provenance, assigned Draft refusal and synthetic UI request controls.
- make test passed in 9.23 seconds; make check passed in 18.93 seconds. Elapsed
  9 minutes 29 seconds from 16:56:58 to 17:06:27 UTC before checkpoint handoff.
  Tests mock all database/provider boundaries and launch no real providers/Docker.
- No deployed browser/screenshot, live multi-controller or provider acceptance
  performed. This is partial GOV-08: automatic agent-originated proposal channels,
  bounded delegation, resource approvals and independent objections remain pending.
  No full governance model or P0 release gate is certified.
- Last notification: proposal capture/review/Draft batch verified locally and ready
  for owner signing/publication with one commit message. Nothing staged, committed,
  pushed, merged or posted by the agent. Await owner publication/review on this batch.

## PR #58 personal-room replay review fix

- Current reviewed head: `71e67624d3336d5fc8ad3a3aa317268f2d18ebea`, branch
  `codex/governance-task-proposals`. PR #58 open; Greptile completed with new
  valid P1 `4362796233`: derived personal rooms were missing at proposal replay
  validation, making a successful capture poison subsequent state loading.
- Projection now materializes personal rooms from the owner and previously
  replayed actors/connections before proposal scope validation. No permissions
  or historical room-existence checks are bypassed; future actors cannot repair
  an invalid earlier proposal.
- Authenticated regressions cover owner, member and connection-agent personal
  rooms, first capture without profile/Help events, state/list reads, approval,
  rejection, restart and feature-off history. Unknown/future personal-room scope
  remains refused; private sources remain absent from shared-room lists.
- Prior missing-file findings `4362619352`, `4362619359`, `4362619374` on
  `2f44cb17186f47cfced9923664c562c47373a015` were resolved by owner publication;
  all five implementation/test/doc files are tracked in current head.
- make test passed in 9.19 seconds; make check passed in 18.81 seconds. Elapsed
  3 minutes 5 seconds from 18:22:05 to 18:25:10 UTC before checkpoint handoff.
  All database/provider boundaries mocked; live deployment/browser/provider
  acceptance remains unverified and no P0 release gate is certified.
- Last notification: personal-room replay fix verified locally, ready for owner
  signing/publication on the same PR. Nothing staged, committed, pushed, merged
  or posted by the agent. Await current-head rereview after owner publication.

## PR #58 derived personal-room label review fix

- Current reviewed head `545246aaff1ff01cd31bee210ada874afa599df9`, branch
  `codex/governance-task-proposals`; prior P1 personal-room replay fix published.
  PR #58 open; Greptile completed with valid non-blocking P2 `4362975561`:
  early projection froze personal-room labels before later actor renames.
- Existing derived personal-room labels now refresh from current actor names;
  connection-derived agent names follow current connections. Room IDs, personal
  ownership, saved settings and actor authority remain unchanged. Historical
  proposal source authorship is never renamed retroactively.
- Mocked shared/private proposal replay and restart regressions cover later
  member, owner and connection renames. A repeated-projection control verifies
  custom shared labels, personal settings and disabled actor permissions persist.
- make test passed in 9.25 seconds; make check passed in 18.74 seconds. Elapsed
  2 minutes 9 seconds from 18:30:48 to 18:32:57 UTC before checkpoint handoff.
  No real database/provider/browser or live deployment exercised.
- Last notification: derived-label review fix verified locally, ready for owner
  signing/publication on the same PR. Nothing staged, committed, pushed, merged
  or posted by the agent; await current-head rereview after owner publication.

## PR #58 custom personal-room name review fix

- Reviewed head `7765e8d878df4e7fe007392178ad2a69fd7d5381`, branch
  `codex/governance-task-proposals`. PR #58 open; Greptile completed with valid
  non-blocking P2 `4363039881`: refreshing every derived personal-room label
  overwrote custom names explicitly saved through Room settings.
- Replay now distinguishes automatic labels from custom names using historical
  room.updated values. This metadata is private/controller-derived, never trusted
  from clients. Personal rooms exist before applying their historical updates;
  saved personal ownership and other room fields are preserved.
- Only automatic names follow actor renames. Custom names persist through later
  settings saves, proposal capture, actor rename, restart and state refresh.
  Tests cover updates before/after proposal capture and unchanged automatic-name
  controls. Prior unknown/future-room and source-isolation regressions still pass.
- make test passed in 9.19 seconds; make check passed in 18.85 seconds. Observed
  wall-clock span 18:37:53–21:08:12 UTC (2 hours 30 minutes 19 seconds); active
  execution time was not separately measured. Tests use mocked boundaries;
  no real database/provider/browser or live deployment acceptance performed.
- Last notification: custom-name persistence fix verified locally, ready for
  owner signing/publication on the same PR. Nothing staged, committed, pushed,
  merged or posted by the agent; await current-head rereview after publication.
