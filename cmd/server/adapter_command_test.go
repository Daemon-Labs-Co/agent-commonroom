package main

import (
	"context"
	"strings"
	"testing"
)

func TestCodexCommandArgumentIsolation(t *testing.T) {
	t.Parallel()
	// Inspect argv only: never start a provider executable, shell, or SSH session.
	arg := "text with ' quotes; $(touch should-never-run)"
	local := codexCommand(context.Background(), "codex", "exec", arg)
	if len(local.Args) != 3 || local.Args[1] != "exec" || local.Args[2] != arg {
		t.Fatal("local prompt interpreted as command syntax")
	}
	remote := codexCommand(context.Background(), "ssh://jerry", "exec", arg)
	if len(remote.Args) != 8 || remote.Args[6] != "jerry" {
		t.Fatalf("unexpected SSH argv: %q", remote.Args)
	}
	want := "exec ~/.local/bin/codex 'exec' " + shellArgument(arg)
	if remote.Args[7] != want {
		t.Fatal("remote argument not safely quoted", remote.Args[7])
	}
	if !strings.Contains(strings.Join(remote.Args, " "), "BatchMode=yes") {
		t.Fatal("interactive SSH enabled")
	}
}
