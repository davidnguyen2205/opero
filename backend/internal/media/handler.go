// Package media implements the uploadMedia operation: it accepts a multipart
// file upload from an authenticated caller and stores it in object storage
// under the caller's tenant, returning a URL to reference it.
//
// It is a thin handler over the platform storage interface. It deliberately
// uses object storage keyed by tenant_id (from the JWT) rather than a tenant
// DB pool, so /media is a secured route but NOT a tenant-data route.
package media

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"github.com/davidnguyen2205/opero/backend/gen/oapi"
	appmw "github.com/davidnguyen2205/opero/backend/internal/platform/middleware"
	"github.com/davidnguyen2205/opero/backend/internal/platform/storage"
)

// maxUploadBytes caps a single upload to keep memory and storage bounded.
const maxUploadBytes = 15 << 20 // 15 MiB

// formField is the multipart field name the spec defines for the file.
const formField = "file"

// Handler implements the uploadMedia slice of the oapi ServerInterface.
type Handler struct {
	store  storage.ObjectStore
	logger *slog.Logger
}

func NewHandler(store storage.ObjectStore, logger *slog.Logger) *Handler {
	return &Handler{store: store, logger: logger}
}

// UploadMedia handles POST /media.
func (h *Handler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	claims, present := appmw.ClaimsFromContext(r.Context())
	if !present {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}
	tenantID := claims.TenantIDValue()
	if tenantID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "token is not scoped to a tenant")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	file, header, err := r.FormFile(formField)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "expected a multipart form with a \"file\" field")
		return
	}
	defer func() { _ = file.Close() }()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	key := objectKey(tenantID, header.Filename)

	if err := h.store.Upload(r.Context(), key, contentType, header.Size, file); err != nil {
		h.logger.ErrorContext(r.Context(), "media upload failed", slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to store file")
		return
	}

	url, err := h.store.URL(r.Context(), key)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "media url failed", slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to produce file url")
		return
	}

	writeJSON(w, http.StatusCreated, oapi.MediaUploadResponse{Url: url})
}

// objectKey builds a tenant-namespaced, collision-resistant key, preserving the
// original file extension (sanitized) for content sniffing/serving.
func objectKey(tenantID uuid.UUID, filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if !safeExt(ext) {
		ext = ""
	}
	return fmt.Sprintf("%s/%s%s", tenantID.String(), uuid.NewString(), ext)
}

// safeExt allows only a short, alphanumeric extension to avoid path tricks.
func safeExt(ext string) bool {
	if ext == "" || ext[0] != '.' || len(ext) > 6 {
		return false
	}
	for _, c := range ext[1:] {
		isLower := c >= 'a' && c <= 'z'
		isDigit := c >= '0' && c <= '9'
		if !isLower && !isDigit {
			return false
		}
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, oapi.Error{Code: code, Message: message})
}
