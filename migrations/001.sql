CREATE TABLE IF NOT EXISTS events (
 sequence BIGSERIAL PRIMARY KEY,
 id TEXT NOT NULL UNIQUE,
 kind TEXT NOT NULL,
 payload JSONB NOT NULL,
 committed_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
 previous_hash TEXT NOT NULL,
 hash TEXT NOT NULL
);
CREATE OR REPLACE FUNCTION reject_history_mutation() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN RAISE EXCEPTION 'History is append-only'; END $$;
DROP TRIGGER IF EXISTS immutable_events ON events;
CREATE TRIGGER immutable_events BEFORE UPDATE OR DELETE OR TRUNCATE ON events FOR EACH STATEMENT EXECUTE FUNCTION reject_history_mutation();
