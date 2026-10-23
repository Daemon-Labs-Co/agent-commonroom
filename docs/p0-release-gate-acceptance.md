# P0 release-gate acceptance specification

Status: acceptance contract, not execution evidence. Prepared 22 October 2026.
Audience: implementers, test authors, deployment operators and the release owner.
Scope: the nine open gates in [release readiness](release-readiness.md), grounded
in [FUNCTIONAL_SPEC.md](../FUNCTIONAL_SPEC.md). P1 collaboration and explicitly
later secrets-editor features are outside this release decision.

## Result and evidence rules

Each gate starts **NOT RUN**. Use **BLOCKED** when required implementation,
reviewed images, accounts or environments are missing; **FAIL** when an executed
assertion is violated; **PASS** only when every required automated and live check
has evidence for the candidate revision. A skipped check is never PASS. An
expected refusal passes only when its healthy positive control also passes.
No gate is currently certified by this specification.

For each run record: gate/case ID, UTC timestamp, operator, candidate commit,
controller and worker binary hashes, immutable image references and actual IDs,
provider/harness version, enabled flags and defaults, synthetic account/room/job/
repository/revision identifiers, action, expected and actual result, ordered
security effects, redacted evidence locations, cleanup outcome and reviewer.
Record BLOCKED dependencies explicitly. Never retain tokens, private transcripts,
credential file contents or raw secret-bearing output in reports or repository.
Store bounded redacted evidence outside Git; link only approved shareable evidence.
A rerun after a relevant change supersedes earlier results; do not reuse stale
image, binary or timing evidence to certify a changed candidate.

## Common setup, controls and teardown

1. Automated drivers call real application boundaries with every database,
   Docker, transport and provider interaction mocked. Use synthetic credentials,
   per-test temporary files and fake clocks. Never launch providers, SSH, Docker
   workloads or database connections from the automated suite. Observe actual
   effects; never copy catalog expectations into passing observations.
2. Live runs are separate, explicit operator procedures on a dedicated Linux
   rootless worker and dedicated synthetic repositories/provider accounts. Use
   a disposable acceptance installation, never production messages or databases.
   Confirm trusted SSH host identity, reviewed immutable engine/tool images,
   binary provenance and enforceable cgroup limits before model execution.
3. Test owner, non-owner, foreign-room and foreign-connection actors. For every
   optional capability run documented default/off, authorized on, unauthorized
   changes and disable-during-work cases. Disabling preserves data/history and
   follows documented running-work policy. Missing server enforcement is FAIL.
   Mandatory isolation, authentication, authorization, audit, quotas and retention
   have no bypass flag, paid exemption, diagnostic exception or host fallback.
4. Use two synthetic repositories/revisions and a synthetic credential canary.
   Inventory allocations before and after each run. Capture exact role IDs,
   workspace identities, durable receipts and account ledger revisions.
5. On completion, cancellation or failure, revoke tool authority and reconcile
   exact role IDs. Prove absence using fresh inventory; a removal request alone
   is insufficient. Preserve source and storage charges unless separately
   authorized exact disposal succeeds. Unknown absence retains reservations and
   blocks replacement. Never clean unrelated containers, volumes or history.
6. Restore test-only settings and revoke temporary acceptance grants. Record
   retained resources and cleanup uncertainty rather than hiding leftovers.

## Dependency and execution order

| Order | Gate | Prerequisites |
| --- | --- | --- |
| 1 | G09 Live platform foundation | Dedicated worker, reviewed images and synthetic installation |
| 2 | G04 Storage | Inventory and physical quota backend on that worker |
| 3 | G02 Providers; G03 Credentials; G05 Upgrades | Bridge-only engine builds, dedicated accounts and signed review publication |
| 4 | G01 Production dispatch | Foundation, provider, credential, storage and review admission boundaries |
| 5 | G06 Idle; G07 Commands/artifacts; G08 Disposal | Production allocation and durable authenticated audit path |
| 6 | G09 End-to-end failure acceptance | All previous gates integrated on the same candidate |

G09 is checked twice: its platform checks enable safe testing; its final failure
matrix closes the release. Internal component success may precede dependencies,
but cannot substitute for the final assembled production path.

## G01 — Production dispatch

Requirements: FR-09, FR-13, FR-15–17, FR-20, FR-32–33.
Catalog: admission (8 cases).
Prerequisites: authenticated synthetic installation, approved repository allowlist,
reviewed engine/tool images, trusted worker and durable audit/storage admission.

| ID | Procedure | Required result/evidence |
| --- | --- | --- |
| G01-A | Submit bounded repository work as its authorized human through the intended production entry point. | One durable allocation binds company/human/room/participant/connection/job/repository/deadline; capacity reserved before create; fresh distinct roles/workspace; authenticated callback audit; observed result and exact fencing. Internal helper invocation alone is insufficient. |
| G01-B | Repeat as unauthorized/foreign actor, flag off, unreviewed image, unapproved worker and non-rootless worker. | Refusal before any engine/tool start or unauthorized durable allocation; actionable sanitized reason. Healthy authorized control succeeds. |
| G01-C | Fail audit intent and outcome persistence separately and disconnect the authenticated audit transport. | No action before durable intent; no result before durable outcome acknowledgment; authority revoked and integrity stop/uncertainty preserved across controller restart. |
| G01-D | Refresh/disconnect/reconnect the browser before and after submission acknowledgment; replay the operation identity. | Same committed job/attempt, no duplicate dispatch/result; human stop wins; uncertain work reconciles before retry. |
| G01-E | Change membership, repository authorization, deadline or stop state between admission and execution. | Revalidation denies stale authority without fallback; unauthorized effects never occur. |

