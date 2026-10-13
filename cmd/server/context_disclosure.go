package main

import (
	"agentroom/internal/room"
	"encoding/json"
	"errors"
)

// Export restrictions are mandatory. The editor flag only controls configuration
// UX; disabling it never grants a model access to previously excluded material.
func contextPolicyEditorEnabled(st room.State) bool {
	return st.FeatureFlags["contextPolicyEditor"] || !flagDefined(st, "contextPolicyEditor")
}
func flagDefined(st room.State, name string) bool { _, ok := st.FeatureFlags[name]; return ok }
func validExportProviders(providers []string) bool {
	if len(providers) > 2 {
		return false
	}
	seen := map[string]bool{}
	for _, provider := range providers {
		if (provider != "openai" && provider != "claude") || seen[provider] {
			return false
		}
		seen[provider] = true
	}
	return true
}
func containsProvider(providers []string, provider string) bool {
	for _, v := range providers {
		if v == provider {
			return true
		}
	}
	return false
}
func contextDestination(st room.State, p room.Participant) (string, *room.ContextExportPolicy, error) {
	provider := ""
	for _, c := range st.Connections {
		if c.ID == p.ConnectionID {
			provider = providerForEndpoint(c.Endpoint)
			if c.Protocol != "CLI" {
				return "", nil, errors.New("Unsupported context destination")
			}
		}
	}
	if provider == "" {
		return "", nil, errors.New("Context destination connection missing")
	}
	for _, r := range st.Rooms {
		if r.ID == p.RoomID {
			if r.CompanyID != "" && p.CompanyID != "" && r.CompanyID != p.CompanyID {
				return "", nil, errors.New("Context company mismatch")
			}
			policy := r.ContextExport
			if policy != nil && !containsProvider(policy.Providers, provider) {
				return "", nil, errors.New("Room context export to this provider is disabled")
			}
			return provider, policy, nil
		}
	}
	return "", nil, errors.New("Context room missing")
}
func memoryExportAllowed(st room.State, m room.Memory, p room.Participant, provider string) bool {
	if m.AgentExport != nil && !*m.AgentExport {
		return false
	}
	if len(m.ExportProviders) > 0 && !containsProvider(m.ExportProviders, provider) {
		return false
	}
	// Project sharing is not a permission boundary until a real Project model exists.
	if m.RoomID == p.RoomID {
		return true
	}
	if m.Scope != "Application" || !containsProvider(m.ExportProviders, provider) {
		return false
	}
	for _, r := range st.Rooms {
		if r.ID == m.RoomID && r.PersonalActorID != "" {
			return false
		}
	}
	return true
}

// referenceJSON keeps source metadata distinct from lower-trust content. Labels
// support model interpretation; actual exclusion happens before serialization.
func referenceJSON(value any) string { raw, _ := json.Marshal(value); return string(raw) }

// Legacy native sessions may contain context exported before disclosure checks.
// Start a new session once; preserved application history is filtered normally.
func nativeContextThread(p room.Participant) string {
	if p.DisclosureVersion != 1 {
		return ""
	}
	return p.ThreadID
}
