-- Neighbourhood-level geo zones share an LGA parent. Names are unique within
-- that parent regardless of case so providers converge on one catalog row
-- instead of creating near-duplicates such as "Ariaria" and "ariaria".
CREATE UNIQUE INDEX geo_zones_area_name_unique
  ON geo_zones (parent_id, lower(name))
  WHERE parent_id IS NOT NULL;

-- A route's drop-off hub can be pinned to one exact branch. This is what a
-- customer sees when they search in drop-off mode.
ALTER TABLE pricing_zone_rates
  ADD COLUMN dropoff_branch_id uuid REFERENCES branches (id) ON DELETE SET NULL;

CREATE INDEX pricing_zone_rates_dropoff_branch_idx
  ON pricing_zone_rates (dropoff_branch_id);

-- Quote snapshots keep the zone labels and hub details at search time.
ALTER TABLE marketplace_quotes
  ADD COLUMN origin_zone_name text,
  ADD COLUMN destination_zone_name text,
  ADD COLUMN dropoff_branch_name text,
  ADD COLUMN dropoff_branch_address text;
