CREATE EXTENSION IF NOT EXISTS citext;

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ---------------------------------------------------------------------------
-- Identity
-- ---------------------------------------------------------------------------

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

-- ---------------------------------------------------------------------------
-- Companies and tenancy
-- ---------------------------------------------------------------------------

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

-- ---------------------------------------------------------------------------
-- Roles and permissions (global templates + company-scoped roles)
-- ---------------------------------------------------------------------------

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

-- ---------------------------------------------------------------------------
-- Company configuration
-- ---------------------------------------------------------------------------

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

-- ---------------------------------------------------------------------------
-- Authentication artefacts
-- ---------------------------------------------------------------------------

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

-- ---------------------------------------------------------------------------
-- Audit trail
-- ---------------------------------------------------------------------------

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
