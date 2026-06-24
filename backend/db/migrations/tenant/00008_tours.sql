-- +goose Up
-- +goose StatementBegin
CREATE TABLE tours (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name            text NOT NULL,
    category        text NOT NULL
                        CHECK (category IN ('walking', 'day_trip', 'food', 'driving', 'evening')),
    meeting_point   text,
    duration_min    integer NOT NULL DEFAULT 120 CHECK (duration_min >= 0),
    max_guests      integer NOT NULL DEFAULT 10 CHECK (max_guests >= 0),
    guides_needed   integer NOT NULL DEFAULT 1 CHECK (guides_needed >= 0),
    drivers_needed  integer NOT NULL DEFAULT 0 CHECK (drivers_needed >= 0),
    departure_times text[] NOT NULL DEFAULT '{}',
    price_cents     integer NOT NULL DEFAULT 0 CHECK (price_cents >= 0),
    rating          double precision CHECK (rating >= 0 AND rating <= 5),
    active          boolean NOT NULL DEFAULT true,
    color           text,
    description     text,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX tours_category_idx ON tours (category);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER tours_set_updated_at BEFORE UPDATE ON tours
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tours;
-- +goose StatementEnd
