package room

import (
	"agentroom/internal/taskcontract"
	"encoding/json"
	"regexp"
	"strings"
	"time"
)

type Connection struct {
	Revision  int    `json:"revision"`
	Location  string `json:"location"`
	CompanyID string `json:"companyId"`
	Provider  string `json:"provider"`
	ID        string `json:"id"`
	Name      string `json:"name"`
	Protocol  string `json:"protocol"`
	Endpoint  string `json:"endpoint"`
}
type Participant struct {
	DisclosureVersion int                  `json:"disclosureVersion,omitempty"`
	Revision          int                  `json:"revision"`
	Background        []BackgroundActivity `json:"background,omitempty"`
	StatusDetails     *AgentStatusDetails  `json:"statusDetails,omitempty"`
	ActorID           string               `json:"actorId"`
	ExecutionMode     string               `json:"executionMode,omitempty"`
	CompanyID         string               `json:"companyId"`
	Discipline        string               `json:"discipline"`
	Level             string               `json:"level"`
	Speciality        string               `json:"speciality"`
	Assignment        string               `json:"assignment"`
	Challenge         *int                 `json:"challenge"`
	RoomID            string               `json:"roomId"`
	ID                string               `json:"id"`
	ConnectionID      string               `json:"connectionId"`
	Name              string               `json:"name"`
	Role              string               `json:"role"`
	Instructions      string               `json:"instructions"`
	Rank              int                  `json:"rank"`
	Accept            bool                 `json:"accept"`
	AdditionalContext string               `json:"additionalContext"`
	ContextSource     string               `json:"contextSource"`
	ThreadID          string               `json:"threadId"`
	Status            string               `json:"status"`
}

type BackgroundActivity struct {
	ParticipantID string    `json:"participantId"`
	RoomID        string    `json:"roomId"`
	TurnID        string    `json:"turnId"`
	JobID         string    `json:"jobId,omitempty"`
	Stage         string    `json:"stage"`
	Label         string    `json:"label"`
	ObservedAt    time.Time `json:"observedAt"`
}
type Memory struct {
	AgentExport     *bool    `json:"agentExport,omitempty"`
	ExportProviders []string `json:"exportProviders,omitempty"`
	SourceActorID   string   `json:"sourceActorId,omitempty"`
	RoomID          string   `json:"roomId"`
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	Body            string   `json:"body"`
	Scope           string   `json:"scope"`
	Revision        int      `json:"revision"`
}
type Message struct {
	HumanRequest            string    `json:"humanRequest,omitempty"`
	AttentionAcknowledgedBy string    `json:"attentionAcknowledgedBy,omitempty"`
	ActorID                 string    `json:"actorId"`
	ActorType               string    `json:"actorType"`
	Provider                string    `json:"provider,omitempty"`
	RoomID                  string    `json:"roomId"`
	ID                      string    `json:"id"`
	Author                  string    `json:"author"`
	Text                    string    `json:"text"`
	Created                 time.Time `json:"created"`
}

