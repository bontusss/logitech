DROP INDEX IF EXISTS pricing_zone_rates_company_direction_unique;
DROP INDEX IF EXISTS pricing_zone_rates_company_idx;

ALTER TABLE marketplace_quotes
  DROP CONSTRAINT IF EXISTS marketplace_quotes_service_key_check;
ALTER TABLE marketplace_quotes
  ADD CONSTRAINT marketplace_quotes_service_key_check
  CHECK (service_key IN ('standard', 'express', 'same_day', 'last_mile'));

ALTER TABLE pricing_zone_rates
  DROP COLUMN IF EXISTS status,
  DROP COLUMN IF EXISTS additional_weight_fee_minor,
  DROP COLUMN IF EXISTS included_weight_kg,
  DROP COLUMN IF EXISTS company_id;

DROP INDEX IF EXISTS service_zones_company_geo_unique;
DROP INDEX IF EXISTS service_zones_company_idx;

ALTER TABLE service_zones
  DROP COLUMN IF EXISTS geo_zone_id,
  DROP COLUMN IF EXISTS company_id;

DROP TABLE IF EXISTS geo_zones;
