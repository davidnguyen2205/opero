package liveview

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/davidnguyen2205/opero/backend/gen/oapi"
	"github.com/davidnguyen2205/opero/backend/internal/roster"
)

type Handler struct {
	svc    *Service
	logger *slog.Logger
}

func NewHandler(svc *Service, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

func (h *Handler) GetLiveView(w http.ResponseWriter, r *http.Request, params oapi.GetLiveViewParams) {
	from, to := resolveWindow(params, time.Now().UTC())

	entries, err := h.svc.LiveView(r.Context(), from, to)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "live view failed", slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}

	out := make([]oapi.LiveViewEntry, 0, len(entries))
	for _, e := range entries {
		out = append(out, oapi.LiveViewEntry{
			EmployeeId:       e.EmployeeID,
			EmployeeName:     e.EmployeeName,
			Shift:            toShift(e.Shift),
			AttendanceStatus: oapi.LiveViewEntryAttendanceStatus(e.AttendanceStatus),
			CheckInAt:        e.CheckInAt,
			CheckOutAt:       e.CheckOutAt,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

// resolveWindow defaults an omitted from/to to the current UTC day. The server
// has no tenant timezone; clients should pass local-day bounds for correctness.
func resolveWindow(params oapi.GetLiveViewParams, now time.Time) (time.Time, time.Time) {
	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if params.From != nil {
		from = *params.From
	}
	to := from.Add(24 * time.Hour)
	if params.To != nil {
		to = *params.To
	}
	return from, to
}

func toShift(s roster.Shift) oapi.Shift {
	return oapi.Shift{
		Id:         s.ID,
		EmployeeId: s.EmployeeID,
		LocationId: s.LocationID,
		StartsAt:   s.StartsAt,
		EndsAt:     s.EndsAt,
		Notes:      s.Notes,
		Status:     oapi.ShiftStatus(s.Status),
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, oapi.Error{Code: code, Message: message})
}
