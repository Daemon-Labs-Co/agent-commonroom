# internal/toolbroker/delegation_test.go

[Module index](internal.md) · [Open source](../../../internal/toolbroker/delegation_test.go)

Automated regression checks for delegation; read test names for individual contracts.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `delegationAuditProbe`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{ called bool }`.
- **`(*delegationAuditProbe).RecordIntent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, Intent) error`.
- **`(*delegationAuditProbe).RecordOutcome`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, Intent, Outcome) error`.
- **`(*delegationAuditProbe).RecordDelegation`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, Scope, DelegationRequest) error`.
- **`TestPrivateDelegationFramesRejectMixedOrForgedScope`** — Regression contract: TestPrivateDelegationFramesRejectMixedOrForgedScope.
  Signature: `func(t *testing.T)`.
  Calls: `ServeAudit`, `append`, `auditRead`, `auditWrite`, `bufio.NewReader`, `cancel`, `context.WithTimeout`, `json.Marshal`, `left.Close`, `left.Write`, `len`, `make`, `net.Pipe`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.Now`, `time.Now().UTC`, `time.Now().UTC().Add`.
- **Type `operationSnapshotProbe`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { delegationAuditProbe operations uint8 }`.
- **`(*operationSnapshotProbe).RecordDelegation`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(_ context.Context, _ Scope, r DelegationRequest) error`.
- **`TestSupervisorDelegationOverwritesClaimedOperationPermissions`** — Regression contract: TestSupervisorDelegationOverwritesClaimedOperationPermissions.
  Signature: `func(t *testing.T)`.
  Calls: `New`, `NewAuditClient`, `ServeAudit`, `cancel`, `context.WithCancel`, `left.Close`, `make`, `net.Pipe`, `session.ProposeDelegation`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().UTC`, `time.Now().UTC().Add`.
