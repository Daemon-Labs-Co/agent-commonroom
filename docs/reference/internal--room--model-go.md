# internal/room/model.go

[Module index](internal.md) · [Open source](../../internal/room/model.go)

Go implementation for model; declarations and calls are listed below.

Package: `room`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `Connection`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Revision int 'json:"revision"' Location string 'json:"location"' CompanyID string 'json:"companyId"' Provider string 'json:"provider"' ID string 'json:"id"' Name string 'json:"name"' Protocol string 'json:"protocol"' Endpoint string 'json:"endpoint"' }`.
- **Type `Participant`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { DisclosureVersion int 'json:"disclosureVersion,omitempty"' Revision int 'json:"revision"' Background []BackgroundActivity 'json:"background,omitempty"' StatusDetails *AgentStatusDetails 'json:"statusDetails,omitempty"' ActorID string 'json:"actorId"' ExecutionMode string 'json:"executionMode,omitempty"' CompanyID string 'json:"companyId"' Discipline string 'json:"discipline"' Level string 'json:"level"' Speciality string 'json:"speciality"' Assignment string 'json:"assignment"' Challenge *int 'json:"challenge"' RoomID string 'json:"roomId"' ID string 'json:"id"' ConnectionID string 'json:"connectionId"' Name string 'json:"name"' Role string 'json:"role"' Instructions string 'json:"instructions"' Rank int 'json:"rank"' Accept bool 'json:"accept"' AdditionalContext string 'json:"additionalContext"' ContextSource string 'json:"contextSource"' ThreadID string 'json:"threadId"' Status string 'json:"status"' }`.
- **Type `BackgroundActivity`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { ParticipantID string 'json:"participantId"' RoomID string 'json:"roomId"' TurnID string 'json:"turnId"' JobID string 'json:"jobId,omitempty"' Stage string 'json:"stage"' Label string 'json:"label"' ObservedAt time.Time 'json:"observedAt"' }`.
- **Type `Memory`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { AgentExport *bool 'json:"agentExport,omitempty"' ExportProviders []string 'json:"exportProviders,omitempty"' SourceActorID string 'json:"sourceActorId,omitempty"' RoomID string 'json:"roomId"' ID string 'json:"id"' Title string 'json:"title"' Body string 'json:"body"' Scope string 'json:"scope"' Revision int 'json:"revision"' }`.
- **Type `Message`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { HumanRequest string 'json:"humanRequest,omitempty"' AttentionAcknowledgedBy string 'json:"attentionAcknowledgedBy,omitempty"' ActorID string 'json:"actorId"' ActorType string 'json:"actorType"' Provider string 'json:"provider,omitempty"' RoomID string 'json:"roomId"' ID string 'json:"id"' Author string 'json:"author"' Text string 'json:"text"' Created time.Time 'json:"created"' }`.
- **Type `ContextExportPolicy`**: ContextExportPolicy is owner-managed disclosure authorization, never model instructions.
  Declaration: `struct { Providers []string 'json:"providers"' ExcludeMessages bool 'json:"excludeMessages"' ExcludeMemory bool 'json:"excludeMemory"' ExcludeReferences bool 'json:"excludeReferences"' }`.
- **Type `Room`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Revision int 'json:"revision"' ContextExport *ContextExportPolicy 'json:"contextExport,omitempty"' PendingHumanRequests int 'json:"pendingHumanRequests,omitempty"' PersonalActorID string 'json:"personalActorId"' CompanyID string 'json:"companyId"' HumanPresent *bool 'json:"humanPresent"' HumanRank *int 'json:"humanRank"' HumanName string 'json:"humanName"' ID string 'json:"id"' Name string 'json:"name"' Description string 'json:"description"' }`.
- **Type `Task`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { DispatchID string 'json:"dispatchId,omitempty"' DispatchStatus string 'json:"dispatchStatus,omitempty"' DispatchDetail string 'json:"dispatchDetail,omitempty"' ID string 'json:"id"' RoomID string 'json:"roomId"' Title string 'json:"title"' Criteria string 'json:"criteria"' Assignee string 'json:"assignee"' Reviewer string 'json:"reviewer"' Status string 'json:"status"' Revision int 'json:"revision"' }`.
- **Type `Evidence`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { ID string 'json:"id"' RoomID string 'json:"roomId"' TaskID string 'json:"taskId"' Kind string 'json:"kind"' Title string 'json:"title"' Body string 'json:"body"' Source string 'json:"source"' Created time.Time 'json:"created"' }`.
- **Type `Company`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Revision int 'json:"revision"' ID string 'json:"id"' Name string 'json:"name"' Description string 'json:"description"' }`.
- **Type `OrganizationChart`**: OrganizationChart is owner-managed company metadata. Reporting relationships replace numbered decision seniority but do not grant permissions or tool access.
  Declaration: `struct { Revision int 'json:"revision"' Nodes []OrganizationNode 'json:"nodes"' }`.
- **Type `OrganizationNode`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { ActorID string 'json:"actorId"' ReportsTo string 'json:"reportsTo"' X float64 'json:"x"' Y float64 'json:"y"' }`.
- **Type `Actor`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { HelpDisplay string 'json:"helpDisplay,omitempty"' Image string 'json:"image,omitempty"' ID string 'json:"id"' Name string 'json:"name"' Title string 'json:"title"' Kind string 'json:"kind"' Role string 'json:"role"' Enabled bool 'json:"enabled"' PublicKey string 'json:"publicKey,omitempty"' RoomIDs []string 'json:"roomIds"' Revision int 'json:"revision"' ChangedBy string 'json:"changedBy"' }`.
- **Type `ManagedJob`**: ManagedJob is a safe, durable execution binding. Model data and arguments never belong here; ShutdownConfirmed comes only from a fenced worker proof.
  Declaration: `struct { Execution string 'json:"execution,omitempty"' RepositoryDigest string 'json:"repositoryDigest,omitempty"' InitiatorID string 'json:"initiatorId,omitempty"' CompanyID string 'json:"companyId,omitempty"' Version int 'json:"version"' JobID string 'json:"jobId"' RoomID string 'json:"roomId"' ParticipantID string 'json:"participantId"' ConnectionID string 'json:"connectionId"' Worker string 'json:"worker"' Provider string 'json:"provider"' Status string 'json:"status"' DiagnosticCode string 'json:"diagnosticCode"' AllocatedAt time.Time 'json:"allocatedAt"' UpdatedAt time.Time 'json:"updatedAt"' Deadline time.Time 'json:"deadline"' ShutdownConfirmed bool 'json:"shutdownConfirmed"' }`.
- **Type `OperationReceipt`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Key string 'json:"key"' Fingerprint string 'json:"fingerprint"' Status int 'json:"status"' Response json.RawMessage 'json:"response"' }`.
- **Type `ToolCallAudit`**: ToolCallAudit is replayed intent/outcome evidence for a broker-mediated tool call (internal/toolbroker). It is durable audit metadata only — never raw tool output, capability tokens or Docker launch details. Status "pending" means an intent was recorded with no matching outcome yet; replay never infers completion or retries on its own. Commit timestamps are not surfaced here yet; add them from the event's protected _auditRecord metadata only when a consumer actually needs to display them.
  Declaration: `struct { CallID string 'json:"callId"' JobID string 'json:"jobId"' RoomID string 'json:"roomId"' ParticipantID string 'json:"participantId"' ConnectionID string 'json:"connectionId"' Provider string 'json:"provider"' Operation string 'json:"operation"' PayloadDigest string 'json:"payloadDigest"' Status string 'json:"status"' OutputBytes int 'json:"outputBytes"' OutputDigest string 'json:"outputDigest"' }`.
