# internal/workspacecapture/channel_test.go

[Module index](internal.md) · [Open source](../../../internal/workspacecapture/channel_test.go)

Automated regression checks for channel; read test names for individual contracts.

Package: `workspacecapture`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `channelResult`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { report Report err error }`.
- **`TestCaptureChannelActualCustodyTransfer`** — Regression contract: TestCaptureChannelActualCustodyTransfer.
  Signature: `func(t *testing.T)`.
  Calls: `NewCaptureChannel`, `ServeCaptureChannel`, `captureFixture`, `captureToCustody`, `client.Close`, `client.Finish`, `filepath.Join`, `len`, `make`, `net.Pipe`, `os.ReadFile`, `stopped.Add`, `stopped.Load`, `string`, `t.Context`, `t.Fatal`, `t.Parallel`, `transferRoot`.
- **`TestCaptureChannelDenialAndUncertainCustody`** — Regression contract: TestCaptureChannelDenialAndUncertainCustody.
  Signature: `func(t *testing.T)`.
  Calls: `NewCaptureChannel`, `ServeCaptureChannel`, `captureFixture`, `captureToCustody`, `client.Close`, `errors.Is`, `errors.New`, `len`, `make`, `net.Pipe`, `os.ReadDir`, `stopped.Add`, `stopped.Load`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `transferRoot`.
- **`TestCaptureChannelRejectsMalformedFramesBeforeAuthority`** — Regression contract: TestCaptureChannelRejectsMalformedFramesBeforeAuthority.
  Signature: `func(t *testing.T)`.
  Calls: `RequestDigest`, `ServeCaptureChannel`, `append`, `binary.BigEndian.PutUint32`, `captureFixture`, `controlWrite`, `left.Close`, `left.Write`, `len`, `make`, `net.Pipe`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `uint32`, `wire.Bytes`, `wire.Reset`, `wire.Write`.
- **`TestCaptureChannelCancellationAndWrongAcknowledgment`** — Regression contract: TestCaptureChannelCancellationAndWrongAcknowledgment.
  Signature: `func(t *testing.T)`.
  Calls: `NewCaptureChannel`, `cancel`, `captureFixture`, `client.AuthorizeFrozenSource`, `client.Close`, `close`, `context.WithCancel`, `controlRead`, `controlWrite`, `make`, `net.Pipe`, `right.Close`, `stopped.Add`, `stopped.Load`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.After`.
- **`TestCaptureChannelRejectsInvalidLifecycleTransitions`** — Regression contract: TestCaptureChannelRejectsInvalidLifecycleTransitions.
  Signature: `func(t *testing.T)`.
  Calls: `RequestDigest`, `ServeCaptureChannel`, `accept`, `append`, `captureFixture`, `controlRead`, `controlWrite`, `left.Close`, `make`, `net.Pipe`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Helper`, `t.Parallel`, `t.Run`, `uint64`.
- **`TestCaptureChannelBoundsObservationMessages`** — Regression contract: TestCaptureChannelBoundsObservationMessages.
  Signature: `func(t *testing.T)`.
  Calls: `RequestDigest`, `ServeCaptureChannel`, `captureFixture`, `controlRead`, `controlWrite`, `left.Close`, `len`, `make`, `net.Pipe`, `t.Context`, `t.Fatal`, `t.Parallel`, `uint64`.
- **Type `supervisedCaptureTestStream`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { net.Conn entered chan struct{} reading atomic.Bool closed atomic.Bool }`.
- **`(*supervisedCaptureTestStream).Read`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(raw []byte) (int, error)`.
  Calls: `close`, `s.Conn.Read`, `s.reading.Swap`.
- **`(*supervisedCaptureTestStream).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
  Calls: `s.Conn.Close`, `s.closed.Store`.
- **`TestSupervisedCaptureChannelStillClosesOnFailure`** — Regression contract: TestSupervisedCaptureChannelStillClosesOnFailure.
  Signature: `func(t *testing.T)`.
  Calls: `RequestDigest`, `ServeSupervisedCaptureChannel`, `cancel`, `captureFixture`, `context.WithCancel`, `controlWrite`, `left.Close`, `make`, `net.Pipe`, `stream.closed.Load`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.After`.
