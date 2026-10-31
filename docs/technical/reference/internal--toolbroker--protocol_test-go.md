# internal/toolbroker/protocol_test.go

[Module index](internal.md) · [Open source](../../../internal/toolbroker/protocol_test.go)

Automated regression checks for protocol; read test names for individual contracts.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`fixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*Authority, Request, time.Time)`.
  Calls: `New`, `now.Add`, `strings.Repeat`, `t.Fatal`, `t.Helper`, `time.Date`.
- **`TestDecodeStrict`** — Regression contract: TestDecodeStrict.
  Signature: `func(t *testing.T)`.
  Calls: `Decode`, `bytes.NewReader`, `fixture`, `json.Marshal`, `len`, `string`, `strings.NewReader`, `strings.Repeat`, `strings.Replace`, `t.Fatal`, `t.Fatalf`, `t.Parallel`.
- **`TestScopeAndCapabilityRejection`** — Regression contract: TestScopeAndCapabilityRejection.
  Signature: `func(t *testing.T)`.
  Calls: `a.Reserve`, `a.Revoke`, `change`, `fixture`, `r.Scope.Deadline.Add`, `strings.Repeat`, `t.Fatal`, `t.Parallel`.
- **`TestReplayConcurrencyAndBudget`** — Regression contract: TestReplayConcurrencyAndBudget.
  Signature: `func(t *testing.T)`.
  Calls: `a.Reserve`, `accepted.Add`, `accepted.Load`, `fixture`, `fmt.Sprintf`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `wg.Add`, `wg.Done`, `wg.Wait`.
- **`TestOperationAndSizeBounds`** — Regression contract: TestOperationAndSizeBounds.
  Signature: `func(t *testing.T)`.
  Calls: `Decode`, `New`, `a.Reserve`, `bytes.NewReader`, `change`, `fixture`, `json.Marshal`, `strings.Repeat`, `t.Fatal`, `t.Parallel`.
- **`TestPolicyValidationAndCopy`** — Regression contract: TestPolicyValidationAndCopy.
  Signature: `func(t *testing.T)`.
  Calls: `New`, `a.Reserve`, `fixture`, `now.Add`, `t.Fatal`, `t.Parallel`.
- **`TestTypedOperations`** — Regression contract: TestTypedOperations.
  Signature: `func(t *testing.T)`.
  Calls: `New`, `a.Reserve`, `fixture`, `now.Add`, `t.Fatal`, `t.Fatalf`, `t.Parallel`.
