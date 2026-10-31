# cmd/server/tool_command_runner_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/tool_command_runner_test.go)

Automated regression checks for tool_command_runner; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestCommandProcessCapturesBothStreamsAndClientStatus`** — Regression contract: TestCommandProcessCapturesBothStreamsAndClientStatus.
  Signature: `func(t *testing.T)`.
  Calls: `bytes.Contains`, `exec.CommandContext`, `output.Bytes`, `output.Len`, `runDockerCommand`, `slices.Equal`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`.
- **`TestCommandProcessRejectsCanceledContext`** — Regression contract: TestCommandProcessRejectsCanceledContext.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.WithCancel`, `runDockerCommand`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestIsolatedCommandsSettingDefaultsOffAndFailsClosed`** — Regression contract: TestIsolatedCommandsSettingDefaultsOffAndFailsClosed.
  Signature: `func(t *testing.T)`.
  Calls: `isolatedCommandsSetting`, `t.Fatal`, `t.Parallel`.