Pass: all rows through the public production path, automated denial/race controls
and a live authorized round trip. BLOCKED while dispatch is only internal.

## G02 — Required providers and capability contracts

Requirements: FR-03, FR-13, FR-16, FR-29, FR-33.
Catalog: providers (9 cases).
Prerequisites: reviewed isolated adapters/images for Codex, Claude, Gemini,
GitHub Copilot and llama.cpp; each provider's supported dedicated authentication
and billing configuration (local inference may not require user authentication).

| ID | Procedure | Required result/evidence |
| --- | --- | --- |
| G02-A | For each of the five providers, admit fresh bounded work and request an authorized edit/check through its broker callback. | Correct declared provider/version, fresh pair/workspace, actual authenticated callback, independently recorded launches proving no host model/tool execution, observed result and both-role absence. Run all five; one provider cannot stand in for another. |
| G02-B | Remove required authentication, request unsupported tools and register a synthetic future read-only adapter. | Missing auth blocks before start; undeclared capabilities refuse; declared read resolves through the real contract; no silent API or billing fallback. Future-adapter test does not require arbitrary plugin product execution. |
| G02-C | Cancel a running provider and interrupt its stream. Request subsequent work after reconciliation. | Durable cancellation, no late success, exact pair fencing, retained source and fresh resources for later authorized work; old resources never restarted/adopted. |
| G02-D | Ask the engine to invoke native shell/filesystem, MCP/plugin or alternate execution routes outside the declared broker. | Routes unavailable/refused; synthetic engine credential cannot reach tool workspace, audit or artifacts. Inspect image contents/policy, not just version output. |

Pass: automated contracts plus each provider's live callback/refusal/cancellation
record. A stock CLI or metadata/version command does not pass bridge isolation.

## G03 — Credential renewal and recovery

Requirements: FR-03, FR-13, FR-17, FR-20.
Catalog: credentials (4 cases).
Prerequisites: reviewed credential custody/export path and a dedicated expiring
provider session. Use synthetic credential files for automated failure tests.

| ID | Procedure | Required result/evidence |
| --- | --- | --- |
| G03-A | Refresh a dedicated session during isolated execution, export privately and persist after exact fencing. | Atomic validated update belongs to the original connection; no partial overwrite; final credential remains private; recorded outcome contains no values. |
| G03-B | Race two jobs refreshing the same connection; try an export for another connection. | Lease serializes custody; foreign update refused; no lost update or cross-connection overwrite. |
| G03-C | Fail write/rename/export acknowledgment and interrupt transport or crash at custody milestones. Restart. | Prior credential remains valid or state explicitly uncertain; new execution blocked; no invented persistence success or blind retry. |
| G03-D | Perform explicit supported sign-in recovery after uncertainty. | Required actor is authorized; recovery re-establishes verified custody before admission resumes; secrets absent from audit/logs/artifacts. |

Pass: all four hermetic catalog cases and live refresh/recovery on applicable
required adapters. A directory's existence is not authentication evidence.

## G04 — Aggregate storage and container capacity

Requirements: FR-13, FR-20, FR-32–33.
Catalog: storage (10 cases).
Prerequisites: complete trusted inventory, durable account journal, physical
worker quota backend and production allocation/disposal bindings.

| ID | Procedure | Required result/evidence |
| --- | --- | --- |
| G04-A | Inventory workspace, staged-bundle and artifact usage; allocate at exactly 5,000,000,000 bytes, then request one more byte. | Exact boundary allowed, extra byte denied before allocation; retained data charged; checked integer arithmetic rejects negative/overflow/malformed evidence. |
| G04-B | Race final-byte and final-slot allocations across rooms/workers, then restart controller. Use independent accounts as a control. | One winner at capacity; no oversubscription after replay; account identity enforced. Four physical managed containers maximum; a pair consumes two. Independent accounts remain independent. |
| G04-C | Fail or truncate inventory; attempt owner/paid/diagnostic override. | Unknown usage blocks allocation and unmetered writes; no missing inventory interpreted as zero; no bypass. |
| G04-D | Write through every supported tool/output/export path until the physical limit, including concurrent jobs. | Actual worker storage allocation is bounded at the account ceiling, not just journal reservations. Record backend measurements and refusal; no host/shared-store escape. |
| G04-E | Reconcile proven absent engine/tool roles while retaining source. Then perform separately approved disposal. | Container slots released only on proven absence; retained bytes remain charged. Exact storage credit released only after authorized verified disposal. Ambiguity retains capacity. |

Pass: all accounting cases, live hard-limit measurements and production allocation
integration. G04-E is intentionally not satisfied by the current conservative
version-4 behavior that retains slots until workspace disposal.

## G05 — Reviewed harness upgrades

Requirements: FR-13, FR-17, FR-30, FR-33.
Catalog: upgrades (7 cases).
Prerequisites: trusted review key, upstream discovery, compatibility review,
immutable image build/publication/staging and production instantiation checks.

| ID | Procedure | Required result/evidence |
| --- | --- | --- |
| G05-A | Discover a compatible candidate; review/build/publish it and instantiate a fresh agent. Repeat with the current approved version. | Signed review checked at instantiation; actual staged repository digest matches approved image; selected version/outcome audited; fresh pair executes callback and fences. |
| G05-B | Supply incompatible/unreviewed candidate, unavailable review source, wrong signature, rollback or mismatched digest/repository metadata. | New execution blocks with a sanitized actionable reason; no tag substitution, unreviewed fallback or mutation of review history. Exercise authorized reviewed rollback under the documented rollback policy separately. |
| G05-C | Publish an update while an existing reviewed engine is active. | Running image/config/allocation remain immutable; only a fresh agent selects the newly reviewed release. |
| G05-D | Inspect built binary and test callback/native-route refusal using the exact published image. | Source, compiler, binary and image provenance linked; compatibility is observed, not inferred from a version string. |

