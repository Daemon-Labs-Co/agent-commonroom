package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// Connection credentials are isolated from other connections. Agent execution
// remains disposable and separately scoped; no host CLI is used for sign-in.
type workerAuthRequest struct {
	Version    int    `json:"version"`
	Connection string `json:"connection"`
	Provider   string `json:"provider"`
}

func connectionSignInEnabled(flags map[string]bool) bool {
	v, ok := flags["connectionSignIn"]
	return !ok || v
}
func connectionAuthKey(connection string) string {
	digest := sha256.Sum256([]byte(connection))
	return hex.EncodeToString(digest[:])
}
func executionConnectionID(key string) (string, error) {
	parts := strings.SplitN(key, "/", 4)
	if len(parts) != 4 || !workerAliasPattern.MatchString(parts[2]) {
		return "", errors.New("Execution requires a valid connection identity")
	}
	return parts[2], nil
}
func encodeWorkerAuth(v workerAuthRequest) (string, error) {
	if v.Version != 1 || !workerAliasPattern.MatchString(v.Connection) || (v.Provider != "codex" && v.Provider != "claude") {
		return "", errors.New("Invalid authentication request")
	}
	raw, _ := json.Marshal(v)
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
func decodeWorkerAuth(encoded string) (workerAuthRequest, error) {
	var v workerAuthRequest
	if len(encoded) > 1024 {
		return v, errors.New("Authentication request too large")
	}
	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return v, err
	}
	d := json.NewDecoder(strings.NewReader(string(raw)))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil || d.Decode(new(any)) != io.EOF {
		return v, errors.New("Invalid authentication request")
	}
	_, err = encodeWorkerAuth(v)
	return v, err
}
func interactiveWorkerLogin(alias, encoded string) string {
	return "ssh -tt -o BatchMode=yes -o StrictHostKeyChecking=yes -o ConnectTimeout=10 " + shellArgument(alias) + " " + shellArgument("exec ~/.local/bin/commonroom-worker auth-login-v1 "+shellArgument(encoded))
}
func providerCredentialName(provider string) string {
	if provider == "claude" {
		return ".credentials.json"
	}
	return "auth.json"
}

