// Package codexbridge drives the reviewed bridge-only Codex app-server protocol
// from rust-v0.162.1. It launches no process and grants no host tool authority.
// The worker must supply an isolated reviewed engine stream, fresh controlled
// configuration and a handler backed only by the audited credential-free broker.
package codexbridge

import (
	"agentroom/internal/strictjson"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"strings"
	"unicode/utf8"
)

const maxFrame = 256 * 1024
const maxStream = 8 * 1024 * 1024
const maxText = 16 * 1024

var ErrProtocol = errors.New("reviewed Codex bridge refused or interrupted; do not replay")
var identifier = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)
var modelName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$`)

type Result struct{ ThreadID, TurnID, Text string }

// Handler receives no capability or launch options from the model. The trusted
// caller binds scope/capability and dispatches through the original ToolSession.
// A source fork CAN supply a host handler, BUT violates mandatory isolation.
type Handler func(context.Context, string, json.RawMessage) ([]byte, error)
type message struct {
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   json.RawMessage `json:"error,omitempty"`
	JSONRPC string          `json:"jsonrpc,omitempty"`
}
type driver struct {
	ctx           context.Context
	stream        io.ReadWriteCloser
	reader        *bufio.Reader
	bytes, frames int
	pending       []message
	result        Result
	seen          map[string]bool
	requests      map[string]bool
	handler       Handler
}

// Run owns one fresh thread and one turn, never resume/fork/config/review methods.
// Close MUST interrupt both stream directions; container shutdown remains the
// supervisor's responsibility. Cancellation closes the stream without replay.
func Run(ctx context.Context, stream io.ReadWriteCloser, prompt, model string, handler Handler) (Result, error) {
	return RunWithFinalizer(ctx, stream, prompt, model, handler, nil)
}

// RunWithFinalizer gives trusted worker custody one bounded post-turn operation
// before closing the attachment. It runs only after validated completion, never
// on a provider failure. No model selects this callback. Failure withholds text;
// cancellation still interrupts both stream directions and supervisor fencing
// remains mandatory. The finalizer must honor ctx and keep credentials private.
func RunWithFinalizer(ctx context.Context, stream io.ReadWriteCloser, prompt, model string, handler Handler, finalize func(context.Context) error) (Result, error) {
	if stream == nil {
		return Result{}, ErrProtocol
	}
	defer stream.Close()
	if ctx == nil || ctx.Err() != nil || handler == nil || len(prompt) == 0 || len(prompt) > 64*1024 || !utf8.ValidString(prompt) || !modelName.MatchString(model) {
		return Result{}, ErrProtocol
	}
	stop := context.AfterFunc(ctx, func() { stream.Close() })
	defer stop()
	d := &driver{ctx: ctx, stream: stream, reader: bufio.NewReaderSize(stream, maxFrame+1), seen: map[string]bool{}, requests: map[string]bool{}, handler: handler}
	if d.send(map[string]any{"id": 1, "method": "initialize", "params": map[string]any{"clientInfo": map[string]string{"name": "commonroom", "version": "0.1"}, "capabilities": map[string]bool{"experimentalApi": true}}}) != nil {
		return Result{}, ErrProtocol
	}
	if _, err := d.response("1"); err != nil {
		return Result{}, ErrProtocol
	}
	if d.send(map[string]any{"method": "initialized", "params": map[string]any{}}) != nil {
		return Result{}, ErrProtocol
	}
	schema := map[string]any{"type": "object", "additionalProperties": false, "required": []string{"operation", "path"}, "properties": map[string]any{"operation": map[string]any{"type": "string", "enum": []string{"read", "write"}}, "path": map[string]string{"type": "string"}, "content": map[string]string{"type": "string"}}}
	params := map[string]any{"model": model, "cwd": "/tmp/commonroom-engine", "sandbox": "read-only", "approvalPolicy": "never", "ephemeral": true, "allowProviderModelFallback": false, "dynamicTools": []any{map[string]any{"type": "function", "name": "commonroom_tool", "deferLoading": false, "description": "Read or write approved isolated workspace files through the audited broker.", "inputSchema": schema}}}
	if d.send(map[string]any{"id": 2, "method": "thread/start", "params": params}) != nil {
		return Result{}, ErrProtocol
	}
	thread, err := d.response("2")
	if err != nil {
		return Result{}, ErrProtocol
	}
	var started struct {
		Thread struct {
			ID string `json:"id"`
		} `json:"thread"`
		Approval string          `json:"approvalPolicy"`
		Sandbox  json.RawMessage `json:"sandbox"`
		Cwd      string          `json:"cwd"`
	}
	if json.Unmarshal(thread, &started) != nil || !identifier.MatchString(started.Thread.ID) || started.Approval != "never" || !readOnlySandbox(started.Sandbox) || started.Cwd != "/tmp/commonroom-engine" {
		return Result{}, ErrProtocol
	}
	d.result.ThreadID = started.Thread.ID
	if d.send(map[string]any{"id": 3, "method": "turn/start", "params": map[string]any{"threadId": d.result.ThreadID, "input": []any{map[string]any{"type": "text", "text": prompt}}, "approvalPolicy": "never"}}) != nil {
		return Result{}, ErrProtocol
	}
	turn, err := d.response("3")
	if err != nil {
		return Result{}, ErrProtocol
	}
	var accepted struct {
		Turn struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"turn"`
	}
	if json.Unmarshal(turn, &accepted) != nil || !identifier.MatchString(accepted.Turn.ID) || accepted.Turn.Status != "inProgress" {
		return Result{}, ErrProtocol
	}
	d.result.TurnID = accepted.Turn.ID
	for {
		var m message
		if len(d.pending) > 0 {
			m = d.pending[0]
			d.pending = d.pending[1:]
		} else {
			m, err = d.read()
			if err != nil {
				return Result{}, ErrProtocol
			}
		}
		done, err := d.handle(m)
		if err != nil || ctx.Err() != nil {
			return Result{}, ErrProtocol
		}
		if done {
			if finalize != nil && finalize(ctx) != nil {
				return Result{}, ErrProtocol
			}
			if ctx.Err() != nil {
				return Result{}, ErrProtocol
			}
			return d.result, nil
		}
	}
}

