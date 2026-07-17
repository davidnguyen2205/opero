package demoseed

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/davidnguyen2205/opero/backend/gen/oapi"
	"github.com/davidnguyen2205/opero/backend/internal/controlplane"
	"github.com/davidnguyen2205/opero/backend/internal/platform/middleware"
)

// tenantGetter resolves a tenant id to its control-plane record (for the slug
// gate). Satisfied by controlplane.Service.
type tenantGetter interface {
	PlatformGetTenant(ctx context.Context, id uuid.UUID) (controlplane.Tenant, error)
}

type Handler struct {
	svc     *Service
	tenants tenantGetter
	// demoSlug is the only tenant slug allowed to seed (DEMO_TENANT_SLUG).
	// Empty disables the endpoint entirely.
	demoSlug string
	logger   *slog.Logger
}

func NewHandler(svc *Service, tenants tenantGetter, demoSlug string, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, tenants: tenants, demoSlug: demoSlug, logger: logger}
}

func (h *Handler) SeedLiveViewDemoData(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}
	if h.demoSlug == "" {
		writeError(w, http.StatusForbidden, "forbidden", "demo seeding is not enabled on this server")
		return
	}
	tenant, err := h.tenants.PlatformGetTenant(r.Context(), claims.TenantIDValue())
	if err != nil {
		h.logger.ErrorContext(r.Context(), "demo seed tenant lookup failed", slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}
	if tenant.Slug != h.demoSlug {
		writeError(w, http.StatusForbidden, "forbidden", "demo seeding is only available to the demo tenant")
		return
	}

	res, err := h.svc.Seed(r.Context())
	if err != nil {
		h.logger.ErrorContext(r.Context(), "demo seed failed", slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}
	h.logger.InfoContext(r.Context(), "demo live view re-seeded",
		slog.Int("shifts", res.Shifts), slog.Int("attendance_records", res.AttendanceRecords))
	writeJSON(w, http.StatusOK, oapi.SeedLiveViewResponse{
		Shifts:            res.Shifts,
		AttendanceRecords: res.AttendanceRecords,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, oapi.Error{Code: code, Message: message})
}
