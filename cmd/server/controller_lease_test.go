package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestControllerLeaseExcludesConcurrentStartup(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "data", "controller.lock")
	first, err := acquireControllerLease(path)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	before, err := os.Stat(path)
	if err != nil || before.Mode().Perm() != 0600 {
		t.Fatal("lease must be private", err)
	}
	second, err := acquireControllerLease(path)
	if err == nil || second != nil {
		if second != nil {
			second.Close()
		}
		t.Fatal("concurrent controller admitted")
	}
	first.Close()
	first.Close() // Cleanup is safe after an explicit release.
	third, err := acquireControllerLease(path)
	if err != nil {
		t.Fatal("released lease unavailable", err)
	}
	defer third.Close()
	after, err := os.Stat(path)
	if err != nil || !os.SameFile(before, after) {
		t.Fatal("lease inode replaced or removed", err)
	}
}

func TestControllerLeaseRejectsLinkedLock(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"symlink", "hardlink"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			target, path := filepath.Join(dir, "target"), filepath.Join(dir, "controller.lock")
			if err := os.WriteFile(target, []byte("synthetic"), 0600); err != nil {
				t.Fatal(err)
			}
			var err error
			if kind == "symlink" {
				err = os.Symlink(target, path)
			} else {
				err = os.Link(target, path)
			}
			if err != nil {
				t.Fatal(err)
			}
			lease, err := acquireControllerLease(path)
			if err == nil || lease != nil {
				if lease != nil {
					lease.Close()
				}
				t.Fatal("linked lease admitted")
			}
			raw, err := os.ReadFile(target)
			if err != nil || string(raw) != "synthetic" {
				t.Fatal("target changed", err)
			}
		})
	}
}
