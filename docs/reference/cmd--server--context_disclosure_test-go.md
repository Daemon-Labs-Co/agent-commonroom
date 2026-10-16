# cmd/server/context_disclosure_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/context_disclosure_test.go)

Automated regression checks for context_disclosure; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`disclosureFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() room.State`.
- **`TestDisclosureExcludesUnauthorizedSources`** — Regression contract: TestDisclosureExcludesUnauthorizedSources.
  Signature: `func(t *testing.T)`.
  Calls: `chatPrompt`, `disclosureFixture`, `strings.Contains`, `t.Fatal`, `t.Parallel`.
- **`TestDisclosurePolicyCannotBeBypassedByEditorFlagOrReferenceText`** — Regression contract: TestDisclosurePolicyCannotBeBypassedByEditorFlagOrReferenceText.
  Signature: `func(t *testing.T)`.
  Calls: `chatPrompt`, `disclosureFixture`, `t.Fatal`, `t.Parallel`, `workflowPrompt`.
- **`TestDisclosureCategoriesAndUnknownDestinations`** — Regression contract: TestDisclosureCategoriesAndUnknownDestinations.
  Signature: `func(t *testing.T)`.
  Calls: `chatPrompt`, `disclosureFixture`, `strings.Contains`, `t.Fatal`, `t.Parallel`.
- **`TestDisclosureConfigurationOwnerAndFlag`** — Regression contract: TestDisclosureConfigurationOwnerAndFlag.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `len`, `mockRequest`, `mockServer`, `s.load`, `t.Context`, `t.Fatal`, `t.Parallel`, `w.Body.String`.
- **`TestDisclosureResetsSessionsAndRejectsMemberGrantChanges`** — Regression contract: TestDisclosureResetsSessionsAndRejectsMemberGrantChanges.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `mockRequest`, `mockServer`, `s.load`, `t.Context`, `t.Fatal`, `t.Parallel`, `w.Body.String`.
- **`TestLegacyNativeContextDoesNotResume`** — Regression contract: TestLegacyNativeContextDoesNotResume.
  Signature: `func(t *testing.T)`.
  Calls: `nativeContextThread`, `t.Fatal`, `t.Parallel`.
- **`TestPersonalMemoryIsNotExposedByHumanRoomProjection`** — Regression contract: TestPersonalMemoryIsNotExposedByHumanRoomProjection.
  Signature: `func(t *testing.T)`.
  Calls: `disclosureFixture`, `len`, `scoped`, `t.Fatal`, `t.Parallel`.
