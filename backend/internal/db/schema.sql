CREATE EXTENSION IF NOT EXISTS citext;

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE users (
  id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  email             citext NOT NULL,
  password_hash     text NOT NULL,
  full_name         text NOT NULL,
  phone             text,
  account_type      text NOT NULL DEFAULT 'customer'
                    CHECK (account_type IN ('platform', 'company', 'customer')),
  platform_role     text
                    CHECK (platform_role IN ('super_admin', 'admin', 'support')),
  email_verified_at timestamptz,
  status            text NOT NULL DEFAULT 'active'
                    CHECK (status IN ('active', 'suspended')),
  last_login_at     timestamptz,
  created_at        timestamptz NOT NULL DEFAULT now(),
  updated_at        timestamptz NOT NULL DEFAULT now(),
  deleted_at        timestamptz,
  CHECK (platform_role IS NULL OR account_type = 'platform')
);

CREATE UNIQUE INDEX users_email_active_unique ON users (email) WHERE deleted_at IS NULL;
CREATE INDEX users_account_type_idx ON users (account_type);

CREATE TRIGGER users_set_updated_at
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE companies (
  id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name              text NOT NULL,
  slug              citext NOT NULL,
  business_email    citext NOT NULL,
  phone             text NOT NULL,
  cac_business_name text,
  country           text NOT NULL DEFAULT 'NG',
  state             text,
  city              text,
  address           text,
  accepted_item_categories text[] NOT NULL DEFAULT ARRAY['cloth', 'electronics', 'document', 'food', 'jewelry', 'other']::text[],
  prohibited_item_categories text[] NOT NULL DEFAULT '{}'::text[],
  status            text NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'under_review', 'verified', 'rejected', 'suspended')),
  onboarding_status text NOT NULL DEFAULT 'incomplete'
                    CHECK (onboarding_status IN ('incomplete', 'pending_review', 'complete')),
  owner_user_id     uuid NOT NULL REFERENCES users (id),
  verified_at       timestamptz,
  created_at        timestamptz NOT NULL DEFAULT now(),
  updated_at        timestamptz NOT NULL DEFAULT now(),
  deleted_at        timestamptz
);

CREATE UNIQUE INDEX companies_slug_active_unique ON companies (slug) WHERE deleted_at IS NULL;
CREATE INDEX companies_owner_idx ON companies (owner_user_id);

CREATE TRIGGER companies_set_updated_at
BEFORE UPDATE ON companies
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE company_members (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id  uuid NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
  user_id     uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  status      text NOT NULL DEFAULT 'active'
              CHECK (status IN ('invited', 'active', 'suspended', 'removed')),
  title       text,
  invited_by  uuid REFERENCES users (id),
  joined_at   timestamptz,
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now(),
  deleted_at  timestamptz
);

CREATE UNIQUE INDEX company_members_company_user_unique
  ON company_members (company_id, user_id);
CREATE INDEX company_members_user_active_idx
  ON company_members (user_id) WHERE deleted_at IS NULL AND status = 'active';

