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
  business_type     text NOT NULL,
  country           text NOT NULL DEFAULT 'NG',
  state             text,
  city              text,
  address           text,
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
  code        citext NOT NULL,
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
  applicable_routes   jsonb NOT NULL DEFAULT '[]'::jsonb,
  applicable_services jsonb NOT NULL DEFAULT '[]'::jsonb,
  applicable_categories jsonb NOT NULL DEFAULT '[]'::jsonb,
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