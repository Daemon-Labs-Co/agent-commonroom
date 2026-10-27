package room

import "agentroom/internal/toolbroker"

// WorkspaceCaptureReservation is private controller helper-capacity history.
// It is not a deliverable, byte-release capability or physical quota observation.
type WorkspaceCaptureReservation struct {
	Version   int                      `json:"version"`
	CompanyID string                   `json:"companyId"`
	RoomID    string                   `json:"roomId"`
	Worker    string                   `json:"worker"`
	ActorID   string                   `json:"actorId"`
	Intent    toolbroker.CaptureIntent `json:"intent"`
	Events    []WorkspaceProbeEvent    `json:"events"`
	Released  bool                     `json:"released"`
}
