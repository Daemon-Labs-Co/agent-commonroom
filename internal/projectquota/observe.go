// Package projectquota observes pre-provisioned XFS project quota enforcement.
// It never enables/changes quotas, writes files, selects paths or authorizes model
// execution. A source fork replacing the kernel reader invalidates this evidence.
package projectquota

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"math"
	"os"
)

var ErrUnavailable = errors.New("physical project quota evidence unavailable")

const (
	xfsMagic          int64  = 0x58465342
	projectInherit    uint32 = 0x200
	projectQuotaFlags uint16 = (1 << 4) | (1 << 5)
	statusBytes              = 160
	quotaBytes               = 112
)

// Expected comes from independently approved volume/custody provisioning, never
// model metadata or the first observed root. Nonzero hard bounds are reservations,
// not permission to consume unaccounted space. Project zero is never accepted.
type Expected struct {
	Device, Inode       uint64
	Project             uint32
	MaxBytes, MaxInodes uint64
}
type Observation struct {
	Device, Inode                                uint64
	Project                                      uint32
	HardBytes, UsedBytes, HardInodes, UsedInodes uint64
}
type rootObservation struct {
	Device, Inode  uint64
	Type           int64
	Project, Flags uint32
}
type kernelReader interface {
	Root(context.Context) (rootObservation, error)
	Status(context.Context) ([]byte, error)
	Quota(context.Context, uint32) ([]byte, error)
	Close() error
}

func validExpected(e Expected) bool {
	return e.Device != 0 && e.Inode != 0 && e.Project != 0 && e.MaxBytes >= 512 && e.MaxInodes > 0
}

// Inspect pins a duplicate of an already opened root directory descriptor for all
// observations. Callers must keep actual creator/writer fences and independently
// prove complete inventory, volume mapping and rootless containment. This is a
// point-in-time read-only check, not a physical quota provisioner or release guard.
// Permission/unsupported/cancellation/uncertainty refuse without fallback.
func Inspect(ctx context.Context, root *os.File, e Expected) (Observation, error) {
	if ctx == nil || ctx.Err() != nil || root == nil || !validExpected(e) {
		return Observation{}, ErrUnavailable
	}
	reader, err := newKernelReader(root)
	if err != nil {
		return Observation{}, ErrUnavailable
	}
	return inspectOwned(ctx, reader, e)
}
func inspectOwned(ctx context.Context, reader kernelReader, e Expected) (observation Observation, err error) {
	defer func() {
		if reader.Close() != nil {
			observation = Observation{}
			err = ErrUnavailable
		}
	}()
	return inspect(ctx, reader, e)
}
func inspect(ctx context.Context, r kernelReader, e Expected) (Observation, error) {
	fail := func() (Observation, error) { return Observation{}, ErrUnavailable }
	if ctx == nil || ctx.Err() != nil || r == nil || !validExpected(e) {
		return fail()
	}
	before, err := r.Root(ctx)
	if err != nil || !validRoot(before, e) {
		return fail()
	}
	status, err := r.Status(ctx)
	if err != nil || !validStatus(status) {
		return fail()
	}
	quota, err := r.Quota(ctx, e.Project)
	if err != nil {
		return fail()
	}
	observation, err := decodeQuota(quota, e)
	if err != nil {
		return fail()
	}
	// Repeated actual queries reject changed limits/usage and lost enforcement.
	again, err := r.Quota(ctx, e.Project)
	if err != nil || !bytes.Equal(quota, again) {
		return fail()
	}
	statusAfter, err := r.Status(ctx)
	if err != nil || !validStatus(statusAfter) || !bytes.Equal(status[56:64], statusAfter[56:64]) {
		return fail()
	}
	after, err := r.Root(ctx)
	if err != nil || after != before || !validRoot(after, e) || ctx.Err() != nil {
		return fail()
	}
	return observation, nil
}
func validRoot(r rootObservation, e Expected) bool {
	return r.Device == e.Device && r.Inode == e.Inode && r.Type == xfsMagic && r.Project == e.Project && r.Flags&projectInherit != 0 && r.Flags&0x101 == 0
}
func validStatus(raw []byte) bool {
	return len(raw) == statusBytes && raw[0] == 1 && binary.LittleEndian.Uint16(raw[2:4])&projectQuotaFlags == projectQuotaFlags && binary.LittleEndian.Uint64(raw[56:64]) != 0
}
func decodeQuota(raw []byte, e Expected) (Observation, error) {
	fail := func() (Observation, error) { return Observation{}, ErrUnavailable }
	if len(raw) != quotaBytes || raw[0] != 1 || raw[1] != 2 || binary.LittleEndian.Uint32(raw[4:8]) != e.Project {
		return fail()
	}
	hard := binary.LittleEndian.Uint64(raw[8:16])
	used := binary.LittleEndian.Uint64(raw[40:48])
	inodes := binary.LittleEndian.Uint64(raw[24:32])
	usedInodes := binary.LittleEndian.Uint64(raw[48:56])
	// XFS quota block counters use 512-byte units, not VFS's 1024-byte limits.
	if hard == 0 || hard > math.MaxUint64/512 || used > math.MaxUint64/512 || hard*512 > e.MaxBytes || used > hard || inodes == 0 || inodes > e.MaxInodes || usedInodes > inodes {
		return fail()
	}
	// Realtime allocation has separate limits and cannot silently escape this bound.
	for _, offset := range []int{72, 80, 88} {
		if binary.LittleEndian.Uint64(raw[offset:offset+8]) != 0 {
			return fail()
		}
	}
	return Observation{e.Device, e.Inode, e.Project, hard * 512, used * 512, inodes, usedInodes}, nil
}
