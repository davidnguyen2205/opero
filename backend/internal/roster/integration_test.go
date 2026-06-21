//go:build integration

// Integration test for the roster store against a real tenant Postgres DB.
// Run with: go test -tags=integration ./internal/roster/...
// Provisions a throwaway tenant database (real CREATE DATABASE + tenant
// migrations 00001..00003) via the controlplane Provisioner, exercises the
// store, and drops it.
package roster_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/davidnguyen2205/opero/backend/internal/controlplane"
	"github.com/davidnguyen2205/opero/backend/internal/identity"
	"github.com/davidnguyen2205/opero/backend/internal/platform/config"
	"github.com/davidnguyen2205/opero/backend/internal/roster"
)

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func TestRosterStoreAgainstRealTenantDB(t *testing.T) {
	cfg := &config.Config{
		DBHost:             envOr("DB_HOST", "localhost"),
		DBPort:             envOr("DB_PORT", "5432"),
		DBUser:             envOr("DB_USER", "opero"),
		DBPassword:         envOr("DB_PASSWORD", "opero"),
		DBSSLMode:          envOr("DB_SSLMODE", "disable"),
		ControlPlaneDBName: envOr("CONTROLPLANE_DB_NAME", "opero_control"),
	}
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	admin, err := pgxpool.New(ctx, cfg.ControlPlaneDSN())
	if err != nil {
		t.Skipf("no control-plane pool: %v", err)
	}
	if err := admin.Ping(ctx); err != nil {
		admin.Close()
		t.Skipf("control-plane unreachable: %v", err)
	}

	dbName := "opero_rostertest_" + uuid.NewString()[:8]
	prov := controlplane.NewProvisioner(admin, cfg.DSN, logger)
	if err := prov.Create(ctx, dbName); err != nil {
		admin.Close()
		t.Fatalf("provision tenant db: %v", err)
	}

	pool, err := pgxpool.New(ctx, cfg.DSN(dbName))
	if err != nil {
		t.Fatalf("tenant pool: %v", err)
	}
	t.Cleanup(func() {
		pool.Close()
		_ = prov.Drop(ctx, dbName)
		admin.Close()
	})

	// A shift needs an employee (FK). Create one via the identity store.
	emp, err := identity.NewStore(pool).CreateEmployee(ctx, identity.CreateEmployeeInput{
		FullName: "Field Worker", EmploymentType: "full_time", Status: "active",
	})
	if err != nil {
		t.Fatalf("seed employee: %v", err)
	}

	store := roster.NewStore(pool)

	// location with lat/lng round-trip
	lat, lng := 10.7769, 106.7009
	loc, err := store.CreateLocation(ctx, roster.CreateLocationInput{Name: "HQ", Lat: &lat, Lng: &lng})
	if err != nil {
		t.Fatalf("CreateLocation: %v", err)
	}
	if loc.Lat == nil || *loc.Lat != lat {
		t.Errorf("lat round-trip failed: %v", loc.Lat)
	}

	start := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	end := start.Add(8 * time.Hour)

	shift, err := store.CreateShift(ctx, roster.CreateShiftInput{
		EmployeeID: emp.ID, LocationID: &loc.ID, StartsAt: start, EndsAt: end,
	}, "draft")
	if err != nil {
		t.Fatalf("CreateShift: %v", err)
	}
	if shift.Status != "draft" {
		t.Errorf("status = %q, want draft", shift.Status)
	}

	// DB CHECK rejects ends_at <= starts_at -> mapped to ErrValidation
	if _, err := store.CreateShift(ctx, roster.CreateShiftInput{
		EmployeeID: emp.ID, StartsAt: end, EndsAt: start,
	}, "draft"); !errors.Is(err, roster.ErrValidation) {
		t.Fatalf("bad time order err = %v, want ErrValidation", err)
	}

	// filter by employee
	byEmp, err := store.ListShifts(ctx, roster.ShiftFilter{EmployeeID: &emp.ID})
	if err != nil || len(byEmp) != 1 {
		t.Fatalf("ListShifts by employee: len=%d err=%v", len(byEmp), err)
	}

	// window filter: [start-1h, start+1h) should include the shift
	from := start.Add(-time.Hour)
	to := start.Add(time.Hour)
	inWindow, err := store.ListShifts(ctx, roster.ShiftFilter{From: &from, To: &to})
	if err != nil || len(inWindow) != 1 {
		t.Fatalf("ListShifts in-window: len=%d err=%v", len(inWindow), err)
	}
	// window entirely before the shift should exclude it
	emptyFrom := start.Add(-3 * time.Hour)
	emptyTo := start.Add(-2 * time.Hour)
	outWindow, err := store.ListShifts(ctx, roster.ShiftFilter{From: &emptyFrom, To: &emptyTo})
	if err != nil || len(outWindow) != 0 {
		t.Fatalf("ListShifts out-window: len=%d err=%v", len(outWindow), err)
	}

	// publish flips status
	pub, err := store.PublishShift(ctx, shift.ID)
	if err != nil || pub.Status != "published" {
		t.Fatalf("PublishShift: status=%q err=%v", pub.Status, err)
	}

	// partial update (notes only) must leave every other field intact —
	// guards the COALESCE-on-narg mapping in UpdateShift.
	newNotes := "rescheduled briefing"
	upd, err := store.UpdateShift(ctx, shift.ID, roster.UpdateShiftInput{Notes: &newNotes})
	if err != nil {
		t.Fatalf("UpdateShift (partial): %v", err)
	}
	if upd.Notes == nil || *upd.Notes != newNotes {
		t.Errorf("notes not updated: %v", upd.Notes)
	}
	if !upd.StartsAt.Equal(pub.StartsAt) || !upd.EndsAt.Equal(pub.EndsAt) {
		t.Errorf("times changed on partial update: %v..%v", upd.StartsAt, upd.EndsAt)
	}
	if upd.EmployeeID != emp.ID {
		t.Errorf("employee changed on partial update: %v", upd.EmployeeID)
	}
	if upd.LocationID == nil || *upd.LocationID != loc.ID {
		t.Errorf("location cleared on partial update: %v", upd.LocationID)
	}
	if upd.Status != "published" {
		t.Errorf("status changed by update (should only change via publish): %q", upd.Status)
	}

	// employee delete is blocked while a shift references them (ON DELETE RESTRICT)
	if err := identity.NewStore(pool).DeleteEmployee(ctx, emp.ID); !errors.Is(err, identity.ErrInUse) {
		t.Fatalf("delete employee with shift err = %v, want identity.ErrInUse", err)
	}

	// delete + not found
	if err := store.DeleteShift(ctx, shift.ID); err != nil {
		t.Fatalf("DeleteShift: %v", err)
	}
	if _, err := store.GetShift(ctx, shift.ID); !errors.Is(err, roster.ErrNotFound) {
		t.Fatalf("GetShift after delete err = %v, want ErrNotFound", err)
	}
}
