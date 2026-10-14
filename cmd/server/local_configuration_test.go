package main

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestLocalConfigurationStrictDecoding(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "config.json")
	got, err := loadLocalConfiguration(path)
	if err != nil || got != (localConfiguration{}) {
		t.Fatal("missing configuration must preserve defaults", err)
	}
	raw := `{"databaseUrl":"postgres://synthetic.invalid/example","remoteWorkerAlias":"synthetic-worker","remoteWorkerEnabled":true}`
	if err = os.WriteFile(path, []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}
	got, err = loadLocalConfiguration(path)
	if err != nil || got.RemoteWorkerAlias != "synthetic-worker" || !got.RemoteWorkerEnabled {
		t.Fatal("valid configuration rejected", err)
	}
	for _, bad := range []string{raw + ` {}`, `{"unknown":"synthetic-secret"}`, `{"databaseUrl":42}`, `[]`, `null`} {
		if err = os.WriteFile(path, []byte(bad), 0600); err != nil {
			t.Fatal(err)
		}
		got, err = loadLocalConfiguration(path)
		if err == nil || got != (localConfiguration{}) {
			t.Fatalf("invalid configuration accepted: %q", bad)
		}
		if strings.Contains(err.Error(), "synthetic-secret") {
			t.Fatal("input leaked in decode error")
		}
	}
}

func TestControllerConfigurationFileBoundaries(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"symlink", "hardlink", "public", "fifo", "oversized"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			path := filepath.Join(dir, "config.json")
			target := filepath.Join(dir, "target")
			if err := os.WriteFile(target, []byte(`{}`), 0600); err != nil {
				t.Fatal(err)
			}
			var err error
			switch kind {
			case "symlink":
				err = os.Symlink(target, path)
			case "hardlink":
				err = os.Link(target, path)
			case "public":
				err = os.WriteFile(path, []byte(`{}`), 0600)
				if err == nil {
					err = os.Chmod(path, 0644)
				}
			case "fifo":
				err = syscall.Mkfifo(path, 0600)
			case "oversized":
				err = os.WriteFile(path, []byte(strings.Repeat(" ", 8193)), 0600)
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, err = loadLocalConfiguration(path); err == nil {
				t.Fatal("unsafe local configuration admitted")
			}
			if _, _, err = loadControllerDevelopmentConfiguration(path, "", ""); err == nil {
				t.Fatal("unsafe development configuration admitted")
			}
		})
	}
}
