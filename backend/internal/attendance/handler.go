package attendance

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/davidnguyen2205/opero/backend/gen/oapi"
	appmw "github.com/davidnguyen2205/opero/backend/internal/platform/middleware"
)

// Handler is the thin HTTP layer implementing the attendance slice of the
// oapi-generated ServerInterface.
type Handler struct {
	svc    *Service
	logger *slog.Logger
}

func NewHandler(svc *Service, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

func (h *Handler) currentUserID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	claims, present := appmw.ClaimsFromContext(r.Context())
	if !present {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return uuid.Nil, false
	}
	uid, err := claims.UserID()
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "invalid token subject")
		return uuid.Nil, false
	}
	return uid, true
}

func (h *Handler) CheckIn(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.currentUserID(w, r)
	if !ok {
		return
	}
	var body oapi.CheckInRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "malformed JSON body")
		return
	}
	rec, created, err := h.svc.CheckIn(r.Context(), userID, CheckInInput{
		ClientID: body.ClientId,
		ShiftID:  body.ShiftId,
		Lat:      body.Lat,
		Lng:      body.Lng,
		PhotoURL: body.PhotoUrl,
	})
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	writeJSON(w, status, toRecord(rec))
}

func (h *Handler) CheckOut(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.currentUserID(w, r)
	if !ok {
		return
	}
	var body oapi.CheckOutRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "malformed JSON body")
		return
	}
	rec, err := h.svc.CheckOut(r.Context(), userID, CheckOutInput{
		ClientID: body.ClientId,
		Lat:      body.Lat,
		Lng:      body.Lng,
		PhotoURL: body.PhotoUrl,
	})
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toRecord(rec))
}

func (h *Handler) SetBreak(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.currentUserID(w, r)
	if !ok {
		return
	}
	var body oapi.SetBreakRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "malformed JSON body")
		return
	}
	rec, err := h.svc.SetBreak(r.Context(), userID, body.ClientId, body.OnBreak)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toRecord(rec))
}

func (h *Handler) ListAttendance(w http.ResponseWriter, r *http.Request, params oapi.ListAttendanceParams) {
	recs, err := h.svc.List(r.Context(), AttendanceFilter{
		EmployeeID: params.EmployeeId,
		Status:     enumToStrPtr(params.Status),
		From:       params.From,
		To:         params.To,
	})
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	out := make([]oapi.AttendanceRecord, 0, len(recs))
	for _, rec := range recs {
		out = append(out, toRecord(rec))
	}
	writeJSON(w, http.StatusOK, out)
}

func toRecord(r Record) oapi.AttendanceRecord {
	return oapi.AttendanceRecord{
		Id:               r.ID,
		EmployeeId:       r.EmployeeID,
		ShiftId:          r.ShiftID,
		ClientId:         r.ClientID,
		CheckInAt:        r.CheckInAt,
		CheckInLat:       r.CheckInLat,
		CheckInLng:       r.CheckInLng,
		CheckInPhotoUrl:  r.CheckInPhotoURL,
		CheckOutAt:       r.CheckOutAt,
		CheckOutLat:      r.CheckOutLat,
		CheckOutLng:      r.CheckOutLng,
		CheckOutPhotoUrl: r.CheckOutPhotoURL,
		Status:           oapi.AttendanceRecordStatus(r.Status),
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
	}
}

func enumToStrPtr[T ~string](v *T) *string {
	if v == nil {
		return nil
	}
	s := string(*v)
	return &s
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
	case errors.Is(err, ErrConflict):
		writeError(w, http.StatusConflict, "conflict", "client_id already used by another employee")
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "attendance record not found")
	default:
		h.logger.ErrorContext(r.Context(), "attendance request failed", slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
