# Agent CommonRoom — authoritative functional specification

Version 3 · 15 October 2026 · owner-confirmed product baseline

This specification supersedes version 2 and conflicting requirements in older
planning notes, demo copy and development history. It records the owner's
confirmed interview decisions. A requirement is not a claim of implementation.
See [the gap assessment](docs/v0.1-gap-assessment-2026-10-15.md) and subsequent
verification records for actual delivery. Complete P0 before P1.

## Purpose and distribution

Agent CommonRoom is an open-source AI harness where a human and agents discuss
problems, challenge proposals, delegate tasks and inspect their results. Different
models collaborating is a hypothesis to evaluate, not a promised quality gain.

FR-01: The free harness supports one human owner and multiple agents. Daemon Labs
also intends a genuinely free hosted harness. Self-hosted operation is required
for v0.1; hosted availability is targeted but may slip without weakening security.
The free hosted service may store conversations, tasks, decisions, audit records
and submitted results with disclosed data flows.

FR-02: Paid multi-human operation and commercial implementations belong in a
separate private repository. Preserve existing historical accounts, identities,
messages and audit records during distribution changes. Do not rewrite or delete
history to enforce product packaging. Security foundations remain free.

## Connections and agents

FR-03: v0.1 requires Claude CLI, Codex CLI and llama.cpp connections. CLI
subscription authentication remains supported; do not silently substitute
separately billed API access. A connection uses a supported CLI/model interface,
not automation of a desktop application's windows or a chat website URL.

FR-04: One connection may supply several agents with independent sessions and
room profiles. Profiles include names, responsibilities, instructions and
constructive/adversarial working styles. Personas, titles and numeric legacy
ranks do not grant permissions. Do not concurrently attach the same live session
to independent memberships.

FR-05: Support self-hosted controller/browser operation and approved local or
remote workers (LAN/VPC where configured). Current supported execution requires a
Linux rootless-Docker worker; native Mac/Windows worker support cannot be claimed
without verification. No direct local/SSH provider launch or uncontained fallback.

FR-06: In hosted use, user-controlled workers retain provider credentials; the
hosted controller must not receive them. Disclose worker-side retention and
session copies accurately. Provider login, credential storage, sign-out and
revocation are distinct operations. Source code leaves the worker only through
explicit sharing or authorized provider context. Keep credentials out of model
tools, audit payloads, logs and exported artifacts.

## Rooms, governing agent and delegation

FR-07: Rooms hold conversations, objectives, profiles and tasks. Agents may
propose approaches, raise objections and divide authorized work across agents
and rooms while preserving permissions, context disclosure and provenance.

FR-08: The first connected agent is the initial governing agent. The human may
appoint a replacement. Persist and audit appointments. The governing agent
coordinates and delegates within granted authority; it never inherits the
human's credentials or permissions merely through seniority.

FR-09: The human authorizes identifiable bounded work and its resources. The
system records task assignments, dependencies, instruction changes and resulting
attempts. Independent authorized work can proceed concurrently. Completion is
not human acceptance; distinguish model reports, human evidence and observed checks.

FR-10: Free agents may debate but have no automatic voting, consensus arbitration
or timeout approval. An unresolved objection blocks the affected task pending a
human decision; dependent work also waits. Independent authorized tasks may
continue. Record alternatives and reasons. Human absence never grants authority.

FR-11: An authorized human's explicit decision resolves the dispute. Agents record
concerns and then follow the selected direction within security boundaries. New
material evidence may justify another objection. Later instructions cannot
silently override earlier instructions without the required authority.

FR-12: Reuse the existing company reporting chart. It describes relationships and
default escalation structure, not execution permissions. Legacy numeric seniority
must not determine decisions, even when the chart UI is disabled.

## Execution, approval and recovery

FR-13: Every managed model turn uses the isolated worker. Development tools run
in a separate credential-free container from the provider engine. Authenticate
supervisor channels independently of engine stdout/tool arguments. Bind authority
to company, initiating human, room, participant, connection, job, approved
repository and deadline. Enforce aggregate quotas and verified pair shutdown.
The mandatory legacy admission barrier stays until the replacement is verified.

