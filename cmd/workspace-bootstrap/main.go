// workspace-bootstrap runs only inside the credential-free Linux tool container.
// Calling it on a host CAN expose host files; doing so breaks managed isolation.
package main

import (
	"agentroom/internal/workspacebootstrap"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

type gitRunner struct{ home string }

func (g gitRunner) command(ctx context.Context, dir string, args ...string) *exec.Cmd {
	fixed := []string{"-c", "core.hooksPath=/dev/null", "-c", "core.fsmonitor=false", "-c", "core.autocrlf=false", "-c", "core.attributesFile=/dev/null", "-c", "core.protectHFS=true", "-c", "core.protectNTFS=true", "-c", "protocol.allow=never", "-c", "protocol.file.allow=always", "-c", "submodule.recurse=false"}
	cmd := exec.CommandContext(ctx, "/usr/bin/git", append(fixed, args...)...)
	cmd.Dir = dir
	cmd.Env = []string{"HOME=" + g.home, "PATH=/usr/bin:/bin", "LANG=C", "LC_ALL=C", "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_TERMINAL_PROMPT=0", "GIT_NO_REPLACE_OBJECTS=1", "GIT_ATTR_NOSYSTEM=1", "GIT_OPTIONAL_LOCKS=0"}
	cmd.WaitDelay = 2 * time.Second
	return cmd
}

func (g gitRunner) Run(ctx context.Context, dir string, args ...string) ([]byte, error) {
	cmd := g.command(ctx, dir, args...)
	out := &limitedWriter{}
	cmd.Stdout = out
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return nil, workspacebootstrap.ErrRefused
	}
	return out.Bytes(), nil
}

type limitedWriter struct{ bytes.Buffer }

func (w *limitedWriter) Write(p []byte) (int, error) {
	if w.Len()+len(p) > workspacebootstrap.MaxTreeBytes {
		return 0, io.ErrShortBuffer
	}
	return w.Buffer.Write(p)
}
func main() {
	if runtime.GOOS != "linux" || len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, workspacebootstrap.ErrRefused)
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	home, err := os.MkdirTemp("/tmp", "commonroom-bootstrap-")
	if err == nil {
		defer os.RemoveAll(home)
		err = workspacebootstrap.Run(ctx, os.Stdin, os.Stdout, "/workspace", os.Args[1], os.Args[2], gitRunner{home})
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, workspacebootstrap.ErrRefused)
		os.Exit(2)
	}
}
