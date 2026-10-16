# internal/toolbroker/pair_resources.go

[Module index](internal.md) · [Open source](../../internal/toolbroker/pair_resources.go)

Go implementation for pair resources; declarations and calls are listed below.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`VerifyPairResources`** — VerifyPairResources validates bounded Docker inspection objects obtained by the trusted worker. Both roles receive at most half the logical job ceiling; absent, unlimited, negative or swap-expanded limits are refused. This checks reported limits only, not daemon enforcement, image policy, mounts or identity labels. Callers must also verify those independently before starting a pair. A source fork CAN skip this check, BUT invalidates aggregate admission claims.
  Signature: `func(engine, tool []byte) error`.
  Calls: `decodeInspection`, `digest`.
