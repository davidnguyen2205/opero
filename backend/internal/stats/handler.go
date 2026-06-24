package stats

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/davidnguyen2205/opero/backend/gen/oapi"
	appmw "github.com/davidnguyen2205/opero/backend/internal/platform/middleware"
)

// Handler is the thin HTTP layer for GET /me/stats.
type Handler struct {
	svc    *Service
	logger *slog.Logger
	now    func() time.Time
}

func NewHandler(svc *Service, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, logger: logger, now: time.Now}
}

func (h *Handler) GetMyStats(w http.ResponseWriter, r *http.Request) {
	claims, ok := appmw.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}
	userID, err := claims.UserID()
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "invalid token subject")
		return
	}
	st, err := h.svc.MyStats(r.Context(), userID, h.now())
	if err != nil {
		h.logger.ErrorContext(r.Context(), "stats request failed", slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}
	out := oapi.MyStats{
		ShiftsThisMonth: st.ShiftsThisMonth,
		HoursThisWeek:   st.HoursThisWeek,
		OnTimePct:       st.OnTimePct,
		TenureDays:      st.TenureDays,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(out)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(oapi.Error{Code: code, Message: message})
}
