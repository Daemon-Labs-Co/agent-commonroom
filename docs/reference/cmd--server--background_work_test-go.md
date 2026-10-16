# cmd/server/background_work_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/background_work_test.go)

Automated regression checks for background_work; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestBackgroundWorkerScopeAndUntrustedStatus`** — Regression contract: TestBackgroundWorkerScopeAndUntrustedStatus.
  Signature: `func(t *testing.T)`.
  Calls: `backgroundWorkerStage`, `json.Marshal`, `strings.Repeat`, `t.Fatal`, `t.Parallel`.
- **`TestCodexTelemetryOnlyMapsOperationalEnums`** — Regression contract: TestCodexTelemetryOnlyMapsOperationalEnums.
  Signature: `func(t *testing.T)`.
  Calls: `codexBackgroundStage`, `t.Fatal`, `t.Parallel`.
- **`TestBackgroundLifecycleBoundedAndFlagEnforced`** — Regression contract: TestBackgroundLifecycleBoundedAndFlagEnforced.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `db.seed`, `db.snapshot`, `emitBackground`, `finish`, `finish2`, `len`, `mockServer`, `s.backgroundContext`, `s.load`, `strings.Repeat`, `t.Fatal`, `t.Parallel`.
- **`TestPublicCallerCannotForgeBackgroundTelemetry`** — Regression contract: TestPublicCallerCannotForgeBackgroundTelemetry.
  Signature: `func(t *testing.T)`.
  Calls: `db.snapshot`, `len`, `mockRequest`, `mockServer`, `t.Fatal`, `t.Parallel`.
- **`TestBackgroundProjectionIsNotDuplicatedInProfileAudit`** — Regression contract: TestBackgroundProjectionIsNotDuplicatedInProfileAudit.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `db.snapshot`, `mockServer`, `s.append`, `string`, `strings.Contains`, `t.Fatal`, `t.Parallel`.
- **`TestBackgroundProgressCannotFloodOrExposeOtherRooms`** — Regression contract: TestBackgroundProgressCannotFloodOrExposeOtherRooms.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `db.seed`, `emitBackground`, `finish`, `len`, `mockServer`, `s.backgroundContext`, `s.load`, `strings.Contains`, `t.Fatal`, `t.Parallel`.
