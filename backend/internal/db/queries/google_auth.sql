-- name: GetUserByGoogleSubject :one
SELECT * FROM users
WHERE google_subject = $1 AND deleted_at IS NULL;

-- name: LinkGoogleSubject :one
UPDATE users
SET google_subject = $1,
    email_verified_at = COALESCE(email_verified_at, now()),
    updated_at = now()
WHERE id = $2
  AND deleted_at IS NULL
  AND (google_subject IS NULL OR google_subject = $1)
RETURNING *;

-- name: CreateGoogleUser :one
INSERT INTO users (
  email, password_hash, full_name, account_type, google_subject, email_verified_at
) VALUES (
  $1, $2, $3, $4, $5, now()
)
RETURNING *;
