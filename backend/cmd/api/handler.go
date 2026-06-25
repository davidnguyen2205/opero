package main

import (
	"net/http"

	"github.com/davidnguyen2205/opero/backend/gen/oapi"
	"github.com/davidnguyen2205/opero/backend/internal/attendance"
	"github.com/davidnguyen2205/opero/backend/internal/controlplane"
	"github.com/davidnguyen2205/opero/backend/internal/identity"
	"github.com/davidnguyen2205/opero/backend/internal/leave"
	"github.com/davidnguyen2205/opero/backend/internal/liveview"
	"github.com/davidnguyen2205/opero/backend/internal/media"
	"github.com/davidnguyen2205/opero/backend/internal/roster"
	"github.com/davidnguyen2205/opero/backend/internal/stats"
	"github.com/davidnguyen2205/opero/backend/internal/tours"
)

// apiHandler composes the per-module handlers into the single ServerInterface
// the generated router expects. (The two handler types can't both be embedded
// anonymously — they'd collide on the promoted field name "Handler" — so
// methods are forwarded explicitly.)
type apiHandler struct {
	cp  *controlplane.Handler
	id  *identity.Handler
	rs  *roster.Handler
	at  *attendance.Handler
	lv  *liveview.Handler
	lv2 *leave.Handler
	st  *stats.Handler
	tr  *tours.Handler
	md  *media.Handler
}

var _ oapi.ServerInterface = (*apiHandler)(nil)

// control-plane
func (a *apiHandler) Signup(w http.ResponseWriter, r *http.Request) { a.cp.Signup(w, r) }
func (a *apiHandler) Login(w http.ResponseWriter, r *http.Request)  { a.cp.Login(w, r) }
func (a *apiHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	a.cp.GetCurrentUser(w, r)
}
func (a *apiHandler) PlatformLogin(w http.ResponseWriter, r *http.Request) {
	a.cp.PlatformLogin(w, r)
}
func (a *apiHandler) GetCurrentPlatformUser(w http.ResponseWriter, r *http.Request) {
	a.cp.GetCurrentPlatformUser(w, r)
}
func (a *apiHandler) PlatformListTenants(w http.ResponseWriter, r *http.Request) {
	a.cp.PlatformListTenants(w, r)
}
func (a *apiHandler) PlatformGetTenant(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	a.cp.PlatformGetTenant(w, r, id)
}
func (a *apiHandler) PlatformUpdateTenant(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	a.cp.PlatformUpdateTenant(w, r, id)
}
func (a *apiHandler) PlatformListUsers(w http.ResponseWriter, r *http.Request, params oapi.PlatformListUsersParams) {
	a.cp.PlatformListUsers(w, r, params)
}
func (a *apiHandler) PlatformUpdateUser(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	a.cp.PlatformUpdateUser(w, r, id)
}
func (a *apiHandler) PlatformListSubscriptions(w http.ResponseWriter, r *http.Request, params oapi.PlatformListSubscriptionsParams) {
	a.cp.PlatformListSubscriptions(w, r, params)
}
func (a *apiHandler) PlatformUpdateSubscription(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	a.cp.PlatformUpdateSubscription(w, r, id)
}
func (a *apiHandler) PlatformGetSystemHealth(w http.ResponseWriter, r *http.Request) {
	a.cp.PlatformGetSystemHealth(w, r)
}
func (a *apiHandler) PlatformListAuditEvents(w http.ResponseWriter, r *http.Request, params oapi.PlatformListAuditEventsParams) {
	a.cp.PlatformListAuditEvents(w, r, params)
}

// identity — departments
func (a *apiHandler) ListDepartments(w http.ResponseWriter, r *http.Request) {
	a.id.ListDepartments(w, r)
}
func (a *apiHandler) CreateDepartment(w http.ResponseWriter, r *http.Request) {
	a.id.CreateDepartment(w, r)
}
func (a *apiHandler) GetDepartment(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	a.id.GetDepartment(w, r, id)
}
func (a *apiHandler) UpdateDepartment(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	a.id.UpdateDepartment(w, r, id)
}
func (a *apiHandler) DeleteDepartment(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	a.id.DeleteDepartment(w, r, id)
}

// identity — employees
func (a *apiHandler) ListEmployees(w http.ResponseWriter, r *http.Request, params oapi.ListEmployeesParams) {
	a.id.ListEmployees(w, r, params)
}
func (a *apiHandler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	a.id.CreateEmployee(w, r)
}
func (a *apiHandler) GetEmployee(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	a.id.GetEmployee(w, r, id)
}
func (a *apiHandler) UpdateEmployee(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	a.id.UpdateEmployee(w, r, id)
}
func (a *apiHandler) DeleteEmployee(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	a.id.DeleteEmployee(w, r, id)
}
func (a *apiHandler) CreateEmployeeLogin(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	a.id.CreateEmployeeLogin(w, r, id)
}

