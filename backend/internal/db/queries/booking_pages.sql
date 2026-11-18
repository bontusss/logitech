-- name: GetBookingPageByCompany :one
SELECT * FROM booking_pages WHERE company_id = $1;

-- name: BookingPageSlugExists :one
SELECT EXISTS (SELECT 1 FROM booking_pages WHERE slug = $1) AS taken;

-- name: UpsertBookingPage :one
INSERT INTO booking_pages (company_id, slug, headline, description, is_published)
VALUES (
  sqlc.arg(company_id),
  sqlc.arg(slug),
  NULLIF(sqlc.arg(headline)::text, ''),
  NULLIF(sqlc.arg(description)::text, ''),
  sqlc.arg(is_published)
)
ON CONFLICT (company_id) DO UPDATE
SET slug = EXCLUDED.slug,
    headline = EXCLUDED.headline,
    description = EXCLUDED.description,
    is_published = EXCLUDED.is_published,
    updated_at = now()
RETURNING *;

-- name: GetPublicBookingPage :one
SELECT
  bp.slug,
  bp.headline,
  bp.description,
  c.id AS company_id,
  c.name,
  c.slug AS company_slug,
  c.business_type,
  c.business_email,
  c.phone,
  c.city,
  c.state,
  c.country,
  c.operating_hours,
  s.subdomain,
  COALESCE(b.public_name, c.name) AS public_name,
  b.logo_url,
  b.primary_color,
  b.secondary_color,
  b.cover_image_url
FROM booking_pages bp
JOIN companies c ON c.id = bp.company_id
  AND c.deleted_at IS NULL
  AND c.status = 'verified'
LEFT JOIN company_branding b ON b.company_id = c.id
LEFT JOIN company_subdomains s ON s.company_id = c.id
WHERE bp.slug = $1 AND bp.is_published;
