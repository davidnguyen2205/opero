package identity

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	identitydb "github.com/davidnguyen2205/opero/backend/gen/sqlc/identity"
)

// Store is the only place that touches the tenant database for this module. It
// is constructed per request from the tenant-scoped pool.
type Store struct {
	q *identitydb.Queries
}

// NewStore binds the generated queries to a tenant DB handle (a *pgxpool.Pool
// from the request context).
func NewStore(db identitydb.DBTX) *Store {
	return &Store{q: identitydb.New(db)}
}

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503": // foreign_key_violation — bad department_id/parent_id/role_id reference
			return fmt.Errorf("%w: referenced record does not exist", ErrValidation)
		case "23514": // check_violation — bad enum value
			return fmt.Errorf("%w: invalid field value", ErrValidation)
		case "23505": // unique_violation — e.g. duplicate role name
			return ErrConflict
		}
	}
	return err
}

// --- pgtype conversions (kept here so the rest of the module uses clean Go types) ---

func toPgUUID(p *uuid.UUID) pgtype.UUID {
	if p == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *p, Valid: true}
}

func fromPgUUID(v pgtype.UUID) *uuid.UUID {
	if !v.Valid {
		return nil
	}
	u := uuid.UUID(v.Bytes)
	return &u
}

func toPgDate(p *time.Time) pgtype.Date {
	if p == nil {
		return pgtype.Date{}
	}
	return pgtype.Date{Time: *p, Valid: true}
}

func fromPgDate(v pgtype.Date) *time.Time {
	if !v.Valid {
		return nil
	}
	t := v.Time
	return &t
}