// Forget only this connection/provider's stored credential; provider-wide revocation is separate.
func clearWorkerCredential(root, provider string) error {
	err := os.Remove(filepath.Join(root, providerCredentialName(provider)))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
func lockWorkerCredential(root string) (func(), error) {
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filepath.Join(root, ".auth.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, errors.New("Connection authentication is busy; finish the current sign-in first")
	}
	return func() { syscall.Flock(int(f.Fd()), syscall.LOCK_UN); f.Close() }, nil
}
func credentialPresent(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil || !fi.Mode().IsRegular() || fi.Mode().Perm()&0077 != 0 || fi.Size() > 65536 {
		return false
	}
	raw, err := os.ReadFile(path)
	return err == nil && len(raw) > 2 && json.Valid(raw)
}

// workerProviderAuth permits only provider login/status in a constrained image.
// Login is a human administrative operation, bounded to fifteen minutes, with
// the same network gateway and resource limits as agent turns. There is no
// credential upload endpoint or arbitrary command/mount option.
func workerProviderAuth(args []string, in io.Reader, out, diagnostics io.Writer) int {
	fail := func(err error) int { fmt.Fprintln(diagnostics, err); return 1 }
	if len(args) != 2 {
		return fail(errors.New("Invalid authentication request"))
	}
	v, err := decodeWorkerAuth(args[1])
	if err != nil {
		return fail(err)
	}
	if err = configureWorker(); err != nil {
		return fail(err)
	}
	root := filepath.Join(os.Getenv("COMMONROOM_AGENT_AUTH_DIR"), connectionAuthKey(v.Connection), v.Provider)
	credential := filepath.Join(root, providerCredentialName(v.Provider))
	if args[0] == "auth-status-v1" {
		json.NewEncoder(out).Encode(map[string]bool{"credentialPresent": credentialPresent(credential)})
		return 0
	}
	if args[0] == "verify-v1" {
		return workerVerifyConnection(v, out, diagnostics)
	}
	if args[0] != "auth-login-v1" && args[0] != "auth-logout-v1" {
		return fail(errors.New("Unsupported authentication action"))
	}
	unlock, err := lockWorkerCredential(root)
	if err != nil {
		return fail(err)
	}
	defer unlock()
	if args[0] == "auth-logout-v1" {
		if err = clearWorkerCredential(root, v.Provider); err != nil {
			return fail(errors.New("Unable to clear connection credentials"))
		}
		json.NewEncoder(out).Encode(map[string]bool{"signedOut": true})
		return 0
	}
	if err = isolationAvailable(v.Provider); err != nil {
		return fail(err)
	}
	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	defer stop()
	ctx, cancel := context.WithTimeout(signalCtx, 15*time.Minute)
	defer cancel()
	info, err := exec.CommandContext(ctx, "docker", "info", "--format", "{{json .}}").Output()
	if err != nil || !dockerSecurityReady(info) {
		return fail(errors.New("Rootless Docker limits unavailable"))
	}
	host, err := exec.CommandContext(ctx, "docker", "context", "inspect", "--format", "{{.Endpoints.docker.Host}}").Output()
	if err != nil || !strings.HasPrefix(strings.TrimSpace(string(host)), "unix://") || os.Getenv("DOCKER_HOST") != "" || os.Getenv("DOCKER_CONTEXT") != "" {
		return fail(errors.New("Local rootless Docker context required"))
	}
	dir, err := os.MkdirTemp("", "commonroom-login-")
	if err != nil {
		return fail(err)
	}
	defer os.RemoveAll(dir)
	gateway, err := newProviderGateway(ctx, filepath.Join(dir, "provider.sock"), v.Provider)
	if err != nil {
		return fail(err)
	}
	defer gateway.Close()
	name := "commonroom-login-" + connectionAuthKey(v.Connection)[:32]
	// Container-name collision serializes login for this connection. A competing
	// launch never removes the other login container.
	volume := name + "-" + id()
	defer func() {
		cleanup, done := context.WithTimeout(context.Background(), 10*time.Second)
		defer done()
		exec.CommandContext(cleanup, "docker", "volume", "rm", volume, volume+"-work").Run()
	}()
	launch := isolatedRunArgs(os.Getenv("COMMONROOM_AGENT_IMAGE"), name, volume, gateway.socket, "/unused", v.Provider, []string{"auth-login"})
	filtered := []string{}
	for i := 0; i < len(launch); i++ {
		if launch[i] == "--rm" {
			continue
		}
		if launch[i] == "--mount" && i+1 < len(launch) && strings.Contains(launch[i+1], "dst=/run/provider-auth.json") {
			i++
			continue
		}
		filtered = append(filtered, launch[i])
	}
	// Create separately so failed name collisions cannot lead to another login
	// being removed. stdin remains interactive for provider authorization codes.
	filtered[0] = "create"
	filtered = append([]string{filtered[0], "-i", "-t"}, filtered[1:]...)
	var result cappedStatusOutput
	create := exec.CommandContext(ctx, "docker", filtered...)
	create.Stdout = &result
	create.Stderr = diagnostics
	if err = create.Run(); err != nil {
		return fail(errors.New("Sign-in container unavailable; another login may be active"))
	}
	defer func() {
		cleanup, done := context.WithTimeout(context.Background(), 10*time.Second)
		defer done()
		if exec.CommandContext(cleanup, "docker", "rm", "--force", name).Run() != nil {
			fmt.Fprintln(diagnostics, "Sign-in container cleanup unconfirmed:", name)
		}
	}()
	run := exec.CommandContext(ctx, "docker", "start", "-ai", name)
	run.Stdin = in
	run.Stdout = out
	run.Stderr = diagnostics
	if run.Run() != nil {
		return fail(errors.New("Provider sign-in did not complete; existing credentials were preserved"))
	}
	exportArgs, err := workerCredentialExportArgs(os.Getenv("COMMONROOM_AGENT_IMAGE"), volume, v.Provider)
	if err != nil {
		return fail(err)
	}
	var exported credentialExportOutput
	export := exec.CommandContext(ctx, "docker", exportArgs...)
	export.Stdout = &exported
	// Export output contains credentials. Never forward either stream to logs.
	if export.Run() != nil {
		return fail(errors.New("Provider credential export failed; existing credentials were preserved"))
	}
	target := filepath.Join(dir, "credential.json")
	if err = os.WriteFile(target, exported.Bytes(), 0600); err != nil {
		return fail(errors.New("Unable to stage private provider credential"))
	}
	if err = persistWorkerCredential(root, v.Provider, target); err != nil {
		return fail(err)
	}
	fmt.Fprintln(out, "Connection sign-in saved. Return to CommonRoom and choose Check sign-in.")
	return 0
}

// credentialExportOutput bounds sensitive output in memory; callers must never
// forward it to diagnostics. Overflow fails the export instead of truncating it.
type credentialExportOutput struct{ bytes.Buffer }

func (b *credentialExportOutput) Write(p []byte) (int, error) {
	if b.Len()+len(p) > 65536 {
		return 0, io.ErrShortBuffer
	}
	return b.Buffer.Write(p)
}

// workerCredentialExportArgs reads one known credential from the connection's
// login volume in a non-root, network-disabled container. The volume is read-only;
// there are no host paths, provider invocation, gateway, or Docker socket mounts.
// Source forks CAN weaken these limits, BUT risk exposing every host credential.
func workerCredentialExportArgs(image, volume, provider string) ([]string, error) {
	path := "/home/agent/.codex/auth.json"
	if provider == "claude" {
		path = "/home/agent/.claude/.credentials.json"
	} else if provider != "codex" {
		return nil, errors.New("Unsupported credential provider")
	}
	// O_NOFOLLOW also rejects symlink substitution between stat and open. Read
	// at most the limit plus one byte, even if the file grows after inspection.
	script := `const fs=require("node:fs");try{const fd=fs.openSync(process.argv[1],fs.constants.O_RDONLY|fs.constants.O_NOFOLLOW);const st=fs.fstatSync(fd);if(!st.isFile()||st.size<=2||st.size>65536)process.exit(78);const b=Buffer.alloc(65537);const n=fs.readSync(fd,b,0,b.length,0);fs.closeSync(fd);if(n<=2||n>65536)process.exit(78);JSON.parse(b.subarray(0,n).toString());process.stdout.write(b.subarray(0,n));}catch{process.exit(78);}`
	return []string{"run", "--rm", "--pull=never", "--network=none", "--read-only", "--user", "1000:1000", "--cap-drop=ALL", "--security-opt=no-new-privileges:true", "--pids-limit=32", "--memory=128m", "--memory-swap=128m", "--cpus=1", "--log-driver=none", "--workdir", "/", "--mount", "type=volume,src=" + volume + ",dst=/home/agent,readonly", "--entrypoint", "node", image, "-e", script, path}, nil
}

// persistWorkerCredential validates only the exported provider file and replaces
// the connection's private credential atomically. Invalid/missing exports leave
// the previous login untouched. No credential bytes are written to diagnostics.
func persistWorkerCredential(root, provider, source string) error {
	if provider != "codex" && provider != "claude" {
		return errors.New("Unsupported credential provider")
	}
	fi, err := os.Lstat(source)
	if err != nil || !fi.Mode().IsRegular() || fi.Size() > 65536 {
		return errors.New("Invalid provider credential file")
	}
	raw, err := os.ReadFile(source)
	if err != nil || len(raw) > 65536 || len(raw) <= 2 || !json.Valid(raw) {
		return errors.New("Invalid provider credential JSON")
	}
	if err = os.MkdirAll(root, 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(root, ".login-")
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
	if err = os.Rename(f.Name(), filepath.Join(root, providerCredentialName(provider))); err != nil {
		return err
	}
	return nil
}

// workerVerifyConnection runs only a fixed, tool-disabled conversation through
// the mandatory container path. No room history or user task is forwarded.
// Diagnostic volumes are disposable; credentials stay connection-scoped.
func workerVerifyConnection(v workerAuthRequest, out, diagnostics io.Writer) int {
	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	defer stop()
	ctx, cancel := context.WithTimeout(signalCtx, 20*time.Second)
	defer cancel()
	key := "diagnostic/" + id() + "/" + v.Connection + "/" + v.Provider
	digest := sha256.Sum256([]byte(key))
	volume := "commonroom-session-" + hex.EncodeToString(digest[:16])
	defer func() {
		cleanup, done := context.WithTimeout(context.Background(), 5*time.Second)
		defer done()
		if exec.CommandContext(cleanup, "docker", "volume", "rm", volume, volume+"-work").Run() != nil {
			fmt.Fprintln(diagnostics, "Diagnostic volume cleanup could not be confirmed")
		}
	}()
	var adapter sessionAdapter = codexAdapter{endpoint: v.Provider, executionKey: key}
	if v.Provider == "claude" {
		adapter = claudeAdapter{endpoint: v.Provider, executionKey: key}
	}
	answer, err := adapter.Run(ctx, "", "Connection check only. Do not use tools. Reply with exactly COMMONROOM_CONNECTION_OK.", func(string) error { return nil })
	json.NewEncoder(out).Encode(map[string]any{"version": 1, "connection": v.Connection, "provider": v.Provider, "verified": err == nil && strings.TrimSpace(answer) == "COMMONROOM_CONNECTION_OK"})
	return 0
}
