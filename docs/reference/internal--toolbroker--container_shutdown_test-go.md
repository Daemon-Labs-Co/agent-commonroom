# internal/toolbroker/container_shutdown_test.go

[Module index](internal.md) · [Open source](../../internal/toolbroker/container_shutdown_test.go)

Automated regression checks for container_shutdown; read test names for individual contracts.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestStopContainerProof`** — Regression contract: TestStopContainerProof.
  Signature: `func(t *testing.T)`.
  Calls: `StopContainer`, `cancel`, `context.Background`, `context.WithCancel`, `errors.New`, `io.WriteString`, `mockDocker`, `strings.Repeat`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `t.Run`.
