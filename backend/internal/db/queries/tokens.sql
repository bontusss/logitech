-- name: CreateEmailVerification :one
INSERT INTO email_verifications (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetEmailVerificationByHash :one
SELECT * FROM email_verifications
WHERE token_hash = $1 AND consumed_at IS NULL AND expires_at > now();

-- name: ConsumeEmailVerification :exec
UPDATE email_verifications
SET consumed_at = now()
WHERE id = $1 AND consumed_at IS NULL;

-- name: CreatePasswordReset :one
INSERT INTO password_resets (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetPasswordResetByHash :one
SELECT * FROM password_resets
WHERE token_hash = $1 AND consumed_at IS NULL AND expires_at > now();

-- name: ConsumePasswordReset :exec
UPDATE password_resets
SET consumed_at = now()
WHERE id = $1 AND consumed_at IS NULL;

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
  user_id, token_hash, family_id, user_agent, ip, expires_at
) VALUES (
  $1, $2, $3, NULLIF($4::text, ''), NULLIF($5::text, ''), $6
)
RETURNING *;

-- name: GetRefreshTokenByHash :one
SELECT * FROM refresh_tokens WHERE token_hash = $1;

-- name: RotateRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = now(), replaced_by = $2
WHERE id = $1 AND revoked_at IS NULL;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = now()
WHERE id = $1 AND revoked_at IS NULL;

-- name: RevokeRefreshTokenFamily :exec
UPDATE refresh_tokens
SET revoked_at = now()
WHERE family_id = $1 AND revoked_at IS NULL;

-- name: RevokeUserRefreshTokens :exec
UPDATE refresh_tokens
SET revoked_at = now()
WHERE user_id = $1 AND revoked_at IS NULL;

-- name: CreateAuditLog :one
INSERT INTO audit_logs (
  actor_user_id, company_id, action, entity_type, entity_id, metadata, ip, user_agent
) VALUES (
  $1, $2, $3, $4, $5, $6, NULLIF($7::text, ''), NULLIF($8::text, '')
)
RETURNING *;
