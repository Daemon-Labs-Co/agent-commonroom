# cmd/server/workflow_browser_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/workflow_browser_test.go)

Automated regression checks for workflow_browser; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestWorkflowBrowserHarness`** — TestWorkflowBrowserHarness is an opt-in manual UI fixture. It serves the real handlers with in-memory event storage and fake providers, never PostgreSQL, real credentials or real agents. Normal suites skip the bounded browser wait.
  Signature: `func(t *testing.T)`.
  Calls: `body.Bytes`, `db.seed`, `documentationHandler`, `executionGraph`, `executionServer`, `fmt.Sprintf`, `host.Close`, `host.ListenAndServe`, `http.Dir`, `http.Error`, `http.FileServer`, `http.NewServeMux`, `int64`, `json.Unmarshal`, `make`, `mux.Handle`, `mux.HandleFunc`, `os.Getenv`, `refreshes.Add`, `reply`, `reviews.Add`, `s.actorGuard`, `s.auth.protect`, `s.state`, `strings.Repeat`, `t.Cleanup`, `t.Fatal`, `t.Log`, `t.Skip`, `time.After`, `time.Now`, `time.Now().UTC`, `time.Unix`, `time.Unix(1, 0).UTC`, `time.Unix(2, 0).UTC`, `time.Unix(3, 0).UTC`, `time.Unix(30, 0).UTC`, `time.Unix(int64(i), 0).UTC`.
