// Command platform-user creates an Opero platform user in the control-plane DB.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/davidnguyen2205/opero/backend/internal/controlplane"
	"github.com/davidnguyen2205/opero/backend/internal/platform/auth"
	"github.com/davidnguyen2205/opero/backend/internal/platform/config"
	"github.com/davidnguyen2205/opero/backend/internal/platform/db"
)

type noopProvisioner struct{}

func (noopProvisioner) Create(context.Context, string) error { return nil }
func (noopProvisioner) Drop(context.Context, string) error   { return nil }

func main() {
	if err := run(); err != nil {
		slog.Error("create platform user failed", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	email := flag.String("email", "", "platform user email")
	password := flag.String("password", "", "platform user password")
	role := flag.String("role", "super_admin", "platform user role: super_admin, support, or ops")
	flag.Parse()

	if *email == "" || *password == "" {
		return fmt.Errorf("-email and -password are required")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.NewControlPlanePool(ctx, cfg.ControlPlaneDSN())
	if err != nil {
		return err
	}
	defer pool.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	store := controlplane.NewStore(pool)
	tokens := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTTTL)
	svc := controlplane.NewService(store, noopProvisioner{}, tokens, cfg.TenantDBPrefix, logger)

	id, err := svc.CreatePlatformUser(ctx, *email, *password, *role)
	if err != nil {
		return err
	}
	fmt.Printf("created platform user %s (%s)\n", id, *role)
	return nil
}