Pass: catalog checks plus actual discovery-to-publication-to-launch evidence and
live compatibility. Disabling optional discovery cannot admit unreviewed code.

## G06 — Idle shutdown and fresh subsequent work

Requirements: FR-16–17, FR-20, FR-31, FR-33.
Catalog: idle (9 cases).
Prerequisites: production policy snapshots, owner settings and authenticated
pre-cleanup announcement. Resolve the three-minute job lifetime versus five-minute
idle-default conflict without relaxing mandatory execution bounds implicitly.

| ID | Procedure | Required result/evidence |
| --- | --- | --- |
| G06-A | Leave an eligible idle agent at default settings; inspect at 299 seconds and at 300 seconds. | No premature stop; at default expiry a durable announcement precedes normal pair fencing. Retain source/history. Use fake clocks automatically and real elapsed policy live. |
| G06-B | Keep active work beyond interval; queue work just before expiry; complete and wait a fresh interval. | Active/queued work protected; completion resets the interval; no old timer races new work. |
| G06-C | Owner changes interval; non-owner attempts it; disable the optional idle policy. | Server authorizes only owner changes; invalid settings refused; immutable allocation snapshot follows documented update policy; off does not disable job deadlines or other safety limits. |
| G06-D | Fail announcement/cleanup/absence inventory and restart; submit later work. | Unacknowledged announcement cannot silently authorize cleanup; uncertainty remains blocked, no remnant reuse. After proven stop, later authorized work gets fresh role IDs/workspace. |

Pass: all nine catalog cases, timer-boundary races and live assembled policy.

## G07 — Commands, approvals, evidence and artifacts

Requirements: FR-09, FR-13–14, FR-17, FR-20, FR-32–33.
Catalog: commands (10 cases) and artifacts (8 cases).
Prerequisites: isolated command grants, revision-bound execution receipts,
external-action approval boundary, artifact store/accounting and authorized export.

| ID | Procedure | Required result/evidence |
| --- | --- | --- |
| G07-A | Run an explicitly granted bounded check at revision r1, with command capability on; repeat off and without permission. | Durable scoped intent before invocation, outcome acknowledgment before result release; receipt binds actual revision, argv/profile and outcome. Off/ungranted runs have no execution. |
| G07-B | Exceed fixed timeout/output limits; escape cwd through a symlink; fail intent/outcome persistence. | Refuse/terminate safely, bounded output, no host effect or unacknowledged result; exact container/descendant termination proof; source retained and stale authority revoked. |
| G07-C | Request an irreversible publish/deploy/message/spend action without approval, then with approval bound to r1 and changed r2/target/scope. | Exact intent announced; absent approval blocks; material change invalidates approval; no external effect. A positive control uses a simulated external-effect recorder in automation and a separately explicitly approved synthetic live target. Silence is never approval. |
| G07-D | Produce observed check/artifact for r1; compare a model-only claim and stale r1 evidence after r2. | Validated audited delivery only for authorized current evidence; observed versus model claim labelled accurately; stale evidence refused. Human acceptance remains separate from completion. |
| G07-E | Export over size bound, foreign-room artifact, traversal/symlink path, credential canary or artifact with exhausted storage. | Each denied before delivery/unmetered storage; no engine secrets in output/audit; ledger and filesystem remain consistent. Healthy authorized bounded export succeeds. |

Pass: all eighteen catalog cases plus live compiler/check, revision binding,
authorized artifact delivery and failure containment. Arbitrary shell support is
not required; the supported bounded profiles must be documented and enforced.

## G08 — Explicit approved disposal

Requirements: FR-13–14, FR-17, FR-20, FR-32–33.
Catalog: disposal (9 cases); shared storage/artifact checks G04/G07.
Prerequisites: stopped-worker identity, exact workspace review, production storage
allocation binding and off-by-default owner disposal capability.

| ID | Procedure | Required result/evidence |
| --- | --- | --- |
| G08-A | Enable disposal as authorized operator; owner reviews exact workspace, waits at least one minute and confirms before one-hour expiry. | Exact identity and current scope checked; durable one-shot claim before removal; absence receipt binds request; only exact prebound storage credit and eligible slots released atomically. Test immediately before/at delay and expiry boundaries. |
| G08-B | Omit/revoke/expire approval, change volume identity, use another actor, replay claim or switch capability off. | No deletion or credit release; durable history preserved; default off enforced on server. |
| G08-C | Fail removal, lose absence inventory or crash after claim/removal before receipt acknowledgment. Restart. | Unknown outcome stays blocked, no blind duplicate deletion or credit; explicit reconciliation proves identity/absence before settlement. |
| G08-D | Cancel/idle-stop ordinary work or request audit-history purge. | Pair fenced without implicit source disposal; audit/history cannot be deleted by disposal approval or a flag. Artifacts follow separately documented retention/accounting ownership. |

Pass: nine catalog cases and live approved/failed disposal using disposable
synthetic volumes. No approval in this document authorizes deleting real data.

## G09 — Live platform and assembled failure acceptance

Requirements: FR-03, FR-13–17, FR-20, FR-29–33.
Catalog: cross-cutting; this gate has additional live cases below and is not a
new alias for any one of the catalog's nine areas.

