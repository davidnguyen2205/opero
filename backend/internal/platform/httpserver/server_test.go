package httpserver

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/davidnguyen2205/opero/backend/gen/oapi"
	"github.com/davidnguyen2205/opero/backend/internal/platform/auth"
)

var testTM = auth.NewTokenManager("test-secret", "opero", time.Hour)

type fakePinger struct{ err error }

func (f fakePinger) Ping(context.Context) error { return f.err }

func newTestHandler(p Pinger) http.Handler {
	return New(Deps{
		Logger:       slog.New(slog.NewTextHandler(io.Discard, nil)),
		ControlPlane: p,
		API:          oapi.Unimplemented{},
		Tokens:       testTM,
		// TenantRegistry/TenantPools are unused by these tests: no tenant route
		// is exercised (TenantRoutePrefixes is empty, so tenant resolution is
		// never applied — the role gate runs and the request reaches the
		// Unimplemented handler).
	})
}

func do(h http.Handler, method, path string) *httptest.ResponseRecorder {
	return doAuth(h, method, path, "")
}

func doAuth(h http.Handler, method, path, authHeader string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	h.ServeHTTP(rec, req)
	return rec
}

func tenantBearer(t *testing.T, role string) string {
	t.Helper()
	tok, _, err := testTM.Issue(uuid.New(), uuid.New(), role, time.Now())
	if err != nil {
		t.Fatalf("Issue(%q): %v", role, err)
	}
	return "Bearer " + tok
}

func TestHealthOK(t *testing.T) {
	if rec := do(newTestHandler(fakePinger{}), http.MethodGet, "/health"); rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestHealthDegraded(t *testing.T) {
	if rec := do(newTestHandler(fakePinger{err: errors.New("db down")}), http.MethodGet, "/health"); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
}

// Secured route without a token must be rejected before reaching the handler.
func TestSecuredRouteRequiresAuth(t *testing.T) {
	if rec := do(newTestHandler(fakePinger{}), http.MethodGet, "/auth/me"); rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

// Public route passes through the auth gate (reaches the Unimplemented handler).
func TestPublicRouteSkipsAuth(t *testing.T) {
	if rec := do(newTestHandler(fakePinger{}), http.MethodPost, "/auth/login"); rec.Code != http.StatusNotImplemented {
		t.Fatalf("status = %d, want 501 (Unimplemented), got auth gate interference?", rec.Code)
	}
}

// TestRoleGateEnforced runs authenticated requests through securedChain for a
// manager-scoped route (GET /departments) and an admin-scoped route
// (POST /departments). An authorized role reaches the Unimplemented handler
// (501); an under-privileged role is stopped at the authorizer (403).
func TestRoleGateEnforced(t *testing.T) {
	h := newTestHandler(fakePinger{})
	cases := []struct {
		name       string
		method     string
		path       string
		role       string
		wantStatus int
	}{
		{"manager route, admin allowed", http.MethodGet, "/departments", "admin", http.StatusNotImplemented},
		{"manager route, manager allowed", http.MethodGet, "/departments", "manager", http.StatusNotImplemented},
		{"manager route, employee denied", http.MethodGet, "/departments", "employee", http.StatusForbidden},
		{"admin route, admin allowed", http.MethodPost, "/departments", "admin", http.StatusNotImplemented},
		{"admin route, manager denied", http.MethodPost, "/departments", "manager", http.StatusForbidden},
		{"admin route, employee denied", http.MethodPost, "/departments", "employee", http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := doAuth(h, tc.method, tc.path, tenantBearer(t, tc.role))
			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
		})
	}
}
