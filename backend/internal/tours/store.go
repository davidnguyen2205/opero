package tours

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	toursdb "github.com/davidnguyen2205/opero/backend/gen/sqlc/tours"
)

// Store is the only place that touches the tenant database for this module.
type Store struct {
	q *toursdb.Queries
}

func NewStore(db toursdb.DBTX) *Store {
	return &Store{q: toursdb.New(db)}
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
		if pgErr.Code == "23514" { // check_violation — category/range constraints
			return fmt.Errorf("%w: invalid field value", ErrValidation)
		}
	}
	return err
}

func i32(v int) int32 { return int32(v) }
func i32p(v *int) *int32 {
	if v == nil {
		return nil
	}
	x := int32(*v)
	return &x
}

func tourFromDB(t toursdb.Tour) Tour {
	times := t.DepartureTimes
	if times == nil {
		times = []string{}
	}
	return Tour{
		ID:             t.ID,
		Name:           t.Name,
		Category:       t.Category,
		MeetingPoint:   t.MeetingPoint,
		DurationMin:    int(t.DurationMin),
		MaxGuests:      int(t.MaxGuests),
		GuidesNeeded:   int(t.GuidesNeeded),
		DriversNeeded:  int(t.DriversNeeded),
		DepartureTimes: times,
		PriceCents:     int(t.PriceCents),
		Rating:         t.Rating,
		Active:         t.Active,
		Color:          t.Color,
		Description:    t.Description,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
	}
}

func (s *Store) Create(ctx context.Context, in CreateInput) (Tour, error) {
	times := in.DepartureTimes
	if times == nil {
		times = []string{}
	}
	t, err := s.q.CreateTour(ctx, toursdb.CreateTourParams{
		Name:           in.Name,
		Category:       in.Category,
		MeetingPoint:   in.MeetingPoint,
		DurationMin:    i32(in.DurationMin),
		MaxGuests:      i32(in.MaxGuests),
		GuidesNeeded:   i32(in.GuidesNeeded),
		DriversNeeded:  i32(in.DriversNeeded),
		DepartureTimes: times,
		PriceCents:     i32(in.PriceCents),
		Rating:         in.Rating,
		Active:         in.Active,
		Color:          in.Color,
		Description:    in.Description,
	})
	if err != nil {
		return Tour{}, fmt.Errorf("create tour: %w", mapErr(err))
	}
	return tourFromDB(t), nil
}

func (s *Store) Get(ctx context.Context, id uuid.UUID) (Tour, error) {
	t, err := s.q.GetTour(ctx, id)
	if err != nil {
		return Tour{}, fmt.Errorf("get tour: %w", mapErr(err))
	}
	return tourFromDB(t), nil
}

func (s *Store) List(ctx context.Context, f Filter) ([]Tour, error) {
	rows, err := s.q.ListTours(ctx, toursdb.ListToursParams{
		Category: f.Category,
		Active:   f.Active,
	})
	if err != nil {
		return nil, fmt.Errorf("list tours: %w", mapErr(err))
	}
	out := make([]Tour, 0, len(rows))
	for _, t := range rows {
		out = append(out, tourFromDB(t))
	}
	return out, nil
}

func (s *Store) Update(ctx context.Context, id uuid.UUID, in UpdateInput) (Tour, error) {
	t, err := s.q.UpdateTour(ctx, toursdb.UpdateTourParams{
		ID:             id,
		Name:           in.Name,
		Category:       in.Category,
		MeetingPoint:   in.MeetingPoint,
		DurationMin:    i32p(in.DurationMin),
		MaxGuests:      i32p(in.MaxGuests),
		GuidesNeeded:   i32p(in.GuidesNeeded),
		DriversNeeded:  i32p(in.DriversNeeded),
		DepartureTimes: in.DepartureTimes,
		PriceCents:     i32p(in.PriceCents),
		Rating:         in.Rating,
		Active:         in.Active,
		Color:          in.Color,
		Description:    in.Description,
	})
	if err != nil {
		return Tour{}, fmt.Errorf("update tour: %w", mapErr(err))
	}
	return tourFromDB(t), nil
}

func (s *Store) Delete(ctx context.Context, id uuid.UUID) error {
	n, err := s.q.DeleteTour(ctx, id)
	if err != nil {
		return fmt.Errorf("delete tour: %w", mapErr(err))
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
