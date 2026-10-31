# cmd/server/object_read_identity_files.go

[Module index](cmd.md) · [Open source](../../../cmd/server/object_read_identity_files.go)

Go implementation for object read identity files; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `objectReadIdentityFiles`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Certificate string 'json:"certificate"' PrivateKey string 'json:"privateKey"' Authorities string 'json:"authorities"' PeerDNSName string 'json:"peerDNSName"' PeerSHA256 string 'json:"peerSHA256"' }`.
- **`objectReadCertificateBundle`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(raw []byte, limit int) ([]*x509.Certificate, error)`.
  Calls: `append`, `bytes.Count`, `bytes.HasPrefix`, `bytes.TrimSpace`, `len`, `pem.Decode`, `x509.ParseCertificate`.
- **`objectReadCertificateCurrent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(c *x509.Certificate, now time.Time) bool`.
  Calls: `now.Before`.
- **`loadObjectReadPeerIdentity`** — loadObjectReadPeerIdentity loads a fresh independent snapshot for one endpoint role (server=false means worker/client). Failure returns no partial identity and a sanitized error. Rotation requires new provisioning and a fresh session; this loader never changes a running session or relaxes TLS/peer/account authorization. It is mandatory security infrastructure, without a flag or host/SSH fallback. A source fork weakening the loader or its trusted parent boundary voids custody.
  Signature: `func(files objectReadIdentityFiles, server bool) (objectReadPeerIdentity, error)`.
  Calls: `bytes.Count`, `bytes.HasPrefix`, `bytes.TrimSpace`, `clear`, `copy`, `fail`, `filepath.Clean`, `filepath.IsAbs`, `hex.DecodeString`, `len`, `objectReadCertificateBundle`, `objectReadCertificateCurrent`, `pem.Decode`, `readPrivateBounded`, `roots.AddCert`, `sha256.Sum256`, `strings.ContainsAny`, `strings.ToLower`, `strings.TrimSpace`, `time.Now`, `tls.X509KeyPair`, `validObjectReadPeerIdentity`, `x509.NewCertPool`.
