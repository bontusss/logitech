-- name: GetStaffByID :one
SELECT * FROM staff
WHERE id = $1 LIMIT 1;

-- name: GetStaffByStaffID :one
SELECT * FROM staff
WHERE staff_id ILIKE $1 LIMIT 1;

-- name: GetStaffByEmail :one
SELECT * FROM staff
WHERE email = $1 LIMIT 1;

-- name: CheckStaffIDExists :one
SELECT EXISTS(SELECT 1 FROM staff WHERE staff_id = $1);

-- name: GetNextStaffIDSequence :one
SELECT nextval('staff_id_seq')::bigint AS next_val;

-- name: CreateStaff :one
INSERT INTO staff (
    id, branch_id, full_name, phone, email, staff_id, otp_hash,
    role, job_role, status, created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING *;

-- name: ListStaff :many
SELECT * FROM staff
ORDER BY created_at DESC;

-- name: ListStaffByBranch :many
SELECT * FROM staff
WHERE branch_id = $1
ORDER BY created_at DESC;

-- name: ListManagers :many
SELECT * FROM staff
WHERE role = 'manager'
ORDER BY created_at DESC;

-- name: ListManagersByBranch :many
SELECT * FROM staff
WHERE role = 'manager' AND branch_id = $1
ORDER BY created_at DESC;

-- name: UpdateStaffDetails :one
UPDATE staff
SET full_name = COALESCE(sqlc.narg('full_name'), full_name),
    phone     = COALESCE(sqlc.narg('phone'), phone),
    email     = COALESCE(sqlc.narg('email'), email),
    job_role  = COALESCE(sqlc.narg('job_role'), job_role),
    updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: UpdateStaffStatus :one
UPDATE staff
SET status = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: RecordFailedLogin :one
UPDATE staff
SET failed_login_attempts = failed_login_attempts + 1,
    locked_at = CASE WHEN failed_login_attempts + 1 >= $2 THEN now() ELSE locked_at END
WHERE id = $1
RETURNING failed_login_attempts, locked_at;

-- name: ResetLoginState :one
UPDATE staff
SET failed_login_attempts = 0
WHERE id = $1
RETURNING failed_login_attempts;

-- name: UnlockStaffAccount :one
UPDATE staff
SET locked_at = NULL,
    failed_login_attempts = 0,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: CompleteFirstLogin :one
UPDATE staff
SET pin_hash = $2,
    otp_hash = NULL,
    first_login_done = true,
    first_login_at = now(),
    failed_login_attempts = 0,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: ChangeOwnPin :one
UPDATE staff
SET pin_hash = $2,
    otp_hash = NULL,
    first_login_done = true,
    first_login_at = COALESCE(first_login_at, now()),
    failed_login_attempts = 0,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: StartPinReset :one
UPDATE staff
SET otp_hash = $2,
    failed_login_attempts = 0,
    locked_at = NULL,
    updated_at = now()
WHERE id = $1
RETURNING *;
