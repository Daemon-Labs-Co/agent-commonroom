# cmd/server/isolation_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/isolation_test.go)

Automated regression checks for isolation; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestProviderAddressDenial`** — Regression contract: TestProviderAddressDenial.
  Signature: `func(t *testing.T)`.
  Calls: `net.ParseIP`, `publicProviderIP`, `t.Fatal`, `t.Parallel`.
- **`TestGatewayRejectsBeforeNetworkAccess`** — Regression contract: TestGatewayRejectsBeforeNetworkAccess.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `g.connect`, `httptest.NewRecorder`, `httptest.NewRequest`, `t.Fatal`, `t.Parallel`.
- **`TestAgentOutputBound`** — Regression contract: TestAgentOutputBound.
  Signature: `func(t *testing.T)`.
  Calls: `b.Len`, `b.String`, `b.Write`, `bytes.Repeat`, `strings.Contains`, `t.Fatal`, `t.Parallel`.
- **`TestLegacySessionCannotLaunch`** — Regression contract: TestLegacySessionCannotLaunch.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `context.WithCancel`, `db.seed`, `len`, `mockServer`, `s.load`, `s.runAgent`, `structParticipantForIsolation`, `t.Fatal`, `t.Parallel`.
- **`structParticipantForIsolation`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() room.Participant`.
- **`TestClientCannotForgeContainerSessionProvenance`** — Regression contract: TestClientCannotForgeContainerSessionProvenance.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `db.seed`, `len`, `mockRequest`, `mockServer`, `s.load`, `t.Fatal`, `t.Parallel`, `w.Body.String`.
