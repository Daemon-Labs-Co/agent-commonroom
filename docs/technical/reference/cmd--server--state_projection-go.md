# cmd/server/state_projection.go

[Module index](cmd.md) · [Open source](../../../cmd/server/state_projection.go)

Go implementation for state projection; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`(*server).load`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) (room.State, error)`.
  Calls: `append`, `cancel`, `coordinationJobAllowed`, `ctx.Err`, `delegationTaskDigest`, `delete`, `errors.New`, `fmt.Errorf`, `hasRoom`, `int`, `intent.Scope.Deadline.Add`, `json.Unmarshal`, `json.Valid`, `len`, `omitUnusedAgentRooms`, `projectActors`, `projectAgentElapsed`, `projectAgentStatusDetails`, `projectArtifactCustody`, `projectArtifactManifest`, `projectBackground`, `projectControllerCapture`, `projectControllerCaptureAuthorization`, `projectControllerObjectRead`, `projectControllerProbe`, `projectControllerProbeAuthorization`, `projectDelegation`, `projectGovernance`, `projectHumanAttention`, `projectIdleNotice`, `projectIdleNoticeMessages`, `projectIdleSettings`, `projectManagedStorage`, `projectRoomOutcomeReview`, `projectTaskAcceptanceReview`, `projectTaskExecution`, `projectTaskProposal`, `projectTaskRepositoryLinks`, `projectToolAudit`, `projectWorkspaceDisposal`, `providerForEndpoint`, `recordTaskDefinitionRevision`, `rememberAgentRoomReferences`, `room.SupportsAgent`, `rows.Close`, `rows.Err`, `rows.Next`, `rows.Scan`, `s.db.Query`, `s.storageContext`, `time.Now`, `time.Now().UTC`, `toolbroker.ValidAuditIntent`, `toolbroker.ValidAuditOutcome`, `uint`, `validManagedJob`, `validateTaskDefinitionProjection`.
