package main

import (
	"agentroom/internal/room"
	"errors"
)

// This is the free edition's mandatory physical-container ceiling. Paid limits
// are not established; there is no flag, worker selection or role-based bypass.
const managedContainerLimit = 4

func managedContainerSlots(job room.ManagedJob) int {
	if pairedLifecycleVersion(job.Version) {
		return 2
	}
	return 1
}

// admitManagedContainerQuota is evaluated against the replayed account journal
// before allocation. The allocation append MUST use that state's EventCount as
// an expected revision: a process-local lock alone cannot serialize controllers.
// Every unresolved allocation reserves its full size, independent of worker,
// room, container inventory or deadline. Only durable shutdown proof releases it.
// Legacy records with no company attribution conservatively count for this
// account; malformed evidence is never permission to allocate more containers.
func admitManagedContainerQuota(st room.State, next room.ManagedJob) error {
	if st.Company.ID == "" || next.CompanyID != st.Company.ID || !validManagedJob(next) || next.ShutdownConfirmed {
		return errors.New("managed container quota identity unavailable")
	}
	slots := managedContainerSlots(next)
	for key, prior := range st.ManagedJobs {
		if key != prior.JobID || !validManagedJob(prior) || prior.JobID == next.JobID {
			return errors.New("managed container quota evidence invalid")
		}
		if prior.CompanyID != "" && prior.CompanyID != st.Company.ID {
			continue
		}
		if !prior.ShutdownConfirmed {
			slots += managedContainerSlots(prior)
		}
		if slots > managedContainerLimit {
			return errors.New("account container capacity reached")
		}
	}
	return nil
}
