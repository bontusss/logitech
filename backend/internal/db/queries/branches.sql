-- name: ListBranches :many
SELECT id, name, code, address, state, phone, status, created_at, updated_at
FROM branches
ORDER BY name ASC;

-- name: GetBranchByID :one
SELECT id, name, code, address, state, phone, status, created_at, updated_at
FROM branches
WHERE id = $1 LIMIT 1;

-- name: GetBranchByCode :one
SELECT id, name, code, address, state, phone, status, created_at, updated_at
FROM branches
WHERE code = $1 LIMIT 1;
