//go:build integration

// Full-lifecycle integration test: boots the real composite API + httpserver
// (auth + TenantMiddleware), signs up a tenant, then creates an employee over
// HTTP and asserts the row lands in the resolved tenant database.
// Run with: go test -tags=integration ./cmd/api/...
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	dbassets "github.com/davidnguyen2205/opero/backend/db"
	"github.com/davidnguyen2205/opero/backend/internal/attendance"
	"github.com/davidnguyen2205/opero/backend/internal/controlplane"
	"github.com/davidnguyen2205/opero/backend/internal/identity"
	"github.com/davidnguyen2205/opero/backend/internal/liveview"
	"github.com/davidnguyen2205/opero/backend/internal/platform/auth"
	"github.com/davidnguyen2205/opero/backend/internal/platform/config"
	platdb "github.com/davidnguyen2205/opero/backend/internal/platform/db"
	"github.com/davidnguyen2205/opero/backend/internal/platform/httpserver"
	"github.com/davidnguyen2205/opero/backend/internal/roster"
)

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func TestFullLifecycleCreateEmployee(t *testing.T) {
	cfg := &config.Config{
		DBHost:             envOr("DB_HOST", "localhost"),
		DBPort:             envOr("DB_PORT", "5432"),
		DBUser:             envOr("DB_USER", "opero"),
		DBPassword:         envOr("DB_PASSWORD", "opero"),
		DBSSLMode:          envOr("DB_SSLMODE", "disable"),
		ControlPlaneDBName: envOr("CONTROLPLANE_DB_NAME", "opero_control"),
		TenantDBPrefix:     "opero_e2e_",
	}
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	pool, err := pgxpool.New(ctx, cfg.ControlPlaneDSN())
	if err != nil {
		t.Skipf("no control-plane pool: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("control-plane unreachable: %v", err)
	}
	ensureControlPlaneMigrated(t, cfg.ControlPlaneDSN())

	tokens := auth.NewTokenManager("e2e-secret", "opero", time.Hour)
	store := controlplane.NewStore(pool)
	prov := controlplane.NewProvisioner(pool, cfg.DSN, logger)
	cpSvc := controlplane.NewService(store, prov, tokens, cfg.TenantDBPrefix, logger)
	idSvc := identity.NewService(logger, cpSvc)
	rsSvc := roster.NewService(logger, idSvc)
	atSvc := attendance.NewService(logger, idSvc, rsSvc)
	resolver := platdb.NewTenantResolver(cfg.DSN)

	api := &apiHandler{
		cp: controlplane.NewHandler(cpSvc, logger),
		id: identity.NewHandler(idSvc, logger),
		at: attendance.NewHandler(atSvc, logger),
	}
	srv := httpserver.New(httpserver.Deps{
		Logger:              logger,
		ControlPlane:        pool,
		API:                 api,
		Tokens:              tokens,
		TenantRegistry:      store,
		TenantPools:         resolver,
		TenantRoutePrefixes: []string{"/departments", "/employees", "/roles", "/shifts", "/locations", "/attendance"},
	})

	slug := "e2e-" + uuid.NewString()[:8]
	dbName := cfg.TenantDBPrefix + strings.ReplaceAll(slug, "-", "_")
	t.Cleanup(func() {
		resolver.Close()
		_, _ = pool.Exec(ctx, "DELETE FROM tenants WHERE slug=$1", slug)
		_ = prov.Drop(ctx, dbName)
		pool.Close()
	})

	// signup -> provisions the tenant DB (with identity schema) and returns a token
	signupBody := `{"company_name":"E2E Co","slug":"` + slug + `","admin_email":"e2e@x.test","admin_password":"password1"}`
	rec := do(srv, http.MethodPost, "/auth/signup", "", signupBody)
	if rec.Code != http.StatusCreated {
		t.Fatalf("signup status=%d body=%s", rec.Code, rec.Body.String())
	}
	var auth1 struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &auth1); err != nil || auth1.Token == "" {
		t.Fatalf("token parse: %v body=%s", err, rec.Body.String())
	}

	// no token -> 401 (auth gate)
	if rec := do(srv, http.MethodPost, "/employees", "", `{"full_name":"X","employment_type":"part_time"}`); rec.Code != http.StatusUnauthorized {
		t.Fatalf("create employee without token status=%d, want 401", rec.Code)
	}

	// create employee over the full stack (auth -> TenantMiddleware -> handler)
	rec = do(srv, http.MethodPost, "/employees", auth1.Token, `{"full_name":"Liz Field","employment_type":"part_time"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create employee status=%d body=%s", rec.Code, rec.Body.String())
	}

	// assert the row exists in the resolved tenant DB (not the control plane)
	tdb, err := sql.Open("pgx", cfg.DSN(dbName))
	if err != nil {
		t.Fatalf("open tenant db: %v", err)
	}
	defer func() { _ = tdb.Close() }()
	var count int
	if err := tdb.QueryRowContext(ctx, "SELECT count(*) FROM employees WHERE full_name='Liz Field'").Scan(&count); err != nil {
		t.Fatalf("query tenant employees: %v", err)
	}
	if count != 1 {
		t.Fatalf("tenant DB employee count = %d, want 1", count)
	}
}

// TestM4AttendanceLifecycle exercises the full M4 chain over HTTP against real
// Postgres: signup -> create employee -> provision staff login -> staff JWT ->
// check-in (idempotent) -> check-out, plus that an account with no employee
// (the admin) cannot check in.
func TestM4AttendanceLifecycle(t *testing.T) {
	cfg := &config.Config{
		DBHost:             envOr("DB_HOST", "localhost"),
		DBPort:             envOr("DB_PORT", "5432"),
		DBUser:             envOr("DB_USER", "opero"),
		DBPassword:         envOr("DB_PASSWORD", "opero"),
		DBSSLMode:          envOr("DB_SSLMODE", "disable"),
		ControlPlaneDBName: envOr("CONTROLPLANE_DB_NAME", "opero_control"),
		TenantDBPrefix:     "opero_m4_",
	}
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	pool, err := pgxpool.New(ctx, cfg.ControlPlaneDSN())
	if err != nil {
		t.Skipf("no control-plane pool: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("control-plane unreachable: %v", err)
	}
	ensureControlPlaneMigrated(t, cfg.ControlPlaneDSN())

	tokens := auth.NewTokenManager("m4-secret", "opero", time.Hour)
	store := controlplane.NewStore(pool)
	prov := controlplane.NewProvisioner(pool, cfg.DSN, logger)
	cpSvc := controlplane.NewService(store, prov, tokens, cfg.TenantDBPrefix, logger)
	idSvc := identity.NewService(logger, cpSvc)
	rsSvc := roster.NewService(logger, idSvc)
	atSvc := attendance.NewService(logger, idSvc, rsSvc)
	resolver := platdb.NewTenantResolver(cfg.DSN)
	api := &apiHandler{
		cp: controlplane.NewHandler(cpSvc, logger),
		id: identity.NewHandler(idSvc, logger),
		at: attendance.NewHandler(atSvc, logger),
	}
	srv := httpserver.New(httpserver.Deps{
		Logger:              logger,
		ControlPlane:        pool,
		API:                 api,
		Tokens:              tokens,
		TenantRegistry:      store,
		TenantPools:         resolver,
		TenantRoutePrefixes: []string{"/employees", "/attendance"},
	})

	slug := "m4-" + uuid.NewString()[:8]
	dbName := cfg.TenantDBPrefix + strings.ReplaceAll(slug, "-", "_")
	t.Cleanup(func() {
		resolver.Close()
		_, _ = pool.Exec(ctx, "DELETE FROM tenants WHERE slug=$1", slug)
		_ = prov.Drop(ctx, dbName)
		pool.Close()
	})

	mustJSON := func(b []byte, v any) {
		t.Helper()
		if err := json.Unmarshal(b, v); err != nil {
			t.Fatalf("json unmarshal: %v body=%s", err, b)
		}
	}

	// signup (admin)
	rec := do(srv, http.MethodPost, "/auth/signup", "",
		`{"company_name":"M4 Co","slug":"`+slug+`","admin_email":"m4@x.test","admin_password":"password1"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("signup status=%d body=%s", rec.Code, rec.Body.String())
	}
	var signup struct {
		Token  string `json:"token"`
		Tenant struct {
			Id string `json:"id"`
		} `json:"tenant"`
	}
	mustJSON(rec.Body.Bytes(), &signup)
	adminToken := signup.Token
	tenantID, err := uuid.Parse(signup.Tenant.Id)
	if err != nil {
		t.Fatalf("parse tenant id: %v", err)
	}

	// create employee (admin)
	rec = do(srv, http.MethodPost, "/employees", adminToken, `{"full_name":"Field Bob","employment_type":"freelance"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create employee status=%d body=%s", rec.Code, rec.Body.String())
	}
	var emp struct {
		Id string `json:"id"`
	}
	mustJSON(rec.Body.Bytes(), &emp)

	// admin has no employee record -> cannot check in
	if rec := do(srv, http.MethodPost, "/attendance/check-in", adminToken,
		`{"client_id":"`+uuid.NewString()+`"}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("admin check-in status=%d, want 400", rec.Code)
	}

	// provision a login for the employee (admin)
	rec = do(srv, http.MethodPost, "/employees/"+emp.Id+"/login", adminToken, `{"email":"bob@x.test","password":"password1"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("provision login status=%d body=%s", rec.Code, rec.Body.String())
	}
	var staff struct {
		Id   string `json:"id"`
		Role string `json:"role"`
	}
	mustJSON(rec.Body.Bytes(), &staff)
	if staff.Role != "employee" {
		t.Errorf("staff role = %q, want employee", staff.Role)
	}
	staffUserID, err := uuid.Parse(staff.Id)
	if err != nil {
		t.Fatalf("parse staff id: %v", err)
	}

	// mint a staff JWT and check in
	staffToken, _, err := tokens.Issue(staffUserID, tenantID, staff.Role, time.Now())
	if err != nil {
		t.Fatalf("issue staff token: %v", err)
	}
	cid := uuid.NewString()
	body := `{"client_id":"` + cid + `","lat":10.77,"lng":106.7,"photo_url":"https://photos.example/x.jpg"}`

	rec = do(srv, http.MethodPost, "/attendance/check-in", staffToken, body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("check-in status=%d body=%s", rec.Code, rec.Body.String())
	}
	var att struct {
		Id     string `json:"id"`
		Status string `json:"status"`
	}
	mustJSON(rec.Body.Bytes(), &att)
	if att.Status != "checked_in" {
		t.Errorf("status = %q, want checked_in", att.Status)
	}

	// idempotent replay -> 200, same record
	rec = do(srv, http.MethodPost, "/attendance/check-in", staffToken, body)
	if rec.Code != http.StatusOK {
		t.Fatalf("replay status=%d, want 200", rec.Code)
	}
	var att2 struct {
		Id string `json:"id"`
	}
	mustJSON(rec.Body.Bytes(), &att2)
	if att2.Id != att.Id {
		t.Errorf("replay returned a different record: %s vs %s", att2.Id, att.Id)
	}

	// check-out
	rec = do(srv, http.MethodPost, "/attendance/check-out", staffToken, `{"client_id":"`+cid+`"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("check-out status=%d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		Status     string  `json:"status"`
		CheckOutAt *string `json:"check_out_at"`
	}
	mustJSON(rec.Body.Bytes(), &out)
	if out.Status != "checked_out" || out.CheckOutAt == nil {
		t.Errorf("after check-out: status=%q check_out_at=%v", out.Status, out.CheckOutAt)
	}
}

// TestM5LiveView exercises the manager live view over HTTP against real
// Postgres: signup -> employee + location + published shift today -> provision
// staff login -> staff checks in against the shift -> GET /live shows one entry
// with status checked_in and the employee name.
func TestM5LiveView(t *testing.T) {
	cfg := &config.Config{
		DBHost:             envOr("DB_HOST", "localhost"),
		DBPort:             envOr("DB_PORT", "5432"),
		DBUser:             envOr("DB_USER", "opero"),
		DBPassword:         envOr("DB_PASSWORD", "opero"),
		DBSSLMode:          envOr("DB_SSLMODE", "disable"),
		ControlPlaneDBName: envOr("CONTROLPLANE_DB_NAME", "opero_control"),
		TenantDBPrefix:     "opero_m5_",
	}
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	pool, err := pgxpool.New(ctx, cfg.ControlPlaneDSN())
	if err != nil {
		t.Skipf("no control-plane pool: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("control-plane unreachable: %v", err)
	}
	ensureControlPlaneMigrated(t, cfg.ControlPlaneDSN())

	tokens := auth.NewTokenManager("m5-secret", "opero", time.Hour)
	store := controlplane.NewStore(pool)
	prov := controlplane.NewProvisioner(pool, cfg.DSN, logger)
	cpSvc := controlplane.NewService(store, prov, tokens, cfg.TenantDBPrefix, logger)
	idSvc := identity.NewService(logger, cpSvc)
	rsSvc := roster.NewService(logger, idSvc)
	atSvc := attendance.NewService(logger, idSvc, rsSvc)
	lvSvc := liveview.NewService(rsSvc, atSvc, idSvc)
	resolver := platdb.NewTenantResolver(cfg.DSN)
	api := &apiHandler{
		cp: controlplane.NewHandler(cpSvc, logger),
		id: identity.NewHandler(idSvc, logger),
		rs: roster.NewHandler(rsSvc, logger),
		at: attendance.NewHandler(atSvc, logger),
		lv: liveview.NewHandler(lvSvc, logger),
	}
	srv := httpserver.New(httpserver.Deps{
		Logger:              logger,
		ControlPlane:        pool,
		API:                 api,
		Tokens:              tokens,
		TenantRegistry:      store,
		TenantPools:         resolver,
		TenantRoutePrefixes: []string{"/employees", "/locations", "/shifts", "/attendance", "/live", "/me/shifts"},
	})

	slug := "m5-" + uuid.NewString()[:8]
	dbName := cfg.TenantDBPrefix + strings.ReplaceAll(slug, "-", "_")
	t.Cleanup(func() {
		resolver.Close()
		_, _ = pool.Exec(ctx, "DELETE FROM tenants WHERE slug=$1", slug)
		_ = prov.Drop(ctx, dbName)
		pool.Close()
	})

	mustJSON := func(b []byte, v any) {
		t.Helper()
		if err := json.Unmarshal(b, v); err != nil {
			t.Fatalf("json unmarshal: %v body=%s", err, b)
		}
	}

	rec := do(srv, http.MethodPost, "/auth/signup", "",
		`{"company_name":"M5 Co","slug":"`+slug+`","admin_email":"m5@x.test","admin_password":"password1"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("signup status=%d body=%s", rec.Code, rec.Body.String())
	}
	var signup struct {
		Token  string `json:"token"`
		Tenant struct {
			Id string `json:"id"`
		} `json:"tenant"`
	}
	mustJSON(rec.Body.Bytes(), &signup)
	admin := signup.Token
	tenantID, _ := uuid.Parse(signup.Tenant.Id)

	rec = do(srv, http.MethodPost, "/employees", admin, `{"full_name":"Ada Guide","employment_type":"freelance"}`)
	var emp struct {
		Id string `json:"id"`
	}
	mustJSON(rec.Body.Bytes(), &emp)

	// shift starting "now" (within the default UTC-day window unless run near midnight UTC;
	// to be deterministic we pass explicit from/to spanning this shift below).
	start := time.Now().UTC().Add(time.Minute).Truncate(time.Second)
	end := start.Add(4 * time.Hour)
	shiftBody := `{"employee_id":"` + emp.Id + `","starts_at":"` + start.Format(time.RFC3339) + `","ends_at":"` + end.Format(time.RFC3339) + `"}`
	rec = do(srv, http.MethodPost, "/shifts", admin, shiftBody)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create shift status=%d body=%s", rec.Code, rec.Body.String())
	}
	var shift struct {
		Id string `json:"id"`
	}
	mustJSON(rec.Body.Bytes(), &shift)
	if rec := do(srv, http.MethodPost, "/shifts/"+shift.Id+"/publish", admin, ""); rec.Code != http.StatusOK {
		t.Fatalf("publish status=%d", rec.Code)
	}

	// provision staff login + token, then check in against the shift
	rec = do(srv, http.MethodPost, "/employees/"+emp.Id+"/login", admin, `{"email":"ada@x.test","password":"password1"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("provision login status=%d body=%s", rec.Code, rec.Body.String())
	}
	var staff struct {
		Id   string `json:"id"`
		Role string `json:"role"`
	}
	mustJSON(rec.Body.Bytes(), &staff)
	staffUserID, _ := uuid.Parse(staff.Id)
	staffToken, _, err := tokens.Issue(staffUserID, tenantID, staff.Role, time.Now())
	if err != nil {
		t.Fatalf("issue staff token: %v", err)
	}
	// field staff lists their own shifts (resolves employee from the token)
	rec = do(srv, http.MethodGet, "/me/shifts", staffToken, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("my-shifts status=%d body=%s", rec.Code, rec.Body.String())
	}
	var myShifts []struct {
		Id string `json:"id"`
	}
	mustJSON(rec.Body.Bytes(), &myShifts)
	if len(myShifts) != 1 || myShifts[0].Id != shift.Id {
		t.Fatalf("my-shifts = %+v, want the one published shift %s", myShifts, shift.Id)
	}

	if rec := do(srv, http.MethodPost, "/attendance/check-in", staffToken,
		`{"client_id":"`+uuid.NewString()+`","shift_id":"`+shift.Id+`"}`); rec.Code != http.StatusCreated {
		t.Fatalf("check-in status=%d body=%s", rec.Code, rec.Body.String())
	}

	// Live view over a window whose lower bound is AFTER the check-in time
	// (check-in happened just now; from = shift start, which is ~1min from now,
	// so check_in_at < from). This guards the by-shift_id join: a check_in_at
	// window would wrongly drop this record and show not_checked_in.
	from := start.Format(time.RFC3339)
	to := start.Add(time.Hour).Format(time.RFC3339)
	rec = do(srv, http.MethodGet, "/live?from="+from+"&to="+to, admin, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("live view status=%d body=%s", rec.Code, rec.Body.String())
	}
	var entries []struct {
		EmployeeName     string `json:"employee_name"`
		AttendanceStatus string `json:"attendance_status"`
		Shift            struct {
			Id string `json:"id"`
		} `json:"shift"`
		CheckInAt *string `json:"check_in_at"`
	}
	mustJSON(rec.Body.Bytes(), &entries)
	if len(entries) != 1 {
		t.Fatalf("live entries = %d, want 1 (body=%s)", len(entries), rec.Body.String())
	}
	e := entries[0]
	if e.Shift.Id != shift.Id {
		t.Errorf("entry shift = %s, want %s", e.Shift.Id, shift.Id)
	}
	if e.EmployeeName != "Ada Guide" {
		t.Errorf("entry name = %q, want Ada Guide", e.EmployeeName)
	}
	if e.AttendanceStatus != "checked_in" {
		t.Errorf("entry status = %q, want checked_in", e.AttendanceStatus)
	}
	if e.CheckInAt == nil {
		t.Errorf("entry check_in_at should be populated")
	}
}

func do(h http.Handler, method, path, token, body string) *httptest.ResponseRecorder {
	var r *http.Request
	if body != "" {
		r = httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
	} else {
		r = httptest.NewRequest(method, path, nil)
	}
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	return rec
}

func ensureControlPlaneMigrated(t *testing.T, dsn string) {
	t.Helper()
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open control-plane: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()
	goose.SetBaseFS(dbassets.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("goose dialect: %v", err)
	}
	if err := goose.UpContext(context.Background(), sqlDB, dbassets.ControlPlaneDir); err != nil {
		t.Fatalf("goose up control-plane: %v", err)
	}
}
