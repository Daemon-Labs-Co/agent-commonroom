# cmd/server/owner_rsa_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/owner_rsa_test.go)

Automated regression checks for owner_rsa; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestRSAOwnerLifecycle`** — Regression contract: TestRSAOwnerLifecycle.
  Signature: `func(t *testing.T)`.
  Calls: `a.prepareRSA`, `a.publicKey.N.BitLen`, `a.verifySignature`, `b.prepareRSA`, `base64.StdEncoding.EncodeToString`, `context.WithValue`, `filepath.Join`, `httptest.NewRecorder`, `httptest.NewRequest`, `json.Marshal`, `loadOwnerKey`, `newOwnerAuth`, `os.IsNotExist`, `os.Stat`, `pem.Decode`, `r.AddCookie`, `r.Context`, `r.Header.Set`, `r.WithContext`, `readPrivateFile`, `rsa.SignPSS`, `s.downloadOwnerKey`, `s.login`, `sha256.Sum256`, `string`, `strings.NewReader`, `t.Fatal`, `t.Parallel`, `t.TempDir`, `time.Now`, `time.Now().Add`, `w.Body.String`, `w.Header`, `w.Header().Get`, `x509.ParsePKCS8PrivateKey`.
