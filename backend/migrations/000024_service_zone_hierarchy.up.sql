ALTER TABLE service_zones
  ADD COLUMN parent_zone_id uuid REFERENCES service_zones(id) ON DELETE SET NULL;

CREATE INDEX service_zones_parent_idx ON service_zones(parent_zone_id);
