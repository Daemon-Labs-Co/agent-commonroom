package room

import "time"

// WorkspaceDisposal records a single owner-reviewed destructive intention.
// Identity binds the entire workspace, including unfinished edits. It grants no
// permission to erase history, credentials or other artifacts. A bound reservation
// may be released only with a verified outcome.
type WorkspaceDisposal struct {
	StorageReservationID string     `json:"storageReservationId,omitempty"`
	Version              int        `json:"version"`
	ID                   string     `json:"id"`
	ActorID              string     `json:"actorId"`
	Job                  ManagedJob `json:"job"`
	WorkspaceName        string     `json:"workspaceName"`
	WorkspaceNonce       string     `json:"workspaceNonce"`
	WorkspaceCreatedAt   string     `json:"workspaceCreatedAt"`
	CreatedAt            time.Time  `json:"createdAt"`
	NotBefore            time.Time  `json:"notBefore"`
	ExpiresAt            time.Time  `json:"expiresAt"`
	ReviewObservedAt     time.Time  `json:"reviewObservedAt,omitempty"`
	ClaimedAt            time.Time  `json:"claimedAt,omitempty"`
	FinishedAt           time.Time  `json:"finishedAt,omitempty"`
	Status               string     `json:"status"`
}
