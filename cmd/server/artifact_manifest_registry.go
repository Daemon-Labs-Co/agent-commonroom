package main

import (
	"agentroom/internal/artifactmanifest"
	"agentroom/internal/room"
	"agentroom/internal/taskcontract"
	"bytes"
	"encoding/json"
)

type artifactManifestEvent struct {
	Version     int                       `json:"version"`
	ActorID     string                    `json:"actorId"`
	Manifest    artifactmanifest.Manifest `json:"manifest"`
	PerformedBy string                    `json:"performedBy"`
	Audit       *auditRecordMetadata      `json:"_auditRecord,omitempty"`
}

func artifactManifestKey(m artifactmanifest.Manifest) string {
	return m.Scope.ExecutionID + "/" + m.OutputName
}

// Replay validates immutable linkage/shape only, without filesystem/network I/O.
// Only trusted actual-byte assembly may append this event; public client events
// cannot create it. A recorded manifest is not fresh export/acceptance authority.
func validateArtifactManifestRecord(st room.State, v artifactManifestEvent) error {
	if _, exists := st.ArtifactManifests[artifactManifestKey(v.Manifest)]; exists {
		return errArtifactCustody
	}
	return validateArtifactManifestLinkage(st, v)
}

// Linkage is historical metadata validation, never current byte/export proof.
func validateArtifactManifestLinkage(st room.State, v artifactManifestEvent) error {
	m := v.Manifest
	if v.Version != 1 || v.PerformedBy != v.ActorID || !artifactRecordedActor(st, v.ActorID, m.Scope.RoomID) || st.Incident || st.Company.ID != m.Scope.CompanyID {
		return errArtifactCustody
	}
	execution, ok := st.TaskExecutions[m.Scope.ExecutionID]
	if !ok || artifactmanifest.ValidateMetadata(execution.Definition, m.Scope, m.OutputName, m) != nil {
		return errArtifactCustody
	}
	count := 0
	for _, object := range st.ArtifactCustody {
		if object.Intent.Scope.ExecutionScope == m.Scope && object.Intent.Scope.OutputName == m.OutputName {
			count++
		}
	}
	if count != len(m.Files) {
		return errArtifactCustody
	}
	for _, file := range m.Files {
		object, exists := st.ArtifactCustody[file.Receipt.Intent.ObjectID]
		if !exists || object.ReadAcknowledgments < 1 || object.Receipt == nil || *object.Receipt != file.Receipt {
			return errArtifactCustody
		}
		receipt := file.Receipt
		event := artifactCustodyEvent{Version: 1, ActorID: v.ActorID, Intent: receipt.Intent, Receipt: &receipt}
		if validateArtifactTransition(st, "artifact.custody.read", event) != nil {
			return errArtifactCustody
		}
	}
	return nil
}
func projectArtifactManifest(st *room.State, raw []byte) error {
	if !taskcontract.ValidJSON(raw, 128*1024) {
		return errArtifactCustody
	}
	var event artifactManifestEvent
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&event) != nil || validateArtifactManifestRecord(*st, event) != nil {
		return errArtifactCustody
	}
	if st.ArtifactManifests == nil {
		st.ArtifactManifests = map[string]artifactmanifest.Manifest{}
	}
	event.Manifest.Files = append([]artifactmanifest.File(nil), event.Manifest.Files...)
	st.ArtifactManifests[artifactManifestKey(event.Manifest)] = event.Manifest
	return nil
}
