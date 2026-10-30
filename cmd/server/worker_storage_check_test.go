package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agentroom/internal/projectquota"
)

func storageCheckFixture(t *testing.T) (string, workerStorageCheckPolicy) {
	t.Helper()
	dir := t.TempDir()
	root := filepath.Join(dir, "root")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	policy := workerStorageCheckPolicy{Version: 1, Roots: []workerStorageCheckRoot{{CompanyID: "company-a", Worker: "worker-a", Path: root, Device: 17, Inode: 23, Project: 7, MaxBytes: 4096, MaxInodes: 16}}}
	return filepath.Join(dir, "policy.json"), policy
}
func writeStorageCheckPolicy(t *testing.T, path string, p workerStorageCheckPolicy) {
	t.Helper()
	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
}
func storageCheckObservation(e projectquota.Expected) projectquota.Observation {
	return projectquota.Observation{Device: e.Device, Inode: e.Inode, Project: e.Project, HardBytes: e.MaxBytes, UsedBytes: 512, HardInodes: e.MaxInodes, UsedInodes: 1}
}
func TestWorkerStorageCheckSuccessAndDescriptorOwnership(t *testing.T) {
	path, policy := storageCheckFixture(t)
	writeStorageCheckPolicy(t, path, policy)
	var out, errout bytes.Buffer
	var opened *os.File
	calls := 0
	inspect := func(ctx context.Context, f *os.File, e projectquota.Expected) (projectquota.Observation, error) {
		calls++
		opened = f
		info, err := f.Stat()
		if err != nil || !info.IsDir() || e != policy.Roots[0].expected() {
			t.Fatal("wrong root/binding")
		}
		return storageCheckObservation(e), nil
	}
	if code := runSingleWorkerStorageCheckForTest(context.Background(), []string{path}, &out, &errout, inspect); code != 0 || calls != 1 || errout.Len() != 0 || !strings.Contains(out.String(), "Point-in-time only") {
		t.Fatalf("code=%d calls=%d out=%q err=%q", code, calls, out.String(), errout.String())
	}
	if _, err := opened.Stat(); !errors.Is(err, os.ErrClosed) {
		t.Fatal("root descriptor remains open")
	}
	if strings.Contains(out.String(), path) {
		t.Fatal("path disclosure")
	}
}
func TestWorkerStorageCheckPolicyRefusalBeforeInspection(t *testing.T) {
	tests := map[string]func(*workerStorageCheckPolicy){
		"version":         func(p *workerStorageCheckPolicy) { p.Version = 2 },
		"empty":           func(p *workerStorageCheckPolicy) { p.Roots = nil },
		"company":         func(p *workerStorageCheckPolicy) { p.Roots[0].CompanyID = "../bad" },
		"worker":          func(p *workerStorageCheckPolicy) { p.Roots[0].Worker = "../bad" },
		"relative":        func(p *workerStorageCheckPolicy) { p.Roots[0].Path = "relative" },
		"unclean":         func(p *workerStorageCheckPolicy) { p.Roots[0].Path += "/../root" },
		"filesystem root": func(p *workerStorageCheckPolicy) { p.Roots[0].Path = "/" },
		"device":          func(p *workerStorageCheckPolicy) { p.Roots[0].Device = 0 },
		"inode":           func(p *workerStorageCheckPolicy) { p.Roots[0].Inode = 0 },
		"project":         func(p *workerStorageCheckPolicy) { p.Roots[0].Project = 0 },
		"bytes":           func(p *workerStorageCheckPolicy) { p.Roots[0].MaxBytes = 511 },
		"inodes":          func(p *workerStorageCheckPolicy) { p.Roots[0].MaxInodes = 0 },
		"duplicate path": func(p *workerStorageCheckPolicy) {
			p.Roots = append(p.Roots, p.Roots[0])
			p.Roots[1].Inode++
			p.Roots[1].Project++
		},
		"duplicate inode": func(p *workerStorageCheckPolicy) {
			p.Roots = append(p.Roots, p.Roots[0])
			p.Roots[1].Path += "2"
			p.Roots[1].Project++
		},
		"shared project": func(p *workerStorageCheckPolicy) {
			p.Roots = append(p.Roots, p.Roots[0])
			p.Roots[1].Path += "2"
			p.Roots[1].Inode++
		},
	}
	for name, change := range tests {
		t.Run(name, func(t *testing.T) {
			path, p := storageCheckFixture(t)
			change(&p)
			writeStorageCheckPolicy(t, path, p)
			var out, errout bytes.Buffer
			if runSingleWorkerStorageCheckForTest(context.Background(), []string{path}, &out, &errout, func(context.Context, *os.File, projectquota.Expected) (projectquota.Observation, error) {
				t.Fatal("invalid policy observed")
				return projectquota.Observation{}, nil
			}) != 1 || out.Len() != 0 {
				t.Fatal("policy accepted")
			}
		})
	}
}
func TestWorkerStorageCheckStrictPrivateJSON(t *testing.T) {
	for _, raw := range []string{`{"version":1,"Version":1,"roots":[]}`, `{"version":1,"roots":[],"extra":true}`, `{"version":1,"roots":[]} {}`, "\xff"} {
		path, _ := storageCheckFixture(t)
		if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := loadWorkerStorageCheckPolicy(path); err == nil {
			t.Fatal("ambiguous JSON accepted")
		}
	}
	path, p := storageCheckFixture(t)
	writeStorageCheckPolicy(t, path, p)
	if err := os.Chmod(path, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadWorkerStorageCheckPolicy(path); err == nil {
		t.Fatal("public policy accepted")
	}
}
func TestWorkerStorageCheckRefusesWrongRootsAndUncertainEvidence(t *testing.T) {
	for _, mode := range []string{"symlink", "file", "missing", "inspect error", "wrong identity", "excess bytes", "excess inodes", "used bytes", "used inodes", "empty", "canceled", "close failure"} {
		t.Run(mode, func(t *testing.T) {
			path, p := storageCheckFixture(t)
			switch mode {
			case "symlink":
				leaf := p.Roots[0].Path + "-link"
				if err := os.Symlink(p.Roots[0].Path, leaf); err != nil {
					t.Fatal(err)
				}
				p.Roots[0].Path = leaf
			case "file":
				leaf := p.Roots[0].Path + "-file"
				if err := os.WriteFile(leaf, nil, 0600); err != nil {
					t.Fatal(err)
				}
				p.Roots[0].Path = leaf
			case "missing":
				p.Roots[0].Path += "-missing"
			}
			writeStorageCheckPolicy(t, path, p)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var out, errout bytes.Buffer
			calls := 0
			var opened *os.File
			inspect := func(_ context.Context, f *os.File, e projectquota.Expected) (projectquota.Observation, error) {
				calls++
				opened = f
				o := storageCheckObservation(e)
				switch mode {
				case "inspect error":
					return o, errors.New("secret canary")
				case "wrong identity":
					o.Inode++
				case "excess bytes":
					o.HardBytes++
				case "excess inodes":
					o.HardInodes++
				case "used bytes":
					o.UsedBytes = o.HardBytes + 1
				case "used inodes":
					o.UsedInodes = o.HardInodes + 1
				case "empty":
					return projectquota.Observation{}, nil
				case "canceled":
					cancel()
				case "close failure":
					if err := f.Close(); err != nil {
						t.Fatal(err)
					}
				}
				return o, nil
			}
			if runSingleWorkerStorageCheckForTest(ctx, []string{path}, &out, &errout, inspect) != 1 || out.Len() != 0 || strings.Contains(errout.String(), "canary") || strings.Contains(errout.String(), path) {
				t.Fatal("uncertain evidence or disclosure")
			}
			if (mode == "symlink" || mode == "file" || mode == "missing") && calls != 0 {
				t.Fatal("invalid root inspected")
			}
			if opened != nil {
				if _, err := opened.Stat(); !errors.Is(err, os.ErrClosed) {
					t.Fatal("leaked descriptor")
				}
			}
		})
	}
}
func TestWorkerStorageCheckAllOrNothingAndUsage(t *testing.T) {
	path, p := storageCheckFixture(t)
	second := p.Roots[0]
	second.Path += "-second"
	second.Inode++
	second.Project++
	if err := os.Mkdir(second.Path, 0700); err != nil {
		t.Fatal(err)
	}
	p.Roots = append(p.Roots, second)
	writeStorageCheckPolicy(t, path, p)
	var out, errout bytes.Buffer
	calls := 0
	inspect := func(_ context.Context, _ *os.File, e projectquota.Expected) (projectquota.Observation, error) {
		calls++
		if calls == 2 {
			return projectquota.Observation{}, projectquota.ErrUnavailable
		}
		return storageCheckObservation(e), nil
	}
	if runSingleWorkerStorageCheckForTest(context.Background(), []string{path}, &out, &errout, inspect) != 1 || calls != 2 || out.Len() != 0 {
		t.Fatal("partial success reported")
	}
	out.Reset()
	errout.Reset()
	if runSingleWorkerStorageCheckForTest(context.Background(), nil, &out, &errout, inspect) != 2 || calls != 2 {
		t.Fatal("usage touched roots")
	}
}

func TestWorkerStorageCheckPreCanceledAndInvalidWholePolicy(t *testing.T) {
	path, p := storageCheckFixture(t)
	invalid := p.Roots[0]
	invalid.Path += "-other"
	invalid.Inode++
	invalid.Project = 0
	p.Roots = append(p.Roots, invalid)
	writeStorageCheckPolicy(t, path, p)
	var out, errout bytes.Buffer
	inspect := func(context.Context, *os.File, projectquota.Expected) (projectquota.Observation, error) {
		t.Fatal("inspection before complete policy validation")
		return projectquota.Observation{}, nil
	}
	if runSingleWorkerStorageCheckForTest(context.Background(), []string{path}, &out, &errout, inspect) != 1 {
		t.Fatal("invalid whole policy accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if runSingleWorkerStorageCheckForTest(ctx, []string{path}, &out, &errout, inspect) != 1 {
		t.Fatal("canceled preflight accepted")
	}
	p.Roots = p.Roots[:1]
	for len(p.Roots) < 33 {
		r := p.Roots[0]
		r.Path += strings.Repeat("x", len(p.Roots))
		r.Inode += uint64(len(p.Roots))
		r.Project += uint32(len(p.Roots))
		p.Roots = append(p.Roots, r)
	}
	writeStorageCheckPolicy(t, path, p)
	if runSingleWorkerStorageCheckForTest(context.Background(), []string{path}, &out, &errout, inspect) != 1 {
		t.Fatal("unbounded root policy accepted")
	}
}

// Mock the complete native set query; all kernel interactions remain synthetic.
func runSingleWorkerStorageCheckForTest(ctx context.Context, args []string, out, errout io.Writer, inspect func(context.Context, *os.File, projectquota.Expected) (projectquota.Observation, error)) int {
	return runWorkerStorageCheck(ctx, args, out, errout, func(ctx context.Context, roots []projectquota.Root, _ uint64) ([]projectquota.Observation, error) {
		observations := make([]projectquota.Observation, 0, len(roots))
		for _, root := range roots {
			o, err := inspect(ctx, root.Directory, root.Expected)
			if err != nil {
				return nil, err
			}
			observations = append(observations, o)
		}
		return observations, nil
	})
}

func TestWorkerStorageCheckAggregateBudgetsAndIndependentCompanies(t *testing.T) {
	for _, mode := range []string{"boundary", "extra byte", "other worker", "independent company", "truncated observations"} {
		t.Run(mode, func(t *testing.T) {
			path, p := storageCheckFixture(t)
			second := p.Roots[0]
			second.Path += "-second"
			second.Inode++
			second.Project++
			if err := os.Mkdir(second.Path, 0700); err != nil {
				t.Fatal(err)
			}
			p.Roots[0].MaxBytes = uint64(managedStorageLimit) - 4096
			switch mode {
			case "extra byte":
				second.MaxBytes++
			case "other worker":
				second.Worker = "worker-b"
				second.MaxBytes++
			case "independent company":
				second.CompanyID = "company-b"
				p.Roots[0].MaxBytes = uint64(managedStorageLimit)
			}
			p.Roots = append(p.Roots, second)
			writeStorageCheckPolicy(t, path, p)
			var out, errout bytes.Buffer
			calls := 0
			var files []*os.File
			code := runWorkerStorageCheck(t.Context(), []string{path}, &out, &errout, func(ctx context.Context, roots []projectquota.Root, budget uint64) ([]projectquota.Observation, error) {
				calls++
				if budget != uint64(managedStorageLimit) {
					t.Fatal("changed mandatory ceiling")
				}
				if mode == "independent company" && len(roots) != 1 || mode != "independent company" && len(roots) != 2 {
					t.Fatal("wrong company grouping", len(roots))
				}
				for _, f := range files {
					if _, err := f.Stat(); err != nil {
						t.Fatal("prior group root closed before complete observation")
					}
				}
				if mode == "truncated observations" {
					return nil, nil
				}
				var observations []projectquota.Observation
				for _, root := range roots {
					files = append(files, root.Directory)
					observations = append(observations, storageCheckObservation(root.Expected))
				}
				return observations, nil
			})
			switch mode {
			case "extra byte", "other worker":
				if code != 1 || calls != 0 || out.Len() != 0 {
					t.Fatal("aggregate capacity bypass", code, calls)
				}
			case "truncated observations":
				if code != 1 || out.Len() != 0 {
					t.Fatal("partial native set accepted")
				}
			default:
				want := 1
				if mode == "independent company" {
					want = 2
				}
				if code != 0 || calls != want {
					t.Fatal("valid company budget refused", code, calls, errout.String())
				}
			}
			for _, f := range files {
				if _, err := f.Stat(); !errors.Is(err, os.ErrClosed) {
					t.Fatal("group root leaked")
				}
			}
		})
	}
}
