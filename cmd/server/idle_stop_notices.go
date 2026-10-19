package main

import (
	"agentroom/internal/room"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var errIdleNotice = errors.New("idle stop announcement unavailable")

func validIdleNotice(n room.IdleStopNotice) bool {
	j := n.Job
	return n.Version == 1 && validWorkerJobID(n.ID) && validManagedJob(j) && j.IdlePolicyVersion == 1 && j.IdleShutdownEnabled && !j.ShutdownConfirmed && (j.Status == "allocated" || j.Status == "submitted") && !n.WorkerStartedAt.IsZero() && !n.WorkerStartedAt.Before(j.AllocatedAt.Add(-5*time.Second)) && !n.RequestedAt.Before(n.WorkerStartedAt.Add(time.Duration(j.IdleMinutes)*time.Minute)) && n.RequestedAt.Before(j.Deadline) && !n.RecordedAt.IsZero() && !n.RequestedAt.After(n.RecordedAt.Add(5*time.Second))
}
func projectIdleNotice(st *room.State, raw []byte) error {
	var n room.IdleStopNotice
	if json.Unmarshal(raw, &n) != nil || !validIdleNotice(n) || n.Job.CompanyID != st.Company.ID {
		return errIdleNotice
	}
	job, ok := st.ManagedJobs[n.Job.JobID]
	if !ok || !sameDisposalJob(job, n.Job) || job.ShutdownConfirmed {
		return errIdleNotice
	}
	if _, exists := st.IdleStopNotices[job.JobID]; exists {
		return errIdleNotice
	}
	if st.IdleStopNotices == nil {
		st.IdleStopNotices = map[string]room.IdleStopNotice{}
	}
	st.IdleStopNotices[job.JobID] = n
	st.Messages = append(st.Messages, idleNoticeMessage(n))
	return nil
}

// newIdleAnnouncer binds a registered job and initiating human to a trusted
// private callback. This is not a public/model/plugin API or peer authentication.
// Production assembly must carry it over a reviewed authenticated control stream.
func (s *server) newIdleAnnouncer(ctx context.Context, jobID string) (workerIdleAnnounce, error) {
	if ctx == nil || ctx.Err() != nil {
		return nil, errIdleNotice
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.load(ctx)
	if err != nil || s.integrityFault {
		return nil, errIdleNotice
	}
	job, ok := st.ManagedJobs[jobID]
	actor, _ := ctx.Value(actorContextKey{}).(string)
	principal, memberFound := member(st, actor)
	if !ok || !validManagedJob(job) || job.IdlePolicyVersion != 1 || !job.IdleShutdownEnabled || job.ShutdownConfirmed || (job.Status != "allocated" && job.Status != "submitted") || actor != job.InitiatorID || !memberFound || !principal.Enabled || principal.Kind != "human" || !canAccess(st, actor, job.RoomID) {
		return nil, errIdleNotice
	}
	return func(action context.Context, record workerJobRecord) error {
		if action == nil || action.Err() != nil || !validWorkerRecord(record) || !sameWorkerBinding(job, workerJobView{workerJobRecord: record}) || record.DiagnosticCode != "idle_shutdown_requested" || record.Status != "running" || record.IdleTimeoutMinutes != job.IdleMinutes {
			return errIdleNotice
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		latest, err := s.load(action)
		if err != nil || s.integrityFault {
			return errIdleNotice
		}
		current, found := latest.ManagedJobs[job.JobID]
		if !found || !sameDisposalJob(current, job) {
			return errIdleNotice
		}
		if prior, exists := latest.IdleStopNotices[job.JobID]; exists {
			if prior.WorkerStartedAt.Equal(record.StartedAt) && prior.RequestedAt.Equal(record.IdleShutdownRequestedAt) {
				return nil
			}
			return errIdleNotice
		}
		notice := room.IdleStopNotice{Version: 1, ID: id(), Job: job, WorkerStartedAt: record.StartedAt, RequestedAt: record.IdleShutdownRequestedAt, RecordedAt: time.Now().UTC()}
		if !validIdleNotice(notice) || current.ShutdownConfirmed {
			return errIdleNotice
		}
		// Attribution comes from the immutable human/job binding, not callback input.
		audit := context.WithValue(action, actorContextKey{}, job.InitiatorID)
		if err := s.append(context.WithValue(audit, expectedEventCountKey{}, latest.EventCount), "worker.idle_stop.announced", notice); err != nil {
			if !errors.Is(err, errAuditConflict) {
				s.tripIntegrityLocked()
			}
			return errIdleNotice
		}
		return nil
	}, nil
}
func idleNoticeMessage(n room.IdleStopNotice) room.Message {
	return room.Message{ID: "idle-stop-" + n.ID, RoomID: n.Job.RoomID, ActorID: "commonroom-system", ActorType: "system", Author: "CommonRoom", Created: n.RecordedAt, Text: fmt.Sprintf("Idle shutdown requested for job %s after %d idle minutes under its recorded policy. This notice does not verify cleanup. Idle shutdown retains source; no capacity release or restart is authorized.", n.Job.JobID, n.Job.IdleMinutes)}
}
func projectIdleNoticeMessages(st *room.State) {
	enabled, exists := st.FeatureFlags["idleStopNotices"]
	if !exists || enabled {
		return
	}
	ids := map[string]bool{}
	for _, n := range st.IdleStopNotices {
		ids["idle-stop-"+n.ID] = true
	}
	kept := st.Messages[:0]
	for _, message := range st.Messages {
		if !(message.ActorType == "system" && message.ActorID == "commonroom-system" && ids[message.ID]) {
			kept = append(kept, message)
		}
	}
	st.Messages = kept
}
