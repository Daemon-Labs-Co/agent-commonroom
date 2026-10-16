# cmd/server/organization_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/organization_test.go)

Automated regression checks for organization; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestOrganizationValidation`** — Regression contract: TestOrganizationValidation.
  Signature: `func(t *testing.T)`.
  Calls: `math.NaN`, `t.Fatal`, `t.Parallel`, `t.Run`, `validateOrganization`.
- **`TestOrganizationSavePermissionsAndProjection`** — Regression contract: TestOrganizationSavePermissionsAndProjection.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `db.seed`, `http.HandlerFunc`, `json.Unmarshal`, `len`, `mockRequest`, `mockServer`, `s.actorGuard`, `s.auth.protect`, `s.load`, `strings.Contains`, `strings.Replace`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`, `w.Body.Bytes`, `w.Body.String`.
- **`TestOrganizationGuidanceScopedToRoom`** — Regression contract: TestOrganizationGuidanceScopedToRoom.
  Signature: `func(t *testing.T)`.
  Calls: `organizationGuidance`, `strings.Contains`, `t.Fatal`, `t.Parallel`.