| ID | Procedure | Required result/evidence |
| --- | --- | --- |
| G09-A | Inspect actual rootless daemon, image metadata/content, users, mounts, read-only roots, capabilities, seccomp, networking, memory/CPU/PID/storage limits; attempt controlled synthetic violations. | Kernel/daemon measurements and observed refusal/limits. Engine has only private provider custody/transport; tool is credential-free with approved workspace only. No socket, host repository/home or unapproved network escape. |
| G09-B | Run each provider's approved bounded task through the production controller, browser, authenticated audit and worker. | One correlated account/job/revision chain, observed checks, quotas, durable outcome and exact cleanup across all components. |
| G09-C | Interrupt audit SSH during a synthetic command; separately drop engine stream, refresh browser and restart controller. | No new tool effect after authority loss, no unacknowledged success, no duplicate job/result; committed state recovered and ambiguity visible. |
| G09-D | Crash worker at journal/create/bind/start, command intent/outcome and credential/disposal milestones; restart/reconcile. | Fresh inventory and durable evidence prevent adoption/restart/unsafe retry; no lost slots/bytes/history; uncertainty blocks replacements. Use a fresh fixture per milestone. |
| G09-E | Exercise resource exhaustion, command descendants, concurrent account capacity, idle expiry, reviewed upgrade and approved disposal together. | Enforced aggregate bounds, both-role/descendant absence, correct retained charges and no bypass between subsystems. Failure in one gate cannot be hidden by a successful other subsystem. |

Pass: all live rows on the exact candidate, all five provider records, successful
teardown or explicitly retained authorized source, and independent release review.
Initial tool-only evidence and downloaded build bases do not satisfy this gate.

## Running checks and recording the release decision

```sh
make test-p0-plan    # Shows bound/pending scenarios; not a release verdict
make test           # Normal hermetic suite
make check          # Required race/static/frontend/documentation checks
make test-p0-ready  # Strict: every catalog driver must exist and match behavior
```

Run make test during implementation and make check before handoff. Record wall
times and elapsed work, failures and meaningful integration gaps. The strict
runner currently intentionally fails for unbound cases. No command above runs
live acceptance or certifies all nine gates. Do not add stub drivers to turn it
green. Writing this specification needs no product toggle; implemented optional
features still require documented defaults, authority, restart/running-work and
both-state behavior under CONTRIBUTING.md.

Copy this row for each gate/case into the acceptance run record:

| Candidate | Gate/case | Mode | Preconditions | Expected | Actual/evidence | Teardown | Result | Reviewer |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| commit + binary/image hashes | Gxx-X / catalog ID | automated/live | setup/flags | contract | redacted observations/link | absence/retention | NOT RUN/BLOCKED/FAIL/PASS | name/date |

Release sign-off requires nine PASS decisions and all 74 strict catalog cases
passing for the same candidate, with no unresolved isolation/audit/authorization/
retention/capacity failure. Mark owns publication and merge. This specification
neither creates implementation nor grants authority to publish, spend or delete.

## Complete catalog mapping

The following mapping covers every existing scenario exactly once by primary
gate. G09 supplies cross-cutting live verification for all of them. Expected
ordered effects and forbidden effects remain authoritative in the linked
[catalog](../cmd/server/testdata/p0_remaining_acceptance.json); the readable
listing below includes those expectations so implementers can plan drivers.

### dispatch-approved — G01

Requirements: FR-13, FR-33.

Setup: owner=authorized; image=reviewed; topology=linux-rootless-ssh.

Action: dispatch approved repository work.

Expected decision: `allow`. Ordered effects: `fresh_pair` → `authenticated_audit` → `observed_result` → `fenced_pair`.

Additional forbidden effects: none beyond global prohibitions.

Required observations: `same_job=true`; `same_repository=true`.

### dispatch-unauthorized — G01

Requirements: FR-13, FR-20.

Setup: owner=unauthorized.

Action: dispatch development.

Expected decision: `deny`. Ordered effects: none.

Additional forbidden effects: `engine_started`, `tool_executed`.

Required observations: decision/effects above.

### dispatch-feature-off — G01

Requirements: FR-13, FR-20.

Setup: execution_feature=off.

Action: dispatch development.

Expected decision: `deny`. Ordered effects: none.

Additional forbidden effects: `engine_started`, `tool_executed`.

Required observations: decision/effects above.

### dispatch-unreviewed-image — G01

Requirements: FR-13, FR-20.

Setup: image=unreviewed.

Action: dispatch development.

Expected decision: `deny`. Ordered effects: none.

Additional forbidden effects: `engine_started`, `tool_executed`.

Required observations: decision/effects above.

### dispatch-unapproved-worker — G01

Requirements: FR-13, FR-20.

Setup: worker=unapproved.

Action: dispatch development.

Expected decision: `deny`. Ordered effects: none.

Additional forbidden effects: `engine_started`, `tool_executed`.

Required observations: decision/effects above.

### dispatch-no-rootless — G01

Requirements: FR-13, FR-20.

Setup: worker=non-rootless.

Action: dispatch development.

Expected decision: `deny`. Ordered effects: none.

Additional forbidden effects: `engine_started`, `tool_executed`.

Required observations: decision/effects above.

### dispatch-audit-unavailable — G01

Requirements: FR-13, FR-17.

Setup: audit=unavailable.

Action: dispatch development.

Expected decision: `block`. Ordered effects: none.

Additional forbidden effects: `tool_executed`, `success_released`.

Required observations: decision/effects above.

### browser-disconnect — G01

Requirements: FR-16.

Setup: job=authorized-running; browser=disconnected.

