-- name: CreateEmployee :one
INSERT INTO employees (
    user_id, full_name, email, phone, employment_type,
    department_id, title, status, hired_at, role_id,
    location, languages, emergency_contact_name, emergency_contact_phone,
    reports_to, employee_code
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
RETURNING *;

-- name: GetEmployee :one
SELECT * FROM employees WHERE id = $1;

-- name: GetEmployeeByUserID :one
SELECT * FROM employees WHERE user_id = $1;

-- name: SetEmployeeUserID :one
UPDATE employees SET user_id = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: ListEmployees :many
SELECT * FROM employees
WHERE (sqlc.narg('department_id')::uuid IS NULL OR department_id = sqlc.narg('department_id'))
  AND (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status'))
ORDER BY full_name;

-- name: UpdateEmployee :one
UPDATE employees SET
    full_name       = COALESCE(sqlc.narg('full_name'), full_name),
    employment_type = COALESCE(sqlc.narg('employment_type'), employment_type),
    email           = COALESCE(sqlc.narg('email'), email),
    phone           = COALESCE(sqlc.narg('phone'), phone),
    department_id   = COALESCE(sqlc.narg('department_id'), department_id),
    title           = COALESCE(sqlc.narg('title'), title),
    status          = COALESCE(sqlc.narg('status'), status),
    hired_at        = COALESCE(sqlc.narg('hired_at'), hired_at),
    role_id         = COALESCE(sqlc.narg('role_id'), role_id),
    location        = COALESCE(sqlc.narg('location'), location),
    languages       = COALESCE(sqlc.narg('languages'), languages),
    emergency_contact_name  = COALESCE(sqlc.narg('emergency_contact_name'), emergency_contact_name),
    emergency_contact_phone = COALESCE(sqlc.narg('emergency_contact_phone'), emergency_contact_phone),
    reports_to      = COALESCE(sqlc.narg('reports_to'), reports_to),
    employee_code   = COALESCE(sqlc.narg('employee_code'), employee_code),
    updated_at      = now()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteEmployee :execrows
DELETE FROM employees WHERE id = $1;