// ContextExportPolicy is owner-managed disclosure authorization, never model instructions.
type ContextExportPolicy struct {
	Providers         []string `json:"providers"`
	ExcludeMessages   bool     `json:"excludeMessages"`
	ExcludeMemory     bool     `json:"excludeMemory"`
	ExcludeReferences bool     `json:"excludeReferences"`
}
type Room struct {
	Outcome            string `json:"outcome,omitempty"`
	AcceptanceCriteria string `json:"acceptanceCriteria,omitempty"`
	NameVersion        string `json:"nameVersion,omitempty"`
	// Reconstructed from explicit name changes, never accepted from clients.
	CustomPersonalName   bool                 `json:"-"`
	Revision             int                  `json:"revision"`
	ContextExport        *ContextExportPolicy `json:"contextExport,omitempty"`
	PendingHumanRequests int                  `json:"pendingHumanRequests,omitempty"`
	PersonalActorID      string               `json:"personalActorId"`
	CompanyID            string               `json:"companyId"`
	HumanPresent         *bool                `json:"humanPresent"`
	HumanRank            *int                 `json:"humanRank"`
	HumanName            string               `json:"humanName"`
	ID                   string               `json:"id"`
	Name                 string               `json:"name"`
	Description          string               `json:"description"`
}
type Task struct {
	Definition     *taskcontract.Definition `json:"definition,omitempty"`
	DelegationID   string                   `json:"delegationId,omitempty"`
	ProposalID     string                   `json:"proposalId,omitempty"`
	DispatchID     string                   `json:"dispatchId,omitempty"`
	DispatchStatus string                   `json:"dispatchStatus,omitempty"`
	DispatchDetail string                   `json:"dispatchDetail,omitempty"`
	ID             string                   `json:"id"`
	RoomID         string                   `json:"roomId"`
	Title          string                   `json:"title"`
	Criteria       string                   `json:"criteria"`
	Assignee       string                   `json:"assignee"`
	Reviewer       string                   `json:"reviewer"`
	Status         string                   `json:"status"`
	Revision       int                      `json:"revision"`
}

// TaskProposal preserves the suggestion separately from a human task decision.
// Source authorship is copied from the original message, never request claims.
type TaskProposal struct {
	ID              string `json:"id"`
	CompanyID       string `json:"companyId"`
	RoomID          string `json:"roomId"`
	SourceMessageID string `json:"sourceMessageId"`
	SourceActorID   string `json:"sourceActorId"`
	SourceActorType string `json:"sourceActorType"`
	SourceAuthor    string `json:"sourceAuthor"`
	CapturedBy      string `json:"capturedBy"`
	ReviewedBy      string `json:"reviewedBy,omitempty"`
	Title           string `json:"title"`
	Criteria        string `json:"criteria"`
	Status          string `json:"status"`
	Revision        int    `json:"revision"`
	TaskID          string `json:"taskId,omitempty"`
}

// TaskDelegation is an assignment proposal, never an execution capability.
type TaskDelegation struct {
	SourceActorRevision        int       `json:"sourceActorRevision"`
	DelegateActorRevision      int       `json:"delegateActorRevision"`
	ParentOperations           uint8     `json:"parentOperations"`
	ID                         string    `json:"id"`
	CompanyID                  string    `json:"companyId"`
	RoomID                     string    `json:"roomId"`
	ParentJobID                string    `json:"parentJobId"`
	SourceParticipantID        string    `json:"sourceParticipantId"`
	SourceConnectionID         string    `json:"sourceConnectionId"`
	SourceParticipantRevision  int       `json:"sourceParticipantRevision"`
	SourceConnectionRevision   int       `json:"sourceConnectionRevision"`
	TaskID                     string    `json:"taskId"`
	TaskRevision               int       `json:"taskRevision"`
	TaskDigest                 string    `json:"taskDigest"`
	DispatchID                 string    `json:"dispatchId"`
	DelegateID                 string    `json:"delegateId"`
	DelegateConnectionID       string    `json:"delegateConnectionId"`
	DelegateRevision           int       `json:"delegateRevision"`
	DelegateConnectionRevision int       `json:"delegateConnectionRevision"`
	RepositoryDigest           string    `json:"repositoryDigest"`
	ExpiresAt                  time.Time `json:"expiresAt"`
	Epoch                      int       `json:"epoch"`
	Status                     string    `json:"status"`
	Revision                   int       `json:"revision"`
	ReviewedBy                 string    `json:"reviewedBy,omitempty"`
}

