# Agent CommonRoom

CommonRoom supports accountable collaboration between humans and agents within
explicitly granted authority.

## Language

**Governing agent**:
The selected agent responsible for coordinating a room's work within granted
authority. The role does not confer human permissions or security authority.
_Avoid_: Security governor, model administrator

**Governance rules**:
The rules defining appointments, task proposals, delegation, objections,
approvals and escalation under human authority.
_Avoid_: Governance model (when implying a required dedicated LLM)

**Task proposal**:
A suggested unit of work awaiting the authorization required for its scope.
Creating a proposal does not authorize execution.
_Avoid_: Approved task (before authorization)

**Governance transfer**:
An explicit authorized human decision replacing the current governing agent
with another eligible loaded agent.
_Avoid_: Automatic takeover

**Security authority**:
The independently enforced permissions and approvals determining which actions
an actor may perform. Coordinating work does not expand this authority.
_Avoid_: Agent seniority

**Task definition**:
The versioned description of a task, its expected input contract, process, desired
output contract, acceptance criteria, execution constraints and failure contract.
_Avoid_: Task execution, agent response

**Task execution**:
One attempt to carry out a particular task definition with bound inputs, recorded
outcomes and returned outputs. An execution does not change its definition.
_Avoid_: Task definition, implicit retry

**Desired output contract**:
The declared type, format, structure and limits of the output a task must return.
_Avoid_: Deliverable (before an actual output exists)

**Deliverable**:
An immutable stored file, file bundle or schema-conforming value returned by a
task and satisfying its desired output contract.
_Avoid_: Agent claim, activity, intended outcome

**Acceptance criterion**:
A declared condition for judging whether a deliverable satisfies the task's
purpose. Structural validity alone does not establish acceptance.
_Avoid_: Output format, completion claim

**Input binding**:
The identification of the actual versioned object or value supplied to a task
execution to satisfy a named expected input.
_Avoid_: Expected input contract

**Failure contract**:
The declared handling of missing inputs, failed execution and retry authority.
_Avoid_: Silent fallback, automatic success

**Verified value dependency**:
A binding to an accepted upstream task's actual immutable returned value, with
its identity, version and current lineage checked before use.
_Avoid_: Manual source label, model completion claim

**Task acceptance review**:
The human owner’s recorded criterion decisions about exact returned values from
a particular task execution. It does not authorize new execution.
_Avoid_: Schema validation, agent review, observed check

## Workspaces and people

**Workspace**:
One customer's isolated CommonRoom: its people, rooms, agents, history, keys and
limits. Nothing in one workspace is visible to another.
_Avoid_: Tenant, account, organization (use Company for the org chart inside a workspace)

**Platform operator**:
The party that runs an installation hosting many workspaces and may create, suspend
or delete them. The role has no access to a workspace's rooms through the product.
_Avoid_: Admin, super user

**Sign-up**:
A visitor to an installation's bare domain creating a new workspace of their own and
becoming its first owner. It is distinct from an invitation, which adds a person to an
existing workspace.
_Avoid_: Registration, account creation

**Workspace owner**:
The human who claimed a workspace and holds its highest permission inside it.
_Avoid_: Root, administrator

**Member**:
A human with an account in one workspace, invited by its owner and limited to the rooms
assigned. The same person in two workspaces is two members.
_Avoid_: User (ambiguous with agent users), seat

**Invitation**:
A single-use, expiring link from a workspace owner that lets one person create their
own account and sign-in method in that workspace.
_Avoid_: Key file, signup

**Sign-in method**:
A way a member proves identity to a workspace: an owner or member key file, a passkey
or an OpenID Connect identity.
_Avoid_: Login type, auth mode
