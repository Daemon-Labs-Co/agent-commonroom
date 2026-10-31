# cmd/server/adapter_command_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/adapter_command_test.go)

Automated regression checks for adapter_command; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestIsolatedCommandPolicy`** — Regression contract: TestIsolatedCommandPolicy.
  Signature: `func(t *testing.T)`.
  Calls: `isolatedRunArgs`, `len`, `strings.Contains`, `strings.Join`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `validAgentImage`.
- **`TestNoRemoteOrLegacyFallback`** — Regression contract: TestNoRemoteOrLegacyFallback.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `isolatedAgentCommand`, `t.Fatal`, `t.Parallel`.
- **`TestProviderEgressPolicy`** — Regression contract: TestProviderEgressPolicy.
  Signature: `func(t *testing.T)`.
  Calls: `providerHostAllowed`, `t.Fatal`, `t.Parallel`.
- **`TestDockerSecurityPrerequisites`** — Regression contract: TestDockerSecurityPrerequisites.
  Signature: `func(t *testing.T)`.
  Calls: `dockerSecurityReady`, `strings.Replace`, `t.Fatal`, `t.Parallel`.
- **`TestIsolatedRunKeepsPromptInputOpen`** — Docker must keep stdin open: Codex receives its bounded prompt over stdin, including fresh connection checks. A missing -i silently drops the message.
  Signature: `func(t *testing.T)`.
  Calls: `isolatedRunArgs`, `len`, `t.Fatal`, `t.Parallel`.
