-- name: CreateDelivery :one
INSERT INTO deliveries (
  company_id, sender_name, sender_phone, receiver_name, receiver_phone,
  pickup_address, dropoff_address, item_description, declared_value_minor,
  delivery_price_minor, payer, is_paid, payment_method, assigned_driver_id,
  status, created_by
) SELECT
  sqlc.arg(company_id), sqlc.arg(sender_name), sqlc.arg(sender_phone),
  sqlc.arg(receiver_name), sqlc.arg(receiver_phone), sqlc.arg(pickup_address),
  sqlc.arg(dropoff_address), sqlc.arg(item_description), sqlc.arg(declared_value_minor),
  sqlc.arg(delivery_price_minor), sqlc.arg(payer), sqlc.arg(is_paid),
  NULLIF(sqlc.arg(payment_method)::text, ''),
  NULLIF(sqlc.arg(assigned_driver_id)::uuid, '00000000-0000-0000-0000-000000000000'::uuid),
  sqlc.arg(status), sqlc.arg(created_by)
WHERE NULLIF(sqlc.arg(assigned_driver_id)::uuid, '00000000-0000-0000-0000-000000000000'::uuid) IS NULL
   OR EXISTS (
     SELECT 1 FROM driver_profiles d
     WHERE d.id = sqlc.arg(assigned_driver_id)::uuid
       AND d.company_id = sqlc.arg(company_id)::uuid
       AND d.status = 'active'
   )
RETURNING *;

-- name: ListCompanyDeliveries :many
SELECT
  d.*,
  driver.full_name AS assigned_driver_name
FROM deliveries d
LEFT JOIN driver_profiles driver ON driver.id = d.assigned_driver_id
WHERE d.company_id = $1
ORDER BY d.created_at DESC
LIMIT 200;

-- name: ListCompanyRequests :many
SELECT d.*, driver.full_name AS assigned_driver_name
FROM deliveries d
LEFT JOIN driver_profiles driver ON driver.id = d.assigned_driver_id
WHERE d.company_id = $1
  AND d.booking_source = 'booking_page'
ORDER BY d.created_at DESC
LIMIT 200;

-- name: GetCompanyDelivery :one
SELECT d.*, driver.full_name AS assigned_driver_name
FROM deliveries d
LEFT JOIN driver_profiles driver ON driver.id = d.assigned_driver_id
WHERE d.company_id = $1 AND d.id = $2;

-- name: AcceptBookingRequest :one
UPDATE deliveries
SET request_status = 'accepted', status = 'created', accepted_at = now(), tracking_code =
  'TRK-' || upper(substr(replace(id::text, '-', ''), 1, 12)), updated_at = now()
WHERE company_id = $1 AND id = $2 AND request_status = 'pending'
RETURNING *;

-- name: DeclineBookingRequest :one
UPDATE deliveries
SET request_status = 'declined', status = 'cancelled', declined_at = now(), updated_at = now()
WHERE company_id = $1 AND id = $2 AND request_status = 'pending'
RETURNING *;

-- name: AssignDeliveryDriver :one
UPDATE deliveries AS delivery
SET assigned_driver_id = NULLIF($3::uuid, '00000000-0000-0000-0000-000000000000'),
    status = CASE WHEN $3::uuid = '00000000-0000-0000-0000-000000000000' THEN 'created'
                  ELSE 'assigned' END,
    updated_at = now()
WHERE delivery.company_id = $1 AND delivery.id = $2
  AND (NULLIF($3::uuid, '00000000-0000-0000-0000-000000000000') IS NULL
       OR EXISTS (SELECT 1 FROM driver_profiles driver
                  WHERE driver.id = $3::uuid AND driver.company_id = $1 AND driver.status = 'active'))
RETURNING *;

-- name: UpdateCompanyDelivery :one
UPDATE deliveries
SET sender_name = $3, sender_phone = $4, receiver_name = $5, receiver_phone = $6,
    pickup_address = $7, dropoff_address = $8, item_description = $9,
    delivery_price_minor = $10, declared_value_minor = $11,
    package_category = $12, weight_kg = $13, additional_instructions = $14,
    updated_at = now()
WHERE company_id = $1 AND id = $2
RETURNING *;

-- name: UpdateCompanyDeliveryStatus :one
UPDATE deliveries
SET status = $3, updated_at = now()
WHERE company_id = $1 AND id = $2
RETURNING *;

-- name: AddDeliveryTimelineEvent :one
INSERT INTO delivery_timeline_events (delivery_id, company_id, status, note, created_by)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListDeliveryTimeline :many
SELECT event.*
FROM delivery_timeline_events event
JOIN deliveries delivery ON delivery.id = event.delivery_id
WHERE delivery.company_id = $1 AND event.delivery_id = $2
ORDER BY event.created_at ASC;
