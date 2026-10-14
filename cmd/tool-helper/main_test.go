//go:build linux || darwin

package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agentroom/internal/toolbroker"
	"agentroom/internal/toolworkspace"
)

func testRoot(t *testing.T) (*os.Root, string) {
	t.Helper()
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { root.Close() })
	return root, dir
}

func wire(t *testing.T, action toolbroker.Action) *bytes.Reader {
	t.Helper()
	raw, err := json.Marshal(action)
	if err != nil {
		t.Fatal(err)
	}
	return bytes.NewReader(raw)
}

func TestFileRoundTripMatchesBrokerWire(t *testing.T) {
	t.Parallel()
	root, dir := testRoot(t)
	content := "synthetic fixture\x00<&>\n"
	var output bytes.Buffer
	if err := run(context.Background(), wire(t, toolbroker.Action{Operation: "write", Path: "file", Content: content}), &output, root); err != nil {
		t.Fatal(err)
	}
	var result fileResult
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256([]byte(content))
	if result.Operation != "write" || result.Bytes != len(content) || result.SHA256 != hex.EncodeToString(digest[:]) || result.Data != "" {
		t.Fatalf("unexpected write metadata: %+v", result)
	}
	actual, err := os.ReadFile(filepath.Join(dir, "file"))
	if err != nil || string(actual) != content {
		t.Fatal("confirmed write missing")
	}
	output.Reset()
	if err := run(context.Background(), wire(t, toolbroker.Action{Operation: "read", Path: "file"}), &output, root); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	decoded, err := base64.StdEncoding.DecodeString(result.Data)
	if err != nil || string(decoded) != content || result.Encoding != "base64" || result.Operation != "read" || result.SHA256 != hex.EncodeToString(digest[:]) {
		t.Fatal("read response failed")
	}
}

func TestProtocolRefusesMalformedOrUndeclaredActions(t *testing.T) {
	t.Parallel()
	valid := `{"Operation":"write","Path":"file","Content":"new","Command":null}`
	cases := []string{
		`null`, `[]`, `{}`, valid + ` {}`, valid + ` false`,
		strings.Replace(valid, `"Operation"`, `"operation"`, 1),
		strings.Replace(valid, `"Command":null`, `"Root":"/tmp","Command":null`, 1),
		strings.Replace(valid, `"Operation":"write"`, `"Operation":"write","\u004fperation":"write"`, 1),
		strings.Replace(valid, `"Path":"file"`, `"Path":null`, 1),
		strings.Replace(valid, `"Content":"new"`, `"Content":null`, 1),
		strings.Replace(valid, `"Command":null`, `"Command":["touch","file"]`, 1),
		strings.Replace(valid, `"Command":null`, `"Command":{}`, 1),
		strings.Replace(valid, `"Command":null`, `"Command":""`, 1),
		strings.Replace(valid, `"write"`, `"command"`, 1),
		strings.Replace(valid, `"write"`, `"patch"`, 1),
		strings.Replace(valid, `"write"`, `"read"`, 1),
		strings.Replace(valid, `"file"`, `"../file"`, 1),
		strings.Replace(valid, `"file"`, `"/file"`, 1),
		strings.Replace(valid, `"file"`, `"a/../file"`, 1),
		strings.Replace(valid, `"file"`, `"file\u0000"`, 1),
		strings.Replace(valid, `"new"`, `"`+strings.Repeat("x", maxInputBytes)+`"`, 1),
		strings.Replace(valid, `"new"`, "\"\xff\"", 1),
		strings.Replace(valid, `"new"`, `"\ud800"`, 1),
		strings.Replace(valid, `"file"`, `"\udfff"`, 1),
	}
	for _, input := range cases {
		root, dir := testRoot(t)
		var output bytes.Buffer
		if err := run(context.Background(), strings.NewReader(input), &output, root); !errors.Is(err, errRefused) {
			t.Fatalf("malformed action accepted: %v", err)
		}
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) != 0 || output.Len() != 0 {
			t.Fatal("refused action produced effect/output")
		}
	}
	if _, err := decodeAction(strings.NewReader(strings.Replace(valid, `null`, `[ ]`, 1))); err != nil {
		t.Fatal("empty command array should be accepted")
	}
}

