-- name: CreateSuperAdminAuditEvent :one
INSERT INTO super_admin_audit_events (
    actor_platform_user_id,
    action,
    target_type,
    target_id,
    tenant_id,
    metadata
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListSuperAdminAuditEvents :many
SELECT
    super_admin_audit_events.id,
    super_admin_audit_events.actor_platform_user_id,
    platform_users.email AS actor_email,
    super_admin_audit_events.action,
    super_admin_audit_events.target_type,
    super_admin_audit_events.target_id,
    super_admin_audit_events.tenant_id,
    tenants.name AS tenant_name,
    tenants.slug AS tenant_slug,
    super_admin_audit_events.metadata,
    super_admin_audit_events.created_at
FROM super_admin_audit_events
JOIN platform_users ON platform_users.id = super_admin_audit_events.actor_platform_user_id
LEFT JOIN tenants ON tenants.id = super_admin_audit_events.tenant_id
WHERE
    (sqlc.narg('tenant_id')::uuid IS NULL OR super_admin_audit_events.tenant_id = sqlc.narg('tenant_id')) AND
    (sqlc.narg('actor_platform_user_id')::uuid IS NULL OR super_admin_audit_events.actor_platform_user_id = sqlc.narg('actor_platform_user_id')) AND
    (sqlc.narg('action')::text IS NULL OR super_admin_audit_events.action = sqlc.narg('action'))
ORDER BY super_admin_audit_events.created_at DESC
LIMIT sqlc.arg('limit');
