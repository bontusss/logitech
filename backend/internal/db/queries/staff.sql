-- name: GetStaffByID :one
SELECT id, branch_id, full_name, phone, email, staff_id, pin, password_hash, role, status, created_at, updated_at
FROM staff
WHERE id = $1 LIMIT 1;

-- name: GetStaffByStaffID :one
SELECT id, branch_id, full_name, phone, email, staff_id, pin, password_hash, role, status, created_at, updated_at
FROM staff
WHERE staff_id = $1 LIMIT 1;

-- name: GetStaffByEmail :one
SELECT id, branch_id, full_name, phone, email, staff_id, pin, password_hash, role, status, created_at, updated_at
FROM staff
WHERE email = $1 LIMIT 1;

-- name: CheckPinExists :one
SELECT EXISTS(SELECT 1 FROM staff WHERE pin = $1);

-- name: CheckStaffIDExists :one
SELECT EXISTS(SELECT 1 FROM staff WHERE staff_id = $1);

-- name: GetNextStaffIDSequence :one
SELECT nextval('staff_id_seq')::bigint AS next_val;

-- name: CreateStaff :one
INSERT INTO staff (
    id, branch_id, full_name, phone, email, staff_id, pin, password_hash, role, status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING id, branch_id, full_name, phone, email, staff_id, pin, password_hash, role, status, created_at, updated_at;

-- name: ListStaff :many
SELECT id, branch_id, full_name, phone, email, staff_id, pin, role, status, created_at, updated_at
FROM staff
ORDER BY created_at DESC;

-- name: ListStaffByBranch :many
SELECT id, branch_id, full_name, phone, email, staff_id, pin, role, status, created_at, updated_at
FROM staff
WHERE branch_id = $1
ORDER BY created_at DESC;

-- name: UpdateStaffStatus :one
UPDATE staff
SET status = $2, updated_at = now()
WHERE id = $1
RETURNING id, branch_id, full_name, phone, email, staff_id, pin, role, status, created_at, updated_at;
