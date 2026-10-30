package projectquota

import (
	"context"
	"encoding/binary"
	"math"
	"os"
	"testing"
)

func quotaSetFixture(t *testing.T) ([]Root, []*syntheticQuotaReader) {
	t.Helper()
	roots := make([]Root, 2)
	readers := make([]*syntheticQuotaReader, 2)
	for i := range roots {
		dir, err := os.Open(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = dir.Close() })
		e, r := quotaFixture()
		e.Device = uint64(i + 1)
		r.root.Device = e.Device
		e.MaxBytes = 4096
		binary.LittleEndian.PutUint64(r.quota[8:], 8)
		binary.LittleEndian.PutUint64(r.quota[40:], 1)
		roots[i] = Root{dir, e}
		readers[i] = r
	}
	return roots, readers
}
func TestInspectSetPinsAllRootsAndBoundsReservedCapacity(t *testing.T) {
	t.Parallel()
	roots, readers := quotaSetFixture(t)
	opens := 0
	for _, r := range readers {
		r.change = func(_ string, _ int) {
			if opens != 2 || readers[0].closes != 0 || readers[1].closes != 0 {
				t.Fatal("set inspected without complete descriptor ownership")
			}
		}
	}
	obs, err := inspectSet(t.Context(), roots, 8192, func(f *os.File) (kernelReader, error) {
		if f != roots[opens].Directory {
			t.Fatal("changed root")
		}
		r := readers[opens]
		opens++
		return r, nil
	})
	if err != nil || len(obs) != 2 || obs[0].HardBytes != 4096 || obs[1].Device != 2 {
		t.Fatal(obs, err)
	}
	for i, r := range readers {
		if r.closes != 1 || len(r.calls) != 12 {
			t.Fatal("missing joined repeated observations", r.closes, r.calls)
		}
		if _, err := roots[i].Directory.Stat(); err != nil {
			t.Fatal("caller root closed")
		}
	}
}
func TestInspectSetRefusesInvalidCompletePlanBeforeOpen(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"empty", "over budget", "overflow", "duplicate root", "duplicate project", "nil directory", "invalid bound", "too many"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			roots, _ := quotaSetFixture(t)
			budget := uint64(8192)
			switch mode {
			case "empty":
				roots = nil
			case "over budget":
				budget--
			case "overflow":
				roots[0].Expected.MaxBytes = math.MaxUint64
				budget = math.MaxUint64
			case "duplicate root":
				roots[1].Expected.Device = 1
				roots[1].Expected.Inode = roots[0].Expected.Inode
			case "duplicate project":
				roots[1].Expected.Device = 1
				roots[1].Expected.Inode++
			case "nil directory":
				roots[1].Directory = nil
			case "invalid bound":
				roots[1].Expected.MaxInodes = 0
			case "too many":
				for len(roots) < 33 {
					roots = append(roots, roots[0])
				}
			}
			obs, err := inspectSet(t.Context(), roots, budget, func(*os.File) (kernelReader, error) { t.Fatal("invalid whole plan opened"); return nil, nil })
			if err == nil || obs != nil {
				t.Fatal("invalid plan accepted")
			}
		})
	}
}
func TestInspectSetFailureWithholdsPrefixAndJoinsReaders(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"second open", "late open error", "first inspect", "second inspect", "changed usage", "changed limit", "lost enforcement", "close error", "cancel during open", "cancel during inspect"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			roots, readers := quotaSetFixture(t)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			opens := 0
			switch mode {
			case "first inspect":
				readers[0].fail = "root"
			case "second inspect":
				readers[1].fail = "quota"
			case "close error":
				readers[1].fail = "close"
			case "changed usage", "changed limit", "lost enforcement":
				readers[0].change = func(_ string, n int) {
					if n == 7 {
						switch mode {
						case "changed usage":
							binary.LittleEndian.PutUint64(readers[0].quota[40:], 2)
						case "changed limit":
							binary.LittleEndian.PutUint64(readers[0].quota[8:], 7)
						case "lost enforcement":
							binary.LittleEndian.PutUint16(readers[0].status[2:], 0)
						}
					}
				}
			case "cancel during inspect":
				readers[1].change = func(_ string, _ int) { cancel() }
			}
			obs, err := inspectSet(ctx, roots, 8192, func(*os.File) (kernelReader, error) {
				i := opens
				opens++
				if i == 1 && mode == "second open" {
					return nil, ErrUnavailable
				}
				if i == 1 && mode == "late open error" {
					return readers[i], ErrUnavailable
				}
				if mode == "cancel during open" {
					cancel()
				}
				return readers[i], nil
			})
			if err == nil || obs != nil {
				t.Fatal("uncertain set returned evidence", obs, err)
			}
			for i, r := range readers {
				want := 0
				if i < opens && !(i == 1 && mode == "second open") {
					want = 1
				}
				if r.closes != want {
					t.Fatal("unjoined reader", i, r.closes, want)
				}
			}
		})
	}
}

func TestInspectSetOwnsPolicySnapshotAndRefusesCanceledAdmission(t *testing.T) {
	t.Parallel()
	roots, readers := quotaSetFixture(t)
	opens := 0
	observations, err := inspectSet(t.Context(), roots, 8192, func(*os.File) (kernelReader, error) {
		i := opens
		opens++
		// A callback cannot retarget later quota expectations after plan validation.
		roots[1].Expected.Device = 99
		roots[1].Expected.MaxBytes = math.MaxUint64
		return readers[i], nil
	})
	if err != nil || len(observations) != 2 || observations[1].Device != 2 {
		t.Fatal("borrowed mutable policy", observations, err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if observations, err = inspectSet(ctx, roots, 8192, func(*os.File) (kernelReader, error) { t.Fatal("canceled set acquired a root"); return nil, nil }); err == nil || observations != nil {
		t.Fatal("canceled set accepted")
	}
}
