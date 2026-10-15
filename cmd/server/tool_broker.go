package main

import (
	"agentroom/internal/toolbroker"
	"context"
	"io"
	"sync"
	"time"
)

// newToolDispatcher assembles the worker-side broker with mandatory controller
// acknowledgements. The supervisor MUST authenticate channel and keep it separate
// from model/stdout traffic. There is no local persistence or no-op audit fallback.
// This component does not launch a pair or authenticate a transport by itself.
// A source fork CAN substitute channel, BUT invalidates durable admission.
func newToolDispatcher(ctx context.Context, scope toolbroker.Scope, allowed []string, image, dockerRoot string, now time.Time, channel io.ReadWriteCloser) (*toolbroker.Dispatcher, string, func(), error) {
	job, ok := ctxJob(ctx)
	if !ok || job != scope.JobID || channel == nil {
		return nil, "", nil, toolbroker.ErrRejected
	}
	authority, token, err := toolbroker.New(scope, allowed, now)
	if err != nil {
		return nil, "", nil, err
	}
	plan, err := toolbroker.NewToolContainerPlan(authority, image, now)
	if err != nil {
		return nil, "", nil, err
	}
	executor, err := toolbroker.NewContainerExecutor(plan, dockerCLIRunner{}, dockerRoot)
	if err != nil {
		return nil, "", nil, err
	}
	audit, err := toolbroker.NewAuditClient(channel, scope)
	if err != nil {
		return nil, "", nil, err
	}
	dispatcher, err := toolbroker.NewDispatcher(authority, audit, executor)
	if err != nil {
		return nil, "", nil, err
	}
	// Lifetime is the worker job, even if a caller later supplies a fresh context
	// to Dispatch. Cleanup is explicit, idempotent and closes blocked channel I/O.
	lifetime, cancel := context.WithDeadline(ctx, scope.Deadline)
	var once sync.Once
	closeLink := func() { once.Do(func() { authority.Revoke(); _ = channel.Close() }) }
	stop := context.AfterFunc(lifetime, func() { closeLink(); cancel() })
	cleanup := func() { stop(); closeLink(); cancel() }
	if lifetime.Err() != nil {
		cleanup()
		return nil, "", nil, toolbroker.ErrRejected
	}
	return dispatcher, token, cleanup, nil
}

func ctxJob(ctx context.Context) (string, bool) {
	if ctx == nil || ctx.Err() != nil {
		return "", false
	}
	job, ok := ctx.Value(workerLocalJobContextKey{}).(string)
	return job, ok && validWorkerJobID(job)
}
