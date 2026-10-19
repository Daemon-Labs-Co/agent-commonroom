package main

import (
	"agentroom/internal/toolbroker"
	"context"
	"errors"
	"path/filepath"
	"time"
)

// workerCleanupOwnershipKey is set only after acquiring the creator job lease.
// Killing a creator's containers while its lease is held does not authorize
// erasing its workspace or confirming completed cleanup.
type workerCleanupOwnershipKey struct{}
type workerWorkspaceDiscardKey struct{}

func (s *workerPairSupervisor) prepareWorkspace(ctx context.Context, adapter workerWorkspaceLaunchAdapter) error {
	s.mu.Lock()
	err := s.available(ctx)
	already := s.record.Pair.Workspace != nil
	job := s.record.JobID
	s.mu.Unlock()
	if err != nil || already {
		return errors.New("workspace creation unavailable")
	}
	intent, err := adapter.WorkspaceIntent(ctx)
	if err != nil || !toolbroker.ValidWorkspaceVolume(job, intent, false) {
		return errors.New("invalid workspace intent")
	}
	if err = s.publishWorkspace(ctx, &workerPairWorkspace{CreationStarted: true, Identity: intent}, false); err != nil {
		return err
	}
	identity, err := adapter.PrepareVolume(ctx, intent)
	if err != nil {
		return errors.New("workspace creation unconfirmed")
	}
	if !toolbroker.ValidWorkspaceVolume(job, identity, true) || identity.Version != intent.Version || identity.Name != intent.Name || identity.Nonce != intent.Nonce {
		return errors.New("workspace identity changed")
	}
	return s.publishWorkspace(ctx, &workerPairWorkspace{CreationStarted: true, Identity: identity}, true)
}
func (s *workerPairSupervisor) publishWorkspace(ctx context.Context, workspace *workerPairWorkspace, bind bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.available(ctx); err != nil {
		return err
	}
	record := s.record
	pair := *record.Pair
	if record.Version != workerWorkspaceLifecycleVersion || (pair.Workspace != nil) != bind {
		return errors.New("workspace journal transition refused")
	}
	if bind && (pair.Workspace.Identity.Nonce != workspace.Identity.Nonce || pair.Workspace.Identity.CreatedAt != "") {
		return errors.New("workspace journal identity immutable")
	}
	pair.Workspace = workspace
	record.Pair = &pair
	record.UpdatedAt = time.Now().UTC()
	if !validWorkerRecord(record) {
		return errors.New("invalid workspace journal")
	}
	if err := writeWorkerRecord(s.dir, record, false); err != nil {
		s.closed = true
		s.cancel()
		return err
	}
	s.record = record
	return nil
}

func stopWorkerWorkspace(ctx context.Context, v workerJobRecord, runner toolbroker.DockerRunner) (workerJobRecord, error) {
	owned, _ := ctx.Value(workerCleanupOwnershipKey{}).(bool)
	if !owned {
		return v, nil
	}
	root, err := pairCapture(ctx, runner, []string{"info", "--format", "{{.DockerRootDir}}"})
	if err != nil || !filepath.IsAbs(root) || filepath.Clean(root) != root || root == "/" {
		return v, errors.New("workspace daemon root unverified")
	}
	var identity *toolbroker.WorkspaceVolumeIdentity
	if v.Pair.Workspace != nil {
		copy := v.Pair.Workspace.Identity
		identity = &copy
	}
	expected, _ := ctx.Value(workerWorkspaceDiscardKey{}).(*toolbroker.WorkspaceVolumeIdentity)
	discard := expected != nil
	if discard && (identity == nil || *expected != *identity || !toolbroker.ValidWorkspaceVolume(v.JobID, *expected, true)) {
		return v, errors.New("workspace disposal binding changed")
	}
	observed, err := toolbroker.RemoveWorkspaceVolume(ctx, runner, v.JobID, v.Pair.Repository, root, identity, discard)
	if v.Pair.Workspace != nil {
		copy := *v.Pair.Workspace
		if observed != nil {
			copy.Identity = *observed
		}
		copy.Removed = err == nil
		v.Pair.Workspace = &copy
	}
	if err != nil {
		return v, errors.New("workspace cleanup unverified")
	}
	return v, nil
}

func sameWorkspaceEvidence(before, after workerJobRecord) bool {
	if before.Pair == nil || after.Pair == nil {
		return before.Pair == nil && after.Pair == nil
	}
	a, b := before.Pair.Workspace, after.Pair.Workspace
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.CreationStarted == b.CreationStarted && a.Identity.Version == b.Identity.Version && a.Identity.Name == b.Identity.Name && a.Identity.Nonce == b.Identity.Nonce && (a.Identity.CreatedAt == "" || a.Identity.CreatedAt == b.Identity.CreatedAt) && (!a.Removed || b.Removed)
}
