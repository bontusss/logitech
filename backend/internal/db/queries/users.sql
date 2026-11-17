-- name: CreateUser :one
INSERT INTO users (
  email, password_hash, full_name, phone, account_type, platform_role, email_verified_at
) VALUES (
  $1, $2, $3, NULLIF($4::text, ''), $5, $6, $7
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 AND deleted_at IS NULL;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateUserProfile :one
UPDATE users
SET full_name = $2, phone = NULLIF($3::text, ''), updated_at = now()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2, updated_at = now()
WHERE id = $1;

-- name: MarkEmailVerified :exec
UPDATE users
SET email_verified_at = COALESCE(email_verified_at, now()), updated_at = now()
WHERE id = $1;

-- name: TouchLastLogin :exec
UPDATE users
SET last_login_at = now(), updated_at = now()
WHERE id = $1;
