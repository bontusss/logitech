ALTER TABLE marketplace_quotes
  DROP COLUMN IF EXISTS estimated_delivery_hours,
  DROP COLUMN IF EXISTS estimated_pickup_hours,
  DROP COLUMN IF EXISTS estimated_pickup,
  DROP COLUMN IF EXISTS service_mode;

ALTER TABLE pricing_zone_rates
  DROP CONSTRAINT IF EXISTS pricing_zone_rates_delivery_hours_check,
  DROP CONSTRAINT IF EXISTS pricing_zone_rates_pickup_hours_check,
  DROP COLUMN IF EXISTS estimated_delivery_hours,
  DROP COLUMN IF EXISTS estimated_pickup_hours,
  DROP COLUMN IF EXISTS dropoff_available,
  DROP COLUMN IF EXISTS pickup_available;
