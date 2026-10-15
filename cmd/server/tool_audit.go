package main

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"context"
	"errors"
	"io"
	"os"
	"time"
)

// controllerToolAuditor owns the durable side of the supervisor's callback link.
// Construction is trusted controller work, not a model/plugin/public HTTP API.
// A source fork CAN substitute a no-op auditor BUT then tools are unaudited.
type controllerToolAuditor struct {
	server              *server
	scope               toolbroker.Scope
	actorID             string
	actor               room.Actor
	participantRevision int
	connection          room.Connection
	companyID           string
	repository          func() (string, error)
}

// serveToolAudit is the controller assembly point for a registered supervisor
// stream, not a public handler. SSH/peer authentication and engine/control-stream
// separation are required before calling it. An absent identity never means owner.
func (s *server) serveToolAudit(ctx context.Context, scope toolbroker.Scope, channel io.ReadWriteCloser) error {
	if channel == nil {
		return toolbroker.ErrAudit
	}
	auditor, err := s.toolAuditor(ctx, scope)
	if err != nil {
		_ = channel.Close()
		return toolbroker.ErrAudit
	}
	return toolbroker.ServeAudit(ctx, channel, scope, auditor)
}

func (s *server) toolAuditor(ctx context.Context, scope toolbroker.Scope) (toolbroker.Auditor, error) {
	if ctx == nil {
		return nil, toolbroker.ErrAudit
	}
	actor, _ := ctx.Value(actorContextKey{}).(string)
	registration, ok := ctx.Value(managedJobContextKey{}).(managedJobRegistration)
	if !ok || registration.JobID != scope.JobID {
		return nil, toolbroker.ErrAudit
	}
	return newControllerToolAuditor(ctx, s, scope, actor, func() (string, error) {
		return controllerDevelopmentRepository(scope.Room, os.Getenv("COMMONROOM_DEVELOPMENT_WORKSPACES"), os.Getenv("COMMONROOM_DEVELOPMENT_REPOSITORIES"))
	})
}

func newControllerToolAuditor(ctx context.Context, s *server, scope toolbroker.Scope, actor string, repository func() (string, error)) (*controllerToolAuditor, error) {
	if ctx == nil || s == nil || actor == "" || repository == nil || !toolbroker.ValidAuditScope(scope, time.Now()) {
		return nil, toolbroker.ErrAudit
	}
	a := &controllerToolAuditor{server: s, scope: scope, actorID: actor, repository: repository}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, principal, participant, connection, err := a.check(ctx, false)
	if err != nil {
		return nil, err
	}
	a.actor, a.participantRevision, a.connection, a.companyID = principal, participant.Revision, connection, st.Company.ID
	return a, nil
}

// check revalidates current membership, connection revision, repository policy,
// job lifetime and stop state. The initiating human never becomes the owner just
// because the selected provider credentials belong to the organization owner.
// Caller holds server.mu; persistence additionally checks the event sequence.
func (a *controllerToolAuditor) check(ctx context.Context, bound bool) (room.State, room.Actor, room.Participant, room.Connection, error) {
	var principal room.Actor
	var participant room.Participant
	var connection room.Connection
	st, err := a.server.load(ctx)
	if err != nil {
		return st, principal, participant, connection, toolbroker.ErrAudit
	}
	repository, err := a.repository()
	job, exists := st.ManagedJobs[a.scope.JobID]
	if job.Version == workerPairLifecycleVersion && job.RepositoryDigest != a.scope.RepositoryDigest {
		return st, principal, participant, connection, toolbroker.ErrAudit
	}
	if ctx.Err() != nil || err != nil || repository == "" || repository != a.scope.RepositoryDigest || !exists || job.InitiatorID != a.actorID || job.CompanyID != st.Company.ID || job.Status != "submitted" || job.ShutdownConfirmed || job.RoomID != a.scope.Room || job.ParticipantID != a.scope.Participant || job.ConnectionID != a.scope.Connection || job.Provider != a.scope.Provider || !job.Deadline.Equal(a.scope.Deadline) || !a.scope.Deadline.After(time.Now()) || a.server.agentsBlockedLocked(st, a.scope.Room) {
		return st, principal, participant, connection, toolbroker.ErrAudit
	}
	principal, exists = member(st, a.actorID)
	if !exists || principal.Kind != "human" || !canAccess(st, a.actorID, a.scope.Room) {
		return st, principal, participant, connection, toolbroker.ErrAudit
	}
	for _, p := range st.Participants {
		if p.ID == a.scope.Participant && p.RoomID == a.scope.Room && p.ConnectionID == a.scope.Connection {
			participant = p
		}
	}
	for _, c := range st.Connections {
		if c.ID == a.scope.Connection {
			connection = c
		}
	}
	roomExists := false
	for _, r := range st.Rooms {
		if r.ID == a.scope.Room && r.CompanyID == st.Company.ID {
			roomExists = true
		}
	}
	alias, provider, endpointErr := remoteWorkerEndpoint(connection.Endpoint, job.Worker, "true")
	if !roomExists || participant.ID == "" || connection.Protocol != "CLI" || participant.CompanyID != st.Company.ID || connection.CompanyID != st.Company.ID || endpointErr != nil || alias != job.Worker || provider != job.Provider {
		return st, principal, participant, connection, toolbroker.ErrAudit
	}
	if _, _, err := contextDestination(st, participant); err != nil {
		return st, principal, participant, connection, toolbroker.ErrAudit
	}
	if bound && (st.Company.ID != a.companyID || principal.Revision != a.actor.Revision || principal.Role != a.actor.Role || participant.Revision != a.participantRevision || connection != a.connection) {
		return st, principal, participant, connection, toolbroker.ErrAudit
	}
	return st, principal, participant, connection, nil
}

