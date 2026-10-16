#!/usr/bin/env bash
set -euo pipefail
[[ "$(uname -s)" == Linux ]] || { echo 'This wrapper requires Linux (including WSL2).' >&2; exit 1; }
exec bash "$(dirname -- "${BASH_SOURCE[0]}")/setup.sh" "$@"
