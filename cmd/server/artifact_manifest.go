//go:build linux || darwin

package main

import (
	"agentroom/internal/artifactmanifest"
	"agentroom/internal/objectcustody"
	"agentroom/internal/taskcontract"
	"context"
	"reflect"
	"sort"
)

// artifactManifestClaims selects exact sealed receipts from replayed controller
// custody, never provider/model claims. It grants no byte/source authority; every
// read still requires the trusted store's current journal and physical guards.
func (s *server) artifactManifestClaims(ctx context.Context, scope objectcustody.ExecutionScope, output string) (taskcontract.Definition, []artifactmanifest.Claim, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	refuse := func() (taskcontract.Definition, []artifactmanifest.Claim, error) {
		return taskcontract.Definition{}, nil, errArtifactCustody
	}
	if ctx == nil || ctx.Err() != nil || s.integrityFault || !objectcustody.ValidExecutionScope(scope) {
		return refuse()
	}
	st, err := s.load(ctx)
	if err != nil || st.Incident || st.Company.ID != scope.CompanyID || !artifactActor(st, ctx, scope.RoomID) || s.agentsBlockedLocked(st, scope.RoomID) {
		return refuse()
	}
	execution, ok := st.TaskExecutions[scope.ExecutionID]
	if !ok || execution.DefinitionDigest != scope.DefinitionDigest || taskcontract.Digest(execution.Definition) != scope.DefinitionDigest {
		return refuse()
	}
	actor, _ := ctx.Value(actorContextKey{}).(string)
	var claims []artifactmanifest.Claim
	for key, object := range st.ArtifactCustody {
		if object.Intent.Scope.ExecutionScope != scope || object.Intent.Scope.OutputName != output {
			continue
		}
		if len(claims) >= 32 || key != object.Intent.ObjectID || object.Receipt == nil {
			return refuse()
		}
		receipt := *object.Receipt
		event := artifactCustodyEvent{Version: 1, ActorID: actor, Intent: object.Intent, Receipt: &receipt}
		if validateArtifactTransition(st, "artifact.custody.read", event) != nil {
			return refuse()
		}
		claims = append(claims, artifactmanifest.Claim{Name: object.Intent.Scope.MemberName, Receipt: receipt})
	}
	if len(claims) == 0 {
		return refuse()
	}
	sort.Slice(claims, func(i, j int) bool { return claims[i].Name < claims[j].Name })
	return execution.Definition, claims, nil
}

// resolveStoredArtifactManifest joins the durable controller's receipt selection
// with actual private stored bytes and finite structural validation. Trusted
// assembly supplies a real isolated store/journal; no default/permissive adapter,
// path selection, public download, task completion or acceptance is introduced.
// Never hold server locks across Store.Get: its journal reacquires them. Recheck
// access, integrity and the COMPLETE receipt set after all member acknowledgments.
// Matching metadata is not a writer fence or independent source/custody proof.
func (s *server) resolveStoredArtifactManifest(ctx context.Context, scope objectcustody.ExecutionScope, output string, store *objectcustody.Store) (artifactmanifest.Manifest, error) {
	refuse := func() (artifactmanifest.Manifest, error) { return artifactmanifest.Manifest{}, errArtifactCustody }
	if s == nil || store == nil || ctx == nil || ctx.Err() != nil {
		return refuse()
	}
	ctx, cancel := s.storageContext(ctx)
	defer cancel()
	definition, claims, err := s.artifactManifestClaims(ctx, scope, output)
	if err != nil {
		return refuse()
	}
	manifest, err := artifactmanifest.Resolve(ctx, definition, scope, output, claims, store)
	if err != nil || ctx.Err() != nil {
		return refuse()
	}
	freshDefinition, freshClaims, err := s.artifactManifestClaims(ctx, scope, output)
	if err != nil || ctx.Err() != nil || !reflect.DeepEqual(definition, freshDefinition) || !reflect.DeepEqual(claims, freshClaims) {
		return refuse()
	}
	return manifest, nil
}
