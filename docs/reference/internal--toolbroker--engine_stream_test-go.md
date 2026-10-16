# internal/toolbroker/engine_stream_test.go

[Module index](internal.md) · [Open source](../../internal/toolbroker/engine_stream_test.go)

Automated regression checks for engine_stream; read test names for individual contracts.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `streamEngineFake`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { *engineDockerFake opens int stream io.ReadWriteCloser openError error onOpen func() }`.
- **`(*streamEngineFake).OpenEngine`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, id string) (io.ReadWriteCloser, error)`.
  Calls: `ctx.Err`, `f.onOpen`.
- **`streamFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*EngineContainerRuntime, *streamEngineFake, net.Conn)`.
  Calls: `client.Close`, `engineFixture`, `net.Pipe`, `peer.Close`, `r.Create`, `r.Start`, `t.Cleanup`, `t.Context`, `t.Fatal`, `t.Helper`.
- **`TestEngineStreamOneShotDuplexAndCancellation`** — Regression contract: TestEngineStreamOneShotDuplexAndCancellation.
  Signature: `func(t *testing.T)`.
  Calls: `errors.New`, `io.ReadFull`, `len`, `make`, `peer.Write`, `r.Attach`, `r.Close`, `stream.Close`, `stream.Read`, `stream.Write`, `streamFixture`, `string`, `t.Context`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `time.After`.
- **`TestEngineStreamRefusesPolicyDriftAndUncertainRetry`** — Regression contract: TestEngineStreamRefusesPolicyDriftAndUncertainRetry.
  Signature: `func(t *testing.T)`.
  Calls: `errors.New`, `r.Attach`, `stream.Close`, `streamFixture`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
