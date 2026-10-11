package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type documentationEntry struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Path  string `json:"path"`
}

var documentationCatalog = []documentationEntry{
	{"index", "Documentation home", "docs/README.md"},
	{"installation", "Installation and getting started", "README.md"},
	{"concepts", "Rooms, agents, profiles and workflows", "docs/concepts.md"},
	{"ssh", "SSH agent connection setup", "docs/agent-ssh.md"},
	{"architecture", "Architecture", "docs/architecture.md"},
	{"api", "HTTP API", "docs/api.md"},
	{"development", "Development", "docs/development.md"},
	{"operations", "Operations and backups", "docs/operations.md"},
	{"debian", "Debian remote installation", "docs/deployment/debian.md"},
	{"nginx", "nginx and existing PostgreSQL", "docs/deployment/nginx-postgresql.md"},
	{"security", "Security policy", "SECURITY.md"},
	{"review", "Security and architecture review", "SECURITY_REVIEW.md"},
	{"contributing", "Contributing", "CONTRIBUTING.md"},
	{"spec", "Functional specification", "FUNCTIONAL_SPEC.md"},
	{"status", "Historical development notes", "DEVELOPMENT_STATUS.md"},
	{"license", "GPLv3 license", "LICENSE"},
}

// documentationHandler serves only the explicit public documentation catalog.
// Private configuration, arbitrary paths, and symlinked files are never served.
func documentationHandler(root string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if r.URL.Path == "/help/catalog" {
			w.Header().Set("Content-Type", "application/json")
			if r.Method == http.MethodGet {
				json.NewEncoder(w).Encode(documentationCatalog)
			}
			return
		}
		name, contentType := "", "text/plain; charset=utf-8"
		switch r.URL.Path {
		case "/help/", "/help":
			name, contentType = "web/help.html", "text/html; charset=utf-8"
		case "/help/reader.js":
			name, contentType = "web/help-reader.js", "text/javascript; charset=utf-8"
		case "/help/reader.css":
			name, contentType = "web/help-reader.css", "text/css; charset=utf-8"
		default:
			for _, entry := range documentationCatalog {
				if r.URL.Path == "/help/document/"+entry.ID {
					name = entry.Path
					break
				}
			}
			for _, asset := range []string{"room.jpg", "workflow.jpg", "connections.jpg"} {
				if r.URL.Path == "/help/image/"+asset {
					name, contentType = "docs/screenshots/"+asset, "image/jpeg"
				}
			}
		}
		if name == "" {
			http.NotFound(w, r)
			return
		}
		// Check every component, so a directory symlink cannot point at private data.
		path := root
		for _, part := range strings.Split(name, "/") {
			path = filepath.Join(path, part)
			info, err := os.Lstat(path)
			if err != nil || info.Mode()&os.ModeSymlink != 0 {
				http.NotFound(w, r)
				return
			}
		}
		f, err := os.Open(path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer f.Close()
		info, err := f.Stat()
		if err != nil || !info.Mode().IsRegular() {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", contentType)
		http.ServeContent(w, r, info.Name(), info.ModTime(), f)
	})
}
