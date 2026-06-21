-- name: CreateDepartment :one
INSERT INTO departments (name, parent_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetDepartment :one
SELECT * FROM departments WHERE id = $1;

-- name: ListDepartments :many
SELECT * FROM departments ORDER BY name;

-- name: UpdateDepartment :one
UPDATE departments SET
    name      = COALESCE(sqlc.narg('name'), name),
    parent_id = COALESCE(sqlc.narg('parent_id'), parent_id),
    updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteDepartment :execrows
DELETE FROM departments WHERE id = $1;
