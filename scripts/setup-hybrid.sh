#!/usr/bin/env bash
# Run on the Linux server; preserve existing nginx, Docker and PostgreSQL state.
set -euo pipefail
[[ "$(uname -s)" == Linux ]] || { echo 'Run this on the Linux server.' >&2; exit 1; }
if [[ ${1:-} == --help || ${1:-} == -h ]]; then
  echo 'Usage: bash scripts/setup-hybrid.sh [--check|--build|--migrate|--start]'
  echo 'Configure nginx, PostgreSQL roles and the isolated worker using docs/quick-start.md.'
  exit 0
fi
for tool in nginx docker psql ssh; do
  command -v "$tool" >/dev/null || { echo "Missing $tool; follow docs/quick-start.md." >&2; exit 1; }
done
echo 'Tools found. Rootless limits, DB privileges, TLS and nginx configuration require separate validation.'
exec bash "$(dirname -- "${BASH_SOURCE[0]}")/setup-linux.sh" "$@"
