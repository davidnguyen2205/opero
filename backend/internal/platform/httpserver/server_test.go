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

	"github.com/davidnguyen2205/opero/backend/gen/oapi"
	"github.com/davidnguyen2205/opero/backend/internal/platform/auth"
)

type fakePinger struct{ err error }

func (f fakePinger) Ping(context.Context) error { return f.err }

func newTestHandler(p Pinger) http.Handler {
	tm := auth.NewTokenManager("test-secret", "opero", time.Hour)
	return New(Deps{
		Logger:       slog.New(slog.NewTextHandler(io.Discard, nil)),
		ControlPlane: p,
		API:          oapi.Unimplemented{},
		Tokens:       tm,
		// TenantRegistry/TenantPools are unused by these tests (no tenant route hit).
	})
}

func do(h http.Handler, method, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	return rec
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
