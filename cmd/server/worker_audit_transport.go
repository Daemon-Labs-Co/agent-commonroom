package main

import (
	"agentroom/internal/toolbroker"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// The audit socket is supervisor-only: never mount its directory into either
// container. SSH authenticates the approved worker account; that account and its
// Docker daemon remain trusted. This is not protection from a hostile host user
// with the same UID. A source fork CAN expose this socket, BUT loses audit trust.
// There is exactly one accepted link per supervisor lifetime; reconnect needs a
// new job and reconciliation, never replay of an uncertain audit outcome.
type workerAuditListener struct {
	listener *net.UnixListener
	parent   context.Context
	scope    toolbroker.Scope
	stop     func() bool
	mu       sync.Mutex
	accepted bool
	closed   bool
	channel  io.ReadWriteCloser
}

func workerAuditSocket(dir, job string) string {
	sum := sha256.Sum256([]byte(job))
	return filepath.Join(dir, ".audit-"+hex.EncodeToString(sum[:12])+".sock")
}

func newWorkerAuditListener(ctx context.Context, dir string, scope toolbroker.Scope) (*workerAuditListener, error) {
	job, ok := ctxJob(ctx)
	fi, err := os.Lstat(dir)
	if !ok || job != scope.JobID || !toolbroker.ValidAuditScope(scope, time.Now()) || err != nil || !fi.IsDir() || fi.Mode()&os.ModeSymlink != 0 || fi.Mode().Perm()&0077 != 0 {
		return nil, toolbroker.ErrAudit
	}
	// Burn a durable, private one-shot marker before creating the socket. A
	// dropped SSH link, failed bind, or supervisor crash cannot reset audit state.
	marker, err := os.OpenFile(filepath.Join(dir, ".audit-"+job+".used"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, toolbroker.ErrAudit
	}
	err = marker.Sync()
	closeErr := marker.Close()
	if err != nil || closeErr != nil || syncWorkerDirectory(dir) != nil {
		return nil, toolbroker.ErrAudit
	}
	// ListenUnix refuses an existing path. Never unlink a previous job's socket.
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: workerAuditSocket(dir, job), Net: "unix"})
	if err != nil {
		return nil, toolbroker.ErrAudit
	}
	if err = listener.SetDeadline(scope.Deadline); err != nil {
		listener.Close()
		return nil, toolbroker.ErrAudit
	}
	if err = os.Chmod(listener.Addr().String(), 0600); err != nil {
		listener.Close()
		return nil, toolbroker.ErrAudit
	}
	l := &workerAuditListener{listener: listener, scope: scope, parent: ctx}
	l.stop = context.AfterFunc(ctx, func() { listener.Close() })
	return l, nil
}

func (l *workerAuditListener) Close() error {
	l.mu.Lock()
	l.closed = true
	channel := l.channel
	l.mu.Unlock()
	if l.stop != nil {
		l.stop()
	}
	err := l.listener.Close()
	if channel != nil {
		channel.Close()
	}
	return err
}

func auditScopeProof(scope toolbroker.Scope) []byte {
	raw, _ := json.Marshal(scope)
	sum := sha256.Sum256(raw)
	return sum[:]
}

func (l *workerAuditListener) Accept(ctx context.Context) (io.ReadWriteCloser, error) {
	l.mu.Lock()
	if l.accepted || l.closed || ctx == nil || ctx.Err() != nil {
		l.mu.Unlock()
		return nil, toolbroker.ErrAudit
	}
	l.accepted = true
	l.mu.Unlock()
	stop := context.AfterFunc(ctx, func() { l.listener.Close() })
	defer stop()
	conn, err := l.listener.AcceptUnix()
	if err != nil {
		return nil, toolbroker.ErrAudit
	}
	// A failed handshake also consumes this link. A peer cannot reset audit state.
	l.listener.Close()
	stopAccept := context.AfterFunc(ctx, func() { conn.Close() })
	stopParent := context.AfterFunc(l.parent, func() { conn.Close() })
	stopConn := func() bool { a := stopAccept(); b := stopParent(); return a && b }
	owned := &workerAuditConn{Conn: conn, stop: stopConn}
	l.mu.Lock()
	if l.closed {
		l.mu.Unlock()
		owned.Close()
		return nil, toolbroker.ErrAudit
	}
	l.channel = owned
	l.mu.Unlock()

	deadline := time.Now().Add(5 * time.Second)
	if l.scope.Deadline.Before(deadline) {
		deadline = l.scope.Deadline
	}
	conn.SetDeadline(deadline)
	proof := make([]byte, 32)
	_, err = io.ReadFull(conn, proof)
	expected := auditScopeProof(l.scope)
	if err == nil && !equalAuditProof(proof, expected) {
		err = toolbroker.ErrAudit
	}
	if err == nil {
		_, err = conn.Write([]byte{1})
	}
	if err != nil || ctx.Err() != nil || l.parent.Err() != nil {
		stopConn()
		conn.Close()
		return nil, toolbroker.ErrAudit
	}
	conn.SetDeadline(l.scope.Deadline)
	return owned, nil
}

