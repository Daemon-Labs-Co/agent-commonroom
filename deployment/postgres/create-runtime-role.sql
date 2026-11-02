-- Administrator connection; role name supplied with -v runtime_role=...
-- Set a password interactively with \password afterwards, not in tracked SQL.
\set ON_ERROR_STOP on
CREATE ROLE :"runtime_role" LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE
  NOREPLICATION NOBYPASSRLS;
