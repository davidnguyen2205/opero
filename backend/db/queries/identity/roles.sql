-- name: CreateRole :one
INSERT INTO roles (name, description, permissions)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetRole :one
SELECT * FROM roles WHERE id = $1;

-- name: ListRoles :many
SELECT * FROM roles ORDER BY name;

-- name: UpdateRole :one
UPDATE roles SET
    name        = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    permissions = COALESCE(sqlc.narg('permissions'), permissions),
    updated_at  = now()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteRole :execrows
DELETE FROM roles WHERE id = $1;
