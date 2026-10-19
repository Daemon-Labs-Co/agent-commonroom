package main

import (
	"agentroom/internal/room"
	"context"
	"encoding/json"
	"errors"
	"maps"
	"strings"
	"testing"
	"time"
)

func TestBoundWorkspaceStorageDisposal(t *testing.T) {
	t.Setenv("COMMONROOM_REMOTE_WORKER", "true")
	t.Setenv("COMMONROOM_REMOTE_WORKER_ALIAS", "worker")
	for _, mode := range []string{"verified", "worker-unknown", "outcome-rollback", "credit-rollback"} {
		t.Run(mode, func(t *testing.T) {
			f := newDisposalFixture(t, true, true)
			p := f.propose(t)
			if p.StorageReservationID == "" {
				t.Fatal("exact reservation not reviewed")
			}
			f.now.Store(p.NotBefore.UnixNano())
			original := f.s.jobFence
			f.s.jobFence = func(ctx context.Context, alias string, r workerFenceRequest) error {
				if err := original(ctx, alias, r); err != nil {
					return err
				}
				if mode == "worker-unknown" {
					return errors.New("uncertain")
				}
				if mode == "outcome-rollback" {
					f.db.failInsertAt = 2
				}
				if mode == "credit-rollback" {
					f.db.failInsertAt = 3
				}
				return nil
			}
			result := f.approve(t, p)
			f.db.failInsertAt = 0
			st, err := (&server{db: f.db}).load(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			total, err := managedStorageTotal(st)
			if err != nil {
				t.Fatal(err)
			}
			if mode == "verified" {
				if result.Code != 200 || total != 0 || len(st.ManagedStorage.Releases) != 1 || len(st.ManagedStorage.Reservations) != 1 {
					t.Fatal("credit/history mismatch", result.Code, total)
				}
				if again := f.approve(t, p); again.Code != 409 || f.calls.Load() != 1 {
					t.Fatal("repeat refund")
				}
				for _, release := range st.ManagedStorage.Releases {
					raw, _ := json.Marshal(release)
					if projectManagedStorage(&st, "storage.released", raw) == nil {
						t.Fatal("duplicate release replay accepted")
					}
				}
			} else {
				if result.Code != 503 || total != 42 || len(st.ManagedStorage.Releases) != 0 {
					t.Fatal("uncertain outcome refunded", result.Code, total)
				}
				if mode != "worker-unknown" && st.ManagedJobs[p.Job.JobID].ShutdownConfirmed {
					t.Fatal("partial outcome committed")
				}
			}
		})
	}
}

func TestStorageReleaseRejectsForgedEvidence(t *testing.T) {
	t.Parallel()
	s, _, ctx := storageFixture(t, storageRecord(10, 42, "workspace"))
	st, err := s.load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	r := room.StorageRelease{Version: 1, ReservationID: storageRecord(10, 42, "workspace").ID, DisposalID: storageRecord(20, 1, "workspace").ID}
	raw, _ := json.Marshal(r)
	if projectManagedStorage(&st, "storage.released", raw) == nil {
		t.Fatal("refund without verified owner disposal")
	}
}

func TestStorageWorkspaceIdentityValidation(t *testing.T) {
	t.Parallel()
	r := storageRecord(10, 42, "workspace")
	r.WorkspaceName = "commonroom-tool-work-" + r.JobID
	r.WorkspaceNonce = strings.Repeat("e", 32)
	r.WorkspaceCreatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	for _, mode := range []string{"valid", "partial", "foreign-job", "artifact", "duplicate-name"} {
		t.Run(mode, func(t *testing.T) {
			candidate := r
			switch mode {
			case "partial":
				candidate.WorkspaceNonce = ""
			case "foreign-job":
				candidate.JobID = strings.Repeat("f", 32)
			case "artifact":
				candidate.Category = "artifact"
			}
			if mode == "duplicate-name" {
				st := room.State{Company: room.Company{ID: r.CompanyID}}
				duplicate := r
				duplicate.ID = strings.Repeat("f", 32)
				raw, _ := json.Marshal(managedStorageInventory{1, r.CompanyID, strings.Repeat("d", 64), []room.StorageReservation{r, duplicate}})
				if projectManagedStorage(&st, "storage.inventory.recorded", raw) == nil {
					t.Fatal("same physical workspace charged twice")
				}
			} else if validStorageReservation(candidate) != (mode == "valid") {
				t.Fatal("workspace identity validation", mode)
			}
		})
	}
}

func TestStorageReleaseScopeCannotChange(t *testing.T) {
	t.Setenv("COMMONROOM_REMOTE_WORKER", "true")
	t.Setenv("COMMONROOM_REMOTE_WORKER_ALIAS", "worker")
	f := newDisposalFixture(t, true, true)
	p := f.propose(t)
	f.now.Store(p.NotBefore.UnixNano())
	if result := f.approve(t, p); result.Code != 200 {
		t.Fatal(result.Code, result.Body.String())
	}
	original, err := f.s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"nonce", "reservation", "worker", "repository", "unconfirmed", "no-shutdown", "foreign-release", "bundle"} {
		t.Run(mode, func(t *testing.T) {
			st := original
			ledger := *original.ManagedStorage
			ledger.Reservations = maps.Clone(ledger.Reservations)
			st.ManagedStorage = &ledger
			st.WorkspaceDisposals = maps.Clone(original.WorkspaceDisposals)
			st.ManagedJobs = maps.Clone(original.ManagedJobs)
			release := ledger.Releases[p.StorageReservationID]
			reservation := ledger.Reservations[p.StorageReservationID]
			outcome := st.WorkspaceDisposals[p.ID]
			job := st.ManagedJobs[p.Job.JobID]
			switch mode {
			case "nonce":
				reservation.WorkspaceNonce = strings.Repeat("f", 32)
			case "reservation":
				outcome.StorageReservationID = strings.Repeat("f", 32)
			case "worker":
				reservation.Worker = "other"
			case "repository":
				reservation.RepositoryDigest = strings.Repeat("f", 64)
			case "unconfirmed":
				outcome.Status = "unconfirmed"
			case "no-shutdown":
				job.ShutdownConfirmed = false
			case "foreign-release":
				release.DisposalID = strings.Repeat("f", 32)
			case "bundle":
				reservation.Category = "bundle"
			}
			ledger.Reservations[reservation.ID] = reservation
			st.WorkspaceDisposals[p.ID] = outcome
			st.ManagedJobs[job.JobID] = job
			if validStorageRelease(st, release) {
				t.Fatal("forged refund", mode)
			}
			if mode != "foreign-release" {
				if _, err := managedStorageTotal(st); err == nil {
					t.Fatal("corrupt refunded ledger loaded", mode)
				}
			}
		})
	}
	// A refund never makes the reservation ID or physical identity reusable.
	duplicate := original.ManagedStorage.Reservations[p.StorageReservationID]
	if _, err := admitManagedStorage(original, duplicate); err == nil {
		t.Fatal("released reservation reused")
	}
	duplicate.ID = strings.Repeat("f", 32)
	if _, err := admitManagedStorage(original, duplicate); err == nil {
		t.Fatal("released physical identity reused")
	}
	fresh := storageRecord(20, managedStorageLimit, "artifact")
	if total, err := admitManagedStorage(original, fresh); err != nil || total != managedStorageLimit {
		t.Fatal("verified refund did not restore capacity", total, err)
	}
}
