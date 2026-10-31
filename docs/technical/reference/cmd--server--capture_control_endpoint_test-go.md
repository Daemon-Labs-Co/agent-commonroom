# cmd/server/capture_control_endpoint_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/capture_control_endpoint_test.go)

Automated regression checks for capture_control_endpoint; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestCaptureControlEndpointLifecycle`** — Regression contract: TestCaptureControlEndpointLifecycle.
  Signature: `func(t *testing.T)`.
  Calls: `a.Close`, `append`, `b.Close`, `boundCapturePeers`, `c.AcknowledgeCapture`, `c.AuthorizeCapture`, `c.Deadline`, `c.StopCaptureAuthority`, `calls.Add`, `calls.Load`, `captureSessionFixture`, `connectCaptureControl`, `db.snapshot`, `g.lifetime.Err`, `len`, `listener.accepts.Load`, `make`, `net.Pipe`, `newSyntheticReadListener`, `serveCaptureControlOnce`, `strings.Repeat`, `syntheticReadDialer`, `t.Error`, `t.Fatal`, `t.Parallel`, `time.Until`.
- **`TestCaptureControlEndpointRejectsBeforeDial`** — Regression contract: TestCaptureControlEndpointRejectsBeforeDial.
  Signature: `func(t *testing.T)`.
  Calls: `boundCapturePeers`, `captureSessionFixture`, `connectCaptureControl`, `db.snapshot`, `errors.New`, `len`, `syntheticReadDialer`, `t.Error`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestCaptureControlEndpointInvalidIdentityAndDialFailure`** — Regression contract: TestCaptureControlEndpointInvalidIdentityAndDialFailure.
  Signature: `func(t *testing.T)`.
  Calls: `a.Close`, `a.Write`, `b.Close`, `boundCapturePeers`, `cancel`, `captureSessionFixture`, `connectCaptureControl`, `context.WithCancel`, `errors.New`, `net.Pipe`, `strings.Repeat`, `syntheticReadDialer`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestCaptureControlListenerCancellation`** — Regression contract: TestCaptureControlListenerCancellation.
  Signature: `func(t *testing.T)`.
  Calls: `boundCapturePeers`, `cancel`, `captureSessionFixture`, `context.WithCancel`, `db.snapshot`, `g.StopCaptureAuthority`, `g.lifetime.Err`, `int32`, `len`, `listener.accepts.Load`, `make`, `newSyntheticReadListener`, `serveCaptureControlOnce`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.After`.
- **`TestCaptureControlListenerCloseFailureRefusesSession`** — Regression contract: TestCaptureControlListenerCloseFailureRefusesSession.
  Signature: `func(t *testing.T)`.
  Calls: `a.Close`, `b.Close`, `boundCapturePeers`, `c.StopCaptureAuthority`, `captureSessionFixture`, `connectCaptureControl`, `db.snapshot`, `g.lifetime.Err`, `len`, `make`, `net.Pipe`, `newSyntheticReadListener`, `serveCaptureControlOnce`, `syntheticReadDialer`, `t.Fatal`, `t.Parallel`.
- **`TestCaptureControlAuthenticationFailureNeverAcceptsReplacement`** — Regression contract: TestCaptureControlAuthenticationFailureNeverAcceptsReplacement.
  Signature: `func(t *testing.T)`.
  Calls: `boundCapturePeers`, `captureSessionFixture`, `connectCaptureControl`, `db.snapshot`, `dials.Add`, `dials.Load`, `g.lifetime.Err`, `left.Close`, `len`, `listener.accepts.Load`, `make`, `net.Pipe`, `newSyntheticReadListener`, `right.Close`, `serveCaptureControlOnce`, `syntheticReadDialer`, `t.Fatal`, `t.Parallel`.
