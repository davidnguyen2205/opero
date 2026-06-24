package controlplane

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
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

func (h *Handler) PlatformLogin(w http.ResponseWriter, r *http.Request) {
	var body oapi.PlatformLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "malformed JSON body")
		return
	}
	res, err := h.svc.PlatformLogin(r.Context(), PlatformLoginInput{
		Email:    string(body.Email),
		Password: body.Password,
	})
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toPlatformAuthResponse(res))
}

func (h *Handler) GetCurrentPlatformUser(w http.ResponseWriter, r *http.Request) {
	actorID, ok := platformActorID(w, r)
	if !ok {
		return
	}
	user, err := h.svc.CurrentPlatformUser(r.Context(), actorID)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, oapi.CurrentPlatformUserResponse{User: toPlatformUserSummary(user)})
}

func (h *Handler) PlatformListTenants(w http.ResponseWriter, r *http.Request) {
	tenants, err := h.svc.PlatformListTenants(r.Context())
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	out := make([]oapi.PlatformTenant, 0, len(tenants))
	for _, tenant := range tenants {
		out = append(out, toPlatformTenant(tenant))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) PlatformGetTenant(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	tenant, err := h.svc.PlatformGetTenant(r.Context(), uuidFromOAPI(id))
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toPlatformTenant(tenant))
}

func (h *Handler) PlatformUpdateTenant(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	actorID, ok := platformActorID(w, r)
	if !ok {
		return
	}
	var body oapi.PlatformUpdateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "malformed JSON body")
		return
	}
	var status *string
	if body.Status != nil {
		s := string(*body.Status)
		status = &s
	}
	tenant, err := h.svc.PlatformUpdateTenant(r.Context(), actorID, uuidFromOAPI(id), body.Name, status, body.Plan)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toPlatformTenant(tenant))
}

func (h *Handler) PlatformListUsers(w http.ResponseWriter, r *http.Request, params oapi.PlatformListUsersParams) {
	var role *string
	if params.Role != nil {
		r := string(*params.Role)
		role = &r
	}
	var status *string
	if params.Status != nil {
		s := string(*params.Status)
		status = &s
	}
	users, err := h.svc.PlatformListUsers(r.Context(), uuidPtrFromOAPI(params.TenantId), role, status)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	out := make([]oapi.PlatformTenantUser, 0, len(users))
	for _, user := range users {
		out = append(out, toPlatformTenantUser(user))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) PlatformUpdateUser(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	actorID, ok := platformActorID(w, r)
	if !ok {
		return
	}
	var body oapi.PlatformUpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "malformed JSON body")
		return
	}
	user, err := h.svc.PlatformUpdateUser(r.Context(), actorID, uuidFromOAPI(id), string(body.Status))
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toUserSummary(user))
}

func (h *Handler) PlatformListSubscriptions(w http.ResponseWriter, r *http.Request, params oapi.PlatformListSubscriptionsParams) {
	subscriptions, err := h.svc.PlatformListSubscriptions(r.Context(), uuidPtrFromOAPI(params.TenantId), params.Plan, params.Status)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	out := make([]oapi.PlatformSubscription, 0, len(subscriptions))
	for _, sub := range subscriptions {
		out = append(out, toPlatformSubscription(sub))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) PlatformUpdateSubscription(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	actorID, ok := platformActorID(w, r)
	if !ok {
		return
	}
	var body oapi.PlatformUpdateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "malformed JSON body")
		return
	}
	sub, err := h.svc.PlatformUpdateSubscription(r.Context(), actorID, uuidFromOAPI(id), body.Plan, body.Status)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toPlatformSubscription(sub))
}

func (h *Handler) PlatformGetSystemHealth(w http.ResponseWriter, r *http.Request) {
	health, err := h.svc.PlatformSystemHealth(r.Context())
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, oapi.PlatformSystemHealth{
		ControlPlane:    oapi.Ok,
		TenantsByStatus: health.TenantsByStatus,
	})
}

