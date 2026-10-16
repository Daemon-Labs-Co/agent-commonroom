package main

import (
	"agentroom/internal/toolbroker"
	"context"
	"io"
	"sync"
	"time"
)

// watchedWorkerAuditLink has one read pump. AuditClient remains the sole frame
// consumer; the pump only forwards bounded chunks through an unbuffered pipe.
// It observes idle EOF without competing with acknowledgement reads. Any link
// loss revokes the original supervisor lifetime, never releases its leases or
// claims container shutdown. This is not peer authentication: construction MUST
// follow the supervisor-owned private transport handshake. A source fork CAN
// replace the link/pump, BUT invalidates mandatory durable-audit guarantees.
type watchedWorkerAuditLink struct {
	raw    io.ReadWriteCloser
	reader *io.PipeReader
	writer *io.PipeWriter
	failed func()
	once   sync.Once
	done   chan struct{}
}

func newWatchedWorkerAuditLink(ctx context.Context, raw io.ReadWriteCloser, failed func()) *watchedWorkerAuditLink {
	reader, writer := io.Pipe()
	link := &watchedWorkerAuditLink{raw: raw, reader: reader, writer: writer, failed: failed, done: make(chan struct{})}
	stop := context.AfterFunc(ctx, link.closeIO)
	go func() {
		defer close(link.done)
		defer stop()
		// Hide optional WriteTo implementations so prefetch remains bounded even
		// when an injected transport implements an unbounded optimized copy path.
		_, _ = io.CopyBuffer(writer, struct{ io.Reader }{raw}, make([]byte, toolbroker.MaxAuditFrameBytes))
		link.closeIO()
	}()
	return link
}
func (l *watchedWorkerAuditLink) Read(p []byte) (int, error)  { return l.reader.Read(p) }
func (l *watchedWorkerAuditLink) Write(p []byte) (int, error) { return l.raw.Write(p) }
func (l *watchedWorkerAuditLink) closeIO() {
	l.once.Do(func() {
		l.raw.Close()
		l.reader.CloseWithError(toolbroker.ErrAudit)
		l.writer.CloseWithError(toolbroker.ErrAudit)
		l.failed()
	})
}
func (l *watchedWorkerAuditLink) Close() error { l.closeIO(); <-l.done; return nil }

// AcceptToolSession consumes the one supervisor-owned authenticated socket and
// builds the actual shared tool session. Scope comes from the listener previously
// opened under this supervisor's lease, never from an engine callback. Trusted
// operator image/root policy and the approved Docker runner remain required.
// The private capability is returned once, not journalled or placed in stdout.
// Production launch must still supply a reviewed engine and fence both roles
// after any failure. Cancelling is NOT shutdown or permission to discard source.
func (s *workerPairSupervisor) AcceptToolSession(ctx context.Context, listener *workerAuditListener, allowed []string, image, dockerRoot string, runner toolbroker.DockerRunner) (*toolbroker.ToolSession, string, error) {
	s.mu.Lock()
	if ctx == nil || s.available(ctx) != nil || listener == nil || listener != s.auditListener || s.leases == nil || !validWorkerAuditScope(s.record, listener.scope) {
		s.mu.Unlock()
		return nil, "", toolbroker.ErrAudit
	}
	lifetime := s.ctx
	scope := listener.scope
	s.mu.Unlock()
	// Do not hold the supervisor mutex while waiting for SSH. Close must be able
	// to cancel attachment without releasing ownership before startup unwinds.
	channel, err := listener.Accept(ctx)
	if err != nil {
		s.cancel()
		listener.Close()
		return nil, "", toolbroker.ErrAudit
	}
	watched := newWatchedWorkerAuditLink(lifetime, channel, s.cancel)
	jobCtx := context.WithValue(lifetime, workerLocalJobContextKey{}, scope.JobID)
	session, token, err := newWorkerToolSession(jobCtx, scope, allowed, image, dockerRoot, time.Now(), watched, runner)
	if err != nil {
		watched.Close()
		return nil, "", toolbroker.ErrAudit
	}
	if ctx.Err() != nil || lifetime.Err() != nil {
		session.Close()
		return nil, "", toolbroker.ErrAudit
	}
	return session, token, nil
}
