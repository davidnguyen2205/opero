package roster

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	rosterdb "github.com/davidnguyen2205/opero/backend/gen/sqlc/roster"
)

// Store is the only place that touches the tenant database for this module. It
// is constructed per request from the tenant-scoped pool.
type Store struct {
	q *rosterdb.Queries
}

// NewStore binds the generated queries to a tenant DB handle (a *pgxpool.Pool
// from the request context).
func NewStore(db rosterdb.DBTX) *Store {
	return &Store{q: rosterdb.New(db)}
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
		case "23503": // foreign_key_violation — bad employee_id/location_id reference
			return fmt.Errorf("%w: referenced record does not exist", ErrValidation)
		case "23514": // check_violation — time order or status constraint
			return fmt.Errorf("%w: invalid field value", ErrValidation)
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

func toPgTimestamptz(p *time.Time) pgtype.Timestamptz {
	if p == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *p, Valid: true}
}

func locationFromDB(l rosterdb.Location) Location {
	return Location{
		ID:        l.ID,
		Name:      l.Name,
		Address:   l.Address,
		Lat:       l.Lat,
		Lng:       l.Lng,
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
	}
}

func shiftFromDB(s rosterdb.Shift) Shift {
	return Shift{
		ID:         s.ID,
		EmployeeID: s.EmployeeID,
		LocationID: fromPgUUID(s.LocationID),
		StartsAt:   s.StartsAt,
		EndsAt:     s.EndsAt,
		Notes:      s.Notes,
		Status:     s.Status,
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
	}
}

// --- locations ---

func (s *Store) CreateLocation(ctx context.Context, in CreateLocationInput) (Location, error) {
	l, err := s.q.CreateLocation(ctx, rosterdb.CreateLocationParams{
		Name:    in.Name,
		Address: in.Address,
		Lat:     in.Lat,
		Lng:     in.Lng,
	})
	if err != nil {
		return Location{}, fmt.Errorf("create location: %w", mapErr(err))
	}
	return locationFromDB(l), nil
}

func (s *Store) GetLocation(ctx context.Context, id uuid.UUID) (Location, error) {
	l, err := s.q.GetLocation(ctx, id)
	if err != nil {
		return Location{}, fmt.Errorf("get location: %w", mapErr(err))
	}
	return locationFromDB(l), nil
}

func (s *Store) ListLocations(ctx context.Context) ([]Location, error) {
	rows, err := s.q.ListLocations(ctx)
	if err != nil {
		return nil, fmt.Errorf("list locations: %w", mapErr(err))
	}
	out := make([]Location, 0, len(rows))
	for _, l := range rows {
		out = append(out, locationFromDB(l))
	}
	return out, nil
}

func (s *Store) UpdateLocation(ctx context.Context, id uuid.UUID, in UpdateLocationInput) (Location, error) {
	l, err := s.q.UpdateLocation(ctx, rosterdb.UpdateLocationParams{
		Name:    in.Name,
		Address: in.Address,
		Lat:     in.Lat,
		Lng:     in.Lng,
		ID:      id,
	})
	if err != nil {
		return Location{}, fmt.Errorf("update location: %w", mapErr(err))
	}
	return locationFromDB(l), nil
}

func (s *Store) DeleteLocation(ctx context.Context, id uuid.UUID) error {
	n, err := s.q.DeleteLocation(ctx, id)
	if err != nil {
		return fmt.Errorf("delete location: %w", mapErr(err))
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// --- shifts ---

func (s *Store) CreateShift(ctx context.Context, in CreateShiftInput, status string) (Shift, error) {
	sh, err := s.q.CreateShift(ctx, rosterdb.CreateShiftParams{
		EmployeeID: in.EmployeeID,
		LocationID: toPgUUID(in.LocationID),
		StartsAt:   in.StartsAt,
		EndsAt:     in.EndsAt,
		Notes:      in.Notes,
		Status:     status,
	})
	if err != nil {
		return Shift{}, fmt.Errorf("create shift: %w", mapErr(err))
	}
	return shiftFromDB(sh), nil
}

func (s *Store) GetShift(ctx context.Context, id uuid.UUID) (Shift, error) {
	sh, err := s.q.GetShift(ctx, id)
	if err != nil {
		return Shift{}, fmt.Errorf("get shift: %w", mapErr(err))
	}
	return shiftFromDB(sh), nil
}

func (s *Store) ListShifts(ctx context.Context, f ShiftFilter) ([]Shift, error) {
	rows, err := s.q.ListShifts(ctx, rosterdb.ListShiftsParams{
		EmployeeID: toPgUUID(f.EmployeeID),
		Status:     f.Status,
		FromTs:     toPgTimestamptz(f.From),
		ToTs:       toPgTimestamptz(f.To),
	})
	if err != nil {
		return nil, fmt.Errorf("list shifts: %w", mapErr(err))
	}
	out := make([]Shift, 0, len(rows))
	for _, sh := range rows {
		out = append(out, shiftFromDB(sh))
	}
	return out, nil
}

func (s *Store) UpdateShift(ctx context.Context, id uuid.UUID, in UpdateShiftInput) (Shift, error) {
	sh, err := s.q.UpdateShift(ctx, rosterdb.UpdateShiftParams{
		EmployeeID: toPgUUID(in.EmployeeID),
		LocationID: toPgUUID(in.LocationID),
		StartsAt:   toPgTimestamptz(in.StartsAt),
		EndsAt:     toPgTimestamptz(in.EndsAt),
		Notes:      in.Notes,
		ID:         id,
	})
	if err != nil {
		return Shift{}, fmt.Errorf("update shift: %w", mapErr(err))
	}
	return shiftFromDB(sh), nil
}

func (s *Store) PublishShift(ctx context.Context, id uuid.UUID) (Shift, error) {
	sh, err := s.q.PublishShift(ctx, id)
	if err != nil {
		return Shift{}, fmt.Errorf("publish shift: %w", mapErr(err))
	}
	return shiftFromDB(sh), nil
}

func (s *Store) DeleteShift(ctx context.Context, id uuid.UUID) error {
	n, err := s.q.DeleteShift(ctx, id)
	if err != nil {
		return fmt.Errorf("delete shift: %w", mapErr(err))
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
