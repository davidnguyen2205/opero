-- name: GetAttendanceByClientID :one
SELECT * FROM attendance_records WHERE client_id = $1;

-- name: GetAttendance :one
SELECT * FROM attendance_records WHERE id = $1;

-- name: CreateCheckIn :one
INSERT INTO attendance_records (
    employee_id, shift_id, client_id,
    check_in_at, check_in_lat, check_in_lng, check_in_photo_url, status
)
VALUES ($1, $2, $3, now(), $4, $5, $6, 'checked_in')
RETURNING *;

-- name: CheckOut :one
UPDATE attendance_records SET
    check_out_at        = now(),
    check_out_lat       = $2,
    check_out_lng       = $3,
    check_out_photo_url = $4,
    status              = 'checked_out',
    break_started_at    = NULL,
    updated_at          = now()
WHERE client_id = $1
RETURNING *;

-- name: SetAttendanceStatus :one
-- Toggle break state for an open record. Only moves between checked_in and
-- on_break (a checked-out/missed record is left unchanged). break_started_at
-- tracks the current break: stamped on entry, cleared on resume.
UPDATE attendance_records SET
    status           = sqlc.arg('status'),
    break_started_at = CASE WHEN sqlc.arg('status') = 'on_break' THEN now() END,
    updated_at       = now()
WHERE client_id = sqlc.arg('client_id')
  AND status IN ('checked_in', 'on_break')
RETURNING *;

-- name: DeleteAttendanceByShiftIDs :execrows
-- Demo tooling: remove attendance linked to the given (seeded) shifts.
DELETE FROM attendance_records WHERE shift_id = ANY($1::uuid[]);

-- name: CreateDemoAttendance :one
-- Demo tooling: insert a fully specified attendance record (fabricated
-- timestamps/state), unlike CreateCheckIn which always stamps now().
INSERT INTO attendance_records (
    employee_id, shift_id, client_id,
    check_in_at, check_in_lat, check_in_lng,
    check_out_at, check_out_lat, check_out_lng,
    break_started_at, status
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: ListAttendanceByShiftIDs :many
-- Fetch attendance linked to any of the given shifts, regardless of check-in
-- time (used by the live view to join shifts to their current attendance state
-- without a check_in_at window dropping early/overnight check-ins).
SELECT * FROM attendance_records
WHERE shift_id = ANY($1::uuid[]);

-- name: ListAttendance :many
SELECT * FROM attendance_records
WHERE (sqlc.narg('employee_id')::uuid IS NULL OR employee_id = sqlc.narg('employee_id'))
  AND (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('from_ts')::timestamptz IS NULL OR check_in_at >= sqlc.narg('from_ts'))
  AND (sqlc.narg('to_ts')::timestamptz IS NULL OR check_in_at < sqlc.narg('to_ts'))
ORDER BY check_in_at DESC NULLS LAST;