CREATE TRIGGER company_members_set_updated_at
BEFORE UPDATE ON company_members
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE roles (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id  uuid REFERENCES companies (id) ON DELETE CASCADE,
  key         text NOT NULL,
  name        text NOT NULL,
  description text,
  is_system   boolean NOT NULL DEFAULT false,
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX roles_company_key_unique
  ON roles (company_id, key) WHERE company_id IS NOT NULL;
CREATE UNIQUE INDEX roles_system_key_unique
  ON roles (key) WHERE company_id IS NULL;

CREATE TRIGGER roles_set_updated_at
BEFORE UPDATE ON roles
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE permissions (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  key         text NOT NULL UNIQUE,
  description text,
  created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE role_permissions (
  role_id       uuid NOT NULL REFERENCES roles (id) ON DELETE CASCADE,
  permission_id uuid NOT NULL REFERENCES permissions (id) ON DELETE CASCADE,
  PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE member_roles (
  member_id uuid NOT NULL REFERENCES company_members (id) ON DELETE CASCADE,
  role_id   uuid NOT NULL REFERENCES roles (id) ON DELETE CASCADE,
  PRIMARY KEY (member_id, role_id)
);

CREATE TABLE company_settings (
  company_id            uuid PRIMARY KEY REFERENCES companies (id) ON DELETE CASCADE,
  timezone              text NOT NULL DEFAULT 'Africa/Lagos',
  currency              text NOT NULL DEFAULT 'NGN',
  weight_unit           text NOT NULL DEFAULT 'kg'
                        CHECK (weight_unit IN ('kg', 'lb')),
  auto_accept_bookings  boolean NOT NULL DEFAULT false,
  notify_email          boolean NOT NULL DEFAULT true,
  created_at            timestamptz NOT NULL DEFAULT now(),
  updated_at            timestamptz NOT NULL DEFAULT now()
);

CREATE TRIGGER company_settings_set_updated_at
BEFORE UPDATE ON company_settings
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE company_branding (
  company_id      uuid PRIMARY KEY REFERENCES companies (id) ON DELETE CASCADE,
  logo_url        text,
  primary_color   text NOT NULL DEFAULT '#2b4bff',
  secondary_color text NOT NULL DEFAULT '#0a1220',
  favicon_url     text,
  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now()
);

CREATE TRIGGER company_branding_set_updated_at
BEFORE UPDATE ON company_branding
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE email_verifications (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  token_hash  text NOT NULL UNIQUE,
  expires_at  timestamptz NOT NULL,
  consumed_at timestamptz,
  created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX email_verifications_user_idx ON email_verifications (user_id);

CREATE TABLE password_resets (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  token_hash  text NOT NULL UNIQUE,
  expires_at  timestamptz NOT NULL,
  consumed_at timestamptz,
  created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX password_resets_user_idx ON password_resets (user_id);

CREATE TABLE refresh_tokens (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  token_hash  text NOT NULL UNIQUE,
  family_id   uuid NOT NULL,
  user_agent  text,
  ip          text,
  expires_at  timestamptz NOT NULL,
  revoked_at  timestamptz,
  replaced_by uuid,
  created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX refresh_tokens_user_idx ON refresh_tokens (user_id);
CREATE INDEX refresh_tokens_family_idx ON refresh_tokens (family_id);

CREATE TABLE email_jobs (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id       uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  company_id    uuid REFERENCES companies (id) ON DELETE CASCADE,
  kind          text NOT NULL CHECK (kind IN ('welcome_customer', 'welcome_provider', 'provider_onboarding')),
  sequence      smallint NOT NULL DEFAULT 0,
  status        text NOT NULL DEFAULT 'pending'
                CHECK (status IN ('pending', 'sending', 'sent', 'cancelled', 'failed')),
  scheduled_at  timestamptz NOT NULL,
  claimed_at    timestamptz,
  sent_at       timestamptz,
  attempts      smallint NOT NULL DEFAULT 0,
  last_error    text,
  created_at    timestamptz NOT NULL DEFAULT now(),
  UNIQUE (user_id, kind, sequence)
);

CREATE INDEX email_jobs_ready_idx ON email_jobs (scheduled_at)
  WHERE status IN ('pending', 'sending');

CREATE TABLE user_phone_verifications (
  user_id           uuid PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
  phone             text NOT NULL,
  token_hash        text NOT NULL,
  expires_at        timestamptz NOT NULL,
  verified_at       timestamptz,
  last_sent_at      timestamptz NOT NULL,
  window_started_at timestamptz NOT NULL,
  sends_in_window   smallint NOT NULL DEFAULT 1,
  attempts          smallint NOT NULL DEFAULT 0,
  created_at        timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE company_invitations (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id  uuid NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
  email       citext NOT NULL,
  role_id     uuid NOT NULL REFERENCES roles (id),
  token_hash  text NOT NULL UNIQUE,
  invited_by  uuid REFERENCES users (id),
  expires_at  timestamptz NOT NULL,
  accepted_at timestamptz,
  revoked_at  timestamptz,
  created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX company_invitations_pending_unique
  ON company_invitations (company_id, email)
  WHERE accepted_at IS NULL AND revoked_at IS NULL;

CREATE TABLE audit_logs (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  actor_user_id uuid REFERENCES users (id) ON DELETE SET NULL,
  company_id    uuid REFERENCES companies (id) ON DELETE SET NULL,
  action        text NOT NULL,
  entity_type   text,
  entity_id     uuid,
  metadata      jsonb NOT NULL DEFAULT '{}'::jsonb,
  ip            text,
  user_agent    text,
  created_at    timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX audit_logs_company_created_idx ON audit_logs (company_id, created_at DESC);
CREATE INDEX audit_logs_actor_created_idx ON audit_logs (actor_user_id, created_at DESC);

ALTER TABLE companies
  ADD COLUMN IF NOT EXISTS description text,
  ADD COLUMN IF NOT EXISTS cover_image_url text,
  ADD COLUMN IF NOT EXISTS operating_hours jsonb NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE company_branding
  ADD COLUMN IF NOT EXISTS cover_image_url text,
  ADD COLUMN IF NOT EXISTS public_name text,
  ADD COLUMN IF NOT EXISTS description text,
  ADD COLUMN IF NOT EXISTS pwa_short_name text;

ALTER TABLE company_invitations
  ADD COLUMN IF NOT EXISTS accepted_by_user_id uuid REFERENCES users (id) ON DELETE SET NULL;

CREATE TABLE branches (
  id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id        uuid NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
  name              text NOT NULL,
  code              citext NOT NULL,
  address           text NOT NULL,
  city              text NOT NULL,
  state             text NOT NULL,
  country           text NOT NULL DEFAULT 'NG',
  phone             text,
  email             citext,
  manager_member_id uuid REFERENCES company_members (id) ON DELETE SET NULL,
  operating_hours   jsonb NOT NULL DEFAULT '{}'::jsonb,
  status            text NOT NULL DEFAULT 'active'
                    CHECK (status IN ('active', 'inactive')),
  created_at        timestamptz NOT NULL DEFAULT now(),
  updated_at        timestamptz NOT NULL DEFAULT now(),
  deleted_at        timestamptz
);

CREATE UNIQUE INDEX branches_company_code_unique
  ON branches (company_id, code) WHERE deleted_at IS NULL;
CREATE INDEX branches_company_status_idx
  ON branches (company_id, status) WHERE deleted_at IS NULL;
CREATE INDEX branches_manager_idx ON branches (manager_member_id);

CREATE TRIGGER branches_set_updated_at
BEFORE UPDATE ON branches
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE branch_members (
  branch_id  uuid NOT NULL REFERENCES branches (id) ON DELETE CASCADE,
  member_id  uuid NOT NULL REFERENCES company_members (id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (branch_id, member_id)
);

CREATE INDEX branch_members_member_idx ON branch_members (member_id);

CREATE TABLE invitation_branches (
  invitation_id uuid NOT NULL REFERENCES company_invitations (id) ON DELETE CASCADE,
  branch_id     uuid NOT NULL REFERENCES branches (id) ON DELETE CASCADE,
  PRIMARY KEY (invitation_id, branch_id)
);

CREATE INDEX invitation_branches_branch_idx ON invitation_branches (branch_id);

CREATE TABLE company_subdomains (
  id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id uuid NOT NULL UNIQUE REFERENCES companies (id) ON DELETE CASCADE,
  subdomain  citext NOT NULL UNIQUE
             CHECK (subdomain ~ '^[a-z0-9][a-z0-9-]{1,28}[a-z0-9]$'),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TRIGGER company_subdomains_set_updated_at
BEFORE UPDATE ON company_subdomains
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- Rate Engine Tables (Task 3)

-- Countries table - catalogue of supported countries
CREATE TABLE countries (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name        text NOT NULL,
  code        citext NOT NULL UNIQUE,
  numeric_code integer NOT NULL,
  status      text NOT NULL DEFAULT 'active'
              CHECK (status IN ('active', 'retired')),
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now(),
  deleted_at  timestamptz
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
  updated_at  timestamptz NOT NULL DEFAULT now(),
  deleted_at  timestamptz
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
  updated_at  timestamptz NOT NULL DEFAULT now(),
  deleted_at  timestamptz
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
  updated_at  timestamptz NOT NULL DEFAULT now(),
  deleted_at  timestamptz
);

CREATE UNIQUE INDEX areas_name_city_unique ON areas (name, city_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_areas_city ON areas (city_id);

-- Terminals
CREATE TABLE terminals (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name        text NOT NULL,
  code        citext NOT NULL,
  city_id     uuid REFERENCES cities (id) ON DELETE SET NULL,
  address     text,
  status      text NOT NULL DEFAULT 'active',
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now(),
  deleted_at  timestamptz
);

CREATE UNIQUE INDEX terminals_code_unique ON terminals (code) WHERE deleted_at IS NULL;
CREATE INDEX idx_terminals_city ON terminals (city_id);

-- ---------------------------------------------------------------------------
-- Booking pages

CREATE TABLE booking_pages (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id   uuid NOT NULL UNIQUE REFERENCES companies (id) ON DELETE CASCADE,
  slug         citext NOT NULL UNIQUE
               CHECK (slug ~ '^[a-z0-9]([a-z0-9-]{0,38}[a-z0-9])?$'),
  headline     text,
  description  text,
  is_published boolean NOT NULL DEFAULT true,
  created_at   timestamptz NOT NULL DEFAULT now(),
  updated_at   timestamptz NOT NULL DEFAULT now()
);

CREATE TRIGGER booking_pages_set_updated_at
BEFORE UPDATE ON booking_pages
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE driver_profiles (
  id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id        uuid NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
  full_name         text NOT NULL,
  phone             text NOT NULL,
  vehicle_plate     text NOT NULL,
  email             citext NOT NULL,
  portal_token_hash text NOT NULL UNIQUE,
  status            text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'inactive')),
  created_at        timestamptz NOT NULL DEFAULT now(),
  updated_at        timestamptz NOT NULL DEFAULT now(),
  UNIQUE (company_id, email),
  UNIQUE (company_id, id)
);

CREATE INDEX driver_profiles_company_status_idx ON driver_profiles (company_id, status);

CREATE TRIGGER driver_profiles_set_updated_at
BEFORE UPDATE ON driver_profiles
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE deliveries (
  id                    uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id            uuid NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
  sender_name           text NOT NULL,
  sender_phone          text NOT NULL,
  receiver_name         text NOT NULL,
  receiver_phone        text NOT NULL,
  pickup_address        text NOT NULL,
  dropoff_address       text NOT NULL,
  item_description      text NOT NULL,
  declared_value_minor  bigint NOT NULL CHECK (declared_value_minor >= 0),
  delivery_price_minor  bigint NOT NULL CHECK (delivery_price_minor >= 0),
  payer                 text NOT NULL CHECK (payer IN ('sender', 'receiver')),
  is_paid               boolean NOT NULL DEFAULT false,
  payment_method        text CHECK (payment_method IN ('cash', 'pos', 'bank_transfer')),
  assigned_driver_id    uuid,
  status                text NOT NULL DEFAULT 'created'
                        CHECK (status IN ('created', 'assigned', 'picked_up', 'in_transit', 'delivered', 'cancelled')),
  created_by            uuid NOT NULL REFERENCES users (id),
  created_at            timestamptz NOT NULL DEFAULT now(),
  updated_at            timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (company_id, assigned_driver_id)
    REFERENCES driver_profiles (company_id, id),
  CHECK ((is_paid AND payment_method IS NOT NULL) OR (NOT is_paid AND payment_method IS NULL))
);

CREATE INDEX deliveries_company_created_idx ON deliveries (company_id, created_at DESC);
CREATE INDEX deliveries_company_status_idx ON deliveries (company_id, status);
CREATE INDEX deliveries_assigned_driver_idx ON deliveries (assigned_driver_id) WHERE assigned_driver_id IS NOT NULL;

CREATE TRIGGER deliveries_set_updated_at
BEFORE UPDATE ON deliveries
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

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

ALTER TABLE company_booking_routes
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
  ADD CONSTRAINT company_booking_routes_weight_check CHECK (max_weight_kg IS NULL OR max_weight_kg >= 0);

CREATE TABLE last_mile_service_areas (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), company_id uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  name text NOT NULL, state text, city text, coverage_status text NOT NULL DEFAULT 'draft' CHECK (coverage_status IN ('draft','active','inactive')),
  pricing_model text NOT NULL DEFAULT 'flat' CHECK (pricing_model IN ('flat','zone','distance','manual')),
  flat_price_minor bigint, base_delivery_fee_minor bigint, price_per_km_minor bigint, minimum_fee_minor bigint,
  maximum_distance_km numeric(10,2), maximum_fee_minor bigint, volumetric_divisor numeric(10,2), price_is_estimate boolean NOT NULL DEFAULT true,
  max_weight_kg numeric(10,2), max_length_cm numeric(10,2), max_width_cm numeric(10,2), max_height_cm numeric(10,2),
  express_available boolean NOT NULL DEFAULT false, pickup_available boolean NOT NULL DEFAULT true, delivery_available boolean NOT NULL DEFAULT true,
  operating_hours jsonb NOT NULL DEFAULT '{}'::jsonb, supported_parcel_types text[] NOT NULL DEFAULT '{}',
  created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX last_mile_service_areas_company_status_idx ON last_mile_service_areas(company_id, coverage_status);
CREATE TRIGGER last_mile_service_areas_set_updated_at BEFORE UPDATE ON last_mile_service_areas FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE geo_zones (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), country_code text NOT NULL DEFAULT 'NG',
  state text NOT NULL, lga text NOT NULL, name text NOT NULL,
  parent_id uuid REFERENCES geo_zones(id) ON DELETE SET NULL,
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active','inactive')),
  created_at timestamptz NOT NULL DEFAULT now(), UNIQUE (country_code, state, lga, name)
);
CREATE INDEX geo_zones_state_lga_idx ON geo_zones (lower(state), lower(lga));
CREATE INDEX geo_zones_parent_idx ON geo_zones(parent_id);

CREATE TABLE service_zones (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), company_id uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  area_id uuid REFERENCES last_mile_service_areas(id) ON DELETE CASCADE, geo_zone_id uuid REFERENCES geo_zones(id) ON DELETE SET NULL,
  name text NOT NULL, neighbourhoods text[] NOT NULL DEFAULT '{}', state text, city text, lga text, custom_location text,
  parent_zone_id uuid REFERENCES service_zones(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(), UNIQUE (area_id, name)
);
CREATE INDEX service_zones_company_idx ON service_zones(company_id);
CREATE UNIQUE INDEX service_zones_company_geo_unique ON service_zones(company_id, geo_zone_id) WHERE geo_zone_id IS NOT NULL;
CREATE INDEX service_zones_parent_idx ON service_zones(parent_zone_id);
CREATE TRIGGER service_zones_set_updated_at BEFORE UPDATE ON service_zones FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE pricing_zone_rates (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), company_id uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  area_id uuid REFERENCES last_mile_service_areas(id) ON DELETE CASCADE,
  from_zone_id uuid NOT NULL REFERENCES service_zones(id) ON DELETE CASCADE, to_zone_id uuid NOT NULL REFERENCES service_zones(id) ON DELETE CASCADE,
  price_minor bigint NOT NULL CHECK (price_minor >= 0),
  included_weight_kg numeric(10,2) NOT NULL DEFAULT 0 CHECK (included_weight_kg >= 0),
  additional_weight_fee_minor bigint NOT NULL DEFAULT 0 CHECK (additional_weight_fee_minor >= 0),
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active','inactive')),
  created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (area_id, from_zone_id, to_zone_id)
);
CREATE INDEX pricing_zone_rates_company_idx ON pricing_zone_rates(company_id, status);
CREATE UNIQUE INDEX pricing_zone_rates_company_direction_unique ON pricing_zone_rates(company_id, from_zone_id, to_zone_id);
CREATE INDEX pricing_zone_rates_area_idx ON pricing_zone_rates(area_id);
CREATE TRIGGER pricing_zone_rates_set_updated_at BEFORE UPDATE ON pricing_zone_rates FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE deliveries
  ADD COLUMN request_status text NOT NULL DEFAULT 'accepted'
    CHECK (request_status IN ('pending', 'accepted', 'declined')),
  ADD COLUMN tracking_code text,
  ADD COLUMN rider_link_token text,
  ADD COLUMN declined_at timestamptz,
  ADD COLUMN accepted_at timestamptz;

UPDATE deliveries
SET request_status = CASE WHEN status = 'requested' THEN 'pending' ELSE 'accepted' END,
    tracking_code = CASE
      WHEN status <> 'requested' THEN 'TRK-' || upper(substr(replace(id::text, '-', ''), 1, 12))
      ELSE NULL
    END;

CREATE UNIQUE INDEX deliveries_tracking_code_unique
  ON deliveries (tracking_code) WHERE tracking_code IS NOT NULL;
CREATE INDEX deliveries_company_request_status_idx
  ON deliveries (company_id, request_status, created_at DESC);

CREATE TABLE delivery_timeline_events (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  delivery_id uuid NOT NULL REFERENCES deliveries(id) ON DELETE CASCADE,
  company_id uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  status text NOT NULL,
  note text,
  created_by uuid REFERENCES users(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX delivery_timeline_delivery_created_idx
  ON delivery_timeline_events(delivery_id, created_at DESC);
-- Marketplace: provider listing, quote snapshots, customer selections and
-- search telemetry. Quotes intentionally snapshot every price component so a
-- later rate change cannot rewrite a saved quote.

CREATE TABLE marketplace_settings (
  company_id          uuid PRIMARY KEY REFERENCES companies (id) ON DELETE CASCADE,
  is_listed           boolean NOT NULL DEFAULT false,
  visibility_status   text NOT NULL DEFAULT 'draft'
                      CHECK (visibility_status IN ('draft', 'listed', 'paused')),
  public_description  text,
  published_at        timestamptz,
  created_at          timestamptz NOT NULL DEFAULT now(),
  updated_at          timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX marketplace_settings_listed_idx
  ON marketplace_settings (visibility_status, is_listed);

CREATE TRIGGER marketplace_settings_set_updated_at
BEFORE UPDATE ON marketplace_settings
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE marketplace_search_logs (
  id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  customer_id         uuid REFERENCES users (id) ON DELETE SET NULL,
  origin_state        text NOT NULL,
  origin_city         text NOT NULL,
  destination_state   text NOT NULL,
  destination_city    text NOT NULL,
  search_metadata     jsonb NOT NULL DEFAULT '{}'::jsonb,
  result_count        integer NOT NULL DEFAULT 0,
  failure_count       integer NOT NULL DEFAULT 0,
  created_at          timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX marketplace_search_logs_created_idx
  ON marketplace_search_logs (created_at DESC);
CREATE INDEX marketplace_search_logs_route_idx
  ON marketplace_search_logs (lower(origin_state), lower(origin_city), lower(destination_state), lower(destination_city));

CREATE TABLE marketplace_quotes (
  id                      uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  quote_reference         text NOT NULL UNIQUE,
  search_id               uuid REFERENCES marketplace_search_logs (id) ON DELETE SET NULL,
  company_id              uuid NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
  provider_name           text NOT NULL,
  provider_logo           text NOT NULL DEFAULT '',
  provider_slug           text NOT NULL,
  route_id                uuid REFERENCES company_booking_routes (id) ON DELETE SET NULL,
  origin_state            text NOT NULL,
  origin_city             text NOT NULL,
  destination_state       text NOT NULL,
  destination_city        text NOT NULL,
  service_key             text NOT NULL CHECK (service_key IN ('standard', 'express', 'same_day', 'last_mile', 'zone_route')),
  service_name            text NOT NULL,
  distance_km             numeric(10,2) NOT NULL,
  package_snapshot        jsonb NOT NULL,
  chargeable_weight_kg    numeric(10,2) NOT NULL,
  base_price_minor        bigint NOT NULL CHECK (base_price_minor >= 0),
  additional_charges_minor bigint NOT NULL DEFAULT 0 CHECK (additional_charges_minor >= 0),
  discount_minor          bigint NOT NULL DEFAULT 0 CHECK (discount_minor >= 0),
  total_price_minor       bigint NOT NULL CHECK (total_price_minor >= 0),
  currency                text NOT NULL DEFAULT 'NGN',
  price_confirmation_type text NOT NULL CHECK (price_confirmation_type IN ('confirmed', 'estimated')),
  price_breakdown         jsonb NOT NULL DEFAULT '[]'::jsonb,
  rate_snapshot           jsonb NOT NULL,
  estimated_delivery      text,
  estimated_duration_days integer,
  pickup_available        boolean NOT NULL DEFAULT false,
  delivery_available      boolean NOT NULL DEFAULT false,
  delivery_method         text NOT NULL DEFAULT 'terminal-to-terminal',
  conditions              text,
  status                  text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'selected', 'expired')),
  expires_at              timestamptz NOT NULL,
  created_at              timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX marketplace_quotes_company_idx ON marketplace_quotes (company_id);
CREATE INDEX marketplace_quotes_search_idx ON marketplace_quotes (search_id);
CREATE INDEX marketplace_quotes_expiry_idx ON marketplace_quotes (expires_at);
CREATE INDEX marketplace_quotes_service_idx ON marketplace_quotes (company_id, service_key);

CREATE TABLE customer_saved_quotes (
  id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  customer_id       uuid REFERENCES users (id) ON DELETE CASCADE,
  guest_token_hash  text,
  quote_id          uuid NOT NULL REFERENCES marketplace_quotes (id) ON DELETE CASCADE,
  selection_status  text NOT NULL DEFAULT 'selected'
                    CHECK (selection_status IN ('selected', 'expired', 'converted', 'discarded')),
  created_at        timestamptz NOT NULL DEFAULT now(),
  updated_at        timestamptz NOT NULL DEFAULT now(),
  CHECK (customer_id IS NOT NULL OR guest_token_hash IS NOT NULL)
);

CREATE UNIQUE INDEX customer_saved_quotes_customer_unique
  ON customer_saved_quotes (customer_id, quote_id) WHERE customer_id IS NOT NULL;
CREATE UNIQUE INDEX customer_saved_quotes_guest_unique
  ON customer_saved_quotes (guest_token_hash, quote_id) WHERE guest_token_hash IS NOT NULL;
CREATE INDEX customer_saved_quotes_customer_created_idx
  ON customer_saved_quotes (customer_id, created_at DESC);

CREATE TRIGGER customer_saved_quotes_set_updated_at
BEFORE UPDATE ON customer_saved_quotes
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Customer self-service bookings: a booking made on a provider's public page
-- can be linked to a customer account, and every booking-page request gets a
-- tracking code so guests can copy a working tracking URL immediately.
ALTER TABLE deliveries
  ADD COLUMN customer_id uuid REFERENCES users (id) ON DELETE SET NULL;

CREATE INDEX deliveries_customer_created_idx
  ON deliveries (customer_id, created_at DESC)
  WHERE customer_id IS NOT NULL;

-- Public provider profile details shown on the booking page, and customer
-- reviews. Reviews are published immediately and tied to the reviewer's
-- account when they are signed in.
ALTER TABLE companies
  ADD COLUMN website_url text,
  ADD COLUMN instagram_url text,
  ADD COLUMN facebook_url text,
  ADD COLUMN twitter_url text,
  ADD COLUMN linkedin_url text,
  ADD COLUMN year_established integer
    CHECK (year_established IS NULL OR year_established BETWEEN 1800 AND 2200),
  ADD COLUMN photo_urls text[] NOT NULL DEFAULT '{}';

CREATE TABLE company_reviews (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id    uuid NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
  customer_id   uuid REFERENCES users (id) ON DELETE SET NULL,
  reviewer_name text NOT NULL,
  rating        smallint NOT NULL CHECK (rating BETWEEN 1 AND 5),
  comment       text NOT NULL DEFAULT '',
  created_at    timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX company_reviews_company_created_idx
  ON company_reviews (company_id, created_at DESC);

-- Driver GPS broadcasts. One row per driver, updated only while the driver has
-- an active assigned delivery and is sharing their location from the portal.
CREATE TABLE driver_locations (
  driver_id       uuid PRIMARY KEY,
  company_id      uuid NOT NULL,
  latitude        double precision NOT NULL CHECK (latitude BETWEEN -90 AND 90),
  longitude       double precision NOT NULL CHECK (longitude BETWEEN -180 AND 180),
  accuracy_m      double precision CHECK (accuracy_m IS NULL OR accuracy_m >= 0),
  heading         double precision CHECK (heading IS NULL OR (heading >= 0 AND heading < 360)),
  speed_kph       double precision CHECK (speed_kph IS NULL OR speed_kph >= 0),
  is_broadcasting boolean NOT NULL DEFAULT true,
  updated_at      timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (company_id, driver_id)
    REFERENCES driver_profiles (company_id, id) ON DELETE CASCADE
);

CREATE INDEX driver_locations_company_updated_idx
  ON driver_locations (company_id, updated_at DESC);

-- Google OAuth identity link. Only customer and company accounts may link a
-- Google identity; platform administrator accounts are rejected in the API.
ALTER TABLE users
  ADD COLUMN google_subject text;

CREATE UNIQUE INDEX users_google_subject_unique
  ON users (google_subject)
  WHERE google_subject IS NOT NULL AND deleted_at IS NULL;
