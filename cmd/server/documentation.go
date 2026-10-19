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
	{"help-display", "Help window preference", "docs/help-display.md"},
	{"index", "Documentation home", "docs/README.md"},
	{"installation", "Installation and getting started", "docs/installation.md"},
	{"quick-start", "Platform quick starts and setup scripts", "docs/quick-start.md"},
	{"installation-check", "Check an installation", "docs/installation-check.md"},
	{"installation-options", "Installation component choices", "docs/installation-options.md"},
	{"extending", "Extending functionality", "docs/extending.md"},
	{"concepts", "Rooms, agents, profiles and workflows", "docs/concepts.md"},
	{"interface", "Interface preferences", "docs/interface.md"},
	{"agent-activity", "Agent activity and recovery", "docs/agent-activity.md"},
	{"history", "History and retention", "docs/history.md"},
	{"organization", "Company organization", "docs/organization.md"},
	{"workflow-templates", "Workflow templates", "docs/workflow-templates.md"},
	{"p0-acceptance-tests", "Remaining P0 acceptance tests", "docs/p0-acceptance-tests.md"},
	{"ssh", "SSH agent connection setup", "docs/agent-ssh.md"},
	{"server-structure", "Server structure and responsibilities", "docs/server-structure.md"},
	{"architecture", "Architecture", "docs/architecture.md"},
	{"connection-testing", "Testing connections", "docs/connection-testing.md"},
	{"connection-sign-in", "Isolated connection sign-in", "docs/connection-sign-in.md"},
	{"worker-credential-recovery", "Recover uncertain worker credentials", "docs/worker-credential-recovery.md"},
	{"engine-credential-custody", "Private engine credential custody", "docs/engine-credential-custody.md"},
	{"isolation", "Mandatory container execution", "docs/container-isolation.md"},
	{"development-workspaces", "Development workspace configuration", "docs/development-workspaces.md"},
	{"workflow-execution", "Running workflows", "docs/workflow-execution.md"},
	{"task-dispatch", "Sending tasks to agents", "docs/task-dispatch.md"},
	{"atomic-operations", "Atomic operations and safe retries", "docs/atomic-operations.md"},
	{"context-disclosure", "Context disclosure controls", "docs/context-disclosure.md"},
	{"storage-deadlines", "Bounded storage waits", "docs/storage-deadlines.md"},
	{"login-availability", "Login availability and session renewal", "docs/login-availability.md"},
	{"private-file-security", "Private file and lease safety", "docs/private-file-security.md"},
	{"provider-tool-separation", "Provider/tool credential separation", "docs/provider-tool-separation.md"},
	{"workspace-bootstrap", "Credential-free workspace bootstrap", "docs/workspace-bootstrap.md"},
	{"workspace-volumes", "Fresh workspace volumes and cleanup", "docs/workspace-volumes.md"},
	{"worker-audit-transport", "Private worker audit transport", "docs/worker-audit-transport.md"},
	{"audited-tool-session", "Shared audited tool session", "docs/audited-tool-session.md"},
	{"live-tool-acceptance", "Tool-only live acceptance evidence", "docs/live-tool-acceptance.md"},
	{"worker-idle-shutdown", "Worker idle shutdown guard", "docs/worker-idle-shutdown.md"},
	{"workspace-disposal", "Owner-reviewed workspace disposal", "docs/workspace-disposal.md"},
	{"workspace-disposal-binding", "Exact workspace disposal binding", "docs/workspace-disposal-binding.md"},
	{"storage-accounting", "Managed storage accounting", "docs/storage-accounting.md"},
	{"isolated-commands", "Bounded isolated Go commands", "docs/isolated-commands.md"},
	{"tool-container-runtime", "Constrained tool-container launch", "docs/tool-container-runtime.md"},
	{"engine-container-runtime", "Reviewed engine runtime", "docs/engine-container-runtime.md"},
	{"codex-engine-bootstrap", "Codex engine bootstrap", "docs/codex-engine-bootstrap.md"},
	{"codex-callback-driver", "Codex callback driver", "docs/codex-callback-driver.md"},
	{"worker-pair-shutdown", "Paired turn shutdown", "docs/worker-pair-shutdown.md"},
	{"worker-pair-assembly", "Worker pair assembly", "docs/worker-pair-assembly.md"},
	{"worker-pairs", "Paired worker lifecycle", "docs/worker-pairs.md"},
	{"container-quotas", "Managed container quotas", "docs/container-quotas.md"},
	{"execution-budgets", "Daily managed-turn budget", "docs/execution-budgets.md"},
	{"worker-jobs", "Free worker job safety and diagnostics", "docs/worker-jobs.md"},
	{"api", "HTTP API", "docs/api.md"},
	{"development", "Development", "docs/development.md"},
	{"operations", "Operations and backups", "docs/operations.md"},
	{"database-security", "Restricted database and audit recovery", "docs/database-security.md"},
	{"security-hardening", "Security hardening and limits", "docs/security-hardening.md"},
	{"debian", "Debian remote installation", "docs/deployment/debian.md"},
	{"nginx", "nginx and existing PostgreSQL", "docs/deployment/nginx-postgresql.md"},
	{"security", "Security policy", "SECURITY.md"},
	{"review", "Security and architecture review", "SECURITY_REVIEW.md"},
	{"contributing", "Contributing", "CONTRIBUTING.md"},
	{"spec", "Functional specification", "FUNCTIONAL_SPEC.md"},
	{"status", "Release readiness", "docs/release-readiness.md"},
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
