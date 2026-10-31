# Artifact-reader authority bridge and worker assembly

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


This internal bridge connects the isolated reader's local journal, controller
account ledger and read-only daemon observations. It preserves authenticated
identity over a supervisor-owned private stream and assembles the fixed container
reader with mandatory lifecycle and byte-release guards. It does not add a public
download endpoint, make file-task dispatch ready, or provision deployment identities.

## User and operator behavior

No new UI, public API, environment setting or feature switch is exposed. Retained
artifact permissions and human acceptance are unchanged. A successful internal read
still means only that verified bytes reached the trusted caller; it is not proof
that a user downloaded them or accepted the task.

An interrupted exchange can leave a reader charged locally, at the controller, or
both. Keep its private local record and controller audit history. Never delete a
record or reset a stream sequence to clear a refusal. A known-ID recovery requires
fresh current cleanup authority and the exact retained identity. Lost-ID incidents,
stop-marker reconciliation and deployment identity provisioning still require production
integration; reconnecting an uncertain session is not a supported recovery shortcut.

For a normal read, the sequence is:

1. Trusted setup authenticates the human and worker independently and holds the
   retained-workspace creator lease. No model/container may choose these identities.
2. Construct local/physical evidence with the approved same-daemon runner and an
   independent quota/export/creator witness; construct the pinned controller session.
3. Provision a separate `objectReadPeerBinding` at each endpoint: trusted peer
   certificate identity, approved worker alias, digest of the complete retained
   record and exact finite read intent. Call `serveAuthenticatedObjectReadSession`
   with the original pinned controller session and `newAuthenticatedObjectReadClient`
   on the worker connection. They authenticate and agree on that binding before
   authority traffic. The directory/socket must never be exposed to model or tool
   containers.
4. Construct `newWorkerObjectReadRuntime` with the retained lease, finite intent,
   authority client and independently pinned byte-release guard. The lease's ONE
   runner must also implement the fixed-ID starter; a separate starter can no
   longer be supplied. Missing dependencies refuse before daemon effects.
5. `Run` reserves local/controller capacity, binds the actual created ID before
   start, validates actual container policy and verifies the finite byte protocol.
6. Successful exit, joined stream, exact owned cleanup, repeated absence checks,
   both journals and current final release permission precede returned bytes.
7. The caller closes and joins its owned authority transport on every path. Success
   neither refunds persistent object storage nor updates task acceptance.

## Pinned controller session

`newControllerObjectReadSession` captures the original authenticated principal,
worker alias, complete retained worker binding, exact receipt/image/volume intent
and a non-nil `controllerObjectReadSessionEvidence`. Later caller contexts cannot
replace that principal. Candidate transitions must be the actual latest valid
checkpoint, including complete local history at acknowledgment. Truncated histories,
substituted requests and old begin events cannot become new authority.

Operation authorization and durable acknowledgment are separate. Authorization
checks current room/human/receipt policy and independently observed evidence, then
reloads the audit head; it grants no capacity by itself. Acknowledgment compares the
exact private local history with the controller candidate, performs independent
lifecycle checks and commits under the same audit revision fence. A stale revision
refuses; genuine controller persistence failure trips the global integrity latch.

Session revocation is scoped, nonblocking and permanent for reads and recovery.
Ordinary stream closure does not create a global integrity incident. A canceled
session may still record a `finish` with **absence false**, using a fresh bounded
call context and current human/creator authority. This preserves uncertainty only:
it cannot confirm absence, release capacity, export bytes or revive recovery.
An unavailable control connection cannot deliver even this metadata; local charges
remain retained. Fresh recovery must be independently authorized.

## Provisioned mutual authentication

`authenticateObjectReadStream` wraps a deployment-owned connection in TLS 1.2 or 1.3
before any authority frames are exchanged. Both sides require a certificate from
the provisioned trust roots, the independently approved peer DNS identity and exact
SHA-256 leaf-certificate fingerprint. The server requires a verified client
certificate. Only the `commonroom-object-read-v1` application protocol is accepted;
TLS versions below 1.2 are refused. TLS 1.2 permits only forward-secret ECDHE
AEAD suites (AES-GCM or ChaCha20-Poly1305); TLS 1.3 uses the standard library
suites. TLS resumption is disabled. Missing identity, untrusted or substituted certificates,
protocol mismatch, cancellation and a five-second handshake timeout close the raw
connection and return a sanitized refusal. Ownership of the connection transfers
on entry, including invalid-configuration paths.

