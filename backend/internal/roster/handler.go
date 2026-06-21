package roster

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/davidnguyen2205/opero/backend/gen/oapi"
	appmw "github.com/davidnguyen2205/opero/backend/internal/platform/middleware"
)

// Handler is the thin HTTP layer implementing the roster slice of the
// oapi-generated ServerInterface.
type Handler struct {
	svc    *Service
	logger *slog.Logger
}

func NewHandler(svc *Service, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

// --- locations ---

func (h *Handler) ListLocations(w http.ResponseWriter, r *http.Request) {
	locs, err := h.svc.ListLocations(r.Context())
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	out := make([]oapi.Location, 0, len(locs))
	for _, l := range locs {
		out = append(out, toLocation(l))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) CreateLocation(w http.ResponseWriter, r *http.Request) {
	var body oapi.CreateLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "malformed JSON body")
		return
	}
	l, err := h.svc.CreateLocation(r.Context(), CreateLocationInput{
		Name:    body.Name,
		Address: body.Address,
		Lat:     body.Lat,
		Lng:     body.Lng,
	})
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toLocation(l))
}

func (h *Handler) GetLocation(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	l, err := h.svc.GetLocation(r.Context(), id)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toLocation(l))
}

func (h *Handler) UpdateLocation(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	var body oapi.UpdateLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "malformed JSON body")
		return
	}
	l, err := h.svc.UpdateLocation(r.Context(), id, UpdateLocationInput{
		Name:    body.Name,
		Address: body.Address,
		Lat:     body.Lat,
		Lng:     body.Lng,
	})
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toLocation(l))
}

func (h *Handler) DeleteLocation(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	if err := h.svc.DeleteLocation(r.Context(), id); err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- shifts ---

func (h *Handler) ListShifts(w http.ResponseWriter, r *http.Request, params oapi.ListShiftsParams) {
	shifts, err := h.svc.ListShifts(r.Context(), ShiftFilter{
		EmployeeID: params.EmployeeId,
		Status:     enumToStrPtr(params.Status),
		From:       params.From,
		To:         params.To,
	})
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	out := make([]oapi.Shift, 0, len(shifts))
	for _, s := range shifts {
		out = append(out, toShift(s))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) ListMyShifts(w http.ResponseWriter, r *http.Request, params oapi.ListMyShiftsParams) {
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
	shifts, err := h.svc.ListMyShifts(r.Context(), userID, ShiftFilter{
		Status: enumToStrPtr(params.Status),
		From:   params.From,
		To:     params.To,
	})
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	out := make([]oapi.Shift, 0, len(shifts))
	for _, s := range shifts {
		out = append(out, toShift(s))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) CreateShift(w http.ResponseWriter, r *http.Request) {
	var body oapi.CreateShiftRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "malformed JSON body")
		return
	}
	s, err := h.svc.CreateShift(r.Context(), CreateShiftInput{
		EmployeeID: body.EmployeeId,
		LocationID: body.LocationId,
		StartsAt:   body.StartsAt,
		EndsAt:     body.EndsAt,
		Notes:      body.Notes,
	})
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toShift(s))
}

func (h *Handler) GetShift(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	s, err := h.svc.GetShift(r.Context(), id)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toShift(s))
}

func (h *Handler) UpdateShift(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	var body oapi.UpdateShiftRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "malformed JSON body")
		return
	}
	s, err := h.svc.UpdateShift(r.Context(), id, UpdateShiftInput{
		EmployeeID: body.EmployeeId,
		LocationID: body.LocationId,
		StartsAt:   body.StartsAt,
		EndsAt:     body.EndsAt,
		Notes:      body.Notes,
	})
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toShift(s))
}

func (h *Handler) DeleteShift(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	if err := h.svc.DeleteShift(r.Context(), id); err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) PublishShift(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	s, err := h.svc.PublishShift(r.Context(), id)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toShift(s))
}

// --- mapping domain -> generated API types ---

func toLocation(l Location) oapi.Location {
	return oapi.Location{
		Id:        l.ID,
		Name:      l.Name,
		Address:   l.Address,
		Lat:       l.Lat,
		Lng:       l.Lng,
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
	}
}

func toShift(s Shift) oapi.Shift {
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

// enumToStrPtr converts any *NamedStringType to *string (nil-safe).
func enumToStrPtr[T ~string](v *T) *string {
	if v == nil {
		return nil
	}
	s := string(*v)
	return &s
}

// --- response helpers ---

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, oapi.Error{Code: code, Message: message})
}

func (h *Handler) writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrValidation):
		writeError(w, http.StatusBadRequest, "validation_failed", err.Error())
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "resource not found")
	default:
		h.logger.ErrorContext(r.Context(), "roster request failed", slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
