package room

import "agentroom/internal/toolbroker"

// WorkspaceProbeReservation is private controller audit/capacity state, never
// source-read authority. A worker's local journal remains independently required.
type WorkspaceProbeEvent struct {
	Stage  string `json:"stage"`
	ID     string `json:"id"`
	Absent bool   `json:"absent"`
}
type WorkspaceProbeReservation struct {
	Version   int                             `json:"version"`
	CompanyID string                          `json:"companyId"`
	RoomID    string                          `json:"roomId"`
	Worker    string                          `json:"worker"`
	ActorID   string                          `json:"actorId"`
	Intent    toolbroker.WorkspaceProbeIntent `json:"intent"`
	Events    []WorkspaceProbeEvent           `json:"events"`
	Released  bool                            `json:"released"`
}
