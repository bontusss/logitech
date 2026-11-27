-- Rider portal links use the vehicle plate (letters and digits only) instead of
-- a long signed token. Enforce uniqueness across providers so /rider/<plate>
-- always resolves to exactly one active driver.
CREATE UNIQUE INDEX driver_profiles_rider_slug_unique
  ON driver_profiles (upper(regexp_replace(vehicle_plate, '[^A-Za-z0-9]', '', 'g')));
