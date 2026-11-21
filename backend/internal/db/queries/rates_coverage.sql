-- name: DeleteCompanyBookingRoute :exec
DELETE FROM company_booking_routes WHERE company_id = $1 AND id = $2;

-- name: DuplicateCompanyBookingRoute :one
INSERT INTO company_booking_routes (
  company_id, origin_state, origin_city, destination_state, destination_city, distance_km,
  standard_price_per_km_minor, express_price_per_km_minor, same_day_price_per_km_minor,
  minimum_charge_minor, transportation_method, pricing_model, origin_branch_id, destination_branch_id,
  estimated_transit_time, departure_schedule, pickup_available, delivery_available, description,
  max_weight_kg, max_length_cm, max_width_cm, max_height_cm, price_is_estimate, status
)
SELECT r.company_id, r.origin_state, r.origin_city, r.destination_state, r.destination_city, r.distance_km,
  r.standard_price_per_km_minor, r.express_price_per_km_minor, r.same_day_price_per_km_minor,
  r.minimum_charge_minor, r.transportation_method, r.pricing_model, r.origin_branch_id, r.destination_branch_id,
  r.estimated_transit_time, r.departure_schedule, r.pickup_available, r.delivery_available, r.description,
  r.max_weight_kg, r.max_length_cm, r.max_width_cm, r.max_height_cm, r.price_is_estimate, 'inactive'
FROM company_booking_routes r WHERE r.company_id = $1 AND r.id = $2
RETURNING *;

-- name: ListLastMileAreas :many
SELECT * FROM last_mile_service_areas WHERE company_id = $1 ORDER BY created_at DESC;

-- name: GetLastMileArea :one
SELECT * FROM last_mile_service_areas WHERE company_id = $1 AND id = $2;

-- name: CreateLastMileArea :one
INSERT INTO last_mile_service_areas (
  company_id, name, state, city, coverage_status, pricing_model, flat_price_minor,
  base_delivery_fee_minor, price_per_km_minor, minimum_fee_minor, maximum_distance_km,
  maximum_fee_minor, volumetric_divisor, price_is_estimate, max_weight_kg,
  max_length_cm, max_width_cm, max_height_cm, express_available, pickup_available,
  delivery_available, operating_hours, supported_parcel_types
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23)
RETURNING *;

-- name: UpdateLastMileArea :one
UPDATE last_mile_service_areas SET name=$3, state=$4, city=$5, coverage_status=$6, pricing_model=$7,
  flat_price_minor=$8, base_delivery_fee_minor=$9, price_per_km_minor=$10, minimum_fee_minor=$11,
  maximum_distance_km=$12, maximum_fee_minor=$13, volumetric_divisor=$14, price_is_estimate=$15,
  max_weight_kg=$16, max_length_cm=$17, max_width_cm=$18, max_height_cm=$19,
  express_available=$20, pickup_available=$21, delivery_available=$22,
  operating_hours=$23, supported_parcel_types=$24, updated_at=now()
WHERE company_id=$1 AND id=$2 RETURNING *;

-- name: DeleteLastMileArea :exec
DELETE FROM last_mile_service_areas WHERE company_id=$1 AND id=$2;

-- name: ListServiceZones :many
SELECT z.* FROM service_zones z JOIN last_mile_service_areas a ON a.id=z.area_id
WHERE a.company_id=$1 AND z.area_id=$2 ORDER BY z.created_at;

-- name: CreateServiceZone :one
INSERT INTO service_zones (area_id,name,neighbourhoods,state,city,lga,custom_location)
SELECT $1,$2,$3,$4,$5,$6,$7 FROM last_mile_service_areas
WHERE id=$1 AND company_id=$8 RETURNING *;

-- name: DeleteServiceZone :exec
DELETE FROM service_zones z USING last_mile_service_areas a
WHERE z.id=$2 AND z.area_id=a.id AND a.company_id=$1;

-- name: ReplacePricingZoneRates :exec
DELETE FROM pricing_zone_rates r USING last_mile_service_areas a
WHERE r.area_id=a.id AND a.company_id=$1 AND r.area_id=$2;

-- name: CreatePricingZoneRate :one
INSERT INTO pricing_zone_rates (area_id,from_zone_id,to_zone_id,price_minor)
SELECT $1,$2,$3,$4 WHERE EXISTS (
  SELECT 1 FROM last_mile_service_areas a WHERE a.id=$1 AND a.company_id=$5
) AND EXISTS (SELECT 1 FROM service_zones z WHERE z.id=$2 AND z.area_id=$1)
AND EXISTS (SELECT 1 FROM service_zones z WHERE z.id=$3 AND z.area_id=$1)
RETURNING *;







-- name: ListZoneRates :many
SELECT r.* FROM pricing_zone_rates r JOIN last_mile_service_areas a ON a.id=r.area_id
WHERE a.company_id=$1 AND r.area_id=$2;
