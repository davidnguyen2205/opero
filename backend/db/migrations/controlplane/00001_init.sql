-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE tenants (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name       text NOT NULL,
    slug       text NOT NULL UNIQUE,
    db_name    text NOT NULL UNIQUE,
    status     text NOT NULL DEFAULT 'provisioning'
                   CHECK (status IN ('active', 'suspended', 'provisioning')),
    plan       text NOT NULL DEFAULT 'free',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER tenants_set_updated_at BEFORE UPDATE ON tenants
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE users (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email         text NOT NULL,
    password_hash text NOT NULL,
    role          text NOT NULL CHECK (role IN ('admin', 'manager', 'employee')),
    status        text NOT NULL DEFAULT 'active'
                      CHECK (status IN ('active', 'disabled')),
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- Email is unique per tenant, case-insensitively.
-- +goose StatementBegin
CREATE UNIQUE INDEX users_tenant_email_unique ON users (tenant_id, lower(email));
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER users_set_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
-- +goose StatementEnd

-- Minimal billing stub; fleshed out in phase 2.
-- +goose StatementBegin
CREATE TABLE subscriptions (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    plan       text NOT NULL DEFAULT 'free',
    status     text NOT NULL DEFAULT 'active',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER subscriptions_set_updated_at BEFORE UPDATE ON subscriptions
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS subscriptions;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS tenants;
-- +goose StatementEnd
-- +goose StatementBegin
DROP FUNCTION IF EXISTS set_updated_at();
-- +goose StatementEnd
