package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

var repositoryDigestPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

// developmentRepository is administrator-owned, never accepted from an event or
// prompt. Only this bounded immutable bundle is mounted, read-only; an agent gets
// a separate writable clone. CAN replacing this with a host worktree expose
// secrets and permit destructive host edits? Yes, BUT it breaks this contract.
type developmentRepository struct {
	Bundle string `json:"bundle"`
	SHA256 string `json:"sha256"`
}

// controllerDevelopmentRepository is the server-enforced off-by-default feature
// flag. Enabling it also requires an explicit room-to-snapshot allowlist. Only a
// deployment administrator can change these environment settings; room members
// cannot request paths, turn tools on, or grant themselves repository access.
func controllerDevelopmentRepository(roomID, enabled, registry string) (string, error) {
	if enabled == "" || enabled == "false" {
		return "", nil
	}
	if enabled != "true" {
		return "", errors.New("COMMONROOM_DEVELOPMENT_WORKSPACES must be true or false")
	}
	var repositories map[string]string
	if json.Unmarshal([]byte(registry), &repositories) != nil {
		return "", errors.New("Configure a valid development repository room allowlist")
	}
	for room, digest := range repositories {
		if !workerAliasPattern.MatchString(room) || !repositoryDigestPattern.MatchString(digest) {
			return "", errors.New("Invalid development repository room allowlist")
		}
	}
	return repositories[roomID], nil
}

// providerTurnArgs defines two exact policies. The development policy uses the
// mandatory outer Docker sandbox as its execution backend, avoiding nested bwrap
// user namespaces. It must NEVER be used by a host launcher. Read-only chat stays
// unchanged, including for connection verification and credential management.
func providerTurnArgs(provider string, development bool) []string {
	if provider == "codex" {
		sandbox := "read-only"
		if development {
			sandbox = "danger-full-access"
		}
		args := []string{"exec", "--json", "--skip-git-repo-check", "--ignore-user-config", "-s", sandbox, "-c", `approval_policy="never"`}
		if development {
			args = append(args, "-c", `shell_environment_policy.inherit="all"`)
		}
		return args
	}
	tools := ""
	if development {
		tools = "Bash,Read,Edit,Write,Glob,Grep"
	}
	args := []string{"--print", "--output-format", "json", "--tools", tools, "--strict-mcp-config", "--mcp-config", `{"mcpServers":{}}`, "--disable-slash-commands", "--safe-mode"}
	if development {
		args = append(args, "--dangerously-skip-permissions")
	}
	return args
}

// developmentTurnArgs recognizes only the exact development policy prefix.
func developmentTurnArgs(provider string, args []string) bool {
	base := providerTurnArgs(provider, true)
	return len(args) >= len(base) && slices.Equal(args[:len(base)], base)
}

// validateDevelopmentRegistry bounds the private worker registry and disallows
// paths whose spelling could change Docker mount-option interpretation.
func validateDevelopmentRegistry(repositories map[string]developmentRepository) error {
	if len(repositories) > 64 {
		return errors.New("Too many development repositories")
	}
	for room, repository := range repositories {
		if !workerAliasPattern.MatchString(room) || !repositoryDigestPattern.MatchString(repository.SHA256) || !filepath.IsAbs(repository.Bundle) || strings.ContainsAny(repository.Bundle, ",\r\n") {
			return errors.New("Invalid approved development bundle configuration")
		}
	}
	return nil
}

