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

-- name: GetGeoAreaZoneByName :one
SELECT * FROM geo_zones
WHERE parent_id = sqlc.arg(parent_id)
  AND lower(name) = lower(sqlc.arg(name))
  AND status = 'active';

-- name: CreateGeoAreaZone :one
INSERT INTO geo_zones (country_code, state, lga, name, parent_id)
SELECT g.country_code, g.state, g.lga, sqlc.arg(name), g.id
FROM geo_zones g
WHERE g.id = sqlc.arg(parent_id)
  AND g.status = 'active'
  AND g.parent_id IS NULL
RETURNING *;

-- name: ListReachableDestinationZones :many
WITH RECURSIVE origin_ancestors AS (
  SELECT gz.id FROM geo_zones gz WHERE gz.id = sqlc.arg(origin_zone_id)
  UNION ALL
  SELECT g.parent_id FROM geo_zones g
  JOIN origin_ancestors a ON g.id = a.id
  WHERE g.parent_id IS NOT NULL
), origin_descendants AS (
  SELECT gz.id FROM geo_zones gz WHERE gz.id = sqlc.arg(origin_zone_id)
  UNION ALL
  SELECT g.id FROM geo_zones g
  JOIN origin_descendants d ON g.parent_id = d.id
  WHERE g.status = 'active'
)
SELECT DISTINCT g.id, g.state, g.lga, g.name, g.parent_id
FROM pricing_zone_rates r
JOIN companies c ON c.id = r.company_id AND c.deleted_at IS NULL AND c.status = 'verified'
JOIN service_zones fz ON fz.id = r.from_zone_id AND fz.company_id = c.id
JOIN service_zones tz ON tz.id = r.to_zone_id AND tz.company_id = c.id
JOIN geo_zones g ON g.id = tz.geo_zone_id
LEFT JOIN marketplace_settings ms ON ms.company_id = c.id
WHERE r.status = 'active'
  AND (
    (sqlc.arg(service_mode)::text = 'dropoff' AND r.dropoff_available)
    OR (sqlc.arg(service_mode)::text <> 'dropoff' AND r.pickup_available)
  )
  AND (ms.company_id IS NULL OR (ms.is_listed AND ms.visibility_status = 'listed'))
  AND fz.geo_zone_id IN (
    SELECT id FROM origin_ancestors
    UNION
    SELECT id FROM origin_descendants WHERE sqlc.arg(service_mode)::text = 'dropoff'
  )
ORDER BY g.state, g.lga, g.name;

-- name: ListDropoffDestinationZones :many
SELECT DISTINCT g.id, g.state, g.lga, g.name, g.parent_id
FROM pricing_zone_rates r
JOIN companies c ON c.id = r.company_id AND c.deleted_at IS NULL AND c.status = 'verified'
JOIN service_zones tz ON tz.id = r.to_zone_id AND tz.company_id = c.id
JOIN geo_zones g ON g.id = tz.geo_zone_id
LEFT JOIN marketplace_settings ms ON ms.company_id = c.id
WHERE r.status = 'active'
  AND r.dropoff_available
  AND (ms.company_id IS NULL OR (ms.is_listed AND ms.visibility_status = 'listed'))
ORDER BY g.state, g.lga, g.name;

-- name: ListCompanyActiveBranches :many
SELECT id, name, address, city, state
FROM branches
WHERE company_id = sqlc.arg(company_id)
  AND status = 'active'
  AND deleted_at IS NULL
ORDER BY name;

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
  r.pickup_available, r.dropoff_available,
  r.estimated_pickup_hours, r.estimated_delivery_hours,
  fz.name AS from_zone_name, fg.state AS from_state, fg.lga AS from_lga,
  tz.name AS to_zone_name, tg.state AS to_state, tg.lga AS to_lga,
  r.dropoff_branch_id, b.name AS dropoff_branch_name, b.address AS dropoff_branch_address,
  r.created_at, r.updated_at
