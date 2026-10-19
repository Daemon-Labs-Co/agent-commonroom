package main

import (
	"agentroom/internal/room"
	"context"
	"encoding/json"
	"errors"
	"math"
	"time"
)

// Mandatory decimal free-account ceiling. Paid limits are not defined; neither
// payment, owner authority nor optional product flags can increase this budget.
const managedStorageLimit int64 = 5_000_000_000

var (
	errStorageUnknown  = errors.New("account storage inventory unavailable")
	errStorageCapacity = errors.New("account storage capacity reached")
	errStorageIdentity = errors.New("account storage identity invalid")
)

type managedStorageInventory struct {
	Version      int                       `json:"version"`
	CompanyID    string                    `json:"companyId"`
	Digest       string                    `json:"digest"`
	Reservations []room.StorageReservation `json:"reservations"`
}

// This trusted adapter must freeze writers and inventory ALL retained volumes,
// staged bundles and artifacts across the account's approved workers. It cannot
// use model claims, polling samples or a partial worker list as complete evidence.
// No production adapter exists yet: unknown inventory must block reservations.
// Source forks CAN supply a dishonest adapter, invalidating accounting/isolation.
type managedStorageInventorySource interface {
	Inventory(context.Context, string) (managedStorageInventory, error)
}

func validStorageReservation(r room.StorageReservation) bool {
	return validStorageWorkspace(r) && r.Version == 1 && workerAliasPattern.MatchString(r.CompanyID) && validWorkerJobID(r.ID) && validWorkerJobID(r.JobID) && workerAliasPattern.MatchString(r.RoomID) && workerAliasPattern.MatchString(r.Worker) && repositoryDigestPattern.MatchString(r.RepositoryDigest) && r.Bytes > 0 && (r.Category == "workspace" || r.Category == "bundle" || r.Category == "artifact")
}

// Storage totals charge reserved capacity, including retained/uncertain data,
// without trusting current process/container status. Never sum unchecked int64s.
func managedStorageTotal(st room.State) (int64, error) {
	l := st.ManagedStorage
	if l == nil {
		return 0, errStorageUnknown
	}
	if l.Version != 1 || l.CompanyID != st.Company.ID || !workerAliasPattern.MatchString(l.CompanyID) || !repositoryDigestPattern.MatchString(l.InventoryDigest) || l.Reservations == nil {
		return 0, errStorageIdentity
	}
	var total int64
	names := map[string]bool{}
	for key, r := range l.Reservations {
		if key != r.ID || !validStorageReservation(r) || r.CompanyID != l.CompanyID {
			return 0, errStorageIdentity
		}
		if r.WorkspaceName != "" {
			if names[r.WorkspaceName] {
				return 0, errStorageIdentity
			}
			names[r.WorkspaceName] = true
		}
		if _, released := l.Releases[key]; !released {
			if r.Bytes > math.MaxInt64-total {
				return 0, errStorageIdentity
			}
			total += r.Bytes
		}
	}
	for key, release := range l.Releases {
		if key != release.ReservationID || !validStorageRelease(st, release) {
			return 0, errStorageIdentity
		}
	}
	return total, nil
}
func admitManagedStorage(st room.State, r room.StorageReservation) (int64, error) {
	if !validStorageReservation(r) || r.CompanyID != st.Company.ID {
		return 0, errStorageIdentity
	}
	total, err := managedStorageTotal(st)
	if err != nil {
		return 0, err
	}
	if _, exists := st.ManagedStorage.Reservations[r.ID]; exists {
		return 0, errStorageIdentity
	}
	if total > managedStorageLimit || r.Bytes > managedStorageLimit-total {
		return 0, errStorageCapacity
	}
	if r.WorkspaceName != "" {
		for _, prior := range st.ManagedStorage.Reservations {
			if prior.WorkspaceName == r.WorkspaceName {
				return 0, errStorageIdentity
			}
		}
	}
	return total + r.Bytes, nil
}

