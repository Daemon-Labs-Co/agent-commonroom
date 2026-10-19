package main

import (
	"agentroom/internal/codexbridge"
	"agentroom/internal/strictjson"
	"agentroom/internal/toolbroker"
	"agentroom/internal/toolcommand"
	"bytes"
	"context"
	"encoding/json"
	"io"
)

// workerCodexToolRequest supplies authority only from trusted session binding.
// Model arguments never contain capability, scope, call ID, Docker options or
// provider settings. The dispatcher still authorizes path, operation and budget.
func workerCodexToolRequest(scope toolbroker.Scope, token, callID string, arguments json.RawMessage) (toolbroker.Request, error) {
	request := toolbroker.Request{Scope: scope, Capability: token, CallID: callID}
	if len(arguments) == 0 || len(arguments) > toolbroker.MaxRequestBytes || !strictjson.ValidEncoding(arguments) {
		return request, toolbroker.ErrRejected
	}
	fields := map[string]json.RawMessage{}
	decoder := json.NewDecoder(bytes.NewReader(arguments))
	start, err := decoder.Token()
	if err != nil || start != json.Delim('{') {
		return request, toolbroker.ErrRejected
	}
	for decoder.More() {
		key, err := decoder.Token()
		name, ok := key.(string)
		if err != nil || !ok {
			return request, toolbroker.ErrRejected
		}
		if _, duplicate := fields[name]; duplicate {
			return request, toolbroker.ErrRejected
		}
		var value json.RawMessage
		if decoder.Decode(&value) != nil {
			return request, toolbroker.ErrRejected
		}
		fields[name] = value
	}
	end, err := decoder.Token()
	if err != nil || end != json.Delim('}') {
		return request, toolbroker.ErrRejected
	}
	if _, err = decoder.Token(); err != io.EOF {
		return request, toolbroker.ErrRejected
	}
	for key, value := range fields {
		if string(bytes.TrimSpace(value)) == "null" {
			return request, toolbroker.ErrRejected
		}
		var target *string
		switch key {
		case "operation":
			target = &request.Operation
		case "path":
			target = &request.Path
		case "content":
			target = &request.Content
		case "command":
			if json.Unmarshal(value, &request.Command) != nil {
				return request, toolbroker.ErrRejected
			}
			continue
		default:
			return request, toolbroker.ErrRejected
		}
		if json.Unmarshal(value, target) != nil {
			return request, toolbroker.ErrRejected
		}
	}
	if request.Operation == "command" {
		if request.Path != "" || request.Content != "" || !toolcommand.Valid(request.Command) {
			return request, toolbroker.ErrRejected
		}
		return request, nil
	}
	if len(request.Command) != 0 {
		return request, toolbroker.ErrRejected
	}
	if (request.Operation != "read" && request.Operation != "write") || request.Path == "" || (request.Operation == "read" && request.Content != "") {
		return request, toolbroker.ErrRejected
	}
	return request, nil
}

// CodexTurn drives only the reviewed isolated app-server stream after startup.
// Engine process creation, controlled configuration and fencing are separate
// required boundaries. No codex exec, host tool, resume or API billing fallback
// is offered. A source fork CAN substitute stream/dispatcher, BUT loses isolation.
func (s *workerPairSupervisor) CodexTurn(ctx context.Context, session *toolbroker.ToolSession, token string, stream io.ReadWriteCloser, prompt, model string) (codexbridge.Result, error) {
	return s.codexTurn(ctx, session, token, stream, prompt, model, nil)
}

func (s *workerPairSupervisor) codexTurn(ctx context.Context, session *toolbroker.ToolSession, token string, stream io.ReadWriteCloser, prompt, model string, finalize func(context.Context) error) (codexbridge.Result, error) {
	s.mu.Lock()
	if ctx == nil || s.available(ctx) != nil || s.record.Status != "running" || s.record.Provider != "codex" || session == nil || session.Dispatcher == nil || stream == nil || !repositoryDigestPattern.MatchString(token) || session != s.toolSession || s.codexTurnStarted || !validWorkerAuditScope(s.record, session.Scope()) {
		s.mu.Unlock()
		return codexbridge.Result{}, toolbroker.ErrRejected
	}
	s.codexTurnStarted = true
	lifetime, cancel := context.WithDeadline(ctx, s.record.Deadline)
	stop := context.AfterFunc(s.ctx, cancel)
	s.mu.Unlock()
	defer cancel()
	defer stop()
	scope := session.Scope()
	result, err := codexbridge.RunWithOptions(lifetime, stream, prompt, model, func(ctx context.Context, callID string, arguments json.RawMessage) ([]byte, error) {
		request, err := workerCodexToolRequest(scope, token, callID, arguments)
		if err != nil {
			return nil, err
		}
		return session.Dispatcher.Dispatch(ctx, request)
	}, finalize, codexbridge.Options{Commands: session.CommandsEnabled()})
	if err != nil {
		s.cancel()
	}
	return result, err
}