func (a *controllerToolAuditor) RecordIntent(ctx context.Context, intent toolbroker.Intent) error {
	if ctx == nil || intent.Scope != a.scope || !toolbroker.ValidAuditIntent(intent, time.Now()) {
		return toolbroker.ErrAudit
	}
	a.server.mu.Lock()
	defer a.server.mu.Unlock()
	st, _, _, _, err := a.check(ctx, true)
	if err != nil {
		return err
	}
	if _, exists := st.ToolAudits[toolAuditKey(intent.Scope.JobID, intent.CallID)]; exists {
		return toolbroker.ErrAudit
	}
	count := 0
	for _, call := range st.ToolAudits {
		if call.JobID == a.scope.JobID {
			if call.Status == "pending" {
				return toolbroker.ErrAudit
			}
			count++
		}
	}
	if count >= toolbroker.MaxCalls {
		return toolbroker.ErrAudit
	}
	now := time.Now().UTC()
	call := room.ManagedToolCall{JobID: a.scope.JobID, CallID: intent.CallID, RoomID: a.scope.Room, ParticipantID: a.scope.Participant, ConnectionID: a.scope.Connection, CompanyID: st.Company.ID, ActorID: a.actorID, Provider: a.scope.Provider, RepositoryDigest: a.scope.RepositoryDigest, Deadline: a.scope.Deadline, Operation: intent.Operation, PayloadDigest: intent.PayloadDigest, Status: "pending", RecordedAt: now, UpdatedAt: now}
	return a.persist(ctx, st, "worker.tool.intent", call)
}

func (a *controllerToolAuditor) RecordOutcome(ctx context.Context, intent toolbroker.Intent, outcome toolbroker.Outcome) error {
	if ctx == nil || intent.Scope != a.scope || !toolbroker.ValidAuditOutcome(outcome) {
		return toolbroker.ErrAudit
	}
	a.server.mu.Lock()
	defer a.server.mu.Unlock()
	st, _, _, _, err := a.check(ctx, true)
	if err != nil {
		return err
	}
	call, exists := st.ToolAudits[toolAuditKey(intent.Scope.JobID, intent.CallID)]
	if !exists || call.Status != "pending" || toolAuditIntent(call) != intent || call.ActorID != a.actorID || call.CompanyID != a.companyID {
		return toolbroker.ErrAudit
	}
	call.Status, call.OutputBytes, call.OutputDigest, call.UpdatedAt = outcome.Status, outcome.OutputBytes, outcome.OutputDigest, time.Now().UTC()
	return a.persist(ctx, st, "worker.tool.outcome", call)
}

func (a *controllerToolAuditor) persist(ctx context.Context, st room.State, kind string, call room.ManagedToolCall) error {
	ctx = context.WithValue(ctx, actorContextKey{}, a.actorID)
	ctx = context.WithValue(ctx, expectedEventCountKey{}, st.EventCount)
	if err := a.server.append(ctx, kind, call); err != nil {
		if !errors.Is(err, errAuditConflict) {
			a.server.tripIntegrityLocked()
		}
		return toolbroker.ErrAudit
	}
	return nil
}
