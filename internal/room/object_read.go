package room

import "agentroom/internal/toolbroker"

// ObjectReadReservation is private controller capacity history, not a byte-release
// capability, physical observation, delivered artifact or task acceptance.
type ObjectReadReservation struct {
	Version   int                         `json:"version"`
	CompanyID string                      `json:"companyId"`
	RoomID    string                      `json:"roomId"`
	Worker    string                      `json:"worker"`
	ActorID   string                      `json:"actorId"`
	Intent    toolbroker.ObjectReadIntent `json:"intent"`
	Events    []WorkspaceProbeEvent       `json:"events"`
	Released  bool                        `json:"released"`
}
