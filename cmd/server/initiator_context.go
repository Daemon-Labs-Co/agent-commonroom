package main

import "context"

// detachedInitiatorContext outlives the HTTP request while copying only its
// authenticated actor, never client payloads or request-scoped authority values.
// Missing provenance remains missing; it must never default to the owner.
func detachedInitiatorContext(request context.Context) context.Context {
	actor, _ := request.Value(actorContextKey{}).(string)
	return withInitiatingActor(context.Background(), actor)
}

// withInitiatingActor binds a trusted request identity or recorded workflow
// initiator to execution. Its parent retains workflow cancellation/deadlines.
func withInitiatingActor(parent context.Context, actor string) context.Context {
	// An empty recorded identity must also shadow any inherited actor.
	return context.WithValue(parent, actorContextKey{}, actor)
}
