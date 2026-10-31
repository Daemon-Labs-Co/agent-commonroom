# cmd/server/governance_commit_race_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/governance_commit_race_test.go)

Automated regression checks for governance_commit_race; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `interleavedEventDB`**: Commits through an independent controller immediately before the original controller's append transaction starts. No goroutine timing or live DB needed.
  Declaration: `struct { eventDatabase beforeBegin func(context.Context) }`.
- **`(*interleavedEventDB).Begin`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) (pgx.Tx, error)`.
  Calls: `d.beforeBegin`, `d.eventDatabase.Begin`.
- **`TestCoordinationWritesRevalidateAfterConcurrentCommit`** — Regression contract: TestCoordinationWritesRevalidateAfterConcurrentCommit.
  Signature: `func(t *testing.T)`.
  Calls: `appendCtx.Value`, `coordinatedContext`, `coordinationFixture`, `db.snapshot`, `errors.Is`, `http.NewServeMux`, `mockRequest`, `mux.HandleFunc`, `other.actorGuard`, `other.auth.protect`, `registration.Register`, `registration.Update`, `replaceCoordinator`, `s.finishAgentTurn`, `s.load`, `s.managedJobContext`, `s.managedJobContext(ctx, p).Value`, `s.recordAgentSession`, `strings.Contains`, `t.Context`, `t.Error`, `t.Fatal`, `t.Parallel`, `t.Run`, `w.Body.String`.
- **`TestIndependentResultSurvivesConcurrentCoordinatorTransfer`** — Existing independent human work is not revoked by a coordinator replacement.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `coordinationFixture`, `http.NewServeMux`, `mux.HandleFunc`, `other.actorGuard`, `other.auth.protect`, `replaceCoordinator`, `s.finishAgentTurn`, `s.load`, `t.Context`, `t.Fatal`, `t.Parallel`.
