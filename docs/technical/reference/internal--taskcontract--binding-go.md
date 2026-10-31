# internal/taskcontract/binding.go

[Module index](internal.md) · [Open source](../../../internal/taskcontract/binding.go)

Go implementation for binding; declarations and calls are listed below.

Package: `taskcontract`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `ValueReference`**: Binding belongs to an execution, never to its reusable definition. File and bundle references require independently verified immutable storage; this initial value-input resolver deliberately refuses unresolved object claims. ValueReference names an immutable returned value. Only the controller may resolve it and attach accepted-source provenance; labels are not proof.
  Declaration: `struct { ExecutionID string 'json:"executionId"' OutputName string 'json:"outputName"' }`.
- **Type `ValueProvenance`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { TaskID string 'json:"taskId"' ExecutionID string 'json:"executionId"' OutputName string 'json:"outputName"' TaskRevision int 'json:"taskRevision"' DefinitionDigest string 'json:"definitionDigest"' OutputDigest string 'json:"outputDigest"' }`.
- **Type `Binding`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Source *ValueReference 'json:"source,omitempty"' ValueJSON string 'json:"valueJson,omitempty"' Name string 'json:"name"' Value json.RawMessage 'json:"value,omitempty"' SourceVersion string 'json:"sourceVersion"' }`.
- **Type `BoundValue`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Provenance *ValueProvenance 'json:"provenance,omitempty"' ValueJSON string 'json:"valueJson"' Name string 'json:"name"' Value json.RawMessage 'json:"value"' SourceVersion string 'json:"sourceVersion"' Digest string 'json:"digest"' }`.
- **`ValidValue`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(schemaID string, raw []byte) bool`.
  Calls: `bytes.NewReader`, `dec.Decode`, `dec.UseNumber`, `json.NewDecoder`, `len`, `math.IsInf`, `math.IsNaN`, `n.Float64`, `n.String`, `scanJSON`, `strconv.ParseInt`, `strictjson.ValidEncoding`.
- **`Bind`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(d Definition, inputs []Binding) ([]BoundValue, error)`.
  Calls: `ValidContractValue`, `Validate`, `append`, `bounded`, `bytesDigest`, `canonicalValue`, `errors.New`, `int64`, `json.RawMessage`, `len`, `string`.
- **`bytesDigest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(raw []byte) string`.
  Calls: `hex.EncodeToString`, `sha256.Sum256`.
- **`ReturnedValues`** — ReturnedValues validates actual returned JSON, not a model claim that an artifact exists. Immutable file/bundle custody belongs to the artifact path.
  Signature: `func(d Definition, raw []byte, version string) ([]BoundValue, error)`.
  Calls: `ValidContractValue`, `append`, `bytesDigest`, `canonicalValue`, `errors.New`, `int64`, `json.Marshal`, `json.Unmarshal`, `len`, `scanJSON`, `sort.Slice`, `strictjson.ValidEncoding`, `string`.
- **`Executable`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(d Definition) error`.
  Calls: `errors.New`.
- **`canonicalValue`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(raw []byte) json.RawMessage`.
  Calls: `bytes.NewReader`, `dec.Decode`, `dec.UseNumber`, `json.Marshal`, `json.NewDecoder`.
- **`ValidContractValue`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(c DataContract, raw []byte) bool`.
  Calls: `ValidData`, `ValidValue`, `json.Unmarshal`, `len`.
