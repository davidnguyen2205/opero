-- +goose Up
-- Roles: access level + department link.
-- +goose StatementBegin
ALTER TABLE roles
    ADD COLUMN access_level text NOT NULL DEFAULT 'web_manager'
        CHECK (access_level IN ('mobile', 'web_manager', 'web_admin')),
    ADD COLUMN department_id uuid REFERENCES departments(id) ON DELETE SET NULL;
-- +goose StatementEnd

-- Departments: presentation + lead metadata.
-- +goose StatementBegin
ALTER TABLE departments
    ADD COLUMN description text,
    ADD COLUMN lead_employee_id uuid REFERENCES employees(id) ON DELETE SET NULL,
    ADD COLUMN icon text,
    ADD COLUMN color text;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE departments
    DROP COLUMN IF EXISTS description,
    DROP COLUMN IF EXISTS lead_employee_id,
    DROP COLUMN IF EXISTS icon,
    DROP COLUMN IF EXISTS color;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE roles
    DROP COLUMN IF EXISTS access_level,
    DROP COLUMN IF EXISTS department_id;
-- +goose StatementEnd
