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
