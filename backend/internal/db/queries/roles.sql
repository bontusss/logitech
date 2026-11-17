-- name: ListRolesForCompany :many
SELECT
  r.id,
  r.company_id,
  r.key,
  r.name,
  r.description,
  r.is_system,
  COALESCE(array_agg(p.key ORDER BY p.key) FILTER (WHERE p.id IS NOT NULL), ARRAY[]::text[])::text[] AS permissions
FROM roles r
LEFT JOIN role_permissions rp ON rp.role_id = r.id
LEFT JOIN permissions p ON p.id = rp.permission_id
WHERE r.company_id IS NULL OR r.company_id = $1
GROUP BY r.id
ORDER BY r.is_system DESC, r.name;

-- name: ListPermissions :many
SELECT key, description FROM permissions ORDER BY key;

-- name: GetCompanyRole :one
SELECT * FROM roles
WHERE id = $1 AND (company_id IS NULL OR company_id = $2);

-- name: CreateCompanyRole :one
INSERT INTO roles (company_id, key, name, description, is_system)
VALUES ($1, $2, $3, NULLIF($4::text, ''), false)
RETURNING *;

-- name: UpdateCompanyRole :one
UPDATE roles
SET name = $3, description = NULLIF($4::text, ''), updated_at = now()
WHERE id = $1 AND company_id = $2 AND is_system = false
RETURNING *;

-- name: DeleteCompanyRole :exec
DELETE FROM roles
WHERE id = $1 AND company_id = $2 AND is_system = false;

-- name: DeleteRolePermissions :exec
DELETE FROM role_permissions WHERE role_id = $1;

-- name: AddRolePermission :exec
INSERT INTO role_permissions (role_id, permission_id)
SELECT $1, p.id FROM permissions p WHERE p.key = $2
ON CONFLICT DO NOTHING;

-- name: ListRolePermissionKeys :many
SELECT p.key
FROM role_permissions rp
JOIN permissions p ON p.id = rp.permission_id
WHERE rp.role_id = $1
ORDER BY p.key;

-- name: CountRoleAssignments :one
SELECT count(*) FROM member_roles WHERE role_id = $1;

-- name: GetRoleByCompanyAndKey :one
SELECT * FROM roles
WHERE key = $1 AND (company_id IS NULL OR company_id = $2)
ORDER BY company_id NULLS LAST
LIMIT 1;
