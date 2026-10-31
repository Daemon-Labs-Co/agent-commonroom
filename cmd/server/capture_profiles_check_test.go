package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

type failedCaptureCheckWriter struct{}

func (failedCaptureCheckWriter) Write([]byte) (int, error) {
	return 0, errors.New("synthetic output failure")
}

func TestCaptureProfileOfflineCheck(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"valid", "reader policy", "missing PEM", "bad path", "output failure", "extra operands"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()
			g, _, _, _ := captureSessionFixture(t)
			path, file := captureProfileFixture(t, g)
			args := []string{path}
			want := 0
			switch kind {
			case "reader policy":
				file.Protocol = objectReadControlALPN
				writeCaptureProfiles(t, path, file)
				want = 1
			case "missing PEM":
				if err := os.WriteFile(file.Profiles[0].Identity.Certificate, []byte("synthetic certificate without PEM markers"), 0600); err != nil {
					t.Fatal(err)
				}
				want = 1
			case "bad path":
				args = []string{"/synthetic/secret-canary"}
				want = 1
			case "output failure":
				want = 1
			case "extra operands":
				args = append(args, "secret-canary")
				want = 2
			}
			var out, diagnostic bytes.Buffer
			var writer io.Writer = &out
			if kind == "output failure" {
				writer = failedCaptureCheckWriter{}
			}
			if code := captureProfileCheckMain(args, writer, &diagnostic); code != want {
				t.Fatalf("code=%d want=%d diagnostic=%s", code, want, diagnostic.String())
			}
			if strings.Contains(out.String()+diagnostic.String(), "secret-canary") || strings.Contains(diagnostic.String(), path) {
				t.Fatal("private operand disclosed")
			}
			if kind == "valid" && (!strings.Contains(out.String(), "2 entries") || diagnostic.Len() != 0 || !strings.Contains(out.String(), "have not been verified")) {
				t.Fatal("configuration check overstated evidence")
			}
			if kind == "missing PEM" && !strings.Contains(diagnostic.String(), "BEGIN CERTIFICATE") {
				t.Fatal("PEM correction missing", diagnostic.String())
			}
		})
	}
}
