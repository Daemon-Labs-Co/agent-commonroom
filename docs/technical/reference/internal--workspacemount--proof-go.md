# internal/workspacemount/proof.go

[Module index](internal.md) · [Open source](../../../internal/workspacemount/proof.go)

Go implementation for proof; declarations and calls are listed below.

Package: `workspacemount`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `Proof`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Version int 'json:"version"' Challenge string 'json:"challenge"' MountID int 'json:"mountId"' ReadOnly bool 'json:"readOnly"' RootReadOnly bool 'json:"rootReadOnly"' }`.
- **`ValidChallenge`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(s string) bool`.
  Calls: `hex.DecodeString`, `hex.EncodeToString`, `len`.
- **`ValidProof`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p Proof, challenge string) bool`.
  Calls: `ValidChallenge`.
- **`mountID`** — mountID accepts only one explicit read-only workspace mount and no submounts. Unicode/path escapes in mountinfo are not needed for this fixed path; decode kernel octal escapes before checking for workspace descendants.
  Signature: `func(raw []byte) (int, error)`.
  Calls: `mountIDAt`.
- **`mountIDAt`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(raw []byte, root string) (int, error)`.
  Calls: `len`, `mountPath`, `strconv.Atoi`, `string`, `strings.Fields`, `strings.HasPrefix`, `strings.Split`, `strings.TrimSuffix`.
- **`mountPath`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(s string) (string, bool)`.
  Calls: `len`, `out.String`, `out.WriteByte`.
