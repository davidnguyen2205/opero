-- name: CreateTenant :one
INSERT INTO tenants (name, slug, db_name, status, plan)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetTenantByID :one
SELECT * FROM tenants WHERE id = $1;

-- name: GetTenantBySlug :one
SELECT * FROM tenants WHERE slug = $1;

-- name: SetTenantStatus :one
UPDATE tenants SET status = $2 WHERE id = $1
RETURNING *;

-- name: DeleteTenant :exec
DELETE FROM tenants WHERE id = $1;

-- name: ListTenants :many
SELECT * FROM tenants ORDER BY created_at;

-- name: UpdateTenantPlatform :one
UPDATE tenants
SET
    name = COALESCE(sqlc.narg('name'), name),
    status = COALESCE(sqlc.narg('status'), status),
    plan = COALESCE(sqlc.narg('plan'), plan)
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: CountTenantsByStatus :many
SELECT status, count(*)::bigint AS count
FROM tenants
GROUP BY status
ORDER BY status;
