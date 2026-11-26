ALTER TABLE marketplace_quotes
  DROP COLUMN IF EXISTS dropoff_branch_address,
  DROP COLUMN IF EXISTS dropoff_branch_name,
  DROP COLUMN IF EXISTS destination_zone_name,
  DROP COLUMN IF EXISTS origin_zone_name;

DROP INDEX IF EXISTS pricing_zone_rates_dropoff_branch_idx;

ALTER TABLE pricing_zone_rates
  DROP COLUMN IF EXISTS dropoff_branch_id;

DROP INDEX IF EXISTS geo_zones_area_name_unique;
