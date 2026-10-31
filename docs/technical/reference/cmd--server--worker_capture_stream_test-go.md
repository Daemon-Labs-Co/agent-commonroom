# cmd/server/worker_capture_stream_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_capture_stream_test.go)

Automated regression checks for worker_capture_stream; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestCaptureStreamSyntheticChild`** — Regression contract: TestCaptureStreamSyntheticChild.
  Signature: `func(t *testing.T)`.
  Calls: `io.Copy`, `os.Exit`, `os.Getenv`, `os.Stderr.WriteString`, `os.Stdout.WriteString`.
- **`captureSyntheticFactory`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(mode string) func(context.Context) *exec.Cmd`.
  Calls: `exec.CommandContext`.
- **`TestCaptureStreamSuccessfulWaitPreservesBufferedAck`** — Regression contract: TestCaptureStreamSuccessfulWaitPreservesBufferedAck.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `captureSyntheticFactory`, `context.WithTimeout`, `io.ReadAll`, `openCaptureCommand`, `stream.Close`, `stream.Wait`, `string`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestCaptureStreamFailureIsSanitized`** — Regression contract: TestCaptureStreamFailureIsSanitized.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `captureSyntheticFactory`, `context.WithTimeout`, `err.Error`, `errors.Is`, `io.ReadAll`, `len`, `openCaptureCommand`, `stream.Close`, `stream.Wait`, `string`, `strings.Contains`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestCaptureStreamDuplexCancelAndCloseReaps`** — Regression contract: TestCaptureStreamDuplexCancelAndCloseReaps.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `captureSyntheticFactory`, `context.WithCancel`, `errors.Is`, `io.ReadFull`, `len`, `make`, `openCaptureCommand`, `stream.Close`, `stream.Wait`, `stream.Write`, `string`, `t.Context`, `t.Fatal`, `t.Parallel`, `time.After`.
- **`TestCaptureStartPolicyAndInvalidInput`** — Regression contract: TestCaptureStartPolicyAndInvalidInput.
  Signature: `func(t *testing.T)`.
  Calls: `(dockerCLIRunner{}).StartCapture`, `cancel`, `captureStartCommand`, `context.WithCancel`, `exec.CommandContext`, `len`, `openCaptureCommand`, `slices.Equal`, `strings.HasPrefix`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`.
