package main

// Controller dependencies and live coordination; eventDatabase is the mocked persistence boundary.

import (
	"agentroom/internal/room"
	"context"
	"sync"
)

type server struct {
	connectionProbe   workerProbe
	verificationRooms map[string]string
	verificationTurn  func(context.Context) (string, error)
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
}
