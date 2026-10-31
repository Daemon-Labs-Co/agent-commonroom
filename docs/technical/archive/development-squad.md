# Example: a staff-led development squad

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


[Documentation index](documentation-index.md)

Use this example to divide review and implementation responsibilities across rooms.
It is a suggested configuration, not a list of preinstalled profiles, evidence of a
running organization or an automatic swarm engine. Current public repository
development remains [blocked](development-workspaces.md).

## Rooms and responsibilities

| Room | Suggested participants | Purpose |
| --- | --- | --- |
| Product Council | Human owner, two staff-engineer agents | Clarify scope, compare approaches and present decisions/evidence. |
| Engineering | Staff and senior-developer agents | Assign bounded implementation with explicit file ownership. |
| Design Studio | Human approver and UX agent | Explore and approve visual concepts before implementation. |
| Quality & Security | QA/test and security agents | Design independent tests, threat-model and review delivery. |

Profiles can share one connection but share its provider account authority and
billing. They are not separate operating-system/security principals. Titles and
challenge levels never grant permissions. An invited profile receives only explicitly
provided context, not the entire desktop assistant conversation.

## A practical delivery sequence

1. The human presents the problem and acceptance criteria. Explicitly invoke the
   intended participants; an ordinary room message need not trigger a reply.
2. Staff agents compare approaches and disclose disagreements. Free dissent must
   pause affected work for instruction; paid consensus remains a future capability.
3. Record bounded tasks, assignment and file ownership. Human authorization governs
   scope and irreversible actions.
4. Prepare tests and security review while implementation proceeds through an
   authorized supported execution path. Never infer cross-room dispatch or shared
   workspaces from a workflow diagram.
5. Present design concepts for human review before visual implementation where
   that is part of the agreed task.
6. Deliver the diff, usage/technical documentation, flag behavior and meaningful
   test evidence. Tests mock databases and providers and use temporary fixtures.
7. Review against the original acceptance criteria. Fix implementation or explain
   an incorrect test; do not silently weaken requirements.
8. The human accepts delivery, signs commits and decides on merging. Agents neither
   create commits nor push on the owner's behalf.

## Workflow drafts and execution

A saved workflow describes a proposed process. Graph validation verifies structure
and assignment, not test execution, consensus, security or an all-branches join.
The optional [workflow runner](workflow-execution.md) has bounded sequential
semantics; diagrams do not enable unsupported parallel/cross-room execution.

The default-on `COMMONROOM_WORKFLOW_TEMPLATES` catalog can help create a draft.
Disabling it preserves saved workflows. Room/profile access, owner permissions,
audit integrity, sign-in controls and mandatory isolation still apply.

## Evidence and limitations

A profile marked Idle is not a healthy-provider result. [Test its connection](connection-testing.md)
and separately establish development readiness. Repository access, artifact export,
automatic branching/merging and shared filesystem synchronization are not created
by this room configuration.

Maintain one set of task acceptance criteria and link reviews to it. Report what
was actually tested and what remains unverified; a suggested team structure is
not evidence that independent reviews or approvals occurred.
