package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"time"

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

// verifyRuntimeDatabaseStartup bounds connection and privilege verification as
// one operation. Never log the raw connection error: drivers/backends may include
// credentials or untrusted server detail. Unknown privileges refuse startup.
func verifyRuntimeDatabaseStartup(ctx context.Context, ping func(context.Context) error, db catalogQuerier) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if ctx.Err() != nil || ping(ctx) != nil {
		return errors.New("Runtime database connection unavailable; inspect database administration logs")
	}
	if err := validateRuntimeDatabase(ctx, db); err != nil {
		return err
	}
	if ctx.Err() != nil {
		return errors.New("Runtime database startup verification timed out")
	}
	return nil
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

// workspaceStorageSQL reports whether the workspace migration is applied, whether
// row-level security is enabled on events, and whether the serving role could
// bypass it (superuser, BYPASSRLS or owner of the table).
const workspaceStorageSQL = `SELECT
 EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name='events' AND column_name='workspace_id'),
 COALESCE((SELECT relrowsecurity FROM pg_class WHERE oid='public.events'::regclass), false),
 COALESCE((SELECT rolsuper OR rolbypassrls FROM pg_roles WHERE rolname=current_user), true),
 COALESCE((SELECT pg_has_role(current_user, relowner, 'USAGE') FROM pg_class WHERE oid='public.events'::regclass), true)`

type workspaceStorage struct{ Present, Protected bool }

// verifyWorkspaceStorage fails closed: once the workspace column exists the
// database must actually enforce isolation against this role. Before the
// migration the previous single-history layout is still served unchanged.
func verifyWorkspaceStorage(ctx context.Context, db catalogQuerier) (workspaceStorage, error) {
	rows, err := db.Query(ctx, workspaceStorageSQL)
	if err != nil {
		return workspaceStorage{}, errors.New("workspace storage could not be verified")
	}
	defer rows.Close()
	var present, rls, bypass, owner bool
	if !rows.Next() || rows.Scan(&present, &rls, &bypass, &owner) != nil || rows.Err() != nil {
		return workspaceStorage{}, errors.New("workspace storage could not be verified")
	}
	if !present {
		return workspaceStorage{}, nil
	}
	if !rls || bypass || owner {
		return workspaceStorage{Present: true}, errors.New("workspace storage is not protected: enable row-level security on events and serve with a role that is not its owner and has no BYPASSRLS or SUPERUSER")
	}
	return workspaceStorage{Present: true, Protected: true}, nil
}

// applyMigrations runs every NNN.sql in dir in name order. Files are idempotent.
func applyMigrations(ctx context.Context, dir string, exec func(context.Context, string) error) error {
	names, err := filepath.Glob(filepath.Join(dir, "[0-9][0-9][0-9].sql"))
	if err != nil || len(names) == 0 {
		return errors.New("migration source unavailable")
	}
	sort.Strings(names)
	for _, name := range names {
		raw, err := os.ReadFile(name)
		if err != nil {
			return errors.New("migration source unavailable")
		}
		if err = exec(ctx, string(raw)); err != nil {
			return errors.New("migration failed; inspect database administration logs")
		}
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
	if err = applyMigrations(ctx, "migrations", func(ctx context.Context, sql string) error {
		_, err := db.Exec(ctx, sql)
		return err
	}); err != nil {
		return err
	}
	return nil
}
