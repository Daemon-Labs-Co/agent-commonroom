package main

// Assembly point for internal/toolbroker's credential/tool-separation
// foundation: it wires the real Auditor (tool_broker_audit.go) and the real
// DockerRunner (tool_broker_runner.go) into a working toolbroker.Dispatcher.
//
// Nothing in cmd/server calls newToolDispatcher outside its own test. No live
// HTTP handler or agent turn dispatches a real tool call through it yet.
// Deciding how a provider's in-container tool call gets intercepted and
// routed here — instead of running directly in the provider's own container,
// as it does today — is a separate transport-protocol design question and
// unstarted follow-up work. See docs/provider-tool-separation.md.

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"time"
)

// toolSeparationEnabled defaults off. It currently gates nothing observable:
// no caller of newToolDispatcher exists yet, so this flag is a rollout
// marker/kill-switch placed ahead of the first real caller, not an active
// behavior gate. Per AGENTS.md's mandatory execution boundary, once a real
// caller exists this flag must never become a way to fall back to the
// credential-colocated single-container tool path — turning tool separation
// off must mean "do not use the broker yet," never "use it unsafely."
func toolSeparationEnabled(st room.State) bool { return st.FeatureFlags["toolSeparation"] }

// newToolDispatcher mints a bounded-lifetime Authority for scope, builds the
// locked-down tool container plan and the real Docker-backed executor, and
// assembles them with the real event-store-backed Auditor into a Dispatcher.
// image and dockerRoot are supplied by the caller from verified worker
// configuration (see ToolContainerPlan/ContainerExecutor doc comments) — this
// function sources neither from an environment variable itself, since no
// caller exists yet to decide where that configuration should come from.
// The returned token is the raw one-time capability secret; the caller must
// deliver it to the tool container's paired entrypoint and never log it.
func (s *server) newToolDispatcher(scope toolbroker.Scope, allowed []string, image, dockerRoot string, now time.Time) (*toolbroker.Dispatcher, string, error) {
	authority, token, err := toolbroker.New(scope, allowed, now)
	if err != nil {
		return nil, "", err
	}
	plan, err := toolbroker.NewToolContainerPlan(authority, image, now)
	if err != nil {
		return nil, "", err
	}
	executor, err := toolbroker.NewContainerExecutor(plan, dockerCLIRunner{}, dockerRoot)
	if err != nil {
		return nil, "", err
	}
	dispatcher, err := toolbroker.NewDispatcher(authority, &storedToolAuditor{s}, executor)
	if err != nil {
		return nil, "", err
	}
	return dispatcher, token, nil
}
