-- name: GetBookingForClaim :one
SELECT id, company_id, customer_id, tracking_code, request_status
FROM deliveries
WHERE id = $1 AND booking_source = 'booking_page';

-- name: ClaimBookingForCustomer :execrows
UPDATE deliveries
SET customer_id = $2, updated_at = now()
WHERE id = $1
  AND booking_source = 'booking_page'
  AND (customer_id IS NULL OR customer_id = $2);

-- name: ListCustomerBookings :many
SELECT
  d.id,
  d.status,
  d.request_status,
  d.tracking_code,
  d.pickup_address,
  d.dropoff_address,
  d.package_category,
  d.item_description,
  d.weight_kg,
  d.delivery_price_minor,
  d.declared_value_minor,
  d.price_is_estimate,
  d.delivery_speed,
  d.distance_km,
  d.created_at,
  d.updated_at,
  c.id AS company_id,
  COALESCE(b.public_name, c.name) AS company_name,
  b.logo_url AS company_logo_url,
  COALESCE(cs.currency, 'NGN') AS currency
FROM deliveries d
JOIN companies c ON c.id = d.company_id AND c.deleted_at IS NULL
LEFT JOIN company_branding b ON b.company_id = c.id
LEFT JOIN company_settings cs ON cs.company_id = c.id
WHERE d.customer_id = $1
ORDER BY d.created_at DESC
LIMIT 200;

-- name: ListCompanyCustomers :many
SELECT
  u.id AS customer_id,
  u.full_name,
  u.email,
  u.phone,
  u.created_at AS joined_at,
  COUNT(d.id)::bigint AS total_bookings,
  COUNT(d.id) FILTER (WHERE d.status = 'delivered')::bigint AS delivered_bookings,
  COUNT(d.id) FILTER (WHERE d.status NOT IN ('delivered', 'cancelled'))::bigint AS active_bookings,
  MIN(d.created_at)::timestamptz AS first_booking_at,
  MAX(d.created_at)::timestamptz AS last_booking_at
FROM deliveries d
JOIN users u ON u.id = d.customer_id
WHERE d.company_id = $1
GROUP BY u.id, u.full_name, u.email, u.phone, u.created_at
ORDER BY MAX(d.created_at) DESC;

-- name: ListCompanyCustomerDeliveries :many
SELECT
  d.*,
  driver.full_name AS assigned_driver_name
FROM deliveries d
LEFT JOIN driver_profiles driver ON driver.id = d.assigned_driver_id
WHERE d.company_id = $1 AND d.customer_id = $2
ORDER BY d.created_at DESC
LIMIT 200;
