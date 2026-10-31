# internal/taskcontract/definition.go

[Module index](internal.md) · [Open source](../../../internal/taskcontract/definition.go)

Go implementation for definition; declarations and calls are listed below.

Package: `taskcontract`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `Definition`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { SchemaVersion int 'json:"schemaVersion"' Revision int 'json:"revision"' Task string 'json:"task"' ExpectedInputs []DataContract 'json:"expectedInput"' Process Process 'json:"process"' DesiredOutputs []DataContract 'json:"desiredOutput"' AcceptanceCriteria []Criterion 'json:"acceptanceCriteria"' ExecutionConstraints Limits 'json:"executionConstraints"' FailureContract Failure 'json:"failureContract"' }`.
- **Type `Process`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Mode string 'json:"mode"' Instructions string 'json:"instructions"' Constraints []string 'json:"constraints"' }`.
- **Type `DataContract`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Name string 'json:"name"' Type string 'json:"type"' // file, bundle, value; no arbitrary provider type. Required bool 'json:"required"' MediaType string 'json:"mediaType,omitempty"' SchemaID string 'json:"schemaId,omitempty"' MaxBytes int64 'json:"maxBytes"' FileName string 'json:"fileName,omitempty"' Members []BundleMember 'json:"members,omitempty"' Properties []ValueProperty 'json:"properties,omitempty"' ItemSchemaID string 'json:"itemSchemaId,omitempty"' MaxItems int 'json:"maxItems,omitempty"' }`.
- **Type `ValueProperty`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Name string 'json:"name"' SchemaID string 'json:"schemaId"' Required bool 'json:"required"' }`.
- **Type `BundleMember`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Name string 'json:"name"' MediaType string 'json:"mediaType"' MaxBytes int64 'json:"maxBytes"' }`.
- **Type `Criterion`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { ID string 'json:"id"' Description string 'json:"description"' Method string 'json:"method"' Output string 'json:"output"' }`.
- **Type `Limits`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { ResourceScope string 'json:"resourceScope"' AllowedOperations []string 'json:"allowedOperations"' MaxDurationSeconds int 'json:"maxDurationSeconds"' MaxResponseBytes int 'json:"maxResponseBytes"' MaxToolCalls int 'json:"maxToolCalls"' CompletionAuthority string 'json:"completionAuthority"' }`.
- **Type `Failure`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { MissingInput string 'json:"missingInput"' ExecutionFailure string 'json:"executionFailure"' Retry string 'json:"retry"' }`.
- **`bounded`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(s string, max int) bool`.
  Calls: `strings.ContainsRune`, `strings.TrimSpace`, `utf8.RuneCountInString`, `utf8.ValidString`.
- **`media`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(s string) bool`.
  Calls: `len`, `mime.ParseMediaType`, `strings.ContainsAny`.
- **`relative`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(s string) bool`.
  Calls: `bounded`, `path.Clean`, `strings.ContainsAny`, `strings.HasPrefix`.
- **`schema`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(s string) bool`.
- **`ValidData`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(c DataContract) error`.
  Calls: `errors.New`, `len`, `media`, `namePattern.MatchString`, `primitive`, `relative`, `schema`, `strings.HasPrefix`.
- **`Validate`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(d Definition) error`.
  Calls: `ValidData`, `bounded`, `errors.New`, `json.Marshal`, `len`, `namePattern.MatchString`.
- **`Digest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(d Definition) string`.
  Calls: `hex.EncodeToString`, `json.Marshal`, `sha256.Sum256`.
- **`Same`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(a, b Definition) bool`.
  Calls: `Digest`.
- **`Allows`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(d Definition, operation string) bool`.
- **`Decode`** — Decode rejects unknown/duplicate object keys, excessive nesting, malformed Unicode and trailing JSON. Size limits apply before any decoding allocations.
  Signature: `func(raw []byte) (Definition, error)`.
  Calls: `Validate`, `bytes.NewReader`, `dec.Decode`, `dec.DisallowUnknownFields`, `errors.New`, `json.NewDecoder`, `len`, `scanJSON`, `strictjson.ValidEncoding`.
- **`CriteriaText`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(d Definition) string`.
  Calls: `append`, `fmt.Sprintf`, `len`, `make`, `strings.Join`.
- **`scanJSON`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(raw []byte) error`.
  Calls: `bytes.NewReader`, `decoder.More`, `decoder.Token`, `errors.New`, `json.NewDecoder`, `scan`.
- **`primitive`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(id string) bool`.
- **`ValidJSON`** — ValidJSON checks bounded grammar, Unicode, duplicates and depth before typed decoding.
  Signature: `func(raw []byte, maxBytes int) bool`.
  Calls: `len`, `scanJSON`, `strictjson.ValidEncoding`.
