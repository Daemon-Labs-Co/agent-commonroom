package main

import (
	"agentroom/internal/toolbroker"
	"agentroom/internal/workspacemount"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
)

// workerWorkspaceProbeGuard binds trusted current controller authorization,
// durable observation/audit and account-wide admission. No production transport
// or permissive implementation exists. A model cannot supply this authority.
type workerWorkspaceProbeGuard interface {
	workerWorkspaceObservationGuard
	workerProbeGuard
}

// guardedWorkspaceProbe is confined to the synchronous creator-lock callback.
// Its journal methods track the one exact helper ID; no handle escapes to callers.
type guardedWorkspaceProbe struct {
	active  bool
	intent  toolbroker.WorkspaceProbeIntent
	id      string
	lease   *workerWorkspaceObservation
	probe   *toolbroker.WorkspaceMountProbe
	journal *workerProbeJournal
	guard   workerWorkspaceProbeGuard
}

func (g *guardedWorkspaceProbe) Authorize(ctx context.Context, i toolbroker.WorkspaceProbeIntent) error {
	if !g.active || i != g.intent || g.lease.checkProbe(ctx, g.probe, g.id) != nil {
		return errWorkerProbeJournal
	}
	return g.guard.Authorize(ctx, i)
}
func (g *guardedWorkspaceProbe) Acknowledge(ctx context.Context, i toolbroker.WorkspaceProbeIntent, e workerProbeEvent) error {
	if !g.active || i != g.intent {
		return errWorkerProbeJournal
	}
	return g.guard.Acknowledge(ctx, i, e)
}
func (g *guardedWorkspaceProbe) StopAuthority() { g.guard.StopAuthority() }
func (g *guardedWorkspaceProbe) Begin(ctx context.Context, i toolbroker.WorkspaceProbeIntent) error {
	if !g.active || i != g.intent {
		return errWorkerProbeJournal
	}
	return g.journal.Begin(ctx, i)
}
func (g *guardedWorkspaceProbe) Bound(ctx context.Context, i toolbroker.WorkspaceProbeIntent, id string) error {
	if !g.active || i != g.intent || g.id != "" || g.probe.VerifyAdmission(ctx, id) != nil {
		return errWorkerProbeJournal
	}
	g.id = id
	return g.journal.Bound(ctx, i, id)
}
func (g *guardedWorkspaceProbe) Finish(ctx context.Context, i toolbroker.WorkspaceProbeIntent, id string, absent bool) error {
	if !g.active || i != g.intent || g.id != "" && g.id != id || id != "" && !probeHex(id, 32) || absent && id == "" {
		return errWorkerProbeJournal
	}
	// Absence is first proven by the runtime's exact-ID cleanup. The journal's
	// final checks then independently require zero references and unchanged volume.
	if absent {
		g.id = ""
	} else {
		g.id = id
	}
	return g.journal.Finish(ctx, i, id, absent)
}

// probeRetainedWorkerWorkspace assembles creator ownership, exact retained
// workspace admission, durable local journal and isolated read-only helper.
// Caller supplies only trusted reviewed image/daemon/controller infrastructure.
// No public route, source bytes, host mount/fallback or FrozenLease is returned.
// A source fork CAN substitute dishonest guards/runners, BUT loses these claims.
func probeRetainedWorkerWorkspace(ctx context.Context, dir string, record workerJobRecord, dockerRoot string, runner toolbroker.DockerRunner, image string, guard workerWorkspaceProbeGuard) (workspacemount.Proof, error) {
	var proof workspacemount.Proof
	if ctx == nil || ctx.Err() != nil || nilWorkerObservationDependency(guard) {
		return proof, errors.New("workspace probe dependencies unavailable")
	}
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return proof, err
	}
	err := withWorkerWorkspaceObservation(ctx, dir, record, dockerRoot, runner, guard, func(lease *workerWorkspaceObservation) error {
		intent := toolbroker.WorkspaceProbeIntent{JobID: lease.record.JobID, Repository: lease.record.Pair.Repository, Volume: lease.record.Pair.Workspace.Identity, Image: image, Challenge: hex.EncodeToString(nonce)}
		g := &guardedWorkspaceProbe{active: true, intent: intent, lease: lease, guard: guard}
		defer func() { g.active = false }()
		journal, err := newWorkerProbeJournal(dir, g)
		if err != nil {
			return err
		}
		g.journal = journal
		probe, err := toolbroker.NewWorkspaceMountProbe(intent, runner, g, dockerRoot)
		if err != nil {
			return err
		}
		g.probe = probe
		proof, err = probe.Run(ctx)
		return err
	})
	if err != nil {
		return workspacemount.Proof{}, err
	}
	return proof, nil
}
