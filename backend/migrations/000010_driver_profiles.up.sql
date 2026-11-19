CREATE TABLE driver_profiles (
  id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id        uuid NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
  full_name         text NOT NULL,
  phone             text NOT NULL,
  vehicle_plate     text NOT NULL,
  email             citext NOT NULL,
  portal_token_hash text NOT NULL UNIQUE,
  status            text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended')),
  created_at        timestamptz NOT NULL DEFAULT now(),
  updated_at        timestamptz NOT NULL DEFAULT now(),
  UNIQUE (company_id, email)
);

CREATE INDEX driver_profiles_company_status_idx ON driver_profiles (company_id, status);

CREATE TRIGGER driver_profiles_set_updated_at
BEFORE UPDATE ON driver_profiles
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
