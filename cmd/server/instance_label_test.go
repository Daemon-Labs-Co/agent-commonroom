package main

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestInstanceLabelUsesMachineHostname(t *testing.T) {
	name, err := os.Hostname()
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(name)
	for _, method := range []string{"GET", "HEAD", "POST"} {
		r := httptest.NewRequest(method, "https://unrelated-website.example/instance-label.js", nil)
		w := httptest.NewRecorder()
		instanceLabel(w, r)
		if method == "POST" {
			if w.Code != 405 {
				t.Fatal(w.Code)
			}
			continue
		}
		if w.Code != 200 || w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal(w.Result())
		}
		if method == "HEAD" {
			if w.Body.Len() != 0 {
				t.Fatal("HEAD body")
			}
			continue
		}
		if !strings.Contains(w.Body.String(), "globalThis.commonroomServerHostname="+string(raw)) || strings.Contains(w.Body.String(), "unrelated-website") {
			t.Fatal(w.Body.String())
		}
	}
}
