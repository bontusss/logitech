DROP TABLE IF EXISTS pricing_quote_versions;
DROP TABLE IF EXISTS pricing_zone_rates;
DROP TABLE IF EXISTS service_zones;
DROP TABLE IF EXISTS last_mile_service_areas;
DROP TABLE IF EXISTS pricing_size_categories;
DROP TABLE IF EXISTS pricing_weight_brackets;

ALTER TABLE company_booking_routes
  DROP COLUMN IF EXISTS transportation_method,
  DROP COLUMN IF EXISTS pricing_model,
  DROP COLUMN IF EXISTS origin_branch_id,
  DROP COLUMN IF EXISTS destination_branch_id,
  DROP COLUMN IF EXISTS estimated_transit_time,
  DROP COLUMN IF EXISTS departure_schedule,
  DROP COLUMN IF EXISTS pickup_available,
  DROP COLUMN IF EXISTS delivery_available,
  DROP COLUMN IF EXISTS description,
  DROP COLUMN IF EXISTS max_weight_kg,
  DROP COLUMN IF EXISTS max_length_cm,
  DROP COLUMN IF EXISTS max_width_cm,
  DROP COLUMN IF EXISTS max_height_cm,
  DROP COLUMN IF EXISTS price_is_estimate;
