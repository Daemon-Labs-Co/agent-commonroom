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
	node --check web/app.js
	node --check web/features.js
	node --check web/workflows.js
	node --check web/provider-auth.js
	node --check web/connection-editor.js
	node tests/connection-editor.test.cjs
	node --check web/history.js
	node --check web/help.js
	node --check web/help-reader.js
	node tests/help-reader.test.cjs

# Gitleaks 8.30.1 was used for the initial repository scan.
secrets:
	@if command -v gitleaks >/dev/null 2>&1; then gitleaks dir --redact --no-banner .; elif [ -x bin/gitleaks ]; then bin/gitleaks dir --redact --no-banner .; else echo "Install Gitleaks before running this check." >&2; exit 1; fi
