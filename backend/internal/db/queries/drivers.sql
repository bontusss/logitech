-- name: ListDrivers :many
SELECT * FROM drivers
ORDER BY full_name ASC;

-- name: GetDriverByID :one
SELECT * FROM drivers
WHERE id = $1 LIMIT 1;

-- name: CreateDriver :one
INSERT INTO drivers (id, branch_id, full_name, phone, license_number, license_expiry, status)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateDriver :one
UPDATE drivers
SET branch_id       = COALESCE(sqlc.narg('branch_id'), branch_id),
    full_name        = COALESCE(sqlc.narg('full_name'), full_name),
    phone            = COALESCE(sqlc.narg('phone'), phone),
    license_number   = COALESCE(sqlc.narg('license_number'), license_number),
    license_expiry   = COALESCE(sqlc.narg('license_expiry'), license_expiry),
    status           = COALESCE(sqlc.narg('status'), status),
    updated_at       = now()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: SoftDeleteDriver :one
UPDATE drivers
SET status = 'inactive', updated_at = now()
WHERE id = $1
RETURNING *;
