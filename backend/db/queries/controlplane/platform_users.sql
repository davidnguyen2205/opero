-- name: CreatePlatformUser :one
INSERT INTO platform_users (email, password_hash, role, status)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetPlatformUserByID :one
SELECT * FROM platform_users WHERE id = $1;

-- name: GetPlatformUserByEmail :one
SELECT * FROM platform_users
WHERE lower(email) = lower($1);

-- name: ListPlatformUsers :many
SELECT * FROM platform_users
ORDER BY created_at DESC;

-- name: UpdatePlatformUserStatus :one
UPDATE platform_users
SET status = $2
WHERE id = $1
RETURNING *;
