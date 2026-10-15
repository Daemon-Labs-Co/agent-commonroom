package room

import "time"

// ManagedToolCall is the durable metadata projection of a scoped broker action.
// Raw actions, model output, capabilities and provider credentials never belong here.
type ManagedToolCall struct {
	JobID            string    `json:"jobId"`
	CallID           string    `json:"callId"`
	RoomID           string    `json:"roomId"`
	ParticipantID    string    `json:"participantId"`
	ConnectionID     string    `json:"connectionId"`
	CompanyID        string    `json:"companyId"`
	ActorID          string    `json:"actorId"`
	Provider         string    `json:"provider"`
	RepositoryDigest string    `json:"repositoryDigest"`
	Deadline         time.Time `json:"deadline"`
	Operation        string    `json:"operation"`
	PayloadDigest    string    `json:"payloadDigest"`
	Status           string    `json:"status"`
	OutputBytes      int       `json:"outputBytes"`
	OutputDigest     string    `json:"outputDigest,omitempty"`
	RecordedAt       time.Time `json:"recordedAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}
