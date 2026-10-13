package main

import (
	"context"
	"errors"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const runtimePrivilegeSQL = `SELECT
 r.rolsuper OR r.rolcreaterole OR r.rolcreatedb OR r.rolreplication OR r.rolbypassrls
 OR EXISTS(SELECT 1 FROM pg_auth_members WHERE member=r.oid)
 OR pg_has_role(current_user,c.relowner,'MEMBER')
 OR pg_has_role(current_user,n.nspowner,'MEMBER')
 OR pg_has_role(current_user,d.datdba,'MEMBER'),
 has_table_privilege(current_user,c.oid,'SELECT'),
 has_table_privilege(current_user,c.oid,'INSERT'),
 has_sequence_privilege(current_user,'public.events_sequence_seq','USAGE'),
 has_table_privilege(current_user,c.oid,'UPDATE,DELETE,TRUNCATE,TRIGGER') OR has_sequence_privilege(current_user,'public.events_sequence_seq','UPDATE'),
 has_schema_privilege(current_user,n.oid,'CREATE')
 FROM pg_roles r JOIN pg_class c ON c.oid='public.events'::regclass
 JOIN pg_namespace n ON n.oid=c.relnamespace
 JOIN pg_database d ON d.datname=current_database()
 WHERE r.rolname=current_user`

type catalogQuerier interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func runtimeDatabaseConfig(url string) (*pgxpool.Config, error) {
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, errors.New("invalid runtime database configuration")
	}
	config.ConnConfig.RuntimeParams["search_path"] = "pg_catalog,public"
	return config, nil
}

// validateRuntimeDatabase prevents the serving process from using a migration
// owner or an account that can mutate audit rows/protections. This has no off
// switch: a source fork CAN bypass it, BUT undermines append-only guarantees.
func validateRuntimeDatabase(ctx context.Context, db catalogQuerier) error {
	rows, err := db.Query(ctx, runtimePrivilegeSQL)
	if err != nil {
		return errors.New("runtime database privileges could not be verified")
	}
	defer rows.Close()
	if !rows.Next() {
		return errors.New("initialize the audit schema with --migrate before serving")
	}
	var broad, read, insert, sequence, mutate, create bool
	if rows.Scan(&broad, &read, &insert, &sequence, &mutate, &create) != nil || rows.Err() != nil {
		return errors.New("runtime database privileges could not be verified")
	}
	if broad || mutate || create || !read || !insert || !sequence {
		return errors.New("use a dedicated SELECT/INSERT audit runtime role with sequence USAGE; migration/owner/DDL privileges are forbidden")
	}
	return nil
}

// runDatabaseMigration uses a distinct administrator-supplied URL. It never
// falls back to runtime credentials, and emits no credential-bearing errors.
func runDatabaseMigration(ctx context.Context) error {
	url := os.Getenv("MIGRATION_DATABASE_URL")
	if url == "" {
		return errors.New("MIGRATION_DATABASE_URL is required for --migrate")
	}
	db, err := pgxpool.New(ctx, url)
	if err != nil {
		return errors.New("migration database configuration unavailable")
	}
	defer db.Close()
	raw, err := os.ReadFile("migrations/001.sql")
	if err != nil {
		return errors.New("migration source unavailable")
	}
	if _, err = db.Exec(ctx, string(raw)); err != nil {
		return errors.New("migration failed; inspect database administration logs")
	}
	return nil
}
