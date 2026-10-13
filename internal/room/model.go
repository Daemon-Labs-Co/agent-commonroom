package room

import (
	"regexp"
	"strings"
	"time"
)

type Connection struct {
	Location  string `json:"location"`
	CompanyID string `json:"companyId"`
	Provider  string `json:"provider"`
	ID        string `json:"id"`
	Name      string `json:"name"`
	Protocol  string `json:"protocol"`
	Endpoint  string `json:"endpoint"`
}
type Participant struct {
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
	RoomID   string `json:"roomId"`
	ID       string `json:"id"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	Scope    string `json:"scope"`
	Revision int    `json:"revision"`
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
type Room struct {
	PendingHumanRequests int    `json:"pendingHumanRequests,omitempty"`
	PersonalActorID      string `json:"personalActorId"`
	CompanyID            string `json:"companyId"`
	HumanPresent         *bool  `json:"humanPresent"`
	HumanRank            *int   `json:"humanRank"`
	HumanName            string `json:"humanName"`
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	Description          string `json:"description"`
}
type Task struct {
	ID       string `json:"id"`
	RoomID   string `json:"roomId"`
	Title    string `json:"title"`
	Criteria string `json:"criteria"`
	Assignee string `json:"assignee"`
	Reviewer string `json:"reviewer"`
	Status   string `json:"status"`
	Revision int    `json:"revision"`
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
	Version           int       `json:"version"`
	JobID             string    `json:"jobId"`
	RoomID            string    `json:"roomId"`
	ParticipantID     string    `json:"participantId"`
	ConnectionID      string    `json:"connectionId"`
	Worker            string    `json:"worker"`
	Provider          string    `json:"provider"`
	Status            string    `json:"status"`
	DiagnosticCode    string    `json:"diagnosticCode"`
	AllocatedAt       time.Time `json:"allocatedAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
	Deadline          time.Time `json:"deadline"`
	ShutdownConfirmed bool      `json:"shutdownConfirmed"`
}

type State struct {
	ManagedJobs      map[string]ManagedJob `json:"-"`
	Organization     OrganizationChart     `json:"organization"`
	FeatureFlags     map[string]bool       `json:"featureFlags"`
	AgentsStopped    bool                  `json:"agentsStopped"`
	AgentRoomStopped bool                  `json:"agentRoomStopped"`
	AgentStopRooms   []string              `json:"-"`
	Actors           []Actor               `json:"actors"`
	CurrentActorID   string                `json:"currentActorId"`
	Company          Company               `json:"company"`
	HumanCompanyID   string                `json:"humanCompanyId"`
	HumanTitle       string                `json:"humanTitle"`
	HumanName        string                `json:"humanName"`
	WorkflowRuns     []WorkflowRun         `json:"workflowRuns"`
	Workflows        []Workflow            `json:"workflows"`
	Rooms            []Room                `json:"rooms"`
	Tasks            []Task                `json:"tasks"`
	Evidence         []Evidence            `json:"evidence"`

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
