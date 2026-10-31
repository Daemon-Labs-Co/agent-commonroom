# cmd/server/object_read_authenticated_stream.go

[Module index](cmd.md) · [Open source](../../../cmd/server/object_read_authenticated_stream.go)

Go implementation for object read authenticated stream; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `objectReadPeerIdentity`**: objectReadPeerIdentity comes exclusively from trusted deployment provisioning, never from a worker request or model metadata. Roots authenticate the issuing authority; the exact leaf pin also separates workers sharing that authority. Certificate rotation requires a fresh one-shot session with a newly approved pin.
  Declaration: `struct { Certificate tls.Certificate Roots *x509.CertPool PeerDNSName string PeerSHA256 [sha256.Size]byte }`.
- **`validObjectReadPeerIdentity`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(identity objectReadPeerIdentity) bool`.
  Calls: `len`, `nilWorkerObservationDependency`.
- **`authenticateObjectReadStream`** — authenticateObjectReadStream owns raw on every path. It authenticates both ends before any authority frame is sent. It does not dial, discover identities, grant export rights, or replace the pinned human/creator/intent session. Production provisioning must keep keys and this stream outside all model containers. No plaintext, host-execution, or alternate-protocol fallback is permitted. A source fork replacing this constructor can bypass the boundary.
  Signature: `func(ctx context.Context, raw net.Conn, identity objectReadPeerIdentity, server bool) (*tls.Conn, error)`.
  Calls: `authenticatePinnedWorkerStream`.
- **`authenticatePinnedWorkerStream`** — Only fixed internal protocol wrappers may select these two disjoint ALPNs. This shares certificate/cipher policy, never operation or account authority.
  Signature: `func(ctx context.Context, raw net.Conn, identity objectReadPeerIdentity, server bool, protocol string) (*tls.Conn, error)`.
  Calls: `cancel`, `conn.HandshakeContext`, `context.WithTimeout`, `ctx.Err`, `identity.Roots.Clone`, `leaf.VerifyHostname`, `len`, `raw.Close`, `refuse`, `sha256.Sum256`, `tls.Client`, `tls.Server`, `validObjectReadPeerIdentity`.
