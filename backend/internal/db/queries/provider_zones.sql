-- name: ListGeoStates :many
SELECT DISTINCT state
FROM geo_zones
WHERE country_code = 'NG' AND status = 'active'
ORDER BY state;

-- name: ListGeoLgas :many
SELECT DISTINCT lga
FROM geo_zones
WHERE country_code = 'NG' AND state = sqlc.arg(state) AND status = 'active'
ORDER BY lga;

-- name: ListGeoZones :many
SELECT *
FROM geo_zones
WHERE country_code = 'NG'
  AND status = 'active'
  AND (sqlc.narg(state)::text IS NULL OR state = sqlc.narg(state))
  AND (sqlc.narg(lga)::text IS NULL OR lga = sqlc.narg(lga))
ORDER BY state, lga, name;

-- name: GetGeoZone :one
SELECT * FROM geo_zones WHERE id = sqlc.arg(id) AND status = 'active';

-- name: ListReachableDestinationZones :many
WITH RECURSIVE origin_ancestors AS (
  SELECT gz.id FROM geo_zones gz WHERE gz.id = sqlc.arg(origin_zone_id)
  UNION ALL
  SELECT g.parent_id FROM geo_zones g
  JOIN origin_ancestors a ON g.id = a.id
  WHERE g.parent_id IS NOT NULL
)
SELECT DISTINCT g.id, g.state, g.lga, g.name
FROM pricing_zone_rates r
JOIN companies c ON c.id = r.company_id AND c.deleted_at IS NULL AND c.status = 'verified'
JOIN service_zones fz ON fz.id = r.from_zone_id AND fz.company_id = c.id
JOIN service_zones tz ON tz.id = r.to_zone_id AND tz.company_id = c.id
JOIN geo_zones g ON g.id = tz.geo_zone_id
LEFT JOIN marketplace_settings ms ON ms.company_id = c.id
WHERE r.status = 'active'
  AND (ms.company_id IS NULL OR (ms.is_listed AND ms.visibility_status = 'listed'))
  AND fz.geo_zone_id IN (SELECT id FROM origin_ancestors)
ORDER BY g.state, g.lga;

-- name: ListCompanyServiceZones :many
SELECT
  z.id, z.company_id, z.geo_zone_id, z.name, z.parent_zone_id,
  g.state AS geo_state, g.lga AS geo_lga, g.name AS geo_name,
  z.created_at, z.updated_at
FROM service_zones z
JOIN geo_zones g ON g.id = z.geo_zone_id
WHERE z.company_id = sqlc.arg(company_id)
ORDER BY g.state, g.lga, g.name;

-- name: GetCompanyServiceZone :one
SELECT
  z.id, z.company_id, z.geo_zone_id, z.name, z.parent_zone_id,
  g.state AS geo_state, g.lga AS geo_lga, g.name AS geo_name,
  z.created_at, z.updated_at
FROM service_zones z
JOIN geo_zones g ON g.id = z.geo_zone_id
WHERE z.company_id = sqlc.arg(company_id) AND z.id = sqlc.arg(id);

-- name: CreateCompanyServiceZone :one
INSERT INTO service_zones (company_id, area_id, geo_zone_id, name, state, lga, parent_zone_id)
SELECT sqlc.arg(company_id), NULL, g.id, g.name, g.state, g.lga, sqlc.narg(parent_zone_id)
FROM geo_zones g
WHERE g.id = sqlc.arg(geo_zone_id) AND g.status = 'active'
RETURNING *;

-- name: UpdateCompanyServiceZone :one
UPDATE service_zones
SET parent_zone_id = sqlc.narg(parent_zone_id), updated_at = now()
WHERE company_id = sqlc.arg(company_id) AND id = sqlc.arg(id)
RETURNING *;

-- name: DeleteCompanyServiceZone :exec
DELETE FROM service_zones WHERE company_id = sqlc.arg(company_id) AND id = sqlc.arg(id);

