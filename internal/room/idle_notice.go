package room

import "time"

// IdleStopNotice records an admitted worker's intention, never shutdown proof.
// It contains fixed lifecycle metadata only, never model text or credentials.
type IdleStopNotice struct {
	Version         int        `json:"version"`
	ID              string     `json:"id"`
	Job             ManagedJob `json:"job"`
	WorkerStartedAt time.Time  `json:"workerStartedAt"`
	RequestedAt     time.Time  `json:"requestedAt"`
	RecordedAt      time.Time  `json:"recordedAt"`
}
