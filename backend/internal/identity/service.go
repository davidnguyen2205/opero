package identity

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	appmw "github.com/davidnguyen2205/opero/backend/internal/platform/middleware"
)

// repo is the tenant-DB persistence the service needs (satisfied by *Store).
// Declared as an interface so the service is unit-testable with fakes.
type repo interface {
	CreateDepartment(ctx context.Context, in CreateDepartmentInput) (Department, error)
	GetDepartment(ctx context.Context, id uuid.UUID) (Department, error)
	ListDepartments(ctx context.Context) ([]Department, error)
	UpdateDepartment(ctx context.Context, id uuid.UUID, in UpdateDepartmentInput) (Department, error)
	DeleteDepartment(ctx context.Context, id uuid.UUID) error
	CreateEmployee(ctx context.Context, in CreateEmployeeInput) (Employee, error)
	GetEmployee(ctx context.Context, id uuid.UUID) (Employee, error)
	GetEmployeeByUserID(ctx context.Context, userID uuid.UUID) (Employee, error)
	SetEmployeeUserID(ctx context.Context, employeeID, userID uuid.UUID) (Employee, error)
	ListEmployees(ctx context.Context, f EmployeeFilter) ([]Employee, error)
	UpdateEmployee(ctx context.Context, id uuid.UUID, in UpdateEmployeeInput) (Employee, error)
	DeleteEmployee(ctx context.Context, id uuid.UUID) error
	CreateRole(ctx context.Context, in CreateRoleInput) (Role, error)
	GetRole(ctx context.Context, id uuid.UUID) (Role, error)
	ListRoles(ctx context.Context) ([]Role, error)
	UpdateRole(ctx context.Context, id uuid.UUID, in UpdateRoleInput) (Role, error)
	DeleteRole(ctx context.Context, id uuid.UUID) error
}

// UserCreator provisions control-plane logins. Satisfied by controlplane.Service.
// Injected so login provisioning can mint a user without identity reaching into
// the control-plane database itself.
type UserCreator interface {
	CreateUser(ctx context.Context, tenantID uuid.UUID, email, password, role string) (uuid.UUID, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
}

// Service holds identity business logic. The tenant store is resolved per
// request from the context pool (placed by TenantMiddleware) via newStore,
// honouring the rule that services use only the request-scoped tenant handle.
type Service struct {
	newStore func(ctx context.Context) (repo, error)
	users    UserCreator
	logger   *slog.Logger
}

func NewService(logger *slog.Logger, users UserCreator) *Service {
	s := &Service{logger: logger, users: users}
	s.newStore = s.tenantStore
	return s
}

func (s *Service) tenantStore(ctx context.Context) (repo, error) {
	pool, ok := appmw.TenantPoolFromContext(ctx)
	if !ok {
		return nil, ErrNoTenant
	}
	return NewStore(pool), nil
}

// --- departments ---

func (s *Service) CreateDepartment(ctx context.Context, in CreateDepartmentInput) (Department, error) {
	if strings.TrimSpace(in.Name) == "" {
		return Department{}, fmt.Errorf("%w: name is required", ErrValidation)
	}
	st, err := s.newStore(ctx)
	if err != nil {
		return Department{}, err
	}
	return st.CreateDepartment(ctx, in)
}

func (s *Service) GetDepartment(ctx context.Context, id uuid.UUID) (Department, error) {
	st, err := s.newStore(ctx)
	if err != nil {
		return Department{}, err
	}
	return st.GetDepartment(ctx, id)
}

func (s *Service) ListDepartments(ctx context.Context) ([]Department, error) {
	st, err := s.newStore(ctx)
	if err != nil {
		return nil, err
	}
	return st.ListDepartments(ctx)
}

func (s *Service) UpdateDepartment(ctx context.Context, id uuid.UUID, in UpdateDepartmentInput) (Department, error) {
	if in.Name != nil && strings.TrimSpace(*in.Name) == "" {
		return Department{}, fmt.Errorf("%w: name must not be empty", ErrValidation)
	}
	st, err := s.newStore(ctx)
	if err != nil {
		return Department{}, err
	}
	return st.UpdateDepartment(ctx, id, in)
}

func (s *Service) DeleteDepartment(ctx context.Context, id uuid.UUID) error {
	st, err := s.newStore(ctx)
	if err != nil {
		return err
	}
	return st.DeleteDepartment(ctx, id)
}

// --- employees ---

func (s *Service) CreateEmployee(ctx context.Context, in CreateEmployeeInput) (Employee, error) {
	if strings.TrimSpace(in.FullName) == "" {
		return Employee{}, fmt.Errorf("%w: full_name is required", ErrValidation)
	}
	if !validEmploymentTypes[in.EmploymentType] {
		return Employee{}, fmt.Errorf("%w: invalid employment_type", ErrValidation)
	}
	if in.Status == "" {
		in.Status = "active"
	} else if !validStatuses[in.Status] {
		return Employee{}, fmt.Errorf("%w: invalid status", ErrValidation)
	}
	st, err := s.newStore(ctx)
	if err != nil {
		return Employee{}, err
	}
	return st.CreateEmployee(ctx, in)
}

func (s *Service) GetEmployee(ctx context.Context, id uuid.UUID) (Employee, error) {
	st, err := s.newStore(ctx)
	if err != nil {
		return Employee{}, err
	}
	return st.GetEmployee(ctx, id)
}

func (s *Service) ListEmployees(ctx context.Context, f EmployeeFilter) ([]Employee, error) {
	if f.Status != nil && !validStatuses[*f.Status] {
		return nil, fmt.Errorf("%w: invalid status filter", ErrValidation)
	}
	st, err := s.newStore(ctx)
	if err != nil {
		return nil, err
	}
	return st.ListEmployees(ctx, f)
}

func (s *Service) UpdateEmployee(ctx context.Context, id uuid.UUID, in UpdateEmployeeInput) (Employee, error) {
	if in.FullName != nil && strings.TrimSpace(*in.FullName) == "" {
		return Employee{}, fmt.Errorf("%w: full_name must not be empty", ErrValidation)
	}
	if in.EmploymentType != nil && !validEmploymentTypes[*in.EmploymentType] {
		return Employee{}, fmt.Errorf("%w: invalid employment_type", ErrValidation)
	}
	if in.Status != nil && !validStatuses[*in.Status] {
		return Employee{}, fmt.Errorf("%w: invalid status", ErrValidation)
	}
	st, err := s.newStore(ctx)
	if err != nil {
		return Employee{}, err
	}
	return st.UpdateEmployee(ctx, id, in)
}

func (s *Service) DeleteEmployee(ctx context.Context, id uuid.UUID) error {
	st, err := s.newStore(ctx)
	if err != nil {
		return err
	}
	return st.DeleteEmployee(ctx, id)
}

// EmployeeIDByUserID resolves the employee linked to a control-plane user.
// Satisfies the attendance module's EmployeeResolver. The bool is false (with a
// nil error) when the user has no employee record in this tenant.
func (s *Service) EmployeeIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, bool, error) {
	st, err := s.newStore(ctx)
	if err != nil {
		return uuid.Nil, false, err
	}
	emp, err := st.GetEmployeeByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return uuid.Nil, false, nil
		}
		return uuid.Nil, false, err
	}
	return emp.ID, true, nil
}

