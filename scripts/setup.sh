#!/usr/bin/env bash
# Build the controller; optionally migrate or start using operator-provided URLs.
set -euo pipefail
usage() {
  echo 'Usage: bash scripts/setup.sh [--check|--build|--migrate|--start]'
  echo 'Default: --build. Provision restricted database roles before --migrate/start.'
  echo 'After starting: python3 scripts/check-installation.py (separate terminal).' 
}
mode=${1:---build}
if [[ $# -gt 1 ]]; then usage >&2; exit 2; fi
case "$mode" in
  --help|-h) usage; exit 0 ;;
  --check|--build|--migrate|--start) ;;
  *) usage >&2; exit 2 ;;
esac
cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.."
case "$(uname -s)" in
  Darwin|Linux) ;;
  *) echo 'Use macOS, Linux, or the Windows WSL2 wrapper.' >&2; exit 1 ;;
esac
command -v go >/dev/null || { echo 'Install Go 1.26+; see docs/installation.md.' >&2; exit 1; }
# Go enforces the go.mod toolchain requirement during the build.
go version
if [[ "$mode" == --check ]]; then
  echo 'Build prerequisites found. Database, browser and worker are separate checks.'
  exit 0
fi
if [[ "$mode" == --migrate && -z "${MIGRATION_DATABASE_URL:-}" ]]; then
  echo 'Set MIGRATION_DATABASE_URL for the separate migration owner.' >&2; exit 1
fi
if [[ "$mode" == --start && -z "${DATABASE_URL:-}" ]]; then
  echo 'Set DATABASE_URL for the restricted runtime role.' >&2; exit 1
fi
mkdir -p bin
go build -o bin/commonroom ./cmd/server
case "$mode" in
  --migrate) exec ./bin/commonroom --migrate ;;
  --start)
    unset MIGRATION_DATABASE_URL
    echo 'Starting controller. Open http://127.0.0.1:8787 (default address).'
    exec ./bin/commonroom ;;
  *) echo 'Built bin/commonroom. Next: database setup, --migrate, grants, --start.' ;;
esac