Trusted deployment setup supplies a `tls.Certificate`, private CA roots, peer DNS
name and leaf fingerprint in `objectReadPeerIdentity` separately at each endpoint.
For example, the worker pins the controller certificate and controller DNS name;
the controller independently pins the worker certificate and worker DNS name.
These values must never come from model output, a request body or the peer itself.
The bound session assembly described below performs authentication before private
control traffic and retains the original human/worker/creator/intent binding.
Close and join it on every outcome. No listener or certificate issuance,
key loading, rotation service, automatic reconnect or fallback is installed here.
Provisioned rotation requires fresh approved fingerprints and a fresh session.
Keep private keys and connection handles outside model containers; a certificate
alone grants no artifact export or cleanup permission.

This mandatory security adapter has no optional feature flag or disable setting.
Production service wiring and real deployment authentication acceptance remain
outstanding. Synthetic ephemeral-certificate tests validate mutual identity, trust,
exact pins, DNS names, protocol/version refusal, cancellation, stalled peers and
closed failed connections without contacting a provider or external service.

## Loading provisioned identities from private files

`loadObjectReadPeerIdentity` loads an independent identity snapshot from three
operator-selected absolute file paths: certificate chain, matching unencrypted
private key and trusted peer CA bundle. Supply a separately approved peer DNS
name and exact64-character lowercase SHA-256 leaf fingerprint. Pass `server=true`
for the controller/server certificate role or `false` for worker/client. Put the
returned identity in that endpoint’s independently provisioned peer binding.
This internal loader introduces no environment variable, HTTP route or automatic
startup configuration; trusted service assembly must explicitly call it.

Keep all three files outside repositories and agent/container mounts, under
operator-controlled parent directories. Paths must be absolute, clean and distinct.
The existing private-file boundary validates each opened inode: current process
owner, owner-only permissions (for example0600 or0400), regular, single link,
no-follow/nonblocking open and bounded reads. Symlinks, hardlinks, directories,
public permissions and oversized inputs refuse. Parent directory control remains
a deployment prerequisite; the final-inode checks do not secure hostile ancestors.

Certificate and key files are limited to64KiB each; the CA bundle to256KiB.
Certificate chains allow at most8 PEM certificates and trust bundles at most16.
Only exact certificate PEM blocks and one unencrypted PKCS#8/PKCS#1/SEC1 key block
are accepted, with no extra headers, embedded skipped blocks or trailing junk.
Certificate/key matching is checked before use. The local leaf must be current,
not a CA and usable for the requested endpoint role under standard EKU semantics
(no EKU restrictions or any-usage are permitted). Trust entries must be current
signing CAs, with no duplicates. Wrong roles, expired/not-yet-valid identities,
empty/zero/malformed peer pins and wildcard/whitespace peer names refuse.

Missing or malformed certificate PEM markers return this actionable error:
`invalid certificate format: expected PEM with BEGIN CERTIFICATE and END CERTIFICATE markers`.
The certificate must contain the exact `-----BEGIN CERTIFICATE-----` and
`-----END CERTIFICATE-----` lines around its base64 content. Bare base64, binary
DER, empty files and other PEM block types are refused rather than repaired.
Certificate format is checked before reading the key or CA files, so a missing
key does not mask this certificate-format error. Other failures keep the generic
sanitized authentication refusal. Every failure returns no partial identity;
paths, certificate bodies and key material are never included in the error/audit.
Raw key buffers are cleared after parsing, but Go memory copies and the live
private-key object prevent any guaranteed memory-zeroization claim. The loader
neither generates certificates nor persists or exports keys. Actual peer chain,
name and exact pin verification remain mandatory at every TLS handshake; loading
files is not proof of peer identity, physical storage evidence or export permission.

Rotate through independently approved replacement files and fingerprints, then
load a fresh snapshot for a fresh session. Existing sessions retain their original
identity; editing files does not replace trust in a running session. A malformed
replacement refuses on reload. Revoke/close existing authority explicitly when
rotation must terminate an active session. No automatic reload, issuance, peer
registry or expiry-monitoring service is installed, and there is no security
bypass flag. Configure issuance, file custody, renewal and service ownership
before production use.

Tests generate temporary synthetic CAs/certificates/keys, perform mutual TLS with
loaded identities and run the complete isolated reader scenario matrix through
file loading, one-shot endpoint setup and the bound session. Failure coverage
includes unsafe files, key mismatch, malformed PEM/trust, duplicate/non-signing CA,
expiry, future validity and incompatible client/server roles. No real operator
key, provider credential or private data is used. See [private file safety](private-file-security.md).

