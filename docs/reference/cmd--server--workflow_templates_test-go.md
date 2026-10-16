# cmd/server/workflow_templates_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/workflow_templates_test.go)

Automated regression checks for workflow_templates; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestWorkflowTemplateFlagAndPermissions`** — Regression contract: TestWorkflowTemplateFlagAndPermissions.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `db.seed`, `db.snapshot`, `len`, `mockRequest`, `mockServer`, `s.load`, `strings.Contains`, `strings.Replace`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.Now`, `time.Now().Add`, `w.Body.String`.
- **`TestWorkflowTemplateAssetFlag`** — Regression contract: TestWorkflowTemplateAssetFlag.
  Signature: `func(t *testing.T)`.
  Calls: `filepath.Join`, `httptest.NewRecorder`, `httptest.NewRequest`, `optionalScriptHandler`, `optionalScriptHandler(enabled, root, "workflow-templates.js").ServeHTTP`, `os.WriteFile`, `t.Fatal`, `t.Parallel`, `t.TempDir`, `w.Body.Len`.
- **`TestAdditionalTemplateApplications`** — Regression contract: TestAdditionalTemplateApplications.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `db.snapshot`, `len`, `mockRequest`, `mockServer`, `s.load`, `t.Fatal`, `t.Parallel`, `t.Run`, `w.Body.String`.
