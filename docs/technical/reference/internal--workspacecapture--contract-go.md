# internal/workspacecapture/contract.go

[Module index](internal.md) · [Open source](../../../internal/workspacecapture/contract.go)

Go implementation for contract; declarations and calls are listed below.

Package: `workspacecapture`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`ContractRequest`** — ContractRequest derives finite capture members from the exact task definition. All required file/bundle outputs are included. optionalOutputs may name only explicitly chosen optional file/bundle outputs, never values or arbitrary paths. This validates metadata only: current human/task/export authority, frozen source, quotas and durable journals remain mandatory. No root, launch or read is opened.
  Signature: `func(definition taskcontract.Definition, binding workspacesnapshot.Binding, challenge string, optionalOutputs []string) (Request, error)`.
  Calls: `RequestDigest`, `append`, `delete`, `len`, `sort.Slice`, `taskcontract.Digest`, `taskcontract.Validate`.
- **`ValidateContractRequest`** — ValidateContractRequest requires the exact canonical selection, definition digest, required outputs, member identities and limits. A trusted parent authority MUST compare against its independently loaded current definition, never one supplied by the helper/model. It grants no authority and performs no I/O.
  Signature: `func(definition taskcontract.Definition, request Request) error`.
  Calls: `ContractRequest`, `append`, `len`, `reflect.DeepEqual`, `taskcontract.Validate`.
