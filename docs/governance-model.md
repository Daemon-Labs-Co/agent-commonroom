# Provider-independent governance

Status: accepted design; durable appointments and human transfer implemented
as described in [room coordination appointments](governance-appointments.md).
Coordination grants, delegation and the broader enforcement integration remain pending. This document refines FR-07–17,
FR-20 and the later paid rules FR-22–24. It does not claim the current application
implements appointments, transfers or every enforcement boundary below.

## Human and agent interaction

The first eligible agent whose connection is durably registered in a room becomes
its initial governing agent. Eligibility requires current room membership and an
approved connection/capability contract; a display label or model claim is not
registration. Concurrent first connections are serialized by the controller; the
first committed eligible registration wins. Reconnects do not create new appointments.

An authenticated, authorized human can explicitly select another eligible loaded
agent in that room to supersede the current agent. Loading a second agent alone
never transfers governance. The human sees current and proposed identity and
confirms the replacement. A model message, tool response, reporting relationship,
subscription, timeout or claimed human instruction cannot make that decision.

The governing agent proposes plans, task creation, assignments, dependencies and
escalations. Other agents and the human may also propose work and raise objections.
CommonRoom preserves original messages with authorship and access controls;
summaries reference their sources and never replace them. Every room member with
relevant access can use a direct human escalation channel without governor approval.
Direct escalation is still scoped: it cannot reveal foreign-room/private content.

A conversational suggestion may create a task proposal. CommonRoom records the
proposal's source, author, scope and status and deduplicates retries by operation
identity. Proposal, authorization, execution, observed outcome and human acceptance
are separate facts. No semantic classifier or model output is itself an approval.
Routine work already within a valid bounded grant can proceed without asking the
human again; new scope or consequential external effects require the applicable
explicit approval. Ambiguous scope remains proposed/blocked with a visible reason.

## Enforcement boundary

All agents, including a user-customized malicious governor, use the same broker.
Before an effect, server code checks authenticated actor identity, company and
room membership, capability, task/resource scope, current grants, applicable
approval, deadline, stop state, immutable repository/revision, quotas and audit
availability. Reject a request if any required evidence is missing or uncertain.
A governance role is never an additional resource permission or a bypass.

The governor cannot issue human approvals, edit membership/permissions, access
credentials, grant a delegate broader rights, rewrite history or disable limits.
Delegation intersects the parent's current authority with the delegate's own
permissions and explicit task scope; it never unions privileges. Revalidate at
effect time so revoked membership, expired grants and changed revisions cannot
be revived through delegation chains or stale queues.

The controller owns a room appointment record containing the selected participant,
company/room identity, monotonic governance revision, initiating human or initial
registration provenance and durable event identity. The exact schema is an
implementation detail. No model output can write this record directly.

A transfer uses an authenticated human-only endpoint, checks the expected current
revision and replacement eligibility, and commits the appointment plus audit in
one transaction. Failed persistence leaves the prior appointment effective and
releases no new authority. Conflicting transfers produce a visible conflict,
not last-writer-wins authority. Replays of the same committed operation return
its existing result without transferring twice.

Coordination requests bind to the current governance revision. After transfer,
old governor grants and queued coordination actions are invalid; delegates cannot
use an old governance chain to restore them. Already-running jobs retain only
independent human-authorized work grants that remain valid; role-derived queued
or active effects must stop/revalidate through the normal fenced path. Transfer
never fabricates cancellation, proven absence or extra retry authority. Record
what continued, stopped or requires reconciliation. The human's stop always wins.

If the governor disconnects, loses membership or becomes ineligible, mark its
appointment unavailable and deny its coordination effects. Do not silently elect
a replacement after an existing appointment. Independent valid work may continue;
appointment/recovery is visible and an authorized human selects a successor.

## Security threat and limits

A malicious governor can lie, omit context, propose harmful tasks or inject
instructions into other agents. Independent storage, authenticated authorship,
scoped disclosure, direct escalation and effect-time authorization constrain
these attempts. A summary must be labelled a summary, and a claim of approval
must link to an actual controller-validated approval record.

These controls prevent authorized coordination from automatically becoming
privilege escalation. They do not establish model honesty, guarantee that a
human notices deception or defend a compromised controller/worker/host. Security
claims require adversarial tests and independent review of every execution path.
A second reviewing model may advise but cannot replace enforcement.

