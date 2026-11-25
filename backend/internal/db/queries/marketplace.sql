-- name: GetMarketplaceSettings :one
SELECT * FROM marketplace_settings WHERE company_id = sqlc.arg(company_id);

-- name: UpsertMarketplaceSettings :one
INSERT INTO marketplace_settings (
  company_id, is_listed, visibility_status, public_description, published_at
) VALUES (
  sqlc.arg(company_id), sqlc.arg(is_listed), sqlc.arg(visibility_status),
  sqlc.narg(public_description), sqlc.narg(published_at)
)
ON CONFLICT (company_id) DO UPDATE SET
  is_listed = EXCLUDED.is_listed,
  visibility_status = EXCLUDED.visibility_status,
  public_description = EXCLUDED.public_description,
  published_at = COALESCE(EXCLUDED.published_at, marketplace_settings.published_at)
RETURNING *;

-- name: ListMarketplaceRoutesForLane :many
SELECT
  c.id AS company_id,
  c.name AS company_name,
  c.slug AS company_slug,
  COALESCE(b.logo_url, '') AS logo_url,
  COALESCE(ms.public_description, '') AS public_description,
  r.id AS route_id,
  r.origin_state,
  r.origin_city,
  r.destination_state,
  r.destination_city,
  r.distance_km,
  r.standard_price_per_km_minor,
  r.express_price_per_km_minor,
  r.same_day_price_per_km_minor,
  r.minimum_charge_minor,
  r.estimated_transit_time,
  r.departure_schedule,
  r.pickup_available,
  r.delivery_available,
  r.description,
  r.max_weight_kg,
  r.max_length_cm,
  r.max_width_cm,
  r.max_height_cm,
  r.price_is_estimate,
  COALESCE(bp.slug, '') AS booking_slug,
  c.accepted_item_categories,
  c.prohibited_item_categories
FROM companies c
LEFT JOIN marketplace_settings ms ON ms.company_id = c.id
JOIN company_booking_routes r ON r.company_id = c.id AND r.status = 'active'
LEFT JOIN company_branding b ON b.company_id = c.id
LEFT JOIN booking_pages bp ON bp.company_id = c.id AND bp.is_published
WHERE c.deleted_at IS NULL
  AND c.status = 'verified'
  AND (ms.company_id IS NULL OR (ms.is_listed AND ms.visibility_status = 'listed'))
  AND lower(r.origin_state) = lower(sqlc.arg(origin_state))
  AND lower(r.origin_city) = lower(sqlc.arg(origin_city))
  AND lower(r.destination_state) = lower(sqlc.arg(destination_state))
  AND lower(r.destination_city) = lower(sqlc.arg(destination_city))
ORDER BY c.name, r.distance_km;

-- name: ListMarketplaceLastMileAreas :many
SELECT
  c.id AS company_id,
  c.name AS company_name,
  c.slug AS company_slug,
  COALESCE(b.logo_url, '') AS logo_url,
  COALESCE(ms.public_description, COALESCE(b.description, c.description, '')) AS public_description,
  a.id AS area_id,
  a.name AS area_name,
  COALESCE(a.state, c.state) AS state,
  COALESCE(a.city, c.city) AS city,
  a.pricing_model,
  a.flat_price_minor,
  a.base_delivery_fee_minor,
  a.price_per_km_minor,
  a.minimum_fee_minor,
  a.maximum_distance_km,
  a.maximum_fee_minor,
  a.volumetric_divisor,
  a.price_is_estimate,
  a.max_weight_kg,
  a.max_length_cm,
  a.max_width_cm,
  a.max_height_cm,
  a.express_available,
  a.pickup_available,
  a.delivery_available,
  a.operating_hours,
  a.supported_parcel_types,
  COALESCE(bp.slug, '') AS booking_slug,
  c.accepted_item_categories,
  c.prohibited_item_categories
FROM last_mile_service_areas a
JOIN companies c ON c.id = a.company_id
  AND c.deleted_at IS NULL
  AND c.status = 'verified'
LEFT JOIN marketplace_settings ms ON ms.company_id = c.id
LEFT JOIN company_branding b ON b.company_id = c.id
LEFT JOIN booking_pages bp ON bp.company_id = c.id AND bp.is_published
WHERE a.coverage_status = 'active'
  AND a.pricing_model IN ('flat', 'distance')
  AND (ms.company_id IS NULL OR (ms.is_listed AND ms.visibility_status = 'listed'))
  AND (
    lower(trim(COALESCE(a.city, c.city))) = lower(trim(sqlc.arg(city)))
    OR lower(trim(sqlc.arg(city))) LIKE lower(trim(COALESCE(a.city, c.city))) || ' %'
  )
  AND (
    COALESCE(a.state, c.state) IS NULL
    OR lower(trim(regexp_replace(COALESCE(a.state, c.state), '\s*state$', '', 'i')))
       = lower(trim(regexp_replace(sqlc.arg(state), '\s*state$', '', 'i')))
  )
