package main

import (
	"agentroom/internal/room"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type syntheticObjectReadWitness struct {
	operations, lifecycles int
	deny, revoke           bool
	change                 func()
}

func (w *syntheticObjectReadWitness) ValidateObjectReadOperation(ctx context.Context, _ room.State, r workerJobRecord, _ workerObjectReadIntent, _ workerProbeEvent) error {
	w.operations++
	if w.deny || w.revoke && w.operations > 1 {
		return errControllerObjectRead
	}
	if w.change != nil && w.operations == 1 {
		w.change()
	}
	r.Pair.Workspace.Removed = true
	return ctx.Err()
}
func (w *syntheticObjectReadWitness) ValidateObjectReadLifecycle(ctx context.Context, _ room.State, r workerJobRecord, local workerObjectReadRecord, _ workerProbeEvent) error {
	w.lifecycles++
	if w.deny || w.revoke && w.lifecycles > 1 {
		return errControllerObjectRead
	}
	if w.change != nil && w.lifecycles == 1 {
		w.change()
	}
	r.Pair.Workspace.Removed = true
	local.Events[0].Stage = "mutated"
	return ctx.Err()
}

type physicalObjectReadFixture struct {
	probe                       *assembledProbeFixture
	intent                      workerObjectReadIntent
	phase, mode                 string
	calls, inspections, effects int
}

func (f *physicalObjectReadFixture) Run(ctx context.Context, args []string, in []byte, out io.Writer) error {
	f.probe.checkLock()
	f.calls++
	if ctx.Err() != nil || len(in) != 0 {
		return errControllerObjectRead
	}
	p, i := f.probe, f.intent
	switch args[0] {
	case "inspect":
		if args[len(args)-1] != p.id {
			return errControllerObjectRead
		}
		f.inspections++
		c := p.inspection()
		c["Name"] = "/commonroom-read-" + p.record.JobID + "-" + i.Request.Challenge
		c["Image"] = i.Image
		c["State"].(map[string]any)["Status"] = f.phase
		c["State"].(map[string]any)["Running"] = f.phase == "running"
		config := c["Config"].(map[string]any)
		raw, _ := json.Marshal(i.Request)
		sum := sha256.Sum256(raw)
		config["Image"] = i.Image
		config["WorkingDir"] = "/objects"
		config["OpenStdin"] = true
		config["Entrypoint"] = []string{"/opt/commonroom/object-read"}
		config["Cmd"] = []string{base64.RawURLEncoding.EncodeToString(raw)}
		config["Labels"] = map[string]string{"commonroom.managed": "true", "commonroom.execution": "object-read-v1", "commonroom.job": p.record.JobID, "commonroom.repository": p.record.Pair.Repository, "commonroom.read": hex.EncodeToString(sum[:])}
		c["Mounts"] = []map[string]any{{"Type": "volume", "Name": i.Objects.Name, "Source": syntheticObservationDockerRoot + "/volumes/" + i.Objects.Name + "/_data", "Destination": "/objects", "Driver": "local", "RW": f.mode == "writable", "Propagation": ""}}
		if f.mode == "foreign ID" {
			c["Id"] = strings.Repeat("9", 64)
		}
		if f.mode == "record changed" {
			r := p.record
			r.UpdatedAt = r.UpdatedAt.Add(time.Second)
			if err := writeWorkerRecord(p.dir, r, false); err != nil {
				p.t.Fatal(err)
			}
		}
		return json.NewEncoder(out).Encode(c)
	case "ps":
		answer := ""
		if f.phase != "absent" && f.phase != "vacant" {
			answer = p.id
		}
		if f.mode == "extra reference" {
			answer += "\n" + strings.Repeat("9", 64)
		}
		_, err := io.WriteString(out, answer)
		return err
	case "volume":
		if args[1] == "ls" {
			name := i.Objects.Name
			if f.mode == "extra catalog" {
				name += "\nother"
			}
			_, err := io.WriteString(out, name)
			return err
		}
		if args[1] != "inspect" || args[len(args)-1] != i.Objects.Name {
			return errControllerObjectRead
		}
		created := i.Objects.CreatedAt
		if f.mode == "volume replaced" {
			created = "2000-01-01T00:00:00Z"
		}
		return json.NewEncoder(out).Encode(map[string]any{"Name": i.Objects.Name, "Driver": "local", "Scope": "local", "CreatedAt": created, "Mountpoint": syntheticObservationDockerRoot + "/volumes/" + i.Objects.Name + "/_data", "Labels": map[string]string{"commonroom.managed": "true", "commonroom.execution": "object-custody-v1", "commonroom.job": p.record.JobID, "commonroom.repository": p.record.Pair.Repository, "commonroom.custody.nonce": i.Objects.Nonce}})
	}
	f.effects++
	p.t.Error("observation attempted mutation", args)
	return errControllerObjectRead
}
func objectReadPhysicalState(t *testing.T, j *workerObjectReadJournal, p *assembledProbeFixture) room.State {
	t.Helper()
	s, _, ctx := physicalControllerFixture(t, p)
	st := probeControllerState(t, s, ctx)
	job := st.ManagedJobs[j.record.JobID]
	b := j.intent.Request.Receipt.Intent.Scope
	job.TaskID = b.TaskID
	job.TaskDispatchID = b.ExecutionID
	job.TaskDefinitionDigest = b.DefinitionDigest
	st.ManagedJobs[job.JobID] = job
	return st
}
func TestWorkspaceObjectReadEvidenceOperation(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"begin-pre", "begin-post", "bound-pre", "bound-post", "finish-pre", "finish-post", "recovery-pre", "recovery-post", "denied", "revoked", "writable", "false absence", "extra reference", "extra catalog", "volume replaced", "record changed", "history changed", "foreign record", "foreign worker", "expired", "stopped", "old event", "unknown adoption"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			j, g, l, _ := objectReadJournalFixture(t)
			id := strings.Repeat("1", 64)
			event := workerProbeEvent{Stage: "bound", ID: id}
			phase := "created"
			if mode != "begin-pre" {
				beginObjectReadJournal(t, j)
			}
			if strings.HasPrefix(mode, "begin-") {
				event = workerProbeEvent{Stage: "begin"}
				phase = "vacant"
			}
			if mode == "bound-post" || strings.HasPrefix(mode, "finish-") || strings.HasPrefix(mode, "recovery-") || mode == "false absence" || mode == "old event" {
				if j.BindObjectRead(t.Context(), j.intent.Request, id) != nil {
					t.Fatal("bind")
				}
			}
			if strings.HasPrefix(mode, "finish-") || mode == "false absence" {
				event = workerProbeEvent{Stage: "finish", ID: id, Absent: true}
				phase = "absent"
			}
			if mode == "finish-post" || mode == "old event" {
				if j.FinishObjectRead(t.Context(), j.intent.Request, id, true) != nil {
					t.Fatal("finish")
				}
			}
			if strings.HasPrefix(mode, "recovery-") {
				if j.FinishObjectRead(t.Context(), j.intent.Request, id, false) != nil {
					t.Fatal("uncertain")
				}
				event = workerProbeEvent{Stage: "recovery", ID: id}
				if mode == "recovery-post" && j.BeginObjectReadRecovery(t.Context(), j.intent.Request, j.intent.Objects, j.intent.Image, id) != nil {
					t.Fatal("recovery")
				}
			}
			if mode == "unknown adoption" {
				if j.FinishObjectRead(t.Context(), j.intent.Request, "", false) != nil {
					t.Fatal("unknown")
				}
				event = workerProbeEvent{Stage: "recovery", ID: id}
			}
			f := &physicalObjectReadFixture{probe: g.f, intent: j.intent, phase: phase, mode: mode}
			if mode == "false absence" {
				f.phase = "created"
			}
			l.runner = f
			st := objectReadPhysicalState(t, j, g.f)
			w := &syntheticObjectReadWitness{deny: mode == "denied", revoke: mode == "revoked"}
			e, err := newWorkspaceObjectReadEvidence(l, "worker", j.intent, w)
			if err != nil {
				t.Fatal(err)
			}
			r := detachedProbeWorkerRecord(j.record)
			switch mode {
			case "foreign record":
				r.Pair.Workspace.Removed = true
			case "foreign worker":
				e.worker = "other"
			case "expired":
				dead, cancel := context.WithCancel(t.Context())
				cancel()
				e.lifetime = dead
			case "stopped":
				if err = os.WriteFile(filepath.Join(j.dir, workerProbeStop), []byte("stop"), 0600); err != nil {
					t.Fatal(err)
				}
			case "history changed":
				w.change = func() {
					records, err := readWorkerObjectReadRecords(j.dir)
					if err != nil {
						t.Fatal(err)
					}
					local := records[objectReadRecordName(j.intent)]
					local.Events = append(local.Events, event)
					raw, _ := json.Marshal(local)
					if err = os.WriteFile(filepath.Join(j.dir, objectReadRecordName(j.intent)), raw, 0600); err != nil {
						t.Fatal(err)
					}
				}
			}
			err = e.ValidateObjectReadOperation(t.Context(), st, r, j.intent, event)
			healthy := strings.HasSuffix(mode, "-pre") || strings.HasSuffix(mode, "-post")
			if (err == nil) != healthy {
				t.Fatal(mode, err)
			}
			if healthy && w.operations != 2 {
				t.Fatal("witness not rechecked")
			}
			if mode == "denied" && f.calls != 0 {
				t.Fatal("daemon before permission")
			}
			if f.effects != 0 || j.record.Pair.Workspace.Removed {
				t.Fatal("effects or callback escaped")
			}
		})
	}
}
func TestWorkspaceObjectReadEvidenceLifecycle(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"begin", "bound", "finish", "recovery", "recovery-finish", "false absence", "unknown finish", "denied", "revoked", "extra reference", "volume replaced", "history mismatch", "history changed", "corrupt history", "expired"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			j, g, l, _ := objectReadJournalFixture(t)
			beginObjectReadJournal(t, j)
			id := strings.Repeat("1", 64)
			phase := "vacant"
			if mode != "begin" && mode != "denied" && mode != "unknown finish" {
				if j.BindObjectRead(t.Context(), j.intent.Request, id) != nil {
					t.Fatal("bind")
				}
				phase = "created"
			}
			if mode == "finish" || mode == "false absence" {
				if j.FinishObjectRead(t.Context(), j.intent.Request, id, true) != nil {
					t.Fatal("finish")
				}
				phase = "absent"
			}
			if mode == "recovery" || mode == "recovery-finish" {
				if j.FinishObjectRead(t.Context(), j.intent.Request, id, false) != nil || j.BeginObjectReadRecovery(t.Context(), j.intent.Request, j.intent.Objects, j.intent.Image, id) != nil {
					t.Fatal("recovery")
				}
			}
			if mode == "recovery-finish" {
				if j.FinishObjectReadRecovery(t.Context(), j.intent.Request, id, true) != nil {
					t.Fatal("recovered")
				}
				phase = "absent"
			}
			if mode == "unknown finish" {
				if j.FinishObjectRead(t.Context(), j.intent.Request, "", false) != nil {
					t.Fatal("unknown finish")
				}
			}
			if mode == "false absence" {
				phase = "created"
			}
			f := &physicalObjectReadFixture{probe: g.f, intent: j.intent, phase: phase, mode: mode}
			l.runner = f
			st := objectReadPhysicalState(t, j, g.f)
			records, err := readWorkerObjectReadRecords(j.dir)
			if err != nil {
				t.Fatal(err)
			}
			events := objectReadControllerEvents(records[objectReadRecordName(j.intent)])
			candidate := room.ObjectReadReservation{Version: 1, CompanyID: st.Company.ID, RoomID: j.record.RoomID, Worker: "worker", ActorID: ownerID, Intent: j.intent, Events: events, Released: releasedProbeEvents(events)}
			w := &syntheticObjectReadWitness{deny: mode == "denied", revoke: mode == "revoked"}
			e, err := newWorkspaceObjectReadEvidence(l, "worker", j.intent, w)
			if err != nil {
				t.Fatal(err)
			}
			switch mode {
			case "expired":
				dead, cancel := context.WithCancel(t.Context())
				cancel()
				e.lifetime = dead
			case "history mismatch":
				candidate.Events[len(events)-1].ID = strings.Repeat("9", 64)
			case "corrupt history":
				if err = os.WriteFile(filepath.Join(j.dir, objectReadRecordName(j.intent)), []byte(`{}`), 0600); err != nil {
					t.Fatal(err)
				}
			case "history changed":
				w.change = func() {
					local := records[objectReadRecordName(j.intent)]
					local.Events = append(local.Events, workerProbeEvent{Stage: "finish", ID: id})
					raw, _ := json.Marshal(local)
					if err = os.WriteFile(filepath.Join(j.dir, objectReadRecordName(j.intent)), raw, 0600); err != nil {
						t.Fatal(err)
					}
				}
			}
			event := candidate.Events[len(candidate.Events)-1]
			err = e.ValidateObjectReadTransition(t.Context(), st, candidate, event)
			healthy := mode == "begin" || mode == "bound" || mode == "finish" || mode == "recovery" || mode == "recovery-finish" || mode == "unknown finish"
			if (err == nil) != healthy {
				t.Fatal(mode, err)
			}
			if healthy && w.lifecycles != 2 {
				t.Fatal("witness not rechecked")
			}
			if mode == "denied" && f.calls != 0 {
				t.Fatal("inspection before permission")
			}
			if f.effects != 0 {
				t.Fatal("lifecycle observer mutated daemon")
			}
		})
	}
}
func TestWorkspaceObjectReadEvidenceRequiresDependencies(t *testing.T) {
	var e *workspaceObjectReadEvidence
	if e.ValidateObjectReadOperation(t.Context(), room.State{}, workerJobRecord{}, workerObjectReadIntent{}, workerProbeEvent{}) == nil || e.ValidateObjectReadTransition(t.Context(), room.State{}, room.ObjectReadReservation{}, workerProbeEvent{}) == nil {
		t.Fatal("nil evidence accepted")
	}
	j, _, l, _ := objectReadJournalFixture(t)
	if _, err := newWorkspaceObjectReadEvidence(l, "worker", j.intent, nil); err == nil {
		t.Fatal("nil witness accepted")
	}
	if _, err := newWorkspaceObjectReadEvidence(nil, "worker", j.intent, &syntheticObjectReadWitness{}); err == nil {
		t.Fatal("nil owner accepted")
	}
}
