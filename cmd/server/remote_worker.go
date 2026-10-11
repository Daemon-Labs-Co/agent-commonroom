package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
)

var workerAliasPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,63}$`)
var workerJobPattern = regexp.MustCompile(`^[a-f0-9]{32}$`)

func validWorkerJobID(v string) bool { return workerJobPattern.MatchString(v) }

// remoteWorkerEndpoint admits only the explicitly configured worker alias.
// SSH transports a constrained worker protocol, never arbitrary model commands.
// Disabling remote transport never restores direct host execution.
func remoteWorkerEndpoint(endpoint, configured, enabled string) (string, string, error) {
	if enabled != "true" {
		return "", "", errors.New("Remote worker transport is disabled; direct SSH execution is forbidden")
	}
	if !workerAliasPattern.MatchString(configured) {
		return "", "", errors.New("Configure COMMONROOM_REMOTE_WORKER_ALIAS")
	}
	provider := "codex"
	alias := strings.TrimPrefix(endpoint, "ssh://")
	if strings.HasSuffix(alias, "/claude") {
		provider = "claude"
		alias = strings.TrimSuffix(alias, "/claude")
	}
	if !strings.HasPrefix(endpoint, "ssh://") || alias != configured {
		return "", "", errors.New("Connection does not target the approved isolated worker")
	}
	return alias, provider, nil
}

type workerRunRequest struct {
	Version    int      `json:"version"`
	JobID      string   `json:"jobId"`
	Provider   string   `json:"provider"`
	Key        string   `json:"key"`
	Args       []string `json:"args"`
	Repository string   `json:"repository,omitempty"`
}

func encodeWorkerRequest(v workerRunRequest) (string, error) {
	if err := validateWorkerRun(v); err != nil {
		return "", err
	}
	raw, _ := json.Marshal(v)
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
func decodeWorkerRequest(encoded string) (workerRunRequest, error) {
	var v workerRunRequest
	if len(encoded) > 16384 {
		return v, errors.New("Worker request too large")
	}
	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return v, errors.New("Invalid worker request")
	}
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&v); err != nil {
		return v, errors.New("Invalid worker request")
	}
	if decoder.Decode(new(any)) != io.EOF {
		return v, errors.New("Trailing worker request data")
	}
	return v, validateWorkerRun(v)
}

// validateWorkerRun forbids arbitrary CLI options, mounts, environment and shell
// commands. Provider arguments must match the exact chat policy or the explicit
// development policy bound to an approved snapshot. The worker independently
// verifies its administrator-owned room registry before any launch.
func validateWorkerRun(v workerRunRequest) error {
	if v.Version != 1 || !validWorkerJobID(v.JobID) || strings.TrimSpace(v.Key) == "" || len(v.Key) > 1024 {
		return errors.New("Invalid worker identity")
	}
	base := []string{}
	suffix := []string{}
	threadFlag := ""
	if v.Repository != "" && !repositoryDigestPattern.MatchString(v.Repository) {
		return errors.New("Invalid development repository digest")
	}
	if v.Provider != "codex" && v.Provider != "claude" {
		return errors.New("Unsupported worker provider")
	}
	base = providerTurnArgs(v.Provider, v.Repository != "")
	if v.Provider == "codex" {
		suffix = []string{"-"}
		threadFlag = "resume"
	} else {
		threadFlag = "--resume"
	}

	n := len(base) + len(suffix)
	if len(v.Args) != n && len(v.Args) != n+2 {
		return errors.New("Forbidden worker arguments")
	}
	for i, arg := range base {
		if v.Args[i] != arg {
			return errors.New("Forbidden worker arguments")
		}
	}
	for i, arg := range suffix {
		if v.Args[len(v.Args)-len(suffix)+i] != arg {
			return errors.New("Forbidden worker arguments")
		}
	}
	if len(v.Args) == n+2 {
		thread := v.Args[len(base)+1]
		if v.Args[len(base)] != threadFlag || len(thread) > 128 || !regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`).MatchString(thread) {
			return errors.New("Invalid worker session")
		}
	}
	return nil
}
func workerSSHCommand(ctx context.Context, alias, action, value string) *exec.Cmd {
	// action and value originate from validated enum/base64/hex values; quoting
	// still protects the remote shell. Host keys must already be trusted.
	remote := "exec ~/.local/bin/commonroom-worker " + action
	if value != "" {
		remote += " " + shellArgument(value)
	}
	return exec.CommandContext(ctx, "ssh", "-T", "-o", "BatchMode=yes", "-o", "StrictHostKeyChecking=yes", "-o", "ConnectTimeout=10", "-o", "ServerAliveInterval=10", "-o", "ServerAliveCountMax=2", alias, remote)
}
func remoteAgentCommand(ctx context.Context, endpoint, key string, args []string) (*exec.Cmd, func(), error) {
	alias, provider, err := remoteWorkerEndpoint(endpoint, os.Getenv("COMMONROOM_REMOTE_WORKER_ALIAS"), os.Getenv("COMMONROOM_REMOTE_WORKER"))
	if err != nil {
		return nil, nil, err
	}
	repository := ""
	if developmentTurnArgs(provider, args) {
		repository, err = controllerDevelopmentRepository(strings.SplitN(key, "/", 2)[0], os.Getenv("COMMONROOM_DEVELOPMENT_WORKSPACES"), os.Getenv("COMMONROOM_DEVELOPMENT_REPOSITORIES"))
		if err != nil || repository == "" {
			return nil, nil, errors.New("Development workspace is not enabled for this room")
		}
	}
	request := workerRunRequest{Version: 1, JobID: id(), Provider: provider, Key: key, Args: args, Repository: repository}
	encoded, err := encodeWorkerRequest(request)
	if err != nil {
		return nil, nil, err
	}
	var once sync.Once
	cleanup := func() {
		once.Do(func() {
			cancelCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			var result limitedAgentOutput
			cmd := workerSSHCommand(cancelCtx, alias, "cancel-v1", request.JobID)
			cmd.Stdout = &result
			if err := cmd.Run(); err != nil {
				fmt.Fprintln(os.Stderr, "Remote worker cancellation unconfirmed:", request.JobID)
				return
			}
			var ack struct {
				JobID   string `json:"jobId"`
				Stopped bool   `json:"stopped"`
			}
			if json.Unmarshal(result.Bytes(), &ack) != nil || ack.JobID != request.JobID || !ack.Stopped {
				fmt.Fprintln(os.Stderr, "Remote worker cancellation unconfirmed:", request.JobID)
			}
		})
	}
	cmd := workerSSHCommand(ctx, alias, "run-v1", encoded)
	cmd.WaitDelay = 2 * time.Second
	return cmd, cleanup, nil
}

