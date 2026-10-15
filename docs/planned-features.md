# Separate feature work

The org chart is implemented on its own feature branch; the other entries remain requested features, not shipped capabilities. Implement each on its own feature branch with tests, user/technical documentation and an owner-controlled server-enforced flag. Do not mix them into the current UI refinement branch or commit/push on behalf of the owner.

## Secret exposure alerts and managed secrets

Owner-requested 15 October 2026; **not implemented**. Authoritative requirements:
[FR-25–FR-28](../FUNCTIONAL_SPEC.md#secret-exposure-alerts-and-managed-secrets--requested-not-implemented).
Existing P0 execution/security gates remain first; this entry does not enable a
new credential path or expand the current runtime's authority.

- All agents watch for credential exposure and notify the human without echoing
  values. Add deterministic server-side checks before supported persistence and
  provider-export boundaries; do not rely on model compliance alone.
- Distinguish suspected exposure from confirmed exposure. Prompt rotation/revocation
  and save replacements through an authorized key/value secrets editor, never chat.
- Keep usernames, passwords, tokens, API keys and return URLs in a dedicated store
  outside ordinary events/model context. Owner reveal/edit/delete/export actions
  need explicit permissions, actor/action-only auditing and concurrency protection.
  Example: "Mark edited the secrets" plus the standard timestamp; never log key
  names, values, changed fields, exported contents or hashes of values.
- Export safely encoded values directly to an authorized user-controlled secrets
  file or explicit download, with safe targets, private permissions and atomic
  replacement. Never execute the exported file or include it in Git/artifacts.
- Before implementation, finalize protected storage/encryption, dotenv encoding,
  key/reference exposure and scoped runtime consumption. Preserve the mandatory
  provider/tool credential separation contract.
- Proposed editor/export flag: `managedSecrets`, default **off**, owner-only,
  server-enforced; disabling preserves stored secrets. Detection remains a
  separate protective requirement. This is a proposed contract, not a currently
  accepted server flag. Tests use synthetic credentials/temp files only.

## Worker process monitoring and callbacks

Implemented slices: bounded active-turn admission, elapsed dispatch age warnings,
timeout cause preservation, internal adapter callbacks and approved SSH worker
status polling with an expandable background timeline. Durable reconciliation,
exhaustive provider activity and development-specific execution budgets below
remain unimplemented. See [Background work](concepts.md#background-work).

- Add authenticated, job-scoped lifecycle events: starting, running, progress,
  completed, failed, and cancellation confirmed. Reject duplicate, out-of-order,
  foreign-job and post-completion updates.
- Show last confirmed activity and elapsed time in squad status and job details.
  Distinguish worker liveness from useful model progress; silence is not proof of
  a hung process.
- Bound and sanitize provider progress. Never publish raw credentials, full
  command output or private paths as status text. Preserve room permissions,
  audit integrity, mandatory isolation and human-controlled recovery.
- Keep a hard maximum execution budget. Design development-specific budgets and
  stalled-job handling instead of silently extending a turn on every heartbeat.
- Ship on a separate feature branch with an owner-controlled server-enforced
  flag, mocked transport/database tests and user/technical documentation.

## Jev decision-service integration

Requested provider: [TypeSafe AI Jev](https://docs.typesafe.ai/introduction).
Use an optional external decision adapter for narrowly scoped task classification,
workflow routing or assessment rubrics. Typed Choice/Score answers and confidence
must be schema-validated and evaluated against synthetic labelled examples before
enabling automatic routing. Low confidence, unavailable service or invalid output
must defer to existing human/agent review. Confidence is not proof of correctness.

Deliver on its own feature branch behind a default-off, owner-controlled,
server-enforced flag. Keep API credentials server-side; explicitly configure what
room context may leave the application, with bounded requests, timeouts and rate
limits. Never give this service authority over permissions, isolation, audit
integrity, final human acceptance or security-policy overrides. No external Jev
calls or application integration are implemented yet.

## Visual company organization chart

Suggested branch: `feature/company-org-chart`.

- Under Settings → Company, provide a workflow-style canvas with all human and agent identities, connectors and a selected-node properties panel.
- Distinguish a registered agent from its room-specific profiles. Do not invent reporting relationships from decision seniority.
- Company reporting relationships replace numbered authority; managers outrank reports, peers need consensus and humans retain final override. See [the implemented contract](concepts.md#company-organization-and-decision-authority).
- Include keyboard navigation and a readable list alternative to the canvas.

## Automatic tasks from discussions

Suggested branch: `feature/discussion-task-creation`.

- Make task creation from room discussion an explicit, configurable feature. The user requested it enabled when delivered.
- Preserve a source-message reference for each task; deduplicate repeated delivery/retries, expose creation in history and support review/correction.
- Define eligible messages, human/agent permissions, task extraction and protection against malicious messages creating privileged work. Task extraction must not implicitly run agents or approve work.

## Development branch creation

Suggested branch: `feature/development-branch-creation`.

- Offer human-managed branch creation or automatic creation as development proceeds, independently of discussion-task creation.
- Require an explicit repository and branch policy. Preserve unrelated edits and existing branches; never develop directly on main.
- A branch creation feature does not authorize commits, signing, pushes or merges. The owner continues signing commits manually.
- Define isolated-worker repository access and branch creation permissions before enabling automation. Test Git interactions with temporary repositories only.

## Anonymous feature adoption reporting

Suggested branch: `feature/anonymous-feature-statistics`. Report aggregate feature-flag adoption externally, with documented consent, an owner-controlled flag, data minimization and installation-versus-person counting. Collector configuration and opt-in/default behavior still need agreement. Not implemented.

## Consistent action-button spacing

Requested 12 October 2026. Connection-row Edit, Delete, Sign out and Test connection
buttons visually touch. Apply consistent horizontal/vertical gaps and wrapping to
all action groups; preserve destructive-action distinction and keyboard focus.
Deliver separately with responsive UI checks, documentation and the required flag.
This backlog entry does not change current button layout.

## Tasks arising from workflow execution

These are documented follow-up tasks, not automatically created CommonRoom tasks.
Source: [the shipped sequential runner and its limits](workflow-execution.md).

| Task | Acceptance target | Status |
| --- | --- | --- |
| WF-01: Reviewable artifact handoff | Export a bounded patch and artifacts from the isolated workspace, identify the exact source revision, and make them available to an independent reviewer without broadening repository access. | Backlog |
| WF-02: Verified execution evidence | Capture actual command/test outcomes and bind them to the reviewed revision; distinguish collected evidence from model assertions and human notes. | Backlog |
| WF-03: Parallel workflow joins | Execute independent branches and join them only when required branches succeed; enforce cancellation, timeout and failed-branch behavior. | Backlog |
| WF-04: Cross-room handoff | Transfer task context and approved evidence to a destination room while preserving membership, repository and approval boundaries. | Backlog |
| WF-05: Durable controller recovery | Reconcile orphaned workers and use durable leases without silently repeating mutating steps or accepting stale results. | Backlog |

Each implementation needs its own scoped design, meaningful mock-only tests,
user/technical documentation and an enforced flag where applicable. A backlog
entry does not authorize automatic assignment, execution, commit, push or merge.

## Follow-up tasks from human attention requests

- HA-01: Evaluate detection precision/recall on synthetic examples and improve the
  adapter response contract; do not treat quoted instructions as authoritative.
- HA-02: Define optional per-human ownership and notification preferences before
  adding reminders, escalation or desktop notifications. Current acknowledgment
  is room-wide and never approval.

Source: [human attention behavior and limitations](concepts.md#messages-needing-human-attention).
These are repository backlog entries; automatic CommonRoom task creation is not
implemented by this feature.

## Follow-up tasks from History refinement

- HI-01: Define consistent event performer/subject/scope attribution for future
  records. Do not infer an editor from an affected participant identity or rewrite
  old events to invent missing attribution.
- HI-02: Design bounded server-side search and export with permission checks,
  privacy controls and retention safeguards. Current search covers loaded pages.

Source: [Organized History](concepts.md#organized-history). These remain repository
backlog entries, not automatically created application tasks.
