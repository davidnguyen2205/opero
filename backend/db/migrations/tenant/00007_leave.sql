-- +goose Up
-- +goose StatementBegin
CREATE TABLE leave_requests (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id  uuid NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    type         text NOT NULL CHECK (type IN ('holiday', 'sick', 'personal')),
    start_date   date NOT NULL,
    end_date     date NOT NULL,
    note         text,
    status       text NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'approved', 'rejected')),
    -- Control-plane user id of the reviewing manager (no FK; different database,
    -- mirrors employees.user_id).
    reviewed_by  uuid,
    reviewed_at  timestamptz,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT leave_date_order CHECK (end_date >= start_date)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX leave_requests_employee_idx ON leave_requests (employee_id, start_date);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX leave_requests_status_idx ON leave_requests (status);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER leave_requests_set_updated_at BEFORE UPDATE ON leave_requests
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
-- +goose StatementEnd

-- Per-employee, per-year entitlement. Used days are computed live from approved
-- requests; this table only carries the entitlement (a default applies if absent).
-- +goose StatementBegin
CREATE TABLE leave_balances (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id   uuid NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    year          integer NOT NULL,
    entitled_days integer NOT NULL DEFAULT 22 CHECK (entitled_days >= 0),
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    UNIQUE (employee_id, year)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER leave_balances_set_updated_at BEFORE UPDATE ON leave_balances
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS leave_balances;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS leave_requests;
-- +goose StatementEnd
