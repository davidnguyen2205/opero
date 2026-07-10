// Command platform-user creates an Opero platform user in the control-plane DB.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
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
	role := flag.String("role", "super_admin", "platform user role: super_admin, support, or ops")
	flag.Parse()

	if *email == "" {
		return fmt.Errorf("-email is required")
	}

	// The password is read from the PLATFORM_USER_PASSWORD env var, or from
	// stdin if unset — never a CLI flag, which would leak into `ps` output and
	// shell history.
	password, err := readPassword()
	if err != nil {
		return err
	}
	if password == "" {
		return fmt.Errorf("password is required (set PLATFORM_USER_PASSWORD or pipe it on stdin)")
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

	id, err := svc.CreatePlatformUser(ctx, *email, password, *role)
	if err != nil {
		return err
	}
	fmt.Printf("created platform user %s (%s)\n", id, *role)
	return nil
}

// readPassword returns the platform user's password from the
// PLATFORM_USER_PASSWORD env var, or the first line of stdin if the var is
// unset. Kept off the command line so it never appears in `ps` or shell history.
func readPassword() (string, error) {
	if v := os.Getenv("PLATFORM_USER_PASSWORD"); v != "" {
		return v, nil
	}
	fmt.Fprint(os.Stderr, "Password: ")
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && line == "" {
		return "", fmt.Errorf("read password from stdin: %w", err)
	}
	return strings.TrimRight(line, "\r\n"), nil
}
