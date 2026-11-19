ALTER TABLE deliveries
  ALTER COLUMN created_by DROP NOT NULL,
  DROP CONSTRAINT deliveries_status_check,
  ADD CONSTRAINT deliveries_status_check
    CHECK (status IN ('requested', 'created', 'assigned', 'picked_up', 'in_transit', 'delivered', 'cancelled')),
  ADD COLUMN package_category text NOT NULL DEFAULT 'other'
    CHECK (package_category IN ('cloth', 'electronics', 'document', 'food', 'jewelry', 'other')),
  ADD COLUMN weight_kg numeric(10,2) NOT NULL DEFAULT 0 CHECK (weight_kg >= 0),
  ADD COLUMN additional_instructions text NOT NULL DEFAULT '',
  ADD COLUMN package_photo_url text,
  ADD COLUMN pickup_latitude numeric(10,7),
  ADD COLUMN pickup_longitude numeric(10,7),
  ADD COLUMN dropoff_latitude numeric(10,7),
  ADD COLUMN dropoff_longitude numeric(10,7),
  ADD COLUMN selected_route_id uuid,
  ADD COLUMN selected_rate_card_id uuid REFERENCES rate_cards (id) ON DELETE SET NULL,
  ADD COLUMN price_is_estimate boolean NOT NULL DEFAULT true,
  ADD COLUMN booking_source text NOT NULL DEFAULT 'dashboard'
    CHECK (booking_source IN ('dashboard', 'booking_page'));

ALTER TABLE deliveries
  ADD CONSTRAINT deliveries_pickup_coordinates_check
    CHECK ((pickup_latitude IS NULL AND pickup_longitude IS NULL)
      OR (pickup_latitude BETWEEN -90 AND 90 AND pickup_longitude BETWEEN -180 AND 180)),
  ADD CONSTRAINT deliveries_dropoff_coordinates_check
    CHECK ((dropoff_latitude IS NULL AND dropoff_longitude IS NULL)
      OR (dropoff_latitude BETWEEN -90 AND 90 AND dropoff_longitude BETWEEN -180 AND 180));

UPDATE rate_cards
SET is_published = true,
    published_at = COALESCE(published_at, now())
WHERE status = 'published'
  AND NOT is_published;

CREATE INDEX deliveries_company_requests_idx ON deliveries (company_id, created_at DESC)
  WHERE status = 'requested';
