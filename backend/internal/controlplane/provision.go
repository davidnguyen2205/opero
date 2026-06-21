package controlplane

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // database/sql driver "pgx" for goose
	"github.com/pressly/goose/v3"

	dbassets "github.com/davidnguyen2205/opero/backend/db"
)

// Provisioner creates and tears down tenant databases. Creating a database
// cannot run inside a transaction, so CREATE/DROP DATABASE are issued on the
// control-plane (admin) pool, and tenant migrations are then applied to the new
// database via goose.
type Provisioner struct {
	admin    *pgxpool.Pool
	buildDSN func(dbName string) string
	logger   *slog.Logger
}

func NewProvisioner(admin *pgxpool.Pool, buildDSN func(dbName string) string, logger *slog.Logger) *Provisioner {
	return &Provisioner{admin: admin, buildDSN: buildDSN, logger: logger}
}

// Create creates the logical database (idempotently) and runs tenant
// migrations against it. dbName must already be a safe identifier; it is also
// quoted defensively via pgx.Identifier.
func (p *Provisioner) Create(ctx context.Context, dbName string) error {
	ident := pgx.Identifier{dbName}.Sanitize()

	// CREATE DATABASE cannot be parameterized and cannot run in a transaction.
	if _, err := p.admin.Exec(ctx, "CREATE DATABASE "+ident); err != nil {
		// 42P04 = duplicate_database; treat an existing DB as already provisioned.
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != "42P04" {
			return fmt.Errorf("create database %s: %w", dbName, err)
		}
	}

	if err := p.migrate(ctx, dbName); err != nil {
		return fmt.Errorf("migrate tenant db %s: %w", dbName, err)
	}
	return nil
}

// Drop removes a tenant database. Best-effort; used to clean up after a failed
// provisioning. It will fail if connections are still open to the database.
func (p *Provisioner) Drop(ctx context.Context, dbName string) error {
	ident := pgx.Identifier{dbName}.Sanitize()
	if _, err := p.admin.Exec(ctx, "DROP DATABASE IF EXISTS "+ident); err != nil {
		return fmt.Errorf("drop database %s: %w", dbName, err)
	}
	return nil
}

func (p *Provisioner) migrate(ctx context.Context, dbName string) error {
	// Skip entirely if no tenant migrations exist yet (none until M2).
	hasMigrations, err := dirHasSQL(dbassets.FS, dbassets.TenantDir)
	if err != nil {
		return err
	}
	if !hasMigrations {
		p.logger.InfoContext(ctx, "no tenant migrations to apply", slog.String("db", dbName))
		return nil
	}

	sqlDB, err := sql.Open("pgx", p.buildDSN(dbName))
	if err != nil {
		return fmt.Errorf("open tenant db: %w", err)
	}
	defer func() { _ = sqlDB.Close() }()

	goose.SetBaseFS(dbassets.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose dialect: %w", err)
	}
	if err := goose.UpContext(ctx, sqlDB, dbassets.TenantDir); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}

func dirHasSQL(fsys fs.FS, dir string) (bool, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return false, fmt.Errorf("read migrations dir %s: %w", dir, err)
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			return true, nil
		}
	}
	return false, nil
}
