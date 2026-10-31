# internal/room/model_test.go

[Module index](internal.md) · [Open source](../../../internal/room/model_test.go)

Automated regression checks for model; read test names for individual contracts.

Package: `room`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestAdapterEndpointBoundaries`** — Regression contract: TestAdapterEndpointBoundaries.
  Signature: `func(t *testing.T)`.
  Calls: `SupportsAgent`, `SupportsCodex`, `t.Fatalf`, `t.Parallel`, `t.Run`.
