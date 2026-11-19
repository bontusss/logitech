CREATE TABLE company_booking_routes (
  id                          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id                  uuid NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
  origin_state                text NOT NULL,
  origin_city                 text NOT NULL,
  destination_state           text NOT NULL,
  destination_city            text NOT NULL,
  distance_km                 numeric(10,2) NOT NULL CHECK (distance_km > 0),
  standard_price_per_km_minor bigint CHECK (standard_price_per_km_minor >= 0),
  express_price_per_km_minor  bigint CHECK (express_price_per_km_minor >= 0),
  same_day_price_per_km_minor bigint CHECK (same_day_price_per_km_minor >= 0),
  minimum_charge_minor        bigint NOT NULL DEFAULT 0 CHECK (minimum_charge_minor >= 0),
  status                      text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
  created_at                  timestamptz NOT NULL DEFAULT now(),
  updated_at                  timestamptz NOT NULL DEFAULT now(),
  CHECK (
    standard_price_per_km_minor IS NOT NULL
    OR express_price_per_km_minor IS NOT NULL
    OR same_day_price_per_km_minor IS NOT NULL
  )
);

CREATE UNIQUE INDEX company_booking_routes_lane_unique
  ON company_booking_routes (company_id, lower(origin_state), lower(origin_city), lower(destination_state), lower(destination_city));
CREATE INDEX company_booking_routes_company_status_idx
  ON company_booking_routes (company_id, status);

CREATE TRIGGER company_booking_routes_set_updated_at
BEFORE UPDATE ON company_booking_routes
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE deliveries
  ADD COLUMN delivery_speed text NOT NULL DEFAULT 'standard'
    CHECK (delivery_speed IN ('standard', 'express', 'same_day')),
  ADD COLUMN distance_km numeric(10,2) NOT NULL DEFAULT 0 CHECK (distance_km >= 0);
