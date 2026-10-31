# cmd/server/task_evidence_disclosure_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/task_evidence_disclosure_test.go)

Automated regression checks for task_evidence_disclosure; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestTaskEvidenceDisclosureHonorsReferencePolicyIndependentlyOfOptionalMetadata`** — Regression contract: TestTaskEvidenceDisclosureHonorsReferencePolicyIndependentlyOfOptionalMetadata.
  Signature: `func(t *testing.T)`.
  Calls: `strings.Contains`, `strings.Join`, `t.Fatal`, `t.Parallel`, `t.Run`, `taskDispatchPrompt`.
- **`TestExcludedOversizedEvidenceDoesNotBlockButIncludedOversizedEvidenceRefuses`** — Regression contract: TestExcludedOversizedEvidenceDoesNotBlockButIncludedOversizedEvidenceRefuses.
  Signature: `func(t *testing.T)`.
  Calls: `string`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `taskDispatchPrompt`.
- **`TestActualTypedDispatchExcludesEvidenceButRetainsExplicitBindings`** — Regression contract: TestActualTypedDispatchExcludesEvidenceButRetainsExplicitBindings.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `definedDispatchBody`, `dispatchRequest`, `make`, `strings.Contains`, `structuredTaskFixture`, `t.Fatal`, `t.Parallel`, `time.After`, `w.Body.String`, `waitDispatch`.
