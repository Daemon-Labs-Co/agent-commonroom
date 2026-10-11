package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func localClaudeConnection(stConnectionProtocol, endpoint string) bool {
	return stConnectionProtocol == "CLI" && endpoint == "claude"
}
func shellArgument(value string) string { return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'" }

type cappedStatusOutput struct{ bytes.Buffer }

func (b *cappedStatusOutput) Write(p []byte) (int, error) {
	if b.Len()+len(p) > 16384 {
		return 0, io.ErrShortBuffer
	}
	return b.Buffer.Write(p)
}
func (s *server) providerAuth(w http.ResponseWriter, r *http.Request) {
	if requestActor(r) != ownerID {
		http.Error(w, "Owner permission required", 403)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	if !s.validCSRF(r) {
		http.Error(w, "Invalid request token", 403)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var req struct {
		Action       string `json:"action"`
		ConnectionID string `json:"connectionId"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		http.Error(w, "Invalid request", 400)
		return
	}
	st, e := s.load(r.Context())
	if e != nil {
		http.Error(w, "State unavailable", 503)
		return
	}
	found := false
	for _, c := range st.Connections {
		if c.ID == req.ConnectionID && localClaudeConnection(c.Protocol, c.Endpoint) {
			found = true
		}
	}
	if !found {
		http.Error(w, "This action supports registered local Claude CLI connections only", 400)
		return
	}
	if st.Incident {
		http.Error(w, "Provider operations blocked during an integrity incident", 409)
		return
	}
	binary, e := exec.LookPath("claude")
	if e != nil {
		http.Error(w, "Claude CLI is not installed on this computer", 503)
		return
	}
	binary, e = filepath.Abs(binary)
	if e != nil {
		http.Error(w, "Claude executable unavailable", 503)
		return
	}
	switch req.Action {
	case "status":
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, binary, "auth", "status")
		var out cappedStatusOutput
		cmd.Stdout = &out
		cmd.Stderr = io.Discard
		processErr := cmd.Run()
		var status struct {
			LoggedIn bool `json:"loggedIn"`
		}
		if json.Unmarshal(out.Bytes(), &status) != nil {
			http.Error(w, "Unable to check Claude sign-in status", 502)
			return
		}
		if processErr != nil && status.LoggedIn {
			http.Error(w, "Unable to confirm Claude sign-in", 502)
			return
		}
		reply(w, map[string]bool{"loggedIn": status.LoggedIn})
	case "login":
		if runtime.GOOS != "darwin" {
			http.Error(w, "In-app login launch currently supports macOS only", 501)
			return
		}
		// Run the official interactive flow in Terminal; never proxy credentials or OAuth output.
		if e = os.MkdirAll(".data", 0700); e != nil {
			http.Error(w, "Unable to prepare sign-in", 500)
			return
		}
		script, e := os.CreateTemp(".data", "claude-login-*.command")
		if e != nil {
			http.Error(w, "Unable to prepare sign-in", 500)
			return
		}
		path, pathErr := filepath.Abs(script.Name())
		if pathErr != nil {
			script.Close()
			os.Remove(script.Name())
			http.Error(w, "Unable to prepare sign-in", 500)
			return
		}
		if e = script.Chmod(0700); e == nil {
			_, e = script.WriteString("#!/bin/sh\ntrap 'rm -- \"$0\"' EXIT\n" + shellArgument(binary) + " auth login --claudeai\n")
		}
		closeErr := script.Close()
		if e != nil || closeErr != nil {
			os.Remove(path)
			http.Error(w, "Unable to prepare sign-in", 500)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if e = exec.CommandContext(ctx, "/usr/bin/open", "-a", "Terminal", path).Run(); e != nil {
			os.Remove(path)
			http.Error(w, "Unable to open Claude sign-in in Terminal", 502)
			return
		}
		if e = s.append(r.Context(), "provider.login_launched", map[string]string{"actorId": ownerID, "connectionId": req.ConnectionID, "provider": "claude"}); e != nil {
			http.Error(w, "Sign-in launched but audit recording failed", 500)
			return
		}
		reply(w, map[string]bool{"launched": true})
	default:
		http.Error(w, "Unsupported sign-in action", 400)
	}
}
