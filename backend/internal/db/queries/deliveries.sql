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
