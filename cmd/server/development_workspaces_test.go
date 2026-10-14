package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDevelopmentRoomFlagAndAllowlist(t *testing.T) {
	t.Parallel()
	digest := strings.Repeat("a", 64)
	registry := `{"engineering":"` + digest + `"}`
	for _, flag := range []string{"", "false", "true", "invalid"} {
		got, err := controllerDevelopmentRepository("engineering", flag, registry)
		if flag == "invalid" {
			if err == nil {
				t.Fatal("invalid flag allowed")
			}
			continue
		}
		if err != nil || (got != "") != (flag == "true") {
			t.Fatal(flag, got, err)
		}
	}
	if got, err := controllerDevelopmentRepository("other", "true", registry); err != nil || got != "" {
		t.Fatal("unapproved room enabled")
	}
	for _, invalid := range []string{"", `{"../room":"` + digest + `"}`, `{"engineering":"bad"}`} {
		if _, err := controllerDevelopmentRepository("engineering", "true", invalid); err == nil {
			t.Fatal("bad registry allowed")
		}
	}
}
func TestDevelopmentWorkerIndependentAuthorizationAndIntegrity(t *testing.T) {
	t.Parallel()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "repo.bundle")
	raw := []byte("synthetic bundle fixture; no live repository")
	if err := os.WriteFile(file, raw, 0644); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	digest := hex.EncodeToString(sum[:])
	registry, _ := json.Marshal(map[string]developmentRepository{"engineering": {Bundle: file, SHA256: digest}})
	for _, c := range []struct {
		key, digest, flag string
		ok                bool
	}{
		{"engineering/agent/connection/ssh://worker", digest, "true", true},
		{"engineering/agent/connection/ssh://worker", digest, "false", false},
		{"other/agent/connection/ssh://worker", digest, "true", false},
		{"engineering/agent/connection/ssh://worker", strings.Repeat("b", 64), "true", false},
	} {
		_, err := workerDevelopmentBundle(c.key, c.digest, c.flag, string(registry))
		if (err == nil) != c.ok {
			t.Fatal(c, err)
		}
	}
	if err := os.Chmod(file, 0666); err != nil {
		t.Fatal(err)
	}
	if _, err := workerDevelopmentBundle("engineering/a/c/w", digest, "true", string(registry)); err == nil {
		t.Fatal("world writable source allowed")
	}
	os.Chmod(file, 0644)
	os.WriteFile(file, []byte("tampered"), 0644)
	if _, err := workerDevelopmentBundle("engineering/a/c/w", digest, "true", string(registry)); err == nil {
		t.Fatal("tampered source allowed")
	}
	link := filepath.Join(dir, "link.bundle")
	os.Symlink(file, link)
	registry, _ = json.Marshal(map[string]developmentRepository{"engineering": {Bundle: link, SHA256: digest}})
	if _, err := workerDevelopmentBundle("engineering/a/c/w", digest, "true", string(registry)); err == nil {
		t.Fatal("symlink source allowed")
	}
}
func TestDevelopmentProtocolRequiresSnapshotAndExactToolPolicy(t *testing.T) {
	t.Parallel()
	for _, provider := range []string{"codex", "claude"} {
		v := sampleWorkerRequest()
		v.Provider = provider
		v.Repository = strings.Repeat("a", 64)
		v.Args = providerTurnArgs(provider, true)
		if provider == "codex" {
			v.Args = append(v.Args, "-")
		}
		if _, err := encodeWorkerRequest(v); err != errCredentialBoundary {
			t.Fatal("legacy credential-bearing development must remain unavailable", provider, err)
		}
		v.Repository = ""
		if _, err := encodeWorkerRequest(v); err == nil {
			t.Fatal("development policy without approved snapshot accepted")
		}
		v.Repository = strings.Repeat("a", 64)
		v.Args = append(v.Args, "--add-dir", "/")
		if _, err := encodeWorkerRequest(v); err == nil {
			t.Fatal("arbitrary options allowed")
		}
	}
}
func TestDevelopmentMountPreservesContainment(t *testing.T) {
	t.Parallel()
	args := isolatedRunArgs("image", "name", "volume", "/socket", "/credential", "codex", providerTurnArgs("codex", true))
	args = developmentRunArgs(args, "/approved/repo.bundle", strings.Repeat("a", 64), "commonroom-session-"+strings.Repeat("b", 32))
	joined := strings.Join(args, " ")
	for _, required := range []string{"--network=none", "--read-only", "--cap-drop=ALL", "--security-opt=no-new-privileges:true", "dst=/run/repository.bundle,readonly"} {
		if !strings.Contains(joined, required) {
			t.Fatal("lost boundary", required)
		}
	}
	for _, forbidden := range []string{"--privileged", "SYS_ADMIN", "seccomp=unconfined", "--network=host", "dst=/workspace,readonly"} {
		if strings.Contains(joined, forbidden) {
			t.Fatal(forbidden)
		}
	}
}

func TestDevelopmentSessionAuthorityNeverCrossesModesOrSnapshots(t *testing.T) {
	t.Parallel()
	digest := strings.Repeat("a", 64)
	for _, c := range []struct {
		previous, repository string
		retain, reject       bool
	}{
		{"container-v1", "", true, false},
		{"container-v1", digest, false, false},
		{"container-development-v1:" + digest, digest, true, false},
		{"container-development-v1:" + digest, "", false, false},
		{"container-development-v1:" + strings.Repeat("b", 64), digest, false, false},
		{"host", digest, false, true},
	} {
		_, thread, err := agentExecutionSession(c.previous, "fixture-session", c.repository)
		if (err != nil) != c.reject || (err == nil && (thread != "") != c.retain) {
			t.Fatal(c, thread, err)
		}
	}
}

func TestDevelopmentConfigurationPersistenceAndEnvironmentOverride(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "development.json")
	digest := strings.Repeat("a", 64)
	raw := `{"enabled":true,"repositories":{"engineering":"` + digest + `"}}`
	if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}
	enabled, registry, err := loadControllerDevelopmentConfiguration(path, "", "")
	got, _ := controllerDevelopmentRepository("engineering", enabled, registry)
	if err != nil || got != digest {
		t.Fatal("persistent setup unavailable", err)
	}
	enabled, registry, err = loadControllerDevelopmentConfiguration(path, "false", "")
	got, _ = controllerDevelopmentRepository("engineering", enabled, registry)
	if err != nil || got != "" {
		t.Fatal("explicit disabled flag ignored")
	}
	if err := os.Chmod(path, 0644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := loadControllerDevelopmentConfiguration(path, "", ""); err == nil {
		t.Fatal("shared config admitted")
	}
	os.Chmod(path, 0600)
	for _, invalid := range []string{raw + ` {}`, `{"enabled":true,"mount":"/"}`, `{"enabled":true,"repositories":{"engineering":"bad"}}`} {
		os.WriteFile(path, []byte(invalid), 0600)
		if _, _, err := loadControllerDevelopmentConfiguration(path, "", ""); err == nil {
			t.Fatal("invalid config admitted")
		}
	}
	if _, _, err := loadControllerDevelopmentConfiguration(filepath.Join(t.TempDir(), "absent"), "", ""); err != nil {
		t.Fatal("missing config must default off")
	}
}
