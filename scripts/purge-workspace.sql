-- Permanently removes a DELETED workspace's history. Administrator-only: it needs the
-- table owner or a superuser, because the application's role can never delete history.
--
--   1. Mark the workspace deleted through the platform API first (it stops serving at once).
--   2. Take a backup if you may need it:  pg_dump -Fc -f backup.dump <database>
--   3. psql -v ON_ERROR_STOP=1 -v workspace=acme -d <database> -f scripts/purge-workspace.sql
--   4. Remove its local files: rm -r .data/workspaces/acme
--
-- It refuses the default workspace and any workspace that is not marked deleted. The
-- append-only trigger is disabled only inside this transaction and is re-enabled before
-- commit; the registry row stays as a tombstone so the id is never reused.
\set ON_ERROR_STOP on
BEGIN;
SELECT set_config('purge.workspace', :'workspace', true);
DO $$
DECLARE target text := current_setting('purge.workspace');
BEGIN
  IF target = '' OR target = 'default' THEN
    RAISE EXCEPTION 'refusing to purge workspace "%"', target;
  END IF;
  IF NOT EXISTS (SELECT 1 FROM workspaces WHERE id = target AND status = 'deleted') THEN
    RAISE EXCEPTION 'workspace "%" is not marked deleted in the registry', target;
  END IF;
END $$;
ALTER TABLE events DISABLE TRIGGER immutable_events;
WITH removed AS (DELETE FROM events WHERE workspace_id = current_setting('purge.workspace') RETURNING 1)
  SELECT count(*) AS events_removed FROM removed;
ALTER TABLE events ENABLE TRIGGER immutable_events;
COMMIT;
