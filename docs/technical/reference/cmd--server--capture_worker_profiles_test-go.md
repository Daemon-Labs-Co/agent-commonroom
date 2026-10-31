# cmd/server/capture_worker_profiles_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/capture_worker_profiles_test.go)

Automated regression checks for capture_worker_profiles; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`captureProfileFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, g *controllerCaptureSession) (string, captureWorkerProfileFile)`.
  Calls: `filepath.Join`, `provisionedObjectReadTestFiles`, `t.Helper`, `t.TempDir`, `writeCaptureProfiles`.
- **`writeCaptureProfiles`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, path string, file captureWorkerProfileFile)`.
  Calls: `json.MarshalIndent`, `os.WriteFile`, `t.Fatal`, `t.Helper`.
- **`TestCaptureWorkerServiceLifecycle`** — Regression contract: TestCaptureWorkerServiceLifecycle.
  Signature: `func(t *testing.T)`.
  Calls: `a.Close`, `artifactState`, `b.Close`, `c.AcknowledgeCapture`, `c.AuthorizeCapture`, `c.CaptureAuthorityLifetime`, `c.CaptureAuthorityLifetime().Err`, `c.StopCaptureAuthority`, `captureProfileFixture`, `captureSessionFixture`, `close`, `controllerCaptureKey`, `db.snapshot`, `g.lifetime.Err`, `len`, `make`, `net.Pipe`, `newCaptureWorkerService`, `newSyntheticReadListener`, `service.Close`, `service.Connect`, `service.Serve`, `syntheticReadDialer`, `t.Fatal`, `t.Parallel`, `time.After`.
- **`TestCaptureWorkerProfilesRefuseInvalidPolicy`** — Regression contract: TestCaptureWorkerProfilesRefuseInvalidPolicy.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `captureProfileFixture`, `captureSessionFixture`, `filepath.Join`, `newCaptureWorkerService`, `os.Chmod`, `os.ReadFile`, `os.WriteFile`, `service.Close`, `service.Connect`, `string`, `strings.Replace`, `syntheticReadDialer`, `t.Error`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`, `writeCaptureProfiles`.
- **`TestCaptureWorkerServiceRejectsWrongListenerAndJoinsAccept`** — Regression contract: TestCaptureWorkerServiceRejectsWrongListenerAndJoinsAccept.
  Signature: `func(t *testing.T)`.
  Calls: `captureProfileFixture`, `captureSessionFixture`, `listener.accepts.Load`, `make`, `newCaptureWorkerService`, `newSyntheticReadListener`, `service.Close`, `service.Serve`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.After`.
- **`TestCaptureWorkerServiceShutdownDuringDial`** — Regression contract: TestCaptureWorkerServiceShutdownDuringDial.
  Signature: `func(t *testing.T)`.
  Calls: `c.Done`, `c.Err`, `captureProfileFixture`, `captureSessionFixture`, `close`, `make`, `newCaptureWorkerService`, `service.Close`, `service.Connect`, `syntheticReadDialer`, `t.Fatal`, `t.Parallel`, `time.After`.
- **`TestCaptureWorkerServiceListenerFactoryFailureAndCancellation`** — Regression contract: TestCaptureWorkerServiceListenerFactoryFailureAndCancellation.
  Signature: `func(t *testing.T)`.
  Calls: `c.Deadline`, `c.Done`, `c.Err`, `cancel`, `captureProfileFixture`, `captureSessionFixture`, `close`, `context.WithCancel`, `g.lifetime.Err`, `listener.accepts.Load`, `make`, `newCaptureWorkerService`, `newSyntheticReadListener`, `service.Close`, `service.ListenAndServe`, `syntheticReadListenerFactory`, `t.Error`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.After`.
- **`TestCaptureWorkerServiceJoinsRefusedSessionRevocation`** — Regression contract: TestCaptureWorkerServiceJoinsRefusedSessionRevocation.
  Signature: `func(t *testing.T)`.
  Calls: `captureProfileFixture`, `captureSessionFixture`, `close`, `g.lifetime.Done`, `g.server.mu.Lock`, `g.server.mu.Unlock`, `listener.accepts.Load`, `make`, `newCaptureWorkerService`, `newSyntheticReadListener`, `service.Close`, `service.Serve`, `t.Fatal`, `t.Parallel`, `time.After`.