ORDER BY c.name, a.name;

-- name: ListMarketplaceProviders :many
SELECT
  c.id,
  c.name,
  c.slug,
  COALESCE(b.logo_url, '') AS logo_url,
  COALESCE(b.cover_image_url, '') AS cover_image_url,
  COALESCE(ms.public_description, '') AS public_description,
  COUNT(r.id)::int AS route_count,
  MIN(r.minimum_charge_minor)::bigint AS from_price_minor
FROM companies c
LEFT JOIN marketplace_settings ms ON ms.company_id = c.id
LEFT JOIN company_branding b ON b.company_id = c.id
LEFT JOIN company_booking_routes r ON r.company_id = c.id AND r.status = 'active'
WHERE c.deleted_at IS NULL
  AND c.status = 'verified'
  AND (ms.company_id IS NULL OR (ms.is_listed AND ms.visibility_status = 'listed'))
  AND (
    sqlc.narg(state)::text IS NULL
    OR lower(r.origin_state) = lower(sqlc.narg(state)::text)
    OR lower(r.destination_state) = lower(sqlc.narg(state)::text)
  )
  AND (
    sqlc.narg(city)::text IS NULL
    OR lower(r.origin_city) = lower(sqlc.narg(city)::text)
    OR lower(r.destination_city) = lower(sqlc.narg(city)::text)
  )
GROUP BY c.id, c.name, c.slug, b.logo_url, b.cover_image_url, ms.public_description
ORDER BY c.name;

-- name: GetMarketplaceProviderBySlug :one
SELECT
  c.id,
  c.name,
  c.slug,
  c.state,
  c.city,
  c.country,
  c.business_email,
  c.phone,
  c.verified_at,
  COALESCE(b.logo_url, '') AS logo_url,
  COALESCE(b.cover_image_url, '') AS cover_image_url,
  COALESCE(b.primary_color, '#2b4bff') AS primary_color,
  COALESCE(b.secondary_color, '#0a1220') AS secondary_color,
  COALESCE(ms.public_description, '') AS public_description,
  ms.published_at,
  s.subdomain,
  COALESCE(cs.currency, 'NGN') AS currency
FROM companies c
LEFT JOIN marketplace_settings ms ON ms.company_id = c.id
LEFT JOIN company_branding b ON b.company_id = c.id
LEFT JOIN company_subdomains s ON s.company_id = c.id
LEFT JOIN company_settings cs ON cs.company_id = c.id
WHERE c.slug = sqlc.arg(slug)
  AND c.deleted_at IS NULL
  AND c.status = 'verified'
  AND (ms.company_id IS NULL OR (ms.is_listed AND ms.visibility_status = 'listed'));

-- name: ListMarketplaceProviderRoutes :many
SELECT
  r.id,
  r.origin_state,
  r.origin_city,
  r.destination_state,
  r.destination_city,
  r.distance_km,
  r.standard_price_per_km_minor,
  r.express_price_per_km_minor,
  r.same_day_price_per_km_minor,
  r.minimum_charge_minor,
  r.estimated_transit_time,
  r.departure_schedule,
  r.pickup_available,
  r.delivery_available,
  r.description,
  r.price_is_estimate
FROM company_booking_routes r
WHERE r.company_id = sqlc.arg(company_id) AND r.status = 'active'
ORDER BY r.origin_state, r.origin_city, r.destination_state, r.destination_city;

-- name: ListMarketplaceLocations :many
SELECT DISTINCT
  r.origin_state AS state,
  r.origin_city AS city
FROM companies c
LEFT JOIN marketplace_settings ms ON ms.company_id = c.id
JOIN company_booking_routes r ON r.company_id = c.id AND r.status = 'active'
WHERE c.deleted_at IS NULL AND c.status = 'verified'
  AND (ms.company_id IS NULL OR (ms.is_listed AND ms.visibility_status = 'listed'))
UNION
SELECT DISTINCT
  r.destination_state AS state,
  r.destination_city AS city
FROM companies c
LEFT JOIN marketplace_settings ms ON ms.company_id = c.id
JOIN company_booking_routes r ON r.company_id = c.id AND r.status = 'active'
WHERE c.deleted_at IS NULL AND c.status = 'verified'
  AND (ms.company_id IS NULL OR (ms.is_listed AND ms.visibility_status = 'listed'))
ORDER BY state, city;

