//go:build integration

// Integration test for the identity store against a real tenant Postgres DB.
// Run with: go test -tags=integration ./internal/identity/...
// Provisions a throwaway tenant database (real CREATE DATABASE + tenant
// migrations) via the controlplane Provisioner, exercises the store, and drops it.
package identity_test

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
)

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func TestStoreCRUDAgainstRealTenantDB(t *testing.T) {
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

	dbName := "opero_idtest_" + uuid.NewString()[:8]
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

	store := identity.NewStore(pool)

	// department
	dept, err := store.CreateDepartment(ctx, identity.CreateDepartmentInput{Name: "Operations"})
	if err != nil {
		t.Fatalf("CreateDepartment: %v", err)
	}

	// employee in that department, with a nullable date + email round-trip
	hired := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	email := "bob@itest.test"
	emp, err := store.CreateEmployee(ctx, identity.CreateEmployeeInput{
		FullName:       "Bob Ops",
		EmploymentType: "full_time",
		Status:         "active",
		DepartmentID:   &dept.ID,
		Email:          &email,
		HiredAt:        &hired,
	})
	if err != nil {
		t.Fatalf("CreateEmployee: %v", err)
	}
	if emp.DepartmentID == nil || *emp.DepartmentID != dept.ID {
		t.Errorf("department_id mismatch: %v", emp.DepartmentID)
	}
	if emp.HiredAt == nil || !emp.HiredAt.Equal(hired) {
		t.Errorf("hired_at round-trip failed: %v", emp.HiredAt)
	}
	if emp.Email == nil || *emp.Email != email {
		t.Errorf("email round-trip failed: %v", emp.Email)
	}

	// filtered list
	list, err := store.ListEmployees(ctx, identity.EmployeeFilter{DepartmentID: &dept.ID})
	if err != nil || len(list) != 1 {
		t.Fatalf("ListEmployees by dept: len=%d err=%v", len(list), err)
	}

	// update (clears nothing, sets title)
	title := "Lead"
	upd, err := store.UpdateEmployee(ctx, emp.ID, identity.UpdateEmployeeInput{Title: &title})
	if err != nil || upd.Title == nil || *upd.Title != "Lead" {
		t.Fatalf("UpdateEmployee: title=%v err=%v", upd.Title, err)
	}

	// role: create, unique-name conflict, assign to employee
	role, err := store.CreateRole(ctx, identity.CreateRoleInput{Name: "Guide", Permissions: []string{"roster.read"}})
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}
	if len(role.Permissions) != 1 || role.Permissions[0] != "roster.read" {
		t.Errorf("permissions round-trip failed: %v", role.Permissions)
	}
	if _, err := store.CreateRole(ctx, identity.CreateRoleInput{Name: "guide"}); !errors.Is(err, identity.ErrConflict) {
		t.Fatalf("duplicate role name err = %v, want ErrConflict", err)
	}
	assigned, err := store.UpdateEmployee(ctx, emp.ID, identity.UpdateEmployeeInput{RoleID: &role.ID})
	if err != nil {
		t.Fatalf("assign role: %v", err)
	}
	if assigned.RoleID == nil || *assigned.RoleID != role.ID {
		t.Errorf("role_id not assigned: %v", assigned.RoleID)
	}

	// deleting the role clears the employee's role_id (ON DELETE SET NULL)
	if err := store.DeleteRole(ctx, role.ID); err != nil {
		t.Fatalf("DeleteRole: %v", err)
	}
	cleared, err := store.GetEmployee(ctx, emp.ID)
	if err != nil {
		t.Fatalf("GetEmployee after role delete: %v", err)
	}
	if cleared.RoleID != nil {
		t.Errorf("role_id should be null after role delete, got %v", cleared.RoleID)
	}

	// delete + not found
	if err := store.DeleteEmployee(ctx, emp.ID); err != nil {
		t.Fatalf("DeleteEmployee: %v", err)
	}
	if _, err := store.GetEmployee(ctx, emp.ID); !errors.Is(err, identity.ErrNotFound) {
		t.Fatalf("GetEmployee after delete err = %v, want ErrNotFound", err)
	}
	if err := store.DeleteEmployee(ctx, emp.ID); !errors.Is(err, identity.ErrNotFound) {
		t.Fatalf("second DeleteEmployee err = %v, want ErrNotFound", err)
	}
}
