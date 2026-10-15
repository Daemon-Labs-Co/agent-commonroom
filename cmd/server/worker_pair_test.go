package main

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

type pairFakeContainer struct{ name, execution, repository, job string }
type pairFakeDocker struct {
	containers    map[string]pairFakeContainer
	removed       []string
	failRemove    string
	failInventory bool
	replaceName   bool
}

func (f *pairFakeDocker) Run(ctx context.Context, args []string, _ []byte, out io.Writer) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	switch args[0] {
	case "ps":
		if f.failInventory {
			return errors.New("synthetic daemon unavailable")
		}
		filter := args[4]
		ids := []string{}
		for id, c := range f.containers {
			if filter == "id="+id || filter == "name=^/"+c.name+"$" || filter == "label=commonroom.job="+c.job {
				ids = append(ids, id)
			}
		}
		sort.Strings(ids)
		_, err := io.WriteString(out, strings.Join(ids, "\n"))
		return err
	case "inspect":
		id := args[len(args)-1]
		c, ok := f.containers[id]
		if !ok {
			return errors.New("not found")
		}
		_, err := io.WriteString(out, id+"|/"+c.name+"|true|"+c.job+"|"+c.execution+"|"+c.repository)
		return err
	case "rm":
		id := args[len(args)-1]
		if !validPairContainerID(id) {
			return errors.New("cleanup used mutable name")
		}
		if id == f.failRemove {
			return errors.New("synthetic removal failure")
		}
		c := f.containers[id]
		delete(f.containers, id)
		f.removed = append(f.removed, id)
		if f.replaceName {
			f.replaceName = false
			f.containers[strings.Repeat("d", 64)] = c
		}
		return nil
	}
	return errors.New("unexpected Docker invocation")
}
func pairRequest() workerRunRequest {
	r := sampleWorkerRequest()
	r.Version = workerPairLifecycleVersion
	r.Repository = strings.Repeat("b", 64)
	return r
}
func pairFixture(t *testing.T) (string, *workerPairSupervisor, *pairFakeDocker) {
	t.Helper()
	dir := t.TempDir()
	s, err := admitWorkerPair(t.Context(), dir, pairRequest(), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Close)
	runtime := &pairFakeDocker{containers: map[string]pairFakeContainer{}}
	for i, role := range []string{"engine", "tool"} {
		if err := s.BeginCreate(t.Context(), role); err != nil {
			t.Fatal(err)
		}
		id := strings.Repeat([]string{"c", "e"}[i], 64)
		runtime.containers[id] = pairFakeContainer{name: pairContainerName(s.record.JobID, role), execution: "separated-" + role + "-v1", repository: s.record.Pair.Repository, job: s.record.JobID}
		if err := s.BindCreated(t.Context(), role, id); err != nil {
			t.Fatal(err)
		}
	}
	return dir, s, runtime
}
func pairFence(s *workerPairSupervisor) workerFenceRequest {
	return workerFenceRequest{Version: 3, Repository: s.record.Pair.Repository, JobID: s.record.JobID, Provider: s.record.Provider, Key: pairRequest().Key}
}
func fencePair(ctx context.Context, dir string, s *workerPairSupervisor, runtime *pairFakeDocker) (workerJobRecord, error) {
	return fenceWorkerJobContainers(ctx, dir, pairFence(s), func(ctx context.Context, v workerJobRecord) (workerJobRecord, error) {
		return stopWorkerPair(ctx, v, runtime)
	})
}
func TestWorkerPairCreationJournalBindsImmutableRoles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	request := pairRequest()
	s, err := admitWorkerPair(t.Context(), dir, request, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err = s.BindCreated(t.Context(), "engine", strings.Repeat("c", 64)); err == nil {
		t.Fatal("ID bound without durable create intent")
	}
	if err = s.BeginCreate(t.Context(), "engine"); err != nil {
		t.Fatal(err)
	}
	record, err := readWorkerRecord(dir, request.JobID)
	if err != nil || !record.Pair.Engine.CreationStarted || record.Pair.Engine.ID != "" {
		t.Fatal("intent not durable", err)
	}
	id := strings.Repeat("c", 64)
	if err = s.BindCreated(t.Context(), "engine", id); err != nil {
		t.Fatal(err)
	}
	if err = s.BindCreated(t.Context(), "engine", strings.Repeat("d", 64)); err == nil {
		t.Fatal("immutable engine ID replaced")
	}
	if err = s.BeginCreate(t.Context(), "tool"); err != nil {
		t.Fatal(err)
	}
	if err = s.BindCreated(t.Context(), "tool", id); err == nil {
		t.Fatal("roles shared container ID")
	}
	if err = s.BeginCreate(t.Context(), "engine"); err == nil {
		t.Fatal("duplicate creation attempted")
	}
	if err = os.WriteFile(filepath.Join(dir, request.JobID+".cancel"), []byte("cancel"), 0600); err != nil {
		t.Fatal(err)
	}
	if err = s.BindCreated(t.Context(), "tool", strings.Repeat("e", 64)); err == nil {
		t.Fatal("cancelled pair gained identity")
	}
	record, err = readWorkerRecord(dir, request.JobID)
	if err != nil || record.Pair.Engine.ID != id || record.Pair.Tool.ID != "" {
		t.Fatal("journal changed", err)
	}
}
func TestWorkerPairShutdownRequiresBothIDsAndReleasedLease(t *testing.T) {
	t.Parallel()
	dir, s, runtime := pairFixture(t)
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Millisecond)
	defer cancel()
	if _, err := fencePair(ctx, dir, s, runtime); err == nil {
		t.Fatal("pair absence bypassed held supervisor")
	}
	record, err := readWorkerRecord(dir, s.record.JobID)
	if err != nil || record.ShutdownConfirmed {
		t.Fatal("held lease settled", err)
	}
	s.Close()
	record, err = fencePair(t.Context(), dir, s, runtime)
	if err != nil || !record.ShutdownConfirmed || !record.Pair.Engine.Absent || !record.Pair.Tool.Absent {
		t.Fatal("released pair not reconciled", err)
	}
	if len(runtime.removed) != 2 {
		t.Fatal("both roles not stopped by ID", runtime.removed)
	}
	next := pairRequest()
	next.JobID = strings.Repeat("f", 32)
	admitted, err := admitWorkerPair(t.Context(), dir, next, nil)
	if err != nil {
		t.Fatal("confirmed pair blocked replacement", err)
	}
	admitted.Close()
}
func TestWorkerPairPartialShutdownRemainsUnresolved(t *testing.T) {
	t.Parallel()
	dir, s, runtime := pairFixture(t)
	runtime.failRemove = s.record.Pair.Engine.ID
	s.Close()
	record, err := fencePair(t.Context(), dir, s, runtime)
	if err == nil || record.ShutdownConfirmed || record.Pair.Engine.Absent || !record.Pair.Tool.Absent {
		t.Fatal("partial cleanup misreported", record, err)
	}
	saved, err := readWorkerRecord(dir, s.record.JobID)
	if err != nil || saved.ShutdownConfirmed || saved.Status != "shutdown_unconfirmed" {
		t.Fatal("partial failure not durable", err)
	}
	next := pairRequest()
	next.JobID = strings.Repeat("f", 32)
	if admitted, err := admitWorkerPair(t.Context(), dir, next, nil); err == nil {
		admitted.Close()
		t.Fatal("surviving role admitted replacement")
	}
	runtime.failRemove = ""
	if record, err = fencePair(t.Context(), dir, s, runtime); err != nil || !record.ShutdownConfirmed {
		t.Fatal("explicit reconciliation failed", err)
	}
}
func TestWorkerPairCreationCrashRecoversOnlyVerifiedIdentity(t *testing.T) {
	t.Parallel()
	for _, badBinding := range []bool{false, true} {
		t.Run(map[bool]string{false: "matching", true: "foreign repository"}[badBinding], func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			s, err := admitWorkerPair(t.Context(), dir, pairRequest(), nil)
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			if err = s.BeginCreate(t.Context(), "engine"); err != nil {
				t.Fatal(err)
			}
			// Docker-create succeeded but the supervisor died before publishing the ID.
			id := strings.Repeat("c", 64)
			repo := s.record.Pair.Repository
			if badBinding {
				repo = strings.Repeat("d", 64)
			}
			runtime := &pairFakeDocker{containers: map[string]pairFakeContainer{id: {name: pairContainerName(s.record.JobID, "engine"), execution: "separated-engine-v1", repository: repo, job: s.record.JobID}}}
			s.Close()
			record, err := fencePair(t.Context(), dir, s, runtime)
			if badBinding {
				if err == nil || record.ShutdownConfirmed || len(runtime.removed) != 0 {
					t.Fatal("foreign identity removed or settled")
				}
				return
			}
			if err != nil || !record.ShutdownConfirmed || record.Pair.Engine.ID != id || !record.Pair.Engine.Absent || record.Pair.Tool.CreationStarted {
				t.Fatal("creation crash unresolved", err)
			}
		})
	}
}
func TestWorkerPairRefusesDaemonFailureAndNameReplacement(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"daemon unavailable", "name replaced"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			dir, s, runtime := pairFixture(t)
			s.Close()
			runtime.failInventory = mode == "daemon unavailable"
			runtime.replaceName = mode == "name replaced"
			record, err := fencePair(t.Context(), dir, s, runtime)
			if err == nil || record.ShutdownConfirmed {
				t.Fatal("unverified pair marked absent")
			}
			if mode == "name replaced" && len(runtime.removed) != 2 {
				t.Fatal("cleanup skipped unaffected role", runtime.removed)
			}
		})
	}
}
func TestWorkerPairAdmissionCountsLogicalJobsAndRejectsOrphans(t *testing.T) {
	t.Parallel()
	next := fakeWorkerRecord(t, strings.Repeat("f", 32), "another/new/connection/worker")
	records := []workerJobRecord{}
	names := []string{}
	for _, suffix := range []string{"a", "b", "c", "d", "e", "1", "2"} {
		r := fakeWorkerRecord(t, strings.Repeat(suffix, 32), "room/agent"+suffix+"/connection/worker")
		r.Version = workerPairLifecycleVersion
		r.Execution = workerPairExecution
		r.Pair = &workerPairRecord{Version: 1, Repository: strings.Repeat("b", 64)}
		records = append(records, r)
		names = append(names, pairContainerName(r.JobID, "engine"), pairContainerName(r.JobID, "tool"))
	}
	if err := workerAdmission(records, names, next); err != nil {
		t.Fatal("two containers counted as two logical jobs", err)
	}
	if err := workerAdmission(records, append(names, "commonroom-tool-"+strings.Repeat("9", 32)), next); err == nil {
		t.Fatal("orphan admitted")
	}
	r := fakeWorkerRecord(t, strings.Repeat("8", 32), "room/agent8/connection/worker")
	records = append(records, r)
	if err := workerAdmission(records, names, next); err == nil {
		t.Fatal("aggregate job limit bypassed")
	}
}
func TestWorkerPairCannotForgeTerminalOrCrossRoleEvidence(t *testing.T) {
	t.Parallel()
	_, s, _ := pairFixture(t)
	record := s.record
	record.Status = "finished"
	record.ShutdownConfirmed = true
	if validWorkerRecord(record) || workerJobSettled(record) {
		t.Fatal("terminal label substituted for pair absence")
	}
	pair := *record.Pair
	record.Pair = &pair
	pair.Engine.Absent = true
	pair.Tool.Absent = true
	if !validWorkerRecord(record) || !workerJobSettled(record) {
		t.Fatal("valid dual absence refused")
	}
	pair.Tool.ID = pair.Engine.ID
	if validWorkerRecord(record) {
		t.Fatal("cross-role ID alias accepted")
	}
	pair.Tool.ID = strings.Repeat("e", 64)
	pair.Version = 2
	if validWorkerRecord(record) {
		t.Fatal("unknown pair version accepted")
	}
}
func TestWorkerPairRuntimeEntryPointRemainsUnavailable(t *testing.T) {
	t.Parallel()
	request := pairRequest()
	request.ToolSeparation = true
	if err := validateWorkerRun(request); !errors.Is(err, errCredentialBoundary) {
		t.Fatal("lifecycle component enabled development", err)
	}
}

