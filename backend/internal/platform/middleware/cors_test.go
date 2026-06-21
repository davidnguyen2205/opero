package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func corsHandler(origins []string) http.Handler {
	return CORS(origins)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
}

func TestCORSAllowedOriginSetsHeaders(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	corsHandler([]string{"http://localhost:5173"}).ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Errorf("ACAO = %q, want the origin", got)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestCORSPreflightAllowed(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/x", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	corsHandler([]string{"http://localhost:5173"}).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("preflight status = %d, want 204", rec.Code)
	}
}

func TestCORSDisallowedOrigin(t *testing.T) {
	// Non-preflight from a disallowed origin: pass through, no ACAO header.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Origin", "http://evil.example")
	corsHandler([]string{"http://localhost:5173"}).ServeHTTP(rec, req)
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("ACAO should not be set for a disallowed origin")
	}

	// Preflight from a disallowed origin: rejected.
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodOptions, "/x", nil)
	req2.Header.Set("Origin", "http://evil.example")
	corsHandler([]string{"http://localhost:5173"}).ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusForbidden {
		t.Errorf("disallowed preflight status = %d, want 403", rec2.Code)
	}
}

func TestCORSDisabledNoOp(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	corsHandler(nil).ServeHTTP(rec, req)
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("ACAO should not be set when CORS is disabled")
	}
}
