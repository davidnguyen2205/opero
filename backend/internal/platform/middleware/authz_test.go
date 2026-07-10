package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidnguyen2205/opero/backend/internal/platform/auth"
)

func TestRoleSatisfies(t *testing.T) {
	cases := []struct {
		name     string
		actual   string
		required []string
		want     bool
	}{
		{"empty required allows employee", "employee", nil, true},
		{"empty required allows unknown", "whatever", []string{}, true},
		{"admin meets admin", "admin", []string{"admin"}, true},
		{"manager below admin", "manager", []string{"admin"}, false},
		{"employee below admin", "employee", []string{"admin"}, false},
		{"admin meets manager", "admin", []string{"manager"}, true},
		{"manager meets manager", "manager", []string{"manager"}, true},
		{"employee below manager", "employee", []string{"manager"}, false},
		{"employee meets employee", "employee", []string{"employee"}, true},
		{"manager meets employee", "manager", []string{"employee"}, true},
		{"admin meets employee", "admin", []string{"employee"}, true},
		{"unknown role denied", "superuser", []string{"employee"}, false},
		{"empty actual denied", "", []string{"employee"}, false},
		{"multiple takes most restrictive - manager fails", "manager", []string{"employee", "admin"}, false},
		{"multiple takes most restrictive - admin passes", "admin", []string{"employee", "admin"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := roleSatisfies(tc.actual, tc.required); got != tc.want {
				t.Fatalf("roleSatisfies(%q, %v) = %v, want %v", tc.actual, tc.required, got, tc.want)
			}
		})
	}
}

// runAuthz applies Authorizer(required) to a probe handler, optionally injecting
// claims into the request context, and reports the status code and whether the
// protected handler was reached.
func runAuthz(required []string, claims *auth.Claims) (int, bool) {
	reached := false
	h := Authorizer(quietLogger(), required)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if claims != nil {
		req = req.WithContext(context.WithValue(req.Context(), claimsContextKey, claims))
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Code, reached
}

func TestAuthorizer(t *testing.T) {
	t.Run("no claims -> 401 and handler not reached", func(t *testing.T) {
		code, reached := runAuthz([]string{"manager"}, nil)
		if code != http.StatusUnauthorized || reached {
			t.Fatalf("code=%d reached=%v, want 401 and handler not reached", code, reached)
		}
	})

	t.Run("sufficient role -> 200 and handler reached", func(t *testing.T) {
		code, reached := runAuthz([]string{"manager"}, &auth.Claims{Kind: "tenant", Role: "admin"})
		if code != http.StatusOK || !reached {
			t.Fatalf("code=%d reached=%v, want 200 and handler reached", code, reached)
		}
	})

	t.Run("exact role -> 200 and handler reached", func(t *testing.T) {
		code, reached := runAuthz([]string{"manager"}, &auth.Claims{Kind: "tenant", Role: "manager"})
		if code != http.StatusOK || !reached {
			t.Fatalf("code=%d reached=%v, want 200 and handler reached", code, reached)
		}
	})

	t.Run("insufficient role -> 403 and handler not reached", func(t *testing.T) {
		code, reached := runAuthz([]string{"manager"}, &auth.Claims{Kind: "tenant", Role: "employee"})
		if code != http.StatusForbidden || reached {
			t.Fatalf("code=%d reached=%v, want 403 and handler not reached", code, reached)
		}
	})

	t.Run("admin-only denies manager -> 403", func(t *testing.T) {
		code, reached := runAuthz([]string{"admin"}, &auth.Claims{Kind: "tenant", Role: "manager"})
		if code != http.StatusForbidden || reached {
			t.Fatalf("code=%d reached=%v, want 403 and handler not reached", code, reached)
		}
	})

	t.Run("empty scopes allow any authenticated caller", func(t *testing.T) {
		code, reached := runAuthz([]string{}, &auth.Claims{Kind: "tenant", Role: "employee"})
		if code != http.StatusOK || !reached {
			t.Fatalf("code=%d reached=%v, want 200 and handler reached", code, reached)
		}
	})
}
