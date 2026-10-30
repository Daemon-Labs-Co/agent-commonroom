//go:build linux || darwin

package main

import (
	"agentroom/internal/artifactmanifest"
	"agentroom/internal/objectcustody"
	"agentroom/internal/taskcontract"
	"context"
	"reflect"
)

// recordedManifestSnapshot is private authenticated setup, not a public listing.
// Exact digest pins the version reviewed by the caller. Metadata never proves
// current bytes; actual release MUST still use the approved isolated store guard.
func (s *server) recordedManifestSnapshot(ctx context.Context, execution, output, digest string) (taskcontract.Definition, artifactmanifest.Manifest, error) {
	refuse := func() (taskcontract.Definition, artifactmanifest.Manifest, error) {
		return taskcontract.Definition{}, artifactmanifest.Manifest{}, errArtifactCustody
	}
	if ctx == nil || ctx.Err() != nil || !validWorkerJobID(execution) || !membershipPartPattern.MatchString(output) || !repositoryDigestPattern.MatchString(digest) {
		return refuse()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.integrityFault {
		return refuse()
	}
	st, err := s.load(ctx)
	if err != nil {
		return refuse()
	}
	manifest, ok := st.ArtifactManifests[execution+"/"+output]
	if !ok || manifest.Digest != digest || !artifactActor(st, ctx, manifest.Scope.RoomID) || s.agentsBlockedLocked(st, manifest.Scope.RoomID) {
		return refuse()
	}
	actor, _ := ctx.Value(actorContextKey{}).(string)
	if validateArtifactManifestLinkage(st, artifactManifestEvent{Version: 1, ActorID: actor, PerformedBy: actor, Manifest: manifest}) != nil {
		return refuse()
	}
	manifest.Files = append([]artifactmanifest.File(nil), manifest.Files...)
	return st.TaskExecutions[execution].Definition, manifest, nil
}

// readRecordedArtifactMember revalidates ALL finite members before returning one
// exact declared member's actual sealed bytes. No caller-supplied path/receipt,
// permissive source, host fallback or public route exists. Trusted assembly must
// supply actual isolated custody with independent physical/export/writer fences.
// Never hold server locks across Store.Get; its journal owns current authorization.
// A manifest digest or prior audit record alone can NEVER authorize delivery.
func (s *server) readRecordedArtifactMember(ctx context.Context, execution, output, digest, member string, store *objectcustody.Store) ([]byte, error) {
	refuse := func() ([]byte, error) { return nil, errArtifactCustody }
	if s == nil || ctx == nil || ctx.Err() != nil || store == nil {
		return refuse()
	}
	ctx, cancel := s.storageContext(ctx)
	defer cancel()
	definition, manifest, err := s.recordedManifestSnapshot(ctx, execution, output, digest)
	if err != nil {
		return refuse()
	}
	var receipt objectcustody.Receipt
	found := false
	for _, file := range manifest.Files {
		if file.Name == member {
			receipt = file.Receipt
			found = true
			break
		}
	}
	if !found {
		return refuse()
	} // unknown names must cause NO custody reads
	if artifactmanifest.Revalidate(ctx, definition, manifest.Scope, output, manifest, store) != nil || ctx.Err() != nil {
		return refuse()
	}
	raw, err := store.Get(ctx, receipt.Intent.Scope, receipt)
	if err != nil || ctx.Err() != nil {
		clear(raw)
		return refuse()
	}
	freshDefinition, fresh, err := s.recordedManifestSnapshot(ctx, execution, output, digest)
	if err != nil || ctx.Err() != nil || !reflect.DeepEqual(definition, freshDefinition) || !reflect.DeepEqual(manifest, fresh) {
		clear(raw)
		return refuse()
	}
	return raw, nil
}
