package main

import (
	"strings"
	"testing"
)

func TestClaudeSignInBoundaries(t *testing.T) {
	t.Parallel()
	if !localClaudeConnection("CLI", "claude") {
		t.Fatal("local Claude rejected")
	}
	for _, endpoint := range []string{"codex", "ssh://claude/claude", "claude; echo bad", "/tmp/claude"} {
		if localClaudeConnection("CLI", endpoint) {
			t.Fatal("unexpected executable accepted")
		}
	}
	if localClaudeConnection("HTTP", "claude") {
		t.Fatal("HTTP accepted")
	}
	if shellArgument("a'b") != "'a'\\''b'" {
		t.Fatal("shell quote unsafe")
	}
	var b cappedStatusOutput
	if _, e := b.Write([]byte(strings.Repeat("x", 16385))); e == nil {
		t.Fatal("unbounded status output")
	}
}