FROM pricing_zone_rates r
JOIN service_zones fz ON fz.id = r.from_zone_id
JOIN service_zones tz ON tz.id = r.to_zone_id
JOIN geo_zones fg ON fg.id = fz.geo_zone_id
JOIN geo_zones tg ON tg.id = tz.geo_zone_id
LEFT JOIN branches b ON b.id = r.dropoff_branch_id AND b.company_id = r.company_id AND b.deleted_at IS NULL
WHERE r.company_id = sqlc.arg(company_id)
ORDER BY fg.state, fg.lga, tg.state, tg.lga;

-- name: GetCompanyZoneRoute :one
SELECT * FROM pricing_zone_rates
WHERE company_id = sqlc.arg(company_id) AND id = sqlc.arg(id);

-- name: CreateCompanyZoneRoute :one
INSERT INTO pricing_zone_rates (
  company_id, area_id, from_zone_id, to_zone_id, price_minor,
  included_weight_kg, additional_weight_fee_minor, status,
  pickup_available, dropoff_available,
  estimated_pickup_hours, estimated_delivery_hours, dropoff_branch_id
)
SELECT sqlc.arg(company_id), NULL, sqlc.arg(from_zone_id), sqlc.arg(to_zone_id),
  sqlc.arg(price_minor), sqlc.arg(included_weight_kg), sqlc.arg(additional_weight_fee_minor), sqlc.arg(status),
  sqlc.arg(pickup_available), sqlc.arg(dropoff_available),
  sqlc.narg(estimated_pickup_hours), sqlc.narg(estimated_delivery_hours), sqlc.narg(dropoff_branch_id)
