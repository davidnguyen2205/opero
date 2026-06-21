-- +goose Up
-- +goose StatementBegin
CREATE TABLE locations (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name       text NOT NULL,
    address    text,
    lat        double precision,
    lng        double precision,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER locations_set_updated_at BEFORE UPDATE ON locations
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE shifts (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id uuid NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    location_id uuid REFERENCES locations(id) ON DELETE SET NULL,
    starts_at   timestamptz NOT NULL,
    ends_at     timestamptz NOT NULL,
    notes       text,
    status      text NOT NULL DEFAULT 'draft'
                    CHECK (status IN ('draft', 'published')),
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT shifts_time_order CHECK (ends_at > starts_at)
);
-- +goose StatementEnd

-- Supports the common list filters: by employee and by start-time window.
-- +goose StatementBegin
CREATE INDEX shifts_employee_starts_idx ON shifts (employee_id, starts_at);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX shifts_starts_idx ON shifts (starts_at);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER shifts_set_updated_at BEFORE UPDATE ON shifts
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS shifts;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS locations;
-- +goose StatementEnd
