# Agent team loop and extension layer — proposed design

Proposal for review. Nothing on this page is implemented unless the
[product inventory](product-functionality.md) says so. It does not change any flag,
default or authority.

## Input

Goal: two or more agents, on different providers and models, each with its own
persona, work together in one room towards a defined goal. They respond to each
other, object when they disagree, escalate what they cannot settle to the person or
agent they report to, and turn agreed work into tasks that are executed and reviewed.
Separately deployed services must be able to add capabilities without their code
living in, or being imported by, this repository.

Current behaviour this builds on:

| Area | Today |
| --- | --- |
| Communication | Agents share the room: each turn receives recent room messages, including other agents' replies, and is told that peers must reach consensus, that managers outrank their reports and that unresolved disagreements must be reported. |
| Turns | A turn starts only from a human `send` or `broadcast`, a workflow step or a task dispatch. An agent's reply does not start another agent's turn, so an exchange between agents advances one human prompt at a time. |
| Persona | Name, role, discipline, level, challenge level, instructions and model per agent. |
| Goal | Room outcome text and review (`roomOutcomes`, `roomOutcomeReview`). |
| Reporting | Organization chart with `reportsTo`; it grants no access. |
| Governing agent | Appointment record per room (`agentCoordination`); no behaviour of its own. |
| Disagreement | Challenge level (supportive to red team) and the consensus guidance shape what agents say. A disagreement is visible as message text; it is not recorded as an objection and blocks nothing. |
| Escalation | `[HUMAN ACTION REQUIRED]` first-line protocol raises a human-attention item. |
| Tasks | Owner captures a proposal from a message, approves it, then dispatches by hand. |
| Containment | Mandatory container isolation, controlled egress, turn budget, permanent log. |

Constraints: every existing security contract stays. Agent text is untrusted
reference, never an approval or an execution grant. Humans approve; agents propose
and do. No extension may weaken authentication, authorization, isolation, audit or
retention.

## Output

Three pieces, each usable without the next.

1. **Team loop (core).** Bounded discussion rounds between agents, a stance
   protocol, recorded objections, an escalation path and proposals that flow into
   the existing task contract, dispatch and review.
2. **Governing agent (core).** The appointed agent gains suggest-only behaviour:
   it drafts task proposals, suggests assignee and reviewer, and summarises or
   flags stalled work. It runs as an ordinary managed turn on any connection.
3. **Extension layer (core seam).** A registry, a scoped service identity, a
   versioned contract API, an event feed and sandboxed interface slots, so a
   separate service can do what the governing agent does, or add other capabilities,
   through the same checks as everyone else.

Refusals stay explicit: a round that hits a limit stops with a visible room notice;
an objection blocks only the work it names; an extension call outside its scope
returns 403 and is logged.

## Functionality

### Discussion rounds

Agents already read and answer each other; what is missing is continuing without a
human prompt for every reply. After a human message, a round lets each selected
agent reply in turn, seeing the earlier replies of that round. Rounds repeat until one stop condition is met:
every agent's stance is `AGREE`; any agent escalates; the round limit or the
per-discussion turn limit is reached; a human posts or stops the
room. Each turn is a normal managed turn: isolated worker, turn budget, context
sharing policy and context boundary all apply. One discussion runs per room at a
time. A human can stop it at any point with the existing stop control.

The round ceiling is 3. The owner may lower it per room (minimum 1) but cannot
raise it. A registered extension may raise the ceiling for an installation through
the contract, to at most 10, which is a fixed safety maximum in the core.

A discussion starts when a human asks for one, or when an agent starts its reply
with `[DISCUSS] topic`. An agent-started discussion counts against the same round,
turn and daily budgets, and an agent cannot start another until a human has posted
in the room, so two agents cannot keep each other running.

### Stance protocol

An agent's reply may start with one first-line marker, parsed by the server the
same way as `[HUMAN ACTION REQUIRED]`:

| Marker | Effect |
| --- | --- |
| `[AGREE]` | Counts towards ending the discussion. |
| `[OBJECT] reason` | Records an objection against the proposal or task it answers. |
| `[ESCALATE] question` | Ends the discussion and raises the question to the escalation target. |
| `[PROPOSE TASK] title` | Records a task proposal with this message as its source. |
| `[DISCUSS] topic` | Starts a discussion with the room's agents. |
| `[ASSIGN name] title` | A manager assigns a task to one of its reports (see below). |

A marker is a request, not a decision. A reply without a marker is plain
conversation. Markers inside quoted or reference text are ignored.

### Objections and escalation

An objection is a stored record linked to its source message and target. While it
is open, the targeted proposal cannot be approved and the targeted task cannot be
dispatched; unrelated work continues. It is closed by the objecting agent
withdrawing it or by a decision from the escalation target.

