package main

// Synthetic Docker/engine protocol peers exercise real worker assembly, session,
// runtime inspection, audit framing, callback dispatch and cleanup together.
// No real Docker, provider, database or credentials are used.
import (
	"agentroom/internal/toolbroker"
	"agentroom/internal/workspacebootstrap"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

const syntheticRefreshedCredential = `{"auth_mode":"chatgpt","OPENAI_API_KEY":null,"tokens":{"access_token":"synthetic-new-token","refresh_token":"synthetic-new-refresh"}}`

const syntheticDockerRoot = "/home/synthetic/.local/share/docker"

type codexIntegrationDocker struct {
	containers                        map[string]map[string]any
	volume                            map[string]any
	scope                             toolbroker.Scope
	engineDone                        chan error
	audit                             *transportTestAuditor
	creates, starts, helpers, removed int
	bootstrapped                      bool
	mode                              string
}

type codexIntegrationAuditor struct {
	counter *transportTestAuditor
	mode    string
}

func (a codexIntegrationAuditor) RecordIntent(ctx context.Context, intent toolbroker.Intent) error {
	a.counter.RecordIntent(ctx, intent)
	if a.mode == "intent failure" {
		return toolbroker.ErrAudit
	}
	return nil
}
func (a codexIntegrationAuditor) RecordOutcome(ctx context.Context, intent toolbroker.Intent, outcome toolbroker.Outcome) error {
	a.counter.RecordOutcome(ctx, intent, outcome)
	if a.mode == "outcome failure" {
		return toolbroker.ErrAudit
	}
	return nil
}

func fixtureFlag(args []string, key string) string {
	for i, v := range args {
		if v == key && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}
func fixtureLabels(args []string) map[string]string {
	labels := map[string]string{}
	for i, v := range args {
		if v == "--label" {
			k, value, _ := strings.Cut(args[i+1], "=")
			labels[k] = value
		}
	}
	return labels
}
func (f *codexIntegrationDocker) Run(ctx context.Context, args []string, input []byte, out io.Writer) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	write := func(v string) error { _, err := io.WriteString(out, v); return err }
	encode := func(v any) error { return json.NewEncoder(out).Encode(v) }
	switch args[0] {
	case "info":
		return write(syntheticDockerRoot)
	case "volume":
		switch args[1] {
		case "ls":
			if f.volume != nil {
				return write(f.volume["Name"].(string))
			}
			return nil
		case "create":
			name := args[len(args)-1]
			f.volume = map[string]any{"Name": name, "Driver": "local", "Scope": "local", "Mountpoint": path.Join(syntheticDockerRoot, "volumes", name, "_data"), "CreatedAt": time.Now().UTC().Format(time.RFC3339Nano), "Labels": fixtureLabels(args)}
			return write(name)
		case "inspect":
			if f.volume == nil {
				return errors.New("synthetic volume missing")
			}
			return encode(f.volume)
		default:
			return errors.New("workspace disposal was not authorized")
		}
	case "ps":
		filters := []string{}
		for i, v := range args {
			if v == "--filter" {
				filters = append(filters, args[i+1])
			}
		}
		ids := []string{}
		for id, c := range f.containers {
			match := true
			for _, filter := range filters {
				switch {
				case strings.HasPrefix(filter, "name=^/"):
					match = match && c["Name"] == "/"+strings.TrimSuffix(strings.TrimPrefix(filter, "name=^/"), "$")
				case strings.HasPrefix(filter, "id="):
					match = match && id == strings.TrimPrefix(filter, "id=")
				case strings.HasPrefix(filter, "volume="):
					match = match && c["Config"].(map[string]any)["Labels"].(map[string]string)["commonroom.execution"] == "separated-tool-v1"
				case strings.HasPrefix(filter, "label="):
					k, v, _ := strings.Cut(strings.TrimPrefix(filter, "label="), "=")
					match = match && c["Config"].(map[string]any)["Labels"].(map[string]string)[k] == v
				default:
					return errors.New("unsupported synthetic inventory filter")
				}
			}
			if match {
				ids = append(ids, id)
			}
		}
		slices.Sort(ids)
		return write(strings.Join(ids, "\n"))
	case "image":
		return encode(map[string]any{"Id": "sha256:" + strings.Repeat("c", 64), "Config": map[string]any{"User": "1000:1000", "WorkingDir": "/tmp", "Entrypoint": []string{"/usr/local/bin/node", "/opt/commonroom/engine-entrypoint.cjs"}}})
	case "create":
		f.creates++
		id := strings.Repeat("d", 64)
		if strings.Contains(fixtureFlag(args, "--name"), "engine") {
			id = strings.Repeat("e", 64)
		}
		if f.containers[id] != nil {
			return errors.New("synthetic identity reused")
		}
		env := []string{}
		mounts := []map[string]any{}
		for i, arg := range args {
			if arg == "--env" {
				env = append(env, args[i+1])
			}
			if arg == "--mount" {
				fields := map[string]string{}
				for _, part := range strings.Split(args[i+1], ",") {
					k, v, _ := strings.Cut(part, "=")
					fields[k] = v
				}
				mount := map[string]any{"Type": fields["type"], "Destination": fields["dst"], "RW": fields["readonly"] == "" && !strings.Contains(args[i+1], ",readonly"), "Propagation": "rprivate"}
				if fields["type"] == "volume" {
					mount["Name"] = fields["src"]
					mount["Source"] = path.Join(syntheticDockerRoot, "volumes", fields["src"], "_data")
					mount["Driver"] = "local"
				} else {
					mount["Source"] = fields["src"]
				}
				mounts = append(mounts, mount)
			}
		}
		entry := slices.Index(args, "--entrypoint")
		if entry < 0 || len(args) < entry+3 {
			return errors.New("missing synthetic entrypoint")
		}
		config := map[string]any{"Image": args[entry+2], "User": fixtureFlag(args, "--user"), "WorkingDir": fixtureFlag(args, "--workdir"), "Entrypoint": []string{args[entry+1]}, "Cmd": args[entry+3:], "Env": env, "Labels": fixtureLabels(args), "Healthcheck": map[string]any{"Test": []string{"NONE"}}, "Tty": false, "OpenStdin": slices.Contains(args, "-i"), "StdinOnce": false}
		f.containers[id] = map[string]any{"Id": id, "Name": "/" + fixtureFlag(args, "--name"), "Image": "sha256:" + strings.Repeat("c", 64), "RestartCount": 0, "Config": config, "State": map[string]any{"Running": false, "Paused": false, "Restarting": false, "Dead": false, "Status": "created", "Error": ""}, "Mounts": mounts, "HostConfig": map[string]any{"Privileged": false, "ReadonlyRootfs": true, "Init": true, "AutoRemove": false, "RestartPolicy": map[string]any{"Name": "no"}, "NetworkMode": "none", "IpcMode": "private", "CgroupnsMode": "private", "Memory": int64(536870912), "MemorySwap": int64(536870912), "NanoCpus": int64(500000000), "PidsLimit": int64(64), "CapDrop": []string{"ALL"}, "SecurityOpt": []string{"no-new-privileges:true"}, "Tmpfs": map[string]string{"/tmp": "rw,nosuid,nodev,noexec,size=64m,mode=1777"}, "LogConfig": map[string]any{"Type": "none"}, "Ulimits": []map[string]any{{"Name": "nofile", "Soft": 512, "Hard": 512}}}}
		return write(id)
	case "inspect":
		c := f.containers[args[len(args)-1]]
		if c == nil {
			return errors.New("synthetic container missing")
		}
		if strings.Contains(fixtureFlag(args, "--format"), "|") {
			labels := c["Config"].(map[string]any)["Labels"].(map[string]string)
			return write(strings.Join([]string{c["Id"].(string), c["Name"].(string), labels["commonroom.managed"], labels["commonroom.job"], labels["commonroom.execution"], labels["commonroom.repository"]}, "|"))
		}
		return encode(c)
	case "start":
		c := f.containers[args[1]]
		if c == nil {
			return errors.New("synthetic container missing")
		}
		if strings.Contains(c["Name"].(string), "engine") && !f.bootstrapped {
			return errors.New("engine started before bootstrap")
		}
		f.starts++
		state := c["State"].(map[string]any)
		state["Running"] = true
		state["Status"] = "running"
		return write(args[1])
	case "rm":
		delete(f.containers, args[len(args)-1])
		f.removed++
		return nil
	case "exec":
		if len(args) == 9 && args[6] == "/usr/local/bin/node" && args[7] == "-e" {
			if args[5] != strings.Repeat("e", 64) || !strings.Contains(args[8], "COMMONROOM-CREDENTIAL-EXPORT-V1") {
				return errors.New("wrong private export binding")
			}
			if f.mode == "refresh export failure" {
				return errors.New("synthetic export failure")
			}
			if f.mode == "invalid refresh" {
				return write(`{"auth_mode":"apikey","OPENAI_API_KEY":"synthetic-not-a-key"}`)
			}
			if f.mode == "foreign refresh" {
				return write(`{"auth_mode":"chatgpt","tokens":{"access_token":"synthetic-only","account_id":"foreign"}}`)
			}
			return write(syntheticRefreshedCredential)
		}
		if args[len(args)-1] != "/opt/commonroom/tool-call.cjs" || args[len(args)-3] != strings.Repeat("d", 64) || f.audit.calls.Load() != 1 || !f.bootstrapped {
			return errors.New("helper bypassed audited tool identity")
		}
		f.helpers++
		data := []byte("synthetic source")
		sum := sha256.Sum256(data)
		return encode(map[string]any{"operation": "read", "bytes": len(data), "sha256": hex.EncodeToString(sum[:]), "encoding": "base64", "data": base64.StdEncoding.EncodeToString(data)})
	}
	return fmt.Errorf("unsupported synthetic Docker operation %s", args[0])
}
func (f *codexIntegrationDocker) RunInput(ctx context.Context, args []string, input io.Reader, out io.Writer) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if len(args) != 10 || args[6] != strings.Repeat("d", 64) || args[7] != "/opt/commonroom/workspace-bootstrap" || args[8] != f.scope.JobID || args[9] != f.scope.RepositoryDigest {
		return errors.New("bootstrap escaped scope")
	}
	raw, err := io.ReadAll(input)
	if err != nil || string(raw) != "synthetic approved bundle" {
		return errors.New("snapshot mismatch")
	}
	f.bootstrapped = true
	return json.NewEncoder(out).Encode(workspacebootstrap.Receipt{Version: 1, Job: f.scope.JobID, Repository: f.scope.RepositoryDigest, Commit: strings.Repeat("a", 40), Branch: "feature/commonroom-" + f.scope.JobID})
}
func (f *codexIntegrationDocker) OpenEngine(ctx context.Context, id string) (io.ReadWriteCloser, error) {
	if ctx.Err() != nil || id != strings.Repeat("e", 64) || !f.bootstrapped {
		return nil, errors.New("unprepared synthetic engine")
	}
	client, peer := net.Pipe()
	go func() {
		defer peer.Close()
		peer.SetDeadline(time.Now().Add(3 * time.Second))
		f.engineDone <- serveSyntheticCodex(peer, f.mode)
	}()
	return client, nil
}
func serveSyntheticCodex(stream net.Conn, mode string) error {
	magic := make([]byte, len("COMMONROOM-AUTH-V1\n"))
	if _, err := io.ReadFull(stream, magic); err != nil || string(magic) != "COMMONROOM-AUTH-V1\n" {
		return errors.New("credential magic mismatch")
	}
	size := make([]byte, 4)
	if _, err := io.ReadFull(stream, size); err != nil {
		return err
	}
	if binary.BigEndian.Uint32(size) != uint32(len(syntheticEngineCredential)) {
		return errors.New("credential bound mismatch")
	}
	raw := make([]byte, len(syntheticEngineCredential))
	if _, err := io.ReadFull(stream, raw); err != nil || string(raw) != syntheticEngineCredential {
		return errors.New("private provisioning mismatch")
	}
	decoder := json.NewDecoder(stream)
	encoder := json.NewEncoder(stream)
	receive := func(method string) error {
		var msg struct{ Method string }
		if err := decoder.Decode(&msg); err != nil {
			return err
		}
		if msg.Method != method {
			return fmt.Errorf("expected %s got %s", method, msg.Method)
		}
		return nil
	}
	send := func(v any) error { return encoder.Encode(v) }
	if err := receive("initialize"); err != nil {
		return err
	}
	if err := send(map[string]any{"id": 1, "result": map[string]any{}}); err != nil {
		return err
	}
	if err := receive("initialized"); err != nil {
		return err
	}
	if err := receive("thread/start"); err != nil {
		return err
	}
	if err := send(map[string]any{"id": 2, "result": map[string]any{"thread": map[string]string{"id": "synthetic-thread"}, "approvalPolicy": "never", "sandbox": map[string]string{"type": "readOnly"}, "cwd": "/tmp/commonroom-engine"}}); err != nil {
		return err
	}
	if err := receive("turn/start"); err != nil {
		return err
	}
	if mode == "engine drop" {
		return nil
	}
	if err := send(map[string]any{"id": 3, "result": map[string]any{"turn": map[string]any{"id": "synthetic-turn", "status": "inProgress", "items": []any{}}}}); err != nil {
		return err
	}
	operation := "read"
	if mode == "forbidden callback" {
		operation = "command"
	}
	if err := send(map[string]any{"id": 100, "method": "item/tool/call", "params": map[string]any{"threadId": "synthetic-thread", "turnId": "synthetic-turn", "callId": "synthetic-call", "tool": "commonroom_tool", "arguments": map[string]string{"operation": operation, "path": "file.go"}}}); err != nil {
		return err
	}
	var reply map[string]json.RawMessage
	if err := decoder.Decode(&reply); err != nil {
		if mode != "success" {
			return nil
		}
		return err
	}
	if string(reply["id"]) != "100" || !strings.Contains(string(reply["result"]), "base64") {
		return errors.New("callback result missing")
	}
	return send(map[string]any{"method": "turn/completed", "params": map[string]any{"threadId": "synthetic-thread", "turn": map[string]any{"id": "synthetic-turn", "status": "completed", "items": []any{map[string]string{"type": "agentMessage", "phase": "final_answer", "text": "Synthetic turn completed"}}}}})
}

