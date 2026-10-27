# Private sealed-object read channel

This internal protocol transfers one finite sealed object from isolated custody to
an authorized receiving supervisor. It is a foundation for artifact delivery, not a
public download endpoint, task submission or proof of human acceptance. No new user
feature is enabled. Mandatory authorization/isolation has no bypass flag.

## Setup and use

Trusted private session setup must authenticate both peers and pin the same exact
`objectcustody.Scope` and `Receipt`, including company, room, task, execution, job,
repository/revision/definition hashes, output/member identity, random object ID,
maximum/actual size and content digest. Decoding those fields is not authentication.
Neither side accepts an arbitrary filesystem path, URI or object selected by a model.

The serving supervisor calls `objectcustody.ServeReadChannel` with an owned private
stream and a trusted `SealedReader`. Production readers must use the actual isolated,
private, quota-backed custody store and its current controller journal. Store `Get`
checks actual inode/size/content and durably acknowledges release before returning
bytes. The channel additionally verifies the returned size and digest. Never mount
a host object root as a fallback, use engine credential volumes, or expose this stream
as model stdin/stdout or a container-accessible control socket.

The receiving supervisor calls `objectcustody.ReadChannel` with a separate mandatory
`ReadReleaseGuard`. That guard must pin the authenticated principal/session and
recheck current room, exact sealed receipt, revision and export policy. It must
respect server-enforced feature/retention/integrity policy. Its final acknowledgment
must durably authorize downstream release before returning. It must cooperate with
context cancellation; `StopReadAuthority` must be local, nonblocking and must not
reenter the same stream. A no-op or synthetic guard cannot admit production delivery.

## Wire and release contract

Each stream handles one object and closes. A four-byte big-endian length precedes
a canonical JSON header bounded to4096bytes. Request/response operations differ;
version and exact receipt must match. Duplicate/unknown fields, noncanonical JSON,
wrong scope/receipt or oversized headers are refused. Before the object header, a credential-free isolated reader may relay exactly one
source authorization and one durable source acknowledgment over this same private
stream. The receiver enforces that order and refuses skipped/repeated phases. Direct
trusted readers can instead use their independently authenticated custody journal.
The body is raw bytes whose length is fixed by the pinned receipt, bounded by the16MiB
custody maximum. The
receiver reads exactly that amount, recomputes SHA-256, and requires EOF; trailing
bytes, another frame or an incomplete/unconfirmed end fail closed.

Each side imposes a30-second context deadline and closes its owned stream on
cancellation. The receiver first obtains current read authorization, then verifies
the complete response and actual body, and finally obtains a durable release
acknowledgment. It returns nil bytes on every authorization, integrity, audit,
cancellation or transport failure and stops its receiving-side authority. No
reconnect, automatic replay, quota refund or object deletion is performed.

Source-side acknowledgment records bytes authorized onto this private channel.
Receiving-side acknowledgment records permission to release verified bytes beyond
that channel. Neither proves a remote user received them, a task was completed or a
human accepted the result. A source acknowledgment may persist even if a later
transfer or receiving acknowledgment fails; retain that history and stored charge.

## Verification and deployment limits

Tests use actual private temporary synthetic custody objects and local in-memory
pipes, with mocked authority/account journals. They cover empty/binary content,
source and receiving authorization/audit denial, exact request pinning, digest
mismatch, truncation, trailing bytes, wrong versions/operations, hostile headers,
receipt bounds, missing guards and cancellation. No real database, Docker daemon,
provider, credential or private message is used.

Deployed authenticated transport, isolated reader/helper orchestration and concrete
receiving-side controller guard assembly remain incomplete. Physical quota/account
witnesses and volume provisioning, end-to-end task artifact delivery/human review,
reviewed image/latency/live acceptance and recovery incident handling remain required.
Required file-task dispatch remains blocked; synthetic byte transfer is not live P0
acceptance. Source forks can substitute the reader/guards or host roots and invalidate
the isolation/accounting contract. There is no supported uncontained fallback.

See [source-to-custody transfer](source-custody-transfer.md),
[controller capture sessions](controller-capture-session.md) and
[the mandatory isolation contract](architecture.md#mandatory-container-isolation-design-contract).

[Read-only source journal relay](sealed-read-journal-relay.md) supplies the fixed
reader-side journal client without database credentials or local acknowledgment.
