// Command api is the Opero server entrypoint.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/davidnguyen2205/opero/backend/internal/attendance"
	"github.com/davidnguyen2205/opero/backend/internal/controlplane"
	"github.com/davidnguyen2205/opero/backend/internal/identity"
	"github.com/davidnguyen2205/opero/backend/internal/leave"
	"github.com/davidnguyen2205/opero/backend/internal/liveview"
	"github.com/davidnguyen2205/opero/backend/internal/platform/auth"
	"github.com/davidnguyen2205/opero/backend/internal/platform/config"
	"github.com/davidnguyen2205/opero/backend/internal/platform/db"
	"github.com/davidnguyen2205/opero/backend/internal/platform/httpserver"
	"github.com/davidnguyen2205/opero/backend/internal/roster"
	"github.com/davidnguyen2205/opero/backend/internal/stats"
	"github.com/davidnguyen2205/opero/backend/internal/tours"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server exited with error", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	controlPool, err := db.NewControlPlanePool(ctx, cfg.ControlPlaneDSN())
	if err != nil {
		return err
	}
	defer controlPool.Close()

	// Tenant pool resolver — maps each tenant's db_name to a cached pool. Used
	// by TenantMiddleware to scope tenant-data routes (identity, M2+).
	tenantResolver := db.NewTenantResolver(cfg.DSN)
	defer tenantResolver.Close()

	tokens := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTTTL)
	store := controlplane.NewStore(controlPool)
	provisioner := controlplane.NewProvisioner(controlPool, cfg.DSN, logger)
	cpService := controlplane.NewService(store, provisioner, tokens, cfg.TenantDBPrefix, logger)
	cpHandler := controlplane.NewHandler(cpService, logger)

	identityService := identity.NewService(logger, cpService) // cpService provisions logins
	identityHandler := identity.NewHandler(identityService, logger)

	rosterService := roster.NewService(logger, identityService) // resolves employee for /me/shifts
	rosterHandler := roster.NewHandler(rosterService, logger)

	attendanceService := attendance.NewService(logger, identityService) // resolves employee from user
	attendanceHandler := attendance.NewHandler(attendanceService, logger)

	// liveview owns no tables; it composes roster + attendance + identity.
	liveviewService := liveview.NewService(rosterService, attendanceService, identityService)
	liveviewHandler := liveview.NewHandler(liveviewService, logger)

	leaveService := leave.NewService(logger, identityService) // resolves employee for /me/leave
	leaveHandler := leave.NewHandler(leaveService, logger)

	// stats owns no tables; it composes roster + attendance + identity.
	statsService := stats.NewService(rosterService, attendanceService, identityService)
	statsHandler := stats.NewHandler(statsService, logger)

	toursService := tours.NewService(logger)
	toursHandler := tours.NewHandler(toursService, logger)

	api := &apiHandler{
		cp: cpHandler, id: identityHandler, rs: rosterHandler, at: attendanceHandler,
		lv: liveviewHandler, lv2: leaveHandler, st: statsHandler, tr: toursHandler,
	}

	handler := httpserver.New(httpserver.Deps{
		Logger:              logger,
		ControlPlane:        controlPool,
		API:                 api,
		Tokens:              tokens,
		TenantRegistry:      store,          // controlplane.Store: tenant_id -> db_name
		TenantPools:         tenantResolver, // db.TenantResolver: db_name -> pool
		CORSAllowedOrigins:  cfg.CORSAllowedOrigins,
		TenantRoutePrefixes: []string{"/departments", "/employees", "/roles", "/shifts", "/locations", "/attendance", "/live", "/leave", "/tours", "/me/shifts", "/me/leave", "/me/stats"},
	})

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("graceful shutdown failed", slog.Any("error", err))
		}
	}()

	logger.Info("starting http server", slog.String("addr", cfg.HTTPAddr))
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	logger.Info("server stopped")
	return nil
}