Action: refresh browser and reconnect.

Expected decision: `allow`. Ordered effects: `same_job_recovered`.

Additional forbidden effects: `duplicate_dispatch`.

Required observations: `attempt_count=1`.

### provider-codex — G02

Requirements: FR-03, FR-29.

Setup: provider=codex; transport=synthetic; auth=supported-user-auth.

Action: run isolated callback turn.

Expected decision: `allow`. Ordered effects: `fresh_pair` → `callback_dispatched` → `fenced_pair`.

Additional forbidden effects: `silent_api_fallback`.

Required observations: `provider=codex`; `host_launches=0`.

### provider-claude — G02

Requirements: FR-03, FR-29.

Setup: provider=claude; transport=synthetic; auth=supported-user-auth.

Action: run isolated callback turn.

Expected decision: `allow`. Ordered effects: `fresh_pair` → `callback_dispatched` → `fenced_pair`.

Additional forbidden effects: `silent_api_fallback`.

Required observations: `provider=claude`; `host_launches=0`.

### provider-gemini — G02

Requirements: FR-03, FR-29.

Setup: provider=gemini; transport=synthetic; auth=supported-user-auth.

Action: run isolated callback turn.

Expected decision: `allow`. Ordered effects: `fresh_pair` → `callback_dispatched` → `fenced_pair`.

Additional forbidden effects: `silent_api_fallback`.

Required observations: `provider=gemini`; `host_launches=0`.

### provider-copilot — G02

Requirements: FR-03, FR-29.

Setup: provider=copilot; transport=synthetic; auth=supported-user-auth.

Action: run isolated callback turn.

Expected decision: `allow`. Ordered effects: `fresh_pair` → `callback_dispatched` → `fenced_pair`.

Additional forbidden effects: `silent_api_fallback`.

Required observations: `provider=copilot`; `host_launches=0`.

### provider-llama-cpp — G02

Requirements: FR-03, FR-29.

Setup: provider=llama.cpp; transport=synthetic; auth=supported-user-auth.

Action: run isolated callback turn.

Expected decision: `allow`. Ordered effects: `fresh_pair` → `callback_dispatched` → `fenced_pair`.

Additional forbidden effects: `silent_api_fallback`.

Required observations: `provider=llama.cpp`; `host_launches=0`.

### provider-missing-auth — G02

Requirements: FR-03, FR-29.

Setup: provider=gemini; auth=missing.

Action: instantiate agent.

Expected decision: `block`. Ordered effects: none.

Additional forbidden effects: `engine_started`, `silent_api_fallback`.

Required observations: decision/effects above.

### provider-unsupported-capability — G02

Requirements: FR-29.

Setup: provider=synthetic-custom; capabilities=chat-only.

Action: request workspace command.

Expected decision: `deny`. Ordered effects: none.

Additional forbidden effects: `host_execution`, `tool_executed`.

Required observations: decision/effects above.

### provider-custom-registration — G02

Requirements: FR-29.

Setup: provider=synthetic-future-adapter; adapter=test-only-capability-contract; capabilities=read.

Action: resolve declared adapter capabilities.

Expected decision: `allow`. Ordered effects: `capability_contract_resolved`.

Additional forbidden effects: none beyond global prohibitions.

Required observations: `undeclared_capabilities_denied=true`.

### provider-cancel — G02

Requirements: FR-16, FR-33.

Setup: provider=copilot; job=running.

Action: cancel and reconcile.

Expected decision: `allow`. Ordered effects: `cancel_recorded` → `fenced_pair`.

Additional forbidden effects: `success_released`.

Required observations: `old_container_reusable=false`.

### refresh-atomic — G03

Requirements: FR-03, FR-13.

Setup: auth=synthetic-subscription; refresh=new-valid-token.

Action: persist engine refresh after turn.

Expected decision: `allow`. Ordered effects: `private_refresh_validated` → `credential_persisted` → `credential_lease_released`.

Additional forbidden effects: `credential_in_audit`.

Required observations: `mode=0600`; `connection=original`.

### refresh-concurrent — G03

Requirements: FR-13.

Setup: connection=same; workers=two; refresh=concurrent.

Action: refresh credentials concurrently.

Expected decision: `allow`. Ordered effects: `refresh_serialized`.

Additional forbidden effects: none beyond global prohibitions.

Required observations: `overlapping_refreshes=0`; `lost_updates=0`.

### refresh-persist-fails — G03

Requirements: FR-03, FR-13.

Setup: refresh=valid; persistence=fails.

Action: save refreshed credentials.

Expected decision: `block`. Ordered effects: `auth_recovery_required`.

Additional forbidden effects: `refresh_success_reported`.

Required observations: `prior_credential_preserved=true`.

### refresh-foreign-connection — G03

Requirements: FR-13.

Setup: refresh_connection=foreign.

Action: save refreshed credentials.

Expected decision: `deny`. Ordered effects: none.

Additional forbidden effects: `credential_replaced`.

Required observations: decision/effects above.

### storage-exact-boundary — G04

Requirements: FR-32.

Setup: account_used_bytes=4999999999; reserve_bytes=1.

Action: reserve storage.

Expected decision: `allow`. Ordered effects: `storage_reserved`.

Additional forbidden effects: none beyond global prohibitions.

Required observations: `used_plus_reserved_bytes=5000000000`.

### storage-one-byte-over — G04

Requirements: FR-32.

Setup: account_used_bytes=5000000000; reserve_bytes=1.

Action: reserve storage.

Expected decision: `deny`. Ordered effects: none.

Additional forbidden effects: `storage_allocated`, `tool_executed`.

