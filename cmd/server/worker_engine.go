package main

import (
	"agentroom/internal/codexbridge"
	"agentroom/internal/toolbroker"
	"context"
)

// workerCodexEngine implements the reviewed engine role only. Private socket
// setup and connection credential leasing must precede construction; a model never supplies paths,
// image, runner or Docker flags. The runtime verifies reported isolation policy.
// A source fork CAN replace these dependencies, BUT invalidates isolation. There
// is no CLI/host fallback, implicit image, automatic pull or native-tool exception.
type workerCodexEngine struct {
	scope      toolbroker.Scope
	runtime    *toolbroker.EngineContainerRuntime
	credential *workerEngineCredential
	gateway    *workerEngineGateway
}

func newWorkerCodexEngine(ctx context.Context, scope toolbroker.Scope, image string, gateway *workerEngineGateway, credential *workerEngineCredential, runner toolbroker.DockerRunner) (*workerCodexEngine, error) {
	job, ok := ctxJob(ctx)
	if !ok || job != scope.JobID || credential == nil || gateway == nil {
		return nil, toolbroker.ErrRejected
	}
	scope.Deadline = scope.Deadline.UTC().Round(0)
	if gateway.ready(ctx, scope) != nil {
		return nil, toolbroker.ErrRejected
	}
	credential.mu.Lock()
	valid := credential.scope == scope && !credential.released && !credential.provisioned && credential.original != nil && credential.original.Err() == nil
	credential.mu.Unlock()
	if !valid {
		return nil, toolbroker.ErrRejected
	}
	runtime, err := toolbroker.NewEngineContainerRuntime(ctx, scope, image, gateway.gateway.socket, runner)
	if err != nil {
		return nil, err
	}
	return &workerCodexEngine{scope: scope, runtime: runtime, credential: credential, gateway: gateway}, nil
}
func (e *workerCodexEngine) identity(v workerPairIdentity) bool {
	return v.Role == "engine" && v.Job == e.scope.JobID && v.Repository == e.scope.RepositoryDigest
}
func (e *workerCodexEngine) Create(ctx context.Context, identity workerPairIdentity) (string, error) {
	if !e.identity(identity) || e.gateway.ready(ctx, e.scope) != nil {
		return "", toolbroker.ErrRejected
	}
	return e.runtime.Create(ctx)
}
func (e *workerCodexEngine) Verify(ctx context.Context, identity workerPairIdentity, id string) error {
	if !e.identity(identity) || e.gateway.ready(ctx, e.scope) != nil {
		return toolbroker.ErrRejected
	}
	return e.runtime.VerifyCreated(ctx, id)
}
func (e *workerCodexEngine) Inspect(ctx context.Context, id string) ([]byte, error) {
	if e.gateway.ready(ctx, e.scope) != nil {
		return nil, toolbroker.ErrRejected
	}
	return e.runtime.Inspect(ctx, id)
}
func (e *workerCodexEngine) Start(ctx context.Context, id string) error {
	if e.gateway.ready(ctx, e.scope) != nil {
		return toolbroker.ErrRejected
	}
	return e.runtime.Start(ctx, id)
}

var _ workerEngineLaunch = (*workerCodexEngine)(nil)

func (e *workerCodexEngine) Scope() toolbroker.Scope { return e.scope }

// Turn connects the inspected running engine to the original tool session.
// Invoke inside supervisor.Run's turn callback so all outcomes receive fenced
// shutdown. Neither model arguments nor public requests select the stream/ID.
func (e *workerCodexEngine) Turn(ctx context.Context, s *workerPairSupervisor, session *toolbroker.ToolSession, token, prompt, model string) (codexbridge.Result, error) {
	if s == nil || session == nil || session.Scope() != e.scope {
		return codexbridge.Result{}, toolbroker.ErrRejected
	}
	s.mu.Lock()
	if ctx == nil || s.available(ctx) != nil || s.record.Status != "running" || s.toolSession != session || s.codexTurnStarted || !validWorkerAuditScope(s.record, e.scope) {
		s.mu.Unlock()
		return codexbridge.Result{}, toolbroker.ErrRejected
	}
	id := s.record.Pair.Engine.ID
	s.mu.Unlock()
	if e.gateway.ready(ctx, e.scope) != nil {
		s.cancel()
		return codexbridge.Result{}, toolbroker.ErrRejected
	}
	stream, err := e.runtime.Attach(ctx, id)
	if err != nil {
		s.cancel()
		return codexbridge.Result{}, err
	}
	defer stream.Close()
	if err = e.credential.Provision(ctx, stream); err != nil {
		s.cancel()
		return codexbridge.Result{}, err
	}
	return s.codexTurn(ctx, session, token, stream, prompt, model, func(finalize context.Context) error {
		raw, err := e.runtime.ExportCredential(finalize, id)
		if err != nil {
			return err
		}
		defer clear(raw)
		return e.credential.StageRefresh(e.scope, raw)
	})
}
