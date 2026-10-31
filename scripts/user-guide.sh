#!/usr/bin/env bash
# Read-only navigation helper; no installation, configuration or task side effects.
set -euo pipefail
if [[ $# != 1 ]]; then
  echo 'Usage: bash scripts/user-guide.sh getting-started|rooms|agents|tasks|workflows|results|storage|recovery|settings' >&2
  exit 2
fi
case "$1" in
  getting-started|rooms|agents|tasks|workflows|results|storage|recovery|settings) topic=$1 ;;
  *) echo 'Unknown guide. Use a documented functionality name.' >&2; exit 2 ;;
esac
command -v curl >/dev/null || { echo 'curl is required; open docs/user/README.md instead.' >&2; exit 1; }
if ! curl --fail --silent --show-error --max-time 5 --output /dev/null 'http://127.0.0.1:8787/help/'; then
  echo 'Start your controller first, or read docs/user/README.md offline.' >&2
  exit 1
fi
printf 'Open http://127.0.0.1:8787/help/?doc=user-%s\n' "$topic"
