package tours

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/davidnguyen2205/opero/backend/gen/oapi"
)

// Handler is the thin HTTP layer implementing the tours slice of the
// oapi-generated ServerInterface.
type Handler struct {
	svc    *Service
	logger *slog.Logger
}

func NewHandler(svc *Service, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

func (h *Handler) ListTours(w http.ResponseWriter, r *http.Request, params oapi.ListToursParams) {
	tours, err := h.svc.List(r.Context(), Filter{
		Category: categoryToStrPtr(params.Category),
		Active:   params.Active,
	})
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	out := make([]oapi.Tour, 0, len(tours))
	for _, t := range tours {
		out = append(out, toTour(t))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) CreateTour(w http.ResponseWriter, r *http.Request) {
	var body oapi.CreateTourRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "malformed JSON body")
		return
	}
	in := CreateInput{
		Name:           body.Name,
		Category:       string(body.Category),
		MeetingPoint:   body.MeetingPoint,
		DurationMin:    derefOr(body.DurationMin, 120),
		MaxGuests:      derefOr(body.MaxGuests, 10),
		GuidesNeeded:   derefOr(body.GuidesNeeded, 1),
		DriversNeeded:  derefOr(body.DriversNeeded, 0),
		PriceCents:     derefOr(body.PriceCents, 0),
		Rating:         body.Rating,
		Active:         derefBool(body.Active, true),
		Color:          body.Color,
		Description:    body.Description,
		DepartureTimes: derefSlice(body.DepartureTimes),
	}
	t, err := h.svc.Create(r.Context(), in)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toTour(t))
}

func (h *Handler) GetTour(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	t, err := h.svc.Get(r.Context(), id)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toTour(t))
}

func (h *Handler) UpdateTour(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	var body oapi.UpdateTourRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "malformed JSON body")
		return
	}
	in := UpdateInput{
		Name:          body.Name,
		MeetingPoint:  body.MeetingPoint,
		DurationMin:   body.DurationMin,
		MaxGuests:     body.MaxGuests,
		GuidesNeeded:  body.GuidesNeeded,
		DriversNeeded: body.DriversNeeded,
		PriceCents:    body.PriceCents,
		Rating:        body.Rating,
		Active:        body.Active,
		Color:         body.Color,
		Description:   body.Description,
	}
	if body.Category != nil {
		c := string(*body.Category)
		in.Category = &c
	}
	if body.DepartureTimes != nil {
		in.DepartureTimes = *body.DepartureTimes
	}
	t, err := h.svc.Update(r.Context(), id, in)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toTour(t))
}

func (h *Handler) DeleteTour(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	if err := h.svc.Delete(r.Context(), id); err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- mapping + helpers ---

func toTour(t Tour) oapi.Tour {
	return oapi.Tour{
		Id:             t.ID,
		Name:           t.Name,
		Category:       oapi.TourCategory(t.Category),
		MeetingPoint:   t.MeetingPoint,
		DurationMin:    t.DurationMin,
		MaxGuests:      t.MaxGuests,
		GuidesNeeded:   t.GuidesNeeded,
		DriversNeeded:  t.DriversNeeded,
		DepartureTimes: t.DepartureTimes,
		PriceCents:     t.PriceCents,
		Rating:         t.Rating,
		Active:         t.Active,
		Color:          t.Color,
		Description:    t.Description,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
	}
}

func categoryToStrPtr(c *oapi.TourCategory) *string {
	if c == nil {
		return nil
	}
	v := string(*c)
	return &v
}

func derefOr(p *int, def int) int {
	if p == nil {
		return def
	}
	return *p
}

func derefBool(p *bool, def bool) bool {
	if p == nil {
		return def
	}
	return *p
}

func derefSlice(p *[]string) []string {
	if p == nil {
		return []string{}
	}
	return *p
}

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
		h.logger.ErrorContext(r.Context(), "tours request failed", slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
