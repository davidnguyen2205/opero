-- name: ListSubscriptionsPlatform :many
SELECT
    subscriptions.id,
    subscriptions.tenant_id,
    subscriptions.plan,
    subscriptions.status,
    subscriptions.created_at,
    subscriptions.updated_at,
    tenants.name AS tenant_name,
    tenants.slug AS tenant_slug
FROM subscriptions
JOIN tenants ON tenants.id = subscriptions.tenant_id
WHERE
    (sqlc.narg('tenant_id')::uuid IS NULL OR subscriptions.tenant_id = sqlc.narg('tenant_id')) AND
    (sqlc.narg('status')::text IS NULL OR subscriptions.status = sqlc.narg('status')) AND
    (sqlc.narg('plan')::text IS NULL OR subscriptions.plan = sqlc.narg('plan'))
ORDER BY subscriptions.created_at DESC;

-- name: UpdateSubscriptionPlatform :one
UPDATE subscriptions
SET
    plan = COALESCE(sqlc.narg('plan'), plan),
    status = COALESCE(sqlc.narg('status'), status)
WHERE id = sqlc.arg('id')
RETURNING *;
