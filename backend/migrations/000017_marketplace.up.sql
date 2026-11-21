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
  service_key             text NOT NULL CHECK (service_key IN ('standard', 'express', 'same_day')),
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
