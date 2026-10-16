package main

import (
	"agentroom/internal/toolbroker"
	"agentroom/internal/workspacebootstrap"
	"context"
	"errors"
	"io"
	"sync"
)

// workerEngineLaunch is a reviewed credential-bearing engine policy, never a
// model-selected command. Implementations MUST create fresh stopped containers,
// verify immutable image/mount/network/credential policy, start only by ID and
// respect cancellation. No local/SSH provider fallback is permitted. A source
// fork CAN substitute host execution here, BUT invalidates mandatory isolation.
// Codex supplies reported-policy checks; reviewed images, private mount staging
// and live acceptance are still required before production admission.
type workerEngineLaunch interface {
	Scope() toolbroker.Scope
	Create(context.Context, workerPairIdentity) (string, error)
	Verify(context.Context, workerPairIdentity, string) error
	Inspect(context.Context, string) ([]byte, error)
	Start(context.Context, string) error
}

type workerPairToolRuntime interface {
	WorkspaceIntent(context.Context) (toolbroker.WorkspaceVolumeIdentity, error)
	PrepareVolume(context.Context, toolbroker.WorkspaceVolumeIdentity) (toolbroker.WorkspaceVolumeIdentity, error)
	Create(context.Context) (string, error)
	VerifyCreated(context.Context, string) error
	Inspect(context.Context, string) ([]byte, error)
	Start(context.Context, string) error
	Bootstrap(context.Context, io.ReadCloser) (workspacebootstrap.Receipt, error)
}

// workerAssembledPair routes the credential-free tool role through the exact
// audited session's runtime. Engine and tool never share images, credentials or
// dispatchers. Bootstrap owns the approved bundle, not a model-controlled path.
// This is launch assembly, not authentication or shutdown proof. The supervisor
// retains creation journals/leases and owns fencing after any failure.
type workerAssembledPair struct {
	mu                                         sync.Mutex
	ctx                                        context.Context
	scope                                      toolbroker.Scope
	engine                                     workerEngineLaunch
	tool                                       workerPairToolRuntime
	bundle                                     io.ReadCloser
	engineID, toolID                           string
	bootstrapAttempted                         bool
	bootstrapped                               bool
	toolCreateAttempted, engineCreateAttempted bool
	toolVerified, engineVerified               bool
	toolStarted                                bool
	toolStartAttempted, engineStartAttempted   bool
}

// AssemblePair may only bind a live workspace-aware supervisor to its exact
// private audited session. It does not expose a public launch or relax admission.
// The caller owns session lifetime and must close it when the supervisor ends.
func (s *workerPairSupervisor) AssemblePair(session *toolbroker.ToolSession, engine workerEngineLaunch, bundle io.ReadCloser) (*workerAssembledPair, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.available(s.ctx) != nil || s.assembled || session == nil || session.Runtime == nil || session.Dispatcher == nil || engine == nil || bundle == nil || engine.Scope() != session.Scope() || !validWorkerAuditScope(s.record, session.Scope()) {
		return nil, errors.New("pair assembly binding refused")
	}
	s.assembled = true
	s.toolSession = session
	// Once accepted, close an unconsumed snapshot even if startup is cancelled
	// before Bootstrap. session.Close revokes the shared dispatch authority.
	input := &workerPairBundle{ReadCloser: bundle}
	context.AfterFunc(s.ctx, func() { input.Close(); session.Close() })
	return &workerAssembledPair{ctx: s.ctx, scope: session.Scope(), engine: engine, tool: session.Runtime, bundle: input}, nil
}

type workerPairBundle struct {
	io.ReadCloser
	once sync.Once
	err  error
}

func (b *workerPairBundle) Close() error {
	b.once.Do(func() { b.err = b.ReadCloser.Close() })
	return b.err
}

