package main

import (
	"agentroom/internal/toolbroker"
	"context"
	"net"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

// workerEngineGateway owns one fresh per-job socket in a private host directory.
// Only its socket inode is mounted into the engine, never the parent directory.
// Socket mode 0666 supports the container's rootless subordinate UID; the private
// mode-0700 host directory prevents access by other host accounts. This is NOT a
// credential file or a replacement for provider destination/address restrictions.
// The worker account/host remain trusted. Source forks can replace this boundary.
type workerEngineGateway struct {
	mu                 sync.Mutex
	scope              toolbroker.Scope
	original, lifetime context.Context
	cancel             context.CancelFunc
	gateway            *providerGateway
	dir                string
	directory, socket  os.FileInfo
	closed             bool
	once               sync.Once
	stop               func() bool
	closeError         error
}

func privateGatewayDirectory(path string) (os.FileInfo, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return nil, toolbroker.ErrRejected
	}
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !fi.IsDir() || fi.Mode()&os.ModeSymlink != 0 || fi.Mode().Perm() != 0700 || !ok || st.Uid != uint32(os.Geteuid()) {
		return nil, toolbroker.ErrRejected
	}
	return fi, nil
}

// root is an existing private worker-controlled staging root. Its ancestors and
// daemon selection are deployment trust. No caller/model may choose root/socket.
// A long Unix socket path is refused, never moved to a shared-directory fallback.
func newWorkerEngineGateway(ctx context.Context, scope toolbroker.Scope, root string) (*workerEngineGateway, error) {
	job, ok := ctxJob(ctx)
	if !ok || job != scope.JobID || scope.Provider != "codex" || !toolbroker.ValidAuditScope(scope, time.Now()) || !filepath.IsAbs(root) || filepath.Clean(root) != root {
		return nil, toolbroker.ErrRejected
	}
	if _, err := privateGatewayDirectory(root); err != nil {
		return nil, err
	}
	scope.Deadline = scope.Deadline.UTC().Round(0)
	dir, err := os.MkdirTemp(root, ".g-")
	if err != nil {
		return nil, toolbroker.ErrExecution
	}
	directory, err := privateGatewayDirectory(dir)
	if err != nil {
		os.Remove(dir)
		return nil, toolbroker.ErrRejected
	}
	lifetime, cancel := context.WithDeadline(ctx, scope.Deadline)
	gateway, err := newProviderGateway(lifetime, filepath.Join(dir, "p.sock"), scope.Provider)
	if err != nil {
		cancel()
		os.Remove(dir)
		return nil, toolbroker.ErrExecution
	}
	// Own unlinking explicitly: UnixListener's automatic unlink could remove a
	// replacement at the same pathname after the original inode was moved.
	gateway.listener.(*gatewayListener).Listener.(*net.UnixListener).SetUnlinkOnClose(false)
	socket, err := os.Lstat(gateway.socket)
	g := &workerEngineGateway{scope: scope, original: ctx, lifetime: lifetime, cancel: cancel, gateway: gateway, dir: dir, directory: directory, socket: socket}
	if err != nil || g.ready(ctx, scope) != nil {
		g.shutdown()
		return nil, toolbroker.ErrRejected
	}
	g.stop = context.AfterFunc(lifetime, func() { g.shutdown() })
	return g, nil
}

func (g *workerEngineGateway) ready(ctx context.Context, scope toolbroker.Scope) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	job, ok := ctxJob(ctx)
	if !ok || job != scope.JobID || scope != g.scope || g.closed || g.original.Err() != nil || g.lifetime.Err() != nil || g.gateway.ctx.Err() != nil || !toolbroker.ValidAuditScope(scope, time.Now()) {
		return toolbroker.ErrRejected
	}
	dir, err := privateGatewayDirectory(g.dir)
	if err != nil || !os.SameFile(dir, g.directory) {
		return toolbroker.ErrRejected
	}
	socket, err := os.Lstat(g.gateway.socket)
	if err != nil || g.socket == nil || socket.Mode()&os.ModeSocket == 0 || socket.Mode().Perm() != 0666 || !os.SameFile(socket, g.socket) {
		return toolbroker.ErrRejected
	}
	stat, ok := socket.Sys().(*syscall.Stat_t)
	if !ok || stat.Uid != uint32(os.Geteuid()) || stat.Nlink != 1 {
		return toolbroker.ErrRejected
	}
	return nil
}

// Close revokes network access; it never proves either container has stopped.
// Only the owned empty directory is removed, never recursively deleted. Changed
// ancestry/inodes or extra files leave an explicit cleanup error for recovery.
func (g *workerEngineGateway) Close() error {
	if g.stop != nil {
		g.stop()
	}
	g.shutdown()
	return g.closeError
}
func (g *workerEngineGateway) shutdown() {
	g.once.Do(func() {
		g.mu.Lock()
		g.closed = true
		g.mu.Unlock()
		g.cancel()
		g.gateway.Close()
		fi, err := privateGatewayDirectory(g.dir)
		if err != nil || !os.SameFile(fi, g.directory) {
			g.closeError = toolbroker.ErrShutdownUnconfirmed
			return
		}
		socket, err := os.Lstat(g.gateway.socket)
		if err == nil {
			if g.socket == nil || !os.SameFile(socket, g.socket) || socket.Mode()&os.ModeSocket == 0 || os.Remove(g.gateway.socket) != nil {
				g.closeError = toolbroker.ErrShutdownUnconfirmed
				return
			}
		} else if !os.IsNotExist(err) {
			g.closeError = toolbroker.ErrShutdownUnconfirmed
			return
		}
		if os.Remove(g.dir) != nil {
			g.closeError = toolbroker.ErrShutdownUnconfirmed
		}
	})
}
