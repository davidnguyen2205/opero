package attendance

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	attendancedb "github.com/davidnguyen2205/opero/backend/gen/sqlc/attendance"
)

// Store is the only place that touches the tenant database for this module.
type Store struct {
	q *attendancedb.Queries
}

func NewStore(db attendancedb.DBTX) *Store {
	return &Store{q: attendancedb.New(db)}
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
		case "23503": // bad employee_id/shift_id reference
			return fmt.Errorf("%w: referenced record does not exist", ErrValidation)
		case "23505": // duplicate client_id
			return ErrConflict
		}
	}
	return err
}

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

func fromPgTimestamptz(v pgtype.Timestamptz) *time.Time {
	if !v.Valid {
		return nil
	}
	t := v.Time
	return &t
}

func recordFromDB(a attendancedb.AttendanceRecord) Record {
	return Record{
		ID:               a.ID,
		EmployeeID:       a.EmployeeID,
		ShiftID:          fromPgUUID(a.ShiftID),
		ClientID:         a.ClientID,
		CheckInAt:        fromPgTimestamptz(a.CheckInAt),
		CheckInLat:       a.CheckInLat,
		CheckInLng:       a.CheckInLng,
		CheckInPhotoURL:  a.CheckInPhotoUrl,
		CheckOutAt:       fromPgTimestamptz(a.CheckOutAt),
		CheckOutLat:      a.CheckOutLat,
		CheckOutLng:      a.CheckOutLng,
		CheckOutPhotoURL: a.CheckOutPhotoUrl,
		BreakStartedAt:   fromPgTimestamptz(a.BreakStartedAt),
		Status:           a.Status,
		CreatedAt:        a.CreatedAt,
		UpdatedAt:        a.UpdatedAt,
	}
}

func (s *Store) CreateDemoRecord(ctx context.Context, in DemoRecordInput) (Record, error) {
	rec, err := s.q.CreateDemoAttendance(ctx, attendancedb.CreateDemoAttendanceParams{
		EmployeeID:     in.EmployeeID,
		ShiftID:        toPgUUID(in.ShiftID),
		ClientID:       uuid.New(),
		CheckInAt:      toPgTimestamptz(in.CheckInAt),
		CheckInLat:     in.CheckInLat,
		CheckInLng:     in.CheckInLng,
		CheckOutAt:     toPgTimestamptz(in.CheckOutAt),
		CheckOutLat:    in.CheckOutLat,
		CheckOutLng:    in.CheckOutLng,
		BreakStartedAt: toPgTimestamptz(in.BreakStartedAt),
		Status:         in.Status,
	})
	if err != nil {
		return Record{}, fmt.Errorf("create demo attendance: %w", mapErr(err))
	}
	return recordFromDB(rec), nil
}

func (s *Store) DeleteByShiftIDs(ctx context.Context, shiftIDs []uuid.UUID) (int64, error) {
	n, err := s.q.DeleteAttendanceByShiftIDs(ctx, shiftIDs)
	if err != nil {
		return 0, fmt.Errorf("delete attendance by shift ids: %w", mapErr(err))
	}
	return n, nil
}

func (s *Store) GetByClientID(ctx context.Context, clientID uuid.UUID) (Record, error) {
	a, err := s.q.GetAttendanceByClientID(ctx, clientID)
	if err != nil {
		return Record{}, fmt.Errorf("get attendance by client id: %w", mapErr(err))
	}
	return recordFromDB(a), nil
}

func (s *Store) CreateCheckIn(ctx context.Context, employeeID uuid.UUID, in CheckInInput) (Record, error) {
	a, err := s.q.CreateCheckIn(ctx, attendancedb.CreateCheckInParams{
		EmployeeID:      employeeID,
		ShiftID:         toPgUUID(in.ShiftID),
		ClientID:        in.ClientID,
		CheckInLat:      in.Lat,
		CheckInLng:      in.Lng,
		CheckInPhotoUrl: in.PhotoURL,
	})
	if err != nil {
		return Record{}, fmt.Errorf("create check-in: %w", mapErr(err))
	}
	return recordFromDB(a), nil
}

func (s *Store) CheckOut(ctx context.Context, in CheckOutInput) (Record, error) {
	a, err := s.q.CheckOut(ctx, attendancedb.CheckOutParams{
		ClientID:         in.ClientID,
		CheckOutLat:      in.Lat,
		CheckOutLng:      in.Lng,
		CheckOutPhotoUrl: in.PhotoURL,
	})
	if err != nil {
		return Record{}, fmt.Errorf("check-out: %w", mapErr(err))
	}
	return recordFromDB(a), nil
}

func (s *Store) SetStatus(ctx context.Context, clientID uuid.UUID, status string) (Record, error) {
	a, err := s.q.SetAttendanceStatus(ctx, attendancedb.SetAttendanceStatusParams{
		ClientID: clientID,
		Status:   status,
	})
	if err != nil {
		return Record{}, fmt.Errorf("set attendance status: %w", mapErr(err))
	}
	return recordFromDB(a), nil
}

// ListByShiftIDs returns attendance records linked to any of the given shifts,
// independent of check-in time. Used by the live view's shift⋈attendance join.
func (s *Store) ListByShiftIDs(ctx context.Context, shiftIDs []uuid.UUID) ([]Record, error) {
	if len(shiftIDs) == 0 {
		return nil, nil
	}
	rows, err := s.q.ListAttendanceByShiftIDs(ctx, shiftIDs)
	if err != nil {
		return nil, fmt.Errorf("list attendance by shift ids: %w", mapErr(err))
	}
	out := make([]Record, 0, len(rows))
	for _, a := range rows {
		out = append(out, recordFromDB(a))
	}
	return out, nil
}

func (s *Store) List(ctx context.Context, f AttendanceFilter) ([]Record, error) {
	rows, err := s.q.ListAttendance(ctx, attendancedb.ListAttendanceParams{
		EmployeeID: toPgUUID(f.EmployeeID),
		Status:     f.Status,
		FromTs:     toPgTimestamptz(f.From),
		ToTs:       toPgTimestamptz(f.To),
	})
	if err != nil {
		return nil, fmt.Errorf("list attendance: %w", mapErr(err))
	}
	out := make([]Record, 0, len(rows))
	for _, a := range rows {
		out = append(out, recordFromDB(a))
	}
	return out, nil
}