- **Type `State`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { WorkspaceDisposals map[string]WorkspaceDisposal 'json:"-"' ManagedStorage *ManagedStorage 'json:"-"' HumanRevision int 'json:"humanRevision"' EventCount int64 'json:"-"' OperationReceipts map[string]OperationReceipt 'json:"-"' ManagedJobs map[string]ManagedJob 'json:"-"' ToolAudits map[string]ManagedToolCall 'json:"-"' ToolCalls map[string]ToolCallAudit 'json:"-"' Organization OrganizationChart 'json:"organization"' FeatureFlags map[string]bool 'json:"featureFlags"' AgentsStopped bool 'json:"agentsStopped"' AgentRoomStopped bool 'json:"agentRoomStopped"' AgentStopRooms []string 'json:"-"' Actors []Actor 'json:"actors"' CurrentActorID string 'json:"currentActorId"' Company Company 'json:"company"' HumanCompanyID string 'json:"humanCompanyId"' HumanTitle string 'json:"humanTitle"' HumanName string 'json:"humanName"' WorkflowRuns []WorkflowRun 'json:"workflowRuns"' Workflows []Workflow 'json:"workflows"' Rooms []Room 'json:"rooms"' Tasks []Task 'json:"tasks"' Evidence []Evidence 'json:"evidence"' Connections []Connection 'json:"connections"' Participants []Participant 'json:"participants"' Memories []Memory 'json:"memories"' Messages []Message 'json:"messages"' Incident bool 'json:"incident"' }`.
- **Type `AgentAdapter`**: AgentAdapter separates provider behavior from delivery. Capability discovery precedes invocation.
  Declaration: `interface{ Capabilities() Capabilities }`.
- **Type `Capabilities`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{ Launch, Attach, Resume, Cancel, Steer bool }`.
- **`SupportsCodex`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(c Connection) bool`.
  Calls: `regexp.MustCompile`, `regexp.MustCompile('^[A-Za-z0-9][A-Za-z0-9_.-]*$').MatchString`, `strings.HasPrefix`, `strings.TrimPrefix`.
- **`SupportsAgent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(c Connection) bool`.
  Calls: `SupportsCodex`, `regexp.MustCompile`, `regexp.MustCompile('^ssh://[A-Za-z0-9][A-Za-z0-9_.-]*/claude$').MatchString`.
- **Type `WorkflowNode`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { ID string 'json:"id"' Type string 'json:"type"' Title string 'json:"title"' X float64 'json:"x"' Y float64 'json:"y"' AgentID string 'json:"agentId"' Instructions string 'json:"instructions"' Timeout int 'json:"timeout"' Retries int 'json:"retries"' Evidence string 'json:"evidence"' }`.
- **Type `WorkflowEdge`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { ID string 'json:"id"' From string 'json:"from"' To string 'json:"to"' Port string 'json:"port"' }`.
- **Type `Workflow`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { ID string 'json:"id"' RoomID string 'json:"roomId"' Name string 'json:"name"' Revision int 'json:"revision"' MaxIterations int 'json:"maxIterations"' Nodes []WorkflowNode 'json:"nodes"' Edges []WorkflowEdge 'json:"edges"' }`.
- **Type `WorkflowRun`**: WorkflowRun is an append-only projection of a pinned graph and its execution. Outputs are provider reports, not independently verified filesystem evidence.
  Declaration: `struct { ID string 'json:"id"' RoomID string 'json:"roomId"' Workflow Workflow 'json:"workflow"' Revision int 'json:"revision"' Status string 'json:"status"' NodeID string 'json:"nodeId"' Input string 'json:"input"' Detail string 'json:"detail"' ActorID string 'json:"actorId"' Updated time.Time 'json:"updated"' Visits map[string]int 'json:"visits"' Steps []WorkflowStep 'json:"steps"' }`.
- **Type `WorkflowStep`**: WorkflowStep pins each completed attempt's output, hash, identity and evidence.
  Declaration: `struct { NodeID string 'json:"nodeId"' AgentID string 'json:"agentId"' Output string 'json:"output"' SHA256 string 'json:"sha256"' Evidence string 'json:"evidence"' Port string 'json:"port"' Completed time.Time 'json:"completed"' }`.
- **Type `AgentStatusDetails`**: AgentStatusDetails is a safe display projection, never raw provider output.
  Declaration: `struct { Activity string 'json:"activity"' Reason string 'json:"reason"' Next string 'json:"next"' ElapsedSeconds int 'json:"elapsedSeconds,omitempty"' LongRunning bool 'json:"longRunning,omitempty"' }`.
