# cmd/server/task_repository_links_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/task_repository_links_test.go)

Automated regression checks for task_repository_links; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`repositoryLinksFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*server, *mockDB, http.Handler, room.Task)`.
  Calls: `http.NewServeMux`, `mux.HandleFunc`, `s.actorGuard`, `s.auth.protect`, `structuredTaskFixture`, `t.Helper`.
- **`repositoryLinksCall`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(h http.Handler, method, actor, key, body string) *httptest.ResponseRecorder`.
  Calls: `h.ServeHTTP`, `httptest.NewRecorder`, `httptest.NewRequest`, `r.AddCookie`, `r.Header.Set`, `strings.NewReader`.
- **`repositoryLinksBody`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(task room.Task, revision int, links []room.TaskRepositoryLink) string`.
  Calls: `json.Marshal`, `string`.
- **`TestTaskRepositoryLinksLifecycleFlagsAndReplay`** — Regression contract: TestTaskRepositoryLinksLifecycleFlagsAndReplay.
  Signature: `func(t *testing.T)`.
  Calls: `again.Body.Bytes`, `clear.Body.String`, `db.seed`, `db.snapshot`, `got.Body.Bytes`, `got.Body.String`, `json.Unmarshal`, `len`, `mockRequest`, `reflect.DeepEqual`, `repositoryLinksBody`, `repositoryLinksCall`, `repositoryLinksFixture`, `s.load`, `t.Context`, `t.Fatal`, `t.Parallel`, `taskRepositoryLinksKey`, `w.Body.Bytes`, `w.Body.String`.
- **`TestTaskRepositoryLinksPermissionsAndIntegrity`** — Regression contract: TestTaskRepositoryLinksPermissionsAndIntegrity.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `db.snapshot`, `h.ServeHTTP`, `httptest.NewRecorder`, `httptest.NewRequest`, `len`, `mockRequest`, `r.AddCookie`, `repositoryLinksBody`, `repositoryLinksCall`, `repositoryLinksFixture`, `strings.Contains`, `strings.NewReader`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.Now`, `time.Now().Add`, `w.Body.String`.
- **`TestTaskRepositoryLinkURLBoundaries`** — Regression contract: TestTaskRepositoryLinkURLBoundaries.
  Signature: `func(t *testing.T)`.
  Calls: `strings.Repeat`, `t.Fatal`, `t.Parallel`, `validTaskRepositoryLink`.
- **`TestTaskRepositoryLinksInvalidWritesHaveNoEffects`** — Regression contract: TestTaskRepositoryLinksInvalidWritesHaveNoEffects.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `db.snapshot`, `len`, `make`, `mockRequest`, `repositoryLinksBody`, `repositoryLinksCall`, `repositoryLinksFixture`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `w.Body.String`.
