-- +goose Up
-- +goose StatementBegin
CREATE TABLE roles (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name        text NOT NULL,
    description text,
    -- Permission keys granted by this role. Stored now; enforcement is a
    -- separate later pass (no taxonomy pinned yet).
    permissions text[] NOT NULL DEFAULT '{}',
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- Role names are unique per tenant, case-insensitively.
-- +goose StatementBegin
CREATE UNIQUE INDEX roles_name_unique ON roles (lower(name));
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER roles_set_updated_at BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE employees ADD COLUMN role_id uuid REFERENCES roles(id) ON DELETE SET NULL;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX employees_role_id_idx ON employees (role_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE employees DROP COLUMN IF EXISTS role_id;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS roles;
-- +goose StatementEnd
