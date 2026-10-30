package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStartupCommandRejectsAmbiguousIntent(t *testing.T) {
	t.Parallel()
	for _, args := range [][]string{{"--check-worker-storag"}, {"unexpected"}, {"--migrate", "extra"}, {"--help", "secret-canary"}, {"--check-reader-profiles"}, {"--check-worker-storage", "one", "two"}, {"--check-reader-profiles=secret-canary"}} {
		var out, diagnostic bytes.Buffer
		handled, code := startupCommand(args, &out, &diagnostic)
		if !handled || code != 2 || out.Len() != 0 || diagnostic.Len() == 0 || strings.Contains(diagnostic.String(), "canary") {
			t.Fatal("ambiguous startup accepted/disclosed", args, handled, code)
		}
	}
	for _, args := range [][]string{nil, {"--migrate"}, {"--agent-worker", "internal"}} {
		var out, diagnostic bytes.Buffer
		if handled, code := startupCommand(args, &out, &diagnostic); handled || code != 0 || out.Len() != 0 || diagnostic.Len() != 0 {
			t.Fatal("existing explicit startup changed")
		}
	}
}
func TestStartupCommandHelpAndPrivateValidation(t *testing.T) {
	t.Parallel()
	for _, flag := range []string{"--help", "-h"} {
		var out, diagnostic bytes.Buffer
		if handled, code := startupCommand([]string{flag}, &out, &diagnostic); !handled || code != 0 || !strings.Contains(out.String(), "--check-worker-storage") || diagnostic.Len() != 0 {
			t.Fatal("help unavailable")
		}
	}
	g, _, _, _ := objectReadSessionFixture(t)
	path, _ := objectReadProfileFixture(t, g)
	var out, diagnostic bytes.Buffer
	if handled, code := startupCommand([]string{"--check-reader-profiles", path}, &out, &diagnostic); !handled || code != 0 || !strings.Contains(out.String(), "2 entries") || diagnostic.Len() != 0 {
		t.Fatal("private preflight failed")
	}
}
func TestStartupCommandProcessChild(t *testing.T) {
	if os.Getenv("COMMONROOM_SYNTHETIC_STARTUP_CHILD") != "1" {
		return
	}
	os.Args = append([]string{"commonroom"}, strings.Split(os.Getenv("COMMONROOM_SYNTHETIC_STARTUP_ARGS"), "\n")...)
	main()
	os.Exit(99)
}
func TestStartupCommandProcessDoesNotStartController(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		args    []string
		code    int
		message string
	}{
		{[]string{"--help"}, 0, "Usage:"},
		{[]string{"--check-worker-storag"}, 2, "Invalid command"},
		{[]string{"--migrate", "secret-canary"}, 2, "Invalid command"},
		{[]string{"--check-reader-profiles"}, 2, "Usage:"},
		{[]string{"--check-worker-storage", "/synthetic/nonexistent-policy"}, 1, "validation refused"},
	} {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestStartupCommandProcessChild$")
			cmd.Dir = dir
			// No inherited credentials/configuration. Invalid DB syntax is an additional
			// fail-before-dial guard if startup ordering ever regresses; no DB is contacted.
			cmd.Env = []string{"COMMONROOM_SYNTHETIC_STARTUP_CHILD=1", "COMMONROOM_SYNTHETIC_STARTUP_ARGS=" + strings.Join(tc.args, "\n"), "DATABASE_URL=://synthetic-invalid", "HOME=" + dir, "TMPDIR=" + dir}
			output, err := cmd.CombinedOutput()
			code := 0
			if err != nil {
				var exit *exec.ExitError
				if !errors.As(err, &exit) {
					t.Fatal(err)
				}
				code = exit.ExitCode()
			}
			if ctx.Err() != nil || code != tc.code || !strings.Contains(string(output), tc.message) || strings.Contains(string(output), "canary") {
				t.Fatalf("code=%d output=%s err=%v", code, output, err)
			}
			if _, err := os.Stat(filepath.Join(dir, ".data")); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("operator command touched controller state", err)
			}
		})
	}
}
