// Package tours owns the tour catalog in the tenant database — the experiences a
// tenant offers. Like every tenant-scoped module it only ever touches the tenant
// DB via the request-context pool placed by TenantMiddleware.
package tours

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrValidation = errors.New("validation failed")
	ErrNotFound   = errors.New("not found")
	ErrNoTenant   = errors.New("no tenant in context")
)

var validCategories = map[string]bool{
	"walking": true, "day_trip": true, "food": true, "driving": true, "evening": true,
}

// ValidCategory reports whether c is an allowed tour category.
func ValidCategory(c string) bool { return validCategories[c] }

// Tour is the domain catalog entry. Integer fields use Go int (mapped to/from the
// generated sqlc int32 at the store boundary) so they line up with the API types.
type Tour struct {
	ID             uuid.UUID
	Name           string
	Category       string
	MeetingPoint   *string
	DurationMin    int
	MaxGuests      int
	GuidesNeeded   int
	DriversNeeded  int
	DepartureTimes []string
	PriceCents     int
	Rating         *float64
	Active         bool
	Color          *string
	Description    *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CreateInput struct {
	Name           string
	Category       string
	MeetingPoint   *string
	DurationMin    int
	MaxGuests      int
	GuidesNeeded   int
	DriversNeeded  int
	DepartureTimes []string
	PriceCents     int
	Rating         *float64
	Active         bool
	Color          *string
	Description    *string
}

type UpdateInput struct {
	Name           *string
	Category       *string
	MeetingPoint   *string
	DurationMin    *int
	MaxGuests      *int
	GuidesNeeded   *int
	DriversNeeded  *int
	DepartureTimes []string
	PriceCents     *int
	Rating         *float64
	Active         *bool
	Color          *string
	Description    *string
}

type Filter struct {
	Category *string
	Active   *bool
}
