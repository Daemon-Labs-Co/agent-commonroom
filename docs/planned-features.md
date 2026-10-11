# Separate feature work

These are requested features, not shipped capabilities. Implement each on its own feature branch with tests, user/technical documentation and an owner-controlled server-enforced flag. Do not mix them into the current UI refinement branch or commit/push on behalf of the owner.

## Visual company organization chart

Suggested branch: `feature/company-org-chart`.

- Under Settings → Company, provide a workflow-style canvas with all human and agent identities, connectors and a selected-node properties panel.
- Distinguish a registered agent from its room-specific profiles. Do not invent reporting relationships from decision seniority.
- Define how an owner edits reporting relationships and whether agent reporting is company-wide or profile-specific. Preserve room authorization when displaying assignments; do not reveal private room content.
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
