package main

import (
	"encoding/json"
	"net/http"
	"os"
)

// instanceLabel identifies the application host, independently of proxy/DNS URLs.
func instanceLabel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	name, err := os.Hostname()
	if err != nil || name == "" {
		name = "Server name unavailable"
	}
	raw, _ := json.Marshal(name)
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if r.Method == http.MethodHead {
		return
	}
	w.Write([]byte("globalThis.commonroomServerHostname=" + string(raw) + ";document.querySelector('#instance-address').textContent=globalThis.commonroomServerHostname;"))
}