## Bound authenticated session assembly

`newAuthenticatedObjectReadClient` and `serveAuthenticatedObjectReadSession` own
one supplied connection each. Trusted local provisioning associates its peer
certificate with an `objectReadPeerBinding`: worker alias, SHA-256 digest of the
complete retained worker record and immutable finite read intent. The worker
constructor compares its binding with the original record/intent before TLS; the
controller also compares it with the original human-pinned session and worker.
Never populate this provisioning structure from a received frame or model output.

After mutual TLS, both sides exchange an exact canonical `object-read-session-v1`
hello with version1, worker, record digest and intent, under the existing32-KiB
frame limit. A five-second bound (or earlier parent deadline) covers the binding
exchange. The server reads/compares first and replies only with its independently
expected binding; the client requires that exact reply. Wrong workers, stale
records, substituted intents, protocol/version/encoding errors or canceled setup
close the stream before the controller operation/evidence/audit callbacks run.
TLS encryption/pinning protects these frames; they are not a replacement for the
certificate identity or the controller’s current account/room/export permission.
After agreement, the existing `object-read-v1` sequence starts at1.

The worker client follows its original setup context even while idle. The server
follows both its serving context and the original pinned controller-session
lifetime. EOF, revocation, parent cancellation and any serving exit revoke scoped
authority. The worker owner must call `StopObjectReadAuthority` to join its read
pump; the controller owner must join its serving call before releasing the actual
creator lease. Closing aborts the underlying connection before TLS shutdown, so
close-notify cannot delay cancellation. Only full durable framed acknowledgments
settle lifecycle charges; EOF/connection closure never proves cleanup or absence.
There is no reconnection, automatic adoption, permissive witness or isolation flag.

This internal assembly accepts an independently provisioned listener through the
one-shot endpoint API below, but installs no listener configuration, peer registry,
certificate provisioner, production service route or public artifact endpoint. It accepts no
model-selected network address or credentials. Deployment still must supply the
correct independent identities and real physical/creator/export evidence.

## One-shot provisioned connection endpoints

`connectObjectReadControl` adds worker-side connection setup using an independently
provisioned `objectReadControlEndpoint` and a deployment-owned `net.Dialer`. Its
address must be a literal unicast IPv4/IPv6 address with a nonzero port, such as
`192.0.2.10:443` or `[2001:db8::10]:443` (documentation examples, not configured
services). Hostnames, URLs, unspecified/multicast addresses, IPv6 zone selectors
and IPv4-mapped IPv6 are refused before dialing. Peer DNS identity in the TLS
certificate remains required independently of this literal network address.
The endpoint and peer binding must come from trusted operator provisioning;
never copy an address from model output, room content or a request parameter.

The dialer receives exactly one `tcp` call to that address with a five-second
maximum context. Invalid identity, record/intent binding or canceled setup refuses
before dialing. Dial errors, returned connections accompanying an error, and
connections returned after cancellation are closed and sanitized. There is no
DNS discovery, redirect, alternate address, retry or reconnect. After dialing,
mutual TLS and exact session binding are still required; the short dial context
is not used as the successful session’s lifetime. Original worker setup authority
continues to govern the returned client while idle.

`serveObjectReadControlOnce` consumes an already provisioned listener and the
original pinned controller session. The listener must implement the normal
`net.TCPListener` contract: concurrent Close unblocks Accept. Its address/firewall
and private-key custody remain deployment responsibilities. Session identity and
binding are checked before Accept. One connection is accepted under a five-second
maximum wait; parent cancellation or pinned-session revocation closes the listener.
The listener closes before TLS begins. A listener-close error refuses the session;
operators must investigate an uncertain listener rather than open a replacement
automatically. Every accepted connection is owned until the serving call returns.
Every exit stops scoped controller authority. Callers must join the serving call
before relinquishing the actual creator lease.

For deployment assembly, provision a listener once for the exact pinned session,
run `serveObjectReadControlOnce`, and supply its approved literal address and
independently pinned peer binding to `connectObjectReadControl`. The latter returns
the same runtime guard used by `newWorkerObjectReadRuntime`. No certificate, worker
or address is learned from the accepted socket. This API does not install an HTTP
route, manage firewall rules, issue certificates, load keys, register peers or start
a network service automatically. It adds no feature flag or host/SSH execution
fallback: authentication and mandatory container isolation cannot be disabled.

