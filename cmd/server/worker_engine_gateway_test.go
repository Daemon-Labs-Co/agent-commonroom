package main

import (
	"agentroom/internal/toolbroker"
	"bufio"
	"context"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func shortGatewayRoot(t *testing.T) string {
	t.Helper()
	root, err := os.MkdirTemp("", "crg-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(root) })
	return root
}
func engineGatewayFixture(t *testing.T, ctx context.Context, scope toolbroker.Scope) *workerEngineGateway {
	t.Helper()
	g, err := newWorkerEngineGateway(ctx, scope, shortGatewayRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { g.Close() })
	return g
}
func TestEngineGatewayFreshPrivateScopedAndCancelled(t *testing.T) {
	t.Parallel()
	scope := sampleToolScope(t)
	ctx, cancel := context.WithCancel(context.WithValue(t.Context(), workerLocalJobContextKey{}, scope.JobID))
	defer cancel()
	root := shortGatewayRoot(t)
	first, err := newWorkerEngineGateway(ctx, scope, root)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := newWorkerEngineGateway(ctx, scope, root)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	if first.dir == second.dir || first.gateway.socket == second.gateway.socket {
		t.Fatal("gateway directory reused")
	}
	if first.ready(ctx, scope) != nil {
		t.Fatal("fresh socket refused")
	}
	foreign := scope
	foreign.Room = "foreign"
	if first.ready(ctx, foreign) == nil {
		t.Fatal("foreign scope accepted")
	}
	conn, err := net.DialTimeout("unix", first.gateway.socket, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(time.Second))
	// A forbidden destination is handled locally; no DNS or external dial occurs.
	io.WriteString(conn, "CONNECT example.invalid:443 HTTP/1.1\r\nHost: example.invalid:443\r\n\r\n")
	status, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil || status != "HTTP/1.1 403 Forbidden\r\n" {
		t.Fatalf("destination restriction lost: %q %v", status, err)
	}
	cancel()
	if first.ready(context.WithValue(t.Context(), workerLocalJobContextKey{}, scope.JobID), scope) == nil {
		t.Fatal("fresh context revived cancelled gateway")
	}
	if err = first.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Lstat(first.dir); !os.IsNotExist(err) {
		t.Fatal("owned directory retained", err)
	}
	if first.Close() != nil {
		t.Fatal("close not idempotent")
	}
	if _, err = net.DialTimeout("unix", first.gateway.socket, time.Second); err == nil {
		t.Fatal("gateway still listening")
	}
}
func TestEngineGatewayRefusesUnsafeRootAndChangedSocket(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"public root", "symlink root", "missing job", "foreign scope", "socket replaced", "directory public", "extra file"} {
		t.Run(mode, func(t *testing.T) {
			scope := sampleToolScope(t)
			ctx := context.WithValue(t.Context(), workerLocalJobContextKey{}, scope.JobID)
			root := shortGatewayRoot(t)
			switch mode {
			case "public root":
				if err := os.Chmod(root, 0755); err != nil {
					t.Fatal(err)
				}
			case "symlink root":
				link := filepath.Join(shortGatewayRoot(t), "link")
				if err := os.Symlink(root, link); err != nil {
					t.Fatal(err)
				}
				root = link
			case "missing job":
				ctx = t.Context()
			}
			g, err := newWorkerEngineGateway(ctx, scope, root)
			if mode == "public root" || mode == "symlink root" || mode == "missing job" {
				if err == nil {
					g.Close()
					t.Fatal("unsafe gateway admitted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			defer g.Close()
			switch mode {
			case "foreign scope":
				scope.Connection = "foreign"
			case "socket replaced":
				if err := os.Remove(g.gateway.socket); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(g.gateway.socket, []byte("synthetic"), 0600); err != nil {
					t.Fatal(err)
				}
			case "directory public":
				if err := os.Chmod(g.dir, 0755); err != nil {
					t.Fatal(err)
				}
			case "extra file":
				if err := os.WriteFile(filepath.Join(g.dir, "retained"), []byte("synthetic"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if mode != "extra file" && g.ready(ctx, scope) == nil {
				t.Fatal("changed gateway accepted")
			}
			if mode == "socket replaced" {
				if g.Close() == nil {
					t.Fatal("replacement socket cleanup acknowledged")
				}
				if raw, err := os.ReadFile(g.gateway.socket); err != nil || string(raw) != "synthetic" {
					t.Fatal("replacement file removed", err)
				}
			}
			if mode == "extra file" {
				if g.Close() == nil {
					t.Fatal("unknown file silently discarded")
				}
				if _, err := os.Stat(filepath.Join(g.dir, "retained")); err != nil {
					t.Fatal("unknown file removed")
				}
			}
		})
	}
}
