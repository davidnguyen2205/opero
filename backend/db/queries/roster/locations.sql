-- name: CreateLocation :one
INSERT INTO locations (name, address, lat, lng)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetLocation :one
SELECT * FROM locations WHERE id = $1;

-- name: ListLocations :many
SELECT * FROM locations ORDER BY name;

-- name: UpdateLocation :one
UPDATE locations SET
    name    = COALESCE(sqlc.narg('name'), name),
    address = COALESCE(sqlc.narg('address'), address),
    lat     = COALESCE(sqlc.narg('lat'), lat),
    lng     = COALESCE(sqlc.narg('lng'), lng),
    updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteLocation :execrows
DELETE FROM locations WHERE id = $1;
