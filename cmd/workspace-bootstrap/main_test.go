package main

import (
	"agentroom/internal/workspacebootstrap"
	"context"
	"slices"
	"strings"
	"testing"
)

func TestGitCommandHasFixedExecutableAndConfiguration(t *testing.T) {
	t.Parallel()
	cmd := (gitRunner{home: "/tmp/synthetic-home"}).command(context.Background(), "/workspace/repository", "ls-tree", "-r", "-l", "-z", strings.Repeat("a", 40))
	if cmd.Path != "/usr/bin/git" || cmd.Dir != "/workspace/repository" || cmd.Stdin != nil {
		t.Fatal("unsafe command")
	}
	for _, value := range []string{"core.hooksPath=/dev/null", "core.fsmonitor=false", "protocol.allow=never", "protocol.file.allow=always", "submodule.recurse=false"} {
		if !slices.Contains(cmd.Args, value) {
			t.Fatal("missing mandatory Git setting")
		}
	}
	expected := []string{"HOME=/tmp/synthetic-home", "PATH=/usr/bin:/bin", "LANG=C", "LC_ALL=C", "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_TERMINAL_PROMPT=0", "GIT_NO_REPLACE_OBJECTS=1", "GIT_ATTR_NOSYSTEM=1", "GIT_OPTIONAL_LOCKS=0"}
	if !slices.Equal(cmd.Env, expected) {
		t.Fatal("ambient environment inherited")
	}
}
func TestGitOutputBound(t *testing.T) {
	t.Parallel()
	w := &limitedWriter{}
	if _, err := w.Write(make([]byte, workspacebootstrap.MaxTreeBytes)); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("x")); err == nil || w.Len() != workspacebootstrap.MaxTreeBytes {
		t.Fatal("unbounded Git metadata")
	}
}
