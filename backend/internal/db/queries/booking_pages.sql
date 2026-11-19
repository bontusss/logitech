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
  b.cover_image_url,
  COALESCE(cs.currency, 'NGN') AS currency
FROM booking_pages bp
JOIN companies c ON c.id = bp.company_id
  AND c.deleted_at IS NULL
  AND c.status = 'verified'
LEFT JOIN company_branding b ON b.company_id = c.id
LEFT JOIN company_subdomains s ON s.company_id = c.id
LEFT JOIN company_settings cs ON cs.company_id = c.id
WHERE bp.slug = $1 AND bp.is_published;

-- name: ListPublicBookingRates :many
SELECT DISTINCT ON (rc.id)
  rc.id AS rate_card_id,
  rc.name,
  rc.currency,
  pt.price_cents
FROM rate_cards rc
JOIN rate_card_versions v ON v.rate_card_id = rc.id
  AND v.is_active
  AND v.deleted_at IS NULL
  AND v.effective_from <= now()
  AND (v.effective_until IS NULL OR v.effective_until > now())
JOIN pricing_tiers pt ON pt.rate_card_version_id = v.id
  AND pt.deleted_at IS NULL
  AND pt.min_weight_kg <= sqlc.arg(weight_kg)
  AND pt.max_weight_kg >= sqlc.arg(weight_kg)
WHERE rc.company_id = sqlc.arg(company_id)
  AND rc.status = 'published'
  AND rc.is_published
  AND rc.deleted_at IS NULL
  AND (rc.effective_start IS NULL OR rc.effective_start <= now())
  AND (rc.effective_end IS NULL OR rc.effective_end > now())
ORDER BY rc.id, pt.min_weight_kg DESC;

-- name: GetPublicBookingRate :one
SELECT rc.id AS rate_card_id, rc.name, rc.currency, pt.price_cents
FROM rate_cards rc
JOIN rate_card_versions v ON v.rate_card_id = rc.id
  AND v.is_active
  AND v.deleted_at IS NULL
  AND v.effective_from <= now()
  AND (v.effective_until IS NULL OR v.effective_until > now())
JOIN pricing_tiers pt ON pt.rate_card_version_id = v.id
  AND pt.deleted_at IS NULL
  AND pt.min_weight_kg <= sqlc.arg(weight_kg)
  AND pt.max_weight_kg >= sqlc.arg(weight_kg)
WHERE rc.company_id = sqlc.arg(company_id)
  AND rc.id = sqlc.arg(rate_card_id)
  AND rc.status = 'published'
  AND rc.is_published
  AND rc.deleted_at IS NULL
  AND (rc.effective_start IS NULL OR rc.effective_start <= now())
  AND (rc.effective_end IS NULL OR rc.effective_end > now())
ORDER BY pt.min_weight_kg DESC
LIMIT 1;

-- name: CreatePublicBookingDelivery :one
INSERT INTO deliveries (
  company_id, sender_name, sender_phone, receiver_name, receiver_phone,
  pickup_address, dropoff_address, item_description, declared_value_minor,
  delivery_price_minor, payer, is_paid, payment_method, assigned_driver_id,
  status, created_by, package_category, weight_kg, additional_instructions,
  package_photo_url, pickup_latitude, pickup_longitude, dropoff_latitude,
  dropoff_longitude, selected_route_id, selected_rate_card_id, price_is_estimate, booking_source,
  delivery_speed, distance_km
)
SELECT
  sqlc.arg(company_id), sqlc.arg(sender_name), sqlc.arg(sender_phone),
  sqlc.arg(receiver_name), sqlc.arg(receiver_phone), sqlc.arg(pickup_address),
  sqlc.arg(dropoff_address), sqlc.arg(item_description), sqlc.arg(declared_value_minor),
  sqlc.arg(delivery_price_minor), 'sender', false, NULL, NULL,
  'requested', NULL, sqlc.arg(package_category), sqlc.arg(weight_kg),
  sqlc.arg(additional_instructions), NULLIF(sqlc.arg(package_photo_url)::text, ''),
  sqlc.narg(pickup_latitude), sqlc.narg(pickup_longitude),
  sqlc.narg(dropoff_latitude), sqlc.narg(dropoff_longitude),
  sqlc.narg(selected_route_id), sqlc.narg(selected_rate_card_id),
  sqlc.arg(price_is_estimate), 'booking_page', sqlc.arg(delivery_speed), sqlc.arg(distance_km)
WHERE sqlc.narg(selected_route_id)::uuid IS NULL
   OR EXISTS (
     SELECT 1 FROM company_booking_routes r
     WHERE r.id = sqlc.narg(selected_route_id)::uuid
       AND r.company_id = sqlc.arg(company_id)::uuid
       AND r.status = 'active'
   )
RETURNING *;