type Evidence struct {
	ID      string    `json:"id"`
	RoomID  string    `json:"roomId"`
	TaskID  string    `json:"taskId"`
	Kind    string    `json:"kind"`
	Title   string    `json:"title"`
	Body    string    `json:"body"`
	Source  string    `json:"source"`
	Created time.Time `json:"created"`
}
type Company struct {
	Revision    int    `json:"revision"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// OrganizationChart is owner-managed company metadata. Reporting relationships
// replace numbered decision seniority but do not grant permissions or tool access.
type OrganizationChart struct {
	Revision int                `json:"revision"`
	Nodes    []OrganizationNode `json:"nodes"`
}
type OrganizationNode struct {
	ActorID   string  `json:"actorId"`
	ReportsTo string  `json:"reportsTo"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
}
type Actor struct {
	HelpDisplay string   `json:"helpDisplay,omitempty"`
	Image       string   `json:"image,omitempty"`
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Title       string   `json:"title"`
	Kind        string   `json:"kind"`
	Role        string   `json:"role"`
	Enabled     bool     `json:"enabled"`
	PublicKey   string   `json:"publicKey,omitempty"`
	RoomIDs     []string `json:"roomIds"`
	Revision    int      `json:"revision"`
	ChangedBy   string   `json:"changedBy"`
}

// ManagedJob is a safe, durable execution binding. Model data and arguments
// never belong here; ShutdownConfirmed comes only from a fenced worker proof.
type ManagedJob struct {
	DelegationID         string `json:"delegationId,omitempty"`
	TaskID               string `json:"taskId,omitempty"`
	TaskDispatchID       string `json:"taskDispatchId,omitempty"`
	TaskDefinitionDigest string `json:"taskDefinitionDigest,omitempty"`
	TaskDefinitionEpoch  int    `json:"taskDefinitionEpoch,omitempty"`
	CoordinationEpoch    int    `json:"coordinationEpoch,omitempty"`
	// Optional role dependency; never an additional resource permission.
	CoordinatorID       string           `json:"coordinatorId,omitempty"`
	GovernanceRevision  int              `json:"governanceRevision,omitempty"`
	Harness             HarnessSelection `json:"harness,omitempty"`
	IdlePolicyVersion   int              `json:"idlePolicyVersion,omitempty"`
	IdlePolicyRevision  int              `json:"idlePolicyRevision,omitempty"`
	IdleShutdownEnabled bool             `json:"idleShutdownEnabled,omitempty"`
	IdleMinutes         int              `json:"idleMinutes,omitempty"`
	Execution           string           `json:"execution,omitempty"`
	RepositoryDigest    string           `json:"repositoryDigest,omitempty"`
	InitiatorID         string           `json:"initiatorId,omitempty"`
	CompanyID           string           `json:"companyId,omitempty"`
	Version             int              `json:"version"`
	JobID               string           `json:"jobId"`
	RoomID              string           `json:"roomId"`
	ParticipantID       string           `json:"participantId"`
	ConnectionID        string           `json:"connectionId"`
	Worker              string           `json:"worker"`
	Provider            string           `json:"provider"`
	Status              string           `json:"status"`
	DiagnosticCode      string           `json:"diagnosticCode"`
	AllocatedAt         time.Time        `json:"allocatedAt"`
	UpdatedAt           time.Time        `json:"updatedAt"`
	Deadline            time.Time        `json:"deadline"`
	ShutdownConfirmed   bool             `json:"shutdownConfirmed"`
}

