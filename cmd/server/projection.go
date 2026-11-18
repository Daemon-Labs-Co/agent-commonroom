package main

// The projection turns the recorded events into the current application state. Each kind of
// event has one reducer (projection_*.go); the projection also keeps the bookkeeping that
// the final presentation step needs (agent activity, failures, acknowledgements and so on).
// It performs no persistence and calls no provider.

import (
	"agentroom/internal/room"
	"time"
)

type reducer func(p *projection, kind string, data []byte) error

type projection struct {
	st                   room.State
	referencedAgentRooms map[string]bool
	activities           map[string]string
	background           map[string][]room.BackgroundActivity
	activityStarts       map[string]time.Time
	failures             map[string]string
	acknowledged         map[string]string
	objectionResolutions map[string]objectionResolution
	dismissedSuggestions map[string]bool
}

func newProjection() *projection {
	return &projection{
		st:                   room.State{Company: room.Company{ID: "company", Name: "My company"}, HumanCompanyID: "company", Workflows: []room.Workflow{}, Rooms: []room.Room{{ID: "engineering", Name: "Engineering room"}}, Tasks: []room.Task{}, Evidence: []room.Evidence{}, Connections: []room.Connection{}, Participants: []room.Participant{}, Memories: []room.Memory{}, Messages: []room.Message{}},
		referencedAgentRooms: map[string]bool{},
		activities:           map[string]string{},
		background:           map[string][]room.BackgroundActivity{},
		activityStarts:       map[string]time.Time{},
		failures:             map[string]string{},
		acknowledged:         map[string]string{},
		objectionResolutions: map[string]objectionResolution{},
		dismissedSuggestions: map[string]bool{},
	}
}

// projectionReducers is built once and never changed.
var projectionReducers = buildReducers()