// identity — roles
func (a *apiHandler) ListRoles(w http.ResponseWriter, r *http.Request)  { a.id.ListRoles(w, r) }
func (a *apiHandler) CreateRole(w http.ResponseWriter, r *http.Request) { a.id.CreateRole(w, r) }
func (a *apiHandler) GetRole(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	a.id.GetRole(w, r, id)
}
func (a *apiHandler) UpdateRole(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	a.id.UpdateRole(w, r, id)
}
func (a *apiHandler) DeleteRole(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	a.id.DeleteRole(w, r, id)
}

// roster — locations
func (a *apiHandler) ListLocations(w http.ResponseWriter, r *http.Request) {
	a.rs.ListLocations(w, r)
}
func (a *apiHandler) CreateLocation(w http.ResponseWriter, r *http.Request) {
	a.rs.CreateLocation(w, r)
}
func (a *apiHandler) GetLocation(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	a.rs.GetLocation(w, r, id)
}
func (a *apiHandler) UpdateLocation(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	a.rs.UpdateLocation(w, r, id)
}
func (a *apiHandler) DeleteLocation(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	a.rs.DeleteLocation(w, r, id)
}

// roster — shifts
func (a *apiHandler) ListShifts(w http.ResponseWriter, r *http.Request, params oapi.ListShiftsParams) {
	a.rs.ListShifts(w, r, params)
}
func (a *apiHandler) ListMyShifts(w http.ResponseWriter, r *http.Request, params oapi.ListMyShiftsParams) {
	a.rs.ListMyShifts(w, r, params)
}
func (a *apiHandler) CreateShift(w http.ResponseWriter, r *http.Request) {
	a.rs.CreateShift(w, r)
}
func (a *apiHandler) GetShift(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	a.rs.GetShift(w, r, id)
}
func (a *apiHandler) UpdateShift(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	a.rs.UpdateShift(w, r, id)
}
func (a *apiHandler) DeleteShift(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	a.rs.DeleteShift(w, r, id)
}
func (a *apiHandler) PublishShift(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	a.rs.PublishShift(w, r, id)
}

// attendance
func (a *apiHandler) ListAttendance(w http.ResponseWriter, r *http.Request, params oapi.ListAttendanceParams) {
	a.at.ListAttendance(w, r, params)
}
func (a *apiHandler) CheckIn(w http.ResponseWriter, r *http.Request)  { a.at.CheckIn(w, r) }
func (a *apiHandler) CheckOut(w http.ResponseWriter, r *http.Request) { a.at.CheckOut(w, r) }
func (a *apiHandler) SetBreak(w http.ResponseWriter, r *http.Request) { a.at.SetBreak(w, r) }

// liveview
func (a *apiHandler) GetLiveView(w http.ResponseWriter, r *http.Request, params oapi.GetLiveViewParams) {
	a.lv.GetLiveView(w, r, params)
}

// leave
func (a *apiHandler) ListMyLeave(w http.ResponseWriter, r *http.Request) { a.lv2.ListMyLeave(w, r) }
func (a *apiHandler) CreateMyLeave(w http.ResponseWriter, r *http.Request) {
	a.lv2.CreateMyLeave(w, r)
}
func (a *apiHandler) GetMyLeaveBalance(w http.ResponseWriter, r *http.Request) {
	a.lv2.GetMyLeaveBalance(w, r)
}
func (a *apiHandler) ListLeave(w http.ResponseWriter, r *http.Request, params oapi.ListLeaveParams) {
	a.lv2.ListLeave(w, r, params)
}
func (a *apiHandler) ApproveLeave(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	a.lv2.ApproveLeave(w, r, id)
}
func (a *apiHandler) RejectLeave(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	a.lv2.RejectLeave(w, r, id)
}

// stats
func (a *apiHandler) GetMyStats(w http.ResponseWriter, r *http.Request) { a.st.GetMyStats(w, r) }

// media
func (a *apiHandler) UploadMedia(w http.ResponseWriter, r *http.Request) { a.md.UploadMedia(w, r) }

// tours
func (a *apiHandler) ListTours(w http.ResponseWriter, r *http.Request, params oapi.ListToursParams) {
	a.tr.ListTours(w, r, params)
}
func (a *apiHandler) CreateTour(w http.ResponseWriter, r *http.Request) { a.tr.CreateTour(w, r) }
func (a *apiHandler) GetTour(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	a.tr.GetTour(w, r, id)
}
func (a *apiHandler) UpdateTour(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	a.tr.UpdateTour(w, r, id)
}
func (a *apiHandler) DeleteTour(w http.ResponseWriter, r *http.Request, id oapi.IdParam) {
	a.tr.DeleteTour(w, r, id)
}
