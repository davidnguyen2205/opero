//go:build integration

// Integration test for the M1 request lifecycle against a real Postgres.
// Run with: go test -tags=integration ./internal/controlplane/...
// Requires a reachable Postgres (defaults match docker-compose: opero/opero).
package controlplane_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	dbassets "github.com/davidnguyen2205/opero/backend/db"
	"github.com/davidnguyen2205/opero/backend/gen/oapi"
	"github.com/davidnguyen2205/opero/backend/internal/controlplane"
	"github.com/davidnguyen2205/opero/backend/internal/platform/auth"
	"github.com/davidnguyen2205/opero/backend/internal/platform/config"
	platdb "github.com/davidnguyen2205/opero/backend/internal/platform/db"
	"github.com/davidnguyen2205/opero/backend/internal/platform/httpserver"
	appmw "github.com/davidnguyen2205/opero/backend/internal/platform/middleware"
)

// testAPI satisfies the full generated ServerInterface for this test: real
// control-plane handlers, with the identity routes left Unimplemented (this
// test only exercises /auth/* and a standalone tenant-resolution route).
type testAPI struct {
	oapi.Unimplemented
	cp *controlplane.Handler
}

func (a testAPI) Signup(w http.ResponseWriter, r *http.Request)         { a.cp.Signup(w, r) }
func (a testAPI) Login(w http.ResponseWriter, r *http.Request)          { a.cp.Login(w, r) }
func (a testAPI) GetCurrentUser(w http.ResponseWriter, r *http.Request) { a.cp.GetCurrentUser(w, r) }

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func TestSignupLoginMeLifecycle(t *testing.T) {
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

	pool, err := pgxpool.New(ctx, cfg.ControlPlaneDSN())
	if err != nil {
		t.Skipf("no control-plane pool: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("control-plane DB unreachable: %v", err)
	}

	ensureControlPlaneMigrated(t, cfg.ControlPlaneDSN())

	tokens := auth.NewTokenManager("itest-secret", "opero", time.Hour)
	store := controlplane.NewStore(pool)
	prov := controlplane.NewProvisioner(pool, cfg.DSN, logger)
	svc := controlplane.NewService(store, prov, tokens, "opero_itest_", logger)
	handler := controlplane.NewHandler(svc, logger)
	resolver := platdb.NewTenantResolver(cfg.DSN)
	api := httpserver.New(httpserver.Deps{
		Logger:         logger,
		ControlPlane:   pool,
		API:            testAPI{cp: handler},
		Tokens:         tokens,
		TenantRegistry: store,
		TenantPools:    resolver,
	})

	// Unique slug per run so repeated runs don't collide.
	slug := fmt.Sprintf("itest-%d", time.Now().UnixNano())
	dbName := "opero_itest_" + strings.ReplaceAll(slug, "-", "_")

	t.Cleanup(func() {
		resolver.Close() // close tenant pools before dropping the DB
		_, _ = pool.Exec(ctx, "DELETE FROM tenants WHERE slug=$1", slug)
		_ = prov.Drop(ctx, dbName)
		pool.Close()
	})

	// --- signup ---
	signupBody := fmt.Sprintf(`{"company_name":"ITest Co","slug":%q,"admin_email":"it@itest.test","admin_password":"password1"}`, slug)
	rec := doReq(api, http.MethodPost, "/auth/signup", "", signupBody)
	if rec.Code != http.StatusCreated {
		t.Fatalf("signup status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var auth1 struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &auth1); err != nil || auth1.Token == "" {
		t.Fatalf("signup token parse: %v body=%s", err, rec.Body.String())
	}

	// --- login ---
	loginBody := fmt.Sprintf(`{"tenant_slug":%q,"email":"it@itest.test","password":"password1"}`, slug)
	if rec := doReq(api, http.MethodPost, "/auth/login", "", loginBody); rec.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", rec.Code, rec.Body.String())
	}

	// --- /auth/me without token ---
	if rec := doReq(api, http.MethodGet, "/auth/me", "", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("me-without-token status = %d, want 401", rec.Code)
	}

	// --- /auth/me with token ---
	rec = doReq(api, http.MethodGet, "/auth/me", auth1.Token, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("me status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "it@itest.test") {
		t.Errorf("me body missing email: %s", rec.Body.String())
	}

	// --- tenant resolution path (Authenticator -> TenantResolver -> tenant pool) ---
	tr := chi.NewRouter()
	tr.With(
		appmw.Authenticator(tokens, logger),
		appmw.TenantResolver(store, resolver, logger),
	).Get("/tenant-check", func(w http.ResponseWriter, r *http.Request) {
		p, ok := appmw.TenantPoolFromContext(r.Context())
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := p.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	if rec := doReq(tr, http.MethodGet, "/tenant-check", auth1.Token, ""); rec.Code != http.StatusOK {
		t.Fatalf("tenant-check status = %d (tenant resolution/ping failed)", rec.Code)
	}
}

func doReq(h http.Handler, method, path, token, body string) *httptest.ResponseRecorder {
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
