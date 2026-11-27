-- name: ListCustomerAddresses :many
SELECT *
FROM customer_addresses
WHERE customer_id = $1
ORDER BY default_pickup DESC, default_delivery DESC, created_at DESC;

-- name: GetCustomerAddress :one
SELECT *
FROM customer_addresses
WHERE id = $1 AND customer_id = $2;

-- name: ClearCustomerPickupDefault :exec
UPDATE customer_addresses
SET default_pickup = false
WHERE customer_id = $1 AND default_pickup;

-- name: ClearCustomerDeliveryDefault :exec
UPDATE customer_addresses
SET default_delivery = false
WHERE customer_id = $1 AND default_delivery;

-- name: CreateCustomerAddress :one
INSERT INTO customer_addresses (
  customer_id, label, address_type, contact_name, phone, state, city, area,
  street_address, directions, latitude, longitude, default_pickup, default_delivery
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
)
RETURNING *;

-- name: UpdateCustomerAddress :one
UPDATE customer_addresses
SET label = $3,
    address_type = $4,
    contact_name = $5,
    phone = $6,
    state = $7,
    city = $8,
    area = $9,
    street_address = $10,
    directions = $11,
    latitude = $12,
    longitude = $13,
    default_pickup = $14,
    default_delivery = $15
WHERE id = $1 AND customer_id = $2
RETURNING *;

-- name: DeleteCustomerAddress :execrows
DELETE FROM customer_addresses
WHERE id = $1 AND customer_id = $2;

-- name: ListCustomerFavourites :many
SELECT
  c.id,
  c.name,
  c.slug,
  COALESCE(b.logo_url, '') AS logo_url,
  COALESCE(ms.public_description, '') AS description,
  COALESCE((
    SELECT ROUND(AVG(r.rating)::numeric, 1)
    FROM company_reviews r
    WHERE r.company_id = c.id
  ), 0)::float8 AS rating,
  (
    SELECT COUNT(*)::bigint
    FROM pricing_zone_rates p
    WHERE p.company_id = c.id AND p.status = 'active'
  ) AS route_count,
  f.created_at
FROM customer_favourite_providers f
JOIN companies c ON c.id = f.company_id AND c.deleted_at IS NULL
LEFT JOIN company_branding b ON b.company_id = c.id
LEFT JOIN marketplace_settings ms ON ms.company_id = c.id
WHERE f.customer_id = $1
ORDER BY f.created_at DESC;

-- name: AddCustomerFavourite :execrows
INSERT INTO customer_favourite_providers (customer_id, company_id)
SELECT sqlc.arg(customer_id), c.id
FROM companies c
WHERE c.id = sqlc.arg(company_id) AND c.deleted_at IS NULL AND c.status = 'verified'
ON CONFLICT DO NOTHING;

-- name: RemoveCustomerFavourite :execrows
DELETE FROM customer_favourite_providers
WHERE customer_id = $1 AND company_id = $2;

-- name: CreateDeliveryIssue :one
INSERT INTO delivery_issues (delivery_id, customer_id, issue_type, description)
SELECT sqlc.arg(delivery_id)::uuid, sqlc.arg(customer_id)::uuid,
       sqlc.arg(issue_type)::text, sqlc.arg(description)::text
WHERE EXISTS (
  SELECT 1 FROM deliveries d
  WHERE d.id = sqlc.arg(delivery_id) AND d.customer_id = sqlc.arg(customer_id)
)
RETURNING *;

-- name: ListCustomerDeliveryIssues :many
SELECT
  i.id,
  i.delivery_id,
  i.issue_type,
  i.description,
  i.status,
  i.created_at,
  d.tracking_code,
  d.pickup_address,
  d.dropoff_address,
  COALESCE(b.public_name, c.name) AS company_name
FROM delivery_issues i
JOIN deliveries d ON d.id = i.delivery_id
JOIN companies c ON c.id = d.company_id
LEFT JOIN company_branding b ON b.company_id = c.id
WHERE i.customer_id = $1
ORDER BY i.created_at DESC
LIMIT 100;

-- name: ListCustomerRecurringShipments :many
SELECT *
FROM customer_recurring_shipments
WHERE customer_id = $1
ORDER BY created_at DESC;

-- name: CreateCustomerRecurringShipment :one
INSERT INTO customer_recurring_shipments (
  customer_id, label, frequency, search_payload
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: DeleteCustomerRecurringShipment :execrows
DELETE FROM customer_recurring_shipments
WHERE id = $1 AND customer_id = $2;
