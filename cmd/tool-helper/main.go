//go:build linux || darwin

// tool-helper is the planned tool-container's bounded file protocol, not a host
// execution API. Managed use MUST remain inside the verified credential-free
// container and audited broker. Calling it on a host CAN bypass isolation, BUT
// exposes host files and invalidates the mandatory execution contract.
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
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"agentroom/internal/strictjson"
	"agentroom/internal/toolworkspace"
)

const maxInputBytes = 64 * 1024
const maxOutputBytes = 16 * 1024

var errRefused = errors.New("tool file operation refused")
var errUnconfirmed = errors.New("tool write outcome unconfirmed; do not retry automatically")

type fileAction struct {
	operation string
	path      string
	content   string
}

// decodeAction accepts exactly the broker Action wire fields. Revalidating here
// does not authenticate a caller: only the broker may authorize/admit operations.
func decodeAction(in io.Reader) (fileAction, error) {
	var action fileAction
	if in == nil {
		return action, errRefused
	}
	raw, err := io.ReadAll(io.LimitReader(in, maxInputBytes+1))
	if err != nil || len(raw) > maxInputBytes || !strictjson.ValidEncoding(raw) {
		return action, errRefused
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	start, err := decoder.Token()
	if err != nil || start != json.Delim('{') {
		return action, errRefused
	}
	seen := make(map[string]bool, 4)
	for decoder.More() {
		key, err := decoder.Token()
		name, ok := key.(string)
		if err != nil || !ok || seen[name] || (name != "Operation" && name != "Path" && name != "Content" && name != "Command") {
			return action, errRefused
		}
		seen[name] = true
		var value json.RawMessage
		if decoder.Decode(&value) != nil {
			return action, errRefused
		}
		if name == "Command" {
			if bytes.Equal(value, []byte("null")) {
				continue
			}
			var args []json.RawMessage
			if json.Unmarshal(value, &args) != nil || len(args) != 0 || len(value) == 0 || value[0] != '[' {
				return action, errRefused
			}
			continue
		}
		var text string
		if bytes.Equal(value, []byte("null")) || json.Unmarshal(value, &text) != nil {
			return action, errRefused
		}
		switch name {
		case "Operation":
			action.operation = text
		case "Path":
			action.path = text
		case "Content":
			action.content = text
		}
	}
	end, err := decoder.Token()
	if err != nil || end != json.Delim('}') || len(seen) != 4 {
		return action, errRefused
	}
	if _, err = decoder.Token(); err != io.EOF {
		return action, errRefused
	}
	if action.operation != "read" && action.operation != "write" {
		return action, errRefused
	}
	if action.operation == "read" && action.content != "" {
		return action, errRefused
	}
	return action, nil
}

type fileResult struct {
	Operation string `json:"operation"`
	Bytes     int    `json:"bytes"`
	SHA256    string `json:"sha256"`
	Encoding  string `json:"encoding,omitempty"`
	Data      string `json:"data,omitempty"`
}

// run borrows a trusted root for one operation. There is no command, patch,
// directory creation or alternate-root fallback. Output failure after a write
// is ambiguous and must not cause replay. The supervisor owns hard cancellation.
func run(ctx context.Context, in io.Reader, out io.Writer, root *os.Root) error {
	if ctx == nil || ctx.Err() != nil || out == nil {
		return errRefused
	}
	action, err := decodeAction(in)
	if err != nil {
		return errRefused
	}
	workspace, err := toolworkspace.New(root)
	if err != nil {
		return errRefused
	}
	var data []byte
	if action.operation == "read" {
		data, err = workspace.Read(ctx, action.path)
	} else {
		data = []byte(action.content)
		err = workspace.Write(ctx, action.path, data)
	}
	if errors.Is(err, toolworkspace.ErrWriteUnconfirmed) {
		return errUnconfirmed
	}
	if err != nil {
		return errRefused
	}
	digest := sha256.Sum256(data)
	result := fileResult{Operation: action.operation, Bytes: len(data), SHA256: hex.EncodeToString(digest[:])}
	if action.operation == "read" {
		// Base64 carries arbitrary file bytes without unbounded JSON escaping.
		result.Encoding = "base64"
		result.Data = base64.StdEncoding.EncodeToString(data)
	}
	encoded, err := json.Marshal(result)
	if err != nil || len(encoded)+1 > maxOutputBytes || ctx.Err() != nil {
		if action.operation == "write" {
			return errUnconfirmed
		}
		return errRefused
	}
	encoded = append(encoded, '\n')
	n, err := out.Write(encoded)
	if err != nil || n != len(encoded) || ctx.Err() != nil {
		if action.operation == "write" {
			return errUnconfirmed
		}
		return errRefused
	}
	return nil
}

// openWorkspace uses a fixed container mount; neither argv nor the payload may
// select roots. os.Root is path containment, not container/mount attestation.
func openWorkspace() (*os.Root, error) {
	parent, err := os.OpenRoot("/workspace")
	if err != nil {
		return nil, errRefused
	}
	defer parent.Close()
	return repositoryRoot(parent)
}

func repositoryRoot(parent *os.Root) (*os.Root, error) {
	if parent == nil {
		return nil, errRefused
	}
	info, err := parent.Lstat("repository")
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, errRefused
	}
	root, err := parent.OpenRoot("repository")
	if err != nil {
		return nil, errRefused
	}
	opened, err := root.Stat(".")
	if err != nil || !os.SameFile(info, opened) {
		root.Close()
		return nil, errRefused
	}
	return root, nil
}

func main() {
	if runtime.GOOS != "linux" || len(os.Args) != 1 {
		fmt.Fprintln(os.Stderr, errRefused)
		os.Exit(2)
	}
	root, err := openWorkspace()
	if err != nil {
		fmt.Fprintln(os.Stderr, errRefused)
		os.Exit(2)
	}
	defer root.Close()
	signals, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(signals, 3*time.Minute)
	defer cancel()
	if err = run(ctx, os.Stdin, os.Stdout, root); err != nil {
		fmt.Fprintln(os.Stderr, err)
		if errors.Is(err, errUnconfirmed) {
			os.Exit(3)
		}
		os.Exit(2)
	}
}
