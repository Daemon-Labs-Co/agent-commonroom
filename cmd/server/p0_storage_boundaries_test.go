package main

import (
	"agentroom/internal/room"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// This exercises the real internal journal decoder, not a new public numeric API.
func observeStorageReservationEncoding(t *testing.T, encoded string) pendingP0Observation {
	t.Helper()
	s, db, ctx := storageFixture(t)
	state, err := s.load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(storageRecord(20, 1, "artifact"))
	if err != nil {
		t.Fatal(err)
	}
	raw = bytes.Replace(raw, []byte(`"bytes":1`), []byte(`"bytes":`+encoded), 1)
	before, _ := json.Marshal(state.ManagedStorage)
	journal := db.snapshot()
	err = projectManagedStorage(&state, "storage.reserved", raw)
	out := pendingP0Observation{State: "allow", Values: map[string]string{}}
	if errors.Is(err, errStorageIdentity) {
		out.State = "deny"
	} else if err != nil {
		t.Fatal("unexpected decoder failure", err)
	}
	after, _ := json.Marshal(state.ManagedStorage)
	if !bytes.Equal(before, after) {
		out.Effects = append(out.Effects, "storage_reserved")
	}
	if !reflect.DeepEqual(journal, db.snapshot()) {
		out.Effects = append(out.Effects, "history_mutated")
	}
	return out
}

func driveP0StorageOverflow(t *testing.T, scenario pendingP0Case) pendingP0Observation {
	t.Helper()
	return observeStorageReservationEncoding(t, scenario.Given["reserve_bytes"])
}

func TestStorageReservationDecoderRejectsUnrepresentableEvidence(t *testing.T) {
	t.Parallel()
	for _, encoded := range []string{"9223372036854775808", "18446744073709551616", "1e100", "1.5", `"1"`} {
		t.Run(encoded, func(t *testing.T) {
			out := observeStorageReservationEncoding(t, encoded)
			if out.State != "deny" || len(out.Effects) != 0 {
				t.Fatal("invalid reservation encoding changed state", out)
			}
		})
	}
	// A valid decoded reservation must actually update the projection.
	if out := observeStorageReservationEncoding(t, "1"); out.State != "allow" || !reflect.DeepEqual(out.Effects, []string{"storage_reserved"}) {
		t.Fatal("valid reservation decoder control failed", out)
	}
}

func TestStorageOverflowEvidenceStopsJournalReplay(t *testing.T) {
	t.Parallel()
	s, db, ctx := storageFixture(t)
	raw, err := json.Marshal(storageRecord(20, 1, "artifact"))
	if err != nil {
		t.Fatal(err)
	}
	raw = bytes.Replace(raw, []byte(`"bytes":1`), []byte(`"bytes":18446744073709551616`), 1)
	// Synthetic append-only evidence with a valid mock hash chain but an invalid
	// numeric payload must be refused by the actual replay boundary.
	db.seed("storage.reserved", json.RawMessage(raw))
	before := db.snapshot()
	if _, err := s.load(ctx); err == nil {
		t.Fatal("overflow evidence was accepted during replay")
	}
	if !reflect.DeepEqual(before, db.snapshot()) {
		t.Fatal("failed replay rewrote evidence")
	}
}

func storageAccountBoundaryFixture(t *testing.T, company string, used int64) (*server, *mockDB, context.Context) {
	t.Helper()
	s, db, _ := mockServer(t)
	db.seed("company.updated", room.Company{ID: company, Name: "Synthetic account"})
	job := quotaJob(1, true)
	job.CompanyID = company
	db.seed("worker.job.updated", job)
	ctx := context.WithValue(t.Context(), actorContextKey{}, ownerID)
	initial := []room.StorageReservation{}
	if used > 0 {
		r := storageRecord(10, used, "workspace")
		r.CompanyID = company
		initial = append(initial, r)
	}
	if err := s.initializeManagedStorage(ctx, company, storageInventoryFunc(func(context.Context, string) (managedStorageInventory, error) {
		return managedStorageInventory{1, company, strings.Repeat("d", 64), initial}, nil
	})); err != nil {
		t.Fatal(err)
	}
	return s, db, ctx
}

// Separate installation journals model independent accounts. This does not
// establish shared-database multi-tenancy or physical cross-worker quotas.
func driveP0StorageCrossAccount(t *testing.T, scenario pendingP0Case) pendingP0Observation {
	t.Helper()
	if scenario.Given["account"] != "b" {
		t.Fatal("unknown target account")
	}
	usedA, errA := strconv.ParseInt(scenario.Given["account_a_used_bytes"], 10, 64)
	usedB, errB := strconv.ParseInt(scenario.Given["account_b_used_bytes"], 10, 64)
	requested, errR := strconv.ParseInt(scenario.Given["reserve_bytes"], 10, 64)
	if errA != nil || errB != nil || errR != nil {
		t.Fatal("invalid account boundary fixture")
	}
	_, dbA, ctxA := storageAccountBoundaryFixture(t, "account-a", usedA)
	b, dbB, ctxB := storageAccountBoundaryFixture(t, "account-b", usedB)
	beforeA, beforeB := dbA.snapshot(), dbB.snapshot()
	r := storageRecord(20, requested, "artifact")
	r.CompanyID = "account-b"
	_, err := b.reserveManagedStorage(ctxB, r)
	out := pendingP0Observation{State: "allow", Values: map[string]string{}}
	if err != nil {
		out.State = "deny"
	}
	for _, event := range dbB.snapshot()[len(beforeB):] {
		if event.kind == "storage.reserved" {
			out.Effects = append(out.Effects, "storage_reserved")
		} else {
			out.Effects = append(out.Effects, event.kind)
		}
	}
	if !reflect.DeepEqual(beforeA, dbA.snapshot()) {
		out.Effects = append(out.Effects, "foreign_account_changed")
	}
	for _, account := range []struct {
		label string
		db    *mockDB
		ctx   context.Context
	}{{"a", dbA, ctxA}, {"b", dbB, ctxB}} {
		state, err := (&server{db: account.db}).load(account.ctx)
		if err != nil {
			t.Fatal(err)
		}
		total, err := managedStorageTotal(state)
		if err != nil {
			t.Fatal(err)
		}
		out.Values["account_"+account.label+"_used_bytes"] = strconv.FormatInt(total, 10)
	}
	return out
}

func TestStorageReservationCannotBorrowAnotherAccountIdentity(t *testing.T) {
	t.Parallel()
	for _, local := range []string{"account-a", "account-b"} {
		t.Run(local, func(t *testing.T) {
			s, db, ctx := storageAccountBoundaryFixture(t, local, 0)
			r := storageRecord(20, 1, "artifact")
			r.CompanyID = "foreign-account"
			before := db.snapshot()
			if _, err := s.reserveManagedStorage(ctx, r); !errors.Is(err, errStorageIdentity) || !reflect.DeepEqual(before, db.snapshot()) {
				t.Fatal("foreign account reserved capacity", err)
			}
		})
	}
}

func TestRetainedWorkspaceReconciliationKeepsStorageAndContainerReservations(t *testing.T) {
	t.Parallel()
	dir, worker, adapter := workspacePair(t)
	if err := worker.Start(t.Context(), adapter); err != nil {
		t.Fatal(err)
	}
	docker := newWorkspaceCleanup(t, worker)
	worker.Close() // Release the synthetic supervisor lease before reconciliation.
	record := worker.record
	for role, id := range map[string]string{"engine": record.Pair.Engine.ID, "tool": record.Pair.Tool.ID} {
		docker.containers[id] = pairFakeContainer{
			name: pairContainerName(record.JobID, role), execution: "separated-" + role + "-v1",
			repository: record.Pair.Repository, job: record.JobID,
		}
	}
	job := quotaJob(1, true)
	job.JobID, job.RoomID, job.ParticipantID, job.ConnectionID = record.JobID, record.RoomID, record.ParticipantID, record.ConnectionID
	job.Provider, job.RepositoryDigest, job.Deadline = record.Provider, record.Pair.Repository, record.Deadline
	job.Status = "submitted"
	s, db, _ := mockServer(t)
	db.seed("worker.job.updated", job)
	r := storageRecord(10, managedStorageLimit, "workspace")
	r.JobID, r.RoomID, r.Worker, r.RepositoryDigest = job.JobID, job.RoomID, job.Worker, job.RepositoryDigest
	r.WorkspaceName, r.WorkspaceNonce, r.WorkspaceCreatedAt = record.Pair.Workspace.Identity.Name, record.Pair.Workspace.Identity.Nonce, record.Pair.Workspace.Identity.CreatedAt
	ctx := context.WithValue(t.Context(), actorContextKey{}, ownerID)
	if err := s.initializeManagedStorage(ctx, job.CompanyID, storageInventoryFunc(func(context.Context, string) (managedStorageInventory, error) {
		return managedStorageInventory{1, job.CompanyID, strings.Repeat("d", 64), []room.StorageReservation{r}}, nil
	})); err != nil {
		t.Fatal(err)
	}
	before := db.snapshot()
	s.jobFence = func(ctx context.Context, alias string, request workerFenceRequest) error {
		if alias != job.Worker || request != managedJobFence(job) || request.DiscardWorkspace {
			t.Fatal("reconciliation changed binding or disposal authority")
		}
		_, err := fenceWorkerJobContainers(ctx, dir, request, func(ctx context.Context, current workerJobRecord) (workerJobRecord, error) {
			return stopWorkerPair(ctx, current, docker)
		})
		return err
	}
	shutdownErr, auditErr := s.reconcileManagedJob(ctx, job)
	if shutdownErr == nil || auditErr != nil {
		t.Fatal("retained source was mistaken for complete shutdown", shutdownErr, auditErr)
	}
	if !slices.Equal(docker.pairFakeDocker.removed, []string{record.Pair.Engine.ID, record.Pair.Tool.ID}) {
		t.Fatal("reconciliation did not remove both exact role IDs", docker.pairFakeDocker.removed)
	}
	stored, err := readWorkerRecord(dir, job.JobID)
	if err != nil || stored.ShutdownConfirmed || !stored.Pair.Engine.Absent || !stored.Pair.Tool.Absent || stored.Pair.Workspace.Removed || !docker.volume || docker.removed != 0 || len(docker.containers) != 0 {
		t.Fatal("reconciliation lost independent role/source evidence", err)
	}
	state, err := (&server{db: db}).load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if total, err := managedStorageTotal(state); err != nil || total != managedStorageLimit || len(state.ManagedStorage.Releases) != 0 || state.ManagedStorage.Reservations[r.ID] != r {
		t.Fatal("retained source received storage credit", total, err)
	}
	if state.ManagedJobs[job.JobID].ShutdownConfirmed || state.ManagedJobs[job.JobID].Status != "shutdown_unconfirmed" {
		t.Fatal("controller fabricated settled lifecycle")
	}
	other := quotaJob(2, true)
	state.ManagedJobs[other.JobID] = other
	if admitManagedContainerQuota(state, quotaJob(3, true)) == nil {
		t.Fatal("retained unresolved pair released physical slot reservations")
	}
	after := db.snapshot()
	if len(after) < len(before) || !reflect.DeepEqual(before, after[:len(before)]) {
		t.Fatal("reconciliation rewrote history")
	}
	for _, event := range after[len(before):] {
		if event.kind == "storage.released" || event.kind == "workspace.disposal.succeeded" {
			t.Fatal("reconciliation fabricated disposal or credit")
		}
	}
}
