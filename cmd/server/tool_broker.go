package main

import (
	"agentroom/internal/toolbroker"
	"context"
	"io"
	"time"
)

// newToolDispatcher assembles the worker-side broker with mandatory controller
// acknowledgements. The supervisor MUST authenticate channel and keep it separate
// from model/stdout traffic. There is no local persistence or no-op audit fallback.
// This component does not launch a pair or authenticate a transport by itself.
// A source fork CAN substitute channel, BUT invalidates durable admission.
func newToolDispatcher(ctx context.Context, scope toolbroker.Scope, allowed []string, image, dockerRoot string, now time.Time, channel io.ReadWriteCloser) (*toolbroker.Dispatcher, string, func(), error) {
	session, token, err := newWorkerToolSession(ctx, scope, allowed, image, dockerRoot, now, channel, dockerCLIRunner{})
	if err != nil {
		return nil, "", nil, err
	}
	return session.Dispatcher, token, session.Close, nil
}

// newWorkerToolSession preserves the worker-job check while sharing readiness,
// authority and lifetime across runtime and durable audit. Runner injection is
// trusted assembly/test work, never a public/model choice or host fallback.
func newWorkerToolSession(ctx context.Context, scope toolbroker.Scope, allowed []string, image, dockerRoot string, now time.Time, channel io.ReadWriteCloser, runner toolbroker.DockerRunner) (*toolbroker.ToolSession, string, error) {
	job, ok := ctxJob(ctx)
	if !ok || job != scope.JobID || channel == nil {
		return nil, "", toolbroker.ErrRejected
	}
	return toolbroker.NewToolSession(ctx, scope, allowed, image, dockerRoot, now, channel, runner)
}

func ctxJob(ctx context.Context) (string, bool) {
	if ctx == nil || ctx.Err() != nil {
		return "", false
	}
	job, ok := ctx.Value(workerLocalJobContextKey{}).(string)
	return job, ok && validWorkerJobID(job)
}
