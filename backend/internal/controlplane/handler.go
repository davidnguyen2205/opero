package controlplane

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/davidnguyen2205/opero/backend/gen/oapi"
	appmw "github.com/davidnguyen2205/opero/backend/internal/platform/middleware"
)

// Handler is the thin HTTP layer implementing the oapi-generated
// ServerInterface. It decodes/validates input and delegates to the Service.
type Handler struct {
	svc    *Service
	logger *slog.Logger
}

func NewHandler(svc *Service, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var body oapi.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "malformed JSON body")
		return
	}
	res, err := h.svc.Signup(r.Context(), SignupInput{
		CompanyName:   body.CompanyName,
		Slug:          deref(body.Slug),
		AdminFullName: deref(body.AdminFullName),
		AdminEmail:    string(body.AdminEmail),
		AdminPassword: body.AdminPassword,
	})
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toAuthResponse(res))
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var body oapi.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "malformed JSON body")
		return
	}
	res, err := h.svc.Login(r.Context(), LoginInput{
		TenantSlug: body.TenantSlug,
		Email:      string(body.Email),
		Password:   body.Password,
	})
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toAuthResponse(res))
}

func (h *Handler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
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
	res, err := h.svc.CurrentUser(r.Context(), userID)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, oapi.CurrentUserResponse{
		User:   toUserSummary(res.User),
		Tenant: toTenantSummary(res.Tenant),
	})
}

// writeServiceError maps the module's sentinel errors to HTTP statuses.
func (h *Handler) writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrValidation):
		writeError(w, http.StatusBadRequest, "validation_failed", err.Error())
	case errors.Is(err, ErrConflict):
		writeError(w, http.StatusConflict, "conflict", "a tenant with this slug or email already exists")
	case errors.Is(err, ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "invalid credentials")
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "resource not found")
	default:
		// Don't leak internal detail; log it server-side.
		h.logger.ErrorContext(r.Context(), "controlplane request failed", slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

// --- mapping to generated API types ---

func toUserSummary(u User) oapi.UserSummary {
	return oapi.UserSummary{
		Id:     openapi_types.UUID(u.ID),
		Email:  openapi_types.Email(u.Email),
		Role:   oapi.UserSummaryRole(u.Role),
		Status: oapi.UserSummaryStatus(u.Status),
	}
}

func toTenantSummary(t Tenant) oapi.TenantSummary {
	return oapi.TenantSummary{
		Id:     openapi_types.UUID(t.ID),
		Name:   t.Name,
		Slug:   t.Slug,
		Status: oapi.TenantSummaryStatus(t.Status),
		Plan:   t.Plan,
	}
}

func toAuthResponse(r AuthResult) oapi.AuthResponse {
	return oapi.AuthResponse{
		Token:     r.Token,
		TokenType: oapi.Bearer,
		ExpiresAt: r.ExpiresAt,
		User:      toUserSummary(r.User),
		Tenant:    toTenantSummary(r.Tenant),
	}
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

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