// HarnessSelection is the immutable, nonsecret result of a signed release review.
// Zero values preserve historical allocations made before review admission.
type HarnessSelection struct {
	Sequence     int64     `json:"sequence"`
	Version      string    `json:"version"`
	Image        string    `json:"image"`
	ReviewDigest string    `json:"reviewDigest"`
	CheckedAt    time.Time `json:"checkedAt"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

type OperationReceipt struct {
	Key         string          `json:"key"`
	Fingerprint string          `json:"fingerprint"`
	Status      int             `json:"status"`
	Response    json.RawMessage `json:"response"`
}

// ToolCallAudit is replayed intent/outcome evidence for a broker-mediated tool
// call (internal/toolbroker). It is durable audit metadata only — never raw
// tool output, capability tokens or Docker launch details. Status "pending"
// means an intent was recorded with no matching outcome yet; replay never
// infers completion or retries on its own. Commit timestamps are not surfaced
// here yet; add them from the event's protected _auditRecord metadata only
// when a consumer actually needs to display them.
type ToolCallAudit struct {
	CallID        string `json:"callId"`
	JobID         string `json:"jobId"`
	RoomID        string `json:"roomId"`
	ParticipantID string `json:"participantId"`
	ConnectionID  string `json:"connectionId"`
	Provider      string `json:"provider"`
	Operation     string `json:"operation"`
	PayloadDigest string `json:"payloadDigest"`
	Status        string `json:"status"`
	OutputBytes   int    `json:"outputBytes"`
	OutputDigest  string `json:"outputDigest"`
}
type GovernanceAppointment struct {
	CompanyID     string `json:"companyId"`
	RoomID        string `json:"roomId"`
	ParticipantID string `json:"participantId"`
	Revision      int    `json:"revision"`
}

type RoomOutcomeReview struct {
	CompanyID  string    `json:"companyId"`
	RoomID     string    `json:"roomId"`
	Revision   int       `json:"revision"`
	Snapshot   string    `json:"snapshot"`
	Status     string    `json:"status"`
	Note       string    `json:"note"`
	ReviewedBy string    `json:"reviewedBy"`
	ReviewedAt time.Time `json:"reviewedAt"`
}

// TaskAcceptanceReview records an owner decision against exact returned values;
// schema checks are structural evidence, never a model's verification claim.
type TaskCriterionAssessment struct {
	CriterionID  string `json:"criterionId"`
	Method       string `json:"method"`
	OutputName   string `json:"outputName"`
	OutputDigest string `json:"outputDigest,omitempty"`
	Decision     string `json:"decision"`
	Note         string `json:"note"`
}
type TaskAcceptanceReview struct {
	CompanyID    string                    `json:"companyId"`
	RoomID       string                    `json:"roomId"`
	TaskID       string                    `json:"taskId"`
	Revision     int                       `json:"revision"`
	TaskRevision int                       `json:"taskRevision"`
	ExecutionID  string                    `json:"executionId"`
	DispatchID   string                    `json:"dispatchId"`
	Snapshot     string                    `json:"snapshot"`
	Action       string                    `json:"action"`
	Note         string                    `json:"note"`
	Assessments  []TaskCriterionAssessment `json:"assessments"`
	ReviewedBy   string                    `json:"reviewedBy"`
	ReviewedAt   time.Time                 `json:"reviewedAt"`
}

type TaskExecution struct {
	ContextPolicyEpoch             int                       `json:"contextPolicyEpoch,omitempty"`
	ConfigurationVersion           int                       `json:"configurationVersion,omitempty"`
	ParticipantRevision            int                       `json:"participantRevision,omitempty"`
	ParticipantConfigurationDigest string                    `json:"participantConfigurationDigest,omitempty"`
	ConnectionID                   string                    `json:"connectionId,omitempty"`
	ConnectionRevision             int                       `json:"connectionRevision,omitempty"`
	ConnectionConfigurationDigest  string                    `json:"connectionConfigurationDigest,omitempty"`
	InputAcceptanceEpoch           int                       `json:"inputAcceptanceEpoch,omitempty"`
	ValueDependencyEpoch           int                       `json:"valueDependencyEpoch,omitempty"`
	DelegationID                   string                    `json:"delegationId,omitempty"`
	DefinitionEpoch                int                       `json:"definitionEpoch"`
	ParentExecutionID              string                    `json:"parentExecutionId,omitempty"`
	ID                             string                    `json:"id"`
	CompanyID                      string                    `json:"companyId"`
	RoomID                         string                    `json:"roomId"`
	TaskID                         string                    `json:"taskId"`
	TaskRevision                   int                       `json:"taskRevision"`
	ParticipantID                  string                    `json:"participantId"`
	InitiatorID                    string                    `json:"initiatorId"`
	Mode                           string                    `json:"mode"`
	Definition                     taskcontract.Definition   `json:"definition"`
	DefinitionDigest               string                    `json:"definitionDigest"`
	Inputs                         []taskcontract.BoundValue `json:"inputs"`
	Outputs                        []taskcontract.BoundValue `json:"outputs"`
	Status                         string                    `json:"status"`
	ValidationError                string                    `json:"validationError,omitempty"`
	ResponseDigest                 string                    `json:"responseDigest,omitempty"`
	StartedAt                      time.Time                 `json:"startedAt"`
	FinishedAt                     time.Time                 `json:"finishedAt,omitempty"`
}

// TaskDefinitionRevision is a projection of an acknowledged task.revised event.
// EventPosition is replay order, not a timestamp or observed execution result.
type TaskDefinitionRevision struct {
	CompanyID     string                  `json:"companyId"`
	RoomID        string                  `json:"roomId"`
	TaskID        string                  `json:"taskId"`
	Definition    taskcontract.Definition `json:"definition"`
	Digest        string                  `json:"digest"`
	EventPosition int64                   `json:"eventPosition"`
}

type State struct {
	WorkspaceCaptures       map[string]WorkspaceCaptureReservation `json:"-"`
	WorkspaceProbes         map[string]WorkspaceProbeReservation   `json:"-"`
	ArtifactCustody         map[string]ArtifactCustodyObject       `json:"-"`
	TaskContextPolicyEpochs map[string]int                         `json:"-"`
	TaskRepositoryLinks     map[string]TaskRepositoryLinks         `json:"-"`
	TaskDefinitionHistory   map[string][]TaskDefinitionRevision    `json:"-"`
	TaskAcceptanceReviews   map[string][]TaskAcceptanceReview      `json:"-"`
	TaskAcceptanceEpoch     int                                    `json:"-"`
	ValueDependencyEpoch    int                                    `json:"-"`
	TaskDefinitionEpoch     int                                    `json:"-"`
	TaskExecutions          map[string]TaskExecution               `json:"-"`
	OutcomeReviews          map[string]RoomOutcomeReview           `json:"-"`
	Delegations             map[string]TaskDelegation              `json:"-"`
	DelegationEpoch         int                                    `json:"-"`
	TaskProposals           map[string]TaskProposal                `json:"-"`
	CoordinationEpoch       int                                    `json:"-"`
	Governance              map[string]GovernanceAppointment       `json:"-"`
	IdleStopNotices         map[string]IdleStopNotice              `json:"-"`
	AgentIdleSettings       AgentIdleSettings                      `json:"agentIdleSettings"`
	WorkspaceDisposals      map[string]WorkspaceDisposal           `json:"-"`
	ManagedStorage          *ManagedStorage                        `json:"-"`
	HumanRevision           int                                    `json:"humanRevision"`
	EventCount              int64                                  `json:"-"`
	OperationReceipts       map[string]OperationReceipt            `json:"-"`
	ManagedJobs             map[string]ManagedJob                  `json:"-"`
	ToolAudits              map[string]ManagedToolCall             `json:"-"`
	ToolCalls               map[string]ToolCallAudit               `json:"-"`
	Organization            OrganizationChart                      `json:"organization"`
	FeatureFlags            map[string]bool                        `json:"featureFlags"`
	AgentsStopped           bool                                   `json:"agentsStopped"`
	AgentRoomStopped        bool                                   `json:"agentRoomStopped"`
	AgentStopRooms          []string                               `json:"-"`
	Actors                  []Actor                                `json:"actors"`
	CurrentActorID          string                                 `json:"currentActorId"`
	Company                 Company                                `json:"company"`
	HumanCompanyID          string                                 `json:"humanCompanyId"`
	HumanTitle              string                                 `json:"humanTitle"`
	HumanName               string                                 `json:"humanName"`
	WorkflowRuns            []WorkflowRun                          `json:"workflowRuns"`
	Workflows               []Workflow                             `json:"workflows"`
	Rooms                   []Room                                 `json:"rooms"`
	Tasks                   []Task                                 `json:"tasks"`
	Evidence                []Evidence                             `json:"evidence"`

	Connections  []Connection  `json:"connections"`
	Participants []Participant `json:"participants"`
	Memories     []Memory      `json:"memories"`
	Messages     []Message     `json:"messages"`
	Incident     bool          `json:"incident"`
}

// AgentAdapter separates provider behavior from delivery. Capability discovery precedes invocation.
type AgentAdapter interface{ Capabilities() Capabilities }
type Capabilities struct{ Launch, Attach, Resume, Cancel, Steer bool }

func SupportsCodex(c Connection) bool {
	return c.Protocol == "CLI" && (c.Endpoint == "codex" || (strings.HasPrefix(c.Endpoint, "ssh://") && regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]*$`).MatchString(strings.TrimPrefix(c.Endpoint, "ssh://"))))
}

