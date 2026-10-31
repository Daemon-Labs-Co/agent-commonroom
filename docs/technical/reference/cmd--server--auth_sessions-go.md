# cmd/server/auth_sessions.go

[Module index](cmd.md) · [Open source](../../../cmd/server/auth_sessions.go)

Go implementation for auth sessions; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `savedHumanSessions`**: Session persistence saves hashes of opaque cookies, never PEM keys or cookie tokens. The private server file is bound to the owner public key; normal expiry, CSRF, origin and actor authorization checks still apply after restart.
  Declaration: `struct { Version int 'json:"version"' Identity string 'json:"identity"' Sessions map[string]ownerSession 'json:"sessions"' }`.
- **`sessionDigest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(token string) string`.
  Calls: `hex.EncodeToString`, `sha256.Sum256`.
- **`(*ownerAuth).sessionIdentity`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() string`.
  Calls: `sessionDigest`, `string`, `x509.MarshalPKCS1PublicKey`.
- **`(*ownerAuth).configureSessionPersistence`** — configureSessionPersistence is called once after RSA setup. Unsafe files or corrupt content fail startup closed. Disabled persistence invalidates the previous store so switching it back on cannot resurrect signed-out sessions.
  Signature: `func(path string, enabled bool) error`.
  Calls: `a.persistSessionsLocked`, `a.sessionIdentity`, `errors.New`, `hex.DecodeString`, `json.Unmarshal`, `len`, `now.Add`, `now.Before`, `os.IsNotExist`, `os.Remove`, `readPrivateBounded`, `session.Expires.After`, `strings.HasPrefix`, `syncParentDirectory`, `time.Now`.
- **`(*ownerAuth).persistSessionsLocked`** — persistSessionsLocked requires a.mu; atomic replacement makes revocation durable before a successful logout response. Storage failure fails the action.
  Signature: `func() error`.
  Calls: `a.sessionIdentity`, `filepath.Dir`, `json.Marshal`, `os.MkdirAll`, `replacePrivateDurably`, `sessionDigest`, `strings.HasPrefix`, `time.Now`, `time.Now().Before`.
- **`persistentHumanSessionsEnabled`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() (bool, error)`.
  Calls: `os.Getenv`, `strconv.ParseBool`.
