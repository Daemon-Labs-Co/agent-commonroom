# cmd/server/owner_rsa.go

[Module index](cmd.md) · [Open source](../../../cmd/server/owner_rsa.go)

Go implementation for owner rsa; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `ownerChallenge`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Value string Expires time.Time }`.
- **`writePrivateFile`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(path string, data []byte) error`.
  Calls: `f.Close`, `f.Write`, `os.OpenFile`.
- **`readPrivateFile`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(path string) ([]byte, error)`.
  Calls: `readPrivateBounded`.
- **`(*ownerAuth).prepareRSA`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(path string) error`.
  Calls: `a.publicKey.N.BitLen`, `fmt.Errorf`, `os.IsNotExist`, `os.Remove`, `pem.Decode`, `pem.EncodeToMemory`, `readPrivateFile`, `rsa.GenerateKey`, `writePrivateFile`, `x509.MarshalPKCS8PrivateKey`, `x509.MarshalPKIXPublicKey`, `x509.ParsePKCS8PrivateKey`, `x509.ParsePKIXPublicKey`.
- **`(*server).challenge`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w http.ResponseWriter, r *http.Request)`.
  Calls: `a.mu.Lock`, `a.mu.Unlock`, `delete`, `http.Error`, `id`, `len`, `now.Add`, `now.Before`, `reply`, `time.Now`.
- **`(*ownerAuth).verifySignature`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(challengeID, signature string) bool`.
  Calls: `base64.StdEncoding.DecodeString`, `delete`, `rsa.VerifyPSS`, `sha256.Sum256`, `time.Now`, `time.Now().Before`.
- **`(*server).downloadOwnerKey`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w http.ResponseWriter, r *http.Request)`.
  Calls: `a.mu.Lock`, `a.mu.Unlock`, `http.Error`, `readPrivateFile`, `requestActor`, `s.validCSRF`, `w.Header`, `w.Header().Set`, `w.Write`.
