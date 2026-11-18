-- name: CreateCompany :one
INSERT INTO companies (
  name, slug, business_email, phone, business_type, country, state, city, address, owner_user_id
) VALUES (
  $1, $2, $3, $4, $5, $6, NULLIF($7::text, ''), NULLIF($8::text, ''), NULLIF($9::text, ''), $10
)
RETURNING *;

-- name: GetCompanyByID :one
SELECT * FROM companies
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateCompany :one
UPDATE companies
SET name = $2,
    business_email = $3,
    phone = $4,
    business_type = $5,
    country = $6,
    state = NULLIF($7::text, ''),
    city = NULLIF($8::text, ''),
    address = NULLIF($9::text, ''),
    description = NULLIF($10::text, ''),
    cover_image_url = NULLIF($11::text, ''),
    operating_hours = $12,
    updated_at = now()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateCompanyCACName :one
UPDATE companies
SET cac_business_name = NULLIF(sqlc.arg(cac_business_name)::text, ''),
    updated_at = now()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateCompanyVerification :one
UPDATE companies
SET status = $2,
    onboarding_status = $3,
    verified_at = $4,
    updated_at = now()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: CreateCompanyMember :one
INSERT INTO company_members (company_id, user_id, status, title, invited_by, joined_at)
VALUES ($1, $2, $3, NULLIF($4::text, ''), $5, $6)
RETURNING *;

-- name: GetCompanyMember :one
SELECT * FROM company_members
WHERE company_id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: GetCompanyMemberByID :one
SELECT * FROM company_members
WHERE company_id = $1 AND id = $2 AND deleted_at IS NULL;

-- name: GetActiveCompanyMember :one
SELECT * FROM company_members
WHERE company_id = $1 AND user_id = $2
  AND deleted_at IS NULL AND status = 'active';

-- name: ListCompanyMembers :many
SELECT
  m.id,
  m.user_id,
  u.full_name,
  u.email,
  m.status,
  m.title,
  m.joined_at,
  COALESCE((
    SELECT array_agg(r.key ORDER BY r.key)
    FROM member_roles mr
    JOIN roles r ON r.id = mr.role_id
    WHERE mr.member_id = m.id
  ), ARRAY[]::text[])::text[] AS roles,
  COALESCE((
    SELECT array_agg(b.id::text ORDER BY b.name)
    FROM branch_members bm
    JOIN branches b ON b.id = bm.branch_id AND b.deleted_at IS NULL
    WHERE bm.member_id = m.id
  ), ARRAY[]::text[])::text[] AS branch_ids
FROM company_members m
JOIN users u ON u.id = m.user_id
WHERE m.company_id = $1 AND m.deleted_at IS NULL
ORDER BY m.created_at;

-- name: ListMemberRoles :many
SELECT r.key
FROM member_roles mr
JOIN roles r ON r.id = mr.role_id
WHERE mr.member_id = $1
ORDER BY r.key;

-- name: DeleteMemberRoles :exec
DELETE FROM member_roles WHERE member_id = $1;

-- name: AddMemberRole :exec
INSERT INTO member_roles (member_id, role_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: GetSystemRoleByKey :one
SELECT * FROM roles
WHERE company_id IS NULL AND key = $1;

-- name: CountCompanyOwners :one
SELECT count(*)
FROM company_members m
JOIN member_roles mr ON mr.member_id = m.id
JOIN roles r ON r.id = mr.role_id
WHERE m.company_id = $1
  AND m.deleted_at IS NULL
  AND m.status = 'active'
  AND r.key = 'owner';

-- name: SoftDeleteMember :exec
UPDATE company_members
SET status = 'removed', deleted_at = now(), updated_at = now()
WHERE id = $1 AND company_id = $2 AND deleted_at IS NULL;

-- name: UpdateMemberStatus :exec
UPDATE company_members
SET status = $2, updated_at = now()
WHERE id = $1 AND company_id = $3 AND deleted_at IS NULL;

-- name: ListUserMemberships :many
SELECT
  c.id AS company_id,
  c.name,
  c.slug,
  c.status,
  c.onboarding_status,
  m.id AS member_id,
  m.status AS member_status,
  COALESCE(array_agg(r.key) FILTER (WHERE r.id IS NOT NULL), ARRAY[]::text[])::text[] AS roles
FROM company_members m
JOIN companies c ON c.id = m.company_id
LEFT JOIN member_roles mr ON mr.member_id = m.id
LEFT JOIN roles r ON r.id = mr.role_id
WHERE m.user_id = $1
  AND m.deleted_at IS NULL
  AND m.status = 'active'
  AND c.deleted_at IS NULL
GROUP BY c.id, c.name, c.slug, c.status, c.onboarding_status, m.id, m.status
ORDER BY c.name;

-- name: CountActiveMemberships :one
SELECT count(*)
FROM company_members m
JOIN companies c ON c.id = m.company_id
WHERE m.user_id = $1 AND m.deleted_at IS NULL AND m.status = 'active' AND c.deleted_at IS NULL;

-- name: UserHasPermission :one
SELECT EXISTS (
  SELECT 1
  FROM company_members m
  JOIN member_roles mr ON mr.member_id = m.id
  JOIN role_permissions rp ON rp.role_id = mr.role_id
  JOIN permissions p ON p.id = rp.permission_id
  WHERE m.user_id = $1
    AND m.company_id = $2
    AND m.deleted_at IS NULL
    AND m.status = 'active'
    AND p.key = $3
) AS allowed;

-- name: ListUserPermissions :many
SELECT DISTINCT p.key
FROM company_members m
JOIN member_roles mr ON mr.member_id = m.id
JOIN role_permissions rp ON rp.role_id = mr.role_id
JOIN permissions p ON p.id = rp.permission_id
WHERE m.user_id = $1
  AND m.company_id = $2
  AND m.deleted_at IS NULL
  AND m.status = 'active'
ORDER BY p.key;

-- name: GetCompanySettings :one
SELECT * FROM company_settings WHERE company_id = $1;

-- name: UpdateCompanySettings :one
UPDATE company_settings
SET timezone = $2,
    currency = $3,
    weight_unit = $4,
    auto_accept_bookings = $5,
    notify_email = $6,
    updated_at = now()
WHERE company_id = $1
RETURNING *;

-- name: GetCompanyBranding :one
SELECT * FROM company_branding WHERE company_id = $1;

-- name: UpdateCompanyBranding :one
UPDATE company_branding
SET logo_url = NULLIF($2::text, ''),
    primary_color = $3,
    secondary_color = $4,
    favicon_url = NULLIF($5::text, ''),
    cover_image_url = NULLIF($6::text, ''),
    public_name = NULLIF($7::text, ''),
    description = NULLIF($8::text, ''),
    pwa_short_name = NULLIF($9::text, ''),
    updated_at = now()
WHERE company_id = $1
RETURNING *;

-- name: CreateInvitation :one
INSERT INTO company_invitations (
  company_id, email, role_id, token_hash, invited_by, expires_at
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetInvitationByToken :one
SELECT * FROM company_invitations
WHERE token_hash = $1 AND accepted_at IS NULL AND revoked_at IS NULL AND expires_at > now();

-- name: RevokePendingInvitations :exec
UPDATE company_invitations
SET revoked_at = now()
WHERE company_id = $1 AND email = $2 AND accepted_at IS NULL AND revoked_at IS NULL;

-- name: CreateCompanySettings :one
INSERT INTO company_settings (company_id)
VALUES ($1)
RETURNING *;

-- name: CreateCompanyBranding :one
INSERT INTO company_branding (company_id)
VALUES ($1)
RETURNING *;
