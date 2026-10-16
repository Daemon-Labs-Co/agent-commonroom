# cmd/server/atomic_operations_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/atomic_operations_test.go)

Automated regression checks for atomic_operations; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`keyedEvent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(h http.Handler, key, body string) *httptest.ResponseRecorder`.
  Calls: `h.ServeHTTP`, `httptest.NewRecorder`, `httptest.NewRequest`, `r.AddCookie`, `r.Header.Set`, `strings.NewReader`.
- **`TestOperationReplayAndKeyConflict`** — Regression contract: TestOperationReplayAndKeyConflict.
  Signature: `func(t *testing.T)`.
  Calls: `a.Body.Bytes`, `a.Body.String`, `b.Body.Bytes`, `b.Body.String`, `b.Header`, `b.Header().Get`, `db.snapshot`, `json.RawMessage`, `keyedEvent`, `len`, `mockServer`, `referenceJSON`, `t.Fatal`, `t.Parallel`.
- **`TestBatchRollbackAtEveryInsertAndCommit`** — Regression contract: TestBatchRollbackAtEveryInsertAndCommit.
  Signature: `func(t *testing.T)`.
  Calls: `context.WithValue`, `db.snapshot`, `json.RawMessage`, `len`, `newMockDB`, `s.appendBatch`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestExpectedHeadRejectsConcurrentBusinessChange`** — Regression contract: TestExpectedHeadRejectsConcurrentBusinessChange.
  Signature: `func(t *testing.T)`.
  Calls: `context.WithValue`, `db.snapshot`, `len`, `newMockDB`, `s.append`, `s.appendBatch`, `s.load`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestBatchMaintainsCanonicalAuditChain`** — Regression contract: TestBatchMaintainsCanonicalAuditChain.
  Signature: `func(t *testing.T)`.
  Calls: `db.snapshot`, `len`, `newMockDB`, `s.appendBatch`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestStaleMutableAggregateEdits`** — Regression contract: TestStaleMutableAggregateEdits.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `db.snapshot`, `len`, `mockRequest`, `mockServer`, `t.Fatal`, `t.Parallel`, `t.Run`, `w.Body.String`.
- **`TestFailedOperationCanRetrySameKeyWithoutPartialReceipt`** — Regression contract: TestFailedOperationCanRetrySameKeyWithoutPartialReceipt.
  Signature: `func(t *testing.T)`.
  Calls: `db.snapshot`, `keyedEvent`, `len`, `mockServer`, `t.Fatal`, `t.Parallel`, `w.Body.String`.
