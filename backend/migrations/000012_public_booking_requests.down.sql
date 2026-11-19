DROP INDEX IF EXISTS deliveries_company_requests_idx;

UPDATE deliveries d
SET status = 'created'
WHERE d.status = 'requested';

UPDATE deliveries d
SET created_by = c.owner_user_id
FROM companies c
WHERE d.company_id = c.id AND d.created_by IS NULL;

ALTER TABLE deliveries
  DROP CONSTRAINT IF EXISTS deliveries_pickup_coordinates_check,
  DROP CONSTRAINT IF EXISTS deliveries_dropoff_coordinates_check,
  DROP COLUMN IF EXISTS package_category,
  DROP COLUMN IF EXISTS weight_kg,
  DROP COLUMN IF EXISTS additional_instructions,
  DROP COLUMN IF EXISTS package_photo_url,
  DROP COLUMN IF EXISTS pickup_latitude,
  DROP COLUMN IF EXISTS pickup_longitude,
  DROP COLUMN IF EXISTS dropoff_latitude,
  DROP COLUMN IF EXISTS dropoff_longitude,
  DROP COLUMN IF EXISTS selected_route_id,
  DROP COLUMN IF EXISTS selected_rate_card_id,
  DROP COLUMN IF EXISTS price_is_estimate,
  DROP COLUMN IF EXISTS booking_source;

ALTER TABLE deliveries
  ALTER COLUMN created_by SET NOT NULL,
  DROP CONSTRAINT deliveries_status_check,
  ADD CONSTRAINT deliveries_status_check
    CHECK (status IN ('created', 'assigned', 'picked_up', 'in_transit', 'delivered', 'cancelled'));
