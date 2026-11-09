package main

// Per-agent model choice. The model is a validated name passed to the provider
// CLI as one option; it cannot change isolation, sandbox or tool policy, and the
// worker independently accepts only the same strict name pattern.

import (
	"agentroom/internal/room"
	"fmt"
	"regexp"
	"strings"
)

var modelNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:/-]{0,63}$`)

// agentModelEnabled is a UX toggle. Disabling it rejects new or changed model
// choices; models already recorded keep being used so agents do not silently
// switch models.
func agentModelEnabled(st room.State) bool {
	return st.FeatureFlags["agentModelSelection"] || !flagDefined(st, "agentModelSelection")
}

// checkAgentModel normalizes and validates the requested model on an invitation
// or profile update. An empty model means the provider default.
func checkAgentModel(st room.State, previous string, v *room.Participant) (string, bool) {
	v.Model = strings.TrimSpace(v.Model)
	if v.Model == "" || v.Model == previous {
		return "", true
	}
	if !agentModelEnabled(st) {
		return "Choosing an agent model is disabled", false
	}
	if !modelNamePattern.MatchString(v.Model) {
		return "Model name may use letters, digits and . _ : / - (maximum 64 characters)", false
	}
	return "", true
}

// modelArgs is the optional provider option for a chosen model; empty means the
// provider default. It is placed before any session-resume arguments.
func modelArgs(provider, model string) []string {
	if model == "" || !modelNamePattern.MatchString(model) {
		return nil
	}
	if provider == "codex" {
		return []string{"-m", model}
	}
	return []string{"--model", model}
}

// explainModelFailure points at the agent's model when a provider error
// mentions it, so a mistyped name does not look like a generic failure.
func explainModelFailure(model string, err error) error {
	if err == nil || model == "" || !strings.Contains(strings.ToLower(err.Error()), "model") {
		return err
	}
	return fmt.Errorf("Model %q may not be available on this connection; change it in the agent's profile (%w)", model, err)
}
