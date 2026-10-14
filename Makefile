.PHONY: run build test test-race coverage check secrets

run:
	go run ./cmd/server

build:
	go build -o bin/commonroom ./cmd/server

TEST_PARALLEL ?= 8

test:
	go test -parallel $(TEST_PARALLEL) -timeout 30s ./...

test-race:
	go test -race -parallel $(TEST_PARALLEL) -timeout 30s ./...

coverage:
	@mkdir -p coverage
	go test -parallel $(TEST_PARALLEL) -timeout 30s -coverprofile=coverage/go.out ./...
	go tool cover -func=coverage/go.out

check: test-race
	go vet ./...
	node --check docker/agent/entrypoint.cjs
	node --check docker/tools/tool-call.cjs
	node --test tests/tool-call.test.cjs
	node --test tests/tool-image.test.cjs
	node --test tests/agent-entrypoint.test.cjs tests/credential-export.test.cjs tests/workspace-readiness.test.cjs
	node --check web/human-attention.js
	node --test tests/human-attention.test.cjs
	node --check web/squad-menus.js
	node --test tests/squad-menus.test.cjs
	node --check web/message-identity.js
	node --test tests/message-identity.test.cjs
	node --check web/security-incident.js
	node --test tests/security-incident.test.cjs
	node --check web/agent-recovery.js
	node --test tests/agent-recovery.test.cjs
	node --check web/job-diagnostics.js
	node --test tests/job-diagnostics.test.cjs
	node --check web/background-work.js
	node --test tests/background-work.test.cjs
	node --check web/agent-status.js
	node --test tests/agent-status.test.cjs
	node --check web/app.js
	node --test tests/operation-retry.test.cjs tests/context-disclosure.test.cjs
	node --check web/features.js
	node --test tests/squad-count.test.cjs tests/task-dispatch.test.cjs
	node --check web/squad-profiles.js
	node --check web/agent-invitation.js
	node --check web/help-icons.js
	node tests/help-icons.test.cjs
	node --check web/feature-flag-tabs.js
	node tests/feature-flag-tabs.test.cjs
	node --check web/settings-tabs.js
	node tests/settings-tabs.test.cjs
	node --check web/organization-chart.js
	node tests/organization-chart.test.cjs
	node --check web/viewport-alignment.js
	node --check web/overlay-layout.js
	node tests/overlay-layout.test.cjs
	node --check web/human-profile-details.js
	node tests/human-profile-details.test.cjs
	node --check web/room-sections.js
	node tests/room-sections.test.cjs
	node tests/room-presentation.test.cjs
	node --test tests/room-message-order.test.cjs
	node --check web/room-viewport.js
	node --test tests/room-scroll-stability.test.cjs
	node --check web/feature-switches.js
	node --test tests/ui-switches-viewport.test.cjs
	node --test tests/workflow-execution.test.cjs
	node --check web/workflows.js
	node --check web/workflow-templates.js
	node tests/workflow-templates.test.cjs
	node --check web/provider-auth.js
	node tests/provider-auth.test.cjs
	node --check web/connection-testing.js
	node tests/connection-testing.test.cjs
	node --check web/connection-editor.js
	node tests/connection-editor.test.cjs
	node --check web/profile-images.js
	node tests/profile-images.test.cjs
	node --check web/agent-safety.js
	node tests/agent-safety.test.cjs
	node --check web/history-presentation.js
	node --test tests/history-presentation.test.cjs
	node --check web/history.js
	node --check web/history-date-groups.js
	node tests/history-date-groups.test.cjs
	node --test tests/history-pagination.test.cjs
	node --check web/help.js
	node tests/help-preferences.test.cjs
	node --check web/help-reader.js
	node tests/help-reader.test.cjs

# Gitleaks 8.30.1 was used for the initial repository scan.
secrets:
	@if command -v gitleaks >/dev/null 2>&1; then gitleaks dir --redact --no-banner .; elif [ -x bin/gitleaks ]; then bin/gitleaks dir --redact --no-banner .; else echo "Install Gitleaks before running this check." >&2; exit 1; fi
