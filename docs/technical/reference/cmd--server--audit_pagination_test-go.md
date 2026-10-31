# cmd/server/audit_pagination_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/audit_pagination_test.go)

Automated regression checks for audit_pagination; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestAuditPaginationBounds`** — Regression contract: TestAuditPaginationBounds.
  Signature: `func(t *testing.T)`.
  Calls: `auditPageBounds`, `t.Fatal`, `t.Parallel`, `url.ParseQuery`.
- **`TestAuditPagesAndDateFilters`** — Regression contract: TestAuditPagesAndDateFilters.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `db.snapshot`, `httptest.NewRecorder`, `httptest.NewRequest`, `json.Unmarshal`, `len`, `mockRequest`, `mockServer`, `s.events`, `t.Fatal`, `t.Parallel`, `w.Body.Bytes`, `w.Body.String`.
- **`TestHistoryPagesKeepOwnerAuthorization`** — Regression contract: TestHistoryPagesKeepOwnerAuthorization.
  Signature: `func(t *testing.T)`.
  Calls: `h.ServeHTTP`, `httptest.NewRecorder`, `httptest.NewRequest`, `mockServer`, `t.Fatal`, `t.Parallel`.
