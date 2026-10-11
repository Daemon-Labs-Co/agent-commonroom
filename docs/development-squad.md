# Staff-led development squad

This is the operational setup requested for developing Agent CommonRoom inside
Agent CommonRoom. It is configuration of existing rooms, profiles and saved
workflow drafts, not a new automatic execution engine.

## Rooms and people

The owner is the sole human, with final scope, visual-design and delivery approval.
The owner is present in all four working rooms. Existing unrelated rooms and
historical participants remain intact.

| Room | Profiles | Responsibility |
| --- | --- | --- |
| Product Council | Owner, Codex (Staff Software Engineer), Claude (Staff Software Engineer) | Discuss ideas, agree scope, propose and assign tasks, coordinate delivery and present final evidence. |
| Engineering room | Owner, Codex and Claude (Staff), Maya and Daniel (Senior Software Engineers) | Staff engineers take hard work; senior developers handle bounded simpler tasks. |
| Design Studio | Owner, Elena (Staff UX Designer) | Critique functionality, develop visual concepts and obtain human approval before visual implementation. |
| Quality & Security | Owner, Priya (Senior QA Test Writer), Alex (Staff QA Engineer), Riley (Staff Red-team Specialist) | Concurrent test design and threat modelling, followed by independent verification of delivered code. |

These are separate room profiles and sessions using the already registered Codex
and Claude connections. They share their connection's provider account and billing.
The current company model identifies a bot by connection: these profiles are not
eight independent company identities or security principals. Titles do not create
access permissions. A new Codex profile does not transplant the entire desktop
conversation; only explicitly supplied starting context is included.

Profiles use constructive, critical, adversarial and red-team challenge levels
according to responsibility. Challenge levels never authorize hostile external
activity, credential access or bypassing the mandatory execution boundary.

## Delivery process

1. The human presents an idea in Product Council. Use Everyone when both staff
   agents should respond; an ordinary room message does not invoke them.
2. Codex, Claude and the human discuss value, functionality, feasibility, effort,
   risks and acceptance criteria. Unsupported assumptions should be challenged.
3. Save proposed work as unassigned tasks, then agree assignment. Codex and Claude
   split difficult work; Maya and Daniel receive simpler bounded tasks. Define
   explicit file ownership before concurrent edits. The human approves scope.
4. Start implementation, test authoring, visual concepts and concept security
   assessment concurrently through explicit room handoffs referencing the same
   task and acceptance criteria.
5. Elena presents concepts or mockups to the human. Record explicit visual
   approval before implementing those visual choices. CLI image generation is
   not assumed: request an appropriate capability when unavailable.
6. Priya writes meaningful tests during implementation; Riley threat-models the
   concept. Tests must mock every database interaction, be fast and parallel-safe,
   use temporary fixtures and avoid real provider agents.
7. Developers supply the actual diff, documentation, feature-flag behavior and
   test evidence. Alert the human after each completed bug fix with a commit
   message. Report total wall-clock runtime for the full test/check suites whenever
   tests are added or changed, not just individual test execution times.
8. Riley independently reviews the delivered code and fixes. Alex verifies
   acceptance criteria and test results. Failed tests are discussed with developers
   against the original task; change a demonstrably incorrect test or fix the
   implementation, never silently weaken an agreed requirement.
9. Staff engineers reconcile independent QA, security and UI evidence. Failed
   review returns work to the relevant developers and specialist room.
10. Present the human with delivered scope, effort/difficulty, changes, checks,
    total suite runtimes, screenshots, documentation, flag behavior, security
    findings, disagreements and remaining limitations. Only the human accepts
    delivery, signs commits and decides whether to merge.

All development must use feature/fix branches and mandatory isolated containers.
No agent creates commits or pushes. Do not claim tests, reviews or approvals exist
without their corresponding evidence.

## Saved workflow drafts

| Room | Saved workflow |
| --- | --- |
| Product Council | Staff-led idea to reviewed delivery |
| Engineering room | Staff and senior implementation |
| Design Studio | Human-approved UI concepts |
| Quality & Security | Concept and implementation security review |
| Quality & Security | Parallel test authoring and independent QA |

The council diagram contains a scope approval, four parallel tracks, a manual
all-branches evidence gate, staff review, a changes loop and final human approval.
Council agent nodes coordinate the named specialists in their respective rooms;
they do not execute cross-room dispatch themselves. Specialist workflow nodes
have room-local profile assignments.

Graph validation checks structure and assignment. It does not execute tests,
prove consensus, enforce a join barrier or certify security. Run remains disabled.
Cross-room task movement, automatic branch management, repository mounting and
artifact export are not established by this configuration. Current isolated
execution does not grant this squad access to the developer's host checkout.

The separate discussion-tasks branch is not deployed as part of this setup.
Until deployed and configured with an operational invited AI, create tasks manually.
The current running task form has no Backlog status: an unassigned manual task
uses its existing Assigned status. Do not mistake that label for an assignment.

## Controls, persistence and verification

This setup uses existing room/profile APIs and the workflow template catalog.
`COMMONROOM_WORKFLOW_TEMPLATES` defaults to true and controls template serving and
application; disabling it preserves existing saved workflow drafts. Existing
owner authorization, CSRF protection, audit recording, connection sign-in flags,
emergency controls and mandatory isolation remain in force. There is no new
product feature or isolation off-switch introduced by this configuration.

Rooms, profiles and workflow revisions were saved through the authenticated owner
interface; all five workflow drafts passed the UI's graph validation. This is
user-requested live configuration, not database test fixtures. No automated test
data was inserted. Separate owner-interface readiness checks passed for both
existing Codex and Claude connections: worker configuration, isolated runtime,
credentials and a fixed provider response. Those live checks send no room history
and are separate from automated tests. They do not prove repository access,
parallel development or complete delivery execution. Saved profiles marked Idle
alone do not prove their connections work.

## Follow-up room lifecycle requirement

First installation should create an Engineering room; continued use is optional.
The owner should be able to archive/remove ordinary rooms from active use,
including Engineering. Preserve immutable history and references. Permanent purge
must remain human-authorized and retention-controlled. A Product Council task
records this requirement; room deletion is not implemented by this setup.