func (d *driver) send(v any) error {
	raw, err := json.Marshal(v)
	if err != nil || len(raw) > maxFrame || d.ctx.Err() != nil {
		return ErrProtocol
	}
	raw = append(raw, '\n')
	n, err := d.stream.Write(raw)
	if err != nil || n != len(raw) || d.ctx.Err() != nil {
		return ErrProtocol
	}
	return nil
}
func (d *driver) read() (message, error) {
	var m message
	raw, err := d.reader.ReadSlice('\n')
	d.bytes += len(raw)
	d.frames++
	if err != nil || len(raw) > maxFrame || d.bytes > maxStream || d.frames > 2048 || !strictjson.ValidEncoding(raw) || !uniqueJSON(raw) {
		return m, ErrProtocol
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if !exactObject(raw, []string{"id", "method", "params", "result", "error", "jsonrpc"}, nil) || decoder.Decode(&m) != nil || (m.JSONRPC != "" && m.JSONRPC != "2.0") || bytes.Equal(m.ID, []byte("null")) || len(m.Error) > 0 {
		return m, ErrProtocol
	}
	if m.Method != "" {
		if len(m.Params) == 0 || len(m.Result) > 0 {
			return m, ErrProtocol
		}
	} else if len(m.ID) == 0 || len(m.Result) == 0 || len(m.Params) > 0 {
		return m, ErrProtocol
	}
	return m, nil
}
func (d *driver) response(id string) (json.RawMessage, error) {
	for {
		m, err := d.read()
		if err != nil {
			return nil, err
		}
		if m.Method == "" {
			if string(m.ID) != id {
				return nil, ErrProtocol
			}
			return m.Result, nil
		}
		if len(m.ID) > 0 && (m.Method != "item/tool/call" || id != "3") {
			return nil, ErrProtocol
		}
		if len(d.pending) >= 16 {
			return nil, ErrProtocol
		}
		d.pending = append(d.pending, m)
	}
}
func (d *driver) handle(m message) (bool, error) {
	if m.Method == "item/tool/call" {
		if len(m.ID) == 0 || !validRPCID(m.ID) || d.requests[string(m.ID)] || len(d.seen) >= 64 {
			return false, ErrProtocol
		}
		var call struct {
			Arguments json.RawMessage `json:"arguments"`
			CallID    string          `json:"callId"`
			Namespace *string         `json:"namespace"`
			ThreadID  string          `json:"threadId"`
			Tool      string          `json:"tool"`
			TurnID    string          `json:"turnId"`
		}
		decoder := json.NewDecoder(bytes.NewReader(m.Params))
		decoder.DisallowUnknownFields()
		if !exactObject(m.Params, []string{"arguments", "callId", "namespace", "threadId", "tool", "turnId"}, []string{"arguments", "callId", "threadId", "tool", "turnId"}) || decoder.Decode(&call) != nil || call.Namespace != nil || call.ThreadID != d.result.ThreadID || call.TurnID != d.result.TurnID || call.Tool != "commonroom_tool" || !identifier.MatchString(call.CallID) || d.seen[call.CallID] || len(call.Arguments) == 0 || len(call.Arguments) > 64*1024 || call.Arguments[0] != '{' {
			return false, ErrProtocol
		}
		d.seen[call.CallID] = true
		d.requests[string(m.ID)] = true
		output, err := d.handler(d.ctx, call.CallID, append(json.RawMessage(nil), call.Arguments...))
		if err != nil || len(output) > maxText || !utf8.Valid(output) || d.ctx.Err() != nil {
			return false, ErrProtocol
		}
		if d.send(map[string]any{"id": m.ID, "result": map[string]any{"success": true, "contentItems": []any{map[string]string{"type": "inputText", "text": string(output)}}}}) != nil {
			return false, ErrProtocol
		}
		return false, nil
	}
	if len(m.ID) > 0 {
		return false, ErrProtocol
	}
	switch m.Method {
	case "thread/started":
		var p struct {
			Thread struct {
				ID string `json:"id"`
			} `json:"thread"`
		}
		if json.Unmarshal(m.Params, &p) != nil || p.Thread.ID != d.result.ThreadID {
			return false, ErrProtocol
		}
	case "turn/started", "turn/completed":
		var p struct {
			ThreadID string `json:"threadId"`
			Turn     struct {
				ID     string          `json:"id"`
				Status string          `json:"status"`
				Error  json.RawMessage `json:"error"`
				Items  []struct {
					Type  string `json:"type"`
					Text  string `json:"text"`
					Phase string `json:"phase"`
				} `json:"items"`
			} `json:"turn"`
		}
		if json.Unmarshal(m.Params, &p) != nil || p.ThreadID != d.result.ThreadID || p.Turn.ID != d.result.TurnID {
			return false, ErrProtocol
		}
		if m.Method == "turn/started" {
			if p.Turn.Status != "inProgress" {
				return false, ErrProtocol
			}
			return false, nil
		}
		if p.Turn.Status != "completed" || (len(p.Turn.Error) > 0 && string(p.Turn.Error) != "null") {
			return false, ErrProtocol
		}
		for _, item := range p.Turn.Items {
			switch item.Type {
			case "userMessage", "agentMessage", "reasoning", "dynamicToolCall", "plan":
			default:
				return false, ErrProtocol
			}

			if item.Type == "agentMessage" && (item.Phase == "final_answer" || item.Phase == "") {
				if len(d.result.Text)+len(item.Text) > maxText {
					return false, ErrProtocol
				}
				d.result.Text += item.Text
			}
		}
		return true, nil
	case "item/started", "item/completed", "item/agentMessage/delta", "item/reasoning/summaryTextDelta", "item/reasoning/textDelta", "thread/tokenUsage/updated":
		var p struct {
			ThreadID string `json:"threadId"`
			TurnID   string `json:"turnId"`
		}
		if json.Unmarshal(m.Params, &p) != nil || p.ThreadID != d.result.ThreadID || p.TurnID != d.result.TurnID {
			return false, ErrProtocol
		}
		if m.Method == "item/started" || m.Method == "item/completed" {
			var item struct {
				Item struct {
					Type string `json:"type"`
				} `json:"item"`
			}
			if json.Unmarshal(m.Params, &item) != nil {
				return false, ErrProtocol
			}
			switch item.Item.Type {
			case "userMessage", "agentMessage", "reasoning", "dynamicToolCall", "plan":
			default:
				return false, ErrProtocol
			}
		}
	// Neither reasoning nor model descriptions of observed checks are exported.
	default:
		return false, ErrProtocol
	}
	return false, nil
}
func validRPCID(raw []byte) bool {
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return identifier.MatchString(text)
	}
	var n json.Number
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if decoder.Decode(&n) != nil {
		return false
	}
	_, err := n.Int64()
	return err == nil
}