-- name: CreateMarketplaceSearchLog :one
INSERT INTO marketplace_search_logs (
  customer_id, origin_state, origin_city, destination_state, destination_city,
  search_metadata, result_count, failure_count
) VALUES (
  sqlc.narg(customer_id), sqlc.arg(origin_state), sqlc.arg(origin_city),
  sqlc.arg(destination_state), sqlc.arg(destination_city),
  sqlc.arg(search_metadata), sqlc.arg(result_count), sqlc.arg(failure_count)
)
RETURNING *;

-- name: InsertMarketplaceQuote :one
INSERT INTO marketplace_quotes (
  quote_reference, search_id, company_id, provider_name, provider_logo, provider_slug, route_id, origin_state, origin_city,
  destination_state, destination_city, service_key, service_name, distance_km,
  package_snapshot, chargeable_weight_kg, base_price_minor, additional_charges_minor,
  discount_minor, total_price_minor, currency, price_confirmation_type,
  price_breakdown, rate_snapshot, estimated_delivery, estimated_duration_days,
  pickup_available, delivery_available, delivery_method, conditions, status, expires_at
) VALUES (
  sqlc.arg(quote_reference), sqlc.narg(search_id), sqlc.arg(company_id), sqlc.arg(provider_name), sqlc.arg(provider_logo), sqlc.arg(provider_slug), sqlc.narg(route_id),
  sqlc.arg(origin_state), sqlc.arg(origin_city), sqlc.arg(destination_state), sqlc.arg(destination_city),
  sqlc.arg(service_key), sqlc.arg(service_name), sqlc.arg(distance_km),
  sqlc.arg(package_snapshot), sqlc.arg(chargeable_weight_kg), sqlc.arg(base_price_minor),
  sqlc.arg(additional_charges_minor), sqlc.arg(discount_minor), sqlc.arg(total_price_minor),
  sqlc.arg(currency), sqlc.arg(price_confirmation_type), sqlc.arg(price_breakdown),
  sqlc.arg(rate_snapshot), sqlc.narg(estimated_delivery), sqlc.narg(estimated_duration_days),
  sqlc.arg(pickup_available), sqlc.arg(delivery_available), sqlc.arg(delivery_method),
  sqlc.narg(conditions), 'active', sqlc.arg(expires_at)
)
RETURNING *;

-- name: GetMarketplaceQuoteByReference :one
SELECT * FROM marketplace_quotes WHERE quote_reference = sqlc.arg(quote_reference);

-- name: ExpireMarketplaceQuotes :exec
UPDATE marketplace_quotes SET status = 'expired'
WHERE status = 'active' AND expires_at <= now();

-- name: MarkMarketplaceQuoteSelected :exec
UPDATE marketplace_quotes SET status = 'selected'
WHERE id = sqlc.arg(id) AND status = 'active' AND expires_at > now();

-- name: InsertSavedQuote :one
INSERT INTO customer_saved_quotes (customer_id, guest_token_hash, quote_id, selection_status)
VALUES (sqlc.narg(customer_id), sqlc.narg(guest_token_hash), sqlc.arg(quote_id), 'selected')
ON CONFLICT DO NOTHING
RETURNING *;

-- name: GetSavedQuoteForCustomer :one
SELECT * FROM customer_saved_quotes
WHERE customer_id = sqlc.arg(customer_id) AND quote_id = sqlc.arg(quote_id)
LIMIT 1;

-- name: GetSavedQuoteForGuest :one
SELECT * FROM customer_saved_quotes
WHERE guest_token_hash = sqlc.arg(guest_token_hash) AND quote_id = sqlc.arg(quote_id)
LIMIT 1;

-- name: ListCustomerSavedQuotes :many
SELECT
  sq.id AS saved_quote_id,
  sq.selection_status,
  sq.created_at AS selected_at,
  q.*
FROM customer_saved_quotes sq
JOIN marketplace_quotes q ON q.id = sq.quote_id
WHERE sq.customer_id = sqlc.arg(customer_id)
ORDER BY sq.created_at DESC
LIMIT 100;

-- name: GetCustomerSavedQuoteByReference :one
SELECT
  sq.id AS saved_quote_id,
  sq.selection_status,
  sq.created_at AS selected_at,
  q.*
FROM customer_saved_quotes sq
JOIN marketplace_quotes q ON q.id = sq.quote_id
WHERE sq.customer_id = sqlc.arg(customer_id)
  AND q.quote_reference = sqlc.arg(quote_reference)
LIMIT 1;

-- name: CountMarketplaceSavedQuotes :one
SELECT COUNT(*)::bigint AS total FROM customer_saved_quotes WHERE customer_id = sqlc.arg(customer_id);
