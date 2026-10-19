package main

import (
	"agentroom/internal/codexbridge"
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"context"
	"errors"
	"io"
)

// All paths/images are trusted worker deployment configuration. Prompt/model
// still require the controller's normal authorization and feature admission.
// No launch flag or model argument can replace isolation, runner or gateway.
type workerCodexRunConfig struct {
	// Trusted authenticated controller allocation, never a model-selected override.
	ManagedJob                         *room.ManagedJob
	IdleAnnounce                       workerIdleAnnounce
	EngineImage, AuthRoot, GatewayRoot string
	Prompt, Model                      string
}

// RunCodex owns one already-admitted supervisor/session/bundle through credential
// acquisition, private gateway creation, paired assembly, callback turn and fenced
// shutdown. It exposes no public endpoint and does not grant admission or audit
// authority. A source fork CAN bypass it, BUT loses the isolation contract.
// Validation refusal retains caller ownership; after admission all exit paths,
// including panic, close owned resources and reconcile container identities.
func (s *workerPairSupervisor) RunCodex(ctx context.Context, session *toolbroker.ToolSession, token string, bundle io.ReadCloser, config workerCodexRunConfig, runner toolbroker.DockerRunner) (result codexbridge.Result, record workerJobRecord, err error) {
	if config.ManagedJob == nil && config.IdleAnnounce != nil {
		return result, record, toolbroker.ErrRejected
	}
	if ctx == nil || ctx.Err() != nil || session == nil || bundle == nil || runner == nil || !repositoryDigestPattern.MatchString(token) {
		return result, record, toolbroker.ErrRejected
	}
	s.mu.Lock()
	if s.available(ctx) != nil || s.codexRunStarted || s.assembled || s.engineCredential != nil || session.Runtime == nil || session.Dispatcher == nil || session.Scope().Provider != "codex" || !validWorkerAuditScope(s.record, session.Scope()) {
		s.mu.Unlock()
		return result, record, toolbroker.ErrRejected
	}
	if config.ManagedJob != nil {
		job := *config.ManagedJob
		// Release mu before binding the same immutable admission through its helper.
		s.mu.Unlock()
		if err = s.configureJobIdle(job, runner, config.IdleAnnounce); err != nil {
			return result, record, err
		}
		s.mu.Lock()
		if s.available(ctx) != nil || s.codexRunStarted || s.assembled {
			s.mu.Unlock()
			return result, record, toolbroker.ErrRejected
		}
	}
	s.codexRunStarted = true
	scope, journal := session.Scope(), s.dir
	s.mu.Unlock()
	bound := context.WithValue(ctx, workerLocalJobContextKey{}, scope.JobID)
	// Resource admission retains the original supervisor context, not only a
	// derived request context whose cancellation may still be propagating.
	setup := context.WithValue(s.ctx, workerLocalJobContextKey{}, scope.JobID)
	lifetime, cancel := context.WithDeadline(bound, scope.Deadline)
	stop := context.AfterFunc(s.ctx, cancel)
	input := &workerPairBundle{ReadCloser: bundle}
	var credential *workerEngineCredential
	var gateway *workerEngineGateway
	var engine *workerCodexEngine
	ran := false
	defer func() {
		cancel()
		stop()
		if !ran {
			var shutdownErr error
			record, shutdownErr = s.Shutdown(context.Background(), runner)
			err = errors.Join(err, shutdownErr)
		} else {
			// Run performs shutdown even when its callback panics.
			s.mu.Lock()
			record = s.record
			s.mu.Unlock()
		}
		if engine != nil {
			engine.runtime.Close()
		}
		session.Close()
		err = errors.Join(err, input.Close())
		if gateway != nil {
			err = errors.Join(err, gateway.Close())
		}
		if credential != nil {
			custodyErr := credential.Release(record)
			if custodyErr != nil {
				result = codexbridge.Result{}
			}
			err = errors.Join(err, custodyErr)
		}
	}()
	credential, err = acquireWorkerEngineCredential(setup, scope, config.AuthRoot, journal)
	if err != nil {
		return result, record, err
	}
	s.mu.Lock()
	if s.available(lifetime) != nil || s.engineCredential != nil {
		s.mu.Unlock()
		return result, record, toolbroker.ErrRejected
	}
	// Keep an unresolved lease reachable for a later Shutdown/recovery attempt.
	s.engineCredential = credential
	s.mu.Unlock()
	gateway, err = newWorkerEngineGateway(setup, scope, config.GatewayRoot)
	if err != nil {
		return result, record, err
	}
	engine, err = newWorkerCodexEngine(setup, scope, config.EngineImage, gateway, credential, runner)
	if err != nil {
		return result, record, err
	}
	adapter, err := s.AssemblePair(session, engine, input)
	if err != nil {
		return result, record, err
	}
	ran = true
	record, err = s.Run(lifetime, adapter, runner, func(turn context.Context) error {
		var turnErr error
		result, turnErr = engine.Turn(turn, s, session, token, config.Prompt, config.Model)
		return turnErr
	})
	return result, record, err
}
