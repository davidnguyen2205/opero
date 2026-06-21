-- Guard against two concurrent CreateEmployeeLogin calls both linking the same
-- (or duplicate) user to employees: a partial unique index on user_id closes
-- the TOCTOU race in identity.Service.CreateEmployeeLogin at the DB level. One
-- employee per login, and one login per employee. NULLs (employees without a
-- login) are not constrained.

-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX employees_user_id_unique ON employees (user_id) WHERE user_id IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS employees_user_id_unique;
-- +goose StatementEnd