func deptFromDB(d identitydb.Department) Department {
	return Department{
		ID:        d.ID,
		Name:      d.Name,
		ParentID:  fromPgUUID(d.ParentID),
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

func roleFromDB(r identitydb.Role) Role {
	perms := r.Permissions
	if perms == nil {
		perms = []string{}
	}
	return Role{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		Permissions: perms,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

func empFromDB(e identitydb.Employee) Employee {
	return Employee{
		ID:             e.ID,
		UserID:         fromPgUUID(e.UserID),
		RoleID:         fromPgUUID(e.RoleID),
		FullName:       e.FullName,
		Email:          e.Email,
		Phone:          e.Phone,
		EmploymentType: e.EmploymentType,
		DepartmentID:   fromPgUUID(e.DepartmentID),
		Title:          e.Title,
		Status:         e.Status,
		HiredAt:        fromPgDate(e.HiredAt),
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
	}
}

// --- departments ---

func (s *Store) CreateDepartment(ctx context.Context, in CreateDepartmentInput) (Department, error) {
	d, err := s.q.CreateDepartment(ctx, identitydb.CreateDepartmentParams{
		Name:     in.Name,
		ParentID: toPgUUID(in.ParentID),
	})
	if err != nil {
		return Department{}, fmt.Errorf("create department: %w", mapErr(err))
	}
	return deptFromDB(d), nil
}

func (s *Store) GetDepartment(ctx context.Context, id uuid.UUID) (Department, error) {
	d, err := s.q.GetDepartment(ctx, id)
	if err != nil {
		return Department{}, fmt.Errorf("get department: %w", mapErr(err))
	}
	return deptFromDB(d), nil
}

func (s *Store) ListDepartments(ctx context.Context) ([]Department, error) {
	rows, err := s.q.ListDepartments(ctx)
	if err != nil {
		return nil, fmt.Errorf("list departments: %w", mapErr(err))
	}
	out := make([]Department, 0, len(rows))
	for _, d := range rows {
		out = append(out, deptFromDB(d))
	}
	return out, nil
}

func (s *Store) UpdateDepartment(ctx context.Context, id uuid.UUID, in UpdateDepartmentInput) (Department, error) {
	d, err := s.q.UpdateDepartment(ctx, identitydb.UpdateDepartmentParams{
		Name:     in.Name,
		ParentID: toPgUUID(in.ParentID),
		ID:       id,
	})
	if err != nil {
		return Department{}, fmt.Errorf("update department: %w", mapErr(err))
	}
	return deptFromDB(d), nil
}

func (s *Store) DeleteDepartment(ctx context.Context, id uuid.UUID) error {
	n, err := s.q.DeleteDepartment(ctx, id)
	if err != nil {
		return fmt.Errorf("delete department: %w", mapErr(err))
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// --- employees ---

func (s *Store) CreateEmployee(ctx context.Context, in CreateEmployeeInput) (Employee, error) {
	e, err := s.q.CreateEmployee(ctx, identitydb.CreateEmployeeParams{
		UserID:         toPgUUID(in.UserID),
		FullName:       in.FullName,
		Email:          in.Email,
		Phone:          in.Phone,
		EmploymentType: in.EmploymentType,
		DepartmentID:   toPgUUID(in.DepartmentID),
		Title:          in.Title,
		Status:         in.Status,
		HiredAt:        toPgDate(in.HiredAt),
		RoleID:         toPgUUID(in.RoleID),
	})
	if err != nil {
		return Employee{}, fmt.Errorf("create employee: %w", mapErr(err))
	}
	return empFromDB(e), nil
}

func (s *Store) GetEmployee(ctx context.Context, id uuid.UUID) (Employee, error) {
	e, err := s.q.GetEmployee(ctx, id)
	if err != nil {
		return Employee{}, fmt.Errorf("get employee: %w", mapErr(err))
	}
	return empFromDB(e), nil
}

func (s *Store) GetEmployeeByUserID(ctx context.Context, userID uuid.UUID) (Employee, error) {
	e, err := s.q.GetEmployeeByUserID(ctx, toPgUUID(&userID))
	if err != nil {
		return Employee{}, fmt.Errorf("get employee by user id: %w", mapErr(err))
	}
	return empFromDB(e), nil
}

func (s *Store) SetEmployeeUserID(ctx context.Context, employeeID, userID uuid.UUID) (Employee, error) {
	e, err := s.q.SetEmployeeUserID(ctx, identitydb.SetEmployeeUserIDParams{
		ID:     employeeID,
		UserID: toPgUUID(&userID),
	})
	if err != nil {
		return Employee{}, fmt.Errorf("set employee user id: %w", mapErr(err))
	}
	return empFromDB(e), nil
}

func (s *Store) ListEmployees(ctx context.Context, f EmployeeFilter) ([]Employee, error) {
	rows, err := s.q.ListEmployees(ctx, identitydb.ListEmployeesParams{
		DepartmentID: toPgUUID(f.DepartmentID),
		Status:       f.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("list employees: %w", mapErr(err))
	}
	out := make([]Employee, 0, len(rows))
	for _, e := range rows {
		out = append(out, empFromDB(e))
	}
	return out, nil
}

func (s *Store) UpdateEmployee(ctx context.Context, id uuid.UUID, in UpdateEmployeeInput) (Employee, error) {
	e, err := s.q.UpdateEmployee(ctx, identitydb.UpdateEmployeeParams{
		FullName:       in.FullName,
		EmploymentType: in.EmploymentType,
		Email:          in.Email,
		Phone:          in.Phone,
		DepartmentID:   toPgUUID(in.DepartmentID),
		Title:          in.Title,
		Status:         in.Status,
		HiredAt:        toPgDate(in.HiredAt),
		RoleID:         toPgUUID(in.RoleID),
		ID:             id,
	})
	if err != nil {
		return Employee{}, fmt.Errorf("update employee: %w", mapErr(err))
	}
	return empFromDB(e), nil
}

func (s *Store) DeleteEmployee(ctx context.Context, id uuid.UUID) error {
	n, err := s.q.DeleteEmployee(ctx, id)
	if err != nil {
		// A 23503 here means a referencing row (e.g. a shift, ON DELETE
		// RESTRICT) blocked the delete — distinct from the create-time "bad
		// reference" meaning, so surface it as ErrInUse (→ 409), not ErrValidation.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return ErrInUse
		}
		return fmt.Errorf("delete employee: %w", mapErr(err))
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// --- roles ---

func (s *Store) CreateRole(ctx context.Context, in CreateRoleInput) (Role, error) {
	// The column is NOT NULL; an explicit INSERT bypasses its default, so a nil
	// slice would send NULL. Coerce to an empty array. (On UPDATE, nil means
	// "leave unchanged" via COALESCE, so it must not be coerced there.)
	perms := in.Permissions
	if perms == nil {
		perms = []string{}
	}
	r, err := s.q.CreateRole(ctx, identitydb.CreateRoleParams{
		Name:        in.Name,
		Description: in.Description,
		Permissions: perms,
	})
	if err != nil {
		return Role{}, fmt.Errorf("create role: %w", mapErr(err))
	}
	return roleFromDB(r), nil
}

func (s *Store) GetRole(ctx context.Context, id uuid.UUID) (Role, error) {
	r, err := s.q.GetRole(ctx, id)
	if err != nil {
		return Role{}, fmt.Errorf("get role: %w", mapErr(err))
	}
	return roleFromDB(r), nil
}

func (s *Store) ListRoles(ctx context.Context) ([]Role, error) {
	rows, err := s.q.ListRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", mapErr(err))
	}
	out := make([]Role, 0, len(rows))
	for _, r := range rows {
		out = append(out, roleFromDB(r))
	}
	return out, nil
}

func (s *Store) UpdateRole(ctx context.Context, id uuid.UUID, in UpdateRoleInput) (Role, error) {
	r, err := s.q.UpdateRole(ctx, identitydb.UpdateRoleParams{
		Name:        in.Name,
		Description: in.Description,
		Permissions: in.Permissions,
		ID:          id,
	})
	if err != nil {
		return Role{}, fmt.Errorf("update role: %w", mapErr(err))
	}
	return roleFromDB(r), nil
}

func (s *Store) DeleteRole(ctx context.Context, id uuid.UUID) error {
	n, err := s.q.DeleteRole(ctx, id)
	if err != nil {
		return fmt.Errorf("delete role: %w", mapErr(err))
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
