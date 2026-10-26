package main

import (
	"agentroom/internal/toolbroker"
	"agentroom/internal/workspacemount"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

type assembledProbeFixture struct {
	t                                 *testing.T
	dir                               string
	record                            workerJobRecord
	base                              *workspaceCleanupDocker
	mode                              string
	id, name, image, challenge        string
	exists, started, removed, stopped bool
	creates                           int
	events                            []workerProbeEvent
	afterFinishObservations           int
	cancel                            context.CancelFunc
}

func (f *assembledProbeFixture) checkLock() {
	f.t.Helper()
	lease, free, err := workerTryLease(f.dir, ".job-"+f.record.JobID+".lease")
	if lease != nil {
		lease.Close()
	}
	if err != nil || free {
		f.t.Fatal("creator lock not held through lifecycle", err)
	}
}
func (f *assembledProbeFixture) AuthorizeWorkspaceObservation(ctx context.Context, r workerJobRecord) error {
	f.checkLock()
	if r.JobID != f.record.JobID {
		f.t.Fatal("wrong creator")
	}
	if f.mode == "observation denied" {
		return errors.New("synthetic denied")
	}
	if len(f.events) > 0 && f.events[len(f.events)-1].Stage == "finish" {
		f.afterFinishObservations++
		if f.mode == "final check denied" && f.afterFinishObservations > 2 {
			return errors.New("synthetic final revocation")
		}
	}
	return ctx.Err()
}
func (f *assembledProbeFixture) Authorize(ctx context.Context, i toolbroker.WorkspaceProbeIntent) error {
	f.checkLock()
	if i.JobID != f.record.JobID || i.Repository != f.record.Pair.Repository || i.Volume != f.record.Pair.Workspace.Identity {
		f.t.Fatal("wrong probe authority")
	}
	if f.mode == "probe denied" || f.mode == "finish denied" && len(f.events) > 0 && f.events[len(f.events)-1].Stage == "finish" {
		return errors.New("synthetic denied")
	}
	return ctx.Err()
}
func (f *assembledProbeFixture) Acknowledge(ctx context.Context, i toolbroker.WorkspaceProbeIntent, e workerProbeEvent) error {
	f.checkLock()
	records, err := readWorkerProbeRecords(f.dir)
	r := records[probeRecordName(i)]
	if err != nil || len(r.Events) == 0 || r.Events[len(r.Events)-1] != e {
		f.t.Fatal("controller acknowledgment before local persistence", err)
	}
	f.events = append(f.events, e)
	if f.mode == "audit "+e.Stage {
		return errors.New("synthetic audit unavailable")
	}
	return ctx.Err()
}
func (f *assembledProbeFixture) StopAuthority() { f.stopped = true }
func (f *assembledProbeFixture) inspection() map[string]any {
	status := "created"
	if f.started {
		status = "exited"
	}
	ro := true
	if f.mode == "writable" && f.started {
		ro = false
	}
	return map[string]any{
		"Id": f.id, "Name": "/" + f.name, "Image": f.image, "RestartCount": 0,
		"State":      map[string]any{"Status": status, "Running": false, "Paused": false, "Restarting": false, "Dead": false, "OOMKilled": false, "ExitCode": 0, "Error": ""},
		"Config":     map[string]any{"Image": f.image, "User": "1000:1000", "WorkingDir": "/workspace", "Tty": false, "OpenStdin": false, "StdinOnce": false, "Env": []string{"PATH=/"}, "Entrypoint": []string{"/opt/commonroom/workspace-mount-probe"}, "Cmd": []string{f.challenge}, "Healthcheck": map[string]any{"Test": []string{"NONE"}}, "Labels": map[string]string{"commonroom.managed": "true", "commonroom.execution": "workspace-mount-probe-v1", "commonroom.job": f.record.JobID, "commonroom.repository": f.record.Pair.Repository, "commonroom.probe": f.challenge}},
		"HostConfig": map[string]any{"Privileged": false, "ReadonlyRootfs": ro, "Init": true, "AutoRemove": false, "RestartPolicy": map[string]string{"Name": "no"}, "NetworkMode": "none", "IpcMode": "private", "CgroupnsMode": "private", "Memory": 67108864, "MemorySwap": 67108864, "NanoCpus": 250000000, "PidsLimit": 16, "CapDrop": []string{"ALL"}, "SecurityOpt": []string{"no-new-privileges:true"}, "LogConfig": map[string]string{"Type": "none"}, "Ulimits": []map[string]any{{"Name": "nofile", "Soft": 64, "Hard": 64}}},
		"Mounts":     []map[string]any{{"Type": "volume", "Name": f.record.Pair.Workspace.Identity.Name, "Source": syntheticObservationDockerRoot + "/volumes/" + f.record.Pair.Workspace.Identity.Name + "/_data", "Destination": "/workspace", "Driver": "local", "RW": false, "Propagation": ""}},
	}
}
func (f *assembledProbeFixture) Run(ctx context.Context, args []string, in []byte, out io.Writer) error {
	f.checkLock()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if len(in) != 0 {
		f.t.Fatal("unexpected input")
	}
	switch args[0] {
	case "create":
		f.creates++
		f.image = args[len(args)-2]
		f.challenge = args[len(args)-1]
		f.name = "commonroom-export-" + f.record.JobID + "-" + f.challenge
		if !slices.Contains(args, "type=volume,src="+f.record.Pair.Workspace.Identity.Name+",dst=/workspace,readonly,volume-nocopy") {
			f.t.Fatal("workspace mount not fixed read-only/no-copy")
		}
		if len(f.events) != 1 || f.events[0].Stage != "begin" {
			f.t.Fatal("creation before durable acknowledgment")
		}
		f.exists = true
		if f.mode == "lost create" {
			return errors.New("synthetic lost reply")
		}
		_, err := io.WriteString(out, f.id)
		return err
	case "inspect":
		if args[len(args)-1] != f.id || !f.exists {
			f.t.Fatal("inspection escaped exact live probe")
		}
		return json.NewEncoder(out).Encode(f.inspection())
	case "start":
		if !slices.Equal(args, []string{"start", "--attach", f.id}) || len(f.events) != 2 || f.events[1].Stage != "bound" {
			f.t.Fatal("start before exact durable binding")
		}
		f.started = true
		if f.mode == "replacement" {
			f.base.mode = "replacement"
		}
		if f.mode == "cancel" {
			f.cancel()
			return ctx.Err()
		}
		p := workspacemount.Proof{Version: 1, Challenge: f.challenge, MountID: 2, ReadOnly: true, RootReadOnly: true}
		if f.mode == "forged proof" {
			p.Challenge = strings.Repeat("f", 32)
		}
		return json.NewEncoder(out).Encode(p)
	case "rm":
		if !slices.Equal(args, []string{"rm", "--force", f.id}) {
			f.t.Fatal("removed wrong container")
		}
		if f.mode == "cleanup failure" {
			return errors.New("synthetic remove failure")
		}
		f.exists = false
		f.removed = true
		return nil
	case "ps":
		filter := args[4]
		answer := ""
		if f.exists && (filter == "id="+f.id || filter == "volume="+f.record.Pair.Workspace.Identity.Name || filter == "label=commonroom.job="+f.record.JobID || filter == "name=^/"+f.name+"$") {
			answer = f.id
		}
		if f.mode == "extra reference" && f.exists && strings.HasPrefix(filter, "volume=") {
			answer += "\n" + strings.Repeat("9", 64)
		}
		if f.mode == "role returns" && f.exists && filter == "name=^/"+pairContainerName(f.record.JobID, "engine")+"$" {
			answer = strings.Repeat("9", 64)
		}
		if f.mode == "late reference" && f.removed && strings.HasPrefix(filter, "volume=") {
			answer = strings.Repeat("9", 64)
		}
		_, err := io.WriteString(out, answer)
		return err
	default:
		return f.base.Run(ctx, args, in, out)
	}
}
func newAssembledProbeFixture(t *testing.T, mode string) *assembledProbeFixture {
	t.Helper()
	dir, r, base := retainedObservationFixture(t)
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	return &assembledProbeFixture{t: t, dir: dir, record: r, base: base, mode: mode, id: strings.Repeat("1", 64)}
}
func TestGuardedWorkspaceProbeIntegratedLifecycle(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"healthy", "observation denied", "probe denied", "audit begin", "audit bound", "audit finish", "extra reference", "role returns", "late reference", "writable", "replacement", "forged proof", "cleanup failure", "cancel", "lost create", "finish denied", "final check denied"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			f := newAssembledProbeFixture(t, mode)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			f.cancel = cancel
			proof, err := probeRetainedWorkerWorkspace(ctx, f.dir, f.record, syntheticObservationDockerRoot, f, "sha256:"+strings.Repeat("d", 64), f)
			if mode == "healthy" {
				if err != nil || !workspacemount.ValidProof(proof, f.challenge) || !f.removed {
					t.Fatalf("healthy assembled lifecycle refused: %v %+v", err, proof)
				}
			} else if err == nil || proof != (workspacemount.Proof{}) {
				t.Fatalf("unsafe result released: %v %+v", err, proof)
			}
			if (mode == "observation denied" || mode == "probe denied" || mode == "audit begin") && f.creates != 0 {
				t.Fatal("created before authorization/audit")
			}
			if (mode == "audit bound" || mode == "extra reference" || mode == "role returns") && f.started {
				t.Fatal("executed before durable binding")
			}
			if strings.HasPrefix(mode, "audit ") && (!f.stopped || !workerProbeStopped(f.dir)) {
				t.Fatal("audit failure did not stop authority")
			}
			if f.base.removed != 0 {
				t.Fatal("probe disposed retained workspace")
			}
			records, e := readWorkerProbeRecords(f.dir)
			if e != nil {
				t.Fatal(e)
			}
			if mode == "healthy" {
				if len(records) != 1 {
					t.Fatal("probe intent missing")
				}
				for _, r := range records {
					if len(r.Events) != 4 || r.Events[3].Stage != "confirmed" {
						t.Fatal("cleanup confirmation missing")
					}
				}
			}
			if mode == "lost create" {
				if !f.exists || f.removed {
					t.Fatal("unknown create adopted")
				}
				for _, r := range records {
					last := r.Events[len(r.Events)-1]
					if last.Stage != "finish" || last.ID != "" || last.Absent {
						t.Fatal("lost create uncertainty discarded")
					}
				}
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
func TestGuardedWorkspaceProbeBusyAndInvalidDependencies(t *testing.T) {
	t.Parallel()
	f := newAssembledProbeFixture(t, "healthy")
	lease, free, e := workerTryLease(f.dir, ".job-"+f.record.JobID+".lease")
	if e != nil || !free {
		t.Fatal(e)
	}
	defer lease.Close()
	if p, e := probeRetainedWorkerWorkspace(t.Context(), f.dir, f.record, syntheticObservationDockerRoot, f, "sha256:"+strings.Repeat("d", 64), f); e == nil || p != (workspacemount.Proof{}) || f.creates != 0 {
		t.Fatal("busy creator admitted")
	}
	var missing *assembledProbeFixture
	if _, e := probeRetainedWorkerWorkspace(t.Context(), f.dir, f.record, syntheticObservationDockerRoot, f, "sha256:"+strings.Repeat("d", 64), missing); e == nil {
		t.Fatal("typed nil authority accepted")
	}
	if _, e := os.Stat(filepath.Join(f.dir, workerProbeStop)); !os.IsNotExist(e) {
		t.Fatal("busy ownership is not integrity corruption")
	}
}

func TestOrdinaryWorkspaceObservationStillRejectsProbeReference(t *testing.T) {
	t.Parallel()
	f := newAssembledProbeFixture(t, "healthy")
	err := withWorkerWorkspaceObservation(t.Context(), f.dir, f.record, syntheticObservationDockerRoot, f, f, func(l *workerWorkspaceObservation) error {
		f.exists = true
		defer func() { f.exists = false }()
		if l.Check(t.Context()) == nil {
			t.Fatal("ordinary observation accepted attached probe")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
