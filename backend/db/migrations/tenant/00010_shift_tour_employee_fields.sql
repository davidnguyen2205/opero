-- +goose Up
-- Shifts can name the tour they deliver.
-- +goose StatementBegin
ALTER TABLE shifts
    ADD COLUMN tour_id uuid REFERENCES tours(id) ON DELETE SET NULL;
-- +goose StatementEnd

-- Richer employee profile fields.
-- +goose StatementBegin
ALTER TABLE employees
    ADD COLUMN location                text,
    ADD COLUMN languages               text[] NOT NULL DEFAULT '{}',
    ADD COLUMN emergency_contact_name  text,
    ADD COLUMN emergency_contact_phone text,
    ADD COLUMN reports_to              uuid REFERENCES employees(id) ON DELETE SET NULL,
    ADD COLUMN employee_code           text;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE employees
    DROP COLUMN IF EXISTS location,
    DROP COLUMN IF EXISTS languages,
    DROP COLUMN IF EXISTS emergency_contact_name,
    DROP COLUMN IF EXISTS emergency_contact_phone,
    DROP COLUMN IF EXISTS reports_to,
    DROP COLUMN IF EXISTS employee_code;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE shifts DROP COLUMN IF EXISTS tour_id;
-- +goose StatementEnd
