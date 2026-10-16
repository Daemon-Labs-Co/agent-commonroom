# cmd/server/auth.go

[Module index](cmd.md) · [Open source](../../cmd/server/auth.go)

Go implementation for auth; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `ownerSession`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { ActorID string CSRF string Expires time.Time }`.
- **Type `ownerAuth`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { mu sync.Mutex publicKey *rsa.PublicKey setupPath, privatePath string sessionPath string setupPending bool challenges map[string]ownerChallenge key string sessions map[string]ownerSession attempts int window time.Time port string }`.
- **`loadOwnerKey`** — The local bootstrap code is never served through HTTP or included in audit payloads.
  Signature: `func(path string) (string, error)`.
  Calls: `f.Close`, `f.WriteString`, `filepath.Dir`, `fmt.Errorf`, `id`, `len`, `os.IsExist`, `os.MkdirAll`, `os.OpenFile`, `readPrivateBounded`, `string`, `strings.TrimSpace`.
- **`newOwnerAuth`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(key, addr string) (*ownerAuth, error)`.
  Calls: `fmt.Errorf`, `ip.IsLoopback`, `net.ParseIP`, `net.SplitHostPort`.
- **`(*ownerAuth).trustedRequest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r *http.Request) bool`.
  Calls: `net.JoinHostPort`, `r.Header.Get`.
- **`(*ownerAuth).session`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r *http.Request) (ownerSession, bool)`.
  Calls: `a.mu.Lock`, `a.mu.Unlock`, `delete`, `r.Cookie`, `sessionDigest`, `time.Now`, `time.Now().Before`.
- **`(*server).validCSRF`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r *http.Request) bool`.
  Calls: `r.Header.Get`, `s.auth.session`, `subtle.ConstantTimeCompare`.
- **`(*ownerAuth).protect`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(next http.Handler) http.Handler`.
  Calls: `a.session`, `a.trustedRequest`, `http.Error`, `http.HandlerFunc`, `next.ServeHTTP`, `strings.HasPrefix`.
- **`(*server).login`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w http.ResponseWriter, r *http.Request)`.
  Calls: `a.mu.Lock`, `a.mu.Unlock`, `a.persistSessionsLocked`, `a.verifySignature`, `decoder.Decode`, `decoder.DisallowUnknownFields`, `delete`, `http.Error`, `http.MaxBytesReader`, `http.SetCookie`, `id`, `json.NewDecoder`, `len`, `new`, `now.Add`, `now.Before`, `now.Sub`, `os.IsNotExist`, `os.Remove`, `r.Context`, `r.Cookie`, `reply`, `s.verifyMemberLogin`, `sessionDigest`, `subtle.ConstantTimeCompare`, `time.Now`, `w.Header`, `w.Header().Set`.
- **`(*server).logout`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w http.ResponseWriter, r *http.Request)`.
  Calls: `delete`, `http.Error`, `http.SetCookie`, `r.Cookie`, `reply`, `s.auth.mu.Lock`, `s.auth.mu.Unlock`, `s.auth.persistSessionsLocked`, `s.validCSRF`, `sessionDigest`.
