-- name: CreateManifest :one
INSERT INTO manifests (
  id, manifest_number, origin_branch_id, created_by,
  manifest_date, status, confirmed_by
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetManifestByID :one
SELECT * FROM manifests
WHERE id = $1 LIMIT 1;

-- name: ListManifests :many
SELECT * FROM manifests
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ListManifestsByCreator :many
-- statuses is a comma-separated list; an empty string means every status.
SELECT * FROM manifests
WHERE created_by = sqlc.arg('created_by')
  AND (sqlc.arg('statuses') = '' OR status::text = ANY(string_to_array(sqlc.arg('statuses'), ',')))
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CreateManifestItem :one
INSERT INTO manifest_items (
  id, manifest_id, serial_number, description, quantity,
  sender_customer_id, receiver_customer_id, destination_branch_id,
  amount, currency, payment_status
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: ListManifestItems :many
SELECT * FROM manifest_items
WHERE manifest_id = $1
ORDER BY created_at ASC;

-- name: GetCustomerByPhone :one
SELECT * FROM customers
WHERE phone = $1
LIMIT 1;

-- name: CreateCustomer :one
INSERT INTO customers (id, full_name, phone, email, address, customer_type)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;