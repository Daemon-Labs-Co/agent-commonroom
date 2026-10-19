# internal/toolbroker/audit_wire.go

[Module index](internal.md) · [Open source](../../internal/toolbroker/audit_wire.go)

Go implementation for audit wire; declarations and calls are listed below.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`ValidAuditScope`** — ValidAuditScope validates trusted admission metadata against a trusted clock.
  Signature: `func(scope Scope, now time.Time) bool`.
  Calls: `validScope`.
- **Type `AuditFrame`**: AuditFrame belongs to the authenticated supervisor/controller channel, never the model's callback channel. It contains hashes and bounded lifecycle metadata, never actions or capabilities.
  Declaration: `struct { IdleStop *IdleStop 'json:"idleStop,omitempty"' Version int 'json:"version"' Sequence int 'json:"sequence"' Session string 'json:"session"' Stage string 'json:"stage"' Intent Intent 'json:"intent"' Outcome *Outcome 'json:"outcome,omitempty"' }`.
- **Type `auditAck`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Version int 'json:"version"' Digest string 'json:"digest"' Accepted bool 'json:"accepted"' }`.
- **`ValidAuditIntent`** — ValidAuditIntent checks bounded metadata. now is a trusted admission clock; outcome persistence may instead use the original intent's admission instant.
  Signature: `func(intent Intent, now time.Time) bool`.
  Calls: `digest`, `identifier`, `operation`, `validScope`.
- **`ValidAuditOutcome`** — ValidAuditOutcome bounds evidence metadata; hashes are not filesystem attestation.
  Signature: `func(outcome Outcome) bool`.
  Calls: `digest`.
- **`auditFrameDigest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(frame AuditFrame) (string, error)`.
  Calls: `hex.EncodeToString`, `json.Marshal`, `len`, `sha256.Sum256`.
- **`auditDecode`** — auditDecode rejects duplicate/unknown/case-folded/missing members and explicit nulls before typed decoding. All wire fields are mandatory except outcome and the version-2 idleStop extension. Stage validation forbids mixed payloads.
  Signature: `func(raw []byte, target any) error`.
  Calls: `auditShape`, `auditUniqueValue`, `bytes.NewReader`, `decoder.Token`, `json.NewDecoder`, `json.Unmarshal`, `len`, `reflect.TypeOf`, `reflect.TypeOf(target).Elem`, `strictjson.ValidEncoding`.
- **`auditShape`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(value any, schema reflect.Type) bool`.
  Calls: `auditShape`, `bytes.IndexByte`, `field.Tag.Get`, `len`, `reflect.TypeOf`, `schema.Elem`, `schema.Field`, `schema.Kind`, `schema.NumField`.
- **`auditUniqueValue`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(decoder *json.Decoder, depth int) error`.
  Calls: `auditUniqueValue`, `decoder.More`, `decoder.Token`.
- **`auditRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(reader *bufio.Reader, target any) error`.
  Calls: `auditDecode`, `len`, `reader.ReadSlice`.
- **`auditWrite`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(writer io.Writer, value any) error`.
  Calls: `append`, `json.Marshal`, `len`, `writer.Write`.