func TestOwnedCodexSyntheticCallbackAndFencedRetention(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"success", "forbidden callback", "intent failure", "outcome failure", "engine drop", "refresh export failure", "invalid refresh", "foreign refresh"} {
		t.Run(mode, func(t *testing.T) {
			_, s, _ := workspacePair(t)
			r := s.record
			scope := toolbroker.Scope{Version: 1, JobID: r.JobID, Room: r.RoomID, Participant: r.ParticipantID, Connection: r.ConnectionID, Provider: r.Provider, RepositoryDigest: r.Pair.Repository, Deadline: r.Deadline}
			auditor := &transportTestAuditor{}
			runner := &codexIntegrationDocker{containers: map[string]map[string]any{}, scope: scope, engineDone: make(chan error, 1), audit: auditor, mode: mode}
			client, peer := net.Pipe()
			defer peer.Close()
			auditDone := make(chan error, 1)
			go func() { auditDone <- toolbroker.ServeAudit(s.ctx, peer, scope, codexIntegrationAuditor{auditor, mode}) }()
			ctx := context.WithValue(s.ctx, workerLocalJobContextKey{}, scope.JobID)
			session, token, err := newWorkerToolSession(ctx, scope, []string{"read"}, "ghcr.io/example/tool@sha256:"+strings.Repeat("a", 64), syntheticDockerRoot, time.Now(), client, runner)
			if err != nil {
				t.Fatal(err)
			}
			defer session.Close()
			auth := t.TempDir()
			private := filepath.Join(auth, connectionAuthKey(scope.Connection), scope.Provider)
			if err = os.MkdirAll(private, 0700); err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(filepath.Join(private, "auth.json"), []byte(syntheticEngineCredential), 0600); err != nil {
				t.Fatal(err)
			}
			gateway := shortGatewayRoot(t)
			bundle := &pairBundleFixture{Reader: strings.NewReader("synthetic approved bundle")}
			result, record, err := s.RunCodex(t.Context(), session, token, bundle, workerCodexRunConfig{EngineImage: "ghcr.io/example/engine@sha256:" + strings.Repeat("b", 64), AuthRoot: auth, GatewayRoot: gateway, Prompt: "Synthetic request", Model: "reviewed-model"}, runner)
			if err == nil || record.ShutdownConfirmed || record.Pair.Workspace == nil || record.Pair.Workspace.Removed || runner.volume == nil {
				t.Fatalf("unapproved workspace disposal or false lifecycle success: %v %+v", err, record)
			}
			if !record.Pair.Engine.Absent || !record.Pair.Tool.Absent || len(runner.containers) != 0 || runner.removed != 2 {
				t.Fatalf("roles not fenced: %v %+v creates=%d starts=%d removed=%d", err, record, runner.creates, runner.starts, runner.removed)
			}
			if runner.creates != 2 || runner.starts != 2 || !runner.bootstrapped || bundle.closes.Load() != 1 {
				t.Fatal("assembly did not execute exactly once", runner.creates, runner.starts, bundle.closes.Load())
			}
			if mode == "success" {
				if result.Text != "Synthetic turn completed" || auditor.calls.Load() != 2 || runner.helpers != 1 {
					t.Fatalf("callback pipeline incomplete: result=%+v error=%v audit=%d helper=%d", result, err, auditor.calls.Load(), runner.helpers)
				}
			} else {
				wantAudit, wantHelpers := int32(0), 0
				if mode == "intent failure" {
					wantAudit = 1
				}
				if mode == "outcome failure" || mode == "refresh export failure" || mode == "invalid refresh" || mode == "foreign refresh" {
					wantAudit = 2
					wantHelpers = 1
				}
				if result.Text != "" || auditor.calls.Load() != wantAudit || runner.helpers != wantHelpers {
					t.Fatalf("failed turn leaked results or bypassed audit: mode=%s result=%+v audit=%d helpers=%d", mode, result, auditor.calls.Load(), runner.helpers)
				}
			}
			select {
			case engineErr := <-runner.engineDone:
				if engineErr != nil {
					t.Fatal(engineErr)
				}
			case <-time.After(time.Second):
				t.Fatal("synthetic engine peer stuck")
			}
			awaitAuditResult(t, auditDone)
			if s.engineCredential != nil {
				t.Fatal("fenced engine retained credential lease")
			}
			ready := workerCredentialReady(private) && credentialPresent(filepath.Join(private, "auth.json"))
			if ready != (mode == "success") {
				t.Fatal("credential custody did not match observed refresh outcome", mode)
			}
			saved, readErr := readPrivateBounded(filepath.Join(private, "auth.json"), 65536)
			if readErr != nil {
				t.Fatal(readErr)
			}
			expected := syntheticEngineCredential
			if mode == "success" {
				expected = syntheticRefreshedCredential
			}
			if string(saved) != expected {
				t.Fatal("incorrect connection credential persisted")
			}
			unlock, err := lockWorkerCredential(private)
			if err != nil {
				t.Fatal(err)
			}
			unlock()
			if entries, err := os.ReadDir(gateway); err != nil || len(entries) != 0 {
				t.Fatal("private gateway leaked", err)
			}
			if _, _, err = s.RunCodex(t.Context(), session, token, bundle, workerCodexRunConfig{}, runner); err == nil || runner.creates != 2 {
				t.Fatal("completed job resumed")
			}
		})
	}
}
