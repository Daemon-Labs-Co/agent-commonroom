#!/usr/bin/env bash
# Verifies workspace storage against a REAL PostgreSQL server you own: the row-level security
# policy, the scoped database, per-workspace hash chains, the upgrade of existing history and the
# workspace registry. It creates a throwaway database and role, runs everything, and drops both.
#
# It never touches an existing database. Do not point it at a server you do not administer.
#
#   PGHOST=127.0.0.1 PGPORT=5432 PGUSER=postgres PGPASSWORD=... bash scripts/check-workspace-database.sh
#
# Needs: psql, go, and an administrator login that may create databases and roles.
set -euo pipefail
cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.."
command -v psql >/dev/null && command -v go >/dev/null || { echo 'psql and go are required.' >&2; exit 2; }
[[ -n "${PGUSER:-}" ]] || { echo 'Set PGUSER (and PGHOST, PGPORT, PGPASSWORD) to an administrator login.' >&2; exit 2; }
suffix="$$_$(date +%s)"
db="commonroom_scratch_${suffix}"
role="rt_scratch_${suffix}"
admin=(psql -v ON_ERROR_STOP=1 -q -X)
cleanup() {
  "${admin[@]}" -d postgres -c "DROP DATABASE IF EXISTS \"$db\" WITH (FORCE)" >/dev/null 2>&1 || true
  "${admin[@]}" -d postgres -c "DROP ROLE IF EXISTS \"$role\"" >/dev/null 2>&1 || true
}
trap cleanup EXIT
"${admin[@]}" -d postgres -c "CREATE DATABASE \"$db\""
"${admin[@]}" -d postgres -c "CREATE ROLE \"$role\" LOGIN NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE"
run_go() {
  SCRATCH_PHASE="$1" SCRATCH_DATABASE_URL="postgres://${role}@${PGHOST:-127.0.0.1}:${PGPORT:-5432}/${db}?sslmode=disable" \
    go test -tags scratchdb ./cmd/server -run 'Scratch' -count=1
}
grant() {
  "${admin[@]}" -d "$db" -c "GRANT USAGE ON SCHEMA public TO \"$role\"" -c "GRANT SELECT,INSERT ON events TO \"$role\"" \
    -c "GRANT USAGE ON SEQUENCE events_sequence_seq TO \"$role\""
}
echo "== 1/4 previous layout (migration 001 only)"
"${admin[@]}" -d "$db" -f migrations/001.sql 2>/dev/null
grant
run_go pre
echo "== 2/4 apply migrations 002 and 003 (twice, to prove they are idempotent)"
for _ in 1 2; do "${admin[@]}" -d "$db" -f migrations/001.sql -f migrations/002.sql -f migrations/003.sql 2>/dev/null; done
"${admin[@]}" -d "$db" -c "GRANT SELECT,INSERT,UPDATE ON workspaces TO \"$role\""
echo "== 3/4 isolation proof for a restricted role"
# The proof script needs an empty events table, so it runs in a second scratch database.
proof="${db}_proof"
trap 'cleanup; "${admin[@]}" -d postgres -c "DROP DATABASE IF EXISTS \"$proof\" WITH (FORCE)" >/dev/null 2>&1 || true' EXIT
"${admin[@]}" -d postgres -c "CREATE DATABASE \"$proof\""
"${admin[@]}" -d "$proof" -f migrations/001.sql -f migrations/002.sql 2>/dev/null
"${admin[@]}" -d "$proof" -f scripts/check-workspace-isolation.sql | tail -3
echo "== 4/4 application behaviour on the migrated database"
run_go post
echo "All workspace database checks passed."
