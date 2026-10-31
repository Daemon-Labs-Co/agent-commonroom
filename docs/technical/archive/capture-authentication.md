# Authenticated private capture sessions

Retained detailed reference. Start with [user guides](../../user/README.md) for illustrated
steps or [technical contracts](../README.md) for concise current interfaces. Historical
observations remain historical; documented security requirements still apply.


Capture lifecycle authority can now run over a mutually authenticated private TLS
connection pinned to one worker, retained creator record and finite capture intent.
This joins the private capture RPC to certificate identity and the existing
controller session. It is internal mandatory security infrastructure, not a public
capture API, provider connection or enabled file-delivery feature.

## Trusted configuration and use

Operator-controlled setup must independently provision both certificate identities
and the approved worker/record/intent bindings on both peers. Each identity contains
its local certificate/private key, trusted issuing roots, expected peer DNS name
and exact approved peer leaf SHA-256 pin. Sharing an issuing CA does not authorize
another worker's leaf certificate. Keep keys and the control connection outside
model/tool containers. Certificate rotation requires a newly approved identity and
fresh session, not changing an active connection's trust.

The existing private PEM identity loader supplies compatible certificate material;
its [identity provisioning rules](reader-authority-bridge.md) still apply. Certificate
files require explicit BEGIN CERTIFICATE and END CERTIFICATE PEM markers. This
assembly adds no credential storage, key generation, file-upload or model-selected
identity endpoint. Application/bootstrap code must supply trusted configuration;
no capture configuration file or network listener is installed automatically.

For each operation:

1. Authenticate the original human independently, acquire retained creator ownership
   and construct `controllerCaptureSession` with the exact approved worker/record/
   intent and mandatory current physical/account evidence.
2. Construct independent `capturePeerBinding` values for worker and controller.
   Their worker alias, complete retained-record digest and complete finite intent
   must agree. Never populate these expected values from received frames.
3. The controller calls `serveAuthenticatedCaptureSession(ctx, rawConn, binding,
   originalSession)`. The worker calls `newAuthenticatedCaptureClient(ctx, rawConn,
   binding, retainedRecord, intent, localStop)` on its exclusively owned connection.
4. Pass that client to `newOwnedWorkerCapture` or, with independently renewed cleanup
   permission, `recoverRetainedWorkerCapture`. Normal journal authorization and
   durable controller acknowledgments still precede helper effects.
5. Stop the worker client and join the controller serving call on EVERY path.
   Stop callbacks must be local/nonblocking and cannot reenter the stream or join
   their own read pump. The original setup context remains authoritative while idle.

The raw connection is consumed on every outcome, including invalid binding or
handshake failure. Invalid setup cannot leave a plaintext connection available
for fallback. The raw-stream functions neither discover peers nor choose endpoints. The fixed
endpoint adapters below consume their listener/connection and leave deployment
policy provisioning to trusted setup. Trusted setup must
not concurrently mutate inputs during construction; request member lists are
copied into the session binding before use.

The [private capture worker service](capture-worker-profiles.md) can load a
strict capture-only policy and own these sessions through joined shutdown.

## Fixed endpoint transport assembly

Trusted deployment code can supply a `captureControlEndpoint` containing an
approved literal IP:port and the independently constructed capture binding.
`connectCaptureControl` uses an injected deployment `net.Dialer` to dial exactly
that TCP endpoint once, with a five-second setup bound. Hostnames, URLs, missing
or zero ports, unspecified/multicast addresses, scoped IPv6 and IPv4-mapped IPv6
are refused before dialing. This is not DNS discovery or a model-selected URL.
Loopback and private addresses remain operator choices, protected by peer pins.

The controller can pass an exclusively provisioned listener to
`serveCaptureControlOnce`. It consumes that listener on every outcome, accepts at
most one connection, and closes the listener before authenticating the peer.
Original controller-session cancellation or a five-second accept deadline closes
the listener to unblock Accept. A failed listener close refuses and closes any
accepted connection before authority callbacks. Deployment must ensure Close
unblocks Accept and that its listener is actually bound to the approved interface;
this low-level adapter neither creates nor verifies a deployment listener address.

