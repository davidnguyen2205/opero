package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/davidnguyen2205/opero/backend/internal/platform/auth"
)

func testTokens(t *testing.T) *auth.TokenManager {
	t.Helper()
	return auth.NewTokenManager("test-secret", "opero", time.Hour)
}

func quietLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// run applies mw to a handler that records whether it was reached, sends a
// request with the given Authorization header, and returns the status code plus
// whether the protected handler ran.
func run(mw func(http.Handler) http.Handler, authHeader string) (int, bool) {
	reached := false
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Code, reached
}

// TestKindBoundary is the crux of the tenant/platform separation: a token of
// one kind must never be accepted on the other surface.
func TestKindBoundary(t *testing.T) {
	tm := testTokens(t)
	logger := quietLogger()

	tenantTok, _, err := tm.Issue(uuid.New(), uuid.New(), "admin", time.Now())
	if err != nil {
		t.Fatalf("Issue tenant: %v", err)
	}
	platformTok, _, err := tm.IssuePlatform(uuid.New(), "super_admin", time.Now())
	if err != nil {
		t.Fatalf("IssuePlatform: %v", err)
	}

	tenantMW := TenantAuthenticator(tm, logger)
	platformMW := PlatformAuthenticator(tm, logger)

	t.Run("platform token rejected on tenant route", func(t *testing.T) {
		code, reached := run(tenantMW, "Bearer "+platformTok)
		if code != http.StatusUnauthorized || reached {
			t.Fatalf("code=%d reached=%v, want 401 and handler not reached", code, reached)
		}
	})

	t.Run("tenant token rejected on platform route", func(t *testing.T) {
		code, reached := run(platformMW, "Bearer "+tenantTok)
		if code != http.StatusUnauthorized || reached {
			t.Fatalf("code=%d reached=%v, want 401 and handler not reached", code, reached)
		}
	})

	t.Run("tenant token accepted on tenant route", func(t *testing.T) {
		code, reached := run(tenantMW, "Bearer "+tenantTok)
		if code != http.StatusOK || !reached {
			t.Fatalf("code=%d reached=%v, want 200 and handler reached", code, reached)
		}
	})

	t.Run("platform token accepted on platform route", func(t *testing.T) {
		code, reached := run(platformMW, "Bearer "+platformTok)
		if code != http.StatusOK || !reached {
			t.Fatalf("code=%d reached=%v, want 200 and handler reached", code, reached)
		}
	})
}

func TestAuthenticatorRejectsMissingAndBadTokens(t *testing.T) {
	tm := testTokens(t)
	mw := PlatformAuthenticator(tm, quietLogger())

	cases := map[string]string{
		"no header":     "",
		"not bearer":    "Basic abc",
		"garbage token": "Bearer not-a-jwt",
		"empty bearer":  "Bearer ",
	}
	for name, header := range cases {
		t.Run(name, func(t *testing.T) {
			code, reached := run(mw, header)
			if code != http.StatusUnauthorized || reached {
				t.Fatalf("code=%d reached=%v, want 401 and handler not reached", code, reached)
			}
		})
	}
}

// TestPlainAuthenticatorAllowsEitherKind documents that the kind-agnostic
// Authenticator (used for control-plane routes like /auth/me) accepts a valid
// token of any kind — the kind gating lives in the Tenant/Platform variants.
func TestPlainAuthenticatorAllowsEitherKind(t *testing.T) {
	tm := testTokens(t)
	mw := Authenticator(tm, quietLogger())
	tenantTok, _, _ := tm.Issue(uuid.New(), uuid.New(), "admin", time.Now())
	code, reached := run(mw, "Bearer "+tenantTok)
	if code != http.StatusOK || !reached {
		t.Fatalf("code=%d reached=%v, want 200 and handler reached", code, reached)
	}
}
