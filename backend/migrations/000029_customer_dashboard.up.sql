-- Customer-owned saved addresses. A customer can keep several pickup and
-- delivery places and mark one of each as the default.
CREATE TABLE customer_addresses (
  id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  customer_id       uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  label             text NOT NULL,
  address_type      text NOT NULL DEFAULT 'residential'
                    CHECK (address_type IN ('residential', 'business', 'warehouse', 'other')),
  contact_name      text NOT NULL DEFAULT '',
  phone             text NOT NULL DEFAULT '',
  state             text NOT NULL,
  city              text NOT NULL,
  area              text NOT NULL DEFAULT '',
  street_address    text NOT NULL,
  directions        text NOT NULL DEFAULT '',
  latitude          numeric(10, 7),
  longitude         numeric(10, 7),
  default_pickup    boolean NOT NULL DEFAULT false,
  default_delivery  boolean NOT NULL DEFAULT false,
  created_at        timestamptz NOT NULL DEFAULT now(),
  updated_at        timestamptz NOT NULL DEFAULT now(),
  CHECK ((latitude IS NULL AND longitude IS NULL)
    OR (latitude BETWEEN -90 AND 90 AND longitude BETWEEN -180 AND 180))
);

CREATE INDEX customer_addresses_customer_idx
  ON customer_addresses (customer_id, created_at DESC);
CREATE UNIQUE INDEX customer_addresses_default_pickup_unique
  ON customer_addresses (customer_id) WHERE default_pickup;
CREATE UNIQUE INDEX customer_addresses_default_delivery_unique
  ON customer_addresses (customer_id) WHERE default_delivery;

CREATE TRIGGER customer_addresses_set_updated_at
BEFORE UPDATE ON customer_addresses
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Providers a customer wants to find again quickly.
CREATE TABLE customer_favourite_providers (
  customer_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  company_id  uuid NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
  created_at  timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (customer_id, company_id)
);

CREATE INDEX customer_favourite_providers_company_idx
  ON customer_favourite_providers (company_id);

-- Customer reports for delayed, missing or damaged parcels.
CREATE TABLE delivery_issues (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  delivery_id uuid NOT NULL REFERENCES deliveries (id) ON DELETE CASCADE,
  customer_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  issue_type  text NOT NULL CHECK (issue_type IN ('delayed', 'missing', 'damaged', 'other')),
  description text NOT NULL DEFAULT '',
  status      text NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'resolved')),
  created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX delivery_issues_customer_created_idx
  ON delivery_issues (customer_id, created_at DESC);
CREATE INDEX delivery_issues_delivery_idx
  ON delivery_issues (delivery_id);
