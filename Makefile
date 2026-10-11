.PHONY: run build test check secrets

run:
	go run ./cmd/server

build:
	go build -o bin/commonroom ./cmd/server

test:
	go test -race ./...

check: test
	go vet ./...
	node --check web/app.js
	node --check web/features.js
	node --check web/workflows.js
	node --check web/provider-auth.js

# Gitleaks 8.30.1 was used for the initial repository scan.
secrets:
	@if command -v gitleaks >/dev/null 2>&1; then gitleaks dir --redact --no-banner .; elif [ -x bin/gitleaks ]; then bin/gitleaks dir --redact --no-banner .; else echo "Install Gitleaks before running this check." >&2; exit 1; fi