-- name: ListCompanyZoneRoutes :many
SELECT
  r.id, r.company_id, r.from_zone_id, r.to_zone_id, r.price_minor,
  r.included_weight_kg, r.additional_weight_fee_minor, r.status,
  fz.name AS from_zone_name, fg.state AS from_state, fg.lga AS from_lga,
  tz.name AS to_zone_name, tg.state AS to_state, tg.lga AS to_lga,
  r.created_at, r.updated_at
FROM pricing_zone_rates r
JOIN service_zones fz ON fz.id = r.from_zone_id
JOIN service_zones tz ON tz.id = r.to_zone_id
JOIN geo_zones fg ON fg.id = fz.geo_zone_id
JOIN geo_zones tg ON tg.id = tz.geo_zone_id
WHERE r.company_id = sqlc.arg(company_id)
ORDER BY fg.state, fg.lga, tg.state, tg.lga;

-- name: GetCompanyZoneRoute :one
SELECT * FROM pricing_zone_rates
WHERE company_id = sqlc.arg(company_id) AND id = sqlc.arg(id);

-- name: CreateCompanyZoneRoute :one
INSERT INTO pricing_zone_rates (
  company_id, area_id, from_zone_id, to_zone_id, price_minor,
  included_weight_kg, additional_weight_fee_minor, status
)
SELECT sqlc.arg(company_id), NULL, sqlc.arg(from_zone_id), sqlc.arg(to_zone_id),
  sqlc.arg(price_minor), sqlc.arg(included_weight_kg), sqlc.arg(additional_weight_fee_minor), sqlc.arg(status)
WHERE EXISTS (
  SELECT 1 FROM service_zones WHERE id = sqlc.arg(from_zone_id) AND company_id = sqlc.arg(company_id)
)
AND EXISTS (
  SELECT 1 FROM service_zones WHERE id = sqlc.arg(to_zone_id) AND company_id = sqlc.arg(company_id)
)
RETURNING *;

-- name: UpdateCompanyZoneRoute :one
UPDATE pricing_zone_rates r
SET from_zone_id = sqlc.arg(from_zone_id),
    to_zone_id = sqlc.arg(to_zone_id),
    price_minor = sqlc.arg(price_minor),
    included_weight_kg = sqlc.arg(included_weight_kg),
    additional_weight_fee_minor = sqlc.arg(additional_weight_fee_minor),
    status = sqlc.arg(status),
    updated_at = now()
WHERE r.company_id = sqlc.arg(company_id) AND r.id = sqlc.arg(id)
  AND EXISTS (
    SELECT 1 FROM service_zones sz WHERE sz.id = sqlc.arg(from_zone_id) AND sz.company_id = sqlc.arg(company_id)
  )
  AND EXISTS (
    SELECT 1 FROM service_zones sz WHERE sz.id = sqlc.arg(to_zone_id) AND sz.company_id = sqlc.arg(company_id)
  )
RETURNING *;

-- name: SetCompanyZoneRouteStatus :one
UPDATE pricing_zone_rates
SET status = sqlc.arg(status), updated_at = now()
WHERE company_id = sqlc.arg(company_id) AND id = sqlc.arg(id)
RETURNING *;

-- name: DeleteCompanyZoneRoute :exec
DELETE FROM pricing_zone_rates WHERE company_id = sqlc.arg(company_id) AND id = sqlc.arg(id);

-- name: ListPublicBookingZoneRoutes :many
SELECT
  r.id, r.price_minor, r.included_weight_kg, r.additional_weight_fee_minor,
  fz.id AS from_zone_id, fz.name AS from_zone_name, fg.state AS from_state, fg.lga AS from_lga,
  tz.id AS to_zone_id, tz.name AS to_zone_name, tg.state AS to_state, tg.lga AS to_lga
