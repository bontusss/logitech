-- name: UpdateMemberProfile :one
UPDATE company_members
SET title = NULLIF($3::text, ''), status = $4, updated_at = now()
WHERE id = $1 AND company_id = $2 AND deleted_at IS NULL
RETURNING *;

-- name: ListCompanyInvitations :many
SELECT
  i.id,
  i.email,
  i.expires_at,
  i.created_at,
  r.key AS role_key,
  COALESCE((
    SELECT array_agg(ib.branch_id::text)
    FROM invitation_branches ib WHERE ib.invitation_id = i.id
  ), ARRAY[]::text[])::text[] AS branch_ids
FROM company_invitations i
JOIN roles r ON r.id = i.role_id
WHERE i.company_id = $1
  AND i.accepted_at IS NULL
  AND i.revoked_at IS NULL
  AND i.expires_at > now()
ORDER BY i.created_at DESC;

-- name: GetCompanyInvitation :one
SELECT * FROM company_invitations
WHERE id = $1 AND company_id = $2;

-- name: RevokeInvitation :exec
UPDATE company_invitations
SET revoked_at = now()
WHERE id = $1 AND company_id = $2 AND accepted_at IS NULL AND revoked_at IS NULL;

-- name: UpdateInvitationToken :one
UPDATE company_invitations
SET token_hash = $3, expires_at = $4
WHERE id = $1 AND company_id = $2 AND accepted_at IS NULL AND revoked_at IS NULL
RETURNING *;

-- name: MarkInvitationAccepted :exec
UPDATE company_invitations
SET accepted_at = now(), accepted_by_user_id = $2
WHERE id = $1 AND accepted_at IS NULL AND revoked_at IS NULL;

-- name: AddInvitationBranch :exec
INSERT INTO invitation_branches (invitation_id, branch_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: ListInvitationBranchIDs :many
SELECT branch_id FROM invitation_branches WHERE invitation_id = $1;

-- name: GetInvitationDetails :one
SELECT
  i.id,
  i.company_id,
  i.email,
  i.token_hash,
  i.expires_at,
  i.accepted_at,
  i.revoked_at,
  c.name AS company_name,
  c.slug AS company_slug,
  r.key AS role_key,
  r.name AS role_name
FROM company_invitations i
JOIN companies c ON c.id = i.company_id AND c.deleted_at IS NULL
JOIN roles r ON r.id = i.role_id
WHERE i.token_hash = $1;

-- name: ClaimInvitation :one
UPDATE company_invitations
SET accepted_at = now(), accepted_by_user_id = $2
WHERE id = $1 AND accepted_at IS NULL AND revoked_at IS NULL AND expires_at > now()
RETURNING *;

-- name: GetAnyCompanyMember :one
SELECT * FROM company_members
WHERE company_id = $1 AND user_id = $2;

-- name: ReactivateCompanyMember :one
UPDATE company_members
SET status = 'active', deleted_at = NULL, joined_at = now(), updated_at = now()
WHERE id = $1
RETURNING *;