FR-14: Every action potentially causing irreversible external consequences must
announce its exact intention and wait for explicit approval. Bind approval to
scope, action, target and revision; material changes invalidate it. Silence never
approves in free. Approved isolated workspace edits/tests need no repeated approval;
publishing, deployment, messaging, spending and external-data changes require
separate approval. Unknown cost/impact requires approval.

FR-15: Automatically recover connectivity and continue safely where supported.
Check durable job state and reconcile uncertain execution before retrying an
operation that may already have caused consequences. Prevent duplicate dispatch
and results. Unresolved ambiguity must remain visible; do not invent completion,
retry authority or confirmed shutdown. Human stop and integrity stops prevail.

FR-16: Browser refresh/disconnection preserves committed work independently of
browser presence. Show pending, blocked/pending-decision, failed and completed
states accurately. Cancel fences future actions, stops both containers and
requires immutable-ID absence proof; EOF or deadline alone proves no shutdown.

## Audit, memory and integrity

FR-17: Record human and agent decisions, proposals, objections, approvals,
delegations and instruction changes with actor identity and causal task/action
links. Persist intent before tool execution and outcome before releasing results.
Record stated reasons, not a claim to private model reasoning. Historical
corrections append records; prior evidence is never rewritten.

FR-18: Preserve append-only events, restricted database roles, protected metadata,
chain verification and integrity-stop behavior. Failed persistence grants no
execution permission. Independently protected archive/checkpoints are later work;
do not claim administrator-resistant immutability from local hashes/triggers.
Retain existing retention safeguards. No purge without a separately implemented,
authorized protection/receipt contract; disabling a feature never deletes history.

FR-19: PostgreSQL remains the application store. Preserve scoped revisioned shared
memory and source attribution. The owner intends vector-backed base memory;
pgvector installation alone is not semantic retrieval. Do not describe vector
retrieval as working before embeddings, access controls and retrieval are verified.
Fulginn adds enhanced caching/recency-aware memory for paid clients later.

FR-20: Authentication, authorization, audit integrity, retention protections and
container isolation cannot be bypassed by the owner, payment, feature flags,
personas or plugins. Integrity failure/unavailable verification blocks agent work;
human investigation and explicit verified recovery remain available.

## Deferred and paid scope

FR-21: MCP and skills must be visibly marked **coming soon** for v0.1. Existing
registration placeholders do not imply discovery, authentication or execution.
A public plugin interface is future extensibility work, not permission to expose
provider-native tools or unreviewed hooks. Do not confuse repo development skills
with product runtime skills.

FR-22: Later paid capabilities include multiple humans, scoped override authority,
company-defined approval categories/limits and upward escalation. Reporting lines
route requests; explicit limits authorize them. Human titles alone grant no power.

FR-23: Paid consensus uses one vote per rostered participant, majority after
hearing dissent and reconsideration, then human escalation for a split vote.
Authorized humans may override. A governing agent may resolve after a configured
human-response timeout only within explicitly delegated permissions. A paid
subscription or timeout alone never grants authority for irreversible actions.

FR-24: Daemon Labs may provide the paid governing agent; customers may replace it.
Fulginn and Veriscribe integrations are paid and outside v0.1. Veriscribe's
organizational wiki retains fine-grained attribution, distinct source/approval
authority and historical supersession links. Derived summaries and cached memories
must preserve source access restrictions. SOC 2 readiness is an enterprise goal,
not certification achieved by implementing logging.

## Release acceptance and completion

Demonstrate two agents with different personas discussing a task, bounded
governing delegation, isolated implementation, an objection/approval pause,
human resolution, reviewable artifacts and observed checks, and an inspectable
instruction/decision/action history. Demonstrate disconnect recovery without
duplicate effects. Cover Claude CLI, Codex CLI, llama.cpp and the supported
local/remote topology explicitly. Hosted release remains separately gated.

Every implementation requires meaningful fast parallel-safe tests with mocked
database interactions, full user/technical documentation and applicable
server-enforced permission-aware feature flags with both states tested.
Mandatory security boundaries and the single-human free distribution invariant
have no bypass toggle. Run `make test` during development and `make check` before
completion. Never launch real providers or connect to databases in automated tests.
Never claim synthetic tests establish live provider/runtime compatibility.
