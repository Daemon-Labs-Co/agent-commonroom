package room

import "testing"

func TestAdapterEndpointBoundaries(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		endpoint, protocol string
		codex, agent       bool
	}{
		{"codex", "CLI", true, true}, {"claude", "CLI", false, true}, {"ssh://jerry", "CLI", true, true}, {"ssh://debian-claude/claude", "CLI", false, true},
		{"codex", "HTTP", false, false}, {"ssh://-option", "CLI", false, false}, {"ssh://user@host", "CLI", false, false}, {"ssh://host:22", "CLI", false, false}, {"ssh://host/path", "CLI", false, false}, {"ssh://host;touch file", "CLI", false, false}, {"ssh://host/claude?token=x", "CLI", false, false}, {"https://chatgpt.com", "HTTP", false, false},
	} {
		t.Run(c.endpoint+"/"+c.protocol, func(t *testing.T) {
			t.Parallel()
			v := Connection{Protocol: c.protocol, Endpoint: c.endpoint}
			if SupportsCodex(v) != c.codex || SupportsAgent(v) != c.agent {
				t.Fatalf("wrong capabilities: %+v", v)
			}
		})
	}
}
