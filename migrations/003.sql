-- Workspace registry (ADR 0004, 0007). The only table that spans workspaces; the
-- serving role reads and updates it, never deletes (deleted workspaces keep a
-- tombstone row). Idempotent.
CREATE TABLE IF NOT EXISTS workspaces (
  id TEXT PRIMARY KEY CHECK (id ~ '^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$'),
  name TEXT NOT NULL CHECK (char_length(name) BETWEEN 1 AND 100),
  status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'deleted')),
  limits JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
