package identity

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

// Handler is the thin HTTP layer implementing the identity slice of the
// oapi-generated ServerInterface.
type Handler struct {
	svc    *Service
	logger *slog.Logger
}

func NewHandler(svc *Service, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

// --- departments ---

func (h *Handler) ListDepartments(w http.ResponseWriter, r *http.Request) {
	deps, err := h.svc.ListDepartments(r.Context())
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	out := make([]oapi.Department, 0, len(deps))
	for _, d := range deps {
		out = append(out, toDepartment(d))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) CreateDepartment(w http.ResponseWriter, r *http.Request) {
	var body oapi.CreateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "malformed JSON body")
		return
	}
	d, err := h.svc.CreateDepartment(r.Context(), CreateDepartmentInput{
		Name:           body.Name,
		ParentID:       body.ParentId,
		Description:    body.Description,
		LeadEmployeeID: body.LeadEmployeeId,
		Icon:           body.Icon,
		Color:          body.Color,
	})
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toDepartment(d))
}

func (h *Handler) GetDepartment(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	d, err := h.svc.GetDepartment(r.Context(), id)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toDepartment(d))
}

func (h *Handler) UpdateDepartment(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	var body oapi.UpdateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "malformed JSON body")
		return
	}
	d, err := h.svc.UpdateDepartment(r.Context(), id, UpdateDepartmentInput{
		Name:           body.Name,
		ParentID:       body.ParentId,
		Description:    body.Description,
		LeadEmployeeID: body.LeadEmployeeId,
		Icon:           body.Icon,
		Color:          body.Color,
	})
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toDepartment(d))
}

func (h *Handler) DeleteDepartment(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	if err := h.svc.DeleteDepartment(r.Context(), id); err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- employees ---

func (h *Handler) ListEmployees(w http.ResponseWriter, r *http.Request, params oapi.ListEmployeesParams) {
	emps, err := h.svc.ListEmployees(r.Context(), EmployeeFilter{
		DepartmentID: params.DepartmentId,
		Status:       enumToStrPtr(params.Status),
	})
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	out := make([]oapi.Employee, 0, len(emps))
	for _, e := range emps {
		out = append(out, toEmployee(e))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	var body oapi.CreateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "malformed JSON body")
		return
	}
	status := ""
	if body.Status != nil {
		status = string(*body.Status)
	}
	e, err := h.svc.CreateEmployee(r.Context(), CreateEmployeeInput{
		UserID:                body.UserId,
		RoleID:                body.RoleId,
		FullName:              body.FullName,
		Email:                 emailToStr(body.Email),
		Phone:                 body.Phone,
		EmploymentType:        string(body.EmploymentType),
		DepartmentID:          body.DepartmentId,
		Title:                 body.Title,
		Status:                status,
		HiredAt:               dateToTime(body.HiredAt),
		Location:              body.Location,
		Languages:             derefLangs(body.Languages),
		EmergencyContactName:  body.EmergencyContactName,
		EmergencyContactPhone: body.EmergencyContactPhone,
		ReportsTo:             body.ReportsTo,
		EmployeeCode:          body.EmployeeCode,
	})
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toEmployee(e))
}

func (h *Handler) GetEmployee(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	e, err := h.svc.GetEmployee(r.Context(), id)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toEmployee(e))
}

func (h *Handler) UpdateEmployee(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	var body oapi.UpdateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "malformed JSON body")
		return
	}
	e, err := h.svc.UpdateEmployee(r.Context(), id, UpdateEmployeeInput{
		FullName:              body.FullName,
		EmploymentType:        enumToStrPtr(body.EmploymentType),
		Email:                 emailToStr(body.Email),
		Phone:                 body.Phone,
		DepartmentID:          body.DepartmentId,
		Title:                 body.Title,
		Status:                enumToStrPtr(body.Status),
		HiredAt:               dateToTime(body.HiredAt),
		RoleID:                body.RoleId,
		Location:              body.Location,
		Languages:             derefLangs(body.Languages),
		EmergencyContactName:  body.EmergencyContactName,
		EmergencyContactPhone: body.EmergencyContactPhone,
		ReportsTo:             body.ReportsTo,
		EmployeeCode:          body.EmployeeCode,
	})
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toEmployee(e))
}

func (h *Handler) DeleteEmployee(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	if err := h.svc.DeleteEmployee(r.Context(), id); err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- roles ---

func (h *Handler) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.svc.ListRoles(r.Context())
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	out := make([]oapi.Role, 0, len(roles))
	for _, role := range roles {
		out = append(out, toRole(role))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var body oapi.CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "malformed JSON body")
		return
	}
	in := CreateRoleInput{Name: body.Name, Description: body.Description, DepartmentID: body.DepartmentId}
	if body.Permissions != nil {
		in.Permissions = *body.Permissions
	}
	if body.AccessLevel != nil {
		in.AccessLevel = string(*body.AccessLevel)
	}
	role, err := h.svc.CreateRole(r.Context(), in)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toRole(role))
}

