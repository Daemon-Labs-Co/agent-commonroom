package main

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"encoding/json"
	"errors"
	"net/http"
)

func idlePolicyForState(st room.State) (workerIdlePolicy, error) {
	minutes := st.AgentIdleSettings.Minutes
	if minutes == 0 && st.AgentIdleSettings.Revision == 0 {
		minutes = 5
	}
	if minutes < 1 || minutes > 60 || st.AgentIdleSettings.Revision < 0 {
		return workerIdlePolicy{}, errors.New("invalid idle policy")
	}
	enabled, exists := st.FeatureFlags["idleAgentShutdown"]
	return workerIdlePolicy{Enabled: !exists || enabled, Minutes: minutes}, nil
}
func prepareIdleSettings(w http.ResponseWriter, req eventRequest, st room.State) (any, bool) {
	fields, err := fenceObjectFields(req.Payload, "minutes", "revision")
	var v room.AgentIdleSettings
	if err != nil || len(fields) != 2 || json.Unmarshal(req.Payload, &v) != nil || v.Minutes < 1 || v.Minutes > 60 || v.Revision < 0 {
		http.Error(w, "Choose 1–60 idle minutes and the current revision", 400)
		return nil, false
	}
	if v.Revision != st.AgentIdleSettings.Revision {
		http.Error(w, "Idle settings changed; refresh before saving", 409)
		return nil, false
	}
	v.Revision++
	if v.Revision < 1 {
		http.Error(w, "Idle settings revision exhausted", 409)
		return nil, false
	}
	return v, true
}
func projectIdleSettings(st *room.State, raw []byte) error {
	var v room.AgentIdleSettings
	if json.Unmarshal(raw, &v) != nil || v.Minutes < 1 || v.Minutes > 60 || v.Revision < 1 || v.Revision-1 != st.AgentIdleSettings.Revision {
		return errors.New("invalid idle settings transition")
	}
	st.AgentIdleSettings = v
	return nil
}
func validManagedIdlePolicy(job room.ManagedJob) bool {
	if job.IdlePolicyVersion == 0 {
		return job.IdleMinutes == 0 && !job.IdleShutdownEnabled && job.IdlePolicyRevision == 0
	}
	return job.IdlePolicyVersion == 1 && pairedLifecycleVersion(job.Version) && job.IdleMinutes >= 1 && job.IdleMinutes <= 60 && job.IdlePolicyRevision >= 0
}

// configureJobIdle consumes an immutable, durably admitted controller snapshot.
// The reviewed authenticated assembly must pass its exact registered job, never
// room/model-selected data. This method grants no execution/transport authority.
// Omitting it in a source fork loses controller-policy enforcement; legacy local
// assembly retains the worker deployment default until production dispatch lands.
func (s *workerPairSupervisor) configureJobIdle(job room.ManagedJob, runner toolbroker.DockerRunner) error {
	s.mu.Lock()
	match := validManagedJob(job) && job.IdlePolicyVersion == 1 && !job.ShutdownConfirmed && (job.Status == "allocated" || job.Status == "submitted") && sameWorkerBinding(job, workerJobView{workerJobRecord: s.record})
	s.mu.Unlock()
	if !match {
		return errors.New("registered idle policy binding refused")
	}
	return s.configureIdle(workerIdlePolicy{Enabled: job.IdleShutdownEnabled, Minutes: job.IdleMinutes}, runner)
}
