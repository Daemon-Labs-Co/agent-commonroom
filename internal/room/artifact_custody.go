package room

import "agentroom/internal/objectcustody"

// ArtifactCustodyObject retains the full reservation even when physical setup,
// sealing or release is uncertain. These records never authorize file delivery
// by themselves. There is no refund, replacement, deletion or adoption transition.
type ArtifactCustodyObject struct {
	Version             int                    `json:"version"`
	Intent              objectcustody.Intent   `json:"intent"`
	ReservationID       string                 `json:"reservationId"`
	QuotaEstablished    bool                   `json:"quotaEstablished"`
	Receipt             *objectcustody.Receipt `json:"receipt,omitempty"`
	ReadAcknowledgments int64                  `json:"readAcknowledgments"`
}
