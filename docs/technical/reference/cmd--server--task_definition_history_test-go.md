# cmd/server/task_definition_history_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/task_definition_history_test.go)

Automated regression checks for task_definition_history; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`definitionHistoryHandler`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(s *server) http.Handler`.
  Calls: `http.NewServeMux`, `mux.HandleFunc`, `s.actorGuard`, `s.auth.protect`.
- **`TestDefinitionHistoryPreservesUnexecutedVersionsAndDeduplicatesStateEdits`** — Regression contract: TestDefinitionHistoryPreservesUnexecutedVersionsAndDeduplicatesStateEdits.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `definitionHistoryHandler`, `fmt.Sprintf`, `json.Marshal`, `json.Unmarshal`, `len`, `mockRequest`, `s.load`, `string`, `strings.Contains`, `structuredTaskFixture`, `t.Context`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`, `w.Body.Bytes`, `w.Body.String`.
- **`TestDefinitionHistoryPrivateReadsRequireAuditAndLegacyIsEmpty`** — Regression contract: TestDefinitionHistoryPrivateReadsRequireAuditAndLegacyIsEmpty.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `db.snapshot`, `definitionHistoryHandler`, `len`, `mockRequest`, `mockServer`, `strings.Contains`, `t.Fatal`, `t.Parallel`, `w.Body.String`.
- **`TestDefinitionHistoryDoesNotInventFailedWritesOrAcceptMutation`** — Regression contract: TestDefinitionHistoryDoesNotInventFailedWritesOrAcceptMutation.
  Signature: `func(t *testing.T)`.
  Calls: `definitionHistoryHandler`, `fmt.Sprintf`, `json.Marshal`, `len`, `mockRequest`, `s.load`, `structuredTaskFixture`, `t.Context`, `t.Fatal`, `t.Parallel`.
