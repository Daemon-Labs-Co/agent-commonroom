package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReferenceIncludesTypesMethodsAndCalls(t *testing.T) {
	source := []byte(`package example
// Store keeps records.
type Store struct { Count []int }
type Reader interface { Read() error }
// Save validates before writing.
func (s *Store) Save() { validate(); s.write() }
func validate() {}
`)
	got, err := describeGo("example.go", source)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Store", "Count []int", "Read() error", "(*Store).Save", "Save validates", "s.write", "validate"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q", want)
		}
	}
}

func TestGenerateExcludesPrivateStateAndIsRepeatable(t *testing.T) {
	root := t.TempDir()
	for path, body := range map[string]string{"README.md": "# Example\n", "sample.go": "package example\nfunc Run() {}\n", ".data/private-name": "private", ".env.staging": "private", ".pgpass": "private", "bin/app": "generated", "web/widget.js": "function show() {}\nconst hide = () => {};\n"} {
		full := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := generate(root); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "docs/technical/reference/sample-go.md")
	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(first), "private") || !strings.Contains(string(first), "Run") {
		t.Fatal("unexpected reference content")
	}
	if err := generate(root); err != nil {
		t.Fatal(err)
	}
	second, _ := os.ReadFile(path)
	if string(first) != string(second) {
		t.Fatal("reference is not repeatable")
	}
	index, _ := os.ReadFile(filepath.Join(root, "docs/technical/reference/root.md"))
	for _, secretName := range []string{".env.staging", ".pgpass", "private-name"} {
		if strings.Contains(string(index), secretName) {
			t.Errorf("private filename leaked: %s", secretName)
		}
	}
	js, _ := os.ReadFile(filepath.Join(root, "docs/technical/reference/web.md"))
	for _, name := range []string{"show", "hide"} {
		if !strings.Contains(string(js), name) {
			t.Errorf("missing JS symbol %s", name)
		}
	}
}
