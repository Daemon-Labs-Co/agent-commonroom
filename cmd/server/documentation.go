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
	{"mvc-organization", "MVC and commercial add-on responsibility mapping", "docs/technical/mvc-organization.md"},
	{"product-functionality", "Free core and commercial functionality inventory", "docs/technical/product-functionality.md"},
	{"documentation-policy", "Documentation audience contract", "docs/technical/documentation-policy.md"},
	{"user-docs", "User documentation by functionality", "docs/user/README.md"},
	{"technical-docs", "Technical input/output reference", "docs/technical/README.md"},
	{"user-agents", "Connect and check an agent", "docs/user/agents.md"},
	{"user-getting-started", "Install and open CommonRoom", "docs/user/getting-started.md"},
	{"user-recovery", "Stop, diagnose and recover work", "docs/user/recovery.md"},
	{"user-results", "Inspect outputs and accept work", "docs/user/results.md"},
	{"user-rooms", "Rooms and defined outcomes", "docs/user/rooms.md"},
	{"user-settings", "Settings, history and sharing", "docs/user/settings.md"},
	{"user-storage", "Storage limits and retained work", "docs/user/storage.md"},
	{"user-tasks", "Create and dispatch a task", "docs/user/tasks.md"},
	{"user-workflows", "Run a workflow and human gates", "docs/user/workflows.md"},
	{"technical-artifacts", "Artifact capture, manifests and reads", "docs/technical/artifacts.md"},
	{"technical-execution", "Isolated execution and workflows", "docs/technical/execution.md"},
	{"technical-identity", "Identity, settings and access", "docs/technical/identity.md"},
	{"technical-installation", "Controller setup and preflights", "docs/technical/installation.md"},
	{"technical-interface", "Browser and documentation interface", "docs/technical/interface.md"},
	{"technical-persistence", "Events, state projection and atomic operations", "docs/technical/persistence.md"},
	{"technical-publishing", "Repository publishing", "docs/technical/publishing.md"},
	{"technical-recovery", "Cancellation, cleanup and recovery", "docs/technical/recovery.md"},
	{"technical-storage", "Storage accounting and physical enforcement", "docs/technical/storage.md"},
	{"technical-tasks", "Tasks, execution records and acceptance", "docs/technical/tasks.md"},
	{"capture-worker-profiles", "Private capture worker profiles and owned service", "docs/technical/archive/capture-worker-profiles.md"},
	{"capture-authentication", "Authenticated private capture sessions", "docs/technical/archive/capture-authentication.md"},
	{"capture-supervision", "Owned isolated capture lifecycle", "docs/technical/archive/capture-supervision.md"},
	{"artifact-manifests", "Finite artifact manifests and immutable registry", "docs/technical/archive/artifact-manifests.md"},
	{"worker-storage-preflight", "Worker physical storage preflight", "docs/technical/archive/worker-storage-preflight.md"},
	{"physical-project-quota-observer", "Physical project-quota observations", "docs/technical/archive/physical-project-quota-observer.md"},
	{"reader-worker-profiles", "Reader worker profiles and service setup", "docs/technical/archive/reader-worker-profiles.md"},
	{"reader-authority-bridge", "Artifact-reader authority bridge", "docs/technical/archive/reader-authority-bridge.md"},
	{"controller-object-read-ledger", "Controller artifact-reader ledger", "docs/technical/archive/controller-object-read-ledger.md"},
	{"worker-object-read-journal", "Durable artifact-reader journal", "docs/technical/archive/worker-object-read-journal.md"},
	{"object-reader-supervision", "Isolated artifact reads and recovery", "docs/technical/archive/object-reader-supervision.md"},
	{"task-repository-links", "Task repository references (API)", "docs/technical/archive/task-repository-links.md"},
	{"task-definitions", "Task definitions and execution records", "docs/technical/archive/task-definitions.md"},
	{"task-value-inputs", "Accepted task inputs", "docs/technical/archive/task-value-inputs.md"},
	{"task-acceptance", "Task criterion review", "docs/technical/archive/task-acceptance.md"},
	{"task-progress", "Recorded task progress", "docs/technical/archive/task-progress.md"},
	{"room-outcome-review", "Human acceptance of room deliverables", "docs/technical/archive/room-outcome-review.md"},
	{"room-outcomes", "Room outcomes and acceptance criteria", "docs/technical/archive/room-outcomes.md"},
	{"task-delegations", "Bounded task delegation", "docs/technical/archive/task-delegations.md"},
	{"task-proposals", "Conversation task proposals", "docs/technical/archive/task-proposals.md"},
	{"governance", "Room coordination and authority", "docs/technical/archive/governance-appointments.md"},
	{"help-display", "Help window preference", "docs/technical/archive/help-display.md"},
	{"index", "Documentation home", "docs/README.md"},
	{"installation", "Installation and getting started", "docs/technical/archive/installation.md"},
	{"quick-start", "Platform quick starts and setup scripts", "docs/technical/archive/quick-start.md"},
	{"installation-check", "Check an installation", "docs/technical/archive/installation-check.md"},
	{"installation-options", "Installation component choices", "docs/technical/archive/installation-options.md"},
	{"extending", "Extending functionality", "docs/technical/archive/extending.md"},
	{"concepts", "Rooms, agents, profiles and workflows", "docs/technical/archive/concepts.md"},
	{"interface", "Interface preferences", "docs/technical/archive/interface.md"},
	{"agent-activity", "Agent activity and recovery", "docs/technical/archive/agent-activity.md"},
	{"history", "History and retention", "docs/technical/archive/history.md"},
	{"organization", "Company organization", "docs/technical/archive/organization.md"},
	{"workflow-templates", "Workflow templates", "docs/technical/archive/workflow-templates.md"},
	{"p0-acceptance-tests", "Remaining P0 acceptance tests", "docs/technical/archive/p0-acceptance-tests.md"},
	{"ssh", "SSH agent connection setup", "docs/technical/archive/agent-ssh.md"},
	{"server-structure", "Server structure and responsibilities", "docs/technical/archive/server-structure.md"},
	{"architecture", "Architecture", "docs/technical/archive/architecture.md"},
	{"connection-testing", "Testing connections", "docs/technical/archive/connection-testing.md"},
	{"connection-sign-in", "Isolated connection sign-in", "docs/technical/archive/connection-sign-in.md"},
	{"worker-credential-recovery", "Recover uncertain worker credentials", "docs/technical/archive/worker-credential-recovery.md"},
	{"engine-credential-custody", "Private engine credential custody", "docs/technical/archive/engine-credential-custody.md"},
	{"isolation", "Mandatory container execution", "docs/technical/archive/container-isolation.md"},
	{"development-workspaces", "Development workspace configuration", "docs/technical/archive/development-workspaces.md"},
	{"workflow-execution", "Running workflows", "docs/technical/archive/workflow-execution.md"},
	{"task-dispatch", "Sending tasks to agents", "docs/technical/archive/task-dispatch.md"},
	{"atomic-operations", "Atomic operations and safe retries", "docs/technical/archive/atomic-operations.md"},
	{"context-disclosure", "Context disclosure controls", "docs/technical/archive/context-disclosure.md"},
	{"storage-deadlines", "Bounded storage waits", "docs/technical/archive/storage-deadlines.md"},
	{"login-availability", "Login availability and session renewal", "docs/technical/archive/login-availability.md"},
	{"private-file-security", "Private file and lease safety", "docs/technical/archive/private-file-security.md"},
	{"provider-tool-separation", "Provider/tool credential separation", "docs/technical/archive/provider-tool-separation.md"},
	{"workspace-bootstrap", "Credential-free workspace bootstrap", "docs/technical/archive/workspace-bootstrap.md"},
	{"workspace-volumes", "Fresh workspace volumes and cleanup", "docs/technical/archive/workspace-volumes.md"},
	{"worker-audit-transport", "Private worker audit transport", "docs/technical/archive/worker-audit-transport.md"},
	{"audited-tool-session", "Shared audited tool session", "docs/technical/archive/audited-tool-session.md"},
	{"live-tool-acceptance", "Tool-only live acceptance evidence", "docs/technical/archive/live-tool-acceptance.md"},
	{"harness-release-admission", "Reviewed harness releases", "docs/technical/archive/harness-release-admission.md"},
	{"worker-idle-shutdown", "Worker idle shutdown guard", "docs/technical/archive/worker-idle-shutdown.md"},
	{"workspace-disposal", "Owner-reviewed workspace disposal", "docs/technical/archive/workspace-disposal.md"},
	{"workspace-disposal-binding", "Exact workspace disposal binding", "docs/technical/archive/workspace-disposal-binding.md"},
	{"storage-accounting", "Managed storage accounting", "docs/technical/archive/storage-accounting.md"},
	{"isolated-commands", "Bounded isolated Go commands", "docs/technical/archive/isolated-commands.md"},
	{"tool-container-runtime", "Constrained tool-container launch", "docs/technical/archive/tool-container-runtime.md"},
	{"engine-container-runtime", "Reviewed engine runtime", "docs/technical/archive/engine-container-runtime.md"},
	{"codex-engine-bootstrap", "Codex engine bootstrap", "docs/technical/archive/codex-engine-bootstrap.md"},
	{"codex-callback-driver", "Codex callback driver", "docs/technical/archive/codex-callback-driver.md"},
	{"worker-pair-shutdown", "Paired turn shutdown", "docs/technical/archive/worker-pair-shutdown.md"},
	{"worker-pair-assembly", "Worker pair assembly", "docs/technical/archive/worker-pair-assembly.md"},
	{"worker-pairs", "Paired worker lifecycle", "docs/technical/archive/worker-pairs.md"},
	{"container-quotas", "Managed container quotas", "docs/technical/archive/container-quotas.md"},
	{"execution-budgets", "Daily managed-turn budget", "docs/technical/archive/execution-budgets.md"},
	{"worker-jobs", "Free worker job safety and diagnostics", "docs/technical/archive/worker-jobs.md"},
	{"api", "HTTP API", "docs/technical/archive/api.md"},
	{"development", "Development", "docs/technical/archive/development.md"},
	{"operations", "Operations and backups", "docs/technical/archive/operations.md"},
	{"database-security", "Restricted database and audit recovery", "docs/technical/archive/database-security.md"},
	{"security-hardening", "Security hardening and limits", "docs/technical/archive/security-hardening.md"},
	{"debian", "Debian remote installation", "docs/technical/archive/deployment/debian.md"},
	{"nginx", "nginx and existing PostgreSQL", "docs/technical/archive/deployment/nginx-postgresql.md"},
	{"security", "Security policy", "SECURITY.md"},
	{"review", "Security and architecture review", "SECURITY_REVIEW.md"},
	{"contributing", "Contributing", "CONTRIBUTING.md"},
	{"spec", "Functional specification", "FUNCTIONAL_SPEC.md"},
	{"status", "Release readiness", "docs/technical/archive/release-readiness.md"},
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
			for _, asset := range []string{"user-agents.svg", "user-getting-started.svg", "user-recovery.svg", "user-results.svg", "user-rooms.svg", "user-settings.svg", "user-storage.svg", "user-tasks.svg", "user-workflows.svg"} {
				if r.URL.Path == "/help/image/"+asset {
					name, contentType = "docs/images/"+asset, "image/svg+xml"
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
