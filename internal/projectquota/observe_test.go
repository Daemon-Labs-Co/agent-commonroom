package projectquota

import (
	"context"
	"encoding/binary"
	"math"
	"testing"
)

type syntheticQuotaReader struct {
	root          rootObservation
	status, quota []byte
	calls         []string
	change        func(string, int)
	closes        int
	fail          string
}

func (r *syntheticQuotaReader) step(name string) error {
	r.calls = append(r.calls, name)
	if r.change != nil {
		r.change(name, len(r.calls))
	}
	if r.fail == name {
		return ErrUnavailable
	}
	return nil
}
func (r *syntheticQuotaReader) Root(context.Context) (rootObservation, error) {
	err := r.step("root")
	return r.root, err
}
func (r *syntheticQuotaReader) Status(context.Context) ([]byte, error) {
	err := r.step("status")
	return append([]byte(nil), r.status...), err
}
func (r *syntheticQuotaReader) Quota(_ context.Context, project uint32) ([]byte, error) {
	if project != 7 {
		return nil, ErrUnavailable
	}
	err := r.step("quota")
	return append([]byte(nil), r.quota...), err
}
func (r *syntheticQuotaReader) Close() error {
	r.closes++
	if r.fail == "close" {
		return ErrUnavailable
	}
	return nil
}
func quotaFixture() (Expected, *syntheticQuotaReader) {
	e := Expected{Device: 1, Inode: 2, Project: 7, MaxBytes: 5_000_000_000, MaxInodes: 1000}
	r := &syntheticQuotaReader{root: rootObservation{1, 2, xfsMagic, 7, projectInherit}, status: make([]byte, statusBytes), quota: make([]byte, quotaBytes)}
	r.status[0] = 1
	binary.LittleEndian.PutUint16(r.status[2:], projectQuotaFlags)
	binary.LittleEndian.PutUint64(r.status[56:], 12)
	r.quota[0] = 1
	r.quota[1] = 2
	binary.LittleEndian.PutUint32(r.quota[4:], 7)
	binary.LittleEndian.PutUint64(r.quota[8:], 9_765_625)
	binary.LittleEndian.PutUint64(r.quota[24:], 1000)
	binary.LittleEndian.PutUint64(r.quota[40:], 10)
	binary.LittleEndian.PutUint64(r.quota[48:], 5)
	return e, r
}
func TestObserveEnforcedProjectQuota(t *testing.T) {
	t.Parallel()
	e, r := quotaFixture()
	obs, err := inspectOwned(context.Background(), r, e)
	if err != nil || obs.HardBytes != 5_000_000_000 || obs.UsedBytes != 5120 || obs.HardInodes != 1000 || obs.UsedInodes != 5 || r.closes != 1 {
		t.Fatal(obs, err, r.closes)
	}
	want := []string{"root", "status", "quota", "quota", "status", "root"}
	if len(r.calls) != len(want) {
		t.Fatal(r.calls)
	}
	for i := range want {
		if r.calls[i] != want[i] {
			t.Fatal(r.calls)
		}
	}
}
func TestObserveQuotaRefusals(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"wrong root", "wrong filesystem", "wrong project", "no inherit", "realtime inherit", "no enforcement", "no accounting", "status version", "missing quota inode", "quota version", "quota type", "quota ID", "unlimited bytes", "excess bytes", "overflow", "excess usage", "unlimited inodes", "excess inodes", "inode usage", "realtime quota", "short status", "short quota", "root error", "status error", "quota error", "close error"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()
			e, r := quotaFixture()
			switch kind {
			case "wrong root":
				r.root.Inode++
			case "wrong filesystem":
				r.root.Type = 0xef53
			case "wrong project":
				r.root.Project++
			case "no inherit":
				r.root.Flags = 0
			case "realtime inherit":
				r.root.Flags |= 0x100
			case "no enforcement":
				binary.LittleEndian.PutUint16(r.status[2:], 1<<4)
			case "no accounting":
				binary.LittleEndian.PutUint16(r.status[2:], 1<<5)
			case "status version":
				r.status[0] = 2
			case "missing quota inode":
				binary.LittleEndian.PutUint64(r.status[56:], 0)
			case "quota version":
				r.quota[0] = 0
			case "quota type":
				r.quota[1] = 1
			case "quota ID":
				binary.LittleEndian.PutUint32(r.quota[4:], 8)
			case "unlimited bytes":
				binary.LittleEndian.PutUint64(r.quota[8:], 0)
			case "excess bytes":
				binary.LittleEndian.PutUint64(r.quota[8:], 9_765_626)
			case "overflow":
				binary.LittleEndian.PutUint64(r.quota[8:], math.MaxUint64)
			case "excess usage":
				binary.LittleEndian.PutUint64(r.quota[40:], 9_765_626)
			case "unlimited inodes":
				binary.LittleEndian.PutUint64(r.quota[24:], 0)
			case "excess inodes":
				binary.LittleEndian.PutUint64(r.quota[24:], 1001)
			case "inode usage":
				binary.LittleEndian.PutUint64(r.quota[48:], 1001)
			case "realtime quota":
				binary.LittleEndian.PutUint64(r.quota[72:], 1)
			case "short status":
				r.status = r.status[:10]
			case "short quota":
				r.quota = r.quota[:10]
			case "root error":
				r.fail = "root"
			case "status error":
				r.fail = "status"
			case "quota error":
				r.fail = "quota"
			case "close error":
				r.fail = "close"
			}
			obs, err := inspectOwned(context.Background(), r, e)
			if err != ErrUnavailable || obs != (Observation{}) || r.closes != 1 {
				t.Fatal("unsafe evidence accepted", obs, err, r.closes)
			}
		})
	}
}
func TestObserveQuotaChangedOrCanceled(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"root", "quota", "enforcement", "quota inode", "cancel"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()
			e, r := quotaFixture()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			r.change = func(_ string, n int) {
				switch {
				case kind == "quota" && n == 4:
					r.quota[40]++
				case kind == "enforcement" && n == 5:
					r.status[2] = 0
				case kind == "quota inode" && n == 5:
					r.status[56]++
				case kind == "root" && n == 6:
					r.root.Project++
				case kind == "cancel" && n == 6:
					cancel()
				}
			}
			obs, err := inspectOwned(ctx, r, e)
			if err != ErrUnavailable || obs != (Observation{}) || r.closes != 1 {
				t.Fatal(obs, err)
			}
		})
	}
}
func TestInvalidExpectedRefusesBeforeKernelReads(t *testing.T) {
	t.Parallel()
	e, r := quotaFixture()
	e.Project = 0
	if _, err := inspectOwned(context.Background(), r, e); err != ErrUnavailable || len(r.calls) != 0 || r.closes != 1 {
		t.Fatal(err, r.calls)
	}
}