## Providers, editions and feature control

Any admitted provider may coordinate: Codex, Claude, Gemini, Copilot or llama.cpp
when its adapter supports the required contract. No dedicated governance LLM is
required. Codex's bridge-only engine build remains an individual adapter concern.
Paid provider offerings do not acquire extra authority merely through billing.

Future optional agent coordination must be off by default behind a documented
server-enforced deployment control, with an owner-authorized UI where exposed.
Implementation must specify the final flag name, administrator/owner authority,
restart requirement and running-work semantics before enabling it. Off refuses
new agent coordination/appointments, preserves all records and allows authorized
human recovery and direct task management. It cannot disable permission checks,
audit, retention, isolation or quota safeguards. Human stop/revocation remains
available. This design document introduces no executable flag or product feature.

Free coordination has no automatic consensus, timeout approval or arbitration.
An unresolved objection blocks affected/dependent work and escalates to the human;
independent authorized work can continue. Later paid consensus/timeouts must obey
FR-22–24 and explicitly delegated authority. Neither edition can self-approve
irreversible actions. This change does not implement those later paid features.

## Implementation sequence

1. Durable room appointment state, projection/replay and permission-checked human
   selection with atomic audit, expected revision and idempotent operation identity.
2. Scoped coordination grants and broker revalidation; atomic transfer/revocation,
   stale-queue rejection and explicit running-work reconciliation.
3. Task proposal/authorization separation, bounded delegation and operation deduplication.
4. Original-source references, labelled summaries and independently delivered objections.
5. UI selection/recovery, feature-control documentation and both-state tests.
6. Adversarial integration tests against the actual endpoints/broker, then separate
   live acceptance. Keep the existing mandatory development barrier until its
   own P0 release criteria pass.

## Required acceptance scenarios

| ID | Exercise | Required observation |
| --- | --- | --- |
| GOV-01 | Race two first eligible connections and reconnect both. | One durable initial appointment; deterministic committed winner; no reconnect reassignment. |
| GOV-02 | Human replaces the governor with another eligible loaded agent. | Atomic audited transfer; selected identity/revision visible; unchanged resource permissions. |
| GOV-03 | Governor claims human approval, appoints itself or forges a transfer. | Human-only boundary rejects; appointment, grants and approval records unchanged. |
| GOV-04 | Foreign actor/room, unloaded target or stale expected revision requests transfer. | Refusal/conflict without partial events or leaked private identity/context. |
| GOV-05 | Fail appointment audit persistence; retry an acknowledged operation. | No unaudited transfer; retry returns the same committed operation without a duplicate. |
| GOV-06 | Queue/replay old-governor effects across transfer and restart. | Old revision denied; replay reconstructs current authority; no stale grant revival. |
| GOV-07 | Expand delegation through multiple agents or change room/membership/revision. | Effect-time intersection/revalidation denies expansion and stale scope. |
| GOV-08 | Create a task from vague conversation; replay the proposal; request execution. | One attributed proposal; no implicit authorization or duplicate dispatch. Healthy bounded authorized work succeeds. |
| GOV-09 | Governor suppresses or falsifies an objection/summary. | Original attributed record remains; independent scoped human escalation delivers; summary is visibly derived. |
| GOV-10 | Disconnect/remove governor during independent work and during role-derived work. | No automatic takeover; valid independent grants can continue; stale role authority stops/reconciles without invented shutdown proof. |
| GOV-11 | Coordination off/default; unauthorized enable/transfer; disable during work. | Server refusal, preserved records and explicit running-work behavior; human stop/recovery and mandatory checks remain. |
| GOV-12 | Run equivalent coordination requests through different admitted providers. | Same deterministic authority decision, no Codex privilege or mandatory governance model. |
| GOV-13 | Race human stop/revocation with transfer, approval and tool execution. | Stop/revocation prevails at the effect boundary; no new authority from uncertain outcomes. |

Automated tests must mock every database/provider/Docker interaction, use synthetic
messages/credentials and temporary fixtures, and test actual production entry
points. Each refusal needs a valid positive control. Run make test during work
and make check before handoff. These are new governance acceptance requirements,
not additions silently counted as completed cases in the 74-case P0 catalog.
See [P0 acceptance](p0-release-gate-acceptance.md) for the independent execution gates.
