# Read-only sealed-object journal relay

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


The credential-free isolated object reader needs current controller authorization
and durable source release acknowledgment before its store can send bytes. The
read-only journal relay supplies those operations over the same private stream as
[sealed-object delivery](sealed-object-read-channel.md), without database credentials
or a permissive local journal. It adds no user feature, public route or optional
security bypass.

## Assembly

Trusted authenticated private setup pins one exact sealed receipt, owned stream and
finite reader lifetime. Construct `objectcustody.NewReadJournalChannel` with that
context/stream/receipt and a local nonblocking authority-stop callback. Pass it as
the journal to the trusted store opened on the actual isolated read-only private
object root. Then synchronously call `ServeReadChannel` with that same stream/store.
Only the serving protocol owns the initial request and final object/body framing;
the journal owns the two intervening control exchanges during store `Get`.

The parent uses `ReadChannel` and its mandatory current controller release guard.
That guard must authenticate the reviewed running isolated reader, actual sealed
object/root and volume identity, physical quotas/account state and current human
export/retention policy. Decoded receipt fields are not authentication. Do not expose
the stream to a model, share it with another task, mount a host root as fallback or
supply a synthetic reader/journal to production. Actual isolated reader image/mount
inspection and deployed private setup remain incomplete.

## Sequence and failure

1. The receiving supervisor checks current read authority and sends its pinned request.
2. The trusted store asks its relay journal to authorize that exact receipt. The
   parent rechecks current authorization before sending a positive source grant.
3. Store `Get` verifies actual private inode, length and digest, then asks its relay
   journal for durable source release acknowledgment. The parent reauthorizes and
   persists that acknowledgment before its positive response permits store return.
4. The serving protocol sends the exact object header and bytes, then closes.
5. The receiver checks length, digest and complete EOF, then separately obtains its
   final durable downstream release acknowledgment before returning any bytes.

The parent accepts only source-authorize, source-acknowledge, object in that order.
It refuses repeated/skipped controls, substituted receipts and incomplete phases.
Direct trusted readers with their own authenticated journal may send the object
without relaying controls; this never bypasses the receiver's final checks or the
trusted reader's mandatory store journal.

Relay replies require the exact version, expected phase and pinned receipt. Each
exchange has a five-second deadline bounded by the caller's lifetime and cancellation
closes the stream. Any denial, malformed response, wrong order or uncertainty poisons
the relay and stops its local authority once. There is no reconnect or automatic
replay. A fresh context cannot revive a failed stream. Reserve and Seal always refuse,
so a read helper cannot create objects or alter quota through this journal.

A source acknowledgment can persist even if a later transfer or receiver acknowledgment
fails. Keep that authorization history, object and original storage charge. Neither
source nor receiving acknowledgment proves a user received or accepted the artifact.
No failure refunds/deletes objects or creates a fabricated completion record.

## Verification and deployment limits

Tests use actual temporary synthetic custody objects and in-memory pipes, mocked
journals/controller DB and witnesses. They cover relayed real store reads, durable
source/receiving audit, source acknowledgment denial, phase-order attacks, substituted
replies, expired lifetime, poisoned reuse and prohibited writes. A controller integration
test joins its pinned read session and existing custody ledger to this relay. No real
DB, Docker daemon, provider, credential or private message is used.

A fixed isolated reader entrypoint/image, kernel read-only root/mount observation,
container lifecycle/account supervision and deployed authentication still require
assembly. Actual release/quota/account witnesses, physical volume quotas, end-to-end
task artifact delivery/human review, reviewed rootless image/latency/live acceptance
and unknown-ID incidents remain gaps. Required file-task dispatch stays blocked;
synthetic relay integration certifies no live P0 release gate. A source fork can
substitute trusted setup/guards/roots and invalidate isolation guarantees.
