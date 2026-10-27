# Bounded workspace source capture

The internal `workspacesnapshot` package captures actual source content and selected
output bytes for artifact custody. It is not a public export endpoint or a working
production exporter. Managed file and bundle dispatch remains blocked.

## Usage and authority

Trusted worker/controller assembly supplies a pinned `os.Root`, exact execution,
import commit and creator-volume binding, declared members and a `FrozenLease`.
`Capture` refuses missing or typed-nil leases and caller-supplied revision digests.
The lease must authenticate the root and creator, establish a read-only isolated
view, fence every other writer, enforce current access/export policy and durably
authorize bounded source observation. Existing workspace storage must already be
accounted. A read-only handle alone does not fence another writable mount.

The scanner itself supplies no lease. The [fixed-root capture assembly](isolated-source-capture.md)
now combines actual kernel/root checks with mandatory live supervisor authority;
its reviewed helper/runtime and deployed authority transport remain integrations.
Scanner unit tests explicitly use synthetic leases. A source fork supplying a host root or dishonest lease invalidates the
isolation guarantee; there is no default lease, host fallback or isolation toggle.

## Content and bounds

Capture observes the complete tree twice, including empty directories, actual
file modes, byte counts and SHA-256 content digests. Only a safe root `.git`
directory is excluded, without reading its contents or executing Git. Symlinks,
hardlinked files, special files, unsafe ownership/modes, malformed paths and
case variants of `.git` refuse capture. The imported commit is provenance, not
proof that Git HEAD or the workspace is unchanged.

Limits are 10,000 entries including the root, 32 directory levels, 2,048 UTF-8
path bytes, 256 MiB total source, 16 MiB per file, 32 selected members and 16 MiB
aggregate declared selected capacity. Selected names must be unique. Cancellation,
lease revocation, changes between observations and missing members refuse.

The generated revision digest binds execution/creator/import metadata and actual
canonical tree entries. The separate snapshot digest additionally binds selected
members. Selecting another output does not change the source revision identity.
Neither digest establishes useful work or independently observed test success.

## Custody integration

`Source.Open` creates a lazy reader for an exact revision/output/member/maximum.
It does not consume source on construction. The custody store must durably reserve
the new object and acknowledge physical quota before consuming that stream.
The first read rechecks the complete source; subsequent chunks check the lease.
Capture itself is a separately authorized bounded observation of existing source,
not a new on-disk custody object. It buffers only selected bytes.

`VerifyReceipt` rechecks the complete source and exact scope, maximum, actual
length and content digest before sealing. A source change during copying refuses
sealing; uncertain objects and charges remain retained rather than automatically
refunded or retried. Close clears the reader's private copy. The caller retains
the root and lease until use finishes.

Server integration tests connect actual temporary source bytes through the real
custody store and controller journal to finite manifest resolution/revalidation.
All database, worker and authority boundaries are mocked or explicitly synthetic.
No provider, Docker daemon, credential or live database is used.

## Availability and remaining work

This mandatory internal security infrastructure has no product feature flag and
exposes no new user controls. Existing task flags and permissions remain enforced.
Users cannot yet export task files through this package.

Production still needs the fixed-root lease connected to its reviewed running
helper and authenticated supervisor/writer-fencing transport, real
physical quota and complete inventory, current source/seal/release policy assembly,
crash reconciliation, retention/disposal and authenticated audited public delivery.
Live Linux worker acceptance remains unverified. Synthetic mutation tests cannot
prove real writer fencing or certify any P0 release gate. Linux-specific filename
behavior and exhaustive entry/tree-capacity boundaries are not established by the
current macOS test run.

See [object custody](object-custody.md), [controller journal](artifact-custody-journal.md)
and [finite manifests](artifact-manifests.md).

[Retained workspace observation](retained-workspace-observation.md) adds creator-lock and exact unattached-volume admission; it does not yet provide a frozen source mount.
