package main

import "agentroom/internal/toolbroker"

// toolCallOutcomeRecord decodes historical tool.call.outcome events only. New
// actions use worker.tool.intent/outcome via the scoped controller auditor.
// Preserve this wire shape and its projection; historical evidence is never
// rewritten or converted into current execution authority.
type toolCallOutcomeRecord struct {
	Intent  toolbroker.Intent  `json:"intent"`
	Outcome toolbroker.Outcome `json:"outcome"`
}
