package main

import (
	"agentroom/internal/objectcustody"
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"agentroom/internal/workspacecapture"
	"agentroom/internal/workspacemount"
	"agentroom/internal/workspacesnapshot"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

type syntheticCaptureBinding struct {
	f     *physicalCaptureFixture
	calls int
	mode  string
}

func (g *syntheticCaptureBinding) BoundCapture(ctx context.Context, st room.State, worker string, r workerJobRecord, q workspacecapture.Request, image string, o toolbroker.CaptureObjectVolume, p workspacemount.Proof) (string, error) {
	g.calls++
	f := g.f
	f.probe.checkLock()
	a, _ := workspacecapture.RequestDigest(q)
	b, _ := workspacecapture.RequestDigest(f.request)
	if worker != "worker" || probeWorkerRecordDigest(r) != probeWorkerRecordDigest(f.probe.record) || a != b || image != f.probe.image || o != f.objects || p.MountID != 2 {
		return "", errArtifactCustody
	}
	if g.mode == "quota denied" || g.mode == "revoked after inspection" && g.calls > 1 {
		return "", errArtifactCustody
	}
	if g.mode == "unknown ID" {
		return "", nil
	}
	if g.mode == "foreign ID" || g.mode == "changed binding" && g.calls > 1 {
		return strings.Repeat("9", 64), nil
	}
	return f.probe.id, ctx.Err()
}

type physicalCaptureFixture struct {
	probe       *assembledProbeFixture
	request     workspacecapture.Request
	objects     toolbroker.CaptureObjectVolume
	mode        string
	inspections int
}

func (f *physicalCaptureFixture) Run(ctx context.Context, args []string, in []byte, out io.Writer) error {
	f.probe.checkLock()
	if ctx.Err() != nil || len(in) != 0 {
		return errArtifactCustody
	}
	p := f.probe
	job := p.record.JobID
	digest, _ := workspacecapture.RequestDigest(f.request)
	switch args[0] {
	case "inspect":
		if args[len(args)-1] != p.id {
			return errArtifactCustody
		}
		f.inspections++
		c := p.inspection()
		c["State"].(map[string]any)["Status"] = "running"
		c["State"].(map[string]any)["Running"] = true
		if f.mode == "created" {
			c["State"].(map[string]any)["Status"] = "created"
			c["State"].(map[string]any)["Running"] = false
		}
		cfg := c["Config"].(map[string]any)
		cfg["OpenStdin"] = true
		cfg["Entrypoint"] = []string{"/opt/commonroom/workspace-capture"}
		raw, _ := json.Marshal(f.request)
		cfg["Cmd"] = []string{base64.RawURLEncoding.EncodeToString(raw)}
		cfg["Labels"] = map[string]string{"commonroom.managed": "true", "commonroom.execution": "workspace-capture-v1", "commonroom.job": job, "commonroom.repository": p.record.Pair.Repository, "commonroom.capture": digest}
		mounts := c["Mounts"].([]map[string]any)
		mounts = append(mounts, map[string]any{"Type": "volume", "Name": f.objects.Name, "Source": syntheticObservationDockerRoot + "/volumes/" + f.objects.Name + "/_data", "Destination": "/objects", "Driver": "local", "RW": true, "Propagation": ""})
		if f.mode == "writable source" {
			mounts[0]["RW"] = true
		}
		c["Mounts"] = mounts
		if f.mode == "record changed" {
			r := p.record
			r.UpdatedAt = r.UpdatedAt.Add(time.Second)
			if err := writeWorkerRecord(p.dir, r, false); err != nil {
				p.t.Fatal(err)
			}
		}
		return json.NewEncoder(out).Encode(c)
	case "ps":
		if f.mode == "vacant" || f.mode == "absent" {
			return nil
		}
		answer := p.id
		if f.mode == "extra reference" {
			answer += "\n" + strings.Repeat("9", 64)
		}
		_, e := io.WriteString(out, answer)
		return e
	case "volume":
		if args[1] == "ls" {
			name := p.record.Pair.Workspace.Identity.Name
			for _, arg := range args {
				if arg == "label=commonroom.execution=object-custody-v1" {
					name = f.objects.Name
				}
			}
			_, e := io.WriteString(out, name)
			return e
		}
		if args[1] != "inspect" {
			p.t.Fatal("volume mutation attempted")
		}
		name, nonce, created, execution, key := p.record.Pair.Workspace.Identity.Name, p.record.Pair.Workspace.Identity.Nonce, p.record.Pair.Workspace.Identity.CreatedAt, "separated-tool-v1", "commonroom.workspace.nonce"
		if args[len(args)-1] == f.objects.Name {
			name = f.objects.Name
			nonce = f.objects.Nonce
			created = f.objects.CreatedAt
			execution = "object-custody-v1"
			key = "commonroom.custody.nonce"
		} else if args[len(args)-1] != name {
			return errArtifactCustody
		}
		if f.mode == "volume replaced" {
			created = "2000-01-01T00:00:00Z"
		}
		return json.NewEncoder(out).Encode(map[string]any{"Name": name, "Driver": "local", "Scope": "local", "CreatedAt": created, "Mountpoint": syntheticObservationDockerRoot + "/volumes/" + name + "/_data", "Labels": map[string]string{"commonroom.managed": "true", "commonroom.execution": execution, "commonroom.job": job, "commonroom.repository": p.record.Pair.Repository, key: nonce}})
	default:
		p.t.Fatal("physical evidence attempted effect", args)
		return errArtifactCustody
	}
}
func TestWorkspaceCapturePhysicalEvidence(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"healthy", "quota denied", "revoked after inspection", "unknown ID", "foreign ID", "changed binding", "writable source", "extra reference", "volume replaced", "record changed", "forged proof"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			probe := newAssembledProbeFixture(t, "")
			s, _, ctx := physicalControllerFixture(t, probe)
			st := probeControllerState(t, s, ctx)
			j := st.ManagedJobs[probe.record.JobID]
			scope := objectcustody.ExecutionScope{CompanyID: st.Company.ID, RoomID: j.RoomID, TaskID: "task", ExecutionID: strings.Repeat("a", 32), JobID: j.JobID, RepositoryDigest: j.RepositoryDigest, DefinitionDigest: strings.Repeat("b", 64)}
			j.TaskID = scope.TaskID
			j.TaskDispatchID = scope.ExecutionID
			j.TaskDefinitionDigest = scope.DefinitionDigest
			st.ManagedJobs[j.JobID] = j
			volume := probe.record.Pair.Workspace.Identity
			q := workspacecapture.Request{Binding: workspacesnapshot.Binding{Execution: scope, BaseCommit: strings.Repeat("c", 40), VolumeName: volume.Name, VolumeNonce: volume.Nonce, VolumeCreatedAt: volume.CreatedAt}, Members: []workspacesnapshot.Member{{OutputName: "result", Name: "result.txt", MaxBytes: 128}}, Challenge: strings.Repeat("e", 32)}
			probe.image = "sha256:" + strings.Repeat("d", 64)
			probe.name = "commonroom-capture-" + j.JobID + "-" + q.Challenge
			f := &physicalCaptureFixture{probe: probe, request: q, objects: toolbroker.CaptureObjectVolume{Name: "commonroom-custody-" + j.JobID, Nonce: strings.Repeat("f", 32), CreatedAt: volume.CreatedAt}, mode: mode}
			lease, free, err := workerTryLease(probe.dir, ".job-"+j.JobID+".lease")
			if err != nil || !free {
				t.Fatal("creator acquisition", err)
			}
			defer lease.Close()
			lifetime, cancel := context.WithCancel(ctx)
			defer cancel()
			// Real synthetic creator lock/retained files; kernel/quota/daemon evidence mocked.
			l := &workerWorkspaceObservation{ownership: lease, active: true, lifetime: lifetime, dir: probe.dir, dockerRoot: syntheticObservationDockerRoot, record: probe.record, runner: f}
			witness := &syntheticCaptureBinding{f: f, mode: mode}
			e, err := newWorkspaceCapturePhysicalEvidence(l, "worker", probe.image, q, f.objects, witness)
			if err != nil {
				t.Fatal(err)
			}
			proof := workspacemount.Proof{Version: 1, Challenge: q.Challenge, MountID: 2, ReadOnly: true, RootReadOnly: true}
			if mode == "forged proof" {
				proof.MountID = 3
			}
			err = e.ValidateCapture(ctx, st, probe.record, q, proof)
			if mode == "healthy" {
				if err != nil || f.inspections != 2 || witness.calls != 2 {
					t.Fatal("physical admission", err, f.inspections, witness.calls)
				}
				cancel()
				if e.ValidateCapture(context.Background(), st, probe.record, q, proof) == nil {
					t.Fatal("escaped creator authority revived")
				}
			} else if err == nil {
				t.Fatal("unsafe physical evidence admitted")
			}
			if mode == "quota denied" && f.inspections != 0 {
				t.Fatal("inspected before quota/binding authorization")
			}
		})
	}
}
func TestWorkspaceCapturePhysicalMissingOwnership(t *testing.T) {
	var missing *workspaceCapturePhysicalEvidence
	if missing.ValidateCapture(t.Context(), room.State{}, workerJobRecord{}, workspacecapture.Request{}, workspacemount.Proof{}) == nil {
		t.Fatal("nil evidence admitted")
	}
	if _, err := newWorkspaceCapturePhysicalEvidence(nil, "worker", "sha256:"+strings.Repeat("a", 64), workspacecapture.Request{}, toolbroker.CaptureObjectVolume{}, nil); !errors.Is(err, errArtifactCustody) {
		t.Fatal("missing ownership admitted", err)
	}
}
