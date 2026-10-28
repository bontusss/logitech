-- name: GetCustomerByPhone :one
SELECT * FROM customers
WHERE phone = $1
LIMIT 1;

-- name: GetCustomerByID :one
SELECT * FROM customers
WHERE id = $1 LIMIT 1;

-- name: CreateCustomer :one
INSERT INTO customers (id, full_name, phone, email, address, customer_type)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpsertCustomerByPhone :one
INSERT INTO customers (id, full_name, phone, email, address, customer_type)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (phone) WHERE phone IS NOT NULL
DO UPDATE SET full_name = EXCLUDED.full_name, updated_at = now()
RETURNING *;

-- name: UpdateCustomerContact :one
UPDATE customers
SET full_name = COALESCE(sqlc.narg('full_name'), full_name),
    phone     = COALESCE(sqlc.narg('phone'), phone),
    email     = COALESCE(sqlc.narg('email'), email),
    updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING *;
