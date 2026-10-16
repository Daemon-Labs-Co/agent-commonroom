# cmd/server/background_work.go

[Module index](cmd.md) · [Open source](../../cmd/server/background_work.go)

Go implementation for background work; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `backgroundCallbackKey`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{}`.
- **Type `backgroundCallback`**: Named data type; underlying representation is shown in the declaration.
  Declaration: `func(string, string)`.
- **`backgroundEnabled`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State) bool`.
- **`emitBackground`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, stage, job string)`.
  Calls: `cb`, `ctx.Value`.
- **`codexBackgroundStage`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(eventType, itemType string) string`.
- **`(*server).backgroundContext`** — backgroundContext accepts only fixed operational labels. Model text, commands, outputs and credentials must never become telemetry. This is observability, not proof of useful progress or authority to extend execution deadlines.
  Signature: `func(ctx context.Context, p room.Participant) (context.Context, func(error))`.
  Calls: `backgroundCallback`, `backgroundEnabled`, `context.Background`, `context.WithValue`, `ctx.Err`, `emitBackground`, `id`, `s.append`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `time.Now`, `time.Now().UTC`, `validWorkerJobID`.
- **`projectBackground`** — projectBackground bounds delivery; complete retained audit events remain in History. Turning this feature off omits timelines without changing execution.
  Signature: `func(st *room.State, events map[string][]room.BackgroundActivity)`.
  Calls: `append`, `backgroundEnabled`.
