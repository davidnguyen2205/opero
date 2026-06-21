// Package roster owns shift scheduling and the locations shifts happen at, in
// the tenant database. Like every tenant-scoped module it only ever touches the
// tenant DB via the request-context pool placed by TenantMiddleware — never the
// control-plane DB, never an ad-hoc connection.
package roster

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

// Domain types — clean Go (pointers for nullable), mapped to/from the generated
// sqlc and API types at the store and handler boundaries respectively.

type Location struct {
	ID        uuid.UUID
	Name      string
	Address   *string
	Lat       *float64
	Lng       *float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Shift struct {
	ID         uuid.UUID
	EmployeeID uuid.UUID
	LocationID *uuid.UUID
	StartsAt   time.Time
	EndsAt     time.Time
	Notes      *string
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type CreateLocationInput struct {
	Name    string
	Address *string
	Lat     *float64
	Lng     *float64
}

type UpdateLocationInput struct {
	Name    *string
	Address *string
	Lat     *float64
	Lng     *float64
}

type CreateShiftInput struct {
	EmployeeID uuid.UUID
	LocationID *uuid.UUID
	StartsAt   time.Time
	EndsAt     time.Time
	Notes      *string
}

type UpdateShiftInput struct {
	EmployeeID *uuid.UUID
	LocationID *uuid.UUID
	StartsAt   *time.Time
	EndsAt     *time.Time
	Notes      *string
}

type ShiftFilter struct {
	EmployeeID *uuid.UUID
	Status     *string
	From       *time.Time
	To         *time.Time
}