Required observations: decision/effects above.

### storage-concurrent-last-byte — G04

Requirements: FR-32.

Setup: account_used_bytes=4999999999; workers=two; each_reserves_bytes=1.

Action: reserve concurrently.

Expected decision: `allow`. Ordered effects: `reservation_serialized`.

Additional forbidden effects: none beyond global prohibitions.

Required observations: `successful_reservations=1`; `used_plus_reserved_bytes=5000000000`.

### storage-count-retained — G04

Requirements: FR-32.

Setup: retained_workspace_bytes=3000000000; staged_bundle_bytes=1000000000; artifact_bytes=1000000000; reserve_bytes=1.

Action: reserve additional storage.

Expected decision: `deny`. Ordered effects: none.

Additional forbidden effects: `storage_allocated`.

Required observations: decision/effects above.

### storage-metering-unknown — G04

Requirements: FR-32.

Setup: meter=unavailable.

Action: allocate workspace.

Expected decision: `block`. Ordered effects: none.

Additional forbidden effects: `storage_allocated`, `engine_started`.

Required observations: decision/effects above.

### storage-negative — G04

Requirements: FR-32.

Setup: reserve_bytes=-1.

Action: reserve storage.

Expected decision: `deny`. Ordered effects: none.

Additional forbidden effects: `storage_reserved`.

Required observations: decision/effects above.

### storage-overflow — G04

Requirements: FR-32.

Setup: reserve_bytes=18446744073709551616.

Action: decode internal reservation evidence.

Expected decision: `deny`. Ordered effects: none.

Additional forbidden effects: `storage_reserved`.

Required observations: decision/effects above.

### storage-retention-no-credit — G04

Requirements: FR-32, FR-33.

Setup: workspace=retained; containers=absent.

Action: reconcile job.

Expected decision: `allow`. Ordered effects: `container_slots_released`.

Additional forbidden effects: `storage_reservation_released`.

Required observations: `workspace_bytes_counted=true`.

### storage-cross-account — G04

Requirements: FR-32.

Setup: account_a_used_bytes=5000000000; account_b_used_bytes=0; account=b; reserve_bytes=1.

Action: reserve storage.

Expected decision: `allow`. Ordered effects: `storage_reserved`.

Additional forbidden effects: none beyond global prohibitions.

Required observations: `account_a_used_bytes=5000000000`; `account_b_used_bytes=1`.

### storage-no-paid-bypass — G04

Requirements: FR-20, FR-32.

Setup: account_used_bytes=5000000000; reserve_bytes=1; paid_limit=unset; owner=true; diagnostic_flags=off.

Action: reserve storage.

Expected decision: `deny`. Ordered effects: none.

Additional forbidden effects: `storage_allocated`.

Required observations: decision/effects above.

### upgrade-reviewed — G05

Requirements: FR-30.

Setup: installed=old-reviewed; available=new-compatible-reviewed.

Action: instantiate agent.

Expected decision: `allow`. Ordered effects: `upgrade_checked` → `compatibility_verified` → `reviewed_image_selected` → `fresh_pair` → `callback_dispatched` → `fenced_pair` → `workspace_retained`.

Additional forbidden effects: none beyond global prohibitions.

Required observations: `running_version=synthetic-reviewed-2`; `version_audited=true`.

### upgrade-already-current — G05

Requirements: FR-30.

Setup: installed=current-reviewed; available=current-reviewed.

Action: instantiate agent.

Expected decision: `allow`. Ordered effects: `upgrade_checked` → `reviewed_image_selected` → `fresh_pair` → `callback_dispatched` → `fenced_pair` → `workspace_retained`.

Additional forbidden effects: `running_container_mutated`.

Required observations: decision/effects above.

### upgrade-incompatible — G05

Requirements: FR-30.

Setup: available=incompatible.

Action: instantiate agent.

Expected decision: `block`. Ordered effects: `upgrade_failure_reported`.

Additional forbidden effects: `engine_started`, `running_container_mutated`.

Required observations: decision/effects above.

### upgrade-unreviewed — G05

Requirements: FR-30.

Setup: available=unreviewed.

Action: instantiate agent.

Expected decision: `block`. Ordered effects: `upgrade_failure_reported`.

Additional forbidden effects: `engine_started`, `running_container_mutated`.

Required observations: decision/effects above.

### upgrade-unavailable — G05

Requirements: FR-30.

Setup: available=check-fails.

Action: instantiate agent.

Expected decision: `block`. Ordered effects: `upgrade_failure_reported`.

Additional forbidden effects: `engine_started`, `running_container_mutated`.

Required observations: decision/effects above.

### upgrade-bad-digest — G05

Requirements: FR-30.

Setup: available=digest-mismatch.

Action: instantiate agent.

Expected decision: `block`. Ordered effects: `upgrade_failure_reported`.

Additional forbidden effects: `engine_started`, `running_container_mutated`.

Required observations: decision/effects above.

### upgrade-live-agent — G05

Requirements: FR-30, FR-33.

Setup: job=active; available=new-compatible-reviewed.

Action: check next-instantiation upgrade.

Expected decision: `allow`. Ordered effects: `upgrade_checked`.

Additional forbidden effects: `running_container_mutated`, `old_container_restarted`.

Required observations: `active_version_unchanged=true`.

### idle-before-default — G06

Requirements: FR-31.

Setup: clock=fake; idle_seconds=299; owner_config=default.

Action: tick idle timer.

Expected decision: `allow`. Ordered effects: none.

Additional forbidden effects: `shutdown_started`.

Required observations: `running=true`.

### idle-default-expiry — G06

