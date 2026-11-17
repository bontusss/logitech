-- Company management: profile extensions, branches, staff assignments and
-- company subdomains.

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
