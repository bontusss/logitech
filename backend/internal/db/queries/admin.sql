-- name: ListCustomerAccounts :many
SELECT
  id, email, full_name, phone, status, email_verified_at, created_at, last_login_at
FROM users
WHERE account_type = 'customer' AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: ListProviderCompanies :many
SELECT
  c.id,
  c.name,
  c.business_email,
  c.phone,
  c.business_type,
  c.cac_business_name,
  c.state,
  c.city,
  c.status,
  c.onboarding_status,
  c.verified_at,
  c.created_at,
  u.id AS owner_id,
  u.full_name AS owner_name,
  u.email AS owner_email,
  u.email_verified_at AS owner_email_verified_at
FROM companies c
JOIN users u ON u.id = c.owner_user_id
WHERE c.deleted_at IS NULL
ORDER BY
  CASE c.status
    WHEN 'under_review' THEN 0
    WHEN 'pending' THEN 1
    WHEN 'rejected' THEN 2
    ELSE 3
  END,
  c.created_at DESC;

-- name: ApproveCompany :one
UPDATE companies
SET status = 'verified',
    onboarding_status = 'complete',
    verified_at = COALESCE(verified_at, now()),
    updated_at = now()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: ListPlatformAdmins :many
SELECT id, email, full_name, platform_role, status, created_at, last_login_at
FROM users
WHERE account_type = 'platform' AND deleted_at IS NULL
ORDER BY created_at;
