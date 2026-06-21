-- +goose Up
-- +goose StatementBegin
CREATE TABLE attendance_records (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id         uuid NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    shift_id            uuid REFERENCES shifts(id) ON DELETE SET NULL,
    -- Client-generated idempotency key. Unique per tenant DB, so the offline
    -- mobile queue can safely replay check-in/out keyed by it.
    client_id           uuid NOT NULL,
    check_in_at         timestamptz,
    check_in_lat        double precision,
    check_in_lng        double precision,
    check_in_photo_url  text,
    check_out_at        timestamptz,
    check_out_lat       double precision,
    check_out_lng       double precision,
    check_out_photo_url text,
    status              text NOT NULL DEFAULT 'checked_in'
                            CHECK (status IN ('checked_in', 'checked_out', 'missed')),
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE UNIQUE INDEX attendance_client_id_unique ON attendance_records (client_id);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX attendance_employee_checkin_idx ON attendance_records (employee_id, check_in_at);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER attendance_set_updated_at BEFORE UPDATE ON attendance_records
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS attendance_records;
-- +goose StatementEnd
