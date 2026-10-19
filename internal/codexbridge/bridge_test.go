package codexbridge

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func serverSetup(t *testing.T, conn net.Conn) *json.Decoder {
	return serverSetupCommands(t, conn, false)
}
func serverSetupCommands(t *testing.T, conn net.Conn, commands bool) *json.Decoder {
	t.Helper()
	decoder := json.NewDecoder(bufio.NewReader(conn))
	receive := func(method string) message {
		var m message
		if decoder.Decode(&m) != nil || m.Method != method {
			t.Errorf("expected %s, got %s", method, m.Method)
		}
		return m
	}
	send := func(v any) { json.NewEncoder(conn).Encode(v) }
	receive("initialize")
	send(map[string]any{"id": 1, "result": map[string]any{}})
	receive("initialized")
	request := receive("thread/start")
	var params struct {
		Sandbox, Approval, Cwd string
		DynamicTools           []struct {
			Name, Type  string
			InputSchema json.RawMessage
		}
		AllowProviderModelFallback bool
	}
	var raw map[string]json.RawMessage
	json.Unmarshal(request.Params, &raw)
	json.Unmarshal(raw["sandbox"], &params.Sandbox)
	json.Unmarshal(raw["approvalPolicy"], &params.Approval)
	json.Unmarshal(raw["cwd"], &params.Cwd)
	json.Unmarshal(raw["dynamicTools"], &params.DynamicTools)
	if params.Sandbox != "read-only" || params.Approval != "never" || params.Cwd != "/tmp/commonroom-engine" || len(params.DynamicTools) != 1 || params.DynamicTools[0].Name != "commonroom_tool" {
		t.Error("unrestricted thread configuration")
	}
	if len(params.DynamicTools) == 1 {
		advertised := strings.Contains(string(params.DynamicTools[0].InputSchema), `"command"`)
		if advertised != commands {
			t.Error("command schema did not match admitted capability")
		}
	}
	// Exercise notification ordering before the matching RPC response.
	send(map[string]any{"method": "thread/started", "params": map[string]any{"thread": map[string]string{"id": "thread-1"}}})
	send(map[string]any{"id": 2, "result": map[string]any{"thread": map[string]string{"id": "thread-1"}, "approvalPolicy": "never", "sandbox": map[string]string{"type": "readOnly"}, "cwd": "/tmp/commonroom-engine"}})
	receive("turn/start")
	send(map[string]any{"method": "turn/started", "params": map[string]any{"threadId": "thread-1", "turn": map[string]any{"id": "turn-1", "status": "inProgress", "items": []any{}}}})
	send(map[string]any{"id": 3, "result": map[string]any{"turn": map[string]any{"id": "turn-1", "status": "inProgress", "items": []any{}}}})
	return decoder
}
func callback() map[string]any {
	return map[string]any{"id": 100, "method": "item/tool/call", "params": map[string]any{"threadId": "thread-1", "turnId": "turn-1", "callId": "call-1", "tool": "commonroom_tool", "arguments": map[string]string{"operation": "read", "path": "file.go"}}}
}
func completed() map[string]any {
	return map[string]any{"method": "turn/completed", "params": map[string]any{"threadId": "thread-1", "turn": map[string]any{"id": "turn-1", "status": "completed", "items": []any{map[string]string{"type": "reasoning", "text": "not exported"}, map[string]string{"type": "agentMessage", "phase": "final_answer", "text": "Finished"}}}}}
}
func TestRunDrivesBoundedCallbackOnlyTurn(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	client, server := net.Pipe()
	defer server.Close()
	served := make(chan struct{})
	go func() {
		defer close(served)
		decoder := serverSetup(t, server)
		json.NewEncoder(server).Encode(callback())
		var reply message
		if decoder.Decode(&reply) != nil || string(reply.ID) != "100" || !strings.Contains(string(reply.Result), "audited result") {
			t.Error("missing correlated tool response")
		}
		json.NewEncoder(server).Encode(completed())
	}()
	calls := 0
	result, err := Run(ctx, client, "Synthetic request", "reviewed-model", func(_ context.Context, id string, args json.RawMessage) ([]byte, error) {
		calls++
		if id != "call-1" || !strings.Contains(string(args), "file.go") {
			t.Error("callback lost binding")
		}
		return []byte("audited result"), nil
	})
	if err != nil || result.ThreadID != "thread-1" || result.TurnID != "turn-1" || result.Text != "Finished" || calls != 1 {
		t.Fatal(result, err, calls)
	}
	<-served
}
func TestRunRejectsEscapesReplaysAndWrongCorrelation(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"native approval", "wrong thread", "wrong turn", "wrong tool", "namespace", "call replay", "request replay", "handler failure", "native completion", "failed completion", "foreign completion", "duplicate JSON", "oversize", "oversize result", "oversize final", "invalid result UTF8", "case alias"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
			defer cancel()
			client, server := net.Pipe()
			defer server.Close()
			done := make(chan struct{})
			go func() {
				defer close(done)
				decoder := serverSetup(t, server)
				request := callback()
				params := request["params"].(map[string]any)
				switch mode {
				case "case alias":
					request["Method"] = request["method"]
					delete(request, "method")
				case "native approval":
					request["method"] = "item/commandExecution/requestApproval"
				case "wrong thread":
					params["threadId"] = "foreign"
				case "wrong turn":
					params["turnId"] = "foreign"
				case "wrong tool":
					params["tool"] = "shell"
				case "namespace":
					params["namespace"] = "host"
				}
				switch mode {
				case "native completion", "failed completion", "foreign completion", "oversize final":
					event := completed()
					p := event["params"].(map[string]any)
					turn := p["turn"].(map[string]any)
					if mode == "oversize final" {
						turn["items"] = []any{map[string]string{"type": "agentMessage", "text": strings.Repeat("x", maxText+1)}}
					}
					if mode == "native completion" {
						turn["items"] = []any{map[string]string{"type": "commandExecution"}}
					}
					if mode == "failed completion" {
						turn["status"] = "failed"
					}
					if mode == "foreign completion" {
						p["threadId"] = "foreign"
					}
					json.NewEncoder(server).Encode(event)
				case "duplicate JSON":
					io.WriteString(server, "{\"id\":100,\"id\":101,\"method\":\"item/tool/call\",\"params\":{}}\n")
				case "oversize":
					io.WriteString(server, strings.Repeat("x", maxFrame+1)+"\n")
				default:
					json.NewEncoder(server).Encode(request)
					if mode == "call replay" || mode == "request replay" {
						var response message
						decoder.Decode(&response)
						if mode == "request replay" {
							params["callId"] = "call-2"
						} else {
							request["id"] = 101
						}
						json.NewEncoder(server).Encode(request)
					}
				}
			}()
			var calls atomic.Int32
			result, err := Run(ctx, client, "Synthetic", "reviewed-model", func(context.Context, string, json.RawMessage) ([]byte, error) {
				calls.Add(1)
				if mode == "oversize result" {
					return make([]byte, maxText+1), nil
				}
				if mode == "invalid result UTF8" {
					return []byte{255}, nil
				}
				if mode == "handler failure" {
					return nil, errors.New("synthetic audit refusal")
				}
				return []byte("ok"), nil
			})
			if err == nil || result.Text != "" {
				t.Fatal("unsafe protocol succeeded", mode, result)
			}
			want := int32(0)
			if mode == "call replay" || mode == "request replay" || mode == "handler failure" || mode == "oversize result" || mode == "invalid result UTF8" {
				want = 1
			}
			if calls.Load() != want {
				t.Fatal("unexpected dispatches", calls.Load(), want)
			}
			client.Close()
			<-done
		})
	}
}
func TestRunCancellationInterruptsProviderIO(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	client, server := net.Pipe()
	defer server.Close()
	done := make(chan error, 1)
	go func() {
		_, err := Run(ctx, client, "Synthetic", "model", func(context.Context, string, json.RawMessage) ([]byte, error) { return nil, nil })
		done <- err
	}()
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("cancelled protocol succeeded")
		}
	case <-time.After(time.Second):
		t.Fatal("cancel did not interrupt IO")
	}
}

