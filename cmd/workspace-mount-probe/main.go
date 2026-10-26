// workspace-mount-probe is a fixed credential-free isolated Linux helper.
// Running it outside the reviewed container invalidates managed mount claims.
package main

import (
	"agentroom/internal/workspacemount"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func run(ctx context.Context, args []string, out io.Writer, observe func(context.Context, string) (workspacemount.Proof, error)) error {
	if ctx == nil || ctx.Err() != nil || len(args) != 1 || !workspacemount.ValidChallenge(args[0]) || out == nil || observe == nil {
		return workspacemount.ErrRefused
	}
	proof, err := observe(ctx, args[0])
	if err != nil || ctx.Err() != nil || !workspacemount.ValidProof(proof, args[0]) {
		return workspacemount.ErrRefused
	}
	return json.NewEncoder(out).Encode(proof)
}
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := run(ctx, os.Args[1:], os.Stdout, workspacemount.Observe); err != nil {
		fmt.Fprintln(os.Stderr, workspacemount.ErrRefused)
		os.Exit(2)
	}
}
