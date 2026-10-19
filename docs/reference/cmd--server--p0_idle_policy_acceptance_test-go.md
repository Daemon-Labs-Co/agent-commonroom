# cmd/server/p0_idle_policy_acceptance_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/p0_idle_policy_acceptance_test.go)

Automated regression checks for p0_idle_policy_acceptance; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`driveP0IdleUnauthorizedConfig`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, scenario pendingP0Case) pendingP0Observation`.
  Calls: `append`, `db.seed`, `db.snapshot`, `idleSettingsHTTP`, `json.Marshal`, `mockRequest`, `reflect.DeepEqual`, `s.load`, `strconv.ParseFloat`, `string`, `t.Context`, `t.Fatal`, `t.Helper`, `time.Now`, `time.Now().Add`.
