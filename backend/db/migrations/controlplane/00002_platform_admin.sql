-- +goose Up
-- +goose StatementBegin
CREATE TABLE platform_users (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email         text NOT NULL UNIQUE,
    password_hash text NOT NULL,
    role          text NOT NULL CHECK (role IN ('super_admin', 'support', 'ops')),
    status        text NOT NULL DEFAULT 'active'
                      CHECK (status IN ('active', 'disabled')),
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER platform_users_set_updated_at BEFORE UPDATE ON platform_users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE super_admin_audit_events (
    id                     uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_platform_user_id uuid NOT NULL REFERENCES platform_users(id),
    action                 text NOT NULL,
    target_type            text NOT NULL,
    target_id              uuid NULL,
    tenant_id              uuid NULL REFERENCES tenants(id),
    metadata               jsonb NOT NULL DEFAULT '{}',
    created_at             timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX super_admin_audit_events_created_at_idx
    ON super_admin_audit_events (created_at DESC);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX super_admin_audit_events_tenant_id_idx
    ON super_admin_audit_events (tenant_id)
    WHERE tenant_id IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS super_admin_audit_events;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS platform_users;
-- +goose StatementEnd
