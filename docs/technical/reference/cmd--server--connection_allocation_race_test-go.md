# cmd/server/connection_allocation_race_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/connection_allocation_race_test.go)

Automated regression checks for connection_allocation_race; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`connectionRaceFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*server, *mockDB, room.Participant)`.
  Calls: `db.seed`, `mockServer`, `t.Helper`.
- **`connectionRaceBody`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(operation string) string`.
  Calls: `fmt.Sprintf`, `json.Marshal`.
- **`connectionRaceHandler`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(s *server) http.Handler`.
  Calls: `http.NewServeMux`, `mux.HandleFunc`, `s.actorGuard`, `s.auth.protect`.
- **`TestConnectionMutationAndWorkerAllocationCannotBothCommitStaleChecks`** — Regression contract: TestConnectionMutationAndWorkerAllocationCannotBothCommitStaleChecks.
  Signature: `func(t *testing.T)`.
  Calls: `connectionHasActiveAgent`, `connectionRaceBody`, `connectionRaceFixture`, `connectionRaceHandler`, `context.Background`, `ctx.Value`, `errors.Is`, `len`, `mockRequest`, `other.load`, `other.managedJobContext`, `other.managedJobContext(context.Background(), p).Value`, `reg.Register`, `reg.Update`, `s.managedJobContext`, `s.managedJobContext(context.Background(), p).Value`, `t.Context`, `t.Error`, `t.Fatal`, `t.Parallel`, `t.Run`, `w.Body.String`.
