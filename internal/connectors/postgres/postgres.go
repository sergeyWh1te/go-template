package postgres

import (
	"context"
	"fmt"
	"sync"
	"time"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/sergeyWh1te/go-template/internal/env"
)

var (
	dbMu sync.Mutex
	db   *sqlx.DB
)

const (
	MaxOpenConns = 25
	// Idle connections are a subset of open ones: a value above MaxOpenConns is
	// silently clamped by database/sql, so keeping it below says what it means.
	MaxIdleConns    = 10
	ConnMaxLifetime = 30 * time.Minute
	ConnMaxIdleTime = 5 * time.Minute

	pingTimeout = 5 * time.Second
)

// Connect returns the shared database handle, opening it on first use. A failed
// attempt is not cached: sync.Once would keep the broken handle forever and drop
// the error on every later call, which looks like a working connection.
func Connect(config *env.PgConfig) (*sqlx.DB, error) {
	dbMu.Lock()
	defer dbMu.Unlock()

	if db != nil {
		return db, nil
	}

	// The extended protocol (pgx's default) is left on: it keeps prepared
	// statements, which are both faster and immune to injection through
	// parameters. simple_protocol=true is only needed behind a connection pooler
	// in transaction mode, such as pgbouncer — add it there, not here.
	conf, parseErr := pgx.ParseConfig(
		fmt.Sprintf(`host=%s port=%d user=%s password=%s dbname=%s sslmode=%s`,
			config.Host, config.Port, config.Username, config.Password, config.Database, config.SslMode),
	)
	if parseErr != nil {
		return nil, fmt.Errorf("could not parse postgres config: %w", parseErr)
	}

	conf.RuntimeParams = map[string]string{
		"standard_conforming_strings": "on",
	}

	client := sqlx.NewDb(stdlib.OpenDB(*conf), "pgx")

	client.SetMaxOpenConns(MaxOpenConns)
	client.SetMaxIdleConns(MaxIdleConns)
	client.SetConnMaxLifetime(ConnMaxLifetime)
	client.SetConnMaxIdleTime(ConnMaxIdleTime)

	pingCtx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	if pingErr := client.PingContext(pingCtx); pingErr != nil {
		// Drop the handle so a later call dials again instead of handing back a
		// connection that never worked.
		_ = client.Close()

		return nil, fmt.Errorf("could not ping postgres at %s:%d: %w", config.Host, config.Port, pingErr)
	}

	db = client

	return db, nil
}

// Close closes the shared handle and clears it, so a later Connect dials a new
// one instead of handing back the closed handle. database/sql waits for queries
// already in flight, so callers should close only after their request handlers
// have drained.
//
// Safe to call when nothing was ever opened.
func Close() error {
	dbMu.Lock()
	defer dbMu.Unlock()

	if db == nil {
		return nil
	}

	err := db.Close()
	db = nil

	if err != nil {
		return fmt.Errorf("could not close postgres connection: %w", err)
	}

	return nil
}
