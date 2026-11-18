-- Workspace storage (ADR 0004). Run by the migration owner with the other files in
-- this directory (`commonroom --migrate`). Idempotent; never rewrites or deletes
-- history. Existing events become workspace 'default'.
--
-- The runtime role must not own events and must not have BYPASSRLS or SUPERUSER;
-- the server refuses to start otherwise.
ALTER TABLE events ADD COLUMN IF NOT EXISTS workspace_id TEXT NOT NULL DEFAULT 'default';
-- New rows take the transaction's workspace. Without the setting, or with an empty one
-- (a pooled connection after a scoped transaction ends), an insert fails.
ALTER TABLE events ALTER COLUMN workspace_id SET DEFAULT NULLIF(current_setting('app.workspace_id'), '');
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'events_workspace_id_format' AND conrelid = 'public.events'::regclass) THEN
    ALTER TABLE events ADD CONSTRAINT events_workspace_id_format
      CHECK (workspace_id ~ '^(default|[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?)$');
  END IF;
END $$;
CREATE INDEX IF NOT EXISTS events_workspace_sequence ON events (workspace_id, sequence);
ALTER TABLE events ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS workspace_isolation ON events;
CREATE POLICY workspace_isolation ON events
  USING (workspace_id = NULLIF(current_setting('app.workspace_id', true), ''))
  WITH CHECK (workspace_id = NULLIF(current_setting('app.workspace_id', true), ''));
