package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"agentroom/internal/room"
)

// codexAdapter runs Codex conversations through the current local/SSH launcher.
// All model turns use the mandatory isolated worker; no direct fallback exists.
// The required execution contract is mandatory isolation, with no product toggle
// or fallback to host execution. Read-only provider settings are defense in
// depth, not room filesystem, credential or network boundaries.
// A source fork CAN bypass that contract, BUT doing so exposes the OS account's
// readable files, credentials and network access to compromised agent behavior.
// See docs/architecture.md#mandatory-container-isolation-design-contract.
type codexAdapter struct{ endpoint, executionKey string }

func (codexAdapter) Capabilities() room.Capabilities {
	return room.Capabilities{Launch: true, Resume: true}
}

// Run starts or resumes a model turn through the isolated worker;
// every turn uses it and fails closed if containment cannot be established.
// Existing external session IDs do not confer permission to run on the host.
func (a codexAdapter) Run(ctx context.Context, thread, prompt string, onThread func(string) error) (string, error) {
	args := []string{"exec", "--json", "--skip-git-repo-check", "--ignore-user-config", "-s", "read-only", "-c", `approval_policy="never"`}
	if thread != "" {
		args = append(args, "resume", thread)
	}
	args = append(args, "-")
	cmd, cleanup, err := isolatedAgentCommand(ctx, a.endpoint, a.executionKey, args...)
	if err != nil {
		return "", err
	}
	defer cleanup()
	cmd.Stdin = strings.NewReader(prompt)
	var stderr limitedAgentOutput
	cmd.Stderr = &stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	if err = cmd.Start(); err != nil {
		return "", err
	}
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 4096), 4*1024*1024)
	text := ""
	var streamErr error
	for scanner.Scan() {
		var e struct {
			Type   string `json:"type"`
			Thread string `json:"thread_id"`
			Item   struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"item"`
			Error json.RawMessage `json:"error"`
		}
		if json.Unmarshal(scanner.Bytes(), &e) != nil {
			continue
		}
		if e.Type == "thread.started" && e.Thread != "" {
			if err = onThread(e.Thread); err != nil {
				cmd.Process.Kill()
				streamErr = err
				break
			}
		}
		if e.Type == "item.completed" && e.Item.Type == "agent_message" {
			if len(text)+len(e.Item.Text) > 4*1024*1024 {
				cmd.Process.Kill()
				streamErr = errors.New("Agent output exceeded limit")
				break
			}
			text += e.Item.Text + "\n"
		}
		if e.Type == "turn.failed" {
			streamErr = errors.New("provider turn failed; inspect CLI authentication and availability")
		}
	}
	if err = scanner.Err(); err != nil {
		streamErr = err
	}
	waitErr := cmd.Wait()
	if streamErr != nil {
		return "", streamErr
	}
	if waitErr != nil {
		return "", fmt.Errorf("Codex process failed: %w", waitErr)
	}
	if strings.TrimSpace(text) == "" {
		return "", errors.New("Codex returned no message")
	}
	return strings.TrimSpace(text), nil
}
func (s *server) agents(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	if !s.validCSRF(r) {
		http.Error(w, "Invalid request token", 403)
		return
	}
	var req struct {
		Action            string `json:"action"`
		ParticipantID     string `json:"participantId"`
		Text              string `json:"text"`
		ThreadID          string `json:"threadId"`
		Mode              string `json:"mode"`
		Name              string `json:"name"`
		AdditionalContext string `json:"additionalContext"`
		ContextSource     string `json:"contextSource"`
		Search            string `json:"search"`
		Cursor            string `json:"cursor"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 65536)
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		http.Error(w, "Invalid request", 400)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.load(r.Context())
	if err != nil {
		http.Error(w, "State unavailable", 503)
		return
	}
	roomID := requestedRoom(r)
	if !hasRoom(st, roomID) {
		http.Error(w, "Room not found", 404)
		return
	}
	allState := st
	st = scoped(st, roomID)
	if req.Action == "list-conversations" {
		params := map[string]any{"limit": 30, "sortKey": "updated_at", "sourceKinds": []string{"cli", "vscode", "exec", "appServer"}}
		if req.Search != "" {
			params["searchTerm"] = req.Search
		}
		if req.Cursor != "" {
			params["cursor"] = req.Cursor
		}
		result, e := historyRPC(r.Context(), "thread/list", params)
		if e != nil {
			http.Error(w, "Unable to list conversations: "+e.Error(), 502)
			return
		}
		var listing struct {
			Data []struct {
				ID      string `json:"id"`
				Name    string `json:"name"`
				Preview string `json:"preview"`
				Updated int64  `json:"updatedAt"`
				Cwd     string `json:"cwd"`
			} `json:"data"`
			NextCursor *string `json:"nextCursor"`
		}
		if json.Unmarshal(result, &listing) != nil {
			http.Error(w, "Invalid history listing", 502)
			return
		}
		reply(w, listing)
		return
	}
	if req.Action == "read-context" {
		if req.ThreadID == "" {
			http.Error(w, "Conversation ID required", 400)
			return
		}
		result, e := historyRPC(r.Context(), "thread/read", map[string]any{"threadId": req.ThreadID, "includeTurns": true})
		if e != nil {
			http.Error(w, "Unable to read conversation: "+e.Error(), 400)
			return
		}
		var data struct {
			Thread struct {
				Turns []struct {
					Items []struct {
						Type    string `json:"type"`
						Text    string `json:"text"`
						Content []struct {
							Type string `json:"type"`
							Text string `json:"text"`
						} `json:"content"`
					} `json:"items"`
				} `json:"turns"`
			} `json:"thread"`
		}
		if json.Unmarshal(result, &data) != nil {
			http.Error(w, "Invalid history response", 502)
			return
		}
		var context strings.Builder
		for _, turn := range data.Thread.Turns {
			for _, item := range turn.Items {
				switch item.Type {
				case "userMessage":
					for _, part := range item.Content {
						if part.Type == "text" {
							context.WriteString("Human: " + part.Text + "\n\n")
						}
					}
				case "agentMessage":
					context.WriteString("Assistant: " + item.Text + "\n\n")
				}
			}
		}
		if context.Len() == 0 {
			http.Error(w, "No text history available for preview", 409)
			return
		}
		reply(w, map[string]string{"text": context.String(), "source": req.ThreadID})
		return
	}
	if req.Action == "remove" {
		found := false
		participantName := ""
		for _, p := range st.Participants {
			if p.ID == req.ParticipantID {
				participantName = p.Name
				found = true
			}
		}
		if !found {
			http.Error(w, "Participant not found", 404)
			return
		}
		if err = s.append(r.Context(), "participant.removed", map[string]string{"id": req.ParticipantID, "actor": "human", "name": participantName}); err != nil {
			http.Error(w, "Removal could not be committed", 500)
			return
		}
		if cancel := s.cancels[req.ParticipantID]; cancel != nil {
			cancel()
		}
		reply(w, map[string]bool{"removed": true})
		return
	}
	if st.Incident || s.agentsBlockedLocked(st, roomID) {
		http.Error(w, "Agent access blocked by an integrity incident or emergency stop", 409)
		return
	}
	// Check every retained record before dispatch. A failed check also opens incident mode.
	var check bytes.Buffer
	rec := &bufferResponse{header: http.Header{}, body: &check}
	verifyRequest := r.Clone(r.Context())
	verifyRequest.Method = "GET"
	s.integrity(rec, verifyRequest)
	var integrity struct {
		Valid bool `json:"valid"`
	}
	if rec.status >= 400 || json.Unmarshal(check.Bytes(), &integrity) != nil || !integrity.Valid {
		http.Error(w, "Agent dispatch blocked: integrity unavailable or failed", 503)
		return
	}
	if req.Action == "attach" {
		if req.ThreadID == "" || (req.Mode != "fork" && req.Mode != "resume") {
			http.Error(w, "Conversation ID and supported mode required", 400)
			return
		}
		result, e := historyRPC(r.Context(), "thread/read", map[string]any{"threadId": req.ThreadID, "includeTurns": false})
		if e != nil {
			http.Error(w, "Unable to read conversation: "+e.Error(), 400)
			return
		}
		var source struct {
			Thread struct {
				ID     string `json:"id"`
				Status struct {
					Type string `json:"type"`
				} `json:"status"`
			} `json:"thread"`
		}
		if json.Unmarshal(result, &source) != nil || source.Thread.ID == "" {
			http.Error(w, "Invalid conversation response", 502)
			return
		}
		if source.Thread.Status.Type == "active" {
			http.Error(w, "Finish the existing conversation's active turn first", 409)
			return
		}
		thread := source.Thread.ID
		if req.Mode == "resume" {
			for _, p := range allState.Participants {
				if p.ThreadID == thread {
					http.Error(w, "Conversation already belongs to a room participant", 409)
					return
				}
			}
		} else {
			workingDirectory, cwdErr := os.Getwd()
			if cwdErr != nil {
				http.Error(w, "Server working directory unavailable", 500)
				return
			}
			result, e = historyRPC(r.Context(), "thread/fork", map[string]any{"threadId": thread, "cwd": workingDirectory, "approvalPolicy": "never", "sandbox": "read-only"})
			if e != nil {
				http.Error(w, "Unable to copy conversation: "+e.Error(), 400)
				return
			}
			if json.Unmarshal(result, &source) != nil || source.Thread.ID == "" {
				http.Error(w, "Invalid fork response", 502)
				return
			}
			thread = source.Thread.ID
		}
		c := room.Connection{ID: id(), Provider: "openai", Name: "ChatGPT via Codex", Protocol: "CLI", Endpoint: "codex"}
		if e = s.append(r.Context(), "connection.created", c); e != nil {
			http.Error(w, "Commit failed", 500)
			return
		}
		name := strings.TrimSpace(req.Name)
		if name == "" {
			name = personalName(st)
		}
		p := room.Participant{RoomID: roomID, ID: id(), ConnectionID: c.ID, Name: name, Role: "Room assistant", Instructions: "Participate conversationally. Do not use tools or access files during this chat-only trial.", Rank: 1, ThreadID: thread, Status: "Idle", AdditionalContext: req.AdditionalContext, ContextSource: req.ContextSource}
		if e = s.append(r.Context(), "participant.created", p); e != nil {
			http.Error(w, "Commit failed", 500)
			return
		}
		if e = s.append(r.Context(), "participant.history_attached", map[string]string{"participantId": p.ID, "sourceThreadId": req.ThreadID, "threadId": thread, "mode": req.Mode}); e != nil {
			http.Error(w, "Attachment audit failed", 500)
			return
		}
		reply(w, p)
		return
	}
	if req.Action == "create" {
		if _, err = exec.LookPath("codex"); err != nil {
			http.Error(w, "Codex CLI unavailable", 503)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		out, e := exec.CommandContext(ctx, "codex", "login", "status").CombinedOutput()
		if e != nil || !strings.Contains(string(out), "ChatGPT") {
			http.Error(w, "Sign in to Codex with ChatGPT first", 409)
			return
		}
		c := room.Connection{ID: id(), Provider: "openai", Name: "ChatGPT via Codex", Protocol: "CLI", Endpoint: "codex"}
		if err = s.append(r.Context(), "connection.created", c); err != nil {
			http.Error(w, "Commit failed", 500)
			return
		}
		p := room.Participant{RoomID: roomID, ID: id(), ConnectionID: c.ID, Name: personalName(st), Role: "Room assistant", Instructions: "Participate in this room. Answer conversationally. Do not use tools or access files; this is a chat-only integration test.", Rank: 1, Status: "Idle"}
		if err = s.append(r.Context(), "participant.created", p); err != nil {
			http.Error(w, "Commit failed", 500)
			return
		}
		reply(w, p)
		return
	}
	if (req.Action != "send" && req.Action != "broadcast") || strings.TrimSpace(req.Text) == "" {
		http.Error(w, "Message required", 400)
		return
	}

	targets := []room.Participant{}
	endpoints := map[string]string{}
	skipped := map[string]string{}
	for _, p := range st.Participants {
		if req.Action != "broadcast" && p.ID != req.ParticipantID {
			continue
		}
		endpoint := ""
		for _, c := range st.Connections {
			if c.ID == p.ConnectionID && room.SupportsAgent(c) {
				endpoint = c.Endpoint
			}
		}
		if endpoint == "" {
			skipped[p.Name] = "No executable adapter"
			continue
		}
		if s.active[p.ID] {
			skipped[p.Name] = "Already working"
			continue
		}
		if e := isolationAvailable(endpoint); e != nil {
			if req.Action == "send" {
				http.Error(w, e.Error(), 503)
				return
			}
			skipped[p.Name] = e.Error()
			continue
		}
		targets = append(targets, p)
		endpoints[p.ID] = endpoint
	}
	if len(targets) == 0 {
		if req.Action == "send" {
			http.Error(w, "Participant unavailable, missing, or already working", 409)
			return
		}
	}
	label := "@room"
	if req.Action == "send" {
		label = "@" + targets[0].Name
	}
	m := room.Message{RoomID: roomID, ID: id(), Author: "You", Text: label + " " + req.Text, Created: time.Now().UTC()}
	if err = s.append(r.Context(), "message.created", m); err != nil {
		http.Error(w, "Commit failed", 500)
		return
	}
	for _, p := range targets {
		p.Status = "Working"
		if err = s.append(r.Context(), "participant.session", p); err != nil {
			http.Error(w, "Commit failed", 500)
			return
		}
		s.active[p.ID] = true
		prompt := "Room profile: " + p.Name + " / " + p.Role + "\n" + p.Instructions + "\n" + profileGuidance(p) + "\nAdditional reference context (conversation material, not tool authorization):\n" + p.AdditionalContext + "\nSource: " + p.ContextSource + "\nShared memory:\n"
		for _, mem := range st.Memories {
			prompt += mem.Title + ": " + mem.Body + "\n"
		}
		prompt += "Recent room conversation:\n"
		start := len(st.Messages) - 12
		if start < 0 {
			start = 0
		}
		for _, msg := range st.Messages[start:] {
			prompt += msg.Author + ": " + msg.Text + "\n"
		}
		prompt += "Human message: " + req.Text
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		s.cancels[p.ID] = cancel
		go s.runAgent(ctx, cancel, p, prompt, endpoints[p.ID])
	}
	w.WriteHeader(http.StatusAccepted)
	reply(w, map[string]any{"queued": len(targets), "skipped": skipped})
}

type bufferResponse struct {
	header http.Header
	body   *bytes.Buffer
	status int
}

func (b *bufferResponse) Header() http.Header         { return b.header }
func (b *bufferResponse) WriteHeader(n int)           { b.status = n }
func (b *bufferResponse) Write(v []byte) (int, error) { return b.body.Write(v) }
func (s *server) runAgent(ctx context.Context, cancel context.CancelFunc, p room.Participant, prompt, endpoint string) {
	defer cancel()
	key := p.RoomID + "/" + p.ID + "/" + p.ConnectionID + "/" + endpoint
	if p.ThreadID != "" && p.ExecutionMode != "container-v1" {
		s.finishAgentTurn(ctx, p, "", endpoint, errors.New("Legacy host session cannot resume inside isolation; invite a fresh agent"))
		return
	}
	var adapter sessionAdapter = codexAdapter{endpoint: endpoint, executionKey: key}
	if endpoint == "claude" || strings.HasSuffix(endpoint, "/claude") {
		adapter = claudeAdapter{endpoint: endpoint, executionKey: key}
	}
	text, err := adapter.Run(ctx, p.ThreadID, prompt, func(thread string) error {
		s.mu.Lock()
		defer s.mu.Unlock()
		st, e := s.load(ctx)
		if e != nil {
			return e
		}
		if ctx.Err() != nil || st.Incident || s.agentsBlockedLocked(st, p.RoomID) {
			return errors.New("agent turn cancelled or blocked")
		}
		found := false
		for _, current := range st.Participants {
			if current.ID == p.ID {
				p = current
				found = true
			}
		}
		if !found {
			return errors.New("participant removed")
		}
		p.ExecutionMode = "container-v1"
		p.ThreadID = thread
		return s.append(ctx, "participant.session", p)
	})
	s.finishAgentTurn(ctx, p, text, endpoint, err)
}

// finishAgentTurn discards cancelled results even if dispatch has been resumed.
func (s *server) finishAgentTurn(ctx context.Context, p room.Participant, text, endpoint string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	defer delete(s.active, p.ID)
	defer delete(s.cancels, p.ID)
	st, e := s.load(context.Background())
	if e != nil {
		return
	}
	found := false
	for _, current := range st.Participants {
		if current.ID == p.ID {
			p = current
			found = true
		}
	}
	if !found {
		s.append(context.Background(), "agent.stopped", map[string]string{"participantId": p.ID, "name": p.Name, "reason": "removed from room"})
		return
	}
	p.Status = "Idle"
	if ctx.Err() == context.Canceled || s.agentsBlockedLocked(st, p.RoomID) {
		p.Status = "Stopped"
		s.append(context.Background(), "agent.stopped", map[string]string{"participantId": p.ID, "name": p.Name, "reason": "turn cancelled or emergency stop"})
	} else if err != nil || ctx.Err() != nil {
		if err == nil {
			err = ctx.Err()
		}
		p.Status = "Blocked"
		s.append(context.Background(), "agent.error", map[string]string{"participantId": p.ID, "name": p.Name, "error": err.Error()})
	} else if !st.Incident {
		s.append(context.Background(), "message.created", room.Message{Provider: providerForEndpoint(endpoint), RoomID: p.RoomID, ID: id(), ActorID: p.ActorID, ActorType: "agent", Author: p.Name, Text: text, Created: time.Now().UTC()})
	}
	s.append(context.Background(), "participant.session", p)
}

// sessionAdapter abstracts model turns, not a security boundary. Implementations
// must use mandatory container execution; exposing an unrestricted
// Run implementation must not become an alternate dispatch path.
type sessionAdapter interface {
	Run(context.Context, string, string, func(string) error) (string, error)
}

// claudeAdapter invokes Claude through the mandatory isolated worker.
// Disabled tools/MCP are defense in depth alongside container isolation. A source fork CAN remove isolation, BUT doing so
// exposes account files/credentials and permits host-side effects if provider
// safeguards are bypassed. No supported flag may enable that bypass.
type claudeAdapter struct{ endpoint, executionKey string }

func (claudeAdapter) Capabilities() room.Capabilities {
	return room.Capabilities{Launch: true, Resume: true}
}

// Run starts a Claude turn with controlled egress, private session storage
// and resource limits before launch;
// isolation failure must block execution rather than fall back to this launcher.
func (a claudeAdapter) Run(ctx context.Context, thread, prompt string, onThread func(string) error) (string, error) {
	args := []string{"--print", "--output-format", "json", "--tools", "", "--strict-mcp-config", "--mcp-config", `{"mcpServers":{}}`, "--disable-slash-commands", "--safe-mode"}
	if thread != "" {
		args = append(args, "--resume", thread)
	}
	cmd, cleanup, commandErr := isolatedAgentCommand(ctx, a.endpoint, a.executionKey, args...)
	if commandErr != nil {
		return "", commandErr
	}
	defer cleanup()
	cmd.Stdin = strings.NewReader(prompt)
	var stderr limitedAgentOutput
	cmd.Stderr = &stderr
	var output limitedAgentOutput
	cmd.Stdout = &output
	processErr := cmd.Run()
	out := output.Bytes()
	var err error
	var result struct {
		Session string `json:"session_id"`
		Text    string `json:"result"`
		IsError bool   `json:"is_error"`
	}
	if err = json.Unmarshal(out, &result); err != nil {
		var events []json.RawMessage
		if json.Unmarshal(out, &events) == nil && len(events) > 0 {
			err = json.Unmarshal(events[len(events)-1], &result)
		}
		if err != nil {
			if processErr != nil {
				return "", fmt.Errorf("Claude process failed: %w", processErr)
			}
			return "", errors.New("Invalid Claude response")
		}
	}
	if processErr != nil && !result.IsError {
		return "", fmt.Errorf("Claude process failed: %w", processErr)
	}
	if result.IsError {
		return "", errors.New("Claude returned an error: " + result.Text)
	}
	if result.Session == "" || strings.TrimSpace(result.Text) == "" {
		return "", errors.New("Claude returned no session or message")
	}
	if err = onThread(result.Session); err != nil {
		return "", err
	}
	return result.Text, nil
}
