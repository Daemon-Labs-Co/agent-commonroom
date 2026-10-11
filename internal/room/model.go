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
	ActorID           string `json:"actorId"`
	CompanyID         string `json:"companyId"`
	Discipline        string `json:"discipline"`
	Level             string `json:"level"`
	Speciality        string `json:"speciality"`
	Assignment        string `json:"assignment"`
	Challenge         *int   `json:"challenge"`
	RoomID            string `json:"roomId"`
	ID                string `json:"id"`
	ConnectionID      string `json:"connectionId"`
	Name              string `json:"name"`
	Role              string `json:"role"`
	Instructions      string `json:"instructions"`
	Rank              int    `json:"rank"`
	Accept            bool   `json:"accept"`
	AdditionalContext string `json:"additionalContext"`
	ContextSource     string `json:"contextSource"`
	ThreadID          string `json:"threadId"`
	Status            string `json:"status"`
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
	ActorID   string    `json:"actorId"`
	ActorType string    `json:"actorType"`
	Provider  string    `json:"provider,omitempty"`
	RoomID    string    `json:"roomId"`
	ID        string    `json:"id"`
	Author    string    `json:"author"`
	Text      string    `json:"text"`
	Created   time.Time `json:"created"`
}
type Room struct {
	PersonalActorID string `json:"personalActorId"`
	CompanyID       string `json:"companyId"`
	HumanPresent    *bool  `json:"humanPresent"`
	HumanRank       *int   `json:"humanRank"`
	HumanName       string `json:"humanName"`
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
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
type Actor struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Title     string   `json:"title"`
	Kind      string   `json:"kind"`
	Role      string   `json:"role"`
	Enabled   bool     `json:"enabled"`
	PublicKey string   `json:"publicKey,omitempty"`
	RoomIDs   []string `json:"roomIds"`
	Revision  int      `json:"revision"`
	ChangedBy string   `json:"changedBy"`
}
type State struct {
	Actors         []Actor    `json:"actors"`
	CurrentActorID string     `json:"currentActorId"`
	Company        Company    `json:"company"`
	HumanCompanyID string     `json:"humanCompanyId"`
	HumanTitle     string     `json:"humanTitle"`
	HumanName      string     `json:"humanName"`
	Workflows      []Workflow `json:"workflows"`
	Rooms          []Room     `json:"rooms"`
	Tasks          []Task     `json:"tasks"`
	Evidence       []Evidence `json:"evidence"`

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
