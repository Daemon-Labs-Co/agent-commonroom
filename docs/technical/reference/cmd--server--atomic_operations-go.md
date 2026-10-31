# cmd/server/atomic_operations.go

[Module index](cmd.md) · [Open source](../../../cmd/server/atomic_operations.go)

Go implementation for atomic operations; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `expectedEventCountKey`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{}`.
- **Type `operationReceiptKey`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{}`.
- **`prepareOperation`** — Operation receipts contain a request fingerprint and the already committed response. A replay acknowledges existing work and never dispatches it again.
  Signature: `func(w http.ResponseWriter, r *http.Request, st room.State, input any) (context.Context, bool)`.
  Calls: `context.WithValue`, `hex.EncodeToString`, `http.Error`, `json.Marshal`, `len`, `r.Context`, `r.Header.Get`, `requestActor`, `requestedRoom`, `sha256.Sum256`, `strings.ContainsAny`, `w.Header`, `w.Header().Set`, `w.Write`, `w.WriteHeader`.
- **`withOperationResponse`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, status int, response any) context.Context`.
  Calls: `context.WithValue`, `ctx.Value`, `json.Marshal`.
- **`operationError`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w http.ResponseWriter, err error)`.
  Calls: `err.Error`, `errors.Is`, `http.Error`.
