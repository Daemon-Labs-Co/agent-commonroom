# cmd/server/worker_object_read_stream_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_object_read_stream_test.go)

Automated regression checks for worker_object_read_stream; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestObjectReadStreamWaitAndSanitizedExit`** — Regression contract: TestObjectReadStreamWaitAndSanitizedExit.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `captureSyntheticFactory`, `context.WithTimeout`, `err.Error`, `errors.Is`, `io.ReadAll`, `openObjectReadCommand`, `stream.Close`, `stream.Wait`, `string`, `strings.Contains`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestObjectReadStreamCancelCloseAndRefusals`** — Regression contract: TestObjectReadStreamCancelCloseAndRefusals.
  Signature: `func(t *testing.T)`.
  Calls: `(dockerCLIRunner{}).StartObjectRead`, `cancel`, `captureSyntheticFactory`, `context.WithCancel`, `errors.Is`, `io.ReadFull`, `len`, `make`, `openObjectReadCommand`, `stream.Close`, `stream.Wait`, `stream.Write`, `string`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `time.After`.