// Reject duplicate keys at every depth before decoding typed protocol fields.
func uniqueJSON(raw []byte) bool {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	var value func(int) bool
	value = func(depth int) bool {
		if depth > 32 {
			return false
		}
		token, err := decoder.Token()
		if err != nil {
			return false
		}
		delim, ok := token.(json.Delim)
		if !ok {
			return true
		}
		if delim == '{' {
			seen := map[string]bool{}
			for decoder.More() {
				key, err := decoder.Token()
				name, ok := key.(string)
				if err != nil || !ok || seen[strings.ToLower(name)] {
					return false
				}
				seen[strings.ToLower(name)] = true
				if !value(depth + 1) {
					return false
				}
			}
			end, err := decoder.Token()
			return err == nil && end == json.Delim('}')
		}
		if delim == '[' {
			for decoder.More() {
				if !value(depth + 1) {
					return false
				}
			}
			end, err := decoder.Token()
			return err == nil && end == json.Delim(']')
		}
		return false
	}
	if !value(0) {
		return false
	}
	_, err := decoder.Token()
	return err == io.EOF
}

func exactObject(raw []byte, allowed, required []string) bool {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return false
	}
	for key := range fields {
		found := false
		for _, name := range allowed {
			if key == name {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	for _, name := range required {
		if _, ok := fields[name]; !ok {
			return false
		}
	}
	return true
}

func readOnlySandbox(raw []byte) bool {
	if !exactObject(raw, []string{"type", "networkAccess"}, []string{"type"}) {
		return false
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil {
		return false
	}
	var kind string
	if json.Unmarshal(fields["type"], &kind) != nil || kind != "readOnly" {
		return false
	}
	if network, exists := fields["networkAccess"]; exists {
		var enabled bool
		if string(bytes.TrimSpace(network)) == "null" || json.Unmarshal(network, &enabled) != nil || enabled {
			return false
		}
	}
	return true
}
