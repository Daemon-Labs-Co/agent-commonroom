package main

import (
	"agentroom/internal/toolbroker"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

type syntheticProbeGuard struct {
	deny, stopped bool
	failStage     string
	events        []workerProbeEvent
	revokeStage   string
}

func (g *syntheticProbeGuard) Authorize(ctx context.Context, _ toolbroker.WorkspaceProbeIntent) error {
	if g.deny {
		return errors.New("synthetic denied")
	}
	return ctx.Err()
}
func (g *syntheticProbeGuard) Acknowledge(ctx context.Context, i toolbroker.WorkspaceProbeIntent, e workerProbeEvent) error {
	if !validWorkerProbeIntent(i) {
		return errors.New("synthetic invalid binding")
	}
	g.events = append(g.events, e)
	if g.revokeStage == e.Stage {
		g.deny = true
	}
	if g.failStage == e.Stage {
		return errors.New("synthetic audit failure")
	}
	return ctx.Err()
}
func (g *syntheticProbeGuard) StopAuthority() { g.stopped = true }
func probeJournalFixture(t *testing.T) (*workerProbeJournal, toolbroker.WorkspaceProbeIntent, *syntheticProbeGuard) {
	t.Helper()
	dir, r, _ := retainedObservationFixture(t)
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	i := toolbroker.WorkspaceProbeIntent{JobID: r.JobID, Repository: r.Pair.Repository, Image: "sha256:" + strings.Repeat("d", 64), Challenge: strings.Repeat("e", 32), Volume: r.Pair.Workspace.Identity}
	g := &syntheticProbeGuard{}
	j, err := newWorkerProbeJournal(dir, g)
	if err != nil {
		t.Fatal(err)
	}
	return j, i, g
}
func TestProbeJournalDurableRestartAndExactBinding(t *testing.T) {
	t.Parallel()
	j, i, g := probeJournalFixture(t)
	ctx := t.Context()
	id := strings.Repeat("1", 64)
	if err := j.Begin(ctx, i); err != nil {
		t.Fatal(err)
	}
	restarted, err := newWorkerProbeJournal(j.dir, g)
	if err != nil {
		t.Fatal(err)
	}
	if restarted.Begin(ctx, i) == nil {
		t.Fatal("duplicate intent admitted after restart")
	}
	other := i
	other.Challenge = strings.Repeat("f", 32)
	if restarted.Begin(ctx, other) == nil {
		t.Fatal("unresolved creator admitted another probe")
	}
	if restarted.Bound(ctx, i, id) != nil {
		t.Fatal("valid binding refused")
	}
	changed := i
	changed.Volume.Nonce = strings.Repeat("f", 32)
	if restarted.Finish(ctx, changed, id, true) == nil {
		t.Fatal("wrong volume accepted")
	}
	if restarted.Finish(ctx, i, strings.Repeat("2", 64), true) == nil {
		t.Fatal("wrong container accepted")
	}
	if restarted.Finish(ctx, i, id, true) != nil {
		t.Fatal("valid cleanup refused")
	}
	records, err := readWorkerProbeRecords(j.dir)
	r := records[probeRecordName(i)]
	if err != nil || len(r.Events) != 4 || r.Events[3].Stage != "confirmed" || len(g.events) != 3 {
		t.Fatalf("lost journal history: %+v %v", r, err)
	}
	if restarted.Begin(ctx, i) == nil {
		t.Fatal("confirmed identity reused")
	}
	if restarted.Begin(ctx, other) != nil {
		t.Fatal("confirmed slot prevented fresh challenge")
	}
}
func TestProbeJournalRetainsUncertaintyAndRevocation(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"unknown", "known unbound", "no absence", "revoked", "audit begin", "audit bound", "audit finish", "disk begin", "disk bound", "disk finish", "disk confirmed"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			j, i, g := probeJournalFixture(t)
			ctx := t.Context()
			id := strings.Repeat("1", 64)
			if mode == "audit begin" {
				g.failStage = "begin"
			}
			if mode == "disk begin" {
				j.persist = func(string, []byte, func(string) error) error { return errors.New("synthetic disk") }
			}
			err := j.Begin(ctx, i)
			if strings.HasSuffix(mode, "begin") {
				if err == nil || !g.stopped || !workerProbeStopped(j.dir) {
					t.Fatal("begin integrity failure remained usable")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if mode == "audit bound" {
				g.failStage = "bound"
			}
			if mode == "disk bound" {
				j.persist = func(string, []byte, func(string) error) error { return errors.New("synthetic disk") }
			}
			if mode != "unknown" && mode != "known unbound" {
				err = j.Bound(ctx, i, id)
				if strings.HasSuffix(mode, "bound") {
					if err == nil || !g.stopped {
						t.Fatal("bound integrity failure ignored")
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
			}
			if mode == "unknown" {
				id = ""
			}
			if mode == "revoked" {
				g.deny = true
			}
			if mode == "audit finish" {
				g.failStage = "finish"
			}
			if strings.HasPrefix(mode, "disk ") {
				j.persist = func(p string, b []byte, syncDir func(string) error) error {
					if mode == "disk confirmed" && !strings.Contains(string(b), "confirmed") {
						return replacePrivateDurably(p, b, syncDir)
					}
					return errors.New("synthetic disk")
				}
			}
			err = j.Finish(ctx, i, id, mode != "unknown" && mode != "no absence")
			if mode == "unknown" || mode == "known unbound" || mode == "no absence" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil {
				t.Fatal("unsafe release acknowledged")
			}
			records, e := readWorkerProbeRecords(j.dir)
			if e != nil {
				t.Fatal(e)
			}
			r := records[probeRecordName(i)]
			if mode != "known unbound" && r.Events[len(r.Events)-1].Stage == "confirmed" {
				t.Fatal("uncertain/revoked attempt freed slot")
			}
			if strings.HasPrefix(mode, "disk") || strings.HasPrefix(mode, "audit") {
				if !g.stopped || !workerProbeStopped(j.dir) {
					t.Fatal("integrity stop absent")
				}
			}
		})
	}
}
func TestProbeJournalPrivateReplayAndCorruption(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"truncated", "unknown field", "duplicate field", "name mismatch", "symlink", "hardlink", "public", "invalid transition"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			j, i, _ := probeJournalFixture(t)
			if j.Begin(t.Context(), i) != nil {
				t.Fatal("begin")
			}
			p := filepath.Join(j.dir, probeRecordName(i))
			raw, err := os.ReadFile(p)
			if err != nil {
				t.Fatal(err)
			}
			switch mode {
			case "truncated":
				raw = raw[:len(raw)/2]
			case "unknown field":
				raw = []byte(strings.Replace(string(raw), "{", "{\"foreign\":1,", 1))
			case "duplicate field":
				raw = []byte(strings.Replace(string(raw), "{", "{\"version\":1,", 1))
			case "name mismatch":
				raw = []byte(strings.ReplaceAll(string(raw), i.Challenge, strings.Repeat("f", 32)))
			case "invalid transition":
				var r workerProbeRecord
				_ = json.Unmarshal(raw, &r)
				r.Events = []workerProbeEvent{{Stage: "confirmed", ID: strings.Repeat("1", 64), Absent: true}}
				raw, _ = json.Marshal(r)
			case "public":
				if os.Chmod(p, 0644) != nil {
					t.Fatal("chmod")
				}
			case "symlink":
				if os.Rename(p, p+".saved") != nil || os.Symlink(p+".saved", p) != nil {
					t.Fatal("symlink")
				}
			case "hardlink":
				if os.Link(p, filepath.Join(t.TempDir(), "linked")) != nil {
					t.Fatal("link")
				}
			}
			if mode != "symlink" && mode != "hardlink" && mode != "public" {
				if os.WriteFile(p, raw, 0600) != nil {
					t.Fatal("write")
				}
			}
			if _, e := readWorkerProbeRecords(j.dir); e == nil {
				t.Fatal("corrupt evidence replayed")
			}
			records, e := boundedWorkerRecords(j.dir)
			if e != nil {
				t.Fatal(e)
			}
			if workerProbeCapacity(j.dir, records, 1) == nil {
				t.Fatal("corruption freed capacity")
			}
		})
	}
}
func TestProbeJournalSerializesCompetingInstances(t *testing.T) {
	t.Parallel()
	j, i, g := probeJournalFixture(t)
	other, e := newWorkerProbeJournal(j.dir, g)
	if e != nil {
		t.Fatal(e)
	}
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, x := range []*workerProbeJournal{j, other} {
		wg.Add(1)
		go func(j *workerProbeJournal) { defer wg.Done(); results <- j.Begin(t.Context(), i) }(x)
	}
	wg.Wait()
	close(results)
	success := 0
	for e := range results {
		if e == nil {
			success++
		}
	}
	if success != 1 {
		t.Fatal("duplicate concurrent intent admitted", success)
	}
}
func TestProbeJournalStopsNormalAdmission(t *testing.T) {
	t.Parallel()
	j, i, g := probeJournalFixture(t)
	g.failStage = "begin"
	if j.Begin(t.Context(), i) == nil {
		t.Fatal("audit failure")
	}
	r := fakeWorkerRecord(t, strings.Repeat("f", 32), "room-a/bot-a/connection-a/job")
	if _, leases, err := admitWorkerRecord(t.Context(), j.dir, r, nil); err == nil {
		leases.Close()
		t.Fatal("normal worker ignored probe integrity stop")
	}
}
func TestProbeJournalRequiresPrivateDependencies(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if os.Chmod(dir, 0700) != nil {
		t.Fatal("chmod")
	}
	var missing *syntheticProbeGuard
	if _, err := newWorkerProbeJournal(dir, missing); err == nil {
		t.Fatal("typed nil guard accepted")
	}
	if os.Chmod(dir, 0755) != nil {
		t.Fatal("chmod")
	}
	if _, err := newWorkerProbeJournal(dir, &syntheticProbeGuard{}); err == nil {
		t.Fatal("public root accepted")
	}
}

