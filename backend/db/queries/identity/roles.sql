-- name: CreateRole :one
INSERT INTO roles (name, description, department_id, access_level, permissions)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetRole :one
SELECT * FROM roles WHERE id = $1;

-- name: ListRoles :many
SELECT * FROM roles ORDER BY name;

-- name: UpdateRole :one
UPDATE roles SET
    name          = COALESCE(sqlc.narg('name'), name),
    description   = COALESCE(sqlc.narg('description'), description),
    department_id = COALESCE(sqlc.narg('department_id'), department_id),
    access_level  = COALESCE(sqlc.narg('access_level'), access_level),
    permissions   = COALESCE(sqlc.narg('permissions'), permissions),
    updated_at    = now()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteRole :execrows
DELETE FROM roles WHERE id = $1;
