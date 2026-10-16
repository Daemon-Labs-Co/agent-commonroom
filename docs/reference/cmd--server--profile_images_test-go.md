# cmd/server/profile_images_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/profile_images_test.go)

Automated regression checks for profile_images; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`samplePhoto`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, w, h int) string`.
  Calls: `b.Bytes`, `base64.StdEncoding.EncodeToString`, `image.NewRGBA`, `image.Rect`, `png.Encode`, `t.Fatal`, `t.Helper`.
- **`TestProfilePhotoValidation`** — Regression contract: TestProfilePhotoValidation.
  Signature: `func(t *testing.T)`.
  Calls: `normalizeProfileImage`, `samplePhoto`, `strings.HasPrefix`, `strings.Repeat`, `t.Fatal`, `t.Parallel`.
- **`TestHumanPhotoSelfOnlyAndFlag`** — Regression contract: TestHumanPhotoSelfOnlyAndFlag.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `db.seed`, `json.Marshal`, `member`, `mockRequest`, `mockServer`, `s.load`, `samplePhoto`, `string`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`, `w.Body.String`.
- **`TestOwnerFlagsPersistWithoutResumingAgents`** — Regression contract: TestOwnerFlagsPersistWithoutResumingAgents.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `db.seed`, `filepath.Join`, `fresh.agentSafetyScript`, `fresh.agentSafetyScript(root).ServeHTTP`, `httptest.NewRecorder`, `httptest.NewRequest`, `humanProfileImagesEnabled`, `json.Marshal`, `mockRequest`, `mockServer`, `os.WriteFile`, `s.agentSafetyControlsEnabled`, `s.load`, `string`, `t.Fatal`, `t.Parallel`, `t.TempDir`, `w.Body.String`.
