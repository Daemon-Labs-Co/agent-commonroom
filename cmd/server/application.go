package main

// Controller dependencies and live coordination; eventDatabase is the mocked persistence boundary.

import (
	"agentroom/internal/room"
	"context"
	"encoding/json"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

type server struct {
	delegationRepository func(string) (string, error)
	taskDefinitionTurns  map[string]string
	delegationTurns      map[string]string
	coordinationTurns    map[string]room.GovernanceAppointment
	// Instance test seam; nil always uses the mandatory isolated execution path.
	agentTurn func(context.Context, room.Participant, string, string) (string, error)
	// Provider metadata boundaries; nil uses the production implementation.
	agentHistoryRPC           func(context.Context, string, any) (json.RawMessage, error)
	disposalSeen              map[string]time.Time
	disposalNow               func() time.Time
	storageTimeout            time.Duration
	providerLoginSessions     map[string]*providerLoginSession
	deviceLoginStartupTimeout time.Duration
	deviceLoginRun            func(context.Context, string, string, io.Reader, io.Writer) error
	connectionProbe           workerProbe
	verificationRooms         map[string]string
	verificationTurn          func(context.Context) (string, error)

	dailyTurnLimit    int
	taskReady         func(string) error
	taskTurn          func(context.Context, room.Participant, string, string) (string, error)
	activeJobs        map[string]string
	jobSnapshot       func(context.Context, string) (workerJobsSnapshot, error)
	jobFence          func(context.Context, string, workerFenceRequest) error
	workflowLive      map[string]context.CancelFunc
	workflowTurn      func(context.Context, room.Participant, string, string) (string, error)
	workflowReady     func(string) error
	recoveryIdle      func(context.Context, string) (bool, error)
	safetyUIDefault   bool
	stopAll           bool
	integrityFault    bool
	integrityStopPath string
	requestBudgets    map[string]actorRequestBudget
	stopRooms         map[string]bool
	templatesEnabled  bool
	db                eventDatabase
	auth              *ownerAuth
	mu                sync.Mutex
	active            map[string]bool
	cancels           map[string]context.CancelFunc
	discussions       map[string]context.CancelFunc
	extensionRegistry string
	// workspaceID names the workspace this server value belongs to ("default" for a
	// single-workspace installation). hostedExecutionBlocked refuses all managed
	// agent execution until it is namespaced per workspace.
	workspaceID            string
	limits                 *workspaceLimits // nil: unlimited (single-workspace installation)
	hostedExecutionBlocked bool
	// inFlight counts running agent turns together with the reactions that follow
	// them, so callers can tell when a room has fully settled.
	inFlight        atomic.Int64
	extensionLimits extensionLimiter
}
