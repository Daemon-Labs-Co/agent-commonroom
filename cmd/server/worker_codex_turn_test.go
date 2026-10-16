package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestCodexToolRequestBindsPrivateAuthorityOutsideModelArguments(t *testing.T) {
	t.Parallel()
	scope := sampleToolScope(t)
	request, err := workerCodexToolRequest(scope, strings.Repeat("a", 64), "call-1", json.RawMessage(`{"operation":"write","path":"file.go","content":"synthetic"}`))
	if err != nil || request.Scope != scope || request.Capability != strings.Repeat("a", 64) || request.CallID != "call-1" || request.Operation != "write" || request.Content != "synthetic" {
		t.Fatal("trusted binding lost", err)
	}
	for _, raw := range []string{`{"operation":"read","path":"file.go","capability":"model-controlled"}`, `{"operation":"read","path":"file.go","scope":{}}`, `{"operation":"read","path":"file.go","callId":"foreign"}`, `{"operation":"command","path":"file.go","command":["sh"]}`, `{"operation":"read","path":"file.go","content":"unexpected"}`, `{"operation":"read","path":null}`, `{"operation":"read","path":"file.go","Content":"alias"}`, `{"operation":"read","operation":"write","path":"file.go"}`, `null`, `[]`} {
		if _, err = workerCodexToolRequest(scope, strings.Repeat("a", 64), "call-1", json.RawMessage(raw)); err == nil {
			t.Fatal("model authority injection accepted", raw)
		}
	}
}
func TestCodexTurnRefusesBeforeReviewedPairStartup(t *testing.T) {
	t.Parallel()
	_, s, _ := workspacePair(t)
	if _, err := s.CodexTurn(context.Background(), nil, "", nil, "synthetic", "model"); err == nil {
		t.Fatal("unreviewed engine turn admitted")
	}
	if s.codexTurnStarted {
		t.Fatal("refused turn consumed driver ownership")
	}
}