func equalAuditProof(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte
	for i := range a {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}

type workerAuditConn struct {
	net.Conn
	stop func() bool
}

func (c *workerAuditConn) Close() error { c.stop(); return c.Conn.Close() }

func validWorkerAuditScope(record workerJobRecord, scope toolbroker.Scope) bool {
	return validWorkerRecord(record) && record.Version == workerWorkspaceLifecycleVersion && !workerJobSettled(record) && !workerTerminal(record.Status) &&
		record.JobID == scope.JobID && record.RoomID == scope.Room && record.ParticipantID == scope.Participant && record.ConnectionID == scope.Connection && record.Provider == scope.Provider && record.Pair.Repository == scope.RepositoryDigest && record.Deadline.Equal(scope.Deadline) && toolbroker.ValidAuditScope(scope, time.Now())
}

func decodeWorkerAuditScope(encoded string) (toolbroker.Scope, error) {
	var scope toolbroker.Scope
	if len(encoded) > 4096 {
		return scope, toolbroker.ErrAudit
	}
	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil || json.Unmarshal(raw, &scope) != nil || !toolbroker.ValidAuditScope(scope, time.Now()) {
		return toolbroker.Scope{}, toolbroker.ErrAudit
	}
	// Exact canonical encoding rejects unknown, duplicate, omitted and alternative
	// spellings before the request can select a socket.
	canonical, _ := json.Marshal(scope)
	if string(canonical) != string(raw) {
		return toolbroker.Scope{}, toolbroker.ErrAudit
	}
	return scope, nil
}

// workerAttachAudit forwards only a registered private audit socket, over its
// own SSH session. It never executes provider commands or consumes model output.
func workerAttachAudit(ctx context.Context, dir string, scope toolbroker.Scope, in io.Reader, out io.Writer) error {
	record, err := readWorkerRecord(dir, scope.JobID)
	if err != nil || !validWorkerAuditScope(record, scope) {
		return toolbroker.ErrAudit
	}
	lease, free, err := workerTryLease(dir, ".job-"+scope.JobID+".lease")
	if lease != nil {
		lease.Close()
	}
	if err != nil || free {
		return toolbroker.ErrAudit
	}
	fi, err := os.Lstat(dir)
	if err != nil || !fi.IsDir() || fi.Mode().Perm()&0077 != 0 {
		return toolbroker.ErrAudit
	}
	conn, err := (&net.Dialer{}).DialContext(ctx, "unix", workerAuditSocket(dir, scope.JobID))
	if err != nil {
		return toolbroker.ErrAudit
	}
	defer conn.Close()
	stop := context.AfterFunc(ctx, func() { conn.Close() })
	defer stop()
	deadline := time.Now().Add(5 * time.Second)
	if scope.Deadline.Before(deadline) {
		deadline = scope.Deadline
	}
	conn.SetDeadline(deadline)
	if _, err = conn.Write(auditScopeProof(scope)); err != nil {
		return toolbroker.ErrAudit
	}
	ack := []byte{0}
	if _, err = io.ReadFull(conn, ack); err != nil || ack[0] != 1 {
		return toolbroker.ErrAudit
	}
	conn.SetDeadline(scope.Deadline)
	// Read and write are independent. Closing the SSH process ends both; callers
	// must supply owned closable streams, not model pipes or shared stdin.
	return relayWorkerAudit(ctx, conn, in, out)
}

func relayWorkerAudit(ctx context.Context, conn net.Conn, in io.Reader, out io.Writer) error {
	input, ok := in.(io.ReadCloser)
	output, outputOK := out.(io.WriteCloser)
	if !ok || !outputOK {
		return errors.New("audit input must be owned and closable")
	}
	defer input.Close()
	results := make(chan error, 2)
	go func() { _, err := io.Copy(conn, input); results <- err }()
	go func() { _, err := io.Copy(out, conn); results <- err }()
	var err error
	received := 0
	select {
	case err = <-results:
		received = 1
	case <-ctx.Done():
		err = ctx.Err()
	}
	conn.Close()
	input.Close()
	// Both copies must finish before returning. out is an owned SSH pipe whose
	// consumer must drain it or cancel/close the session.
	output.Close()
	for received < 2 {
		<-results
		received++
	}
	if err != nil {
		return toolbroker.ErrAudit
	}
	return nil
}

// commandAuditLink owns a dedicated SSH process and both pipes. No stdout or
// stderr from the model can enter this channel. Close kills and reaps the link;
// it is not proof that the separate managed containers have stopped.
type commandAuditLink struct {
	input   io.WriteCloser
	output  io.ReadCloser
	command *exec.Cmd
	cancel  context.CancelFunc
	once    sync.Once
	done    chan struct{}
}

func (l *commandAuditLink) Read(p []byte) (int, error)  { return l.output.Read(p) }
func (l *commandAuditLink) Write(p []byte) (int, error) { return l.input.Write(p) }
func (l *commandAuditLink) Close() error {
	l.once.Do(func() { l.cancel(); l.input.Close(); l.output.Close() })
	<-l.done
	return nil
}
func startCommandAuditLink(command *exec.Cmd, cancel context.CancelFunc) (*commandAuditLink, error) {
	input, err := command.StdinPipe()
	if err != nil {
		cancel()
		return nil, toolbroker.ErrAudit
	}
	output, err := command.StdoutPipe()
	if err != nil {
		input.Close()
		cancel()
		return nil, toolbroker.ErrAudit
	}
	command.Stderr = io.Discard
	command.WaitDelay = time.Second
	if err = command.Start(); err != nil {
		input.Close()
		output.Close()
		cancel()
		return nil, toolbroker.ErrAudit
	}
	link := &commandAuditLink{input: input, output: output, command: command, cancel: cancel, done: make(chan struct{})}
	go func() { command.Wait(); cancel(); input.Close(); output.Close(); close(link.done) }()
	return link, nil
}

// serveRemoteToolAudit admits only the configured host-key-verified SSH worker
// and the original registered human/job. It never retries a disconnected link.
// Production pair assembly must run this concurrently with its supervisor and
// cancel/fence the pair on any transport failure before admitting more tools.
func (s *server) serveRemoteToolAudit(ctx context.Context, endpoint string, scope toolbroker.Scope) error {
	alias, provider, err := remoteWorkerEndpoint(endpoint, os.Getenv("COMMONROOM_REMOTE_WORKER_ALIAS"), os.Getenv("COMMONROOM_REMOTE_WORKER"))
	if err != nil || provider != scope.Provider {
		return toolbroker.ErrAudit
	}
	auditor, err := s.toolAuditor(ctx, scope)
	bound, ok := auditor.(*controllerToolAuditor)
	if err != nil || !ok || bound.connection.Endpoint != endpoint {
		return toolbroker.ErrAudit
	}
	raw, err := json.Marshal(scope)
	if err != nil {
		return toolbroker.ErrAudit
	}
	linkCtx, cancel := context.WithDeadline(ctx, scope.Deadline)
	command := workerSSHCommand(linkCtx, alias, "audit-link-v1", base64.RawURLEncoding.EncodeToString(raw))
	link, err := startCommandAuditLink(command, cancel)
	if err != nil {
		return toolbroker.ErrAudit
	}
	defer link.Close()
	return toolbroker.ServeAudit(linkCtx, link, scope, auditor)
}

// OpenAuditListener derives ownership from the admitted supervisor, never from
// model-supplied context. Closed, abandoned and legacy supervisors cannot listen.
func (s *workerPairSupervisor) OpenAuditListener(scope toolbroker.Scope) (*workerAuditListener, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.available(s.ctx) != nil || s.leases == nil || !validWorkerAuditScope(s.record, scope) {
		return nil, toolbroker.ErrAudit
	}
	ctx := context.WithValue(s.ctx, workerLocalJobContextKey{}, s.record.JobID)
	listener, err := newWorkerAuditListener(ctx, s.dir, scope)
	if err != nil {
		return nil, err
	}
	s.auditListener = listener
	return listener, nil
}