WHERE EXISTS (
  SELECT 1 FROM service_zones WHERE id = sqlc.arg(from_zone_id) AND company_id = sqlc.arg(company_id)
)
AND EXISTS (
  SELECT 1 FROM service_zones WHERE id = sqlc.arg(to_zone_id) AND company_id = sqlc.arg(company_id)
)
AND (
  sqlc.narg(dropoff_branch_id)::uuid IS NULL
  OR EXISTS (
    SELECT 1 FROM branches b
    WHERE b.id = sqlc.narg(dropoff_branch_id) AND b.company_id = sqlc.arg(company_id) AND b.deleted_at IS NULL
  )
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
    pickup_available = sqlc.arg(pickup_available),
    dropoff_available = sqlc.arg(dropoff_available),
    estimated_pickup_hours = sqlc.narg(estimated_pickup_hours),
    estimated_delivery_hours = sqlc.narg(estimated_delivery_hours),
    dropoff_branch_id = sqlc.narg(dropoff_branch_id),
    updated_at = now()
WHERE r.company_id = sqlc.arg(company_id) AND r.id = sqlc.arg(id)
  AND EXISTS (
    SELECT 1 FROM service_zones sz WHERE sz.id = sqlc.arg(from_zone_id) AND sz.company_id = sqlc.arg(company_id)
  )
  AND EXISTS (
    SELECT 1 FROM service_zones sz WHERE sz.id = sqlc.arg(to_zone_id) AND sz.company_id = sqlc.arg(company_id)
  )
  AND (
    sqlc.narg(dropoff_branch_id)::uuid IS NULL
    OR EXISTS (
      SELECT 1 FROM branches b
      WHERE b.id = sqlc.narg(dropoff_branch_id) AND b.company_id = sqlc.arg(company_id) AND b.deleted_at IS NULL
    )
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
  r.pickup_available, r.dropoff_available,
  r.estimated_pickup_hours, r.estimated_delivery_hours,
  fz.id AS from_zone_id, fz.name AS from_zone_name, fz.geo_zone_id AS from_geo_zone_id,
  fg.state AS from_state, fg.lga AS from_lga,
  tz.id AS to_zone_id, tz.name AS to_zone_name, tz.geo_zone_id AS to_geo_zone_id,
  tg.state AS to_state, tg.lga AS to_lga,
  r.dropoff_branch_id, b.name AS dropoff_branch_name, b.address AS dropoff_branch_address
FROM pricing_zone_rates r
JOIN service_zones fz ON fz.id = r.from_zone_id
JOIN service_zones tz ON tz.id = r.to_zone_id
JOIN geo_zones fg ON fg.id = fz.geo_zone_id
JOIN geo_zones tg ON tg.id = tz.geo_zone_id
LEFT JOIN branches b ON b.id = r.dropoff_branch_id AND b.company_id = r.company_id AND b.deleted_at IS NULL
WHERE r.company_id = sqlc.arg(company_id) AND r.status = 'active'
ORDER BY fg.state, fg.lga, tg.state, tg.lga;

-- name: GetPublicBookingZoneRoute :one
SELECT
  r.id, r.price_minor, r.included_weight_kg, r.additional_weight_fee_minor,
  r.pickup_available, r.dropoff_available,
  r.estimated_pickup_hours, r.estimated_delivery_hours,
  fz.id AS from_zone_id, fz.name AS from_zone_name, fz.geo_zone_id AS from_geo_zone_id,
  fg.state AS from_state, fg.lga AS from_lga,
  tz.id AS to_zone_id, tz.name AS to_zone_name, tz.geo_zone_id AS to_geo_zone_id,
  tg.state AS to_state, tg.lga AS to_lga,
  r.dropoff_branch_id, b.name AS dropoff_branch_name, b.address AS dropoff_branch_address
FROM pricing_zone_rates r
JOIN service_zones fz ON fz.id = r.from_zone_id
JOIN service_zones tz ON tz.id = r.to_zone_id
JOIN geo_zones fg ON fg.id = fz.geo_zone_id
JOIN geo_zones tg ON tg.id = tz.geo_zone_id
LEFT JOIN branches b ON b.id = r.dropoff_branch_id AND b.company_id = r.company_id AND b.deleted_at IS NULL
WHERE r.company_id = sqlc.arg(company_id) AND r.id = sqlc.arg(id) AND r.status = 'active';

-- name: ListMarketplaceZoneRoutesForLane :many
WITH RECURSIVE origin_ancestors AS (
  SELECT gz.id FROM geo_zones gz WHERE gz.id = sqlc.arg(origin_zone_id)
  UNION ALL
  SELECT g.parent_id FROM geo_zones g
  JOIN origin_ancestors a ON g.id = a.id
  WHERE g.parent_id IS NOT NULL
), origin_descendants AS (
  SELECT gz.id FROM geo_zones gz WHERE gz.id = sqlc.arg(origin_zone_id)
  UNION ALL
  SELECT g.id FROM geo_zones g
  JOIN origin_descendants d ON g.parent_id = d.id
  WHERE g.status = 'active'
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
  r.pickup_available,
  r.dropoff_available,
  r.estimated_pickup_hours,
  r.estimated_delivery_hours,
  r.dropoff_branch_id,
  hub.name AS dropoff_branch_name,
  hub.address AS dropoff_branch_address,
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
LEFT JOIN branches hub ON hub.id = r.dropoff_branch_id AND hub.company_id = c.id AND hub.deleted_at IS NULL
WHERE r.status = 'active'
  AND (
    (sqlc.arg(service_mode)::text = 'dropoff' AND r.dropoff_available)
    OR (sqlc.arg(service_mode)::text <> 'dropoff' AND r.pickup_available)
  )
  AND (ms.company_id IS NULL OR (ms.is_listed AND ms.visibility_status = 'listed'))
  AND fz.geo_zone_id IN (
    SELECT id FROM origin_ancestors
    UNION
    SELECT id FROM origin_descendants WHERE sqlc.arg(service_mode)::text = 'dropoff'
  )
  AND tz.geo_zone_id IN (SELECT id FROM destination_ancestors)
ORDER BY c.name;

