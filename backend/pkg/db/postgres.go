package db

import (
	"context"
	"embed"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/postgres/*.sql
var postgresMigrations embed.FS

func openPostgres(url string) (*Store, error) {
	config, err := pgx.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("invalid DATABASE_URL")
	}
	if os.Getenv("VERCEL") != "" && (config.TLSConfig == nil || len(config.Fallbacks) > 0) {
		return nil, fmt.Errorf("DATABASE_URL must require TLS on Vercel")
	}
	// Transaction poolers cannot guarantee a connection for prepared statements.
	config.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	config.RuntimeParams["timezone"] = "UTC"
	config.ConnectTimeout = 10 * time.Second
	raw := stdlib.OpenDB(*config)
	raw.SetMaxOpenConns(5)
	raw.SetMaxIdleConns(2)
	raw.SetConnMaxLifetime(5 * time.Minute)
	raw.SetConnMaxIdleTime(time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err = raw.PingContext(ctx); err != nil {
		raw.Close()
		return nil, fmt.Errorf("PostgreSQL connection failed (check credentials, network and TLS)")
	}
	tx, err := raw.BeginTx(ctx, nil)
	if err != nil {
		raw.Close()
		return nil, err
	}
	defer tx.Rollback()
	// A transaction-scoped lock prevents concurrent cold starts racing migrations.
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(72600420); CREATE SCHEMA IF NOT EXISTS loop; REVOKE ALL ON SCHEMA loop FROM PUBLIC; CREATE TABLE IF NOT EXISTS loop.schema_version(version INTEGER PRIMARY KEY)`); err != nil {
		tx.Rollback()
		raw.Close()
		return nil, err
	}
	var applied bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM loop.schema_version WHERE version=1)`).Scan(&applied); err != nil {
		tx.Rollback()
		raw.Close()
		return nil, err
	}
	if !applied {
		migration, readErr := postgresMigrations.ReadFile("migrations/postgres/000001_initial.sql")
		if readErr != nil {
			tx.Rollback()
			raw.Close()
			return nil, readErr
		}
		if _, err = tx.ExecContext(ctx, string(migration)); err != nil {
			tx.Rollback()
			raw.Close()
			return nil, fmt.Errorf("PostgreSQL migration: %w", err)
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO loop.schema_version VALUES(1)`); err != nil {
			tx.Rollback()
			raw.Close()
			return nil, err
		}
	}
	if err = tx.Commit(); err != nil {
		tx.Rollback()
		raw.Close()
		return nil, err
	}
	return &Store{DB: raw, postgres: true}, nil
}
