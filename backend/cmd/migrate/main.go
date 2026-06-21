// Command migrate applies database migrations.
//
// Usage:
//
//	migrate controlplane   # migrate the shared control-plane DB
//	migrate tenants        # fan out tenant migrations across every tenant DB
//	migrate all            # control-plane, then all tenants (default)
//
// Tenant failures are reported per tenant; the command exits non-zero if any
// tenant migration fails, after attempting all of them.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	dbassets "github.com/davidnguyen2205/opero/backend/db"
	controlplanedb "github.com/davidnguyen2205/opero/backend/gen/sqlc/controlplane"
	"github.com/davidnguyen2205/opero/backend/internal/platform/config"
)

func main() {
	target := "all"
	if len(os.Args) > 1 {
		target = os.Args[1]
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(context.Background(), logger, target); err != nil {
		logger.Error("migrate failed", slog.Any("error", err))
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *slog.Logger, target string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	switch target {
	case "controlplane":
		return migrateControlPlane(ctx, logger, cfg)
	case "tenants":
		return migrateTenants(ctx, logger, cfg)
	case "all":
		if err := migrateControlPlane(ctx, logger, cfg); err != nil {
			return err
		}
		return migrateTenants(ctx, logger, cfg)
	default:
		return fmt.Errorf("unknown target %q (want: controlplane | tenants | all)", target)
	}
}

func migrateControlPlane(ctx context.Context, logger *slog.Logger, cfg *config.Config) error {
	logger.InfoContext(ctx, "migrating control-plane database")
	return runGoose(ctx, cfg.ControlPlaneDSN(), dbassets.ControlPlaneDir)
}

func migrateTenants(ctx context.Context, logger *slog.Logger, cfg *config.Config) error {
	hasSQL, err := dirHasSQL(dbassets.FS, dbassets.TenantDir)
	if err != nil {
		return err
	}
	if !hasSQL {
		logger.InfoContext(ctx, "no tenant migrations to apply yet")
		return nil
	}

	pool, err := pgxpool.New(ctx, cfg.ControlPlaneDSN())
	if err != nil {
		return fmt.Errorf("connect control-plane: %w", err)
	}
	defer pool.Close()

	tenants, err := controlplanedb.New(pool).ListTenants(ctx)
	if err != nil {
		return fmt.Errorf("list tenants: %w", err)
	}

	var failed int
	for _, t := range tenants {
		if err := runGoose(ctx, cfg.DSN(t.DbName), dbassets.TenantDir); err != nil {
			failed++
			logger.ErrorContext(ctx, "tenant migration failed",
				slog.String("slug", t.Slug), slog.String("db", t.DbName), slog.Any("error", err))
			continue
		}
		logger.InfoContext(ctx, "tenant migrated",
			slog.String("slug", t.Slug), slog.String("db", t.DbName))
	}

	logger.InfoContext(ctx, "tenant migration complete",
		slog.Int("total", len(tenants)), slog.Int("failed", failed))
	if failed > 0 {
		return fmt.Errorf("%d of %d tenant migrations failed", failed, len(tenants))
	}
	return nil
}

func runGoose(ctx context.Context, dsn, dir string) error {
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer func() { _ = sqlDB.Close() }()

	goose.SetBaseFS(dbassets.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose dialect: %w", err)
	}
	if err := goose.UpContext(ctx, sqlDB, dir); err != nil {
		return fmt.Errorf("goose up %s: %w", dir, err)
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
