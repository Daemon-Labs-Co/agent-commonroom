package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestObjectReadProfileCheck(t *testing.T) {
	t.Parallel()
	g, _, _, _ := objectReadSessionFixture(t)
	path, file := objectReadProfileFixture(t, g)
	var out, diagnostic bytes.Buffer
	if code := objectReadProfileCheckMain([]string{path}, &out, &diagnostic); code != 0 || diagnostic.Len() != 0 || !strings.Contains(out.String(), "2 entries") || !strings.Contains(out.String(), "have not been verified") {
		t.Fatal(code, out.String(), diagnostic.String())
	}
	out.Reset()
	diagnostic.Reset()
	if code := objectReadProfileCheckMain(nil, &out, &diagnostic); code != 2 || out.Len() != 0 {
		t.Fatal("usage accepted", code)
	}
	out.Reset()
	diagnostic.Reset()
	if err := os.WriteFile(file.Profiles[0].Identity.Certificate, []byte("synthetic certificate canary"), 0600); err != nil {
		t.Fatal(err)
	}
	if code := objectReadProfileCheckMain([]string{path}, &out, &diagnostic); code != 1 || out.Len() != 0 || !strings.Contains(diagnostic.String(), "BEGIN CERTIFICATE") || strings.Contains(diagnostic.String(), "canary") || strings.Contains(diagnostic.String(), path) {
		t.Fatal("unsafe/unclear certificate diagnostic", code, diagnostic.String())
	}
}
