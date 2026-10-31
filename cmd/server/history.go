package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os/exec"
	"time"
)

// historyRPC uses the supported app-server protocol without starting model turns.
// This is a current host-side history utility, not an isolated agent worker.
// Under the mandatory-isolation design, human-authorized history import must be
// a separate trusted capability; agents must not inherit access to the user's
// whole session store. Imported context never authorizes host execution.
// See docs/technical/archive/architecture.md#mandatory-container-isolation-design-contract.
func historyRPC(ctx context.Context, method string, params any) (json.RawMessage, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "codex", "app-server", "--stdio", "-c", "analytics.enabled=false")
	cmd.WaitDelay = time.Second
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = io.Discard
	if err = cmd.Start(); err != nil {
		return nil, err
	}
	defer func() { in.Close(); cancel(); cmd.Wait() }()
	encoder := json.NewEncoder(in)
	scanner := bufio.NewScanner(out)
	scanner.Buffer(make([]byte, 4096), 16*1024*1024)
	request := func(id int, method string, params any) (json.RawMessage, error) {
		if err := encoder.Encode(map[string]any{"id": id, "method": method, "params": params}); err != nil {
			return nil, err
		}
		for scanner.Scan() {
			var e struct {
				ID     *int            `json:"id"`
				Result json.RawMessage `json:"result"`
				Error  *struct {
					Message string `json:"message"`
				} `json:"error"`
			}
			if json.Unmarshal(scanner.Bytes(), &e) != nil || e.ID == nil || *e.ID != id {
				continue
			}
			if e.Error != nil {
				return nil, errors.New(e.Error.Message)
			}
			return e.Result, nil
		}
		if scanner.Err() != nil {
			return nil, scanner.Err()
		}
		return nil, errors.New("Codex history service disconnected")
	}
	if _, err = request(1, "initialize", map[string]any{"clientInfo": map[string]string{"name": "agentchat", "title": "AgentChat", "version": "0.1.0"}}); err != nil {
		return nil, err
	}
	if err = encoder.Encode(map[string]string{"method": "initialized"}); err != nil {
		return nil, err
	}
	return request(2, method, params)
}
