# Isolated capture helper and private control channel

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


The capture helper now connects fixed-root source capture and private object
persistence to live supervisor authorization and durable custody callbacks. The
helper is credential-free and runs only in the reviewed isolated Linux container.
This is not an enabled file-task/export endpoint; reviewed running-view admission,
physical quotas and production controller/worker routing still need assembly.

## Configuration and trusted use

`cmd/workspace-capture` accepts exactly one canonical base64url JSON `Request`, with
finite binding, member list and fresh challenge. The encoded argument is limited
to 64 KiB. Unknown/duplicate fields, alternate encoding, extra arguments and invalid
contracts refuse before roots are opened. There are no source/object path switches,
provider commands, model stdin or host execution fallback.

Production opens only `/workspace` for source and `/objects` for custody. The latter
must be a separate private quota-backed object volume configured by trusted worker
admission; neither the helper nor directory modes create physical quota. The helper
must have UID/GID1000, read-only root/source, exactly approved mounts, no credentials,
no network or extra capabilities and bounded resources. Its stdio is a dedicated
supervisor-owned private stream. It must never be passed through a model or mounted
as a model-accessible socket.

The parent binds the same immutable request before starting
`workspacecapture.ServeCaptureChannel(ctx, stream, request, authority, journal)`.
Its context/human/worker identity must come from authenticated session setup, not
request metadata. `authority` independently checks current human/task/export policy,
the reviewed running helper, exact creator/volume/root and writer exclusion, and
durably acknowledges observation. The actual custody journal must enforce account
admission, physical object quota, actual source/content/format and seal integrity.
Neither callback may reenter the same stream. Missing or synthetic implementations
cannot admit production work.

`docker/capture/Dockerfile` builds a separate scratch image containing the fixed
helper, owned empty mount points and license notices. Its compiler must be a reviewed
immutable digest-pinned Go image with the official `/usr/local/go` layout. Build steps
use local toolchain/offline module settings and no network. No mutable compiler
image default or automatic image approval is supplied. The old metadata probe image
is unchanged; it cannot be used to launch the capture executable.

## Protocol contract

Each message has a four-byte network-order length and canonical JSON. Frames are
limited to 128 KiB and sessions to 4096 messages. Version, strictly increasing
sequence and the SHA-256 of the complete validated request bind every request and
acknowledgment. The digest is non-secret metadata, not authentication. Unknown or
duplicate fields, alternate spellings, replay, changed binding and mixed payloads
refuse before authority/journal callbacks.

Only four operations exist:

- `observe`: current fixed kernel mount proof, with the same mount identity and
  challenge throughout the session. Parent approval precedes helper source reads.
- `reserve`: exact selected member/maxima and one computed revision. Requires a
  prior approved observation and current authority plus durable physical-quota
  admission. Duplicate member/object identities and changed revisions refuse.
- `seal`: exact previously reserved intent and actual bounded byte receipt. Requires
  current authority and durable seal acknowledgment; duplicate/unreserved seals
  refuse. Actual source-byte verification occurs in the source-bound adapter and
  independently in the real journal's seal guard.
- `finish`: the complete exact set of acknowledged sealed receipts, computed source
  scope and snapshot digest. Missing members, altered receipts and changed scope
  refuse. Final current authority must pass before acknowledgment.

No file-byte, provider-command, read/download or release operation exists. Parent
results are private metadata, not proof of public delivery or task acceptance.
The report's source identity relies on the authenticated reviewed helper and its
physical admission; JSON validity or a hash alone is not independent attestation.

The client closes/poisons a stream and stops local authority on denial, uncertain
I/O, malformed acknowledgment or cancellation. There is no reconnect or automatic
replay. Each exchange has a five-second ceiling; the control session lasts at most
thirty seconds. The executable also has a thirty-five-second cancellation context.
Callbacks must cooperate and the supervising worker must enforce actual process
termination and exact-ID cleanup; Go cancellation is not an unconditional kill.

Once a parent journal reservation has been attempted, any later protocol failure
is reported unconfirmed. Durable charges, partial files and individually sealed
objects remain retained. A failed/lost completion acknowledgment never authorizes
retry, deletion or refund. The parent must additionally confirm helper process exit,
cleanup and current policy before publishing a successful capture report.

## Flags, permissions and verification

This mandatory private authorization/custody infrastructure has no product toggle,
public HTTP route, UI setting or direct host launch exception. Existing task flags,
human permissions, audit stops, quotas and retention are enforced by parent authority
and journal in both flag states. File/bundle task dispatch remains blocked until
all required runtime/quota/release integrations are verified.

Tests use synthetic temporary files and private in-memory streams with mocked
kernel/supervisor/quota/database/provider boundaries. Actual scanner/store/channel
integration persists bytes and returns exact receipts. Tests reject malformed,
oversized/truncated/mixed frames, replay, changed mounts/revisions, duplicate IDs,
unreserved/altered seals and incomplete/altered completion; test message limits,
cancellation, wrong acknowledgments, quota/seal denial and retained uncertain data.
Executable tests explicitly mock the kernel capture boundary and exercise argument
refusal, fixed root wiring and the real object store/private channel. No real
provider, database, Docker or SSH is launched by these tests.

An offline Linux cross-build verifies compilation only. No capture image build,
live mount, real quota or public file release is certified. Remaining work includes
reviewed image acceptance, supervised container/stream and running-view validation,
authenticated deployed evidence/controller journal assembly, physical quotas, output
format checks, recovery and authenticated deliverable delivery/review. A source fork
or trusted daemon administrator can invalidate the boundary by replacing the image,
stream, callbacks or physical evidence. No live P0 gate is claimed.

The [capture-container admission contract](capture-container-admission.md) defines
fixed launch arguments and exact inspection policy for a future supervised runtime.
It does not launch the helper or supply physical quota/authority enforcement.
