package main

import (
	"agentroom/internal/harnessrelease"
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func publicationFixture(t *testing.T) ([]string, ed25519.PublicKey, time.Time) {
	t.Helper()
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	review := harnessrelease.Review{Schema: 1, Sequence: 1, Provider: "codex", Version: "synthetic-review-1", Image: "registry.example/engine@sha256:" + strings.Repeat("a", 64), Profile: "codex-separated-callback-v1", AuthProtocol: 1, AuditProtocol: 2, WorkspaceProtocol: 4, IssuedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Hour)}
	payload, err := json.Marshal(review)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	source := filepath.Join(dir, "review.json")
	key := filepath.Join(dir, "key")
	dest := filepath.Join(dir, "signed.json")
	for path, raw := range map[string][]byte{source: payload, key: []byte(base64.StdEncoding.EncodeToString(private) + "\n")} {
		if err := os.WriteFile(path, raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return []string{"--review", source, "--signing-key", key, "--output", dest}, public, now
}
func TestPublishReviewRoundTrip(t *testing.T) {
	t.Parallel()
	args, key, now := publicationFixture(t)
	var out bytes.Buffer
	if err := run(args, &out, now); err != nil {
		t.Fatal(err)
	}
	raw, err := harnessrelease.ReadPrivate(args[5], 4096)
	if err != nil {
		t.Fatal(err)
	}
	selected, err := harnessrelease.Verify(raw, key, now)
	if err != nil || selected.Version != "synthetic-review-1" {
		t.Fatal(selected, err)
	}
	original := append([]byte(nil), raw...)
	if err := run(args, &out, now); err == nil {
		t.Fatal("overwrote existing publication")
	}
	after, _ := os.ReadFile(args[5])
	if !bytes.Equal(original, after) {
		t.Fatal("existing publication changed")
	}
	matches, _ := filepath.Glob(filepath.Join(filepath.Dir(args[5]), ".harness-review-*"))
	if len(matches) != 0 {
		t.Fatal("temporary publication leaked")
	}
}
func TestPublishReviewRefusals(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"public-key-file", "symlink-key", "hardlink-key", "directory-key", "bad-key", "mismatched-key", "expired", "noncanonical", "relative-output", "symlink-output", "missing-parent", "arguments"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			args, _, now := publicationFixture(t)
			var out bytes.Buffer
			switch mode {
			case "public-key-file":
				if err := os.Chmod(args[3], 0644); err != nil {
					t.Fatal(err)
				}
			case "symlink-key":
				original := args[3] + ".original"
				if err := os.Rename(args[3], original); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(original, args[3]); err != nil {
					t.Fatal(err)
				}
			case "hardlink-key":
				if err := os.Link(args[3], args[3]+".link"); err != nil {
					t.Fatal(err)
				}
			case "directory-key":
				args[3] = t.TempDir()
			case "bad-key":
				if err := os.WriteFile(args[3], []byte("invalid-private-key"), 0600); err != nil {
					t.Fatal(err)
				}
			case "mismatched-key":
				encoded, _ := os.ReadFile(args[3])
				raw, _ := base64.StdEncoding.DecodeString(strings.TrimSpace(string(encoded)))
				raw[63] ^= 1
				if err := os.WriteFile(args[3], []byte(base64.StdEncoding.EncodeToString(raw)), 0600); err != nil {
					t.Fatal(err)
				}
			case "expired":
				now = now.Add(2 * time.Hour)
			case "noncanonical":
				raw, _ := os.ReadFile(args[1])
				if err := os.WriteFile(args[1], append(raw, '\n'), 0600); err != nil {
					t.Fatal(err)
				}
			case "relative-output":
				args[5] = "relative-output.json"
			case "symlink-output":
				if err := os.Symlink(args[1], args[5]); err != nil {
					t.Fatal(err)
				}
			case "missing-parent":
				args[5] = filepath.Join(t.TempDir(), "absent", "review.json")
			case "arguments":
				args = append(args, "unexpected")
			}
			if err := run(args, &out, now); err == nil {
				t.Fatal("unsafe publication succeeded")
			}
			if out.Len() != 0 {
				t.Fatal("failure emitted success or key material")
			}
			if mode != "symlink-output" {
				if _, err := os.Stat(args[5]); !os.IsNotExist(err) {
					t.Fatal("failed publication left destination", err)
				}
			}
		})
	}
}
func TestConcurrentPublicationIsExclusive(t *testing.T) {
	t.Parallel()
	args, key, now := publicationFixture(t)
	var wg sync.WaitGroup
	outcomes := make(chan error, 2)
	for range 2 {
		wg.Add(1)
		go func() { defer wg.Done(); outcomes <- run(args, &bytes.Buffer{}, now) }()
	}
	wg.Wait()
	close(outcomes)
	success := 0
	for err := range outcomes {
		if err == nil {
			success++
		}
	}
	if success != 1 {
		t.Fatal("publication was not exclusive", success)
	}
	raw, err := harnessrelease.ReadPrivate(args[5], 4096)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := harnessrelease.Verify(raw, key, now); err != nil {
		t.Fatal(err)
	}
}

func TestGenerateOfflineReviewKey(t *testing.T) {
	t.Parallel()
	args, _, now := publicationFixture(t)
	keyPath := filepath.Join(t.TempDir(), "new-review-key")
	var out bytes.Buffer
	if err := run([]string{"--generate-key", keyPath}, &out, now); err != nil {
		t.Fatal(err)
	}
	encoded, err := harnessrelease.ReadPrivate(keyPath, 128)
	if err != nil {
		t.Fatal(err)
	}
	private, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(encoded)))
	if err != nil || len(private) != 64 {
		t.Fatal("invalid generated key")
	}
	expected := "COMMONROOM_HARNESS_REVIEW_KEY=" + base64.StdEncoding.EncodeToString(ed25519.PrivateKey(private).Public().(ed25519.PublicKey)) + "\n"
	if out.String() != expected || strings.Contains(out.String(), strings.TrimSpace(string(encoded))) {
		t.Fatal("wrong public key output")
	}
	args[3] = keyPath
	if err := run(args, &bytes.Buffer{}, now); err != nil {
		t.Fatal("generated key cannot sign", err)
	}
	if err := run([]string{"--generate-key", keyPath}, &bytes.Buffer{}, now); err == nil {
		t.Fatal("overwrote signing key")
	}
	if err := run([]string{"--generate-key", filepath.Join(t.TempDir(), "key"), "--review", args[1]}, &bytes.Buffer{}, now); err == nil {
		t.Fatal("ambiguous mode accepted")
	}
}
