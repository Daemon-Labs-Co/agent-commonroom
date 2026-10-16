#!/usr/bin/env bash
set -euo pipefail
[[ "$(uname -s)" == Darwin ]] || { echo 'This wrapper requires macOS.' >&2; exit 1; }
exec bash "$(dirname -- "${BASH_SOURCE[0]}")/setup.sh" "$@"