func (h *Handler) PlatformListAuditEvents(w http.ResponseWriter, r *http.Request, params oapi.PlatformListAuditEventsParams) {
	events, err := h.svc.PlatformListAuditEvents(
		r.Context(),
		uuidPtrFromOAPI(params.TenantId),
		uuidPtrFromOAPI(params.ActorPlatformUserId),
		params.Action,
		int32Value(params.Limit),
	)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	out := make([]oapi.SuperAdminAuditEvent, 0, len(events))
	for _, event := range events {
		out = append(out, toSuperAdminAuditEvent(event))
	}
	writeJSON(w, http.StatusOK, out)
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

func toPlatformUserSummary(u PlatformUser) oapi.PlatformUserSummary {
	return oapi.PlatformUserSummary{
		Id:     openapi_types.UUID(u.ID),
		Email:  openapi_types.Email(u.Email),
		Role:   oapi.PlatformUserSummaryRole(u.Role),
		Status: oapi.PlatformUserSummaryStatus(u.Status),
	}
}

func toAuthResponse(r AuthResult) oapi.AuthResponse {
	return oapi.AuthResponse{
		Token:     r.Token,
		TokenType: oapi.AuthResponseTokenTypeBearer,
		ExpiresAt: r.ExpiresAt,
		User:      toUserSummary(r.User),
		Tenant:    toTenantSummary(r.Tenant),
	}
}

func toPlatformAuthResponse(r PlatformAuthResult) oapi.PlatformAuthResponse {
	return oapi.PlatformAuthResponse{
		Token:     r.Token,
		TokenType: oapi.PlatformAuthResponseTokenTypeBearer,
		ExpiresAt: r.ExpiresAt,
		User:      toPlatformUserSummary(r.User),
	}
}

func toPlatformTenant(t Tenant) oapi.PlatformTenant {
	return oapi.PlatformTenant{
		Id:        openapi_types.UUID(t.ID),
		Name:      t.Name,
		Slug:      t.Slug,
		DbName:    t.DBName,
		Status:    oapi.PlatformTenantStatus(t.Status),
		Plan:      t.Plan,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

func toPlatformTenantUser(u PlatformTenantUser) oapi.PlatformTenantUser {
	return oapi.PlatformTenantUser{
		Id:         openapi_types.UUID(u.ID),
		TenantId:   openapi_types.UUID(u.TenantID),
		TenantName: u.TenantName,
		TenantSlug: u.TenantSlug,
		Email:      openapi_types.Email(u.Email),
		Role:       oapi.PlatformTenantUserRole(u.Role),
		Status:     oapi.PlatformTenantUserStatus(u.Status),
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
	}
}

func toPlatformSubscription(s PlatformSubscription) oapi.PlatformSubscription {
	return oapi.PlatformSubscription{
		Id:         openapi_types.UUID(s.ID),
		TenantId:   openapi_types.UUID(s.TenantID),
		TenantName: s.TenantName,
		TenantSlug: s.TenantSlug,
		Plan:       s.Plan,
		Status:     s.Status,
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
	}
}

func toSuperAdminAuditEvent(e SuperAdminAuditEvent) oapi.SuperAdminAuditEvent {
	return oapi.SuperAdminAuditEvent{
		Id:                  openapi_types.UUID(e.ID),
		ActorPlatformUserId: openapi_types.UUID(e.ActorPlatformUserID),
		ActorEmail:          openapi_types.Email(e.ActorEmail),
		Action:              e.Action,
		TargetType:          e.TargetType,
		TargetId:            uuidPtrToOAPI(e.TargetID),
		TenantId:            uuidPtrToOAPI(e.TenantID),
		TenantName:          e.TenantName,
		TenantSlug:          e.TenantSlug,
		Metadata:            e.Metadata,
		CreatedAt:           e.CreatedAt,
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

func platformActorID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	claims, ok := appmw.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return uuid.Nil, false
	}
	id, err := claims.PlatformUserID()
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "invalid token subject")
		return uuid.Nil, false
	}
	return id, true
}

func uuidFromOAPI(id openapi_types.UUID) uuid.UUID {
	return uuid.UUID(id)
}

func uuidPtrFromOAPI(id *openapi_types.UUID) *uuid.UUID {
	if id == nil {
		return nil
	}
	converted := uuid.UUID(*id)
	return &converted
}

func uuidPtrToOAPI(id *uuid.UUID) *openapi_types.UUID {
	if id == nil {
		return nil
	}
	converted := openapi_types.UUID(*id)
	return &converted
}

func int32Value(v *int) int32 {
	if v == nil {
		return 0
	}
	return int32(*v)
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
