-- Shipments (manifest items) and the payments made against them.

-- name: CreateManifestItem :one
INSERT INTO manifest_items (
  id, tracking_number, serial_number, description, quantity,
  sender_customer_id, receiver_customer_id,
  origin_branch_id, destination_branch_id,
  amount, currency, status, recorded_by, photo_url, notes
) VALUES (
  sqlc.arg('id'),
  sqlc.arg('tracking_number'),
  sqlc.narg('serial_number'),
  sqlc.arg('description'),
  sqlc.arg('quantity'),
  sqlc.arg('sender_customer_id'),
  sqlc.arg('receiver_customer_id'),
  sqlc.arg('origin_branch_id'),
  sqlc.arg('destination_branch_id'),
  sqlc.arg('amount'),
  sqlc.arg('currency'),
  sqlc.arg('status'),
  sqlc.arg('recorded_by'),
  sqlc.narg('photo_url'),
  sqlc.narg('notes')
)
RETURNING *;

-- name: GetManifestItemByID :one
SELECT * FROM manifest_items
WHERE id = sqlc.arg('id') LIMIT 1;

-- name: GetManifestItemForUpdate :one
SELECT * FROM manifest_items
WHERE id = sqlc.arg('id') LIMIT 1
FOR UPDATE;

-- name: GetManifestItemByTrackingNumber :one
SELECT * FROM manifest_items
WHERE tracking_number = sqlc.arg('tracking_number') LIMIT 1;

-- name: TrackingNumberExists :one
SELECT EXISTS(SELECT 1 FROM manifest_items WHERE tracking_number = sqlc.arg('tracking_number'));

-- name: UpdateManifestItem :one
UPDATE manifest_items
SET description           = COALESCE(sqlc.narg('description'), description),
    serial_number         = COALESCE(sqlc.narg('serial_number'), serial_number),
    quantity              = COALESCE(sqlc.narg('quantity'), quantity),
    receiver_customer_id  = COALESCE(sqlc.narg('receiver_customer_id'), receiver_customer_id),
    destination_branch_id = COALESCE(sqlc.narg('destination_branch_id'), destination_branch_id),
    amount                = COALESCE(sqlc.narg('amount'), amount),
    currency              = COALESCE(sqlc.narg('currency'), currency),
    notes                 = COALESCE(sqlc.narg('notes'), notes),
    updated_at            = now()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: MarkManifestItemDispatched :one
UPDATE manifest_items
SET status = 'in_transit',
    dispatched_at = COALESCE(dispatched_at, now()),
    updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: MarkManifestItemReceived :one
UPDATE manifest_items
SET status            = sqlc.arg('status'),
    received_by       = sqlc.arg('received_by'),
    received_at       = now(),
    received_quantity = sqlc.arg('received_quantity'),
    condition         = sqlc.arg('condition'),
    notes             = COALESCE(sqlc.narg('notes'), notes),
    updated_at        = now()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: CancelManifestItem :one
UPDATE manifest_items
SET status = 'cancelled',
    updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: ResetManifestItemToTransit :one
UPDATE manifest_items
SET status = 'in_transit',
    received_by = NULL,
    received_at = NULL,
    received_quantity = NULL,
    condition = NULL,
    updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: CreatePayment :one
INSERT INTO payments (
  id, manifest_item_id, amount, currency, method, reference,
  recorded_by, paid_at, notes
) VALUES (
  sqlc.arg('id'),
  sqlc.arg('manifest_item_id'),
  sqlc.arg('amount'),
  sqlc.arg('currency'),
  sqlc.arg('method'),
  sqlc.narg('reference'),
  sqlc.arg('recorded_by'),
  now(),
  sqlc.narg('notes')
)
RETURNING *;

-- name: GetPaymentByID :one
SELECT * FROM payments
WHERE id = sqlc.arg('id') LIMIT 1;

-- name: ListPaymentsByItem :many
SELECT p.*, COALESCE(s.full_name, '') AS recorded_by_name
FROM payments p
LEFT JOIN staff s ON s.id = p.recorded_by
WHERE p.manifest_item_id = sqlc.arg('manifest_item_id')
ORDER BY p.paid_at ASC;

-- name: SumPaymentsByItem :one
SELECT COALESCE(SUM(amount), 0)::float8 AS total
FROM payments
WHERE manifest_item_id = sqlc.arg('manifest_item_id');
