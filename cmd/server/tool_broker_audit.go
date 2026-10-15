package main

// storedToolAuditor is the first real implementation of toolbroker.Auditor. It
// persists broker-mediated tool call intent/outcome evidence through the
// existing append-only audit chain (event_store.go), inheriting that path's
// hash-chain, actor attribution and transaction guarantees for free. It never
// receives or inserts capability tokens, raw tool output or Docker launch
// options — toolbroker.Dispatcher strips those before calling RecordIntent/
// RecordOutcome; only hashes and scope metadata reach this boundary.
//
// This type has no caller yet outside its own test. Wiring a live agent turn
// to dispatch real tool calls through internal/toolbroker is separate,
// unstarted follow-up work; see docs/provider-tool-separation.md.

import (
	"context"
	"time"

	"agentroom/internal/toolbroker"
)

type storedToolAuditor struct{ s *server }

// toolCallOutcomeRecord correlates an outcome with its originating intent in
// one append so replay never has to guess which call an outcome belongs to.
type toolCallOutcomeRecord struct {
	Intent  toolbroker.Intent  `json:"intent"`
	Outcome toolbroker.Outcome `json:"outcome"`
}

func (a *storedToolAuditor) RecordIntent(ctx context.Context, intent toolbroker.Intent) error {
	if !toolbroker.ValidAuditIntent(intent, time.Now()) {
		return toolbroker.ErrAudit
	}
	return a.s.append(ctx, "tool.call.intent", intent)
}

// RecordOutcome does not re-run ValidAuditIntent's deadline-in-future check:
// outcomes for a call that ran right up to (or past) its deadline are expected
// and legitimate, per ValidAuditIntent's own doc comment ("outcome persistence
// may instead use the original intent's admission instant"). The intent value
// itself was already admission-checked by RecordIntent moments earlier in the
// same Dispatch call and is Dispatcher-constructed, not external input.
func (a *storedToolAuditor) RecordOutcome(ctx context.Context, intent toolbroker.Intent, outcome toolbroker.Outcome) error {
	if !toolbroker.ValidAuditOutcome(outcome) {
		return toolbroker.ErrAudit
	}
	return a.s.append(ctx, "tool.call.outcome", toolCallOutcomeRecord{Intent: intent, Outcome: outcome})
}
