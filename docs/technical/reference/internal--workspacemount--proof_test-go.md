# internal/workspacemount/proof_test.go

[Module index](internal.md) · [Open source](../../../internal/workspacemount/proof_test.go)

Automated regression checks for proof; read test names for individual contracts.

Package: `workspacemount`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestMountObservationParser`** — Regression contract: TestMountObservationParser.
  Signature: `func(t *testing.T)`.
  Calls: `mountID`, `strings.Repeat`, `strings.Replace`, `t.Fatalf`, `t.Parallel`, `t.Run`.
- **`TestProofBinding`** — Regression contract: TestProofBinding.
  Signature: `func(t *testing.T)`.
  Calls: `ValidChallenge`, `ValidProof`, `change`, `strings.Repeat`, `t.Fatal`, `t.Parallel`.
- **`TestFixedObjectMountParser`** — Regression contract: TestFixedObjectMountParser.
  Signature: `func(t *testing.T)`.
  Calls: `mountIDAt`, `strings.Replace`, `t.Fatal`, `t.Parallel`, `t.Run`.