FROM pricing_zone_rates r
JOIN service_zones fz ON fz.id = r.from_zone_id
JOIN service_zones tz ON tz.id = r.to_zone_id
JOIN geo_zones fg ON fg.id = fz.geo_zone_id
JOIN geo_zones tg ON tg.id = tz.geo_zone_id
WHERE r.company_id = sqlc.arg(company_id) AND r.status = 'active'
ORDER BY fg.state, fg.lga, tg.state, tg.lga;

-- name: GetPublicBookingZoneRoute :one
SELECT
  r.id, r.price_minor, r.included_weight_kg, r.additional_weight_fee_minor,
  fz.id AS from_zone_id, fz.name AS from_zone_name, fg.state AS from_state, fg.lga AS from_lga,
  tz.id AS to_zone_id, tz.name AS to_zone_name, tg.state AS to_state, tg.lga AS to_lga
FROM pricing_zone_rates r
JOIN service_zones fz ON fz.id = r.from_zone_id
JOIN service_zones tz ON tz.id = r.to_zone_id
JOIN geo_zones fg ON fg.id = fz.geo_zone_id
JOIN geo_zones tg ON tg.id = tz.geo_zone_id
WHERE r.company_id = sqlc.arg(company_id) AND r.id = sqlc.arg(id) AND r.status = 'active';

-- name: ListMarketplaceZoneRoutesForLane :many
WITH RECURSIVE origin_ancestors AS (
  SELECT gz.id FROM geo_zones gz WHERE gz.id = sqlc.arg(origin_zone_id)
  UNION ALL
  SELECT g.parent_id FROM geo_zones g
  JOIN origin_ancestors a ON g.id = a.id
  WHERE g.parent_id IS NOT NULL
), destination_ancestors AS (
  SELECT gz.id FROM geo_zones gz WHERE gz.id = sqlc.arg(destination_zone_id)
  UNION ALL
  SELECT g.parent_id FROM geo_zones g
  JOIN destination_ancestors a ON g.id = a.id
  WHERE g.parent_id IS NOT NULL
)
SELECT
  c.id AS company_id,
  c.name AS company_name,
  c.slug AS company_slug,
  COALESCE(b.logo_url, '') AS logo_url,
  COALESCE(ms.public_description, COALESCE(c.description, '')) AS public_description,
  COALESCE(bp.slug, '') AS booking_slug,
  r.id AS route_id,
  r.price_minor,
  r.included_weight_kg,
  r.additional_weight_fee_minor,
  fz.id AS from_zone_id,
  fz.name AS from_zone_name,
  fg.state AS from_state,
  fg.lga AS from_lga,
  tz.id AS to_zone_id,
  tz.name AS to_zone_name,
  tg.state AS to_state,
  tg.lga AS to_lga,
  c.accepted_item_categories,
  c.prohibited_item_categories,
  COALESCE((
    SELECT ROUND(AVG(cr.rating)::numeric, 1)
    FROM company_reviews cr
    WHERE cr.company_id = c.id
  ), 0)::float8 AS provider_rating
FROM pricing_zone_rates r
JOIN companies c ON c.id = r.company_id
  AND c.deleted_at IS NULL
  AND c.status = 'verified'
JOIN service_zones fz ON fz.id = r.from_zone_id AND fz.company_id = c.id
JOIN service_zones tz ON tz.id = r.to_zone_id AND tz.company_id = c.id
JOIN geo_zones fg ON fg.id = fz.geo_zone_id
JOIN geo_zones tg ON tg.id = tz.geo_zone_id
LEFT JOIN marketplace_settings ms ON ms.company_id = c.id
LEFT JOIN company_branding b ON b.company_id = c.id
LEFT JOIN booking_pages bp ON bp.company_id = c.id AND bp.is_published
WHERE r.status = 'active'
  AND (ms.company_id IS NULL OR (ms.is_listed AND ms.visibility_status = 'listed'))
  AND fz.geo_zone_id IN (SELECT id FROM origin_ancestors)
  AND tz.geo_zone_id IN (SELECT id FROM destination_ancestors)
ORDER BY c.name;
