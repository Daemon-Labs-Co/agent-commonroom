package projectquota

import (
	"context"
	"os"
)

// Root binds an already opened directory to an independently approved reserved
// capacity. Neither the first observed identity nor model metadata is policy.
type Root struct {
	Directory *os.File
	Expected  Expected
}

// InspectSet checks a bounded set of projects against a finite aggregate reserved
// budget. All root descriptors remain pinned until the entire observation joins.
// Callers MUST independently fence all writers and prove inventory completeness:
// matching repeated observations alone do not make an atomic snapshot, establish
// account ownership or authorize allocation/export. No filesystem is modified.
func InspectSet(ctx context.Context, roots []Root, budget uint64) ([]Observation, error) {
	return inspectSet(ctx, roots, budget, newKernelReader)
}
func inspectSet(ctx context.Context, roots []Root, budget uint64, open func(*os.File) (kernelReader, error)) (observations []Observation, err error) {
	refuse := func() ([]Observation, error) { return nil, ErrUnavailable }
	if ctx == nil || ctx.Err() != nil || open == nil || len(roots) == 0 || len(roots) > 32 || budget < 512 {
		return refuse()
	}
	roots = append([]Root(nil), roots...) // own policy values across callbacks
	identities := map[[2]uint64]bool{}
	projects := map[[2]uint64]bool{}
	var reserved uint64
	// Validate all identities/bounds before opening any descriptor. Reservations,
	// including unused capacity, must fit; smaller observed usage cannot erase them.
	for _, root := range roots {
		e := root.Expected
		identity := [2]uint64{e.Device, e.Inode}
		project := [2]uint64{e.Device, uint64(e.Project)}
		if root.Directory == nil || !validExpected(e) || identities[identity] || projects[project] || e.MaxBytes > budget-reserved {
			return refuse()
		}
		identities[identity] = true
		projects[project] = true
		reserved += e.MaxBytes
	}
	readers := make([]kernelReader, 0, len(roots))
	defer func() {
		for i := len(readers) - 1; i >= 0; i-- {
			if readers[i].Close() != nil {
				observations = nil
				err = ErrUnavailable
			}
		}
	}()
	for _, root := range roots {
		if ctx.Err() != nil {
			return refuse()
		}
		reader, openErr := open(root.Directory)
		if reader != nil {
			readers = append(readers, reader)
		}
		if openErr != nil || reader == nil {
			return refuse()
		}
	}
	observations = make([]Observation, len(roots))
	for i, reader := range readers {
		observation, inspectErr := inspect(ctx, reader, roots[i].Expected)
		if inspectErr != nil {
			return refuse()
		}
		observations[i] = observation
	}
	// Recheck in reverse order while every descriptor is still pinned. Any limit,
	// identity or usage change withholds the entire result, never a partial prefix.
	for i := len(readers) - 1; i >= 0; i-- {
		observation, inspectErr := inspect(ctx, readers[i], roots[i].Expected)
		if inspectErr != nil || observation != observations[i] {
			return refuse()
		}
	}
	if ctx.Err() != nil {
		return refuse()
	}
	return observations, nil
}