type workerConfiguration struct {
	Image                   string                           `json:"image"`
	AuthDir                 string                           `json:"authDir"`
	DevelopmentEnabled      bool                             `json:"developmentEnabled,omitempty"`
	DevelopmentRepositories map[string]developmentRepository `json:"developmentRepositories,omitempty"`
}
type workerJobRecord struct {
	JobID    string `json:"jobId"`
	Status   string `json:"status"`
	Provider string `json:"provider"`
}

func workerStateDirectory() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".local/state/commonroom-worker/jobs")
	err = os.MkdirAll(dir, 0700)
	return dir, err
}
func writeWorkerRecord(dir string, v workerJobRecord, exclusive bool) error {
	raw, _ := json.Marshal(v)
	path := filepath.Join(dir, v.JobID+".json")
	if exclusive {
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err = f.Write(raw); err != nil {
			return err
		}
		return f.Sync()
	}
	f, err := os.CreateTemp(dir, ".outcome-")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(raw); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}
func readWorkerRecord(dir, job string) (workerJobRecord, error) {
	var v workerJobRecord
	raw, err := os.ReadFile(filepath.Join(dir, job+".json"))
	if err != nil || len(raw) > 8192 {
		return v, errors.New("Unknown worker job")
	}
	if json.Unmarshal(raw, &v) != nil || v.JobID != job {
		return v, errors.New("Invalid worker record")
	}
	return v, nil
}
func workerTerminal(status string) bool {
	return status == "finished" || status == "failed" || status == "cancelled"
}

// workerStreamWriter bounds remote output independently of the controller.
type workerStreamWriter struct {
	writer    io.Writer
	remaining int
	cancel    context.CancelFunc
}