// workerDevelopmentBundle independently enforces the worker administrator's
// feature flag, room mapping, snapshot identity and file permissions. A forged
// controller request cannot choose another room's bundle or an arbitrary mount.
func workerDevelopmentBundle(key, digest, enabled, registry string) (string, error) {
	if digest == "" {
		return "", nil
	}
	if enabled != "true" {
		return "", errors.New("Development workspaces are disabled on this worker")
	}
	var repositories map[string]developmentRepository
	if json.Unmarshal([]byte(registry), &repositories) != nil || validateDevelopmentRegistry(repositories) != nil {
		return "", errors.New("Invalid development worker registry")
	}
	room := strings.SplitN(key, "/", 2)[0]
	repository, ok := repositories[room]
	if !ok || repository.SHA256 != digest {
		return "", errors.New("No matching approved development snapshot for this room")
	}
	resolved, err := filepath.EvalSymlinks(repository.Bundle)
	if err != nil || resolved != repository.Bundle {
		return "", errors.New("Development bundle must use a canonical path without symlinks")
	}
	file, err := os.Open(repository.Bundle)
	if err != nil {
		return "", errors.New("Approved development bundle unavailable")
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0022 != 0 || info.Size() <= 0 || info.Size() > 256*1024*1024 {
		return "", errors.New("Development bundle must be a bounded regular file without group/world write access")
	}
	hash := sha256.New()
	if _, err = io.Copy(hash, file); err != nil || hex.EncodeToString(hash.Sum(nil)) != digest {
		return "", errors.New("Development bundle checksum mismatch")
	}
	return repository.Bundle, nil
}

// developmentRunArgs adds only the prevalidated snapshot and fixed bootstrap
// settings. Network, capabilities, read-only root and all resource limits remain
// intact. The snapshot is never a writable host mount.
func developmentRunArgs(args []string, bundle, digest, volume string) []string {
	additions := []string{"--mount", "type=bind,src=" + bundle + ",dst=/run/repository.bundle,readonly", "--env", "COMMONROOM_DEVELOPMENT_DIGEST=" + digest, "--env", "COMMONROOM_DEVELOPMENT_BRANCH=feature/" + volume, "--workdir", "/workspace/repository"}
	// The work directory must exist before Docker starts; bootstrap changes cwd.
	additions[len(additions)-1] = "/workspace"
	for i, arg := range args {
		if arg == "--entrypoint" {
			return append(append(slices.Clone(args[:i]), additions...), args[i:]...)
		}
	}
	return nil // impossible for the supported launcher; fail closed on policy drift.
}

// agentExecutionSession never resumes a provider session across workspace mode
// or snapshot changes. Existing room history is supplied separately in prompts.
// Legacy host sessions remain rejected rather than acquiring container authority.
func agentExecutionSession(previous, thread, repository string) (string, string, error) {
	mode := "container-v1"
	if repository != "" {
		mode = "container-development-v1:" + repository
	}
	if previous != mode && strings.HasPrefix(previous, "container") {
		thread = ""
	}
	if thread != "" && previous != mode {
		return "", "", errors.New("Legacy host session cannot resume inside isolation; invite a fresh agent")
	}
	return mode, thread, nil
}

// loadControllerDevelopmentConfiguration preserves operator setup across normal
// restarts without copying the process environment or credentials. Explicit env
// values override this dedicated owner-only, non-secret configuration file.
func loadControllerDevelopmentConfiguration(path, enabled, registry string) (string, string, error) {
	if enabled == "" || registry == "" {
		info, err := os.Lstat(path)
		if err != nil && !os.IsNotExist(err) {
			return "", "", err
		}
		if err == nil {
			if !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 || info.Size() > 8192 {
				return "", "", errors.New("Development configuration must be a bounded owner-only regular file")
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				return "", "", err
			}
			var config struct {
				Enabled      bool              `json:"enabled"`
				Repositories map[string]string `json:"repositories"`
			}
			decoder := json.NewDecoder(strings.NewReader(string(raw)))
			decoder.DisallowUnknownFields()
			if decoder.Decode(&config) != nil || decoder.Decode(new(any)) != io.EOF || len(config.Repositories) > 64 {
				return "", "", errors.New("Invalid development configuration")
			}
			if enabled == "" {
				if config.Enabled {
					enabled = "true"
				} else {
					enabled = "false"
				}
			}
			if registry == "" {
				encoded, _ := json.Marshal(config.Repositories)
				registry = string(encoded)
			}
		}
	}
	if _, err := controllerDevelopmentRepository("", enabled, registry); err != nil {
		return "", "", err
	}
	return enabled, registry, nil
}
