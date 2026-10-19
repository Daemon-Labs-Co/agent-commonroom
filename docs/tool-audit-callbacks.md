# Scoped tool audit callbacks

This mandatory free/core component connects the broker dispatcher to the
controller's append-only event store through bounded acknowledgements. It is not
a public API, a new connection type or a completed paired-worker deployment.

## User behavior, permissions and flags

History summaries describe "Agent tool action requested · read" and "Agent tool
action · succeeded", for example. Records contain the actual initiating human,
company, room, participant, connection, job/call identity, approved repository
digest, deadline and action/result hashes. They contain no action content, paths,
commands, raw output, capability tokens or provider credentials. Hashes correlate
evidence; they do not attest filesystem contents or stop administrator tampering.

There is no audit-off flag, owner exemption or paid bypass. Disabling optional
background-work, job-diagnostics or policy-editor UI cannot disable audit or stored
context-export restrictions. This component grants no new dispatch permissions.

The existing administrator-owned `COMMONROOM_DEVELOPMENT_WORKSPACES` switch
defaults off. `COMMONROOM_DEVELOPMENT_REPOSITORIES` must map the room to its
approved snapshot digest. These environment settings take effect in a newly
started controller process; room members cannot edit them. The callback checks
policy for every acknowledgement. Disabling development preserves history and
refuses future acknowledgements; it does not itself prove running containers have
stopped. The pair supervisor must cancel and verify both containers.

Only an enabled human with current room access may initiate work. The durable job
must name that human and company. Owner-owned provider credentials do not imply
owner identity for their user. Changed role/revision, room access, participant
revision, connection binding/revision, repository policy or stop state prevents
subsequent acknowledgement. Jobs without explicit human/company provenance remain
readable, but cannot be silently upgraded into tool authority.

Legacy credential-bearing development remains blocked. Internal paired Codex
execution and private audit transport are implemented; public application dispatch
and live acceptance remain required before enabling managed development.

## Wire and integration contract

`internal/toolbroker.AuditClient` implements `Auditor` over a supervisor-owned
`io.ReadWriteCloser`; `ServeAudit` implements the controller side. **The caller
must authenticate the peer** and keep this link separate from engine/model stdout
and tool arguments. These types create no listener, HTTP route or authentication.
Channel `Close` must safely interrupt concurrent reads/writes.

Each newline-delimited frame has exact lower-case keys:

- `version`: protocol 1 for tools; protocol 2 for the terminal idle-stop extension.
- `sequence`: consecutive integer, at most 200 tool frames/100 calls plus one reserved terminal idle frame.
- `session`: fresh random 128-bit client nonce, pinned for the entire link.
- `stage`: `intent` followed by `outcome` for that call, or terminal `idle-stop`.
- `intent`: immutable scope, call ID, operation and canonical action SHA-256.
- `outcome`: present only for outcomes; status, output byte count and hash.
- `idleStop`: present only for a terminal version-2 idle stop; bounded timing and policy metadata.

Acknowledgements contain `version`, `digest` and `accepted`. The digest binds the
canonical complete frame, including session, sequence, scope and outcome. An old
call/link acknowledgement cannot authorize a new request. This is not a signature
or replacement for channel authentication.

Frames are capped at 4 KiB. Decoding rejects unknown, missing, case-folded or
duplicate members, nulls, malformed Unicode, trailing values, arrays, invalid
order, reused calls and wrong scope/session/sequence. Declared success without the
matching full-frame digest is refused.

The dispatcher reserves authority and waits for committed intent acknowledgement
**before** invoking the isolated executor. It withholds results until the outcome
is committed and acknowledged. Cancellation closes blocked I/O. Partial writes,
EOF, refusal and malformed responses fence the link/dispatcher authority. There
is no automatic reconnect/replay. EOF/budget exhaustion is not completion or
container shutdown proof.

The controller's `serveToolAudit` assembly method requires the original human
context and matching managed-job registration. `controllerToolAuditor` reloads
state under controller coordination for each phase and appends with an expected
event count. Conflicts are rejected, not retried. It permits one unresolved intent
per job and rejects reused call IDs. Commit failure gives no positive ack and
latches the existing integrity stop. Unresolved intent requires reconciliation,
never replay. Cancellation/deadline can leave an unresolved intent; the component
does not pretend to persist an outcome after its context has been cancelled.

`worker.tool.intent` / `worker.tool.outcome` replay into internal `State.ToolAudits`,
keyed by job plus call ID. Replay rejects orphan outcomes, changed immutable
metadata, duplicate phases and invalid terminal evidence. Managed-job identity,
company and execution bindings are immutable across lifecycle updates. Timestamps
are controller-generated; deadline and repository remain bound to the original job.