func (h *Handler) GetRole(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	role, err := h.svc.GetRole(r.Context(), id)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toRole(role))
}

func (h *Handler) UpdateRole(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	var body oapi.UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "malformed JSON body")
		return
	}
	in := UpdateRoleInput{Name: body.Name, Description: body.Description, DepartmentID: body.DepartmentId}
	if body.Permissions != nil {
		in.Permissions = *body.Permissions
	}
	if body.AccessLevel != nil {
		access := string(*body.AccessLevel)
		in.AccessLevel = &access
	}
	role, err := h.svc.UpdateRole(r.Context(), id, in)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toRole(role))
}

func (h *Handler) DeleteRole(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	if err := h.svc.DeleteRole(r.Context(), id); err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) CreateEmployeeLogin(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	claims, ok := appmw.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}
	var body oapi.CreateLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "malformed JSON body")
		return
	}
	role := ""
	if body.Role != nil {
		role = string(*body.Role)
	}
	res, err := h.svc.CreateEmployeeLogin(r.Context(), claims.TenantID, id, string(body.Email), body.Password, role)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, oapi.UserSummary{
		Id:     res.UserID,
		Email:  openapi_types.Email(res.Email),
		Role:   oapi.UserSummaryRole(res.Role),
		Status: oapi.UserSummaryStatusActive,
	})
}

// --- mapping domain -> generated API types ---

func toDepartment(d Department) oapi.Department {
	return oapi.Department{
		Id:             d.ID,
		Name:           d.Name,
		ParentId:       d.ParentID,
		Description:    d.Description,
		LeadEmployeeId: d.LeadEmployeeID,
		Icon:           d.Icon,
		Color:          d.Color,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
	}
}

func toRole(r Role) oapi.Role {
	return oapi.Role{
		Id:           r.ID,
		Name:         r.Name,
		Description:  r.Description,
		DepartmentId: r.DepartmentID,
		AccessLevel:  oapi.AccessLevel(r.AccessLevel),
		Permissions:  r.Permissions,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}

func toEmployee(e Employee) oapi.Employee {
	return oapi.Employee{
		Id:                    e.ID,
		UserId:                e.UserID,
		RoleId:                e.RoleID,
		FullName:              e.FullName,
		Email:                 strToEmail(e.Email),
		Phone:                 e.Phone,
		EmploymentType:        oapi.EmployeeEmploymentType(e.EmploymentType),
		DepartmentId:          e.DepartmentID,
		Title:                 e.Title,
		Status:                oapi.EmployeeStatus(e.Status),
		HiredAt:               timeToDate(e.HiredAt),
		Location:              e.Location,
		Languages:             langsPtr(e.Languages),
		EmergencyContactName:  e.EmergencyContactName,
		EmergencyContactPhone: e.EmergencyContactPhone,
		ReportsTo:             e.ReportsTo,
		EmployeeCode:          e.EmployeeCode,
		CreatedAt:             e.CreatedAt,
		UpdatedAt:             e.UpdatedAt,
	}
}

// langsPtr returns nil for an empty slice (so it serialises as omitted), else a pointer.
func langsPtr(v []string) *[]string {
	if len(v) == 0 {
		return nil
	}
	return &v
}

// derefLangs unwraps the optional languages array (nil → nil, meaning "unchanged"
// on update / empty on create).
func derefLangs(v *[]string) []string {
	if v == nil {
		return nil
	}
	return *v
}

// --- small type converters between generated and domain nullable types ---

func emailToStr(e *openapi_types.Email) *string {
	if e == nil {
		return nil
	}
	s := string(*e)
	return &s
}

func strToEmail(s *string) *openapi_types.Email {
	if s == nil {
		return nil
	}
	e := openapi_types.Email(*s)
	return &e
}

func dateToTime(d *openapi_types.Date) *time.Time {
	if d == nil {
		return nil
	}
	t := d.Time
	return &t
}

func timeToDate(t *time.Time) *openapi_types.Date {
	if t == nil {
		return nil
	}
	return &openapi_types.Date{Time: *t}
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
	case errors.Is(err, ErrConflict):
		writeError(w, http.StatusConflict, "conflict", "a role with this name already exists")
	case errors.Is(err, ErrInUse):
		writeError(w, http.StatusConflict, "in_use", "cannot delete: still referenced by other records (e.g. shifts)")
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "resource not found")
	default:
		// ErrNoTenant (route not behind TenantMiddleware) or an unexpected error.
		h.logger.ErrorContext(r.Context(), "identity request failed", slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

// compile-time guard: id params are uuid.UUID.
var _ = func(id uuid.UUID) oapi.IdParam { return id }
