package room

import "time"

type TaskRepositoryLink struct {
	Kind string `json:"kind"`
	URL  string `json:"url"`
}

// These are human-maintained references, never repository or execution authority.
type TaskRepositoryLinks struct {
	Version      int                  `json:"version"`
	CompanyID    string               `json:"companyId"`
	RoomID       string               `json:"roomId"`
	TaskID       string               `json:"taskId"`
	TaskRevision int                  `json:"taskRevision"`
	Revision     int                  `json:"revision"`
	Links        []TaskRepositoryLink `json:"links"`
	ChangedBy    string               `json:"changedBy,omitempty"`
	ChangedAt    time.Time            `json:"changedAt,omitempty"`
}