func (w *workerStreamWriter) Write(p []byte) (int, error) {
	if len(p) > w.remaining {
		w.cancel()
		return 0, errors.New("Worker output exceeded limit")
	}
	n, err := w.writer.Write(p)
	w.remaining -= n
	if err != nil {
		w.cancel()
	}
	return n, err
}

func configureWorker() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(home, ".config/commonroom-worker/config.json")
	fi, err := os.Lstat(path)
	if err != nil || !fi.Mode().IsRegular() || fi.Mode().Perm()&0077 != 0 || fi.Size() > 8192 {
		return errors.New("Worker requires owner-only ~/.config/commonroom-worker/config.json")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var c workerConfiguration
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&c) != nil || decoder.Decode(new(any)) != io.EOF || !validAgentImage(c.Image) || !filepath.IsAbs(c.AuthDir) {
		return errors.New("Worker requires a digest-pinned image and absolute authDir")
	}
	if err := validateDevelopmentRegistry(c.DevelopmentRepositories); err != nil {
		return err
	}
	registry, _ := json.Marshal(c.DevelopmentRepositories)
	os.Setenv("COMMONROOM_WORKER_DEVELOPMENT_REPOSITORIES", string(registry))
	os.Setenv("COMMONROOM_WORKER_DEVELOPMENT", fmt.Sprint(c.DevelopmentEnabled))
	os.Setenv("COMMONROOM_AGENT_IMAGE", c.Image)
	os.Setenv("COMMONROOM_AGENT_AUTH_DIR", c.AuthDir)
	return nil
}

