package main

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"context"
)

// sessionIdleAnnouncer uses only the same private authenticated audit stream as
// the admitted tool session. Frames cannot select sockets/peers or carry actions.
// A source fork can self-acknowledge instead, invalidating durable audit trust.
func sessionIdleAnnouncer(session *toolbroker.ToolSession, job room.ManagedJob) workerIdleAnnounce {
	return func(ctx context.Context, record workerJobRecord) error {
		if session == nil || !validManagedJob(job) || job.IdlePolicyVersion != 1 || !job.IdleShutdownEnabled || !validWorkerRecord(record) || !validWorkerAuditScope(record, session.Scope()) || !sameWorkerBinding(job, workerJobView{workerJobRecord: record}) || record.Status != "running" || record.DiagnosticCode != "idle_shutdown_requested" || record.IdleTimeoutMinutes != job.IdleMinutes {
			return errIdleNotice
		}
		stop := toolbroker.IdleStop{StartedAt: record.StartedAt.UTC().Round(0), RequestedAt: record.IdleShutdownRequestedAt.UTC().Round(0), Minutes: job.IdleMinutes, PolicyRevision: job.IdlePolicyRevision}
		if session.RecordIdleStop(ctx, stop) != nil {
			return errIdleNotice
		}
		return nil
	}
}
