# Private capture authority transport

This internal control protocol connects an isolated capture supervisor's local
journal guard to a trusted controller-side guard. It carries finite lifecycle
metadata and acknowledgments, never file contents, provider messages or credentials.
It is infrastructure supporting the mandatory execution boundary, not a completed
user-facing export feature. There is no configuration switch to bypass it.

## Assembly and use

Both peers must first authenticate and pin one retained creator record and one
reviewed capture intent, including the source identity, job, room, finite output
selection, image digest and object-volume identity. The caller supplies an owned
`io.ReadWriteCloser`; this module does not establish or authenticate a network
connection. Never use model stdin/stdout, expose the control socket inside an agent
container, or reuse another operation's stream.

Create `newCaptureAuthorityClient` with the exact record/intent and a local,
nonblocking authority-stop callback, then pass it as the worker capture journal's
mandatory guard. The controller serves `serveCaptureAuthority` with a separately
assembled `workerCaptureJournalGuard`. That guard must enforce current authenticated
human/worker permissions, flags where applicable, source or cleanup authority,
physical quotas and account capacity. Its acknowledgment must persist the exact
controller transition before returning. Wire metadata is not proof of those facts.

The server guard must not reacquire the worker admission lock, call back over this
stream, or depend on a request the waiting worker must process. The journal holds
its admission lease during exchanges. Independent lifecycle evidence may inspect
persisted local history and physical state without reentering that lease.

## Wire and failure contract

Frames contain a four-byte big-endian length followed by canonical JSON. The limit
is 128 KiB per frame. Unknown fields, duplicate fields, alternate encodings, wrong
versions, changed intents and out-of-sequence messages fail closed. Authorize frames
carry one finite transition; acknowledgment frames also carry the exact valid local
history ending in that transition. Worker-only terminal confirmation markers cannot
be sent as controller transitions. The guard independently verifies allowed history
and current evidence; frame validation alone does not authorize a transition.

The client serializes requests and checks the matching positive response. Each
exchange has a five-second deadline; the server session has a two-minute deadline
and a maximum of 64 exchanges. Cancellation closes the owned stream to unblock I/O.
Denial, malformed response or uncertain I/O stops local authority and makes the link
unusable. Any established server session termination revokes its supplied guard's
authority. Stop callbacks must be nonblocking and must not reenter the channel.
There is no reconnect, automatic replay or host execution fallback. A fresh recovery
session requires independently renewed authorization; it cannot discard uncertainty.

The local journal persists before requesting an acknowledgment. Failure afterward
keeps its reservation and trips the integrity stop. A response lost after controller
persistence remains uncertain locally. Neither peer may infer absence, refund
storage or release a container slot merely because the channel closed.

## Permissions, flags and verification

No public route, user setting or socket listener is installed. Authentication and
server-enforced product flags remain the responsibility of the mandatory guard;
cleanup may use its separately authorized flags-off policy. This transport cannot
weaken authentication, audit integrity, retention or mandatory container isolation.
Source forks can replace guards or connect untrusted streams and invalidate these
guarantees; this module is not an operating-system security boundary by itself.

Automated tests use local in-memory pipes, synthetic private temporary journals and
mocked controller guards. They cover the full journal lifecycle, retained uncertain
charges, binding/history substitutions, malformed frames, callback mutation,
revocation/cancellation and poisoned-link reuse. No real database, Docker daemon or
provider is contacted.

Deployment authentication, operation-specific quota/account guard composition,
physical source/object quota provisioning, actual worker orchestration and live
rootless image/latency acceptance remain incomplete. Required file-task dispatch
remains blocked. This protocol does not deliver artifacts or certify a P0 live gate.
See [local journal](worker-capture-journal.md),
[controller ledger](controller-capture-ledger.md) and
[independent lifecycle evidence](capture-lifecycle-evidence.md).
