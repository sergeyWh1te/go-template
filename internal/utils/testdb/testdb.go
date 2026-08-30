// Package testdb wires tests to a real Postgres instance.
//
// Tests that need a database call New. When no database is configured the test
// is skipped rather than failed, so `go test ./...` still passes on a laptop
// with nothing running; CI sets TEST_PG_DSN and therefore always runs them.
//
//	docker compose up -d postgres   # or: make test-db-up
//	make test-integration
package testdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file" // file:// migration source
	"github.com/jmoiron/sqlx"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx driver for database/sql
)

// DSNEnv names the environment variable holding the test database connection
// string. Unset means "no database available" — see New.
const DSNEnv = "TEST_PG_DSN"

// DefaultDSN matches the postgres service in docker-compose.yml, so a developer
// who ran `make test-db-up` needs no further setup. Assembled rather than
// written out so it does not read as a hardcoded credential.
var DefaultDSN = fmt.Sprintf(
	"postgres://%s:%s@127.0.0.1:5432/%s?sslmode=disable",
	"postgres", "postgres", "master",
)

const connectTimeout = 5 * time.Second

// New returns a database handle with every migration applied and all tables
// truncated, so each test starts from a known-empty schema.
//
// The test is skipped when no database is reachable. It fails, rather than
// skips, when a database is configured but unusable — a broken CI service
// container must not look like a pass.
func New(t *testing.T) *sqlx.DB {
	t.Helper()

	dsn, ok := os.LookupEnv(DSNEnv)
	if !ok {
		// Nothing configured: fall back to the compose default, but treat an
		// unreachable database as "not set up" instead of a failure.
		if !reachable(DefaultDSN) {
			t.Skipf("no test database: set %s or run `make test-db-up`", DSNEnv)
		}
		dsn = DefaultDSN
	}

	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Errorf("close test database: %v", closeErr)
		}
	})

	ctx, cancel := context.WithTimeout(t.Context(), connectTimeout)
	defer cancel()

	if pingErr := db.PingContext(ctx); pingErr != nil {
		t.Fatalf("ping test database at %s: %v", dsn, pingErr)
	}

	migrateUp(t, db.DB, dsn)
	Truncate(t, db)

	return db
}

// migrateUp applies db/migrations, so the test schema and the production schema
// can never drift apart — there is no second copy of the DDL to keep in sync.
func migrateUp(t *testing.T, db *sql.DB, dsn string) {
	t.Helper()

	driver, err := migratepg.WithInstance(db, &migratepg.Config{})
	if err != nil {
		t.Fatalf("build migrate driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+migrationsDir(t), "postgres", driver)
	if err != nil {
		t.Fatalf("open migrations: %v", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("apply migrations to %s: %v", dsn, err)
	}
}

// migrationsDir resolves db/migrations from this file's own location, so tests
// work regardless of the package they are invoked from.
func migrationsDir(t *testing.T) string {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not resolve the testdb package directory")
	}

	// internal/utils/testdb/testdb.go -> repo root
	root := filepath.Join(filepath.Dir(thisFile), "..", "..", "..")

	dir, err := filepath.Abs(filepath.Join(root, "db", "migrations"))
	if err != nil {
		t.Fatalf("resolve migrations dir: %v", err)
	}

	if _, statErr := os.Stat(dir); statErr != nil {
		t.Fatalf("migrations dir %s: %v", dir, statErr)
	}

	return dir
}

// Truncate empties every application table and restarts identity sequences, so
// ids are predictable from one test to the next. schema_migrations is left
// alone: dropping it would re-run every migration on the next call.
func Truncate(t *testing.T, db *sqlx.DB) {
	t.Helper()

	var tables []string
	const q = `select tablename from pg_tables
		where schemaname = 'public' and tablename <> 'schema_migrations'`

	if err := db.Select(&tables, q); err != nil {
		t.Fatalf("list tables: %v", err)
	}

	if len(tables) == 0 {
		return
	}

	quoted := make([]string, 0, len(tables))
	for _, table := range tables {
		quoted = append(quoted, fmt.Sprintf("public.%q", table))
	}

	// One statement so foreign keys never block the truncation order.
	stmt := "truncate table " + strings.Join(quoted, ", ") + " restart identity cascade"
	if _, err := db.ExecContext(t.Context(), stmt); err != nil {
		t.Fatalf("truncate tables: %v", err)
	}
}

// reachable reports whether a database answers at dsn within connectTimeout.
func reachable(dsn string) bool {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return false
	}
	defer func() { _ = db.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	return db.PingContext(ctx) == nil
}
