package main

import (
	"agentroom/internal/toolbroker"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
)

type syntheticRecoveryFixture struct {
	*assembledProbeFixture
	intent                  toolbroker.WorkspaceProbeIntent
	deny                    bool
	removals, postRecovered int
}

func (f *syntheticRecoveryFixture) AuthorizeWorkspaceObservation(ctx context.Context, r workerJobRecord) error {
	f.checkLock()
	records, e := readWorkerProbeRecords(f.dir)
	if e != nil {
		return e
	}
	if row, ok := records[probeRecordName(f.intent)]; ok && row.Events[len(row.Events)-1].Stage == "recovered" {
		f.postRecovered++
		if f.mode == "final denied" && f.postRecovered > 1 {
			return errors.New("synthetic final revocation")
		}
	}
	return ctx.Err()
}
func (f *syntheticRecoveryFixture) AuthorizeRecovery(ctx context.Context, i toolbroker.WorkspaceProbeIntent, id string) error {
	f.checkLock()
	if i != f.intent || id != strings.Repeat("1", 64) {
		f.t.Fatal("recovery escaped durable identity")
	}
	if f.mode == "reference returns" && f.removed && !f.exists {
		f.exists = true
	}
	if f.deny {
		return errors.New("synthetic recovery denied")
	}
	return ctx.Err()
}
func (f *syntheticRecoveryFixture) Acknowledge(ctx context.Context, i toolbroker.WorkspaceProbeIntent, e workerProbeEvent) error {
	f.checkLock()
	if i != f.intent {
		f.t.Fatal("wrong acknowledged intent")
	}
	records, err := readWorkerProbeRecords(f.dir)
	r := records[probeRecordName(i)]
	found := false
	for _, old := range r.Events {
		if old == e {
			found = true
		}
	}
	if err != nil || !found {
		f.t.Fatal("recovery audit precedes durable evidence", err)
	}
	f.events = append(f.events, e)
	if f.mode == "deny "+e.Stage {
		f.deny = true
	}
	if f.mode == "audit "+e.Stage {
		return errors.New("synthetic audit failure")
	}
	return ctx.Err()
}
func (f *syntheticRecoveryFixture) Run(ctx context.Context, args []string, in []byte, out io.Writer) error {
	if args[0] == "create" || args[0] == "start" {
		f.t.Fatal("recovery launched a helper")
	}
	if args[0] == "rm" {
		f.checkLock()
		records, e := readWorkerProbeRecords(f.dir)
		r := records[probeRecordName(f.intent)]
		last := r.Events[len(r.Events)-1].Stage
		if e != nil || (last != "recovery" && last != "recovery-finish") || len(f.events) == 0 || f.events[0].Stage != "recovery" {
			f.t.Fatal("removal before durable recovery intent/ack", e)
		}
		f.removals++
		if f.mode == "cancel" {
			f.cancel()
			return ctx.Err()
		}
	}
	if args[0] == "inspect" && (f.mode == "running" || f.mode == "foreign") {
		f.checkLock()
		c := f.inspection()
		if f.mode == "running" {
			c["State"].(map[string]any)["Status"] = "running"
			c["State"].(map[string]any)["Running"] = true
		} else {
			c["Image"] = "sha256:" + strings.Repeat("f", 64)
		}
		return json.NewEncoder(out).Encode(c)
	}
	return f.assembledProbeFixture.Run(ctx, args, in, out)
}
func recoveryFixture(t *testing.T, mode string) *syntheticRecoveryFixture {
	t.Helper()
	base := newAssembledProbeFixture(t, mode)
	i := toolbroker.WorkspaceProbeIntent{JobID: base.record.JobID, Repository: base.record.Pair.Repository, Image: "sha256:" + strings.Repeat("d", 64), Challenge: strings.Repeat("e", 32), Volume: base.record.Pair.Workspace.Identity}
	base.name = "commonroom-export-" + i.JobID + "-" + i.Challenge
	base.image = i.Image
	base.challenge = i.Challenge
	base.exists = mode != "absent"
	base.started = true
	j, err := newWorkerProbeJournal(base.dir, &syntheticProbeGuard{})
	if err != nil {
		t.Fatal(err)
	}
	if err = j.Begin(t.Context(), i); err != nil {
		t.Fatal(err)
	}
	if mode == "unknown" {
		if err = j.Finish(t.Context(), i, "", false); err != nil {
			t.Fatal(err)
		}
	} else if mode == "known finish" {
		if err = j.Finish(t.Context(), i, base.id, false); err != nil {
			t.Fatal(err)
		}
	} else {
		if err = j.Bound(t.Context(), i, base.id); err != nil {
			t.Fatal(err)
		}
	}
	f := &syntheticRecoveryFixture{assembledProbeFixture: base, intent: i}
	if mode == "denied" {
		f.deny = true
	}
	if mode == "replacement" {
		f.base.mode = "replacement"
	}
	return f
}
func TestRecordedProbeRecoveryLifecycle(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"exited", "running", "absent", "known finish", "unknown", "foreign", "replacement", "cleanup failure", "denied", "audit recovery", "audit recovery-finish", "deny recovery", "deny recovery-finish", "cancel", "reference returns", "final denied"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			f := recoveryFixture(t, mode)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			f.cancel = cancel
			err := recoverRetainedWorkspaceProbe(ctx, f.dir, f.record, f.intent, syntheticObservationDockerRoot, f, f)
			healthy := mode == "exited" || mode == "running" || mode == "absent" || mode == "known finish"
			if (err == nil) != healthy {
				t.Fatalf("unexpected recovery outcome %v", err)
			}
			records, e := readWorkerProbeRecords(f.dir)
			if e != nil {
				t.Fatal(e)
			}
			r := records[probeRecordName(f.intent)]
			settled := workerProbeSettled(r)
			if healthy {
				if !settled || f.exists || r.Events[len(r.Events)-1].Stage != "recovered" {
					t.Fatal("recovery not confirmed")
				}
				jobs, e := boundedWorkerRecords(f.dir)
				if e != nil || workerProbeCapacity(f.dir, jobs, 2) != nil {
					t.Fatal("confirmed cleanup slot not freed", e)
				}
			} else if mode != "final denied" && settled {
				t.Fatal("uncertain recovery freed capacity")
			}
			if (mode == "unknown" || mode == "foreign" || mode == "replacement" || mode == "denied" || mode == "audit recovery" || mode == "deny recovery") && f.removals != 0 {
				t.Fatal("unauthorized/unknown helper removed")
			}
			if mode == "absent" && f.removals != 0 {
				t.Fatal("absent ID removed again")
			}
			if f.creates != 0 || f.base.removed != 0 {
				t.Fatal("recovery launched or disposed workspace")
			}
			if strings.HasPrefix(mode, "audit ") && (!f.stopped || !workerProbeStopped(f.dir)) {
				t.Fatal("audit failure did not stop admission")
			}
			lease, free, e := workerTryLease(f.dir, ".job-"+f.record.JobID+".lease")
			if lease != nil {
				lease.Close()
			}
			if e != nil || !free {
				t.Fatal("creator lease leaked")
			}
		})
	}
}
func TestRecordedProbeRecoveryExplicitRetryAfterRevocation(t *testing.T) {
	t.Parallel()
	f := recoveryFixture(t, "deny recovery")
	if recoverRetainedWorkspaceProbe(t.Context(), f.dir, f.record, f.intent, syntheticObservationDockerRoot, f, f) == nil || f.removals != 0 {
		t.Fatal("revoked attempt removed helper")
	}
	f.deny = false
	f.mode = "exited"
	if err := recoverRetainedWorkspaceProbe(t.Context(), f.dir, f.record, f.intent, syntheticObservationDockerRoot, f, f); err != nil {
		t.Fatal(err)
	}
	records, e := readWorkerProbeRecords(f.dir)
	r := records[probeRecordName(f.intent)]
	if e != nil || len(r.Events) != 5 || !workerProbeSettled(r) {
		t.Fatal("retry duplicated/lost lifecycle", e)
	}
	if recoverRetainedWorkspaceProbe(t.Context(), f.dir, f.record, f.intent, syntheticObservationDockerRoot, f, f) == nil {
		t.Fatal("settled identity reused")
	}
}
func TestRecordedProbeRecoveryResumesCleanupOutcome(t *testing.T) {
	t.Parallel()
	f := recoveryFixture(t, "deny recovery-finish")
	if recoverRetainedWorkspaceProbe(t.Context(), f.dir, f.record, f.intent, syntheticObservationDockerRoot, f, f) == nil || f.exists {
		t.Fatal("first cleanup outcome")
	}
	f.deny = false
	f.mode = "absent"
	if err := recoverRetainedWorkspaceProbe(t.Context(), f.dir, f.record, f.intent, syntheticObservationDockerRoot, f, f); err != nil {
		t.Fatal(err)
	}
	if f.removals != 1 {
		t.Fatal("retry repeated removal")
	}
}

func TestRecordedProbeRecoveryRefusesBusyStoppedAndWrongIntent(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"busy", "integrity stop", "wrong image", "typed nil"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			f := recoveryFixture(t, "exited")
			i := f.intent
			var guard workerProbeRecoveryGuard = f
			switch mode {
			case "busy":
				lock, free, e := workerTryLease(f.dir, ".job-"+f.record.JobID+".lease")
				if e != nil || !free {
					t.Fatal(e)
				}
				defer lock.Close()
			case "integrity stop":
				if replacePrivateDurably(f.dir+"/"+workerProbeStop, []byte("stop"), syncParentDirectory) != nil {
					t.Fatal("marker")
				}
			case "wrong image":
				i.Image = "sha256:" + strings.Repeat("f", 64)
			case "typed nil":
				var missing *syntheticRecoveryFixture
				guard = missing
			}
			if recoverRetainedWorkspaceProbe(t.Context(), f.dir, f.record, i, syntheticObservationDockerRoot, f, guard) == nil || f.removals != 0 {
				t.Fatal("invalid recovery admitted")
			}
		})
	}
}