// LoginResult is what CreateEmployeeLogin returns for the API response.
type LoginResult struct {
	UserID uuid.UUID
	Email  string
	Role   string
}

// CreateEmployeeLogin provisions a control-plane login for an employee and links
// it (employees.user_id). Spans the control-plane DB (via UserCreator) and the
// tenant DB (employee link); on link failure the created user is removed.
func (s *Service) CreateEmployeeLogin(ctx context.Context, tenantID, employeeID uuid.UUID, email, password, role string) (LoginResult, error) {
	st, err := s.newStore(ctx)
	if err != nil {
		return LoginResult{}, err
	}
	emp, err := st.GetEmployee(ctx, employeeID)
	if err != nil {
		return LoginResult{}, err // ErrNotFound
	}
	if emp.UserID != nil {
		return LoginResult{}, fmt.Errorf("%w: employee already has a login", ErrConflict)
	}
	if role == "" {
		role = "employee"
	}

	userID, err := s.users.CreateUser(ctx, tenantID, email, password, role)
	if err != nil {
		return LoginResult{}, err // ErrValidation / ErrConflict (dup email) bubble up
	}
	if _, err := st.SetEmployeeUserID(ctx, employeeID, userID); err != nil {
		// Compensation: the login was created but couldn't be linked.
		if delErr := s.users.DeleteUser(ctx, userID); delErr != nil {
			s.logger.ErrorContext(ctx, "cleanup: delete orphan login failed",
				slog.Any("error", delErr))
		}
		return LoginResult{}, err
	}
	return LoginResult{UserID: userID, Email: strings.TrimSpace(email), Role: role}, nil
}

// --- roles ---

func (s *Service) CreateRole(ctx context.Context, in CreateRoleInput) (Role, error) {
	if strings.TrimSpace(in.Name) == "" {
		return Role{}, fmt.Errorf("%w: name is required", ErrValidation)
	}
	st, err := s.newStore(ctx)
	if err != nil {
		return Role{}, err
	}
	return st.CreateRole(ctx, in)
}

func (s *Service) GetRole(ctx context.Context, id uuid.UUID) (Role, error) {
	st, err := s.newStore(ctx)
	if err != nil {
		return Role{}, err
	}
	return st.GetRole(ctx, id)
}

func (s *Service) ListRoles(ctx context.Context) ([]Role, error) {
	st, err := s.newStore(ctx)
	if err != nil {
		return nil, err
	}
	return st.ListRoles(ctx)
}

func (s *Service) UpdateRole(ctx context.Context, id uuid.UUID, in UpdateRoleInput) (Role, error) {
	if in.Name != nil && strings.TrimSpace(*in.Name) == "" {
		return Role{}, fmt.Errorf("%w: name must not be empty", ErrValidation)
	}
	st, err := s.newStore(ctx)
	if err != nil {
		return Role{}, err
	}
	return st.UpdateRole(ctx, id, in)
}

func (s *Service) DeleteRole(ctx context.Context, id uuid.UUID) error {
	st, err := s.newStore(ctx)
	if err != nil {
		return err
	}
	return st.DeleteRole(ctx, id)
}