Requirements: FR-31, FR-33.

Setup: clock=fake; idle_seconds=300; owner_config=default.

Action: tick idle timer.

Expected decision: `allow`. Ordered effects: `idle_announced` → `fenced_pair`.

Additional forbidden effects: `workspace_discarded`.

Required observations: `history_preserved=true`.

### idle-owner-interval — G06

Requirements: FR-31.

Setup: clock=fake; owner_config_seconds=120; idle_seconds=120.

Action: tick idle timer.

Expected decision: `allow`. Ordered effects: `idle_announced` → `fenced_pair`.

Additional forbidden effects: `workspace_discarded`.

Required observations: decision/effects above.

### idle-active-work — G06

Requirements: FR-31.

Setup: clock=fake; elapsed_seconds=600; work=active.

Action: tick idle timer.

Expected decision: `allow`. Ordered effects: none.

Additional forbidden effects: `idle_shutdown`.

Required observations: `running=true`.

### idle-arriving-work — G06

Requirements: FR-31.

Setup: clock=fake; idle_seconds=299; arrival=new-authorized-work.

Action: deliver work then tick.

Expected decision: `allow`. Ordered effects: `idle_timer_reset`.

Additional forbidden effects: `idle_shutdown`.

Required observations: decision/effects above.

### idle-unauthorized-config — G06

Requirements: FR-31, FR-20.

Setup: actor=non-owner; requested_seconds=1.

Action: change idle interval.

Expected decision: `deny`. Ordered effects: none.

Additional forbidden effects: `idle_config_changed`.

Required observations: decision/effects above.

### idle-feature-off — G06

Requirements: FR-31.

Setup: idle_feature=off; idle_seconds=600.

Action: tick idle timer.

Expected decision: `allow`. Ordered effects: none.

Additional forbidden effects: `idle_shutdown`, `history_deleted`.

Required observations: decision/effects above.

### idle-cleanup-uncertain — G06

Requirements: FR-31, FR-33.

Setup: clock=fake; idle_seconds=60; owner_config_seconds=60; cleanup=uncertain.

Action: expire then request replacement.

Expected decision: `block`. Ordered effects: `idle_announced` → `partial_pair_cleanup` → `shutdown_unconfirmed` → `execution_stopped`.

Additional forbidden effects: `replacement_started`, `container_slots_released`.

Required observations: decision/effects above.

### idle-new-work-fresh — G06

Requirements: FR-31, FR-33.

Setup: previous=idle-stopped-proven; work=new-authorized.

Action: dispatch work.

Expected decision: `allow`. Ordered effects: `fresh_pair` → `fresh_workspace`.

Additional forbidden effects: `old_container_restarted`.

Required observations: `reused_container_ids=0`.

### command-approved — G07

Requirements: FR-09, FR-13, FR-17.

Setup: argv=synthetic-compiler --check; permission=granted; revision=r1.

Action: run bounded workspace command.

Expected decision: `allow`. Ordered effects: `intent_durable` → `isolated_command` → `outcome_durable` → `result_released`.

Additional forbidden effects: none beyond global prohibitions.

Required observations: `revision=r1`; `host_launches=0`.

### command-no-permission — G07

Requirements: FR-13, FR-20.

Setup: permission=missing.

Action: run workspace command.

Expected decision: `deny`. Ordered effects: none.

Additional forbidden effects: `tool_executed`.

Required observations: decision/effects above.

### command-feature-off — G07

Requirements: FR-13, FR-20.

Setup: command_feature=off.

Action: run workspace command.

Expected decision: `deny`. Ordered effects: none.

Additional forbidden effects: `tool_executed`.

Required observations: decision/effects above.

### command-timeout — G07

Requirements: FR-13, FR-16.

Setup: clock=fake; command=never-exits; timeout=exceeded.

Action: run command.

Expected decision: `block`. Ordered effects: `intent_durable` → `isolated_command` → `command_cancelled` → `tool_container_absent`.

Additional forbidden effects: `success_released`.

Required observations: decision/effects above.

### command-output-limit — G07

Requirements: FR-13.

Setup: command=synthetic-large-output; output=over-bound.

Action: run command.

Expected decision: `block`. Ordered effects: `intent_durable` → `isolated_command` → `output_limit_reported` → `tool_container_absent`.

Additional forbidden effects: `unbounded_output`, `success_released`.

Required observations: decision/effects above.

### command-audit-intent-fails — G07

Requirements: FR-17.

Setup: audit_intent=fails.

Action: run command.

Expected decision: `block`. Ordered effects: none.

Additional forbidden effects: `tool_executed`, `result_released`.

Required observations: decision/effects above.

### command-audit-outcome-fails — G07

Requirements: FR-17.

Setup: audit_outcome=fails.

Action: run command.

Expected decision: `block`. Ordered effects: `intent_durable` → `isolated_command` → `tool_container_absent`.

Additional forbidden effects: `result_released`.

Required observations: decision/effects above.

### command-path-escape — G07

Requirements: FR-13.

Setup: cwd=symlink-outside-workspace.

Action: run command.

Expected decision: `deny`. Ordered effects: none.

Additional forbidden effects: `tool_executed`, `host_execution`.

Required observations: decision/effects above.

### command-external-unapproved — G07

Requirements: FR-14.

Setup: action=publish; approval=absent.

Action: perform external action.

Expected decision: `block`. Ordered effects: `approval_requested`.

Additional forbidden effects: `external_effect`.

Required observations: decision/effects above.

### command-stale-approval — G07

Requirements: FR-14.

Setup: approval_revision=r1; current_revision=r2.

Action: perform external action.

