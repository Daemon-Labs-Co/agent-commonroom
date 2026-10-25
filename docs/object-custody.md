# Internal immutable byte custody foundation

`internal/objectcustody` is a tested internal storage primitive. It is **not
connected to managed dispatch, provider export, a public HTTP endpoint or task
file/bundle outputs**. Those paths still refuse because real isolated-source,
accounting, physical-quota and custody-journal adapters are missing. This
foundation does not establish artifact origin, content format, secret isolation,
observed checks or production/P0 readiness.

## Contract

Trusted setup lends a descriptor-pinned, owner-only `0700` object root and a
mandatory `Journal`. The root must be a dedicated private quota-backed store,
kept open with exclusive writer control. It cannot be selected by a model or be
a host worktree or engine credential volume. Kernel/container/mount and volume
identity validation remain outside `os.Root`; descriptor containment does not
prove those properties. A dishonest callback or host root in a source fork can
invalidate the guarantees; there is no supported fallback.

`Put(ctx, scope, maximumBytes, source)`:

1. Checks the bounded company/room/task/execution/job and repository/revision/
   definition bindings, exact output/member labels, a non-nil context-aware source, and a 1–16 MiB maximum.
2. Mints an unpredictable object ID. `Journal.Reserve` must verify actual current
   authority and source bindings, durably charge maximum capacity across all
   retained/staged/uncertain account data, and establish a real physical quota
   before the module reads source bytes or creates the object file.
3. Creates a private `0600` regular file under a generated fixed leaf name with
   exclusive creation and no symlink following. A collision is not overwritten,
   adopted, deleted or automatically retried.
4. Streams actual bytes with bounded reads, checking both actual size and context.
   A non-progressing stream fails after 16 empty reads. The caller must supply a
   context-aware transport/reader: the package cannot interrupt an arbitrary
   reader that blocks forever inside its own `Read` implementation.
5. Synchronizes and reads back the actual inode bytes, checks length and SHA-256
   against the written content, closes the file, verifies its private single-link
   inode and synchronizes the directory.
6. `Journal.Seal` must revalidate current job, source revision, policy and authority
   and durably acknowledge the exact byte count/hash/scope. Only then does Put
   return an immutable receipt. Empty content is valid and still charged against
   a positive reservation.

After a reservation succeeds, **all failures retain capacity and any bytes**.
Creation, source, cancellation, oversized content, flush, identity or seal failure
cannot trigger a refund, deletion or success receipt. A canceled admission may
have a durable reservation even when no file was created. Treat uncertainty as
retained accounting state requiring explicit investigation/disposal; do not
blindly retry. Files are not discoverable through Get merely because they exist.

`Get(ctx, requestedScope, receipt)`:

1. Checks receipt shape and exact requested scope.
2. `Journal.AuthorizeRead` must require an existing exact sealed receipt and
   current authenticated room/principal/revision/export policy.
3. Opens only the generated object leaf with no-follow/nonblocking flags and
   verifies private ownership, regular type, single link, exact size and actual
   SHA-256 bytes. Symlinks, hardlinks, changed modes, foreign/forged receipts and
   tampered contents refuse. Actual bytes are buffered within the 16 MiB bound.
4. `Journal.AcknowledgeRead` must revalidate current authority again and durably
   record verified bytes authorized for release. Any refusal, cancellation,
   missing audit acknowledgment or integrity failure returns no raw bytes.

Acknowledgment is permission to release verified bytes, **not proof of remote
client receipt**. A caller changing the returned byte slice does not mutate the
stored object. No model path, URI, arbitrary filename, download, automatic
cleanup or refund API exists. Retained objects remain charged until separately
authorized exact disposal; this component cannot perform that disposal.

## Required production integration

No production `Journal` implementation exists. A real adapter must connect the
existing event-count-fenced storage ledger and mandatory decimal account ceiling,
complete trusted inventory of retained workspace/bundle/artifact data, verified
physical quota backend, actual isolated tool-volume identity, frozen source
revision and current authenticated export policy. A fake capacity counter or
synthetic quota string cannot satisfy these prerequisites. Reservation success
alone is not permission to write without physical enforcement.

A real source adapter must copy actual bounded data from a quiesced approved
credential-free tool workspace under its creator lease, reject path/symlink/
foreign-root and revision changes, and preserve the engine credential boundary.
A model's path, base64 response, provider message or completion claim is not such
an adapter. Content/format validation and independently observed artifact checks
remain additional layers: this package establishes byte identity only.

File/bundle task contracts also need verified manifests, membership/type/size
validation, durable task-output linkage, authorized download/export controls,
rollback/uncertainty recovery, metered retention and approved disposal integration.
The existing task refusal stays in place until those paths are implemented and
verified. No product feature or switch is exposed by this internal foundation;
its accounting, integrity and audit checks are mandatory infrastructure and have
no off switch. A future user-facing export capability must have its own default-
off server-enforced feature policy without bypassing these checks.

## Verification

Tests use only private synthetic temporary directories, byte streams and mocked
journal/accounting callbacks; no database, credentials, provider, Docker daemon
or live workspace is used. Coverage includes exact maximum and overflow bytes,
quota refusal before reading/allocation, real file/read-back integrity, durable
acknowledgment ordering, cancellation, uncertain retention, permission/link/path
refusals, forged/foreign/revoked receipts, current-authority changes before
release, non-progressing streams, empty content, typed-nil refusal, preexisting
object protection and concurrent independent objects/charges.

Filesystem tests on macOS/Linux do not certify a Linux rootless worker, actual
quota, crash-restart durability, volume identity, provider integration, credential
separation or G07 acceptance. Managed execution platform restrictions and the
production admission barrier remain unchanged.

Logical output/member labels are immutable scope metadata and cannot select
filesystem paths. They reject absolute/traversal/noncanonical names; only the
generated object ID selects a leaf. Trusted source adapters must verify those
labels against actual isolated outputs. See [finite manifests](artifact-manifests.md).
