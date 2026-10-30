//go:build linux || darwin

package main

import (
	"agentroom/internal/artifactmanifest"
	"agentroom/internal/objectcustody"
	"context"
	"errors"
)

// recordStoredArtifactManifest privately assembles actual bytes then durably
// records the immutable manifest once for this execution/output. No client/model
// receipt payload, task completion, acceptance or byte download is introduced.
// Ambiguous append outcomes never return success or automatically retry/adopt.
func (s *server) recordStoredArtifactManifest(ctx context.Context, scope objectcustody.ExecutionScope, output string, store *objectcustody.Store) (artifactmanifest.Manifest, error) {
	refuse := func() (artifactmanifest.Manifest, error) { return artifactmanifest.Manifest{}, errArtifactCustody }
	if s == nil || ctx == nil || ctx.Err() != nil || store == nil {
		return refuse()
	}
	ctx, cancel := s.storageContext(ctx)
	defer cancel()
	manifest, err := s.resolveStoredArtifactManifest(ctx, scope, output, store)
	if err != nil {
		return refuse()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.load(ctx)
	if err != nil || ctx.Err() != nil || s.integrityFault || !artifactActor(st, ctx, scope.RoomID) || s.agentsBlockedLocked(st, scope.RoomID) {
		return refuse()
	}
	actor, _ := ctx.Value(actorContextKey{}).(string)
	event := artifactManifestEvent{Version: 1, ActorID: actor, PerformedBy: actor, Manifest: manifest}
	if validateArtifactManifestRecord(st, event) != nil {
		return refuse()
	}
	if err = s.append(context.WithValue(ctx, expectedEventCountKey{}, st.EventCount), "artifact.manifest.recorded", event); err != nil {
		if !errors.Is(err, errAuditConflict) {
			s.tripIntegrityLocked()
		}
		return artifactmanifest.Manifest{}, err
	}
	return manifest, nil
}
