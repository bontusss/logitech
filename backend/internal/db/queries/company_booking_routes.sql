-- name: ListCompanyBookingRoutes :many
SELECT *
FROM company_booking_routes
WHERE company_id = $1
ORDER BY created_at DESC;

-- name: ListPublicBookingRoutes :many
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
  r.price_is_estimate,
  COALESCE(s.currency, 'NGN') AS currency
FROM company_booking_routes r
LEFT JOIN company_settings s ON s.company_id = r.company_id
WHERE r.company_id = $1
  AND r.status = 'active'
ORDER BY r.origin_state, r.origin_city, r.destination_state, r.destination_city
LIMIT 100;

-- name: GetPublicBookingRoute :one
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
  r.price_is_estimate,
  COALESCE(s.currency, 'NGN') AS currency
FROM company_booking_routes r
LEFT JOIN company_settings s ON s.company_id = r.company_id
WHERE r.company_id = $1
  AND r.id = $2
  AND r.status = 'active';

-- name: CreateCompanyBookingRoute :one
INSERT INTO company_booking_routes (
  company_id, origin_state, origin_city, destination_state, destination_city,
  distance_km, standard_price_per_km_minor, express_price_per_km_minor,
  same_day_price_per_km_minor, minimum_charge_minor
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: UpdateCompanyBookingRoute :one
UPDATE company_booking_routes
SET origin_state = $3,
    origin_city = $4,
    destination_state = $5,
    destination_city = $6,
    distance_km = $7,
    standard_price_per_km_minor = $8,
    express_price_per_km_minor = $9,
    same_day_price_per_km_minor = $10,
    minimum_charge_minor = $11,
    updated_at = now()
WHERE company_id = $1 AND id = $2
RETURNING *;

-- name: SetCompanyBookingRouteStatus :one
UPDATE company_booking_routes
SET status = $3, updated_at = now()
WHERE company_id = $1 AND id = $2
RETURNING *;
