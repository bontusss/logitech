DROP INDEX IF EXISTS service_zones_parent_idx;

ALTER TABLE service_zones
  DROP COLUMN IF EXISTS parent_zone_id;