Tests mock both dialer and listener with in-memory private connections, including
exact endpoint/one-attempt behavior, invalid address/identity refusal, dial error
and cancellation cleanup, canceled waiting Accept and failed listener closure.
The complete isolated reader runtime scenarios also use this setup path before
TLS/session agreement. No test dials a documentation IP or external network,
launches Docker/providers or connects to a real database.

## Private control protocol

The transport carries metadata only, never artifact bytes, provider messages,
credentials, model commands or host paths to execute. It requires an already
authenticated supervisor-owned `io.ReadWriteCloser`; it is not itself a handshake,
listener, credential store, public HTTP endpoint or remote worker deployment.

Frames use a four-byte big-endian length followed by canonical JSON, bounded to
32 KiB. Requests and responses carry `protocol: "object-read-v1"`, version1 and a
strictly increasing sequence. Unknown/duplicate/noncanonical fields, cross-protocol
frames, wrong versions/sequences, substituted intents/events/histories and unexpected
operations refuse. Only `authorize` and `acknowledge` operations exist. The session
allows at most64 exchanges over two minutes; each active client exchange is bounded
to five seconds. Calls are serialized; queued calls do not gain permission from an
older acknowledgment. Complete finite histories are independently validated.

Denied, lost, malformed or canceled exchanges poison and close the one-shot link
and revoke scoped authority. A single bounded read pump also observes idle EOF
without competing with frame parsing. The assembled runtime follows both the
control authority lifetime and actual creator lifetime: losing either cancels
active I/O and withholds bytes, even between protocol exchanges. Stop callbacks must be local, nonblocking and must not
reenter the link. The controller guard must not reenter server/worker admission
locks or call the same stream while the worker awaits its response. The client owns
its stream; invalid construction does not transfer that ownership. The caller must
join the serving goroutine rather than leave it behind after a failure.

## Private local and physical evidence

`newWorkspaceObjectReadEvidence` requires an active retained-workspace creator
lease, approved worker/runner/root/image/volume intent and non-nil
`objectReadSessionWitness`. It checks recorded job/task/execution binding, unchanged
retained worker metadata/cancellation latch and exact private local reader records.
Checks compare local state before and after callbacks/daemon inspection, with
independent witness validation repeated on both sides. Callback records are detached.

Operation checks admit only the next valid local transition or the exact latest
one after persistence. Lifecycle acknowledgment requires already-persisted matching
history. Local-only `confirmed`/`recovered` markers are stripped only for comparing
the controller history; they cannot authorize a new lifecycle. Malformed, changed,
foreign, missing or stale records refuse. No lease/admission lock is reacquired by
callbacks while the journal holds it.

Daemon observations are read-only:

- Begin requires the exact retained custody volume with no reader/job/volume/name
  references, including independent catalog checks.
- Bound requires the exact recorded ID in the fixed created-container policy,
  checked in two independent passes before it may start.
- Finish/recovery inspect only the separately recorded immutable ID and exact owned
  labels/image/volume references. Claimed absence must actually be observed. Owned
  policy drift can be inspected for cleanup but grants no read permission.
- `ObjectReadContainerPlan.InspectRecorded` never creates, starts, executes, removes,
  discovers or adopts a helper. Unknown IDs cannot be substituted by name.

Inspection is not physical quota enforcement or authentication. The independent
witness must still attest actual creator ownership, approved daemon/worker/image,
writer exclusion, custody, current operation-specific export/cleanup rights, real
physical quotas and account admission. No permissive implementation is provided.
Evidence lifetime follows the actual creator lease; it cannot be reused afterward.

## Cleanup-only worker assembly

`recoverRetainedWorkerObjectRead` binds a fresh authority lifetime, actual retained
creator and the existing private known-ID journal to the cleanup-only runtime.
It receives no starter, reader, export guard, volume-deletion or storage-refund
capability. Present and already-absent recorded helpers can settle under repeated
independent evidence. A failed cleanup retains the same recovery checkpoint and
capacity; a newly authenticated cleanup session can retry that exact known ID.
Wrong/unknown IDs, denied authority and changed ownership refuse before effects.
A worker integrity-stop marker still requires independent incident reconciliation;
this helper cannot clear it or turn a cleanup session into new read permission.

