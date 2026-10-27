# Isolated read-only source capture

The capture lease now joins the existing finite source scanner to the running
container's fixed read-only workspace and a live supervisor authority. A previously
completed mount probe report cannot open a source root. This component is internal
infrastructure; required file/bundle task dispatch and public file delivery remain
blocked until the complete runtime, quota and custody path is assembled.

## Trusted use

A reviewed credential-free capture helper calls
`workspacecapture.WithFrozenSource(ctx, binding, members, challenge, authority, use)`.
It opens only `/workspace`; callers cannot provide a source directory, host mount
or replacement kernel observer. For example, after supervisor admission a helper
may capture a declared `report.txt` member and lend its lazy reader to the private
custody pipeline inside the synchronous `use` callback.

The binding identifies the exact company/room/task/attempt/job, task definition,
repository/import base and retained volume nonce/creation time. Its revision digest
starts empty: the scanner computes the actual revision from bounded source entries.
Members identify only declared finite outputs with explicit byte ceilings.

`Authority.AuthorizeFrozenSource` is mandatory on every lease verification. It
must authenticate the exact running reviewed helper and original human/task/export
scope, retain actual creator ownership, independently verify complete exclusion
of other writers, verify existing workspace storage accounting and durably authorize
bounded source observation. A model, plugin, old mount report or requested image
cannot provide these guarantees. The authority must honor cancellation and cannot
reenter the same source. No permissive default or deployed authority transport is
provided by this package.

## Root and lifetime checks

Before every verification/read, the lease checks the pinned root handle is the
same directory as the fixed `/workspace` identity, reads actual Linux kernel mount
evidence and requires current supervisor authorization. It then rechecks root
identity and mount evidence after the potentially remote authorization. The mount
ID, challenge and read-only evidence must remain unchanged throughout capture.
Kernel checks require the existing reviewed UID/GID1000, read-only container root,
explicit read-only workspace mount and no workspace submounts.

Source authority is valid for at most thirty seconds, shortened by caller cancellation
or an earlier deadline. That deadline is propagated to supervisor calls. Callback
success, failure, panic or timeout expires the lease before its root is closed.
Retaining a Source or constructing a new reader with a fresh context does not revive
it. A final complete-source revalidation precedes successful return. Trusted callbacks
must cooperate with cancellation; the future supervised helper runtime must also
enforce its process deadline. A Go context alone cannot terminate arbitrary blocked
callback code.

The scanner still enforces bounded actual tree enumeration, double observation,
path and file-type restrictions, ownership, no hardlinks/symlinks/special files,
no Git execution and no reads of root Git metadata. A source mutation or changed
unrelated file refuses. These checks supplement actual writer fencing; repeated
hashing or a read-only local mount alone does not freeze other mounts.

## Custody, permissions and feature switches

Capture buffers only bounded selected bytes for authorized observation and creates
no persisted output object. A lazy member reader must be consumed only after
custody has durably reserved its maximum size and enforced physical object-storage
quota. Actual content/size must still be sealed against the source, and authenticated
release must obtain current authorization and durable read acknowledgment.
A capture digest is neither human acceptance nor an independently observed test.

This mandatory isolation/authority infrastructure has no product off switch, UI
setting, HTTP route, host path parameter or uncontained fallback. The supervisor
must enforce current task/feature permissions; a disabled product feature never
waives authorization, audit, storage or retention. Nothing here enables file tasks,
changes saved history or refunds retained storage. Supported managed capture remains
inside the reviewed Linux container; other platforms have no host fallback.

## Verification and remaining integration

Tests use only synthetic temporary source files with explicitly mocked kernel view
and supervisor checks. They exercise actual scan/revision/member bytes, immutable
scope, revocation before first read, writable/changed mounts, replaced root,
malformed bindings, missing members, symlinks, cancellation, missing/typed-nil
supervisor authority, bounded deadlines and callback error/panic lifetime expiry.
The public entry point always uses fixed kernel/root checks; tests cannot configure
a production host-path fallback.

This is not live Linux acceptance. Remaining integrations include the reviewed
capture executable/image and supervised stream, physical validation of its running
view, authenticated controller/worker routing, physical workspace/object quotas,
custody transfer and authenticated file delivery/review. A hostile trusted daemon
or a source fork that substitutes authority/kernel checks can invalidate the claim.
No live P0 release gate is certified by these synthetic tests.

Controller custody now accepts the fenced retained job phase while the original
attempt and deadline remain current, as well as a submitted job. It still refuses
allocated/confirmed/expired jobs, revoked definitions and denied physical source
proof. This phase compatibility permits future post-quiescence capture without
pretending shutdown confirmation grants source authority.

The [source-to-custody adapter](source-custody-transfer.md) now transfers selected
members through the actual object store with source-bound sealing and conservative
partial failure handling. Real quota and deployed controller/helper assembly remain
required; file task dispatch is still blocked.
