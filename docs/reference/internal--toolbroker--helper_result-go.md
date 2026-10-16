# internal/toolbroker/helper_result.go

[Module index](internal.md) · [Open source](../../internal/toolbroker/helper_result.go)

Go implementation for helper result; declarations and calls are listed below.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `helperResult`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Operation string 'json:"operation"' Bytes int 'json:"bytes"' SHA256 string 'json:"sha256"' Encoding string 'json:"encoding,omitempty"' Data string 'json:"data,omitempty"' }`.
- **`validHelperResult`** — validHelperResult validates bounded evidence before an audited success, not filesystem attestation. A compromised image could forge matching metadata; image provenance, verified mounts and the trusted helper are still required.
  Signature: `func(raw []byte, action Action) bool`.
  Calls: `base64.StdEncoding.EncodeToString`, `base64.StdEncoding.EncodedLen`, `base64.StdEncoding.Strict`, `base64.StdEncoding.Strict().DecodeString`, `bytes.Equal`, `bytes.NewReader`, `decoder.Decode`, `decoder.More`, `decoder.Token`, `digest`, `hex.EncodeToString`, `json.Delim`, `json.NewDecoder`, `json.Unmarshal`, `len`, `make`, `sha256.Sum256`, `strictjson.ValidEncoding`.
