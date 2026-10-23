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