func buildReducers() map[string]reducer {
	// One entry per recorded event kind; a kind with no entry changes nothing.
	return map[string]reducer{
		"task.repository_links.updated":  reduceTaskRepositoryLinksUpdated,
		"task.acceptance.reviewed":       reduceTaskAcceptanceReviewed,
		"task.execution.started":         reduceTaskExecutionStarted,
		"task.execution.finished":        reduceTaskExecutionStarted,
		"room.outcome.reviewed":          reduceRoomOutcomeReviewed,
		"task.delegation.proposed":       reduceTaskDelegationProposed,
		"task.delegation.approved":       reduceTaskDelegationProposed,
		"task.delegation.rejected":       reduceTaskDelegationProposed,
		"governance.appointed":           reduceGovernanceAppointed,
		"worker.idle_stop.announced":     reduceWorkerIdleStopAnnounced,
		"workspace.disposal.proposed":    reduceWorkspaceDisposalProposed,
		"workspace.disposal.claimed":     reduceWorkspaceDisposalProposed,
		"workspace.disposal.revoked":     reduceWorkspaceDisposalProposed,
		"workspace.disposal.succeeded":   reduceWorkspaceDisposalProposed,
		"workspace.disposal.unconfirmed": reduceWorkspaceDisposalProposed,
		"artifact.reader.acknowledged":   reduceArtifactReaderAcknowledged,
		"workspace.capture.acknowledged": reduceWorkspaceCaptureAcknowledged,
		"workspace.capture.authorized":   reduceWorkspaceCaptureAuthorized,
		"workspace.probe.authorized":     reduceWorkspaceProbeAuthorized,
		"workspace.probe.acknowledged":   reduceWorkspaceProbeAcknowledged,
		"artifact.manifest.recorded":     reduceArtifactManifestRecorded,
		"artifact.custody.reserved":      reduceArtifactCustodyReserved,
		"artifact.custody.ready":         reduceArtifactCustodyReserved,
		"artifact.custody.sealed":        reduceArtifactCustodyReserved,
		"artifact.custody.read":          reduceArtifactCustodyReserved,
		"storage.inventory.recorded":     reduceStorageInventoryRecorded,
		"storage.reserved":               reduceStorageInventoryRecorded,
		"storage.released":               reduceStorageInventoryRecorded,
		"worker.tool.intent":             reduceWorkerToolIntent,
		"worker.tool.outcome":            reduceWorkerToolIntent,
		"operation.receipt":              reduceOperationReceipt,
		"tool.call.intent":               reduceToolCallIntent,
		"tool.call.outcome":              reduceToolCallOutcome,
		"worker.job.updated":             reduceWorkerJobUpdated,
		"invitation.created":             reduceInvitationCreated,
		"invitation.revoked":             reduceInvitationCreated,
		"invitation.redeemed":            reduceInvitationCreated,
		"actor.credential.added":         reduceInvitationCreated,
		"actor.credential.used":          reduceInvitationCreated,
		"actor.credential.removed":       reduceInvitationCreated,
		"actor.oidc.linked":              reduceInvitationCreated,
		"actor.oidc.unlinked":            reduceInvitationCreated,
		"actor.chat_style.updated":       reduceActorChatStyleUpdated,
		"actor.agent_highlight.updated":  reduceActorAgentHighlightUpdated,
		"actor.help.updated":             reduceActorHelpUpdated,
		"actor.photo_size.updated":       reduceActorPhotoSizeUpdated,
		"actor.image.updated":            reduceActorImageUpdated,
		"actor.name.updated":             reduceActorNameUpdated,
		"actor.revised":                  reduceActorRevised,
		"company.organization.updated":   reduceCompanyOrganizationUpdated,
		"company.updated":                reduceCompanyUpdated,
		"human.profile.updated":          reduceHumanProfileUpdated,
		"workflow.run.updated":           reduceWorkflowRunUpdated,
		"workflow.revised":               reduceWorkflowRevised,
		"room.created":                   reduceRoomCreated,
		"room.updated":                   reduceRoomCreated,
		"task.proposal.captured":         reduceTaskProposalCaptured,
		"task.proposal.approved":         reduceTaskProposalCaptured,
		"task.proposal.rejected":         reduceTaskProposalCaptured,
		"task.revised":                   reduceTaskRevised,
		"evidence.created":               reduceEvidenceCreated,
		"connection.created":             reduceConnectionCreated,
		"connection.deleted":             reduceConnectionDeleted,
		"connection.updated":             reduceConnectionUpdated,
		"participant.created":            reduceParticipantCreated,
		"agent.activity":                 reduceAgentActivity,
		"agent.background":               reduceAgentBackground,
		"agent.error":                    reduceAgentError,
		"agent.recovered":                reduceAgentRecovered,
		"participant.updated":            reduceParticipantUpdated,
		"participant.session":            reduceParticipantUpdated,
		"room.discussion.started":        reduceRoomDiscussionStarted,
		"room.discussion.ended":          reduceRoomDiscussionEnded,
		"room.discussion_limit.updated":  reduceRoomDiscussionLimitUpdated,
		"room.context.cleared":           reduceRoomContextCleared,
		"room.context_export.updated":    reduceRoomContextExportUpdated,
		"memory.revised":                 reduceMemoryRevised,
		"participant.removed":            reduceParticipantRemoved,
		"task.suggestion.dismissed":      reduceTaskSuggestionDismissed,
		"objection.resolved":             reduceObjectionResolved,
		"message.attention.acknowledged": reduceMessageAttentionAcknowledged,
		"room.assignment_limit.updated":  reduceRoomAssignmentLimitUpdated,
		"agent.assignment":               reduceAgentAssignment,
		"agent.initiated":                reduceAgentInitiated,
		"message.created":                reduceMessageCreated,
		"agents.idle_policy.updated":     reduceAgentsIdlePolicyUpdated,
		"feature.flag.updated":           reduceFeatureFlagUpdated,
		"agents.stop":                    reduceAgentsStop,
		"agents.resume":                  reduceAgentsStop,
		"incident.opened":                reduceIncidentOpened,
		"incident.recovered":             reduceIncidentRecovered,
	}
}

// apply records one event. An error means the event must never have been recorded.
func (p *projection) apply(kind string, data []byte) error {
	rememberAgentRoomReferences(data, p.referencedAgentRooms)
	p.st.EventCount++
	if reduce := projectionReducers[kind]; reduce != nil {
		return reduce(p, kind, data)
	}
	return nil
}

// finish turns the replayed state into the state the application sees: derived views such as
// agent status, elapsed time, attention and suggestions are computed here, from the replayed
// state, as of now. It changes st, so a projection that will keep replaying must finish a
// copy (see snapshot).
func (p *projection) finish(st *room.State, now time.Time) {
	projectAgentStatusDetails(st, p.activities, p.failures)
	projectAgentElapsed(st, p.activityStarts, now)
	projectBackground(st, p.background)
	for i := range st.Rooms {
		st.Rooms[i].CompanyID = st.Company.ID
	}
	for i := range st.Connections {
		st.Connections[i].CompanyID = st.Company.ID
	}
	for i := range st.Participants {
		st.Participants[i].CompanyID = st.Company.ID
	}
	projectActors(st)
	projectIdleNoticeMessages(st)
	projectHumanAttention(st, p.acknowledged)
	projectEscalations(st, p.acknowledged)
	projectStances(st, p.objectionResolutions)
	projectTaskSuggestions(st, p.dismissedSuggestions)
	omitUnusedAgentRooms(st, p.referencedAgentRooms)
}

// snapshot returns the finished state without disturbing the projection: the replayed
// state is copied first, so later events continue from the unfinished state.
func (p *projection) snapshot(now time.Time) room.State {
	st := p.st.Clone()
	p.finish(&st, now)
	return st
}
