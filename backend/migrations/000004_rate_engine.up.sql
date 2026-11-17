-- ---------------------------------------------------------------------------
-- Location Catalogue (reusable, not company-scoped)
-- ---------------------------------------------------------------------------

-- Countries table - catalogue of supported countries
CREATE TABLE countries (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name        text NOT NULL,
  code        citext NOT NULL UNIQUE,
  numeric_code integer NOT NULL,
  status      text NOT NULL DEFAULT 'active'
              CHECK (status IN ('active', 'retired')),
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX countries_code_active_unique ON countries (code) WHERE deleted_at IS NULL;

CREATE INDEX idx_countries_name ON countries (name);

-- States table - Nigerian states and other African countries
CREATE TABLE states (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name        text NOT NULL,
  code        text,
  country_id  uuid NOT NULL REFERENCES countries (id) ON DELETE CASCADE,
  status      text NOT NULL DEFAULT 'active'
              CHECK (status IN ('active', 'retired')),
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX states_code_country_unique ON states (code, country_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_states_country ON states (country_id);

-- Cities table
CREATE TABLE cities (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name        text NOT NULL,
  state_id    uuid REFERENCES states (id) ON DELETE SET NULL,
  country_id  uuid NOT NULL REFERENCES countries (id) ON DELETE CASCADE,
  status      text NOT NULL DEFAULT 'active',
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX cities_name_state_unique ON cities (name, state_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_cities_state ON cities (state_id);
CREATE INDEX idx_cities_country ON cities (country_id);

-- Areas/Districts
CREATE TABLE areas (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name        text NOT NULL,
  city_id     uuid REFERENCES cities (id) ON DELETE SET NULL,
  status      text NOT NULL DEFAULT 'active',
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX areas_name_city_unique ON areas (name, city_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_areas_city ON areas (city_id);

-- Terminals
CREATE TABLE terminals (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name        text NOT NULL,
  code        citex NOT NULL,
  city_id     uuid REFERENCES cities (id) ON DELETE SET NULL,
  address     text,
  status      text NOT NULL DEFAULT 'active',
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX terminals_code_unique ON terminals (code) WHERE deleted_at IS NULL;
CREATE INDEX idx_terminals_city ON terminals (city_id);

-- ---------------------------------------------------------------------------
-- Company Service Areas (company-scoped)
-- ---------------------------------------------------------------------------

-- Company service areas link companies to locations they serve
CREATE TABLE company_service_areas (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id      uuid NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
  location_type   text NOT NULL CHECK (location_type IN ('country', 'state', 'city', 'area', 'terminal')),
  location_id     uuid NOT NULL,
  serves_pickup   boolean NOT NULL DEFAULT false,
  serves_delivery boolean NOT NULL DEFAULT false,
  serves_terminal boolean NOT NULL DEFAULT false,
  is_primary      boolean NOT NULL DEFAULT false,
  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now(),
  UNIQUE (company_id, location_type, location_id)
);

CREATE INDEX idx_company_service_areas_company ON company_service_areas (company_id);
CREATE INDEX idx_company_service_areas_location ON company_service_areas (location_type, location_id);

-- ---------------------------------------------------------------------------
-- Shipping Services
-- ---------------------------------------------------------------------------

CREATE TABLE shipping_services (
  id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id          uuid NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
  name                text NOT NULL,
  description         text,
  service_code        text NOT NULL,
  delivery_type       text NOT NULL CHECK (delivery_type IN ('intercity', 'terminal-to-terminal', 'door-to-door', 'door-to-terminal', 'terminal-to-door')),
  estimated_duration  integer NOT NULL, -- in days
  min_weight_kg       numeric(10,2) NOT NULL DEFAULT 0,
  max_weight_kg       numeric(10,2),
  max_dimension_cm    numeric(10,2) NOT NULL DEFAULT 0, -- longest side in cm
  volumetric_divisor  numeric(10,2) NOT NULL DEFAULT 5000, -- e.g., 5000 for standard
  status              text NOT NULL DEFAULT 'active'
                      CHECK (status IN ('active', 'inactive')),
  is_default          boolean NOT NULL DEFAULT false,
  created_at          timestamptz NOT NULL DEFAULT now(),
  updated_at          timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX shipping_services_code_company_unique ON shipping_services (service_code, company_id);
CREATE INDEX idx_shipping_services_company ON shipping_services (company_id);
CREATE INDEX idx_shipping_services_status ON shipping_services (status);

-- ---------------------------------------------------------------------------
-- Company Routes
-- ---------------------------------------------------------------------------

-- Directional routes between locations for a company
CREATE TABLE company_routes (
  id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id          uuid NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
  origin_location_type text NOT NULL CHECK (origin_location_type IN ('country', 'state', 'city', 'area', 'terminal')),
  origin_location_id  uuid NOT NULL,
  destination_location_type text NOT NULL CHECK (destination_location_type IN ('country', 'state', 'city', 'area', 'terminal')),
  destination_location_id uuid NOT NULL,
  is_round_trip       boolean NOT NULL DEFAULT false,
  status              text NOT NULL DEFAULT 'active'
                      CHECK (status IN ('active', 'inactive')),
  created_at          timestamptz NOT NULL DEFAULT now(),
  updated_at          timestamptz NOT NULL DEFAULT now(),
  UNIQUE (company_id, origin_location_type, origin_location_id, destination_location_type, destination_location_id)
);

CREATE INDEX idx_company_routes_company ON company_routes (company_id);
CREATE INDEX idx_company_routes_origin ON company_routes (origin_location_type, origin_location_id);
CREATE INDEX idx_company_routes_destination ON company_routes (destination_location_type, destination_location_id);

-- ---------------------------------------------------------------------------
-- Rate Cards
-- ---------------------------------------------------------------------------

CREATE TABLE rate_cards (
  id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id          uuid NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
  name                text NOT NULL,
  description         text,
  currency            text NOT NULL DEFAULT 'NGN'
                      CHECK (currency ~ '^[A-Z]{3}$'),
  status              text NOT NULL DEFAULT 'draft'
                      CHECK (status IN ('draft', 'published', 'expired', 'archived')),
  is_published        boolean NOT NULL DEFAULT false,
  effective_start     timestamptz,
  effective_end       timestamptz,
  published_at        timestamptz,
  created_at          timestamptz NOT NULL DEFAULT now(),
  updated_at          timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_rate_cards_company ON rate_cards (company_id);
CREATE INDEX idx_rate_cards_status ON rate_cards (status);
CREATE INDEX idx_rate_cards_published ON rate_cards (is_published, effective_start, effective_end);

-- ---------------------------------------------------------------------------
-- Rate Card Versions (historical snapshots)
-- ---------------------------------------------------------------------------

CREATE TABLE rate_card_versions (
  id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  rate_card_id        uuid NOT NULL REFERENCES rate_cards (id) ON DELETE CASCADE,
  version             integer NOT NULL,
  pricing_json        jsonb NOT NULL DEFAULT '{}'::jsonb,
  additional_charges_json jsonb NOT NULL DEFAULT '{}'::jsonb,
  is_active           boolean NOT NULL DEFAULT false,
  effective_from      timestamptz NOT NULL DEFAULT now(),
  effective_until     timestamptz,
  created_at          timestamptz NOT NULL DEFAULT now(),
  created_by          uuid REFERENCES users (id)
);

CREATE UNIQUE INDEX rate_card_versions_unique_version ON rate_card_versions (rate_card_id, version);
CREATE INDEX idx_rate_card_versions_card ON rate_card_versions (rate_card_id);
CREATE INDEX idx_rate_card_versions_active ON rate_card_versions (rate_card_id, is_active);

-- ---------------------------------------------------------------------------
-- Pricing Rules
-- ---------------------------------------------------------------------------

-- Base pricing rule types: fixed, weight_based, per_kg, volumetric, category
CREATE TABLE pricing_rules (
  id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  rate_card_version_id uuid NOT NULL REFERENCES rate_card_versions (id) ON DELETE CASCADE,
  rule_type           text NOT NULL CHECK (rule_type IN ('fixed', 'weight_based', 'per_kg', 'volumetric', 'category')),
  name                text NOT NULL,
  description         text,

  -- Fixed pricing
  fixed_amount        numeric(12,2),

  -- Weight-based pricing
  weight_tier_id      uuid REFERENCES pricing_tiers (id),

  -- Per-kg pricing
  per_kg_base_charge  numeric(12,2) NOT NULL DEFAULT 0,
  per_kg_rate         numeric(12,2) NOT NULL DEFAULT 0,
  per_kg_minimum      numeric(12,2) NOT NULL DEFAULT 0,
  per_kg_maximum      numeric(12,2),

  -- Volumetric pricing
  volumetric_divisor  numeric(10,2) NOT NULL DEFAULT 5000,

  -- Category pricing
  category_id         uuid REFERENCES package_categories (id),

  -- Universal settings for the rule
  applicable_routes   jsonb NOT NULL DEFAULT '[]'::jsonb, -- array of route condition objects
  applicable_services jsonb NOT NULL DEFAULT '[]'::jsonb, -- array of service condition objects
  applicable_categories jsonb NOT NULL DEFAULT '[]'::jsonb, -- array of category condition objects
  effective_from      timestamptz NOT NULL DEFAULT now(),
  effective_until     timestamptz,
  is_active           boolean NOT NULL DEFAULT true,
  created_at          timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_pricing_rules_version ON pricing_rules (rate_card_version_id);
CREATE INDEX idx_pricing_rules_type ON pricing_rules (rule_type);
CREATE INDEX idx_pricing_rules_active ON pricing_rules (is_active, effective_from, effective_until);

-- ---------------------------------------------------------------------------
-- Pricing Tiers (weight brackets for weight_based pricing)
-- ---------------------------------------------------------------------------

CREATE TABLE pricing_tiers (
  id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  rate_card_version_id uuid NOT NULL REFERENCES rate_card_versions (id) ON DELETE CASCADE,
  min_weight_kg       numeric(10,2) NOT NULL DEFAULT 0,
  max_weight_kg       numeric(10,2) NOT NULL,
  price_cents         integer NOT NULL, -- price in minor units (kobo for NGN)
  name                text NOT NULL,
  description         text,
  created_at          timestamptz NOT NULL DEFAULT now(),
  UNIQUE (rate_card_version_id, min_weight_kg, max_weight_kg)
);

CREATE INDEX idx_pricing_tiers_version ON pricing_tiers (rate_card_version_id);

-- ---------------------------------------------------------------------------
-- Package Categories
-- ---------------------------------------------------------------------------

CREATE TABLE package_categories (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name        text NOT NULL,
  description text,
  sort_order  integer NOT NULL DEFAULT 0,
  is_default  boolean NOT NULL DEFAULT false,
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX package_categories_name_unique ON package_categories (name) WHERE deleted_at IS NULL;

-- Actually let me fix this - should be package_categories not packages

-- ---------------------------------------------------------------------------
-- Additional Charges
-- ---------------------------------------------------------------------------

CREATE TABLE additional_charges (
  id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  rate_card_version_id uuid REFERENCES rate_card_versions (id) ON DELETE SET NULL,
  company_id          uuid REFERENCES companies (id) ON DELETE SET NULL,
  name                text NOT NULL,
  charge_type         text NOT NULL CHECK (charge_type IN ('fixed', 'percentage')),
  amount_cents        integer NOT NULL DEFAULT 0, -- fixed amount in minor units
  percentage          numeric(5,2), -- percentage value
  is_mandatory        boolean NOT NULL DEFAULT false,
  applies_to_routes   jsonb NOT NULL DEFAULT '[]'::jsonb,
  applies_to_services jsonb NOT NULL DEFAULT '[]'::jsonb,
  applies_to_categories jsonb NOT NULL DEFAULT '[]'::jsonb,
  is_included_in_price boolean NOT NULL DEFAULT false, -- whether included in displayed base price
  created_at          timestamptz NOT NULL DEFAULT now(),
  updated_at          timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_additional_charges_version ON additional_charges (rate_card_version_id);
CREATE INDEX idx_additional_charges_company ON additional_charges (company_id);

-- ---------------------------------------------------------------------------
-- Quote Charges (per-quote line items)
-- ---------------------------------------------------------------------------

CREATE TABLE quote_charges (
  id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  shipping_quote_id   uuid NOT NULL REFERENCES shipping_quotes (id) ON DELETE CASCADE,
  charge_id           uuid REFERENCES additional_charges (id) ON DELETE SET NULL,
  name                text NOT NULL,
  charge_type         text NOT NULL CHECK (charge_type IN ('fixed', 'percentage')),
  amount_cents        integer NOT NULL DEFAULT 0,
  percentage          numeric(5,2),
  is_mandatory        boolean NOT NULL DEFAULT false,
  applied_on          text NOT NULL DEFAULT 'base', -- 'base', 'weight', 'dimensions', etc.
  created_at          timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_quote_charges_quote ON quote_charges (shipping_quote_id);

-- ---------------------------------------------------------------------------
-- Shipping Quotes
-- ---------------------------------------------------------------------------

CREATE TABLE shipping_quotes (
  id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id          uuid NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
  customer_id         uuid REFERENCES users (id) ON DELETE SET NULL,
  origin_location_type text NOT NULL CHECK (origin_location_type IN ('country', 'state', 'city', 'area', 'terminal')),
  origin_location_id  uuid NOT NULL,
  destination_location_type text NOT NULL CHECK (destination_location_type IN ('country', 'state', 'city', 'area', 'terminal')),
  destination_location_id uuid NOT NULL,
  selected_service_id uuid NOT NULL REFERENCES shipping_services (id),
  rate_card_version_id uuid REFERENCES rate_card_versions (id),
  quote_reference     text NOT NULL UNIQUE,
  reference_weight_kg numeric(10,2) NOT NULL,
  reference_length_cm numeric(10,2) NOT NULL,
  reference_width_cm  numeric(10,2) NOT NULL,
  reference_height_cm numeric(10,2) NOT NULL,
  reference_category  text NOT NULL DEFAULT 'parcel', -- document, small_parcel, medium_parcel, large_parcel, fragile, electronics, bulk
  chargeable_weight_kg numeric(10,2) NOT NULL,
  base_price_cents    integer NOT NULL, -- base price in minor units
  total_price_cents   integer NOT NULL, -- total price including all charges in minor units
  currency            text NOT NULL DEFAULT 'NGN',
  estimated_duration  integer NOT NULL, -- in days
  status              text NOT NULL DEFAULT 'valid'
                      CHECK (status IN ('valid', 'expired', 'accepted', 'cancelled')),
  expires_at          timestamptz NOT NULL,
  created_at          timestamptz NOT NULL DEFAULT now(),
  company_confirmed   boolean NOT NULL DEFAULT false,
  company_confirm_at  timestamptz
);

CREATE INDEX idx_shipping_quotes_company ON shipping_quotes (company_id);
CREATE INDEX idx_shipping_quotes_customer ON shipping_quotes (customer_id);
CREATE INDEX idx_shipping_quotes_reference ON shipping_quotes (quote_reference);
CREATE INDEX idx_shipping_quotes_status ON shipping_quotes (status);
CREATE INDEX idx_shipping_quotes_expires ON shipping_quotes (expires_at);
CREATE INDEX idx_shipping_quotes_service ON shipping_quotes (selected_service_id);

-- ---------------------------------------------------------------------------
-- Rate Card Publishing History
-- ---------------------------------------------------------------------------

CREATE TABLE rate_card_history (
  id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  rate_card_id        uuid NOT NULL REFERENCES rate_cards (id) ON DELETE CASCADE,
  version_id          uuid NOT NULL REFERENCES rate_card_versions (id) ON DELETE CASCADE,
  changed_by          uuid REFERENCES users (id),
  changed_at          timestamptz NOT NULL DEFAULT now(),
  previous_status     text NOT NULL DEFAULT 'draft',
  new_status          text NOT NULL,
  notes               text,
  ip                 text,
  user_agent          text
);

CREATE INDEX idx_rate_card_history_card ON rate_card_history (rate_card_id);
CREATE INDEX idx_rate_card_history_version ON rate_card_history (version_id);

-- ---------------------------------------------------------------------------
-- Audit trigger for rate card versions
-- ---------------------------------------------------------------------------

CREATE OR REPLACE FUNCTION handle_rate_card_version_updated_at()
RETURNS trigger AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER rate_card_version_set_updated_at
BEFORE UPDATE ON rate_card_versions
FOR EACH ROW EXECUTE FUNCTION handle_rate_card_version_updated_at();

-- ---------------------------------------------------------------------------
-- Nigerian location catalogue seed data
-- ---------------------------------------------------------------------------

-- Countries (at least Nigeria and some neighbors)
INSERT INTO countries (name, code, numeric_code, status) VALUES ('Nigeria', 'NG', 566, 'active');
INSERT INTO countries (name, code, numeric_code, status) VALUES ('Ghana', 'GH', 288, 'active');
INSERT INTO countries (name, code, numeric_code, status) VALUES ('Kenya', 'KE', 404, 'active');
INSERT INTO countries (name, code, numeric_code, status) VALUES ('South Africa', 'ZA', 710, 'active');
INSERT INTO countries (name, code, numeric_code, status) VALUES ('Egypt', 'EG', 818, 'active');

-- Nigerian states
INSERT INTO states (name, code, country_id, status) VALUES ('Lagos', 'LG', (SELECT id FROM countries WHERE code = 'NG'), 'active');
INSERT INTO states (name, code, country_id, status) VALUES ('Abuja', 'AB', (SELECT id FROM countries WHERE code = 'NG'), 'active');
INSERT INTO states (name, code, country_id, status) VALUES ('Kano', 'KN', (SELECT id FROM countries WHERE code = 'NG'), 'active');
INSERT INTO states (name, code, country_id, status) VALUES ('Oyo', 'OY', (SELECT id FROM countries WHERE code = 'NG'), 'active');
INSERT INTO states (name, code, country_id, status) VALUES ('Rivers', 'RI', (SELECT id FROM countries WHERE code = 'NG'), 'active');
INSERT INTO states (name, code, country_id, status) VALUES ('Kaduna', 'KD', (SELECT id FROM countries WHERE code = 'NG'), 'active');
INSERT INTO states (name, code, country_id, status) VALUES ('Anambra', 'AN', (SELECT id FROM countries WHERE code = 'NG'), 'active');
INSERT INTO states (name, code, country_id, status) VALUES ('Kano', 'KN', (SELECT id FROM countries WHERE code = 'NG'), 'active');
INSERT INTO states (name, code, country_id, status) VALUES ('Delta', 'DT', (SELECT id FROM countries WHERE code = 'NG'), 'active');
INSERT INTO states (name, code, country_id, status) VALUES ('Ondo', 'ON', (SELECT id FROM countries WHERE code = 'NG'), 'active');

-- Cities
INSERT INTO cities (name, state_id, country_id, status) VALUES ('Lagos', (SELECT id FROM states WHERE code = 'LG'), (SELECT id FROM countries WHERE code = 'NG'), 'active');
INSERT INTO cities (name, state_id, country_id, status) VALUES ('Abuja', (SELECT id FROM states WHERE code = 'AB'), (SELECT id FROM countries WHERE code = 'NG'), 'active');
INSERT INTO cities (name, state_id, country_id, status) VALUES ('Kano', (SELECT id FROM states WHERE code = 'KN'), (SELECT id FROM countries WHERE code = 'NG'), 'active');
INSERT INTO cities (name, state_id, country_id, status) VALUES ('Ibadan', (SELECT id FROM states WHERE code = 'OY'), (SELECT id FROM countries WHERE code = 'NG'), 'active');
INSERT INTO cities (name, state_id, country_id, status) VALUES ('Port Harcourt', (SELECT id FROM states WHERE code = 'RI'), (SELECT id FROM countries WHERE code = 'NG'), 'active');
INSERT INTO cities (name, state_id, country_id, status) VALUES ('Kaduna', (SELECT id FROM states WHERE code = 'KD'), (SELECT id FROM countries WHERE code = 'NG'), 'active');
INSERT INTO cities (name, state_id, country_id, status) VALUES ('Onitsha', (SELECT id FROM states WHERE code = 'AN'), (SELECT id FROM countries WHERE code = 'NG'), 'active');
INSERT INTO cities (name, state_id, country_id, status) VALUES ('Warri', (SELECT id FROM states WHERE code = 'DT'), (SELECT id FROM countries WHERE code = 'NG'), 'active');
INSERT INTO cities (name, state_id, country_id, status) VALUES ('Benin City', (SELECT id FROM states WHERE code = 'DT'), (SELECT id FROM countries WHERE code = 'NG'), 'active');
INSERT INTO cities (name, state_id, country_id, status) VALUES ('Aba', (SELECT id FROM states WHERE code = 'AN'), (SELECT id FROM countries WHERE code = 'NG'), 'active');
INSERT INTO cities (name, state_id, country_id, status) VALUES ('Owerri', (SELECT id FROM states WHERE code = 'AN'), (SELECT id FROM countries WHERE code = 'NG'), 'active');
INSERT INTO cities (name, state_id, country_id, status) VALUES ('Sapele', (SELECT id FROM states WHERE code = 'DT'), (SELECT id FROM countries WHERE code = 'NG'), 'active');
