# cmd/server/object_read_authenticated_stream_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/object_read_authenticated_stream_test.go)

Automated regression checks for object_read_authenticated_stream; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`objectReadTestPeers`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (objectReadPeerIdentity, objectReadPeerIdentity)`.
  Calls: `big.NewInt`, `ed25519.GenerateKey`, `issue`, `now.Add`, `pem.EncodeToMemory`, `roots.AddCert`, `sha256.Sum256`, `t.Fatal`, `t.Helper`, `time.Now`, `tls.X509KeyPair`, `x509.CreateCertificate`, `x509.MarshalPKCS8PrivateKey`, `x509.NewCertPool`, `x509.ParseCertificate`.
- **`TestObjectReadAuthenticatedStream`** — Regression contract: TestObjectReadAuthenticatedStream.
  Signature: `func(t *testing.T)`.
  Calls: `a.Close`, `authenticateObjectReadStream`, `b.Close`, `cancel`, `conn.Write`, `context.Background`, `context.WithTimeout`, `make`, `net.Pipe`, `objectReadTestPeers`, `other.conn.Read`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `t.Run`, `tc.change`, `x509.NewCertPool`.
- **`TestObjectReadAuthenticatedStreamCanceledAndPlaintext`** — Regression contract: TestObjectReadAuthenticatedStreamCanceledAndPlaintext.
  Signature: `func(t *testing.T)`.
  Calls: `a.Write`, `authenticateObjectReadStream`, `b.Close`, `cancel`, `context.Background`, `context.WithTimeout`, `net.Pipe`, `objectReadTestPeers`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestObjectReadAuthenticatedStreamRejectsOtherProtocols`** — Regression contract: TestObjectReadAuthenticatedStreamRejectsOtherProtocols.
  Signature: `func(t *testing.T)`.
  Calls: `a.Close`, `authenticateObjectReadStream`, `b.Close`, `cancel`, `close`, `context.Background`, `context.WithTimeout`, `make`, `net.Pipe`, `objectReadTestPeers`, `peer.HandshakeContext`, `t.Fatal`, `t.Parallel`, `t.Run`, `tls.Server`.
- **`TestObjectReadAuthenticatedJournalLifecycle`** — Regression contract: TestObjectReadAuthenticatedJournalLifecycle.
  Signature: `func(t *testing.T)`.
  Calls: `a.Close`, `authenticateObjectReadStream`, `b.Close`, `c.StopObjectReadAuthority`, `cancel`, `context.WithTimeout`, `j.BeginObjectRead`, `j.BindObjectRead`, `j.FinishObjectRead`, `make`, `net.Pipe`, `newObjectReadAuthorityClient`, `objectReadTestPeers`, `objectReadTransportFixture`, `serveObjectReadAuthority`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `t.Run`, `workerObjectReadSlots`.
- **`TestObjectReadAuthenticatedStreamSupportedVersions`** — Force each supported version at the independently configured peer. Both adapter roles must accept mutual TLS 1.2 as well as TLS 1.3.
  Signature: `func(t *testing.T)`.
  Calls: `a.Close`, `authenticateObjectReadStream`, `b.Close`, `cancel`, `conn.ConnectionState`, `conn.HandshakeContext`, `context.WithTimeout`, `fmt.Sprintf`, `make`, `net.Pipe`, `objectReadTestPeers`, `t.Context`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `t.Run`, `tls.Client`, `tls.Server`.
