// Command provision creates a tenant's logical database and runs the tenant
// migrations against it.
//
// Usage:
//
//	provision <db_name>
//
// Full tenant onboarding (registering the tenant + admin user) happens through
// the signup API, which reuses the same provisioner. This command is for ops /
// scripted use against a known database name.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/davidnguyen2205/opero/backend/internal/controlplane"
	"github.com/davidnguyen2205/opero/backend/internal/platform/config"
	"github.com/davidnguyen2205/opero/backend/internal/platform/db"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if len(os.Args) < 2 {
		logger.Error("usage: provision <db_name>")
		os.Exit(2)
	}
	if err := run(context.Background(), logger, os.Args[1]); err != nil {
		logger.Error("provision failed", slog.Any("error", err))
		os.Exit(1)
	}
	logger.Info("provisioned", slog.String("db", os.Args[1]))
}

func run(ctx context.Context, logger *slog.Logger, dbName string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	pool, err := db.NewControlPlanePool(ctx, cfg.ControlPlaneDSN())
	if err != nil {
		return fmt.Errorf("connect control-plane: %w", err)
	}
	defer pool.Close()

	return controlplane.NewProvisioner(pool, cfg.DSN, logger).Create(ctx, dbName)
}
