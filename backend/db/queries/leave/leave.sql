-- name: CreateLeaveRequest :one
INSERT INTO leave_requests (employee_id, type, start_date, end_date, note, status)
VALUES ($1, $2, $3, $4, $5, 'pending')
RETURNING *;

-- name: GetLeaveRequest :one
SELECT * FROM leave_requests WHERE id = $1;

-- name: ListLeaveRequests :many
SELECT * FROM leave_requests
WHERE (sqlc.narg('employee_id')::uuid IS NULL OR employee_id = sqlc.narg('employee_id'))
  AND (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status'))
ORDER BY created_at DESC;

-- name: SetLeaveStatus :one
UPDATE leave_requests SET
    status      = sqlc.arg('status'),
    reviewed_by = sqlc.narg('reviewed_by'),
    reviewed_at = now(),
    updated_at  = now()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: SumApprovedLeaveDays :one
-- Calendar days (inclusive) consumed by approved requests overlapping the year.
SELECT COALESCE(SUM(end_date - start_date + 1), 0)::bigint AS used_days
FROM leave_requests
WHERE employee_id = $1
  AND status = 'approved'
  AND start_date >= sqlc.arg('year_start')::date
  AND start_date <= sqlc.arg('year_end')::date;

-- name: GetLeaveEntitlement :one
SELECT entitled_days FROM leave_balances
WHERE employee_id = $1 AND year = $2;