func TestWorkerPairUnknownJobContainerPreventsSettlement(t *testing.T) {
	t.Parallel()
	dir, s, runtime := pairFixture(t)
	s.Close()
	id := strings.Repeat("9", 64)
	runtime.containers[id] = pairFakeContainer{name: "unexpected-supervisor-container", job: s.record.JobID, repository: s.record.Pair.Repository, execution: "unknown"}
	record, err := fencePair(t.Context(), dir, s, runtime)
	if err == nil || record.ShutdownConfirmed {
		t.Fatal("extra job container hidden by dual absence")
	}
	if _, ok := runtime.containers[id]; !ok {
		t.Fatal("cleanup removed unknown container")
	}
	if len(runtime.removed) != 2 {
		t.Fatal("known pair was not stopped")
	}
}

func TestWorkerPairAdmissionRefusesExpiredOrDuplicateJobs(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	request := pairRequest()
	request.Deadline = time.Now().Add(-time.Second)
	if s, err := admitWorkerPair(t.Context(), dir, request, nil); err == nil {
		s.Close()
		t.Fatal("expired pair admitted")
	}
	request = pairRequest()
	s, err := admitWorkerPair(t.Context(), dir, request, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if other, err := admitWorkerPair(t.Context(), dir, request, nil); err == nil {
		other.Close()
		t.Fatal("duplicate pair admitted")
	}
	request.JobID = strings.Repeat("f", 32)
	if other, err := admitWorkerPair(t.Context(), dir, request, nil); err == nil {
		other.Close()
		t.Fatal("membership lease bypassed")
	}
}

func TestWorkerPairRecordCannotDowngradeToLegacyShutdown(t *testing.T) {
	t.Parallel()
	dir, s, _ := pairFixture(t)
	record := s.record
	record.Status = "finished"
	record.ShutdownConfirmed = true
	record.Pair = nil
	if validWorkerRecord(record) || workerJobSettled(record) {
		t.Fatal("pair metadata removal gained legacy shutdown proof")
	}
	if err := writeWorkerRecord(dir, record, false); err == nil {
		t.Fatal("invalid pair downgrade persisted")
	}
	record = s.record
	record.Execution = ""
	if validWorkerRecord(record) {
		t.Fatal("unversioned pair metadata accepted")
	}
	record = s.record
	record.Execution = "future-pair-version"
	if validWorkerRecord(record) {
		t.Fatal("unknown execution version accepted")
	}
}

func TestWorkerPairRequiresExplicitCancellationProtocol(t *testing.T) {
	t.Parallel()
	dir, s, runtime := pairFixture(t)
	s.Close()
	legacy := pairFence(s)
	legacy.Version = 2
	legacy.Repository = ""
	if _, err := fenceWorkerJobContainers(t.Context(), dir, legacy, func(ctx context.Context, v workerJobRecord) (workerJobRecord, error) {
		t.Fatal("legacy cancellation reached pair cleanup")
		return v, nil
	}); err == nil {
		t.Fatal("legacy protocol settled pair")
	}
	if _, err := os.Stat(filepath.Join(dir, s.record.JobID+".cancel")); !os.IsNotExist(err) {
		t.Fatal("legacy request changed pair state")
	}
	request := pairFence(s)
	expected, err := request.record()
	if err != nil {
		t.Fatal(err)
	}
	record, err := fencePair(t.Context(), dir, s, runtime)
	if err != nil {
		t.Fatal(err)
	}
	ack := workerFenceAcknowledgement(request, record)
	if !validWorkerFenceAck(request, expected, ack) {
		t.Fatal("pair acknowledgement refused")
	}
	ack.Version = 2
	if validWorkerFenceAck(request, expected, ack) {
		t.Fatal("legacy acknowledgement substituted for pair proof")
	}
	ack.Version = 3
	ack.Repository = strings.Repeat("9", 64)
	if validWorkerFenceAck(request, expected, ack) {
		t.Fatal("changed repository settled pair")
	}
}

func TestWorkerPairIncompleteCleanupCannotClaimSuccess(t *testing.T) {
	t.Parallel()
	dir, s, _ := pairFixture(t)
	s.Close()
	record, err := fenceWorkerJobContainers(t.Context(), dir, pairFence(s), func(ctx context.Context, v workerJobRecord) (workerJobRecord, error) { return v, nil })
	if err == nil || record.ShutdownConfirmed {
		t.Fatal("missing role absence accepted as success")
	}
	saved, err := readWorkerRecord(dir, s.record.JobID)
	if err != nil || saved.ShutdownConfirmed || saved.Status != "shutdown_unconfirmed" {
		t.Fatal("incomplete proof not retained", err)
	}
}

func TestWorkerPairJournalUsesNonDowngradableLifecycleVersion(t *testing.T) {
	t.Parallel()
	dir, s, _ := pairFixture(t)
	record, err := readWorkerRecord(dir, s.record.JobID)
	if err != nil || record.Version != workerPairLifecycleVersion || validV2WorkerRecord(record) {
		t.Fatal("pair record can be read as legacy lifecycle", err)
	}
	// An old reader explicitly accepts only version 0/2 and therefore rejects this.
	if record.Version == 0 || record.Version == workerLifecycleVersion {
		t.Fatal("pair version not distinguishable by old worker")
	}
	record.Version = workerLifecycleVersion
	if validWorkerRecord(record) {
		t.Fatal("pair silently downgraded to legacy record")
	}
}

func TestWorkerPairCancellationBeforeAdmissionTombstonesPairVersion(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	request := pairRequest()
	fence := workerFenceRequest{Version: 3, JobID: request.JobID, Provider: request.Provider, Key: request.Key, Repository: request.Repository}
	runtime := &pairFakeDocker{containers: map[string]pairFakeContainer{}}
	record, err := fenceWorkerJobContainers(t.Context(), dir, fence, func(ctx context.Context, v workerJobRecord) (workerJobRecord, error) {
		return stopWorkerPair(ctx, v, runtime)
	})
	if err != nil || !record.ShutdownConfirmed || record.Version != workerPairLifecycleVersion || record.Pair == nil {
		t.Fatal("pair tombstone missing", err)
	}
	if admitted, err := admitWorkerPair(t.Context(), dir, request, nil); err == nil {
		admitted.Close()
		t.Fatal("delayed pair admitted after cancellation")
	}
}

func TestWorkerPairCreationCannotOutliveAdmissionContext(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	s, err := admitWorkerPair(ctx, t.TempDir(), pairRequest(), nil)
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	defer s.Close()
	cancel()
	if err := s.BeginCreate(context.Background(), "engine"); err == nil {
		t.Fatal("fresh call context revived cancelled supervisor")
	}
	if s.record.Pair.Engine.CreationStarted {
		t.Fatal("cancelled context changed durable creation state")
	}
}
