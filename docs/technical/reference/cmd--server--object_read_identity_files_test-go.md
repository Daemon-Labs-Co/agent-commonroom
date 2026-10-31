# cmd/server/object_read_identity_files_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/object_read_identity_files_test.go)

Automated regression checks for object_read_identity_files; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`writeObjectReadTestIdentity`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, identity objectReadPeerIdentity) objectReadIdentityFiles`.
  Calls: `append`, `filepath.Join`, `hex.EncodeToString`, `os.WriteFile`, `pem.EncodeToMemory`, `t.Fatal`, `t.Helper`, `t.TempDir`, `x509.MarshalPKCS8PrivateKey`.
- **`provisionedObjectReadTestFiles`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (objectReadIdentityFiles, objectReadIdentityFiles)`.
  Calls: `provisionedObjectReadTestFilesWith`.
- **`provisionedObjectReadTestFilesWith`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, mutate func(*x509.Certificate, bool)) (objectReadIdentityFiles, objectReadIdentityFiles)`.
  Calls: `big.NewInt`, `ed25519.GenerateKey`, `issue`, `mutate`, `now.Add`, `os.WriteFile`, `pem.EncodeToMemory`, `sha256.Sum256`, `t.Fatal`, `t.Helper`, `time.Now`, `writeObjectReadTestIdentity`, `x509.CreateCertificate`, `x509.ParseCertificate`.
- **`TestObjectReadIdentityFilesMutualAuthentication`** — Regression contract: TestObjectReadIdentityFilesMutualAuthentication.
  Signature: `func(t *testing.T)`.
  Calls: `a.Close`, `authenticateObjectReadStream`, `b.Close`, `cancel`, `context.WithTimeout`, `loadObjectReadPeerIdentity`, `make`, `net.Pipe`, `os.WriteFile`, `provisionedObjectReadTestFiles`, `t.Context`, `t.Fatal`, `t.Parallel`, `validObjectReadPeerIdentity`.
- **`TestObjectReadIdentityFilesRefuseUnsafeInputs`** — Regression contract: TestObjectReadIdentityFilesRefuseUnsafeInputs.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `filepath.Join`, `loadObjectReadPeerIdentity`, `os.Chmod`, `os.Link`, `os.ReadFile`, `os.Remove`, `os.Symlink`, `os.WriteFile`, `provisionedObjectReadTestFiles`, `read`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`, `validObjectReadPeerIdentity`, `write`.
- **`TestObjectReadIdentityFilesCertificatePolicy`** — Regression contract: TestObjectReadIdentityFilesCertificatePolicy.
  Signature: `func(t *testing.T)`.
  Calls: `loadObjectReadPeerIdentity`, `provisionedObjectReadTestFilesWith`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.Now`, `time.Now().Add`, `validObjectReadPeerIdentity`.
- **`TestObjectReadIdentityFilesMissingPEMMarkersExplained`** — Regression contract: TestObjectReadIdentityFilesMissingPEMMarkersExplained.
  Signature: `func(t *testing.T)`.
  Calls: `err.Error`, `loadObjectReadPeerIdentity`, `os.ReadFile`, `os.Remove`, `os.WriteFile`, `pem.Decode`, `provisionedObjectReadTestFiles`, `string`, `strings.ReplaceAll`, `t.Fatal`, `t.Parallel`, `t.Run`, `validObjectReadPeerIdentity`.
