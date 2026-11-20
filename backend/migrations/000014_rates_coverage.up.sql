ALTER TABLE company_booking_routes
  DROP CONSTRAINT IF EXISTS company_booking_routes_check;

ALTER TABLE company_booking_routes
  ADD COLUMN transportation_method text NOT NULL DEFAULT 'road',
  ADD COLUMN pricing_model text NOT NULL DEFAULT 'fixed',
  ADD COLUMN origin_branch_id uuid REFERENCES branches(id) ON DELETE SET NULL,
  ADD COLUMN destination_branch_id uuid REFERENCES branches(id) ON DELETE SET NULL,
  ADD COLUMN estimated_transit_time text,
  ADD COLUMN departure_schedule text,
  ADD COLUMN pickup_available boolean NOT NULL DEFAULT true,
  ADD COLUMN delivery_available boolean NOT NULL DEFAULT true,
  ADD COLUMN description text,
  ADD COLUMN max_weight_kg numeric(10,2),
  ADD COLUMN max_length_cm numeric(10,2),
  ADD COLUMN max_width_cm numeric(10,2),
  ADD COLUMN max_height_cm numeric(10,2),
  ADD COLUMN price_is_estimate boolean NOT NULL DEFAULT true,
  ADD CONSTRAINT company_booking_routes_pricing_model_check
    CHECK (pricing_model IN ('fixed','weight','size','volumetric','manual')),
  ADD CONSTRAINT company_booking_routes_weight_check CHECK (max_weight_kg IS NULL OR max_weight_kg >= 0);

CREATE TABLE pricing_weight_brackets (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  route_id uuid NOT NULL REFERENCES company_booking_routes(id) ON DELETE CASCADE,
  min_weight_kg numeric(10,2) NOT NULL CHECK (min_weight_kg >= 0),
  max_weight_kg numeric(10,2) NOT NULL CHECK (max_weight_kg >= min_weight_kg),
  price_minor bigint NOT NULL CHECK (price_minor >= 0),
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX pricing_weight_brackets_route_idx ON pricing_weight_brackets(route_id, min_weight_kg);

CREATE TABLE pricing_size_categories (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  route_id uuid NOT NULL REFERENCES company_booking_routes(id) ON DELETE CASCADE,
  name text NOT NULL,
  price_minor bigint,
  requires_manual_quote boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now(),
  CHECK (price_minor IS NULL OR price_minor >= 0),
  UNIQUE (route_id, name)
);

CREATE TABLE last_mile_service_areas (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  name text NOT NULL,
  state text,
  city text,
  coverage_status text NOT NULL DEFAULT 'draft' CHECK (coverage_status IN ('draft','active','inactive')),
  pricing_model text NOT NULL DEFAULT 'flat' CHECK (pricing_model IN ('flat','zone','distance','manual')),
  flat_price_minor bigint,
  base_delivery_fee_minor bigint,
  price_per_km_minor bigint,
  minimum_fee_minor bigint,
  maximum_distance_km numeric(10,2),
  maximum_fee_minor bigint,
  volumetric_divisor numeric(10,2),
  price_is_estimate boolean NOT NULL DEFAULT true,
  max_weight_kg numeric(10,2),
  max_length_cm numeric(10,2),
  max_width_cm numeric(10,2),
  max_height_cm numeric(10,2),
  express_available boolean NOT NULL DEFAULT false,
  pickup_available boolean NOT NULL DEFAULT true,
  delivery_available boolean NOT NULL DEFAULT true,
  operating_hours jsonb NOT NULL DEFAULT '{}'::jsonb,
  supported_parcel_types text[] NOT NULL DEFAULT '{}',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX last_mile_service_areas_company_status_idx ON last_mile_service_areas(company_id, coverage_status);
CREATE TRIGGER last_mile_service_areas_set_updated_at BEFORE UPDATE ON last_mile_service_areas
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE service_zones (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  area_id uuid NOT NULL REFERENCES last_mile_service_areas(id) ON DELETE CASCADE,
  name text NOT NULL,
  neighbourhoods text[] NOT NULL DEFAULT '{}',
  state text,
  city text,
  lga text,
  custom_location text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (area_id, name)
);
CREATE TRIGGER service_zones_set_updated_at BEFORE UPDATE ON service_zones
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE pricing_zone_rates (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  area_id uuid NOT NULL REFERENCES last_mile_service_areas(id) ON DELETE CASCADE,
  from_zone_id uuid NOT NULL REFERENCES service_zones(id) ON DELETE CASCADE,
  to_zone_id uuid NOT NULL REFERENCES service_zones(id) ON DELETE CASCADE,
  price_minor bigint NOT NULL CHECK (price_minor >= 0),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (area_id, from_zone_id, to_zone_id)
);
CREATE INDEX pricing_zone_rates_area_idx ON pricing_zone_rates(area_id);
CREATE TRIGGER pricing_zone_rates_set_updated_at BEFORE UPDATE ON pricing_zone_rates
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE pricing_quote_versions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  service_type text NOT NULL CHECK (service_type IN ('intercity','lastmile')),
  route_id uuid REFERENCES company_booking_routes(id) ON DELETE SET NULL,
  area_id uuid REFERENCES last_mile_service_areas(id) ON DELETE SET NULL,
  pricing_model text NOT NULL,
  input_snapshot jsonb NOT NULL,
  calculation_snapshot jsonb NOT NULL,
  base_price_minor bigint,
  additional_charges_minor bigint NOT NULL DEFAULT 0,
  total_minor bigint,
  currency text NOT NULL DEFAULT 'NGN',
  is_estimate boolean NOT NULL DEFAULT true,
  physical_inspection_required boolean NOT NULL DEFAULT false,
  confirmation_required boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX pricing_quote_versions_company_created_idx ON pricing_quote_versions(company_id, created_at DESC);
