-- +goose Up
-- Allow an 'on_break' attendance state.
-- +goose StatementBegin
ALTER TABLE attendance_records DROP CONSTRAINT IF EXISTS attendance_records_status_check;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE attendance_records
    ADD CONSTRAINT attendance_records_status_check
    CHECK (status IN ('checked_in', 'checked_out', 'missed', 'on_break'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE attendance_records DROP CONSTRAINT IF EXISTS attendance_records_status_check;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE attendance_records
    ADD CONSTRAINT attendance_records_status_check
    CHECK (status IN ('checked_in', 'checked_out', 'missed'));
-- +goose StatementEnd
