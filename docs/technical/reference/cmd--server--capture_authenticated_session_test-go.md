# cmd/server/capture_authenticated_session_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/capture_authenticated_session_test.go)

Automated regression checks for capture_authenticated_session; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`boundCapturePeers`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, g *controllerCaptureSession) (capturePeerBinding, capturePeerBinding)`.
  Calls: `copyCaptureIntent`, `objectReadTestPeers`, `probeWorkerRecordDigest`, `t.Helper`.
- **`TestAuthenticatedCaptureSessionLifecycleAndPermissions`** — Regression contract: TestAuthenticatedCaptureSessionLifecycleAndPermissions.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `artifactState`, `boundCapturePeers`, `c.AcknowledgeCapture`, `c.AuthorizeCapture`, `c.CaptureAuthorityLifetime`, `c.CaptureAuthorityLifetime().Err`, `c.StopCaptureAuthority`, `captureSessionFixture`, `controllerCaptureKey`, `copyCaptureIntent`, `db.snapshot`, `fmt.Sprint`, `g.CaptureAuthorityLifetime`, `g.CaptureAuthorityLifetime().Err`, `left.Close`, `len`, `make`, `net.Pipe`, `newAuthenticatedCaptureClient`, `right.Close`, `serveAuthenticatedCaptureSession`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestAuthenticatedCaptureRefusesForeignBindingBeforeAuthority`** — Regression contract: TestAuthenticatedCaptureRefusesForeignBindingBeforeAuthority.
  Signature: `func(t *testing.T)`.
  Calls: `boundCapturePeers`, `c.StopCaptureAuthority`, `cancel`, `captureSessionFixture`, `context.WithTimeout`, `db.snapshot`, `left.Close`, `len`, `make`, `net.Pipe`, `newAuthenticatedCaptureClient`, `right.Close`, `serveAuthenticatedCaptureSession`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`, `x509.NewCertPool`.
- **`TestAuthenticatedCaptureRefusesForeignHelloAndReaderProtocol`** — Regression contract: TestAuthenticatedCaptureRefusesForeignHelloAndReaderProtocol.
  Signature: `func(t *testing.T)`.
  Calls: `authenticatePinnedWorkerStream`, `boundCapturePeers`, `cancel`, `captureSessionFixture`, `client.hello`, `context.WithTimeout`, `db.snapshot`, `left.Close`, `len`, `make`, `net.Pipe`, `readCaptureAuthorityFrame`, `right.Close`, `serveAuthenticatedCaptureSession`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`, `writeCaptureAuthorityFrame`.
- **`TestAuthenticatedCaptureOriginalContextRevokesIdleSession`** — Regression contract: TestAuthenticatedCaptureOriginalContextRevokesIdleSession.
  Signature: `func(t *testing.T)`.
  Calls: `boundCapturePeers`, `c.AuthorizeCapture`, `c.CaptureAuthorityLifetime`, `c.CaptureAuthorityLifetime().Done`, `c.StopCaptureAuthority`, `cancel`, `captureSessionFixture`, `context.WithCancel`, `g.CaptureAuthorityLifetime`, `g.CaptureAuthorityLifetime().Err`, `left.Close`, `make`, `net.Pipe`, `newAuthenticatedCaptureClient`, `right.Close`, `serveAuthenticatedCaptureSession`, `t.Fatal`, `t.Parallel`, `time.After`.
- **`TestCaptureMutualTLSSupportedVersions`** — Regression contract: TestCaptureMutualTLSSupportedVersions.
  Signature: `func(t *testing.T)`.
  Calls: `authenticatePinnedWorkerStream`, `cancel`, `conn.ConnectionState`, `conn.HandshakeContext`, `context.WithTimeout`, `fmt.Sprintf`, `left.Close`, `make`, `net.Pipe`, `objectReadTestPeers`, `right.Close`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `tls.Client`, `tls.Server`.
- **`TestAuthenticatedCaptureOwnedRuntimeControllerBridge`** — Regression contract: TestAuthenticatedCaptureOwnedRuntimeControllerBridge.
  Signature: `func(t *testing.T)`.
  Calls: `artifactState`, `boundCapturePeers`, `c.StopCaptureAuthority`, `captureSessionFixture`, `close`, `controllerCaptureKey`, `copyCaptureRequest`, `db.seed`, `detachedProbeWorkerRecord`, `filepath.Join`, `fmt.Sprint`, `left.Close`, `len`, `make`, `net.Pipe`, `newAuthenticatedCaptureClient`, `newOwnedWorkerCapture`, `os.Chmod`, `os.WriteFile`, `right.Close`, `runtime.Run`, `serveAuthenticatedCaptureSession`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`, `withWorkerWorkspaceOwnership`, `workerCaptureSlots`, `writeWorkerRecord`.
