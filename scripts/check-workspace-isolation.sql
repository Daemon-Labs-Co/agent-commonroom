-- Proves, against a SCRATCH database, that workspace row-level security isolates
-- history for a restricted role. Not for production data: it refuses to run unless the
-- database name contains "scratch" or "test" and the events table is empty.
--
--   createdb commonroom_scratch
--   psql -v ON_ERROR_STOP=1 -d commonroom_scratch -f migrations/001.sql -f migrations/002.sql
--   psql -v ON_ERROR_STOP=1 -d commonroom_scratch -f scripts/check-workspace-isolation.sql
--   dropdb commonroom_scratch
--
-- Prints "workspace isolation checks passed" or fails with the first broken guarantee.
\set ON_ERROR_STOP on
DO $$ BEGIN
  IF current_database() !~ '(scratch|test)' THEN RAISE EXCEPTION 'refusing to run outside a scratch database'; END IF;
  IF EXISTS (SELECT 1 FROM events) THEN RAISE EXCEPTION 'refusing to run on a database that already has events'; END IF;
END $$;
CREATE ROLE iso_runtime LOGIN NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE;
GRANT USAGE ON SCHEMA public TO iso_runtime;
GRANT SELECT, INSERT ON events TO iso_runtime;
GRANT USAGE ON SEQUENCE events_sequence_seq TO iso_runtime;
SET ROLE iso_runtime;

CREATE FUNCTION pg_temp.put(ws text, label text) RETURNS void LANGUAGE plpgsql AS $$
BEGIN
  PERFORM set_config('app.workspace_id', ws, true);
  INSERT INTO events(id, kind, payload, previous_hash, hash) VALUES (label, 'check', '{}', '', label);
END $$;
CREATE FUNCTION pg_temp.visible(ws text) RETURNS text LANGUAGE plpgsql AS $$
BEGIN
  PERFORM set_config('app.workspace_id', ws, true);
  RETURN (SELECT coalesce(string_agg(id, ',' ORDER BY id), '') FROM events);
END $$;
CREATE FUNCTION pg_temp.expect_refused(stmt text, why text) RETURNS void LANGUAGE plpgsql AS $$
BEGIN
  EXECUTE stmt;
  RAISE EXCEPTION 'NOT REFUSED: %', why;
EXCEPTION WHEN raise_exception THEN RAISE;
          WHEN others THEN NULL;
END $$;

BEGIN; SELECT pg_temp.put('alpha', 'a1'); SELECT pg_temp.put('alpha', 'a2'); COMMIT;
BEGIN; SELECT pg_temp.put('beta', 'b1'); COMMIT;
DO $$ BEGIN
  IF pg_temp.visible('alpha') <> 'a1,a2' THEN RAISE EXCEPTION 'alpha sees %', pg_temp.visible('alpha'); END IF;
  IF pg_temp.visible('beta') <> 'b1' THEN RAISE EXCEPTION 'beta sees %', pg_temp.visible('beta'); END IF;
  IF pg_temp.visible('gamma') <> '' THEN RAISE EXCEPTION 'an empty workspace sees rows'; END IF;
END $$;
-- Outside a scoped transaction nothing is visible and nothing can be written.
RESET app.workspace_id;
DO $$ BEGIN
  IF (SELECT count(*) FROM events) <> 0 THEN RAISE EXCEPTION 'unscoped session sees rows'; END IF;
  PERFORM pg_temp.expect_refused($s$INSERT INTO events(id,kind,payload,previous_hash,hash) VALUES ('x','k','{}','','x')$s$, 'unscoped insert');
END $$;
-- A scoped session cannot write into another workspace, use a malformed id or an empty one.
BEGIN; SELECT set_config('app.workspace_id', 'alpha', true);
SELECT pg_temp.expect_refused($s$INSERT INTO events(workspace_id,id,kind,payload,previous_hash,hash) VALUES ('beta','x1','k','{}','','x1')$s$, 'cross-workspace insert');
ROLLBACK;
BEGIN; SELECT set_config('app.workspace_id', 'Not Valid', true);
SELECT pg_temp.expect_refused($s$INSERT INTO events(id,kind,payload,previous_hash,hash) VALUES ('x2','k','{}','','x2')$s$, 'malformed workspace id');
ROLLBACK;
-- History stays append-only for every workspace.
BEGIN; SELECT set_config('app.workspace_id', 'alpha', true);
SELECT pg_temp.expect_refused($s$UPDATE events SET kind = 'changed'$s$, 'update');
ROLLBACK;
BEGIN; SELECT set_config('app.workspace_id', 'alpha', true);
SELECT pg_temp.expect_refused($s$DELETE FROM events$s$, 'delete');
ROLLBACK;
RESET ROLE;
-- The migration owner still sees everything for administration and backups.
DO $$ BEGIN
  IF (SELECT count(*) FROM events) <> 3 THEN RAISE EXCEPTION 'owner view is incomplete'; END IF;
END $$;
DROP OWNED BY iso_runtime;
DROP ROLE iso_runtime;
SELECT 'workspace isolation checks passed' AS result;
