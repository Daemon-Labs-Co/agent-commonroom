package main

import (
	"agentroom/internal/room"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"
)

type verificationBindingKey struct{}
type verificationBinding struct {
	Connection  room.Connection
	Participant room.Participant
}

const connectionVerificationPrompt = "Connection check only. Do not use tools. Reply with exactly COMMONROOM_CONNECTION_OK."

func verificationParticipant(rid, cid string) room.Participant {
	digest := sha256.Sum256([]byte(cid))
	return room.Participant{ID: "verify-" + hex.EncodeToString(digest[:12]), RoomID: rid, ConnectionID: cid}
}

// managedVerification verifies a fixed reply through the journalled protocol-v2
// worker. It cannot export room context, resume a session or mount a repository.
// Source forks CAN bypass registration, BUT lose stop/restart ownership guarantees.
func (s *server) managedVerification(ctx context.Context, c room.Connection, rid string) ([]byte, error) {
	ctx, deadlineCancel := context.WithTimeout(ctx, 20*time.Second)
	defer deadlineCancel()
	p := verificationParticipant(rid, c.ID)
	s.mu.Lock()
	st, err := s.load(ctx)
	if err == nil && !hasRoom(st, rid) {
		err = errors.New("Verification room not found")
	}
	if err == nil && !connectionTestingEnabled(st.FeatureFlags) {
		err = errors.New("Connection testing is disabled")
	}
	if err == nil {
		err = s.workflowSafety(ctx, st, rid)
	}
	if err == nil && (s.active[p.ID] || connectionHasActiveAgent(st, s.active, c.ID)) {
		err = errors.New("Connection has active or unreconciled work; inspect job and task diagnostics before verification")
	}
	if err == nil {
		err = s.agentAdmission(st, rid, 1)
	}
	if err != nil {
		s.mu.Unlock()
		return nil, err
	}
	s.active[p.ID] = true
	if s.verificationRooms == nil {
		s.verificationRooms = map[string]string{}
	}
	s.verificationRooms[p.ID] = rid
	turnCtx, cancel := context.WithCancel(ctx)
	s.cancels[p.ID] = cancel
	s.mu.Unlock()
	defer func() {
		cancel()
		s.mu.Lock()
		delete(s.active, p.ID)
		delete(s.cancels, p.ID)
		delete(s.verificationRooms, p.ID)
		s.mu.Unlock()
	}()
	bound := context.WithValue(turnCtx, verificationBindingKey{}, verificationBinding{Connection: c, Participant: p})
	bound = s.managedJobContext(bound, p)
	key := p.RoomID + "/" + p.ID + "/" + p.ConnectionID + "/" + c.Endpoint
	var adapter sessionAdapter = codexAdapter{endpoint: c.Endpoint, executionKey: key}
	_, provider, endpointErr := remoteWorkerEndpoint(c.Endpoint, os.Getenv("COMMONROOM_REMOTE_WORKER_ALIAS"), os.Getenv("COMMONROOM_REMOTE_WORKER"))
	if endpointErr != nil {
		return nil, endpointErr
	}
	if provider == "claude" {
		adapter = claudeAdapter{endpoint: c.Endpoint, executionKey: key}
	}
	run := func(ctx context.Context) (string, error) {
		return adapter.Run(ctx, "", connectionVerificationPrompt, func(string) error { return nil })
	}
	if s.verificationTurn != nil {
		run = s.verificationTurn
	}
	answer, err := run(bound)
	// A successful reply is insufficient: cancellation and durable shutdown proof
	// must also be checked before describing a connection as working.
	s.mu.Lock()
	latest, loadErr := s.load(context.Background())
	jobID := s.activeJobs[p.ID]
	s.mu.Unlock()
	if loadErr != nil {
		s.mu.Lock()
		s.tripIntegrityLocked()
		s.mu.Unlock()
		return nil, errors.New("Verification audit unavailable")
	}
	registration, _ := bound.Value(managedJobContextKey{}).(managedJobRegistration)
	job, exists := latest.ManagedJobs[registration.JobID]
	if turnCtx.Err() != nil {
		return nil, errors.New("Connection verification cancelled")
	}
	if !exists || !job.ShutdownConfirmed || jobID != "" {
		return nil, errors.New("Verification shutdown is unconfirmed; inspect worker job diagnostics")
	}
	if err != nil {
		return nil, errors.New("Provider verification failed; inspect connection and worker diagnostics")
	}
	return json.Marshal(map[string]any{"version": 1, "connection": c.ID, "provider": provider, "verified": strings.TrimSpace(answer) == "COMMONROOM_CONNECTION_OK"})
}
