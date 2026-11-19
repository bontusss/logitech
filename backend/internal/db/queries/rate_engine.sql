-- name: ListCountries :many
SELECT * FROM countries WHERE deleted_at IS NULL;

-- name: CreateCountry :one
INSERT INTO countries (name, code, numeric_code, status) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: ListStatesByCountry :many
SELECT * FROM states WHERE country_id = $1 AND deleted_at IS NULL;

-- name: CreateState :one
INSERT INTO states (name, code, country_id, status) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: ListCitiesByState :many
SELECT * FROM cities WHERE state_id = $1 AND deleted_at IS NULL;

-- name: CreateCity :one
INSERT INTO cities (name, state_id, country_id, status) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: ListTerminalsByCity :many
SELECT * FROM terminals WHERE city_id = $1 AND deleted_at IS NULL;

-- name: CreateTerminal :one
INSERT INTO terminals (name, code, city_id, address, status) VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: ListShippingServices :many
SELECT * FROM shipping_services WHERE company_id = $1 AND deleted_at IS NULL;

-- name: CreateShippingService :one
INSERT INTO shipping_services (company_id, name, description, service_code, delivery_type, estimated_duration, min_weight_kg, max_weight_kg, max_dimension_cm, volumetric_divisor, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING *;

-- name: ListCompanyRoutes :many
SELECT * FROM company_routes WHERE company_id = $1 AND deleted_at IS NULL;

-- name: CreateCompanyRoute :one
INSERT INTO company_routes (company_id, origin_location_type, origin_location_id, destination_location_type, destination_location_id, is_round_trip, status) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: ListRateCards :many
SELECT * FROM rate_cards WHERE company_id = $1 AND deleted_at IS NULL;

-- name: CreateRateCard :one
INSERT INTO rate_cards (company_id, name, description, status, is_published, published_at)
VALUES ($1, $2, $3, $4, $4 = 'published', CASE WHEN $4 = 'published' THEN now() END)
RETURNING *;

-- name: ListRateCardVersions :many
SELECT * FROM rate_card_versions WHERE rate_card_id = $1 AND deleted_at IS NULL;

-- name: CreateRateCardVersion :one
INSERT INTO rate_card_versions (rate_card_id, version, is_active, effective_from, effective_until, created_by) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: ListPricingRules :many
SELECT * FROM pricing_rules WHERE rate_card_version_id = $1 AND deleted_at IS NULL;

-- name: CreatePricingRule :one
INSERT INTO pricing_rules (rate_card_version_id, rule_type, name, description, fixed_amount, per_kg_base_charge, per_kg_rate, per_kg_minimum, per_kg_maximum, volumetric_divisor, category_id, applicable_routes, applicable_services, applicable_categories, effective_from, effective_until, is_active) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17) RETURNING *;

-- name: ListPricingTiers :many
SELECT * FROM pricing_tiers WHERE rate_card_version_id = $1 AND deleted_at IS NULL;

-- name: CreatePricingTier :one
INSERT INTO pricing_tiers (rate_card_version_id, min_weight_kg, max_weight_kg, price_cents, name, description) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: ListPackageCategories :many
SELECT * FROM package_categories WHERE deleted_at IS NULL;

-- name: CreatePackageCategory :one
INSERT INTO package_categories (name, description, sort_order, is_default) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: ListAdditionalCharges :many
SELECT * FROM additional_charges WHERE rate_card_version_id = $1 AND company_id = $2 AND deleted_at IS NULL;

-- name: CreateAdditionalCharge :one
INSERT INTO additional_charges (rate_card_version_id, company_id, name, charge_type, amount_cents, percentage, is_mandatory, is_included_in_price) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *;

-- name: ListShippingQuotes :many
SELECT * FROM shipping_quotes WHERE company_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: CreateShippingQuote :one
INSERT INTO shipping_quotes (company_id, customer_id, origin_location_type, origin_location_id, destination_location_type, destination_location_id, selected_service_id, rate_card_version_id, quote_reference, reference_weight_kg, reference_length_cm, reference_width_cm, reference_height_cm, reference_category, chargeable_weight_kg, base_price_cents, total_price_cents, currency, estimated_duration, status, expires_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21) RETURNING *;