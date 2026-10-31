# cmd/server/audit_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/audit_test.go)

Automated regression checks for audit; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestAuditSummary`** — Regression contract: TestAuditSummary.
  Signature: `func(t *testing.T)`.
  Calls: `auditSummary`, `strings.Contains`, `t.Errorf`, `t.Fatal`, `t.Parallel`.