// initializeManagedStorage records one complete trusted inventory before capacity
// may be reserved. It is internal infrastructure, NOT an HTTP/model operation or
// a way to reset usage. Over-limit existing data is preserved and blocks growth.
// No resources are created here, and no physical quota guarantee is implied.
func (s *server) initializeManagedStorage(ctx context.Context, company string, source managedStorageInventorySource) error {
	if ctx == nil || ctx.Err() != nil || source == nil {
		return errStorageUnknown
	}
	ctx, cancel := s.storageContext(ctx)
	defer cancel()
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.load(ctx)
	if err != nil {
		return err
	}
	actor, _ := ctx.Value(actorContextKey{}).(string)
	a, ok := member(st, actor)
	if s.integrityFault || actor != ownerID || !ok || !a.Enabled || a.Kind != "human" || company != st.Company.ID {
		return errStorageIdentity
	}
	if st.ManagedStorage != nil {
		return errStorageIdentity
	}
	inventory, err := source.Inventory(ctx, company)
	if err != nil || ctx.Err() != nil {
		return errStorageUnknown
	}
	// Validate via exactly the same transition used on restart/replay.
	candidate := st
	raw, err := json.Marshal(inventory)
	if err != nil {
		return errStorageIdentity
	}
	if err = projectManagedStorage(&candidate, "storage.inventory.recorded", raw); err != nil {
		return err
	}
	return s.append(context.WithValue(ctx, expectedEventCountKey{}, st.EventCount), "storage.inventory.recorded", inventory)
}

// reserveManagedStorage commits capacity before any external allocation. The
// expected journal revision is checked under the transactional audit-chain lock,
// so separate controllers cannot both reserve the last byte from stale state.
// Returning success is not permission to write without a physical quota backend.
// No refund/retry is inferred after timeout, failed append or container removal.
func (s *server) reserveManagedStorage(ctx context.Context, r room.StorageReservation) (int64, error) {
	if ctx == nil || ctx.Err() != nil {
		return 0, errStorageIdentity
	}
	ctx, cancel := s.storageContext(ctx)
	defer cancel()
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.load(ctx)
	if err != nil {
		return 0, err
	}
	actor, _ := ctx.Value(actorContextKey{}).(string)
	a, ok := member(st, actor)
	job, found := st.ManagedJobs[r.JobID]
	if !ok || !a.Enabled || a.Kind != "human" || !canAccess(st, actor, r.RoomID) || !found || !validManagedJob(job) || job.ShutdownConfirmed || (job.Status != "allocated" && job.Status != "submitted") || !time.Now().Before(job.Deadline) || job.CompanyID != r.CompanyID || job.RoomID != r.RoomID || job.Worker != r.Worker || job.RepositoryDigest != r.RepositoryDigest || (actor != ownerID && job.InitiatorID != actor) || s.agentsBlockedLocked(st, r.RoomID) {
		return 0, errStorageIdentity
	}
	total, err := admitManagedStorage(st, r)
	if err != nil {
		return 0, err
	}
	if err = s.append(context.WithValue(ctx, expectedEventCountKey{}, st.EventCount), "storage.reserved", r); err != nil {
		return 0, err
	}
	return total, nil
}

func projectManagedStorage(st *room.State, kind string, raw []byte) error {
	switch kind {
	case "storage.inventory.recorded":
		var inventory managedStorageInventory
		if st.ManagedStorage != nil || json.Unmarshal(raw, &inventory) != nil || inventory.Version != 1 || inventory.CompanyID != st.Company.ID || inventory.Reservations == nil || !repositoryDigestPattern.MatchString(inventory.Digest) {
			return errStorageIdentity
		}
		ledger := &room.ManagedStorage{Version: 1, CompanyID: inventory.CompanyID, InventoryDigest: inventory.Digest, Reservations: map[string]room.StorageReservation{}}
		for _, r := range inventory.Reservations {
			if _, exists := ledger.Reservations[r.ID]; exists {
				return errStorageIdentity
			}
			ledger.Reservations[r.ID] = r
		}
		candidate := *st
		candidate.ManagedStorage = ledger
		if _, err := managedStorageTotal(candidate); err != nil {
			return err
		}
		st.ManagedStorage = ledger
	case "storage.released":
		var release room.StorageRelease
		if json.Unmarshal(raw, &release) != nil || !validStorageRelease(*st, release) {
			return errStorageIdentity
		}
		if _, exists := st.ManagedStorage.Releases[release.ReservationID]; exists {
			return errStorageIdentity
		}
		if st.ManagedStorage.Releases == nil {
			st.ManagedStorage.Releases = map[string]room.StorageRelease{}
		}
		st.ManagedStorage.Releases[release.ReservationID] = release
	case "storage.reserved":
		var r room.StorageReservation
		if json.Unmarshal(raw, &r) != nil {
			return errStorageIdentity
		}
		if _, err := admitManagedStorage(*st, r); err != nil {
			return err
		}
		st.ManagedStorage.Reservations[r.ID] = r
	default:
		return errStorageIdentity
	}
	return nil
}
