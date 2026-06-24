package leave

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/davidnguyen2205/opero/backend/gen/oapi"
	appmw "github.com/davidnguyen2205/opero/backend/internal/platform/middleware"
)

// Handler is the thin HTTP layer implementing the leave slice of the
// oapi-generated ServerInterface.
type Handler struct {
	svc    *Service
	logger *slog.Logger
	now    func() time.Time
}

func NewHandler(svc *Service, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, logger: logger, now: time.Now}
}

// --- /me/leave ---

func (h *Handler) ListMyLeave(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.callerID(w, r)
	if !ok {
		return
	}
	reqs, err := h.svc.ListMyLeave(r.Context(), userID)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toLeaveList(reqs))
}

func (h *Handler) CreateMyLeave(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.callerID(w, r)
	if !ok {
		return
	}
	var body oapi.CreateLeaveRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "malformed JSON body")
		return
	}
	req, err := h.svc.CreateMyLeave(r.Context(), userID, CreateInput{
		Type:      string(body.Type),
		StartDate: body.StartDate.Time,
		EndDate:   body.EndDate.Time,
		Note:      body.Note,
	})
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toLeave(req))
}

func (h *Handler) GetMyLeaveBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.callerID(w, r)
	if !ok {
		return
	}
	bal, err := h.svc.MyBalance(r.Context(), userID, h.now())
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, oapi.LeaveBalance{
		Year:          bal.Year,
		EntitledDays:  bal.EntitledDays,
		UsedDays:      bal.UsedDays,
		RemainingDays: bal.RemainingDays,
	})
}

// --- /leave (manager) ---

func (h *Handler) ListLeave(w http.ResponseWriter, r *http.Request, params oapi.ListLeaveParams) {
	var empID *uuid.UUID
	if params.EmployeeId != nil {
		id := uuid.UUID(*params.EmployeeId)
		empID = &id
	}
	reqs, err := h.svc.List(r.Context(), Filter{
		EmployeeID: empID,
		Status:     leaveStatusToStrPtr(params.Status),
	})
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toLeaveList(reqs))
}

func (h *Handler) ApproveLeave(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	h.decide(w, r, id, true)
}

func (h *Handler) RejectLeave(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	h.decide(w, r, id, false)
}

func (h *Handler) decide(w http.ResponseWriter, r *http.Request, id oapi.IdParam, approve bool) {
	userID, ok := h.callerID(w, r)
	if !ok {
		return
	}
	var (
		req Request
		err error
	)
	if approve {
		req, err = h.svc.Approve(r.Context(), id, userID)
	} else {
		req, err = h.svc.Reject(r.Context(), id, userID)
	}
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toLeave(req))
}

// --- helpers ---

func (h *Handler) callerID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
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

func leaveStatusToStrPtr(s *oapi.LeaveStatus) *string {
	if s == nil {
		return nil
	}
	v := string(*s)
	return &v
}

func toLeave(req Request) oapi.LeaveRequest {
	out := oapi.LeaveRequest{
		Id:         req.ID,
		EmployeeId: req.EmployeeID,
		Type:       oapi.LeaveType(req.Type),
		StartDate:  openapi_types.Date{Time: req.StartDate},
		EndDate:    openapi_types.Date{Time: req.EndDate},
		Note:       req.Note,
		Status:     oapi.LeaveStatus(req.Status),
		ReviewedBy: req.ReviewedBy,
		ReviewedAt: req.ReviewedAt,
		CreatedAt:  req.CreatedAt,
		UpdatedAt:  req.UpdatedAt,
	}
	return out
}

func toLeaveList(reqs []Request) []oapi.LeaveRequest {
	out := make([]oapi.LeaveRequest, 0, len(reqs))
	for _, req := range reqs {
		out = append(out, toLeave(req))
	}
	return out
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
		h.logger.ErrorContext(r.Context(), "leave request failed", slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