Source **CAN** substitute the auditor, mix channels or skip checks, **BUT** that
invalidates durable admission and can allow unaudited access to organization
credentials/workspaces. Mandatory container isolation has no host fallback. The
pair supervisor must still enforce images, credential-free mounts, aggregate
quotas, cancellation and immutable-ID absence proof.

## Verification and limits

Tests use in-memory pipes, synthetic metadata, the existing in-memory event-store
mock and fake executors. They cover the entire callback/store/executor ordering,
flags off/on, real human attribution, commit failure, revocation, immutable
bindings, replay, malformed input, blocked-I/O cancellation and exclusion of raw
payloads/tokens. They launch no provider agents, Docker containers or DB connections.

Combined internal Codex tests exercise real session/runtime/bridge code with
mocked Docker and local protocol peers, including refused intent/outcome and
independent container cleanup. Private SSH transport and supervisor attachment
also have component tests. Public dispatch, live-provider/SSH compatibility,
operational recovery and artifact delivery remain open. Keep the development
admission barrier in place.

## Runner integration acceptance contract

The earlier `storedToolAuditor` writer has been removed. `newToolDispatcher`
requires an isolated worker job context and an explicit audit channel, constructs
`AuditClient`, and has no local persistence fallback. Its returned cleanup must
be called by the supervisor. Worker cancellation or deadline revokes authority and
closes blocked audit I/O even if a later dispatch supplies a fresh context. `tool.call.*` replay remains
readable for historical compatibility; new actions use scoped `worker.tool.*`
records. The supervisor must still authenticate and wire the channel before any
production development admission changes. A constructor is not proof of that wiring.

The implementation contracts are:

- **Human provenance and scope:** controller checks bind the initiating human,
  company, membership, connection, repository and deadline on every phase.
- **Separated transport:** private SSH/Unix attachment carries the audit channel;
  engine stdout is exclusively provider protocol traffic.
- **Ordered execution:** dispatcher waits for durable intent before execution and
  durable outcome before releasing results.
- **Cancellation:** transport loss revokes authority; the internal supervisor
  fences both role containers and retains unresolved evidence.
- **Recovery:** uncertain intents are preserved, never automatically replayed.

These contracts have component and synthetic combined-path coverage. Public
production dispatch, real SSH/container/provider behavior and deployment crash
recovery still need acceptance. See [private transport](worker-audit-transport.md),
[paired shutdown](worker-pair-shutdown.md) and
[release readiness](release-readiness.md) for the remaining gates.

## Terminal idle-stop extension

Protocol-2 `idle-stop` frames reuse the same authenticated private link, pinned
scope, sequence and random session nonce. `intent` contains only its registered
`scope` with empty call/action/hash fields; `outcome` must be absent. The required
`idleStop` object contains exactly `startedAt`, `requestedAt`, integer `minutes`
and integer `policyRevision`. No text, path, command, capability, container ID or
new execution authority is transported.

Both peers validate metadata and ordering. No tool intent may remain pending.
The controller's optional `IdleAuditor` must match the immutable admitted policy
and persist `worker.idle_stop.announced` before returning success. The ack uses
version 2 and the digest of the entire canonical frame, including idle metadata,
session, scope and sequence. A protocol-1 ack cannot acknowledge a lifecycle frame.
Old/unsupported auditors or binaries fail closed; there is no downgrade or retry.

`ToolSession.RecordIdleStop` must own the shared dispatch slot and permanently
revokes future tool authority before sending. An occupied slot returns busy
without revoking active work. Acknowledged idle stop is terminal on both ends:
further tools and repeated idle frames are refused. The controller keeps the
link open for the worker to close after consuming the final acknowledgment;
premature EOF could race the watched reader. The extra lifecycle frame does not
increase the 100-call execution budget.

Codex assembly with a registered job automatically uses this session-backed
announcer unless reviewed internal code explicitly supplies a trusted callback.
The controller auditor captures that registered job during authenticated
construction. Neither a frame nor a tool can choose another job/policy. Existing
SSH host-key validation, same-UID trust and private socket ownership remain the
peer-authentication boundary; the framing code supplies no authentication itself.
See [idle guard](worker-idle-shutdown.md#idle-stop-announcements) for retention and
acknowledgment-loss behavior. Production launch/policy transport assembly and live
SSH/provider acceptance remain open.

There is no separate transport off switch: mandatory audit cannot be disabled.
The existing owner idle-action flag controls new job snapshots, and the existing
notice-visibility flag controls delivery while retaining audit evidence.