Expected decision: `block`. Ordered effects: `approval_invalidated`.

Additional forbidden effects: `external_effect`.

Required observations: decision/effects above.

### artifact-observed — G07

Requirements: FR-09, FR-17.

Setup: check=observed-pass; revision=r1; authorized_human=true.

Action: export artifact.

Expected decision: `allow`. Ordered effects: `artifact_validated` → `artifact_audited` → `artifact_delivered`.

Additional forbidden effects: none beyond global prohibitions.

Required observations: `observed=true`; `revision=r1`.

### artifact-model-claim — G07

Requirements: FR-09.

Setup: check=model-says-pass; execution_receipt=absent.

Action: present check result.

Expected decision: `allow`. Ordered effects: `claim_labelled`.

Additional forbidden effects: `observed_pass_recorded`.

Required observations: `observed=false`.

### artifact-revision-mismatch — G07

Requirements: FR-09, FR-14.

Setup: artifact_revision=r1; current_revision=r2.

Action: accept artifact evidence.

Expected decision: `deny`. Ordered effects: none.

Additional forbidden effects: `evidence_accepted`.

Required observations: decision/effects above.

### artifact-too-large — G07

Requirements: FR-32.

Setup: artifact_bytes=over-export-bound.

Action: export artifact.

Expected decision: `deny`. Ordered effects: none.

Additional forbidden effects: `artifact_delivered`.

Required observations: decision/effects above.

### artifact-cross-room — G07

Requirements: FR-13, FR-20.

Setup: requester=foreign-room.

Action: export artifact.

Expected decision: `deny`. Ordered effects: none.

Additional forbidden effects: `artifact_delivered`.

Required observations: decision/effects above.

### artifact-escape — G07

Requirements: FR-13.

Setup: path=../outside.

Action: export artifact.

Expected decision: `deny`. Ordered effects: none.

Additional forbidden effects: `artifact_delivered`.

Required observations: decision/effects above.

### artifact-credential-isolation — G07

Requirements: FR-13, FR-17.

Setup: engine_credential=synthetic-engine-only-canary; artifact_source=tool-workspace.

Action: export isolated workspace artifact.

Expected decision: `allow`. Ordered effects: `artifact_delivered`.

Additional forbidden effects: `secret_delivered`, `credential_in_audit`.

Required observations: `credential_canary_present=false`.

### artifact-storage-fails — G07

Requirements: FR-32.

Setup: quota=exhausted.

Action: store artifact.

Expected decision: `deny`. Ordered effects: none.

Additional forbidden effects: `artifact_delivered`, `unmetered_artifact`.

Required observations: decision/effects above.

### discard-approved — G08

Requirements: FR-14, FR-33.

Setup: approval=explicit-current-owner; workspace=exact-job-volume-nonce; roles=absent.

Action: discard retained workspace.

Expected decision: `allow`. Ordered effects: `discard_approval_checked` → `discard_intent_durable` → `workspace_removed` → `absence_proven` → `storage_reservation_released`.

Additional forbidden effects: `history_deleted`.

Required observations: decision/effects above.

### discard-missing — G08

Requirements: FR-14, FR-33.

Setup: approval=absent; workspace=retained.

Action: discard workspace.

Expected decision: `deny`. Ordered effects: none.

Additional forbidden effects: `workspace_removed`, `storage_reservation_released`.

Required observations: decision/effects above.

### discard-stale — G08

Requirements: FR-14, FR-33.

Setup: approval=stale-revision; workspace=retained.

Action: discard workspace.

Expected decision: `deny`. Ordered effects: none.

Additional forbidden effects: `workspace_removed`, `storage_reservation_released`.

Required observations: decision/effects above.

### discard-wrong-actor — G08

Requirements: FR-14, FR-33.

Setup: approval=unauthorized-human; workspace=retained.

Action: discard workspace.

Expected decision: `deny`. Ordered effects: none.

Additional forbidden effects: `workspace_removed`, `storage_reservation_released`.

Required observations: decision/effects above.

### discard-wrong-volume — G08

Requirements: FR-14, FR-33.

Setup: approval=different-volume-nonce; workspace=retained.

Action: discard workspace.

Expected decision: `deny`. Ordered effects: none.

Additional forbidden effects: `workspace_removed`, `storage_reservation_released`.

Required observations: decision/effects above.

### discard-remove-fails — G08

Requirements: FR-33.

Setup: approval=explicit-current-owner; remove=fails.

Action: discard workspace.

Expected decision: `block`. Ordered effects: `cleanup_unconfirmed`.

Additional forbidden effects: `storage_reservation_released`.

Required observations: decision/effects above.

### discard-absence-unknown — G08

Requirements: FR-33.

Setup: approval=explicit-current-owner; remove=reply-lost; inventory=unavailable.

Action: discard workspace.

Expected decision: `block`. Ordered effects: `cleanup_unconfirmed`.

Additional forbidden effects: `storage_reservation_released`.

Required observations: decision/effects above.

### cancel-does-not-discard — G08

Requirements: FR-16, FR-33.

Setup: job=running; approval=cancel-only.

Action: cancel work.

Expected decision: `allow`. Ordered effects: `fenced_pair` → `workspace_retained`.

Additional forbidden effects: `workspace_removed`.

Required observations: decision/effects above.

### discard-history-protected — G08

Requirements: FR-18, FR-20.

Setup: actor=owner; flags=off; workspace_discard=approved.

Action: purge audit history.

Expected decision: `deny`. Ordered effects: none.

Additional forbidden effects: `history_deleted`, `history_mutated`.

Required observations: decision/effects above.
