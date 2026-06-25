-- name: CreateDepartment :one
INSERT INTO departments (name, parent_id, description, lead_employee_id, icon, color)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetDepartment :one
SELECT * FROM departments WHERE id = $1;

-- name: ListDepartments :many
SELECT * FROM departments ORDER BY name;

-- name: UpdateDepartment :one
UPDATE departments SET
    name             = COALESCE(sqlc.narg('name'), name),
    parent_id        = COALESCE(sqlc.narg('parent_id'), parent_id),
    description      = COALESCE(sqlc.narg('description'), description),
    lead_employee_id = COALESCE(sqlc.narg('lead_employee_id'), lead_employee_id),
    icon             = COALESCE(sqlc.narg('icon'), icon),
    color            = COALESCE(sqlc.narg('color'), color),
    updated_at       = now()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteDepartment :execrows
DELETE FROM departments WHERE id = $1;
