// Package attendance owns field check-in/out in the tenant database. Records are
// keyed for idempotency by a client-supplied id so the offline mobile queue can
// safely replay. Like every tenant-scoped module it only ever touches the tenant
// DB via the request-context pool placed by TenantMiddleware.
package attendance

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrValidation = errors.New("validation failed")
	ErrNotFound   = errors.New("not found")
	ErrConflict   = errors.New("conflict")
	ErrNoTenant   = errors.New("no tenant in context")
)

type Record struct {
	ID               uuid.UUID
	EmployeeID       uuid.UUID
	ShiftID          *uuid.UUID
	ClientID         uuid.UUID
	CheckInAt        *time.Time
	CheckInLat       *float64
	CheckInLng       *float64
	CheckInPhotoURL  *string
	CheckOutAt       *time.Time
	CheckOutLat      *float64
	CheckOutLng      *float64
	CheckOutPhotoURL *string
	BreakStartedAt   *time.Time
	Status           string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// DemoRecordInput is a fully specified attendance record used by the demo
// seeder — fabricated timestamps and state, unlike the real check-in flow
// which always stamps now().
type DemoRecordInput struct {
	EmployeeID     uuid.UUID
	ShiftID        *uuid.UUID
	CheckInAt      *time.Time
	CheckInLat     *float64
	CheckInLng     *float64
	CheckOutAt     *time.Time
	CheckOutLat    *float64
	CheckOutLng    *float64
	BreakStartedAt *time.Time
	Status         string
}

type CheckInInput struct {
	ClientID uuid.UUID
	ShiftID  *uuid.UUID
	Lat      *float64
	Lng      *float64
	PhotoURL *string
}

type CheckOutInput struct {
	ClientID uuid.UUID
	Lat      *float64
	Lng      *float64
	PhotoURL *string
}

type AttendanceFilter struct {
	EmployeeID *uuid.UUID
	Status     *string
	From       *time.Time
	To         *time.Time
}
