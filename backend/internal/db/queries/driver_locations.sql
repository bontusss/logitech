-- name: ListDriversOnRoad :many
SELECT
  d.id AS delivery_id,
  d.status AS delivery_status,
  d.tracking_code,
  d.pickup_address,
  d.dropoff_address,
  driver.id AS driver_id,
  driver.full_name,
  driver.vehicle_plate,
  loc.latitude,
  loc.longitude,
  loc.updated_at AS location_updated_at
FROM deliveries d
JOIN driver_profiles driver
  ON driver.id = d.assigned_driver_id
 AND driver.company_id = d.company_id
LEFT JOIN driver_locations loc
  ON loc.driver_id = driver.id
 AND loc.company_id = d.company_id
 AND loc.is_broadcasting
 AND loc.updated_at > now() - interval '5 minutes'
WHERE d.company_id = $1
  AND d.status IN ('assigned', 'picked_up', 'in_transit')
ORDER BY d.updated_at DESC;

-- name: GetDriverActiveAssignment :one
SELECT id, status, tracking_code, pickup_address, dropoff_address, updated_at
FROM deliveries
WHERE company_id = $1
  AND assigned_driver_id = $2
  AND status IN ('assigned', 'picked_up', 'in_transit')
ORDER BY updated_at DESC
LIMIT 1;

-- name: UpsertDriverLocation :one
INSERT INTO driver_locations (
  driver_id, company_id, latitude, longitude, accuracy_m, heading, speed_kph, is_broadcasting
) VALUES ($1, $2, $3, $4, $5, $6, $7, true)
ON CONFLICT (driver_id) DO UPDATE SET
  company_id = EXCLUDED.company_id,
  latitude = EXCLUDED.latitude,
  longitude = EXCLUDED.longitude,
  accuracy_m = EXCLUDED.accuracy_m,
  heading = EXCLUDED.heading,
  speed_kph = EXCLUDED.speed_kph,
  is_broadcasting = true,
  updated_at = now()
RETURNING *;

-- name: StopDriverLocation :exec
UPDATE driver_locations
SET is_broadcasting = false, updated_at = now()
WHERE driver_id = $1 AND company_id = $2;
