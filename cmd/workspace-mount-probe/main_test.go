package main

import (
	"agentroom/internal/workspacemount"
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

func TestRunWithholdsInvalidObservation(t *testing.T) {
	t.Parallel()
	challenge := strings.Repeat("a", 32)
	for _, mode := range []string{"healthy", "observe error", "wrong challenge", "writable", "cancel", "extra argument"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			args := []string{challenge}
			if mode == "extra argument" {
				args = append(args, "/foreign")
			}
			var out bytes.Buffer
			err := run(ctx, args, &out, func(context.Context, string) (workspacemount.Proof, error) {
				p := workspacemount.Proof{Version: 1, Challenge: challenge, MountID: 2, ReadOnly: true, RootReadOnly: true}
				switch mode {
				case "observe error":
					return p, errors.New("synthetic")
				case "wrong challenge":
					p.Challenge = strings.Repeat("b", 32)
				case "writable":
					p.ReadOnly = false
				case "cancel":
					cancel()
				}
				return p, nil
			})
			if mode == "healthy" {
				if err != nil || !strings.Contains(out.String(), challenge) {
					t.Fatal(err)
				}
			} else if err == nil || out.Len() != 0 {
				t.Fatalf("released invalid observation: %s, %v", out.String(), err)
			}
		})
	}
}
