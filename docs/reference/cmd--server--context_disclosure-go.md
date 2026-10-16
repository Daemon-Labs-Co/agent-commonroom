# cmd/server/context_disclosure.go

[Module index](cmd.md) · [Open source](../../cmd/server/context_disclosure.go)

Go implementation for context disclosure; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`contextPolicyEditorEnabled`** — Export restrictions are mandatory. The editor flag only controls configuration UX; disabling it never grants a model access to previously excluded material.
  Signature: `func(st room.State) bool`.
  Calls: `flagDefined`.
- **`flagDefined`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, name string) bool`.
- **`validExportProviders`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(providers []string) bool`.
  Calls: `len`.
- **`containsProvider`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(providers []string, provider string) bool`.
- **`contextDestination`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, p room.Participant) (string, *room.ContextExportPolicy, error)`.
  Calls: `containsProvider`, `errors.New`, `providerForEndpoint`.
- **`memoryExportAllowed`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, m room.Memory, p room.Participant, provider string) bool`.
  Calls: `containsProvider`, `len`.
- **`referenceJSON`** — referenceJSON keeps source metadata distinct from lower-trust content. Labels support model interpretation; actual exclusion happens before serialization.
  Signature: `func(value any) string`.
  Calls: `json.Marshal`, `string`.
- **`nativeContextThread`** — Legacy native sessions may contain context exported before disclosure checks. Start a new session once; preserved application history is filtered normally.
  Signature: `func(p room.Participant) string`.
