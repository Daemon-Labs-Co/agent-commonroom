package main

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestLegacyDevelopmentCannotReadCredentialsOrLaunch(t *testing.T) {
	t.Parallel()
	for _, provider := range []string{"codex", "claude"} {
		cmd, cleanup, err := localIsolatedDevelopmentCommand(context.Background(), provider, "synthetic", strings.Repeat("a", 32), strings.Repeat("b", 64), false, providerTurnArgs(provider, true, false)...)
		if !errors.Is(err, errCredentialBoundary) || cmd != nil || cleanup != nil {
			t.Fatal("legacy native development preparation reached credential/runtime setup")
		}
	}
}

func TestWorkerRefusesLegacyDevelopmentBeforeAdmission(t *testing.T) {
	t.Parallel()
	v := sampleWorkerRequest()
	v.Repository = strings.Repeat("b", 64)
	v.Args = append(providerTurnArgs("codex", true, false), "-")
	if !errors.Is(validateWorkerRun(v), errCredentialBoundary) {
		t.Fatal("legacy development worker request accepted")
	}
	v.Repository = ""
	v.Args = append(providerTurnArgs("codex", false, false), "-")
	if err := validateWorkerRun(v); err != nil {
		t.Fatal("chat request unexpectedly refused", err)
	}
}

func TestCredentialBoundaryDiagnosticDoesNotSuggestSignIn(t *testing.T) {
	t.Parallel()
	reason, next := safeAgentFailure(errCredentialBoundary.Error())
	if !strings.Contains(reason, "native tools") || !strings.Contains(next, "separate provider/tool bridge") || strings.Contains(next, "check sign-in") {
		t.Fatal("wrong remediation for mandatory separation blocker")
	}
}
