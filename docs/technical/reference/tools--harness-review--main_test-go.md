# tools/harness-review/main_test.go

[Module index](tools.md) · [Open source](../../../tools/harness-review/main_test.go)

Automated regression checks for main; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`publicationFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) ([]string, ed25519.PublicKey, time.Time)`.
  Calls: `base64.StdEncoding.EncodeToString`, `ed25519.GenerateKey`, `filepath.Join`, `json.Marshal`, `now.Add`, `os.WriteFile`, `strings.Repeat`, `t.Fatal`, `t.Helper`, `t.TempDir`, `time.Now`, `time.Now().UTC`.
- **`TestPublishReviewRoundTrip`** — Regression contract: TestPublishReviewRoundTrip.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `bytes.Equal`, `filepath.Dir`, `filepath.Glob`, `filepath.Join`, `harnessrelease.ReadPrivate`, `harnessrelease.Verify`, `len`, `os.ReadFile`, `publicationFixture`, `run`, `t.Fatal`, `t.Parallel`.
- **`TestPublishReviewRefusals`** — Regression contract: TestPublishReviewRefusals.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `base64.StdEncoding.DecodeString`, `base64.StdEncoding.EncodeToString`, `filepath.Join`, `now.Add`, `os.Chmod`, `os.IsNotExist`, `os.Link`, `os.ReadFile`, `os.Rename`, `os.Stat`, `os.Symlink`, `os.WriteFile`, `out.Len`, `publicationFixture`, `run`, `string`, `strings.TrimSpace`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`.
- **`TestConcurrentPublicationIsExclusive`** — Regression contract: TestConcurrentPublicationIsExclusive.
  Signature: `func(t *testing.T)`.
  Calls: `close`, `harnessrelease.ReadPrivate`, `harnessrelease.Verify`, `make`, `publicationFixture`, `run`, `t.Fatal`, `t.Parallel`, `wg.Add`, `wg.Done`, `wg.Wait`.
- **`TestGenerateOfflineReviewKey`** — Regression contract: TestGenerateOfflineReviewKey.
  Signature: `func(t *testing.T)`.
  Calls: `base64.StdEncoding.DecodeString`, `base64.StdEncoding.EncodeToString`, `ed25519.PrivateKey`, `ed25519.PrivateKey(private).Public`, `filepath.Join`, `harnessrelease.ReadPrivate`, `len`, `out.String`, `publicationFixture`, `run`, `string`, `strings.Contains`, `strings.TrimSpace`, `t.Fatal`, `t.Parallel`, `t.TempDir`.
