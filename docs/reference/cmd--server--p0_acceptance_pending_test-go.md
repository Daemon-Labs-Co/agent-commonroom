# cmd/server/p0_acceptance_pending_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/p0_acceptance_pending_test.go)

Automated regression checks for p0_acceptance_pending; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `pendingP0Case`**: This is test-only vocabulary, not a proposed public API or production state enum. Adapters normalize observations from real application entry points, mocked persistence/providers and fake clocks. Never return the Want fields or use real providers/credentials/database connections to satisfy these contracts.
  Declaration: `struct { ID string 'json:"id"' Area string 'json:"area"' Requirements []string 'json:"requirements"' Given map[string]string 'json:"given"' Action string 'json:"action"' Want pendingP0Expectation 'json:"want"' }`.
- **Type `pendingP0Expectation`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { State string 'json:"state"' Effects []string 'json:"effects"' Forbidden []string 'json:"forbidden"' Values map[string]string 'json:"values"' }`.
- **Type `pendingP0Observation`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { State string // Ordered security-relevant effects, including denied/failed attempts. Do // not filter away unexpected effects merely to match the expected sequence. Effects []string Values map[string]string }`.
- **Type `pendingP0Driver`**: Named data type; underlying representation is shown in the declaration.
  Declaration: `func(*testing.T, pendingP0Case) pendingP0Observation`.
- **`loadPendingP0Cases`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) []pendingP0Case`.
  Calls: `bytes.NewReader`, `decoder.Decode`, `decoder.DisallowUnknownFields`, `json.NewDecoder`, `new`, `os.ReadFile`, `t.Fatal`, `t.Helper`.
- **`TestP0RemainingAcceptance`** — Regression contract: TestP0RemainingAcceptance.
  Signature: `func(t *testing.T)`.
  Calls: `driver`, `fmt.Sprintf`, `loadPendingP0Cases`, `os.Getenv`, `pendingP0Mismatches`, `strings.Join`, `t.Error`, `t.Fatal`, `t.Run`, `t.Skip`.
- **`pendingP0Mismatches`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(want pendingP0Expectation, actual pendingP0Observation) []string`.
  Calls: `append`, `fmt.Sprintf`, `reflect.DeepEqual`.
- **`TestP0AcceptanceCatalog`** — Regression contract: TestP0AcceptanceCatalog.
  Signature: `func(t *testing.T)`.
  Calls: `len`, `loadPendingP0Cases`, `os.ReadFile`, `pendingP0Mismatches`, `regexp.MustCompile`, `regexp.MustCompile('(?m)^(FR-[0-9]+):').FindAllStringSubmatch`, `string`, `t.Errorf`, `t.Fatal`, `t.Fatalf`.
- **`TestP0AcceptanceAssertionsRejectIncompleteOrUnsafeObservations`** — Regression contract: TestP0AcceptanceAssertionsRejectIncompleteOrUnsafeObservations.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `len`, `pendingP0Mismatches`, `t.Fatal`.