func TestProbeJournalCapacityIsRetainedUntilConfirmed(t *testing.T) {
	t.Parallel()
	j, i, _ := probeJournalFixture(t)
	if j.Begin(t.Context(), i) != nil {
		t.Fatal("begin")
	}
	jobs, err := boundedWorkerRecords(j.dir)
	if err != nil {
		t.Fatal(err)
	}
	// The reconciled creator retains two role slots; unresolved probe adds one.
	if workerProbeCapacity(j.dir, jobs, 2) == nil {
		t.Fatal("pending probe omitted from pair capacity")
	}
	if workerProbeCapacity(j.dir, jobs, 1) != nil {
		t.Fatal("conservative capacity miscount")
	}
	if j.Finish(t.Context(), i, strings.Repeat("1", 64), true) != nil {
		t.Fatal("finish")
	}
	if workerProbeCapacity(j.dir, jobs, 2) != nil {
		t.Fatal("confirmed absent probe remained charged locally")
	}
}
func TestProbeJournalAmbiguousDurabilityStopsRestart(t *testing.T) {
	t.Parallel()
	j, i, g := probeJournalFixture(t)
	j.persist = func(p string, b []byte, s func(string) error) error {
		if err := replacePrivateDurably(p, b, s); err != nil {
			return err
		}
		return errors.New("synthetic post-publication sync uncertainty")
	}
	if j.Begin(t.Context(), i) == nil || !g.stopped {
		t.Fatal("ambiguous publication acknowledged")
	}
	records, err := readWorkerProbeRecords(j.dir)
	if err != nil || len(records) != 1 {
		t.Fatal("visible uncertain intent lost", err)
	}
	restarted, err := newWorkerProbeJournal(j.dir, &syntheticProbeGuard{})
	if err != nil {
		t.Fatal(err)
	}
	fresh := i
	fresh.Challenge = strings.Repeat("f", 32)
	if restarted.Begin(t.Context(), fresh) == nil {
		t.Fatal("restart erased integrity stop")
	}
}
func TestProbeJournalCurrentAuthorityAtEveryBoundary(t *testing.T) {
	t.Parallel()
	for _, stage := range []string{"before begin", "after begin", "before bound", "after bound", "before finish"} {
		t.Run(stage, func(t *testing.T) {
			t.Parallel()
			j, i, g := probeJournalFixture(t)
			ctx := t.Context()
			id := strings.Repeat("1", 64)
			if stage == "before begin" {
				g.deny = true
			}
			if stage == "after begin" {
				g.revokeStage = "begin"
			}
			err := j.Begin(ctx, i)
			if stage == "before begin" || stage == "after begin" {
				if err == nil {
					t.Fatal("revoked begin accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if stage == "before bound" {
				g.deny = true
			}
			if stage == "after bound" {
				g.revokeStage = "bound"
			}
			err = j.Bound(ctx, i, id)
			if stage == "before bound" || stage == "after bound" {
				if err == nil {
					t.Fatal("revoked bound accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			g.deny = true
			if j.Finish(ctx, i, id, true) == nil {
				t.Fatal("revoked final release accepted")
			}
			records, err := readWorkerProbeRecords(j.dir)
			r := records[probeRecordName(i)]
			if err != nil || r.Events[len(r.Events)-1].Stage != "finish" {
				t.Fatal("cleanup evidence lost on revocation", err)
			}
		})
	}
}

func TestProbeJournalRevocationDuringConfirmationWithholdsAcknowledgment(t *testing.T) {
	t.Parallel()
	j, i, g := probeJournalFixture(t)
	ctx := t.Context()
	id := strings.Repeat("1", 64)
	if j.Begin(ctx, i) != nil || j.Bound(ctx, i, id) != nil {
		t.Fatal("begin/bound")
	}
	j.persist = func(p string, b []byte, s func(string) error) error {
		err := replacePrivateDurably(p, b, s)
		if strings.Contains(string(b), "confirmed") {
			g.deny = true
		}
		return err
	}
	if j.Finish(ctx, i, id, true) == nil {
		t.Fatal("revocation during publication acknowledged release")
	}
	records, err := readWorkerProbeRecords(j.dir)
	r := records[probeRecordName(i)]
	if err != nil || r.Events[len(r.Events)-1].Stage != "confirmed" {
		t.Fatal("acknowledged absence was lost", err)
	}
	jobs, err := boundedWorkerRecords(j.dir)
	if err != nil || workerProbeCapacity(j.dir, jobs, 2) != nil {
		t.Fatal("removed confirmed container recharged", err)
	}
}
