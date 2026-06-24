// Package leave owns time-off requests and per-employee entitlements in the
// tenant database (the v1.1 leave-management slice). Like every tenant-scoped
// module it only ever touches the tenant DB via the request-context pool placed
// by TenantMiddleware — never the control-plane DB, never an ad-hoc connection.
package leave

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrValidation = errors.New("validation failed")
	ErrNotFound   = errors.New("not found")
	// ErrNoTenant means no tenant pool was found in the request context — a
	// programming error (route not behind TenantMiddleware), not a client error.
	ErrNoTenant = errors.New("no tenant in context")
)

// DefaultEntitledDays is used when an employee has no leave_balances row for the
// year, so a balance can always be reported without seeding.
const DefaultEntitledDays = 22

// Leave statuses and types (mirror the DB CHECK constraints and the spec enums).
const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusRejected = "rejected"
)

// Request is a domain leave request — clean Go types (pointers for nullable),
// mapped to/from the generated sqlc and API types at the store/handler edges.
type Request struct {
	ID         uuid.UUID
	EmployeeID uuid.UUID
	Type       string
	StartDate  time.Time
	EndDate    time.Time
	Note       *string
	Status     string
	ReviewedBy *uuid.UUID
	ReviewedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type CreateInput struct {
	Type      string
	StartDate time.Time
	EndDate   time.Time
	Note      *string
}

type Filter struct {
	EmployeeID *uuid.UUID
	Status     *string
}

// Balance is the computed leave balance for an employee in a year.
type Balance struct {
	Year          int
	EntitledDays  int
	UsedDays      int
	RemainingDays int
}
