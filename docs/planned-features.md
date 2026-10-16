# Backlog and future work

This page records requested work that is not delivered. The
[functional specification](../FUNCTIONAL_SPEC.md) defines scope; the
[release-readiness page](release-readiness.md) lists current P0 blockers. Complete
P0 before P1. A backlog entry does not authorize execution, publishing or data deletion.

## P1 collaboration

| Workstream | Expected behavior |
| --- | --- |
| Governing agent | Persist the first appointment and human-selected replacements without granting human permissions. |
| Bounded delegation | Turn approved proposals into scoped tasks with causal message references, dependencies and duplicate prevention. |
| Disagreement | Record objections, block affected/dependent work and wait for an authorized human. Free agents do not vote or resolve disputes by rank. |
| Decision attribution | Link instructions, proposals, approvals, revisions, attempts and observed outcomes. |
| Automatic recovery | Reconcile durable state before continuing; never blindly retry an uncertain external effect. Existing manual recovery remains available. |
| Parallel workflow joins | Run independent branches and join only when required branches succeed, with clear cancellation and failure behavior. |
| Cross-room handoff | Transfer approved context/evidence while preserving room access, repository and approval boundaries. |

Artifact handoff and observed execution evidence are P0 requirements for the
current release, even though older workflow backlogs placed them later. Their
current acceptance contracts are in the [P0 test plan](p0-acceptance-tests.md).

## Secret exposure protection and managed secrets

[FR-25–28](../FUNCTIONAL_SPEC.md#secret-exposure-alerts-and-managed-secrets--requested-not-implemented)
are requested follow-up work after the current P0 execution/security priorities.
They are not implemented.

- Combine agent guidance with server checks at supported input, persistence,
  provider-export and output boundaries. Alerts must not repeat a suspected secret.
- Ask the human to revoke/rotate and save replacements in a protected key/value
  editor, rather than chat. Do not silently rotate credentials.
- Keep secret values outside ordinary events and model context. Audit only the
  authenticated actor, generic operation and standard timestamp—no key names,
  values, changed fields, exports or value-derived hashes.
- Provide explicit permission-checked reveal/edit/delete/export and concurrency
  handling. Define encryption/custody before claiming protected storage complete.
- Export safe dotenv-compatible data atomically to an authorized target, with
  private permissions and path/symlink checks. Never execute the file or bundle
  it into Git, artifacts, memory providers or model context.

The editor/export needs a documented owner-controlled server flag; `managedSecrets`
was proposed with default off, but is not an accepted runtime flag. Exposure
protection must remain independent of optional editor controls and paid access.

## Other product requests

| Request | Boundary before implementation |
| --- | --- |
| Tasks extracted from discussion | Explicit configuration, source attribution, deduplication and human correction; extraction alone never dispatches work. |
| Human-visible branch controls | Repository/branch policy and isolated access; creating a branch grants no commit, push, merge or signing authority. Internal bootstrap already uses a fresh feature branch. |
| Richer process monitoring | Authenticated job-scoped observations with bounded safe summaries; activity never extends execution limits or proves useful progress. |
| History search/export | Bounded permission-checked server search and export. Current search covers loaded pages, not the entire database. |
| Human-attention preferences | Define ownership and notification/escalation behavior. Acknowledgment remains distinct from approval. |
| Anonymous adoption reporting | Agree consent/defaults, collector, minimization and installation-versus-person counting before any transmission. |
| Action-button spacing | Improve responsive grouping and keyboard focus; verify the current layout before treating a reported defect as still present. |
| Optional Jev decision service | Define a narrow schema-validated adapter and disclosure controls; unavailable/uncertain results defer to human review. No integration exists. |

Any external decision service requires explicit configuration and scoped data
sharing. It cannot override permissions, isolation, audit integrity or human
acceptance. Repository development skills do not imply a product plugin runtime.

## Coming soon and separate paid work

MCP/skills product execution and a plugin interface remain future extensibility
work. Registered placeholders must stay labelled **coming soon**.

Paid consensus, multi-human authority/escalation, Fulginn memory enhancements,
Veriscribe attribution and enterprise compliance belong to separate commercial
work. Existing organization-chart records can inform future hierarchy features;
the chart is already implemented and is not a pending feature branch. No chart,
persona, payment or plugin may bypass core safeguards.

## Completion standard

Each implementation needs meaningful fast mocked tests, current user/technical
documentation and applicable owner-controlled server-enforced feature flags with
both states tested. Mandatory security boundaries have no off switch. Use the
[contribution rules](../CONTRIBUTING.md); the owner signs commits manually.
