-- name: ListBranches :many
SELECT * FROM branches
ORDER BY name ASC;

-- name: GetBranchByID :one
SELECT * FROM branches
WHERE id = $1 LIMIT 1;

-- name: GetBranchByCode :one
SELECT * FROM branches
WHERE code = $1 LIMIT 1;

-- name: CreateBranch :one
INSERT INTO branches (id, name, code, address, state, phone, status)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateBranch :one
UPDATE branches
SET name    = COALESCE(sqlc.narg('name'), name),
    address = COALESCE(sqlc.narg('address'), address),
    state   = COALESCE(sqlc.narg('state'), state),
    phone   = COALESCE(sqlc.narg('phone'), phone),
    status  = COALESCE(sqlc.narg('status'), status),
    updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: SoftDeleteBranch :one
UPDATE branches
SET status = 'inactive', updated_at = now()
WHERE id = $1
RETURNING *;
