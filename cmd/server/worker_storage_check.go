package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"agentroom/internal/projectquota"
	"agentroom/internal/strictjson"
)

// Operator-only, read-only deployment preflight. Independent provisioning supplies
// expected identity and reservation bounds; observed values never become policy.
// This command grants no execution/custody/export authority, has no flag bypass,
// and cannot substitute for fenced live inventory or mandatory container isolation.
type workerStorageCheckPolicy struct {
	Version int                      `json:"version"`
	Roots   []workerStorageCheckRoot `json:"roots"`
}
type workerStorageCheckRoot struct {
	CompanyID string `json:"companyId"`
	Worker    string `json:"worker"`
	Path      string `json:"path"`
	Device    uint64 `json:"device"`
	Inode     uint64 `json:"inode"`
	Project   uint32 `json:"project"`
	MaxBytes  uint64 `json:"maxBytes"`
	MaxInodes uint64 `json:"maxInodes"`
}

func (r workerStorageCheckRoot) expected() projectquota.Expected {
	return projectquota.Expected{Device: r.Device, Inode: r.Inode, Project: r.Project, MaxBytes: r.MaxBytes, MaxInodes: r.MaxInodes}
}
func loadWorkerStorageCheckPolicy(path string) (workerStorageCheckPolicy, error) {
	refuse := func() (workerStorageCheckPolicy, error) {
		return workerStorageCheckPolicy{}, projectquota.ErrUnavailable
	}
	if !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return refuse()
	}
	raw, err := readPrivateBounded(path, 64*1024)
	if err != nil || !strictjson.ValidEncoding(raw) || !uniqueObjectReadProfileJSON(raw) {
		return refuse()
	}
	var policy workerStorageCheckPolicy
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if dec.Decode(&policy) != nil || policy.Version != 1 || len(policy.Roots) == 0 || len(policy.Roots) > 32 {
		return refuse()
	}
	var trailing any
	if dec.Decode(&trailing) != io.EOF {
		return refuse()
	}
	paths := map[string]bool{}
	identities := map[[2]uint64]bool{}
	projects := map[[2]uint64]bool{}
	for _, r := range policy.Roots {
		identity := [2]uint64{r.Device, r.Inode}
		project := [2]uint64{r.Device, uint64(r.Project)}
		if !membershipPartPattern.MatchString(r.CompanyID) || !workerAliasPattern.MatchString(r.Worker) || !filepath.IsAbs(r.Path) || filepath.Clean(r.Path) != r.Path || r.Path == string(filepath.Separator) || r.Device == 0 || r.Inode == 0 || r.Project == 0 || r.MaxBytes < 512 || r.MaxInodes == 0 || paths[r.Path] || identities[identity] || projects[project] {
			return refuse()
		}
		paths[r.Path] = true
		identities[identity] = true
		projects[project] = true
	}
	return policy, nil
}

// Parent directories are operator-controlled; never accept paths from a model or
// request. The leaf cannot be a symlink/non-directory, and Inspect independently
// checks the opened descriptor's provisioned device/inode/project. No path retry.
func openWorkerStorageCheckRoot(path string) (*os.File, error) {
	fd, err := syscall.Open(path, syscall.O_RDONLY|syscall.O_DIRECTORY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK|syscall.O_CLOEXEC, 0)
	if err != nil {
		return nil, projectquota.ErrUnavailable
	}
	return os.NewFile(uintptr(fd), "worker-storage-root"), nil
}

type workerStorageSetInspector func(context.Context, []projectquota.Root, uint64) ([]projectquota.Observation, error)

// The complete policy validates before any root opens. Keep all caller descriptors
// pinned until every company group has been inspected; close uncertainty withholds
// the whole result. Operator labels/budgets still do not prove inventory ownership.
func checkWorkerStoragePolicy(ctx context.Context, policy workerStorageCheckPolicy, inspect workerStorageSetInspector) (err error) {
	type group struct {
		roots    []projectquota.Root
		reserved uint64
	}
	var groups []group
	indices := map[string]int{}
	// Check full reserved capacity, not observed low usage; use the free-edition
	// ceiling independently for each company, including roots on different workers.
	for _, root := range policy.Roots {
		index, ok := indices[root.CompanyID]
		if !ok {
			index = len(groups)
			indices[root.CompanyID] = index
			groups = append(groups, group{})
		}
		g := &groups[index]
		if root.MaxBytes > uint64(managedStorageLimit)-g.reserved {
			return projectquota.ErrUnavailable
		}
		g.reserved += root.MaxBytes
	}
	var files []*os.File
	defer func() {
		for i := len(files) - 1; i >= 0; i-- {
			if files[i].Close() != nil {
				err = projectquota.ErrUnavailable
			}
		}
	}()
	for _, root := range policy.Roots {
		if ctx.Err() != nil {
			return projectquota.ErrUnavailable
		}
		f, openErr := openWorkerStorageCheckRoot(root.Path)
		if openErr != nil {
			return projectquota.ErrUnavailable
		}
		files = append(files, f)
		index := indices[root.CompanyID]
		groups[index].roots = append(groups[index].roots, projectquota.Root{Directory: f, Expected: root.expected()})
	}
	for _, g := range groups {
		observations, inspectErr := inspect(ctx, g.roots, uint64(managedStorageLimit))
		if inspectErr != nil || ctx.Err() != nil || len(observations) != len(g.roots) {
			return projectquota.ErrUnavailable
		}
		// Independently reject mismatched/incomplete evidence at the injected seam.
		for i, o := range observations {
			e := g.roots[i].Expected
			if o.Device != e.Device || o.Inode != e.Inode || o.Project != e.Project || o.HardBytes < 512 || o.HardBytes > e.MaxBytes || o.UsedBytes > o.HardBytes || o.HardInodes == 0 || o.HardInodes > e.MaxInodes || o.UsedInodes > o.HardInodes {
				return projectquota.ErrUnavailable
			}
		}
	}
	if ctx.Err() != nil {
		return projectquota.ErrUnavailable
	}
	return nil
}
func workerStorageCheckMain(args []string, stdout, stderr io.Writer) int {
	return runWorkerStorageCheck(context.Background(), args, stdout, stderr, projectquota.InspectSet)
}
func runWorkerStorageCheck(ctx context.Context, args []string, stdout, stderr io.Writer, inspect workerStorageSetInspector) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "Usage: --check-worker-storage /absolute/private/storage.json")
		return 2
	}
	refuse := func() int {
		fmt.Fprintln(stderr, "Worker storage validation refused; check approved root identity, native XFS enforcement, finite limits and observation permissions. No allocation is authorized.")
		return 1
	}
	if ctx == nil || ctx.Err() != nil || inspect == nil {
		return refuse()
	}
	policy, err := loadWorkerStorageCheckPolicy(args[0])
	if err != nil {
		return refuse()
	}
	if checkWorkerStoragePolicy(ctx, policy, inspect) != nil {
		return refuse()
	}
	if ctx.Err() != nil {
		return refuse()
	}
	_, err = fmt.Fprintf(stdout, "Worker storage observations validated (%d roots). Point-in-time only: complete inventory, container custody, account capacity and live acceptance have not been verified.\n", len(policy.Roots))
	if err != nil {
		return 1
	}
	return 0
}