Dial failure, cancellation, nil results or an uncertain returned connection close
that connection and refuse without reconnecting. The dial deadline is not the
retained session lifetime: a successful client follows the original caller context.
No raw connection is returned for plaintext fallback. The shared reader dialer
interface/address validator carry only transport shape, never read authority or
protocol. Source-injected dialers/listeners must honor cancellation; replacing them
with dishonest adapters invalidates network confinement. No public route, policy
file, automatic listener installation or physical witness is introduced.

## Cryptographic and operation binding

Both peers require verified certificate chains, expected DNS identity and exact
leaf pins. TLS 1.2 and TLS 1.3 are supported. TLS 1.2 uses only forward-secret AEAD
cipher suites; session tickets are disabled. The shared cryptographic implementation
accepts only its two fixed internal protocol wrappers. Capture negotiates
`commonroom-capture-control-v1`; object readers negotiate their DIFFERENT fixed
ALPN. Neither service accepts the other's protocol, even with valid certificates.
There is no TLS/plaintext downgrade or alternate-protocol fallback.

After TLS, a bounded canonical `capture-session-v1` hello must agree on exact worker,
record digest and complete intent, including request members, challenge, reviewed
image and retained volume identities. Certificate identity alone cannot select a
new operation. TLS and hello setup have bounded deadlines and honor cancellation.
Unknown/duplicate fields, foreign protocols and mismatched hello values refuse
before lifecycle authority or audit callbacks. Subsequent RPC frames retain the
[private authority transport](capture-authority-transport.md) contract.

The controller uses the ORIGINAL authenticated human/worker session. A worker's
frame or context cannot replace that principal. TLS never supplies physical quota,
creator ownership, current task permission or artifact export rights. A denied
current controller guard remains denied after successful TLS authentication.

## Cancellation, flags and retention

Original worker-context cancellation closes the connection and joins its watched
read pump. Server return or loss of the pinned controller lifetime revokes that
controller session. Idle EOF cancels the client's continuous capture authority;
fresh contexts cannot revive it. Raw transport closes before TLS shutdown so
close-notify cannot indefinitely delay joining. Complete RPC acknowledgments,
never EOF, establish durable transitions.

A failed or lost acknowledgment retains uncertainty, local/controller history and
capacity. Connection loss is not proof of physical cleanup, successful capture,
complete artifacts or user acceptance. Recovery needs independently renewed
identity/current cleanup authority and an already recorded exact helper ID. There
is no automatic reconnect, adoption, deletion or refund.

Authentication and isolation have no optional switch or paid/owner exception.
Existing server task flags and current human/account permissions remain enforced;
switching a product feature off cannot erase history or waive cleanup/retention.
Required file/bundle task dispatch remains blocked pending complete production
physical custody/quota/source/export assembly.

## Verification and deployment gaps

Tests generate temporary synthetic certificates and use in-memory private pipes,
mocked database/controller/daemon/physical guards and the actual TLS, hello, RPC,
worker journal and owned capture runtime. They cover successful lifecycle/audit
binding, current permission denial, editing-enabled/disabled assembly, wrong worker,
record, intent, pins/trust, foreign hello/reader ALPN, original-context idle
cancellation and TLS 1.2/1.3 in both roles. Reader protocol separation is also checked
against the new capture ALPN. Endpoint tests cover one-shot lifecycle/audit,
invalid addresses/identities/bindings before dial, uncertain dial cleanup, accept
cancellation and failed listener closure. These tests launch no real provider or
Docker daemon.

Production must still provision certificates, approved endpoints/firewall policy,
rootless reviewed images, independent creator/account/inventory and physical
quota/source/custody evidence, and real lifecycle/bootstrap wiring. The synthetic
capture helper protocol in assembly tests reports synthetic receipts; it does not
prove actual persisted artifacts. Public output attachment/delivery/review, source
container disposal and live Linux/provider acceptance remain incomplete. A source
fork substituting identities, callbacks or TLS policy invalidates the boundary.
