# internal/artifactmanifest/metadata.go

[Module index](internal.md) · [Open source](../../../internal/artifactmanifest/metadata.go)

Go implementation for metadata; declarations and calls are listed below.

Package: `artifactmanifest`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `Claim`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Name string 'json:"name"' Receipt objectcustody.Receipt 'json:"receipt"' }`.
- **Type `File`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Name string 'json:"name"' MediaType string 'json:"mediaType"' StructuralCheck string 'json:"structuralCheck"' Receipt objectcustody.Receipt 'json:"receipt"' }`.
- **Type `Manifest`**: Manifest contains immutable identities, not bytes, host paths, capabilities or observed build/test results. Its digest is not a substitute for revalidation.
  Declaration: `struct { Version int 'json:"version"' Scope objectcustody.ExecutionScope 'json:"scope"' OutputName string 'json:"outputName"' OutputType string 'json:"outputType"' ContractDigest string 'json:"contractDigest"' Files []File 'json:"files"' TotalBytes int64 'json:"totalBytes"' Digest string 'json:"digest"' }`.
- **Type `format`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{ media, check string }`.
- **`registeredFormat`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(value string) (format, bool)`.
  Calls: `len`, `mime.FormatMediaType`, `mime.ParseMediaType`, `strings.EqualFold`.
- **`validFormatBytes`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(f format, raw []byte) bool`.
  Calls: `len`, `string`, `taskcontract.ValidJSON`, `unicode.IsControl`, `utf8.Valid`.
- **`digest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(value any) string`.
  Calls: `hex.EncodeToString`, `json.Marshal`, `sha256.Sum256`.
- **`manifestDigest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(value Manifest) string`.
  Calls: `digest`.
- **`metadata`** — metadata constructs canonical shape only, without reading bytes or granting authority. Keep all contract/claim checks shared with actual-byte Resolve.
  Signature: `func(definition taskcontract.Definition, scope objectcustody.ExecutionScope, outputName string, claims []Claim) (Manifest, error)`.
  Calls: `append`, `digest`, `len`, `make`, `manifestDigest`, `objectcustody.ValidExecutionScope`, `objectcustody.ValidReceipt`, `registeredFormat`, `sort.Slice`, `taskcontract.Digest`, `taskcontract.Validate`.
- **`ValidateMetadata`** — ValidateMetadata checks exact canonical contract/receipt shape and digest only. Replay may use this without filesystem I/O. It is NOT actual-byte verification, current authorization, delivery evidence or a substitute for Revalidate.
  Signature: `func(definition taskcontract.Definition, scope objectcustody.ExecutionScope, output string, claimed Manifest) error`.
  Calls: `len`, `make`, `metadata`, `reflect.DeepEqual`.
