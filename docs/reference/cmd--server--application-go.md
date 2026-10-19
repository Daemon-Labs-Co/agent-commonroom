# cmd/server/application.go

[Module index](cmd.md) · [Open source](../../cmd/server/application.go)

Go implementation for application; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `server`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { disposalSeen map[string]time.Time disposalNow func() time.Time storageTimeout time.Duration connectionProbe workerProbe verificationRooms map[string]string verificationTurn func(context.Context) (string, error) dailyTurnLimit int taskReady func(string) error taskTurn func(context.Context, room.Participant, string, string) (string, error) activeJobs map[string]string jobSnapshot func(context.Context, string) (workerJobsSnapshot, error) jobFence func(context.Context, string, workerFenceRequest) error workflowLive map[string]context.CancelFunc workflowTurn func(context.Context, room.Participant, string, string) (string, error) workflowReady func(string) error recoveryIdle func(context.Context, string) (bool, error) safetyUIDefault bool stopAll bool integrityFault bool integrityStopPath string requestBudgets map[string]actorRequestBudget stopRooms map[string]bool templatesEnabled bool db eventDatabase auth *ownerAuth mu sync.Mutex active map[string]bool cancels map[string]context.CancelFunc }`.
