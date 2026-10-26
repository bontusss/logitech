-- name: ListFleets :many
SELECT * FROM fleet
ORDER BY registration_number ASC;

-- name: GetFleetByID :one
SELECT * FROM fleet
WHERE id = $1 LIMIT 1;

-- name: GetFleetByRegistration :one
SELECT * FROM fleet
WHERE registration_number = $1 LIMIT 1;

-- name: CreateFleet :one
INSERT INTO fleet (id, branch_id, registration_number, vehicle_type, make_model, capacity_kg, status)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateFleet :one
UPDATE fleet
SET branch_id       = COALESCE(sqlc.narg('branch_id'), branch_id),
    vehicle_type     = COALESCE(sqlc.narg('vehicle_type'), vehicle_type),
    make_model       = COALESCE(sqlc.narg('make_model'), make_model),
    capacity_kg      = COALESCE(sqlc.narg('capacity_kg'), capacity_kg),
    status           = COALESCE(sqlc.narg('status'), status),
    updated_at       = now()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: SoftDeleteFleet :one
UPDATE fleet
SET status = 'inactive', updated_at = now()
WHERE id = $1
RETURNING *;
