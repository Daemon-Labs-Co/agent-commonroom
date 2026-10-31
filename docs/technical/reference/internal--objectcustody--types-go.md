# internal/objectcustody/types.go

[Module index](internal.md) · [Open source](../../../internal/objectcustody/types.go)

Go implementation for types; declarations and calls are listed below.

Package: `objectcustody`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `ExecutionScope`**: Scope is checked against the authenticated principal and actual isolated job by Journal. These identifiers/hashes are evidence bindings, never authority.
  Declaration: `struct { CompanyID string 'json:"companyId"' RoomID string 'json:"roomId"' TaskID string 'json:"taskId"' ExecutionID string 'json:"executionId"' JobID string 'json:"jobId"' RepositoryDigest string 'json:"repositoryDigest"' RevisionDigest string 'json:"revisionDigest"' DefinitionDigest string 'json:"definitionDigest"' }`.
- **Type `Scope`**: Scope adds immutable output/member labels. Labels never become filesystem paths here; the trusted source adapter must verify their actual origin.
  Declaration: `struct { ExecutionScope OutputName string 'json:"outputName"' MemberName string 'json:"memberName"' }`.
- **Type `Intent`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { ObjectID string 'json:"objectId"' Scope Scope 'json:"scope"' MaxBytes int64 'json:"maxBytes"' }`.
- **Type `Receipt`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Intent Intent 'json:"intent"' Bytes int64 'json:"bytes"' Digest string 'json:"digest"' }`.
- **`lowerDigest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(value string, size int) bool`.
  Calls: `hex.DecodeString`, `hex.EncodeToString`, `len`.
- **`ValidExecutionScope`** — ValidExecutionScope validates metadata shape only, never admission/authority.
  Signature: `func(s ExecutionScope) bool`.
  Calls: `identifier.MatchString`, `lowerDigest`.
- **`validScope`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(s Scope) bool`.
  Calls: `ValidExecutionScope`, `identifier.MatchString`, `path.Clean`, `path.IsAbs`, `strings.ContainsAny`, `strings.HasPrefix`, `strings.TrimSpace`, `utf8.RuneCountInString`, `utf8.ValidString`.
- **`validIntent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(v Intent) bool`.
  Calls: `lowerDigest`, `validScope`.
- **`validReceipt`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(v Receipt) bool`.
  Calls: `lowerDigest`, `validIntent`.
- **`ValidIntent`** — ValidIntent validates metadata only; a trusted journal must admit capacity.
  Signature: `func(v Intent) bool`.
  Calls: `validIntent`.
- **`ValidReceipt`** — ValidReceipt validates metadata only; it does not prove stored byte identity.
  Signature: `func(v Receipt) bool`.
  Calls: `validReceipt`.
- **Type `Journal`**: Journal is a mandatory TRUSTED controller/worker integration boundary, never a model/plugin callback. Reserve must durably charge maximum capacity across retained/staged/uncertain account data AND establish an actual physical quota before any bytes are read/created. Seal must revalidate current job, source revision, policy and authority and durably acknowledge the exact content. AuthorizeRead requires an exact existing sealed receipt plus current room, principal, revision and export policy. AcknowledgeRead must revalidate current authority again and durably record bytes authorized for release, not proof that a remote client received them. A missing, no-op or synthetic implementation cannot admit production work. There is deliberately no automatic refund, deletion, adoption or retry API.
  Declaration: `interface { Reserve(context.Context, Intent) error Seal(context.Context, Receipt) error AuthorizeRead(context.Context, Scope, Receipt) error AcknowledgeRead(context.Context, Receipt) error }`.
