# Workspace probe authority channel

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


Workspace inspection needs current human authorization and durable controller
acknowledgments while the worker holds creator ownership. The internal probe
channel connects these checks to the retained-workspace probe lifecycle without
using agent output as authorization. It does not expose a public operation or
make file deliverables available yet.

## Configuration and operation

This is mandatory security infrastructure, with no product switch, HTTP route or
UI control. Existing task flags cannot disable its checks. Operators cannot enable
exports merely by setting an environment variable. Production authenticated peer
setup and independent physical evidence remain required integrations.

Trusted assembly calls `probeRetainedWorkerWorkspaceChannel` with an approved
Docker runner, immutable reviewed helper image, authenticated private stream
connector and local authority-stop callback. For example, a future authenticated
worker session connects its supervisor-owned stream to `serveProbeAuthority` on
the controller side. Both ends must receive the same exact retained creator record
and fresh probe intent from trusted session setup, rather than accepting identity
from a message. The controller guard must independently enforce human/task access,
physical worker/volume/image evidence and durable account admission using the
controller ledger. There is no permissive guard supplied by this channel.

The existing direct internal guard assembly remains available for trusted in-process
use and automated verification. It does not bypass the observation, journal,
physical inventory, authorization or audit checks.

## Protocol and security

Frames use a four-byte network-order length followed by canonical JSON, bounded to
4096 bytes. Version 1 requests contain a strictly increasing sequence, operation,
exact intent and lifecycle event. Unknown/duplicate fields, alternate encoding,
wrong versions, skipped/replayed sequences, changed intent and inappropriate event
fields refuse before invoking controller authority. The controller returns only a
matching version/sequence acceptance after its guard succeeds; errors close the link
without exposing internal diagnostics.

Operations are `authorize`, `observe` and `acknowledge`. Observation sends no
workspace paths or source bytes; the controller uses its session-bound record and
passes a detached copy into each callback. Lifecycle validation and persistence
remain the guard/ledger's responsibility; parsing an event is not physical proof.
The worker journal still records local intent before controller acknowledgment.
Helper creation/start and proof release retain the existing durable ordering.

The stream must be supervisor-owned, authenticated separately and never connected
to model stdin/stdout or mounted into either container. The channel's constructor
checks bindings, not peer identity. There is no reconnect or automatic replay after
uncertain acknowledgment. Each exchange has a five-second ceiling; serving has a
two-minute lifetime ceiling, both shortened by caller cancellation. Denial, invalid
response or uncertain I/O poisons the client, closes the stream and invokes its local
authority stop once. Cancellation unblocks owned I/O. Cleanup still preserves
uncertain outcomes and requires normal reconciliation; link failure proves neither
container absence nor released capacity.

A source fork can replace the trusted connector or controller guard and falsify
its guarantees. Mandatory isolated execution has no off switch, direct provider
launch or host execution fallback.

## Verification and remaining work

Tests use temporary synthetic retained jobs, mocked Docker/persistence and in-memory
streams. They cover the real probe/journal lifecycle over the channel, durable
begin/bound/finish refusal, exact binding, malformed/oversized/truncated frames,
replay, denial, wrong acknowledgments, callback mutation and cancellation. Every
database/provider boundary remains mocked; no real provider or Docker is launched.

Still required: deployed authenticated stream establishment/invocation, production
independent physical evidence, continuous reference enforcement,
live Linux image acceptance, frozen source-root authority, physical storage quota
and authenticated deliverable capture/delivery. This channel alone is not a
FrozenLease, source-byte capability, complete export feature or live P0 acceptance.

The [controller guard assembly](controller-workspace-probe-authority.md) now
implements the human/dependency checks, durable authorization and lifecycle ledger
connection behind this channel. Physical evidence and deployed peer authentication
remain mandatory external infrastructure, with no permissive defaults.