func TestReadBinaryBudgetAndOutsideLink(t *testing.T) {
	t.Parallel()
	root, dir := testRoot(t)
	data := bytes.Repeat([]byte{0xff}, toolworkspace.MaxReadBytes)
	if err := os.WriteFile(filepath.Join(dir, "file"), data, 0600); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := run(context.Background(), wire(t, toolbroker.Action{Operation: "read", Path: "file"}), &output, root); err != nil {
		t.Fatal(err)
	}
	if output.Len() > maxOutputBytes {
		t.Fatal("output exceeded broker limit")
	}
	outside := filepath.Join(t.TempDir(), "canary")
	if err := os.WriteFile(outside, []byte("synthetic canary"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dir, "link")); err != nil {
		t.Fatal(err)
	}
	for _, operation := range []string{"read", "write"} {
		output.Reset()
		if err := run(context.Background(), wire(t, toolbroker.Action{Operation: operation, Path: "link"}), &output, root); !errors.Is(err, errRefused) {
			t.Fatal("outside link accepted")
		}
	}
	actual, _ := os.ReadFile(outside)
	if string(actual) != "synthetic canary" {
		t.Fatal("outside file changed")
	}
}

type countingReader struct{ count int }

func (r *countingReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = ' '
	}
	r.count += len(p)
	return len(p), nil
}

func TestInputReadStopsAtBudget(t *testing.T) {
	t.Parallel()
	reader := &countingReader{}
	if _, err := decodeAction(reader); !errors.Is(err, errRefused) {
		t.Fatal("oversize stream accepted")
	}
	if reader.count != maxInputBytes+1 {
		t.Fatal("stream read exceeded bounded lookahead")
	}
}

func TestOversizeFileRefusedWithoutOutput(t *testing.T) {
	t.Parallel()
	root, dir := testRoot(t)
	if err := os.WriteFile(filepath.Join(dir, "file"), bytes.Repeat([]byte{'x'}, toolworkspace.MaxReadBytes+1), 0600); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := run(context.Background(), wire(t, toolbroker.Action{Operation: "read", Path: "file"}), &output, root); !errors.Is(err, errRefused) || output.Len() != 0 {
		t.Fatal("oversize read leaked output")
	}
}

type failedWriter struct{}

func (failedWriter) Write([]byte) (int, error) { return 0, errors.New("synthetic output failure") }

func TestWriteOutputFailureIsUnconfirmed(t *testing.T) {
	t.Parallel()
	root, dir := testRoot(t)
	err := run(context.Background(), wire(t, toolbroker.Action{Operation: "write", Path: "file", Content: "fixture"}), failedWriter{}, root)
	if !errors.Is(err, errUnconfirmed) {
		t.Fatal("write output failure claimed confirmation")
	}
	data, err := os.ReadFile(filepath.Join(dir, "file"))
	if err != nil || string(data) != "fixture" {
		t.Fatal("test did not exercise post-write ambiguity")
	}
}

func TestCancelledOrMissingDependenciesRefused(t *testing.T) {
	t.Parallel()
	root, _ := testRoot(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, candidate := range []context.Context{nil, ctx} {
		if err := run(candidate, strings.NewReader(`{}`), io.Discard, root); !errors.Is(err, errRefused) {
			t.Fatal("invalid context accepted")
		}
	}
	if err := run(context.Background(), nil, io.Discard, root); !errors.Is(err, errRefused) {
		t.Fatal("nil input accepted")
	}
	if err := run(context.Background(), wire(t, toolbroker.Action{Operation: "read", Path: "file"}), io.Discard, nil); !errors.Is(err, errRefused) {
		t.Fatal("nil root accepted")
	}
}

func TestRepositoryRootRefusesSymlinkOrMissingDirectory(t *testing.T) {
	t.Parallel()
	parent, dir := testRoot(t)
	if root, err := repositoryRoot(parent); err == nil {
		root.Close()
		t.Fatal("missing root accepted")
	}
	if err := os.Symlink(t.TempDir(), filepath.Join(dir, "repository")); err != nil {
		t.Fatal(err)
	}
	if root, err := repositoryRoot(parent); err == nil {
		root.Close()
		t.Fatal("symlink root accepted")
	}
	if err := os.Remove(filepath.Join(dir, "repository")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "repository"), 0700); err != nil {
		t.Fatal(err)
	}
	root, err := repositoryRoot(parent)
	if err != nil {
		t.Fatal(err)
	}
	root.Close()
}
