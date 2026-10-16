# cmd/server/members.go

[Module index](cmd.md) · [Open source](../../cmd/server/members.go)

Go implementation for members; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `actorContextKey`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{}`.
- **`requestActor`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r *http.Request) string`.
  Calls: `r.Context`, `r.Context().Value`.
- **`member`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, id string) (room.Actor, bool)`.
- **`canAccess`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, actor, rid string) bool`.
  Calls: `member`.
- **`(*server).actorGuard`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(next http.Handler) http.Handler`.
  Calls: `canAccess`, `context.WithValue`, `http.Error`, `http.HandlerFunc`, `member`, `next.ServeHTTP`, `r.Context`, `r.WithContext`, `requestedRoom`, `s.admitActorMutation`, `s.auth.session`, `s.load`, `strings.HasPrefix`, `time.Now`, `w.Header`, `w.Header().Set`.
- **`visibleState`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, actor, rid string) room.State`.
  Calls: `append`, `canAccess`, `scoped`.
- **`projectActors`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st *room.State)`.
  Calls: `append`, `member`.
- **`(*server).members`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w http.ResponseWriter, r *http.Request)`.
  Calls: `http.Error`, `http.MaxBytesReader`, `id`, `json.NewDecoder`, `json.NewDecoder(r.Body).Decode`, `key.N.BitLen`, `key.N.Cmp`, `len`, `member`, `old.N.Cmp`, `pem.Decode`, `r.Context`, `reply`, `requestActor`, `s.append`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `s.validCSRF`, `strings.TrimSpace`, `x509.ParsePKIXPublicKey`.
- **`(*server).verifyMemberLogin`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, actor, cid, sig string) bool`.
  Calls: `base64.StdEncoding.DecodeString`, `delete`, `len`, `member`, `pem.Decode`, `rsa.VerifyPSS`, `s.load`, `sha256.Sum256`, `time.Now`, `time.Now().Before`, `x509.ParsePKIXPublicKey`.
