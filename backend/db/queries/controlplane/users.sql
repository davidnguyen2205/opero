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
