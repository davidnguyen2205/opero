-- Change shifts.employee_id from ON DELETE CASCADE to ON DELETE RESTRICT so
-- deleting an employee who still has shifts is blocked (preserving roster and,
-- later, attendance history) rather than silently cascading the deletes away.

-- +goose Up
-- +goose StatementBegin
ALTER TABLE shifts DROP CONSTRAINT shifts_employee_id_fkey;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE shifts ADD CONSTRAINT shifts_employee_id_fkey
    FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE RESTRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE shifts DROP CONSTRAINT shifts_employee_id_fkey;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE shifts ADD CONSTRAINT shifts_employee_id_fkey
    FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE;
-- +goose StatementEnd
