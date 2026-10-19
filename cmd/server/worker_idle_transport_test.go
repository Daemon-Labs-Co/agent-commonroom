package main

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"context"
	"net"
	"strings"
	"testing"
)

type idleTransportAuditor struct {
	*controllerToolAuditor
	drop bool
	peer net.Conn
}

func (a idleTransportAuditor) RecordIdleStop(ctx context.Context, scope toolbroker.Scope, stop toolbroker.IdleStop) error {
	err := a.controllerToolAuditor.RecordIdleStop(ctx, scope, stop)
	if err == nil && a.drop {
		a.peer.Close()
		return toolbroker.ErrAudit
	}
	return err
}
func TestIdleAcknowledgmentAcrossPrivateWorkerTransport(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"success", "notices-off", "audit-failure", "lost-ack", "wrong-revision", "old-controller"} {
		t.Run(mode, func(t *testing.T) {
			f := newIdleNoticeFixture(t, true)
			if mode == "notices-off" {
				f.db.seed("feature.flag.updated", map[string]any{"name": "idleStopNotices", "enabled": false})
				f.db.seed("feature.flag.updated", map[string]any{"name": "idleAgentShutdown", "enabled": false})
			}
			job := f.job
			job.Status = "submitted"
			f.db.seed("worker.job.updated", job)
			f.db.seed("connection.created", room.Connection{ID: job.ConnectionID, Protocol: "CLI", Provider: "openai", Endpoint: "ssh://worker", Revision: 1})
			f.db.seed("participant.created", room.Participant{ID: job.ParticipantID, RoomID: job.RoomID, ConnectionID: job.ConnectionID, Revision: 1})
			scope := toolbroker.Scope{Version: 1, JobID: job.JobID, Room: job.RoomID, Participant: job.ParticipantID, Connection: job.ConnectionID, Provider: job.Provider, RepositoryDigest: job.RepositoryDigest, Deadline: job.Deadline}
			auditor, err := newControllerToolAuditor(t.Context(), f.s, scope, ownerID, func() (string, error) { return job.RepositoryDigest, nil })
			if err != nil {
				t.Fatal(err)
			}
			listener, err := f.worker.OpenAuditListener(scope)
			if err != nil {
				t.Fatal(err)
			}
			defer listener.Close()
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			controller, relay := net.Pipe()
			defer controller.Close()
			defer relay.Close()
			attached := make(chan error, 1)
			served := make(chan error, 1)
			go func() { attached <- workerAttachAudit(ctx, f.worker.dir, scope, relay, relay) }()
			session, _, err := f.worker.AcceptToolSession(ctx, listener, []string{"read"}, "ghcr.io/example/tool@sha256:"+strings.Repeat("a", 64), "/trusted/docker", f.docker)
			if err != nil {
				t.Fatal(err)
			}
			defer session.Close()
			var serving toolbroker.Auditor = idleTransportAuditor{controllerToolAuditor: auditor, drop: mode == "lost-ack", peer: controller}
			if mode == "old-controller" {
				serving = &transportTestAuditor{}
			}
			if mode == "audit-failure" {
				f.db.failAt = "insert"
			}
			go func() { served <- toolbroker.ServeAudit(ctx, controller, scope, serving) }()
			workerJob := job
			if mode == "wrong-revision" {
				workerJob.IdlePolicyRevision++
			}
			f.worker.mu.Lock()
			f.worker.idle.announce = sessionIdleAnnouncer(session, workerJob)
			f.worker.mu.Unlock()
			expired, stopErr := f.worker.expireIdle(f.now)
			// Join controller/relay work before inspecting or resetting shared fixtures.
			cancel()
			session.Close()
			controller.Close()
			relay.Close()
			awaitAuditResult(t, attached)
			awaitAuditResult(t, served)
			if mode == "audit-failure" {
				f.db.failAt = ""
			}
			// Present workspace keeps full lifecycle proof unresolved even on success.
			if !expired || stopErr == nil {
				t.Fatal("idle stop invented full shutdown proof")
			}
			state, err := f.s.load(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			wantNotices := 0
			if mode == "success" || mode == "notices-off" || mode == "lost-ack" {
				wantNotices = 1
			}
			if len(state.IdleStopNotices) != wantNotices || state.ManagedJobs[job.JobID].ShutdownConfirmed || !f.docker.volume || f.docker.removed != 0 {
				t.Fatal("transport lost attribution or source", mode)
			}
			if mode == "notices-off" && len(state.Messages) != 0 {
				t.Fatal("notice visibility bypassed over transport")
			}
			removed := len(f.docker.pairFakeDocker.removed)
			success := mode == "success" || mode == "notices-off"
			if success && removed != 2 || !success && removed != 0 {
				t.Fatal("cleanup was not gated by durable acknowledgment", mode, removed)
			}
			if expired, _ := f.worker.expireIdle(f.now); expired {
				t.Fatal("failed transport automatically retried")
			}
		})
	}
}