func TestRunInvalidInputStillClosesOwnedStream(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"empty prompt", "invalid UTF8", "invalid model", "missing handler"} {
		t.Run(mode, func(t *testing.T) {
			client, peer := net.Pipe()
			defer peer.Close()
			prompt, model := "synthetic", "model"
			handler := Handler(func(context.Context, string, json.RawMessage) ([]byte, error) { return nil, nil })
			switch mode {
			case "empty prompt":
				prompt = ""
			case "invalid UTF8":
				prompt = string([]byte{255})
			case "invalid model":
				model = "--unsafe"
			case "missing handler":
				handler = nil
			}
			if _, err := Run(t.Context(), client, prompt, model, handler); err == nil {
				t.Fatal("invalid configuration accepted")
			}
			peer.SetReadDeadline(time.Now().Add(time.Second))
			if _, err := peer.Read(make([]byte, 1)); err != io.EOF {
				t.Fatal("invalid input leaked owned stream", err)
			}
		})
	}
}

func TestReadOnlySandboxRejectsAmbiguousOrBroadenedPolicy(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{"type":"readOnly","networkAccess":true}`, `{"type":"readOnly","networkAccess":null}`, `{"type":"readOnly","NetworkAccess":false}`, `{"type":"workspaceWrite"}`, `{"type":"externalSandbox"}`, `null`} {
		if readOnlySandbox([]byte(raw)) {
			t.Fatal("broader or ambiguous policy accepted", raw)
		}
	}
	if !readOnlySandbox([]byte(`{"type":"readOnly","networkAccess":false}`)) {
		t.Fatal("read-only policy refused")
	}
}

func TestRunRefusesCallbackBeforeTurnSubmission(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	client, peer := net.Pipe()
	defer peer.Close()
	done := make(chan struct{})
	go func() {
		defer close(done)
		var init message
		json.NewDecoder(peer).Decode(&init)
		json.NewEncoder(peer).Encode(callback())
		peer.SetReadDeadline(time.Now().Add(time.Second))
		if _, err := peer.Read(make([]byte, 1)); err != io.EOF {
			t.Error("premature callback was queued instead of refused", err)
		}
		peer.Close()
	}()
	_, err := Run(ctx, client, "synthetic", "model", func(context.Context, string, json.RawMessage) ([]byte, error) {
		t.Error("premature callback dispatched")
		return nil, nil
	})
	if err == nil {
		t.Fatal("callback before turn submission accepted")
	}
	<-done
}

type finalizerStream struct {
	net.Conn
	closed atomic.Bool
}

func (s *finalizerStream) Close() error { s.closed.Store(true); return s.Conn.Close() }
func TestFinalizerOwnsPostTurnBeforeStreamClose(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"success", "failure", "cancelled", "provider failure"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
			defer cancel()
			client, server := net.Pipe()
			defer server.Close()
			stream := &finalizerStream{Conn: client}
			served := make(chan struct{})
			go func() {
				defer close(served)
				serverSetup(t, server)
				if mode == "provider failure" {
					server.Close()
					return
				}
				json.NewEncoder(server).Encode(completed())
			}()
			calls := 0
			result, err := RunWithFinalizer(ctx, stream, "Synthetic request", "reviewed-model", func(context.Context, string, json.RawMessage) ([]byte, error) { return nil, nil }, func(ctx context.Context) error {
				calls++
				if stream.closed.Load() {
					t.Error("attachment closed before final credential capture")
				}
				if mode == "failure" {
					return errors.New("synthetic private error")
				}
				if mode == "cancelled" {
					cancel()
				}
				return nil
			})
			if !stream.closed.Load() {
				t.Fatal("attachment leaked")
			}
			if mode == "success" {
				if err != nil || result.Text != "Finished" || calls != 1 {
					t.Fatal("finalization failed", err)
				}
			} else if err == nil || result.Text != "" {
				t.Fatal("failed finalization released result")
			}
			if mode == "provider failure" && calls != 0 {
				t.Fatal("failed provider invoked finalizer")
			}
			<-served
		})
	}
}

func TestCommandSchemaRequiresTrustedOptions(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	client, server := net.Pipe()
	defer server.Close()
	done := make(chan struct{})
	go func() {
		defer close(done)
		serverSetupCommands(t, server, true)
		json.NewEncoder(server).Encode(completed())
	}()
	_, err := RunWithOptions(ctx, client, "Synthetic request", "reviewed-model", func(context.Context, string, json.RawMessage) ([]byte, error) {
		t.Error("unexpected callback")
		return nil, nil
	}, nil, Options{Commands: true})
	if err != nil {
		t.Fatal(err)
	}
	<-done
}
