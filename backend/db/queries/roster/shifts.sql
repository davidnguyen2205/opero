-- name: CreateShift :one
INSERT INTO shifts (employee_id, location_id, starts_at, ends_at, notes, status, tour_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetShift :one
SELECT * FROM shifts WHERE id = $1;

-- name: ListShifts :many
SELECT * FROM shifts
WHERE (sqlc.narg('employee_id')::uuid IS NULL OR employee_id = sqlc.narg('employee_id'))
  AND (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('from_ts')::timestamptz IS NULL OR starts_at >= sqlc.narg('from_ts'))
  AND (sqlc.narg('to_ts')::timestamptz IS NULL OR starts_at < sqlc.narg('to_ts'))
ORDER BY starts_at;

-- name: UpdateShift :one
UPDATE shifts SET
    employee_id = COALESCE(sqlc.narg('employee_id'), employee_id),
    location_id = COALESCE(sqlc.narg('location_id'), location_id),
    starts_at   = COALESCE(sqlc.narg('starts_at'), starts_at),
    ends_at     = COALESCE(sqlc.narg('ends_at'), ends_at),
    notes       = COALESCE(sqlc.narg('notes'), notes),
    tour_id     = COALESCE(sqlc.narg('tour_id'), tour_id),
    updated_at  = now()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: PublishShift :one
UPDATE shifts SET status = 'published', updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteShift :execrows
DELETE FROM shifts WHERE id = $1;

-- name: ListShiftIDsByNote :many
-- Demo tooling: find shifts tagged by the seeder (exact note match).
SELECT id FROM shifts WHERE notes = $1;

-- name: DeleteShiftsByNote :execrows
-- Demo tooling: remove shifts tagged by the seeder (exact note match).
DELETE FROM shifts WHERE notes = $1;
