ALTER TABLE deliveries
  DROP COLUMN IF EXISTS delivery_speed,
  DROP COLUMN IF EXISTS distance_km;

DROP TABLE IF EXISTS company_booking_routes;