func (a *workerAssembledPair) context(ctx context.Context) (context.Context, func(), error) {
	if ctx == nil || ctx.Err() != nil || a.ctx.Err() != nil {
		return nil, nil, errors.New("pair assembly expired")
	}
	child, cancel := context.WithDeadline(ctx, a.scope.Deadline)
	stop := context.AfterFunc(a.ctx, cancel)
	return child, func() { stop(); cancel() }, nil
}
func (a *workerAssembledPair) identity(v workerPairIdentity) bool {
	return v.Job == a.scope.JobID && v.Repository == a.scope.RepositoryDigest && (v.Role == "tool" || v.Role == "engine")
}
func (a *workerAssembledPair) WorkspaceIntent(ctx context.Context) (toolbroker.WorkspaceVolumeIdentity, error) {
	child, done, err := a.context(ctx)
	if err != nil {
		return toolbroker.WorkspaceVolumeIdentity{}, err
	}
	defer done()
	return a.tool.WorkspaceIntent(child)
}
func (a *workerAssembledPair) PrepareVolume(ctx context.Context, intent toolbroker.WorkspaceVolumeIdentity) (toolbroker.WorkspaceVolumeIdentity, error) {
	child, done, err := a.context(ctx)
	if err != nil {
		return toolbroker.WorkspaceVolumeIdentity{}, err
	}
	defer done()
	return a.tool.PrepareVolume(child, intent)
}
func (a *workerAssembledPair) Create(ctx context.Context, identity workerPairIdentity) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	child, done, err := a.context(ctx)
	if err != nil {
		return "", err
	}
	defer done()
	if !a.identity(identity) {
		return "", errors.New("pair role scope refused")
	}
	var id string
	if identity.Role == "tool" {
		if a.toolCreateAttempted {
			return "", errors.New("tool already created")
		}
		a.toolCreateAttempted = true
		id, err = a.tool.Create(child)
	} else {
		if a.engineCreateAttempted {
			return "", errors.New("engine already created")
		}
		a.engineCreateAttempted = true
		id, err = a.engine.Create(child, identity)
	}
	if err != nil || !validPairContainerID(id) || id == a.toolID || id == a.engineID || child.Err() != nil {
		return "", errors.New("pair create unconfirmed")
	}
	if identity.Role == "tool" {
		a.toolID = id
	} else {
		a.engineID = id
	}
	return id, nil
}
func (a *workerAssembledPair) Verify(ctx context.Context, identity workerPairIdentity, id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	child, done, err := a.context(ctx)
	if err != nil {
		return err
	}
	defer done()
	if !a.identity(identity) || !validPairContainerID(id) {
		return errors.New("pair verify scope refused")
	}
	if identity.Role == "tool" && id == a.toolID {
		err = a.tool.VerifyCreated(child, id)
		a.toolVerified = err == nil && child.Err() == nil
		if !a.toolVerified {
			return errors.New("tool verification refused")
		}
		return nil
	}
	if identity.Role == "engine" && id == a.engineID {
		err = a.engine.Verify(child, identity, id)
		a.engineVerified = err == nil && child.Err() == nil
		if !a.engineVerified {
			return errors.New("engine verification refused")
		}
		return nil
	}
	return errors.New("pair role identity refused")
}
func (a *workerAssembledPair) Inspect(ctx context.Context, id string) ([]byte, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	child, done, err := a.context(ctx)
	if err != nil {
		return nil, err
	}
	defer done()
	if !validPairContainerID(id) {
		return nil, errors.New("pair inspection identity refused")
	}
	if id == a.toolID {
		return a.tool.Inspect(child, id)
	}
	if id == a.engineID {
		return a.engine.Inspect(child, id)
	}
	return nil, errors.New("unknown pair identity")
}
func (a *workerAssembledPair) Start(ctx context.Context, id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	child, done, err := a.context(ctx)
	if err != nil {
		return err
	}
	defer done()
	if !validPairContainerID(id) {
		return errors.New("pair start identity refused")
	}
	if id == a.toolID && a.toolVerified && a.engineVerified && !a.toolStartAttempted {
		a.toolStartAttempted = true
		err = a.tool.Start(child, id)
		a.toolStarted = err == nil && child.Err() == nil
		if !a.toolStarted {
			return errors.New("tool startup unconfirmed")
		}
		return nil
	}
	if id == a.engineID && a.bootstrapped && a.engineVerified && !a.engineStartAttempted {
		a.engineStartAttempted = true
		err = a.engine.Start(child, id)
		if err != nil || child.Err() != nil {
			return errors.New("engine startup unconfirmed")
		}
		return nil
	}
	return errors.New("pair startup order refused")
}
func (a *workerAssembledPair) Bootstrap(ctx context.Context, id string) (workspacebootstrap.Receipt, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	child, done, err := a.context(ctx)
	if err != nil {
		return workspacebootstrap.Receipt{}, err
	}
	defer done()
	if id != a.toolID || !validPairContainerID(id) || !a.toolStarted || a.bootstrapAttempted {
		return workspacebootstrap.Receipt{}, errors.New("pair bootstrap identity refused")
	}
	a.bootstrapAttempted = true
	defer a.bundle.Close()
	receipt, err := a.tool.Bootstrap(child, a.bundle)
	if err != nil || child.Err() != nil || !workspacebootstrap.ValidReceipt(receipt, a.scope.JobID, a.scope.RepositoryDigest) {
		return workspacebootstrap.Receipt{}, errors.New("pair bootstrap unconfirmed")
	}
	a.bootstrapped = true
	return receipt, nil
}
