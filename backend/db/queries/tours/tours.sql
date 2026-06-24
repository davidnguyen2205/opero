-- name: CreateTour :one
INSERT INTO tours (
    name, category, meeting_point, duration_min, max_guests,
    guides_needed, drivers_needed, departure_times, price_cents, rating, active, color, description
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
RETURNING *;

-- name: GetTour :one
SELECT * FROM tours WHERE id = $1;

-- name: ListTours :many
SELECT * FROM tours
WHERE (sqlc.narg('category')::text IS NULL OR category = sqlc.narg('category'))
  AND (sqlc.narg('active')::boolean IS NULL OR active = sqlc.narg('active'))
ORDER BY name;

-- name: UpdateTour :one
UPDATE tours SET
    name            = COALESCE(sqlc.narg('name'), name),
    category        = COALESCE(sqlc.narg('category'), category),
    meeting_point   = COALESCE(sqlc.narg('meeting_point'), meeting_point),
    duration_min    = COALESCE(sqlc.narg('duration_min'), duration_min),
    max_guests      = COALESCE(sqlc.narg('max_guests'), max_guests),
    guides_needed   = COALESCE(sqlc.narg('guides_needed'), guides_needed),
    drivers_needed  = COALESCE(sqlc.narg('drivers_needed'), drivers_needed),
    departure_times = COALESCE(sqlc.narg('departure_times'), departure_times),
    price_cents     = COALESCE(sqlc.narg('price_cents'), price_cents),
    rating          = COALESCE(sqlc.narg('rating'), rating),
    active          = COALESCE(sqlc.narg('active'), active),
    color           = COALESCE(sqlc.narg('color'), color),
    description     = COALESCE(sqlc.narg('description'), description),
    updated_at      = now()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteTour :execrows
DELETE FROM tours WHERE id = $1;