The escalation target is the actor the agent reports to in the organization chart.
If that is a human, the question becomes a human-attention item. If that is the
room's governing agent, it may answer once; its answer can close an objection only
when the owner has allowed that for the room (default off), and it can never
approve, dispatch or accept work. Anything it does not settle goes to the human.
Seniority and challenge level never settle a disagreement by themselves.

### Manager-assigned tasks

Authority follows the company reporting hierarchy, not numbered seniority, as the
existing guidance already tells agents. When the owner turns this on for a room,
an agent may assign a task to an agent that reports to it, directly or
indirectly, with `[ASSIGN name] title` followed by the task text. The server checks
the reporting relationship itself; the marker is refused for a peer, a manager or
an agent outside the room.

An assigned task is created and dispatched without a human approval, within these
bounds: it runs with the report's existing permissions and containment and grants
nothing new; it counts against the daily turn budget and the daily limit on manager-assigned
tasks (default 5 per room); it cannot approve, accept, publish or push anything, which
stay human decisions; the report may answer with `[OBJECT]` or `[ESCALATE]`, which
pauses the task and raises it to the assigning manager's own manager, ending at a
human. Every assignment is shown in the room and recorded with the assigning
agent as its source. A human can cancel any assigned task.

Because an agent's text now starts work, this is the highest-risk step: a manager
agent that reads hostile content could be steered into issuing assignments. The
limits above bound the effect to contained, budgeted, reversible work.

### Proposals to tasks

`[PROPOSE TASK]` and governing-agent suggestions create entries in the existing
proposals list. A human approves, edits or rejects each one. Approval creates a
task with the existing contract. With `taskDispatch` on, the owner may mark a room
as "dispatch approved tasks": an approved task with an assignee and bound inputs
is then dispatched without a second click. Review uses the existing reviewer and
acceptance flow; reviewer and assignee must differ.

### Extension layer

An extension is a separately deployed service listed in a registry file with a
name, address and public key. It authenticates with a scoped credential and acts as
a system actor that is neither a human nor the owner.

| Capability | Scope |
| --- | --- |
| Read | Room events it is granted, filtered by the room's context sharing policy. |
| Write | Create proposals and post attributed notices. Nothing else in the first version. |
| Interface | A navigation entry or panel rendered in a sandboxed frame with a small message API. |

Every call passes the same validation, permission and audit path as a browser
request. The contract is versioned; core scripts never load extension code. Whether
an extension is licensed, metered or free is the extension's own concern and is
not visible to, or decided by, this repository. Removing an extension leaves all
core behaviour, records and history intact.

### Build order

1. Stance protocol and objection records (no new turn behaviour).
2. Discussion rounds with limits and stop conditions.
3. Escalation target from the organization chart; agent-started discussions.
4. Governing-agent suggestions into the proposals list.
5. Optional dispatch of approved tasks; manager-assigned tasks.
6. Extension registry, service identity and proposal contract.
7. Sandboxed interface slot and event feed.

Each step ships with tests for both flag states, user and technical documentation,
and a live check in a test room before the next begins.

## Properties and state changes

Proposed flags, all with server enforcement and owner-only changes:

| Flag | Default | Disabled behaviour |
| --- | --- | --- |
| `agentDiscussion` | Off | Agents reply only to humans, as today. Recorded discussions stay in history. |
| `agentObjections` | Off | Markers are stored as text only; nothing is blocked. Existing objections stay visible and can be closed. |
| `governingSuggestions` | Off | The governing agent takes no turns of its own. |
| `autoDispatchApproved` | Off | Approved tasks wait for a manual dispatch. |
| `managerAssignedTasks` | Off | `[ASSIGN]` is stored as text only; assignments already made stay visible and cancellable. |

The daily limit on manager-assigned tasks is an owner setting, default 5 per room,
range 0–50, changeable at any time without a restart. It applies only while
`managerAssignedTasks` is on; 0 refuses every assignment. A change takes effect for
the next assignment and is recorded with the owner as its source; assignments
already made that day still count towards the new limit.
| `extensions` | Off | The registry is ignored; extension credentials are refused. |

New persistent state: discussion records (room, round, turn, stop reason),
objection records (source message, target, status, closing decision), the per-room
"governing agent may close objections" and "dispatch approved tasks" settings, and
extension registrations. All are append-only events; nothing existing is rewritten.

Limits: rounds per discussion, turns per discussion, one discussion per room, the
existing daily managed-turn budget, and a per-extension request rate. Reaching a
limit stops the activity with a room notice and never extends another limit.

Decided by the owner: round ceiling 3, lowerable only; discussions may be started
by an agent; a manager in the reporting hierarchy may assign tasks to its reports
and is the escalation target.

Not covered: file and bundle deliverables (blocked on verified storage), the
approved branch push, a built-in local model, and multi-human teams. The loop
is useful without them, but an application cannot be built end to end until the
first two exist.
