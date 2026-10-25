# Documentation

Choose a guide for the job you want to do. CommonRoom is a prototype: the
[release-readiness page](release-readiness.md) separates available behavior,
internal tested components and remaining release gates. The
[functional specification](../FUNCTIONAL_SPEC.md) is the requirements baseline.

## Get started

| Goal | Guide |
| --- | --- |
| Install the controller locally | [Installation](installation.md) |
| Deploy on a Debian server | [Debian deployment](deployment/debian.md) |
| Use nginx and an existing database privately | [nginx and PostgreSQL](deployment/nginx-postgresql.md) |
| Prepare the required database roles | [Database security](database-security.md) |
| Connect an approved worker | [SSH setup](agent-ssh.md), [provider sign-in](connection-sign-in.md), [connection testing](connection-testing.md) |

## Use the workspace

[Task definitions](task-definitions.md) establishes finite input/process/output contracts and separate immutable execution records.
- [Accepted task inputs](task-value-inputs.md): verified same-room value dependencies, provenance and stale-source refusal.

[Bounded delegation](task-delegations.md) documents authenticated assignment
proposals, human review, inherited limits and provider integration gaps.

[Conversation task proposals](task-proposals.md) separates suggestion capture,
human review and draft creation from execution.

[Room coordination appointments](governance-appointments.md) explains the optional
owner controls, revision-bound coordinator requests, and the distinction between
selection and execution authority.

| Goal | Guide |
| --- | --- |
| Understand rooms, connections, profiles and memory | [Product guide](concepts.md) |
| Assign work and request review | [Task dispatch](task-dispatch.md) |
| Start from a process template | [Workflow templates](workflow-templates.md) |
| Run a supported sequential workflow | [Workflow execution](workflow-execution.md) |
| Interpret status, stop or recover an agent | [Agent activity and recovery](agent-activity.md) |
| Review permanent deletion of a stopped workspace | [Owner-reviewed disposal](workspace-disposal.md) |
| Find decisions and understand retention | [History](history.md) |
| Edit company reporting relationships | [Organization chart](organization.md) |
| Adjust profiles and display preferences | [Interface preferences](interface.md), [Help display](help-display.md) |
| Control provider context sharing | [Context disclosure](context-disclosure.md) |

## Operate and troubleshoot

- [Worker idle shutdown guard](worker-idle-shutdown.md): internal paired-worker policy and remaining idle-agent gaps.
- [Operations and backups](operations.md): start, stop, recover and protect an installation.
- [Worker jobs](worker-jobs.md): durable job identity, diagnostics and reconciliation.
- [Container isolation](container-isolation.md): supported worker platform and mandatory limits.
- [Storage accounting](storage-accounting.md): durable capacity reservations and remaining physical quota requirements.
- [Execution budgets](execution-budgets.md) and [container quotas](container-quotas.md): enforced admission ceilings and storage gaps.
- [Login availability](login-availability.md), [private file safety](private-file-security.md) and [storage deadlines](storage-deadlines.md): bounded failures and recovery.
- [Development workspaces](development-workspaces.md) and [development squads](development-squad.md): configuration boundaries while development admission remains gated.

- [P0 release-gate acceptance specification](p0-release-gate-acceptance.md): procedures, evidence and pass/fail rules for all nine gates.

## Understand the implementation

- [Provider-independent governance](governance-model.md): accepted coordination and security boundaries, transfer rules and pending acceptance.

Start with [architecture](architecture.md), [server structure](server-structure.md)
and the [HTTP API](api.md). The execution references describe internal contracts;
they do not imply that an end-user development endpoint is enabled.

| Layer | References |
| --- | --- |
| Credential boundary | [Provider/tool separation](provider-tool-separation.md) |
| Controller intent and recovery | [Managed jobs](managed-job-lifecycle.md), [atomic operations](atomic-operations.md) |
| Workspace disposal safety | [Exact volume binding](workspace-disposal-binding.md) |
| Pair startup and cleanup | [Paired lifecycle](worker-pairs.md), [assembly](worker-pair-assembly.md), [shutdown](worker-pair-shutdown.md) |
| Workspace | [Volumes](workspace-volumes.md), [bootstrap](workspace-bootstrap.md), [tool runtime](tool-container-runtime.md) |
| Bounded commands | [Isolated Go commands](isolated-commands.md) |
| Audited actions | [Tool audit protocol](tool-audit-callbacks.md), [private transport](worker-audit-transport.md), [shared session](audited-tool-session.md) |
| Provider engine | [Engine runtime](engine-container-runtime.md), [bootstrap/image](codex-engine-bootstrap.md), [Codex callback driver](codex-callback-driver.md) |

## Build and evaluate

- [Development guide](development.md) and [contribution rules](../CONTRIBUTING.md).
- [Initial tool-only live evidence](live-tool-acceptance.md): limited manual worker verification, separate from release acceptance.
- [Release readiness](release-readiness.md) and [prepared acceptance scenarios](p0-acceptance-tests.md).
- [Backlog](planned-features.md): follow-up requirements, separate from delivered behavior.
- [Security policy](../SECURITY.md), [current risk review](../SECURITY_REVIEW.md) and [implemented hardening](security-hardening.md).

## Help inside CommonRoom

Open **Help & documentation** or contextual Help beside a form. The viewer is also
available at `/help/` before sign-in. It serves only allowlisted documents and
illustrations, never private state or arbitrary filesystem paths. Choose same-window
or separate-window help through your [Help display preference](help-display.md).

Keep user instructions in the relevant guide, requirements in the specification
and current blockers on the readiness page. Update links and examples when code
changes. Historical implementation narration belongs in version history, not
in a second set of setup instructions. Examples use synthetic data only.

## Setup choices and source navigation

- [Platform quick starts and setup scripts](quick-start.md)
- [Installation components: Go, PostgreSQL, compilation and web serving](installation-options.md)
- [Every repository file and Go API](reference/README.md)
- [Extending functionality](extending.md)

- [Check an installation and find setup failures](installation-check.md)

- [Recover uncertain worker credentials](worker-credential-recovery.md)

- [Private final-credential capture and fenced persistence](engine-credential-custody.md)

- [Reviewed harness release admission](harness-release-admission.md): signed release selection, deployment flag and remaining upgrade gaps.

- [Room outcomes and acceptance criteria](room-outcomes.md): shared-room deliverables, owner configuration and compatibility.

- [Human room deliverable acceptance](room-outcome-review.md): owner decisions, evidence and stale-snapshot safeguards.

- [Task criterion review](task-acceptance.md): explicit human decisions against exact task outputs and criteria.

- [Recorded task progress](task-progress.md): scoped contract, attempt, input and human-review status without claiming live verification.

- [Internal byte custody](object-custody.md): tested bounded storage foundation; production file/bundle delivery remains blocked.

The internal [finite artifact manifest foundation](artifact-manifests.md) binds
exact declared members to verified sealed bytes and registered structural formats.
It is disconnected from managed/public delivery; actual source, quota and durable
journal adapters remain required and file/bundle dispatch continues to refuse.

[Controller artifact custody journal](artifact-custody-journal.md) documents internal durable reservations, quota readiness, immutable receipts, read authorization and remaining production barriers.

[Bounded workspace source capture](workspace-source-capture.md) describes actual tree identity, lazy member streams and the mandatory production frozen-lease gap.
