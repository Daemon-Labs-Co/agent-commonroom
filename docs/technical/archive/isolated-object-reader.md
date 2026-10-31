# Fixed isolated object reader

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


The object reader is a credential-free helper for one finite sealed artifact. It
reads only a fixed `/objects` mount, checks actual Linux read-only kernel state and
uses the controller journal relay before sending bytes. It is mandatory execution
infrastructure, not a public artifact-download feature or optional security switch.

## Runtime and configuration

`cmd/object-read` accepts exactly one canonical unpadded base64url JSON request,
bounded to8192encoded bytes. `objectread.Request` contains the exact sealed receipt
and a canonical32-hex challenge. Unknown/duplicate fields, alternate encodings,
invalid receipts, extra arguments and filesystem paths refuse. No environment option
or command flag selects a root, provider, database credential or host fallback.

The supervisor owns stdin/stdout exclusively. They must not be model pipes or user
terminal input. The helper opens only `/objects` and runs as UID/GID1000 in the
reviewed credential-free Linux container. Its overall context expires after35seconds;
object-read work is bounded to30seconds and journal exchanges to5seconds. Signals
cancel work. Failures emit only a generic refusal on stderr and exit with status2.

The parent MUST authenticate the private session, pin the exact running helper/image/
receipt/challenge and actual volume identity, fence all other writers, account for
its helper slot/storage, enforce physical quotas and independently check current
human export/retention policy. Labels, directory modes and this helper's kernel
observations are not remote authentication or proof of physical quota. Never expose
provider credentials, the Docker socket or unrelated volumes to this image.

## Mount and byte contract

`workspacemount.ObserveObjects` inspects the fixed `/objects` path and actual kernel
mountinfo/statfs. It requires UID/GID1000, owned safe directory metadata, a read-only
root filesystem and one explicit read-only object mount with no descendants/stacked
mounts. Unsupported platforms refuse; the workspace observer retains its existing
fixed `/workspace` contract. The shared parser accepts only these internal fixed
roots; no caller-selected host path API is introduced.

A private lease pins an open object-root descriptor, fixed path inode and mount
proof. It rechecks directory identity and kernel proof before and after controller
journal authorization/acknowledgment and around actual store reads. A replaced root,
changed mount ID, writable view, bad challenge, expired context or revocation returns
no bytes. This local proof grants no source workspace access or public export rights.
Independent supervisor/image/volume binding remains mandatory.

The sealed store additionally requires private0700root/0600regular single-link objects,
exact receipt size/digest and stable inode identity. Its read-only journal relays
current authorization and durable source release acknowledgment to the parent. It
cannot reserve, seal, delete or refund objects. The receiver independently verifies
the complete bytes and obtains its final durable release acknowledgment. Recorded
release authorization never proves user receipt, task completion or human acceptance.

## Image and build

`docker/object-read/Dockerfile` builds a separate scratch image with only the static
helper, fixed private object mountpoint and GPL/Go license notices. A reviewed
`GO_IMAGE` compiler reference must be pinned by lowercase SHA-256 digest; compilation
runs without dependency downloads/network. The runtime image sets UID/GID1000 and a
fixed entrypoint. The parent must still enforce a read-only root and object mount,
no network/capabilities/extra mounts, bounded resources and exact owned lifecycle.
The helper refuses actual writable root/mount state even if launch policy is wrong.

An offline static Linuxamd64 executable was compiled locally for validation. This
is not a built/pulled container image, reviewed image digest, daemon observation,
rootless runtime test or live acceptance result. No production image identity is
invented from this compilation.

## Verification and remaining integration

Automated tests use temporary synthetic actual custody objects, private pipes,
mocked controller journals and mocked kernel mount views. They cover real byte
transfer through the journal relay, writable root/mount refusal, wrong challenge,
mount/root substitution during authorization, revocation, cancellation after source
audit, canonical command decoding and mount parser attacks. No real database,
Docker daemon, provider, credential or private message is used.

Reader planning/admission and [supervised lifecycle/known-ID recovery](object-reader-supervision.md)
now have tested infrastructure. Production lifecycle journal assembly, deployed private
authentication, actual release/quota/account witnesses and physical volume provisioning
remain incomplete. End-to-end task delivery/human review, reviewed image/
latency/live acceptance and unknown-ID incident handling remain required. File-task
dispatch stays blocked. A source fork can replace private views/guards or execute on
a host and invalidate mandatory isolation; no supported fallback permits it.

See [journal relay](sealed-read-journal-relay.md),
[controller read session](controller-artifact-read-session.md) and
[mandatory isolation](architecture.md#mandatory-container-isolation-design-contract).
