package main

import (
	"context"
	"encoding/json"
	"io"
	"time"
)

func backgroundWorkerStage(raw []byte, job, provider string) string {
	var record workerJobRecord
	if len(raw) > 8192 || json.Unmarshal(raw, &record) != nil || record.JobID != job || record.Provider != provider {
		return "unreachable"
	}
	switch record.Status {
	case "starting", "running", "finished", "failed", "cancelled":
		return record.Status
	}
	return "unreachable"
}

// Poll only the approved SSH worker and the job allocated by this controller.
// Same-state observations are bounded to fifteen-second confirmations. They do
// not extend deadlines, authorize tools or prove useful model progress.
func monitorRemoteBackground(ctx context.Context, done <-chan struct{}, alias, job, provider string) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	last := ""
	confirmed := time.Time{}
	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		case <-ticker.C:
		}
		poll, cancel := context.WithTimeout(ctx, 3*time.Second)
		cmd := workerSSHCommand(poll, alias, "status-v1", job)
		var output limitedAgentOutput
		cmd.Stdout = &output
		cmd.Stderr = io.Discard
		cmd.WaitDelay = time.Second
		err := cmd.Run()
		cancel()
		select {
		case <-done:
			return
		case <-ctx.Done():
			return
		default:
		}
		stage := "unreachable"
		if err == nil {
			stage = backgroundWorkerStage(output.Bytes(), job, provider)
		}
		if stage != last || time.Since(confirmed) >= 15*time.Second {
			emitBackground(ctx, stage, job)
			last = stage
			confirmed = time.Now()
		}
		if workerTerminal(stage) {
			return
		}
	}
}
