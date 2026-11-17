-- name: CreateBranch :one
INSERT INTO branches (
  company_id, name, code, address, city, state, country, phone, email,
  manager_member_id, operating_hours
) VALUES (
  $1, $2, $3, $4, $5, $6, $7,
  NULLIF($8::text, ''), NULLIF($9::citext, ''), $10, $11
)
RETURNING *;

-- name: GetBranch :one
SELECT * FROM branches
WHERE id = $1 AND company_id = $2 AND deleted_at IS NULL;

-- name: ListBranches :many
SELECT * FROM branches
WHERE company_id = $1 AND deleted_at IS NULL
  AND ($2::text = '' OR name ILIKE '%' || $2 || '%' OR code::text ILIKE '%' || $2 || '%' OR city ILIKE '%' || $2 || '%')
  AND ($3::text = '' OR status = $3)
ORDER BY name;

-- name: ListBranchesForMember :many
SELECT b.* FROM branches b
JOIN branch_members bm ON bm.branch_id = b.id
WHERE b.company_id = $1 AND bm.member_id = $2 AND b.deleted_at IS NULL
  AND ($3::text = '' OR b.name ILIKE '%' || $3 || '%' OR b.code::text ILIKE '%' || $3 || '%' OR b.city ILIKE '%' || $3 || '%')
  AND ($4::text = '' OR b.status = $4)
ORDER BY b.name;

-- name: UpdateBranch :one
UPDATE branches
SET name = $3,
    code = $4,
    address = $5,
    city = $6,
    state = $7,
    country = $8,
    phone = NULLIF($9::text, ''),
    email = NULLIF($10::text, ''),
    manager_member_id = $11,
    operating_hours = $12,
    updated_at = now()
WHERE id = $1 AND company_id = $2 AND deleted_at IS NULL
RETURNING *;

-- name: SetBranchStatus :one
UPDATE branches
SET status = $3, updated_at = now()
WHERE id = $1 AND company_id = $2 AND deleted_at IS NULL
RETURNING *;

-- name: ListBranchStaff :many
SELECT
  m.id,
  m.user_id,
  u.full_name,
  u.email,
  m.title,
  m.status,
  COALESCE(b.manager_member_id = m.id, false) AS is_manager,
  COALESCE((
    SELECT array_agg(r.key ORDER BY r.key)
    FROM member_roles mr
    JOIN roles r ON r.id = mr.role_id
    WHERE mr.member_id = m.id
  ), ARRAY[]::text[])::text[] AS roles
FROM branch_members bm
JOIN branches b ON b.id = bm.branch_id
JOIN company_members m ON m.id = bm.member_id AND m.deleted_at IS NULL
JOIN users u ON u.id = m.user_id
WHERE bm.branch_id = $1 AND b.company_id = $2 AND b.deleted_at IS NULL
ORDER BY (b.manager_member_id = m.id) DESC, u.full_name;

-- name: AddBranchMember :exec
INSERT INTO branch_members (branch_id, member_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: DeleteBranchMembers :exec
DELETE FROM branch_members WHERE branch_id = $1;

-- name: DeleteMemberBranches :exec
DELETE FROM branch_members WHERE member_id = $1;

-- name: AddMemberBranch :exec
INSERT INTO branch_members (branch_id, member_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: ListMemberBranches :many
SELECT b.id, b.name, b.code, b.status
FROM branch_members bm
JOIN branches b ON b.id = bm.branch_id
WHERE bm.member_id = $1 AND b.company_id = $2 AND b.deleted_at IS NULL
ORDER BY b.name;

-- name: ListMemberBranchIDs :many
SELECT b.id
FROM branch_members bm
JOIN branches b ON b.id = bm.branch_id
WHERE bm.member_id = $1 AND b.company_id = $2 AND b.deleted_at IS NULL
ORDER BY b.name;

-- name: MemberHasBranchAccess :one
SELECT EXISTS (
  SELECT 1
  FROM branch_members bm
  JOIN branches b ON b.id = bm.branch_id
  WHERE bm.branch_id = $1 AND bm.member_id = $2
    AND b.company_id = $3 AND b.deleted_at IS NULL
) AS allowed;

-- name: CountCompanyBranches :one
SELECT count(*) FROM branches
WHERE company_id = $1 AND deleted_at IS NULL;