Runtime and known-ID recovery now register their complete operation with retained
ownership before doing work. Creator teardown cancels the owner lifetime BEFORE
waiting for its state mutex: a journal can hold that mutex while a current-authority
callback needs cancellation to return. Checking the canceled lifetime also refuses
new admissions during that interval. Teardown then marks ownership inactive and
joins admitted operations before unlocking the actual creator lease.

The join includes cleanup and operation-authority revocation. Runtime failure also
revokes the separate byte-release guard and clears/withholds returned bytes. Even
if final durable release succeeded, a subsequent inactive-owner check refuses the
result without erasing its audit history or refunding storage. Successful helper
cleanup alone never overrides that refusal.

Trusted setup should still call runtime/recovery synchronously inside its creator
callback. Expired retained handles cannot revive work. Dependencies must cooperate
with cancellation and join their streams; arbitrary blocking callbacks cannot be
forcibly stopped by a Go context. Cancellation is not proof of physical absence.
Private control transport closure/join remains the caller's responsibility.
Synthetic regressions hold both read and recovery callbacks during cancellation,
then hold authority revocation, and verify the creator lock remains acquired until
both finish. These fixtures provide no live daemon/quota or custody proof.

## Flags and execution boundary

This is mandatory security infrastructure with no optional-feature flag, off switch,
per-connection exception or restart setting. Task-definition editing flags may be
off while current authorized retained reads and cleanup remain interpretable. They
never bypass authentication, audit, retention or export restrictions. The runtime
cannot fall back to host file reads, local/SSH model execution or a different daemon.
Source forks that replace guards, witnesses, starters or persistence invalidate
these guarantees. A starter must attach only the exact durably bound ID on the
same authenticated rootless daemon. `dockerCLIRunner.StartObjectRead` now provides
the fixed-ID `docker start --attach --interactive` adapter after per-call rootless
local-daemon validation, with reduced environment, sanitized errors/stderr, retained
final stdout through Wait and joined process ownership on Close. It shares the
existing fixed-helper process machinery; it cannot select a command from a request.
Its deployment and reviewed-image behavior still require live verification.

## Verification and remaining integration

Tests exercise pinned identities, current policy/revocation, complete history,
independent operation/acknowledgment checks, canceled uncertainty without release,
failed audit, private framing/canonical sizes/sequences/protocol isolation, poison/
cancellation/closure, frame budget and detached callback data. Physical tests cover
pre/post stages, exact IDs, false absence, writable mounts, volume replacement,
extra references/catalog entries, changed/corrupt local history, dead creator
lifetimes and independent witness denial/revocation. Observer tests verify no
lifecycle effects or byte-release callbacks occur.

An assembled synthetic journey connects the actual local journal, private control
protocol, controller session/account ledger, local/physical evidence and fixed
reader runtime with current controller byte-release policy. It verifies actual
finite protocol bytes and withheld data after cleanup/exit/Close/export/audit
failures. Recovery integration covers present/already-absent IDs, failed cleanup
with fresh-session retry, denied/wrong/unknown IDs, foreign observations and editing
flags off, without exporting bytes or refunding storage. Idle control EOF and creator
cancellation withhold active reads. Synthetic subprocess tests cover final buffered
stdout, sanitized failure and cancel/Close joining for the fixed reader CLI stream.
The assembled runtime scenarios now use mutual TLS and the exact session-binding
exchange before journal/reader traffic. Additional session tests verify worker,
record, intent and certificate substitution, controller-side binding refusal,
permission denial despite authenticated TLS, pinned human audit identity and
idle parent/controller/link cancellation without new audit entries.
All DB and Docker boundaries are mocked; channel helpers and filesystem fixtures
are synthetic/temporary, with no real provider, credential or private user data.

Still required: deployment-authenticated transport setup, real physical quota/
custody/export/creator witnesses, reviewed-image and same-daemon starter deployment
verification, production service assembly, independent object-only ownership after
workspace disposal, unknown-ID/
stop-marker incident reconciliation, public delivery/human review wiring and live
rootless acceptance. Required file-task dispatch remains blocked. These synthetic
checks are not live P0 gate observations and certify no production readiness.

See [controller ledger](controller-object-read-ledger.md),
[worker journal](worker-object-read-journal.md),
[reader supervision](object-reader-supervision.md),
[release sessions](controller-artifact-read-session.md) and
[mandatory isolation](architecture.md#mandatory-container-isolation-design-contract).

For private operator policy, offline validation, configured listener/connection setup and joined service shutdown, see [reader worker profiles](reader-worker-profiles.md).