// agentWorkerMain is a database-free entry point, invoked only by the installed
// worker wrapper over authenticated SSH. Provider secrets remain on this host.
// A fork CAN bypass isolation, BUT removes the documented containment boundary.
func agentWorkerMain(args []string, in io.Reader, out, diagnostics io.Writer) int {
	fail := func(err error) int { fmt.Fprintln(diagnostics, err); return 1 }
	if len(args) == 0 {
		return fail(errors.New("Worker action required"))
	}
	if args[0] == "auth-login-v1" || args[0] == "auth-status-v1" || args[0] == "auth-logout-v1" || args[0] == "verify-v1" {
		return workerProviderAuth(args, in, out, diagnostics)
	}
	if args[0] == "ready-v1" && len(args) == 1 {
		if err := configureWorker(); err != nil {
			return fail(err)
		}
		if err := isolationAvailable("codex"); err != nil {
			return fail(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		info, err := exec.CommandContext(ctx, "docker", "info", "--format", "{{json .}}").Output()
		if err != nil || !dockerSecurityReady(info) {
			return fail(errors.New("Rootless Docker or required limits unavailable"))
		}
		if exec.CommandContext(ctx, "docker", "image", "inspect", os.Getenv("COMMONROOM_AGENT_IMAGE")).Run() != nil {
			return fail(errors.New("Configured agent image is unavailable"))
		}
		credentials := map[string]bool{}
		for provider, filename := range map[string]string{"codex": "auth.json", "claude": ".credentials.json"} {
			fi, err := os.Lstat(filepath.Join(os.Getenv("COMMONROOM_AGENT_AUTH_DIR"), provider, filename))
			credentials[provider] = err == nil && fi.Mode().IsRegular() && fi.Mode().Perm()&0077 == 0 && fi.Size() <= 65536
		}
		json.NewEncoder(out).Encode(map[string]any{"version": 1, "configured": true, "runtimeReady": true, "credentials": credentials})
		return 0
	}
	if len(args) != 2 {
		return fail(errors.New("Invalid worker request"))
	}

	if args[0] == "status-v1" && validWorkerJobID(args[1]) {
		dir, err := workerStateDirectory()
		if err != nil {
			return fail(err)
		}
		v, err := readWorkerRecord(dir, args[1])
		if err != nil {
			return fail(err)
		}
		json.NewEncoder(out).Encode(v)
		return 0
	}
	if args[0] == "cancel-v1" {
		job := args[1]
		if !validWorkerJobID(job) {
			return fail(errors.New("Invalid worker job identity"))
		}
		dir, err := workerStateDirectory()
		if err != nil {
			return fail(err)
		}
		if _, err = readWorkerRecord(dir, job); err != nil {
			return fail(err)
		}
		if err = os.WriteFile(filepath.Join(dir, job+".cancel"), []byte("cancel"), 0600); err != nil {
			return fail(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		defer cancel()
		name := "commonroom-turn-" + job
		for {
			v, err := readWorkerRecord(dir, job)
			if err != nil {
				return fail(err)
			}
			var ids limitedAgentOutput
			cmd := exec.CommandContext(ctx, "docker", "ps", "-a", "--filter", "name=^/"+name+"$", "--format", "{{.ID}}")
			cmd.Stdout = &ids
			if cmd.Run() != nil {
				return fail(errors.New("Container cancellation unconfirmed"))
			}
			exists := strings.TrimSpace(ids.String()) != ""
			if exists && exec.CommandContext(ctx, "docker", "rm", "--force", name).Run() != nil {
				return fail(errors.New("Container cancellation unconfirmed"))
			}
			if !exists && workerTerminal(v.Status) {
				json.NewEncoder(out).Encode(map[string]any{"jobId": job, "stopped": true, "cancellationRequested": true})
				return 0
			}
			select {
			case <-ctx.Done():
				return fail(errors.New("Container cancellation unconfirmed"))
			case <-time.After(100 * time.Millisecond):
			}
		}
	}

	if args[0] != "run-v1" {
		return fail(errors.New("Unsupported worker action"))
	}
	request, err := decodeWorkerRequest(args[1])
	if err != nil {
		return fail(err)
	}
	if err = configureWorker(); err != nil {
		return fail(err)
	}
	dir, err := workerStateDirectory()
	if err != nil {
		return fail(err)
	}
	record := workerJobRecord{JobID: request.JobID, Provider: request.Provider, Status: "starting"}
	if err = writeWorkerRecord(dir, record, true); err != nil {
		return fail(errors.New("Duplicate or unavailable worker job; do not resubmit"))
	}
	signalCtx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	defer stopSignals()
	ctx, cancel := context.WithTimeout(signalCtx, 3*time.Minute)
	defer cancel()
	latch := filepath.Join(dir, request.JobID+".cancel")
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if _, err := os.Stat(latch); err == nil {
					cancel()
					return
				}
			}
		}
	}()
	type promptResult struct {
		data []byte
		err  error
	}
	promptCh := make(chan promptResult, 1)
	go func() { data, err := io.ReadAll(io.LimitReader(in, 1024*1024+1)); promptCh <- promptResult{data, err} }()
	var prompt []byte
	select {
	case result := <-promptCh:
		prompt, err = result.data, result.err
	case <-ctx.Done():
		err = ctx.Err()
	}

	if err != nil || len(prompt) > 1024*1024 {
		record.Status = "failed"
		writeWorkerRecord(dir, record, false)
		return fail(errors.New("Worker prompt exceeded limit"))
	}
	cmd, cleanup, err := localIsolatedDevelopmentCommand(ctx, request.Provider, request.Key, request.JobID, request.Repository, request.Args...)
	if err != nil {
		record.Status = "failed"
		writeWorkerRecord(dir, record, false)
		return fail(err)
	}
	if ctx.Err() != nil {
		cleanup()
		record.Status = "cancelled"
		writeWorkerRecord(dir, record, false)
		return fail(errors.New("Worker job cancelled before launch"))
	}
	record.Status = "running"
	if err = writeWorkerRecord(dir, record, false); err != nil {
		cleanup()
		return fail(err)
	}
	cmd.Stdin = strings.NewReader(string(prompt))
	cmd.Stdout = &workerStreamWriter{writer: out, remaining: 4 * 1024 * 1024, cancel: cancel}
	cmd.Stderr = &workerStreamWriter{writer: diagnostics, remaining: 4 * 1024 * 1024, cancel: cancel}
	err = cmd.Run()
	cleanup()
	record.Status = "finished"
	if err != nil {
		record.Status = "failed"
	}
	if ctx.Err() != nil {
		record.Status = "cancelled"
	}
	if writeWorkerRecord(dir, record, false) != nil {
		return fail(errors.New("Worker outcome could not be recorded"))
	}
	if err != nil {
		return fail(errors.New("Isolated worker turn failed"))
	}
	return 0
}
