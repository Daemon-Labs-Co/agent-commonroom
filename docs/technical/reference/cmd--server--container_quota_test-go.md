# cmd/server/container_quota_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/container_quota_test.go)

Automated regression checks for container_quota; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`quotaJob`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(i int, pair bool) room.ManagedJob`.
  Calls: `fmt.Sprintf`, `now.Add`, `strings.Repeat`, `time.Now`, `time.Now().UTC`.
- **`TestManagedContainerQuotaCountsPhysicalReservationsAcrossWorkers`** — Regression contract: TestManagedContainerQuotaCountsPhysicalReservationsAcrossWorkers.
  Signature: `func(t *testing.T)`.
  Calls: `admitManagedContainerQuota`, `quotaJob`, `t.Fatalf`, `t.Parallel`, `t.Run`.
- **`TestManagedContainerQuotaRetainsExpiredAndUnconfirmedReservations`** — Regression contract: TestManagedContainerQuotaRetainsExpiredAndUnconfirmedReservations.
  Signature: `func(t *testing.T)`.
  Calls: `admitManagedContainerQuota`, `fmt.Sprintf`, `job.AllocatedAt.Add`, `job.Deadline.Add`, `quotaJob`, `t.Fatal`, `t.Parallel`.
- **`TestManagedContainerQuotaConservativelyCountsLegacyAttribution`** — Regression contract: TestManagedContainerQuotaConservativelyCountsLegacyAttribution.
  Signature: `func(t *testing.T)`.
  Calls: `admitManagedContainerQuota`, `quotaJob`, `t.Fatal`, `t.Parallel`.
- **`TestAccountContainerQuotaPersistsAcrossRestartAndDisabledFlags`** — Regression contract: TestAccountContainerQuotaPersistsAcrossRestartAndDisabledFlags.
  Signature: `func(t *testing.T)`.
  Calls: `admitManagedContainerQuota`, `append`, `context.Background`, `db.seed`, `fmt.Sprintf`, `mockServer`, `quotaJob`, `registrations[0].Update`, `registrations[4].Register`, `registrations[i].Register`, `restarted.load`, `s.managedJobContext`, `s.managedJobContext(t.Context(), p).Value`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestLastAccountContainerSlotIsAtomicAcrossRegistrations`** — Regression contract: TestLastAccountContainerSlotIsAtomicAcrossRegistrations.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `close`, `db.seed`, `fmt.Sprintf`, `len`, `make`, `mockServer`, `registrations[i].Register`, `s.load`, `s.managedJobContext`, `s.managedJobContext(t.Context(), p).Value`, `t.Context`, `t.Fatal`, `t.Parallel`, `wg.Add`, `wg.Done`, `wg.Wait`.
- **`TestAccountQuotaStaleControllersCannotBothCommitLastSlot`** — Regression contract: TestAccountQuotaStaleControllersCannotBothCommitLastSlot.
  Signature: `func(t *testing.T)`.
  Calls: `admitManagedContainerQuota`, `context.WithValue`, `db.seed`, `len`, `mockServer`, `other.append`, `other.load`, `quotaJob`, `s.append`, `s.load`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestManagedContainerQuotaRejectsUnknownIdentityAndMalformedEvidence`** — Regression contract: TestManagedContainerQuotaRejectsUnknownIdentityAndMalformedEvidence.
  Signature: `func(t *testing.T)`.
  Calls: `admitManagedContainerQuota`, `quotaJob`, `t.Fatal`, `t.Parallel`, `t.Run`.