func SupportsAgent(c Connection) bool {
	if SupportsCodex(c) {
		return true
	}
	return c.Protocol == "CLI" && (c.Endpoint == "claude" || regexp.MustCompile(`^ssh://[A-Za-z0-9][A-Za-z0-9_.-]*/claude$`).MatchString(c.Endpoint))
}

type WorkflowNode struct {
	ID           string  `json:"id"`
	Type         string  `json:"type"`
	Title        string  `json:"title"`
	X            float64 `json:"x"`
	Y            float64 `json:"y"`
	AgentID      string  `json:"agentId"`
	Instructions string  `json:"instructions"`
	Timeout      int     `json:"timeout"`
	Retries      int     `json:"retries"`
	Evidence     string  `json:"evidence"`
}
type WorkflowEdge struct {
	ID   string `json:"id"`
	From string `json:"from"`
	To   string `json:"to"`
	Port string `json:"port"`
}
type Workflow struct {
	ID            string         `json:"id"`
	RoomID        string         `json:"roomId"`
	Name          string         `json:"name"`
	Revision      int            `json:"revision"`
	MaxIterations int            `json:"maxIterations"`
	Nodes         []WorkflowNode `json:"nodes"`
	Edges         []WorkflowEdge `json:"edges"`
}

// WorkflowRun is an append-only projection of a pinned graph and its execution.
// Outputs are provider reports, not independently verified filesystem evidence.
type WorkflowRun struct {
	ID       string         `json:"id"`
	RoomID   string         `json:"roomId"`
	Workflow Workflow       `json:"workflow"`
	Revision int            `json:"revision"`
	Status   string         `json:"status"`
	NodeID   string         `json:"nodeId"`
	Input    string         `json:"input"`
	Detail   string         `json:"detail"`
	ActorID  string         `json:"actorId"`
	Updated  time.Time      `json:"updated"`
	Visits   map[string]int `json:"visits"`
	Steps    []WorkflowStep `json:"steps"`
}

// WorkflowStep pins each completed attempt's output, hash, identity and evidence.
type WorkflowStep struct {
	NodeID    string    `json:"nodeId"`
	AgentID   string    `json:"agentId"`
	Output    string    `json:"output"`
	SHA256    string    `json:"sha256"`
	Evidence  string    `json:"evidence"`
	Port      string    `json:"port"`
	Completed time.Time `json:"completed"`
}

// AgentStatusDetails is a safe display projection, never raw provider output.
type AgentStatusDetails struct {
	Activity       string `json:"activity"`
	Reason         string `json:"reason"`
	Next           string `json:"next"`
	ElapsedSeconds int    `json:"elapsedSeconds,omitempty"`
	LongRunning    bool   `json:"longRunning,omitempty"`
}
