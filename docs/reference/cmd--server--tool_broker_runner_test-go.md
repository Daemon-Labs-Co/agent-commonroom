# cmd/server/tool_broker_runner_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/tool_broker_runner_test.go)

Automated regression checks for tool_broker_runner; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestDockerCLIRunnerRejectsMissingPreconditions`** — dockerCLIRunner shells out to the real "docker" CLI, so these tests cover only its precondition checks (which must fail before any subprocess runs), not live Docker behavior. Per this repo's established convention for every other Docker-touching path (isolation.go), live verification stays a documented manual step, not a mocked exec.Command in the automated suite.
  Signature: `func(t *testing.T)`.
  Calls: `(dockerCLIRunner{}).Run`, `(dockerCLIRunner{}).RunInput`, `canceledContext`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`canceledContext`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) context.Context`.
  Calls: `cancel`, `context.WithCancel`, `t.Context`, `t.Helper`.
