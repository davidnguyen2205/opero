// Package identity owns the people/org core (employees and departments) in the
// tenant database. It only ever touches the tenant DB via the request-context
// pool placed by TenantMiddleware — never the control-plane DB, never an
// ad-hoc connection.
package identity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrValidation = errors.New("validation failed")
	ErrNotFound   = errors.New("not found")
	ErrConflict   = errors.New("already exists")
	// ErrInUse means the record can't be deleted because other records still
	// reference it (e.g. an employee who still has shifts).
	ErrInUse = errors.New("in use")
	// ErrNoTenant means no tenant pool was found in the request context — a
	// programming error (route not behind TenantMiddleware), not a client error.
	ErrNoTenant = errors.New("no tenant in context")
)

// Domain types — clean Go (pointers for nullable), mapped to/from the generated
// sqlc and API types at the store and handler boundaries respectively.

type Department struct {
	ID             uuid.UUID
	Name           string
	ParentID       *uuid.UUID
	Description    *string
	LeadEmployeeID *uuid.UUID
	Icon           *string
	Color          *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Employee struct {
	ID             uuid.UUID
	UserID         *uuid.UUID
	RoleID         *uuid.UUID
	FullName       string
	Email          *string
	Phone          *string
	EmploymentType string
	DepartmentID   *uuid.UUID
	Title          *string
	Status         string
	HiredAt        *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Role struct {
	ID           uuid.UUID
	Name         string
	Description  *string
	DepartmentID *uuid.UUID
	AccessLevel  string
	Permissions  []string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CreateRoleInput struct {
	Name         string
	Description  *string
	DepartmentID *uuid.UUID
	AccessLevel  string
	Permissions  []string
}

type UpdateRoleInput struct {
	Name         *string
	Description  *string
	DepartmentID *uuid.UUID
	AccessLevel  *string
	Permissions  []string
}

type CreateDepartmentInput struct {
	Name           string
	ParentID       *uuid.UUID
	Description    *string
	LeadEmployeeID *uuid.UUID
	Icon           *string
	Color          *string
}

type UpdateDepartmentInput struct {
	Name           *string
	ParentID       *uuid.UUID
	Description    *string
	LeadEmployeeID *uuid.UUID
	Icon           *string
	Color          *string
}

var validAccessLevels = map[string]bool{
	"mobile": true, "web_manager": true, "web_admin": true,
}

type CreateEmployeeInput struct {
	UserID         *uuid.UUID
	RoleID         *uuid.UUID
	FullName       string
	Email          *string
	Phone          *string
	EmploymentType string
	DepartmentID   *uuid.UUID
	Title          *string
	Status         string
	HiredAt        *time.Time
}

type UpdateEmployeeInput struct {
	FullName       *string
	EmploymentType *string
	Email          *string
	Phone          *string
	DepartmentID   *uuid.UUID
	Title          *string
	Status         *string
	HiredAt        *time.Time
	RoleID         *uuid.UUID
}

type EmployeeFilter struct {
	DepartmentID *uuid.UUID
	Status       *string
}

var validEmploymentTypes = map[string]bool{
	"full_time": true, "part_time": true, "freelance": true, "seasonal": true,
}

var validStatuses = map[string]bool{"active": true, "inactive": true}
