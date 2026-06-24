-- name: CreateUser :one
INSERT INTO users (tenant_id, email, password_hash, role, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByTenantAndEmail :one
SELECT * FROM users
WHERE tenant_id = $1 AND lower(email) = lower($2);

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: ListUsersPlatform :many
SELECT
    users.id,
    users.tenant_id,
    users.email,
    users.password_hash,
    users.role,
    users.status,
    users.created_at,
    users.updated_at,
    tenants.name AS tenant_name,
    tenants.slug AS tenant_slug
FROM users
JOIN tenants ON tenants.id = users.tenant_id
WHERE
    (sqlc.narg('tenant_id')::uuid IS NULL OR users.tenant_id = sqlc.narg('tenant_id')) AND
    (sqlc.narg('status')::text IS NULL OR users.status = sqlc.narg('status')) AND
    (sqlc.narg('role')::text IS NULL OR users.role = sqlc.narg('role'))
ORDER BY users.created_at DESC;

-- name: UpdateUserStatusPlatform :one
UPDATE users
SET status = $2
WHERE id = $1
RETURNING *;
