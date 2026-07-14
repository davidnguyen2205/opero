-- +goose Up
-- Track when the current break began so the live view can show break
-- duration. Set when status moves to on_break, cleared on resume/check-out.
-- +goose StatementBegin
ALTER TABLE attendance_records ADD COLUMN break_started_at timestamptz;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE attendance_records DROP COLUMN IF EXISTS break_started_at;
-- +goose StatementEnd
