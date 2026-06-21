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
CREATE TABLE departments (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name       text NOT NULL,
    parent_id  uuid REFERENCES departments(id) ON DELETE SET NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER departments_set_updated_at BEFORE UPDATE ON departments
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE employees (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    -- Links to control-plane users.id. That table lives in a different
    -- database, so there is intentionally no foreign key here.
    user_id         uuid,
    full_name       text NOT NULL,
    email           text,
    phone           text,
    employment_type text NOT NULL
                        CHECK (employment_type IN ('full_time', 'part_time', 'freelance', 'seasonal')),
    department_id   uuid REFERENCES departments(id) ON DELETE SET NULL,
    title           text,
    status          text NOT NULL DEFAULT 'active'
                        CHECK (status IN ('active', 'inactive')),
    hired_at        date,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX employees_department_id_idx ON employees (department_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER employees_set_updated_at BEFORE UPDATE ON employees
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS employees;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS departments;
-- +goose StatementEnd
-- +goose StatementBegin
DROP FUNCTION IF EXISTS set_updated_at();
-- +goose StatementEnd
