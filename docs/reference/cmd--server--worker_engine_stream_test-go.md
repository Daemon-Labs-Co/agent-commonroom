# cmd/server/worker_engine_stream_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/worker_engine_stream_test.go)

Automated regression checks for worker_engine_stream; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestEngineStreamChild`** — Regression contract: TestEngineStreamChild.
  Signature: `func(t *testing.T)`.
  Calls: `io.Copy`, `os.Exit`, `os.Getenv`, `os.Stderr.WriteString`, `os.Stdout.WriteString`.
- **`TestEngineCommandPreservesFinalFrameAfterProcessExit`** — Regression contract: TestEngineCommandPreservesFinalFrameAfterProcessExit.
  Signature: `func(t *testing.T)`.
  Calls: `exec.CommandContext`, `io.ReadAll`, `openEngineCommand`, `stream.Close`, `string`, `t.Context`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `time.After`.
- **`TestEngineCommandDuplexCloseReapsAndDiscardsStderr`** — Regression contract: TestEngineCommandDuplexCloseReapsAndDiscardsStderr.
  Signature: `func(t *testing.T)`.
  Calls: `exec.CommandContext`, `io.ReadFull`, `len`, `make`, `openEngineCommand`, `stream.Close`, `stream.Write`, `string`, `t.Context`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `time.After`.
- **`TestEngineCommandCancelledAndFailedStart`** — Regression contract: TestEngineCommandCancelledAndFailedStart.
  Signature: `func(t *testing.T)`.
  Calls: `(dockerCLIRunner{}).OpenEngine`, `cancel`, `context.WithCancel`, `exec.CommandContext`, `openEngineCommand`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`.
