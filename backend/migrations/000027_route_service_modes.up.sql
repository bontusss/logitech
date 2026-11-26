-- Directional zone routes now declare how the parcel enters the network and
-- how long each leg takes. Providers either pick the parcel up at the sender's
-- address or accept a drop-off at the hub for that route's origin zone.
ALTER TABLE pricing_zone_rates
  ADD COLUMN pickup_available boolean NOT NULL DEFAULT true,
  ADD COLUMN dropoff_available boolean NOT NULL DEFAULT true,
  ADD COLUMN estimated_pickup_hours numeric(6,2),
  ADD COLUMN estimated_delivery_hours numeric(6,2),
  ADD CONSTRAINT pricing_zone_rates_pickup_hours_check
    CHECK (estimated_pickup_hours IS NULL OR estimated_pickup_hours >= 0),
  ADD CONSTRAINT pricing_zone_rates_delivery_hours_check
    CHECK (estimated_delivery_hours IS NULL OR estimated_delivery_hours >= 0);

-- Quotes remember which service mode produced them and the provider's
-- configured pickup/delivery windows so a later route edit cannot rewrite them.
ALTER TABLE marketplace_quotes
  ADD COLUMN service_mode text NOT NULL DEFAULT 'pickup'
    CHECK (service_mode IN ('pickup', 'dropoff')),
  ADD COLUMN estimated_pickup text,
  ADD COLUMN estimated_pickup_hours numeric(6,2),
  ADD COLUMN estimated_delivery_hours numeric(6,2);
