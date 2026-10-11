package main

import (
	"context"
	"strings"
	"testing"
)

func TestIsolatedCommandPolicy(t *testing.T) {
	t.Parallel()
	image := "example/agent@sha256:" + strings.Repeat("a", 64)
	if !validAgentImage("sha256:" + strings.Repeat("a", 64)) {
		t.Fatal("immutable local image ID rejected")
	}
	if !validAgentImage(image) {
		t.Fatal("pinned image rejected")
	}
	for _, bad := range []string{"agent:latest", "-option@sha256:" + strings.Repeat("a", 64), "agent@sha256:bad", "agent\nother@sha256:" + strings.Repeat("a", 64)} {
		if validAgentImage(bad) {
			t.Fatal("unsafe image accepted")
		}
	}
	arg := "text with ' quotes; $(touch should-never-run)"
	args := isolatedRunArgs(image, "commonroom-test", "commonroom-session-test", "/tmp/socket", "/tmp/auth", "codex", []string{"exec", arg})
	joined := strings.Join(args, " ")
	for _, required := range []string{"--network=none", "--read-only", "--cap-drop=ALL", "no-new-privileges:true", "--pids-limit=128", "--memory=1g", "--cpus=1", "--user 1000:1000", "--pull=never", "--log-driver=none"} {
		if !strings.Contains(joined, required) {
			t.Fatal("missing containment setting", required)
		}
	}
	if args[len(args)-1] != arg {
		t.Fatal("arguments interpreted as shell syntax")
	}
	for _, forbidden := range []string{"--privileged", "--network=host", "docker.sock", "/Users/", "/root/.ssh"} {
		if strings.Contains(joined, forbidden) {
			t.Fatal("host privilege leaked", forbidden)
		}
	}
}
func TestNoRemoteOrLegacyFallback(t *testing.T) {
	t.Parallel()
	for _, endpoint := range []string{"ssh://jerry", "ssh://claude/claude", "sh", "http://localhost"} {
		cmd, cleanup, err := isolatedAgentCommand(context.Background(), endpoint, "key", "exec")
		if err == nil || cmd != nil || cleanup != nil {
			t.Fatal("unisolated endpoint accepted")
		}
	}
}
func TestProviderEgressPolicy(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		provider, host string
		allowed        bool
	}{
		{"codex", "api.openai.com", true}, {"claude", "api.anthropic.com", true}, {"claude", "api.openai.com", false}, {"codex", "api.openai.com.attacker.example", false}, {"codex", "localhost", false}, {"codex", "169.254.169.254", false}, {"codex", "api.openai.com.", false}, {"unknown", "api.openai.com", false},
	} {
		if providerHostAllowed(c.provider, c.host) != c.allowed {
			t.Fatal("egress policy incorrect", c)
		}
	}
}

func TestDockerSecurityPrerequisites(t *testing.T) {
	t.Parallel()
	valid := `{"SecurityOptions":["name=rootless","name=seccomp,profile=builtin"],"MemoryLimit":true,"SwapLimit":true,"PidsLimit":true,"CpuCfsQuota":true}`
	if !dockerSecurityReady([]byte(valid)) {
		t.Fatal("valid daemon rejected")
	}
	for _, broken := range []string{strings.Replace(valid, "rootless", "other", 1), strings.Replace(valid, "seccomp", "other", 1), strings.Replace(valid, `"MemoryLimit":true`, `"MemoryLimit":false`, 1), strings.Replace(valid, `"PidsLimit":true`, `"PidsLimit":false`, 1), "{}", "not json"} {
		if dockerSecurityReady([]byte(broken)) {
			t.Fatal("unsafe daemon accepted")
		}
	}
}
