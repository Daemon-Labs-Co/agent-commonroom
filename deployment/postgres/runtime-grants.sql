-- Connect explicitly to the migrated application database before applying this.
-- Supply -v database=... -v runtime_role=...; do not embed credentials here.
\set ON_ERROR_STOP on
BEGIN;
-- Fail before any grant if the connection selects the wrong database.
SELECT 1 / (current_database() = :'database')::integer AS target_database_check;
GRANT CONNECT ON DATABASE :"database" TO :"runtime_role";
REVOKE CREATE ON SCHEMA public FROM PUBLIC;
GRANT USAGE ON SCHEMA public TO :"runtime_role";
REVOKE ALL ON TABLE public.events FROM :"runtime_role";
GRANT SELECT, INSERT ON TABLE public.events TO :"runtime_role";
REVOKE ALL ON SEQUENCE public.events_sequence_seq FROM :"runtime_role";
GRANT USAGE ON SEQUENCE public.events_sequence_seq TO :"runtime_role";
COMMIT;
