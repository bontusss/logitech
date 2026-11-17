-- name: GetSubdomainByCompany :one
SELECT * FROM company_subdomains WHERE company_id = $1;

-- name: GetCompanyIDBySubdomain :one
SELECT company_id FROM company_subdomains WHERE subdomain = $1;

-- name: SubdomainExists :one
SELECT EXISTS (SELECT 1 FROM company_subdomains WHERE subdomain = $1) AS taken;

-- name: UpsertSubdomain :one
INSERT INTO company_subdomains (company_id, subdomain)
VALUES ($1, $2)
ON CONFLICT (company_id) DO UPDATE
SET subdomain = EXCLUDED.subdomain, updated_at = now()
RETURNING *;

-- name: GetPublicCompanyBySubdomain :one
SELECT
  c.id,
  c.name,
  c.slug,
  c.business_type,
  c.description,
  c.city,
  c.state,
  c.country,
  c.status,
  c.verified_at,
  s.subdomain,
  COALESCE(b.public_name, c.name) AS public_name,
  b.logo_url,
  b.primary_color,
  b.secondary_color,
  b.favicon_url,
  b.cover_image_url,
  b.description AS branding_description,
  b.pwa_short_name
FROM company_subdomains s
JOIN companies c ON c.id = s.company_id AND c.deleted_at IS NULL
LEFT JOIN company_branding b ON b.company_id = c.id
WHERE s.subdomain = $1;
