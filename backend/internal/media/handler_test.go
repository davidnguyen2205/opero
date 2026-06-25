package media

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"time"

	"github.com/google/uuid"

	"github.com/davidnguyen2205/opero/backend/gen/oapi"
	"github.com/davidnguyen2205/opero/backend/internal/platform/auth"
	appmw "github.com/davidnguyen2205/opero/backend/internal/platform/middleware"
)

// authed wraps the handler in the real tenant authenticator and signs a token
// for tenantID, exercising the actual auth path. A nil-UUID tenantID issues a
// token with no tenant scope.
func authed(t *testing.T, h *Handler, tenantID uuid.UUID) (http.Handler, string) {
	t.Helper()
	tm := auth.NewTokenManager("test-secret", "opero", time.Hour)
	tok, _, err := tm.Issue(uuid.New(), tenantID, "employee", time.Now())
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	mw := appmw.TenantAuthenticator(tm, slog.New(slog.NewTextHandler(io.Discard, nil)))
	return mw(http.HandlerFunc(h.UploadMedia)), tok
}

// fakeStore records the last upload and returns a canned URL.
type fakeStore struct {
	gotKey         string
	gotContentType string
	gotBytes       []byte
	uploadErr      error
	url            string
	urlErr         error
}

func (f *fakeStore) Upload(_ context.Context, key, contentType string, _ int64, r io.Reader) error {
	if f.uploadErr != nil {
		return f.uploadErr
	}
	f.gotKey = key
	f.gotContentType = contentType
	b, _ := io.ReadAll(r)
	f.gotBytes = b
	return nil
}

func (f *fakeStore) URL(_ context.Context, _ string) (string, error) {
	return f.url, f.urlErr
}

func newMultipart(t *testing.T, field, filename, content string) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile(field, filename)
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := io.WriteString(fw, content); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	return &buf, mw.FormDataContentType()
}

func TestUploadMedia_Success(t *testing.T) {
	store := &fakeStore{url: "https://example.com/opero-media/key"}
	h := NewHandler(store, slog.New(slog.NewTextHandler(io.Discard, nil)))

	body, ct := newMultipart(t, "file", "photo.JPG", "hello")
	tenantID := uuid.New()
	handler, tok := authed(t, h, tenantID)
	req := httptest.NewRequest(http.MethodPost, "/media", body)
	req.Header.Set("Content-Type", ct)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	var resp oapi.MediaUploadResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Url != store.url {
		t.Errorf("url = %q, want %q", resp.Url, store.url)
	}
	if string(store.gotBytes) != "hello" {
		t.Errorf("uploaded bytes = %q, want %q", store.gotBytes, "hello")
	}
	if !strings.HasPrefix(store.gotKey, tenantID.String()+"/") {
		t.Errorf("key %q not namespaced by tenant %q", store.gotKey, tenantID)
	}
	if !strings.HasSuffix(store.gotKey, ".jpg") {
		t.Errorf("key %q should keep lowercased .jpg ext", store.gotKey)
	}
}

func TestUploadMedia_Unauthenticated(t *testing.T) {
	store := &fakeStore{}
	h := NewHandler(store, slog.New(slog.NewTextHandler(io.Discard, nil)))

	body, ct := newMultipart(t, "file", "p.png", "x")
	// No auth wrapper, no claims in context.
	req := httptest.NewRequest(http.MethodPost, "/media", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()

	h.UploadMedia(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestUploadMedia_NilTenant(t *testing.T) {
	store := &fakeStore{}
	h := NewHandler(store, slog.New(slog.NewTextHandler(io.Discard, nil)))

	body, ct := newMultipart(t, "file", "p.png", "x")
	handler, tok := authed(t, h, uuid.Nil)
	req := httptest.NewRequest(http.MethodPost, "/media", body)
	req.Header.Set("Content-Type", ct)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestUploadMedia_MissingFileField(t *testing.T) {
	store := &fakeStore{}
	h := NewHandler(store, slog.New(slog.NewTextHandler(io.Discard, nil)))

	body, ct := newMultipart(t, "wrongfield", "p.png", "x")
	handler, tok := authed(t, h, uuid.New())
	req := httptest.NewRequest(http.MethodPost, "/media", body)
	req.Header.Set("Content-Type", ct)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestObjectKey(t *testing.T) {
	tid := uuid.New()
	tests := []struct {
		name     string
		filename string
		wantExt  string
	}{
		{"jpg", "photo.jpg", ".jpg"},
		{"uppercase", "PHOTO.PNG", ".png"},
		{"no ext", "photo", ""},
		{"unsafe ext stripped", "evil.tar.gz123456", ""},
		{"path traversal name", "../../etc/passwd", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := objectKey(tid, tt.filename)
			if !strings.HasPrefix(key, tid.String()+"/") {
				t.Errorf("key %q not namespaced by tenant", key)
			}
			if strings.Contains(key, "..") || strings.Contains(strings.TrimPrefix(key, tid.String()+"/"), "/") {
				t.Errorf("key %q contains a traversal or nested path", key)
			}
			if tt.wantExt == "" {
				if strings.Contains(strings.TrimPrefix(key, tid.String()+"/"), ".") {
					t.Errorf("key %q should have no extension", key)
				}
			} else if !strings.HasSuffix(key, tt.wantExt) {
				t.Errorf("key %q should end with %q", key, tt.wantExt)
			}
		})
	}
}
