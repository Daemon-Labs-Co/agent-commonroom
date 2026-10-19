package main

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
)

// A reservation may prebind the worker-generated identity before allocation. This
// is accounting evidence, never permission to write without a hard quota. Legacy
// unbound charges cannot be inferred from job IDs or refunded retrospectively.
func validStorageWorkspace(r room.StorageReservation) bool {
	if r.WorkspaceName == "" && r.WorkspaceNonce == "" && r.WorkspaceCreatedAt == "" {
		return true
	}
	return r.Category == "workspace" && toolbroker.ValidWorkspaceVolume(r.JobID, storageWorkspace(r), true)
}
func storageWorkspace(r room.StorageReservation) toolbroker.WorkspaceVolumeIdentity {
	return toolbroker.WorkspaceVolumeIdentity{Version: 1, Name: r.WorkspaceName, Nonce: r.WorkspaceNonce, CreatedAt: r.WorkspaceCreatedAt}
}
func reservationMatchesDisposal(r room.StorageReservation, p room.WorkspaceDisposal) bool {
	j := p.Job
	return validStorageReservation(r) && r.Category == "workspace" && r.JobID == j.JobID && r.CompanyID == j.CompanyID && r.RoomID == j.RoomID && r.Worker == j.Worker && r.RepositoryDigest == j.RepositoryDigest && storageWorkspace(r) == disposalIdentity(p)
}
func disposalStorageReservation(st room.State, p room.WorkspaceDisposal) string {
	if st.ManagedStorage == nil {
		return ""
	}
	var match string
	for id, r := range st.ManagedStorage.Reservations {
		if reservationMatchesDisposal(r, p) {
			if _, released := st.ManagedStorage.Releases[id]; released {
				return ""
			}
			if match != "" {
				return ""
			} // Ambiguous legacy inventory grants no credit.
			match = id
		}
	}
	return match
}
func validStorageRelease(st room.State, release room.StorageRelease) bool {
	if release.Version != 1 || !validWorkerJobID(release.ReservationID) || !validWorkerJobID(release.DisposalID) || st.ManagedStorage == nil {
		return false
	}
	r, ok := st.ManagedStorage.Reservations[release.ReservationID]
	p, found := st.WorkspaceDisposals[release.DisposalID]
	job, registered := st.ManagedJobs[p.Job.JobID]
	return ok && found && registered && validDisposal(p) && p.Status == "succeeded" && p.StorageReservationID == r.ID && reservationMatchesDisposal(r, p) && sameDisposalJob(job, p.Job) && job.ShutdownConfirmed
}
