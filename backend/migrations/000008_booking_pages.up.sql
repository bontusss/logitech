CREATE TABLE booking_pages (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id   uuid NOT NULL UNIQUE REFERENCES companies (id) ON DELETE CASCADE,
  slug         citext NOT NULL UNIQUE
               CHECK (slug ~ '^[a-z0-9][a-z0-9-]{1,38}[a-z0-9]$'),
  headline     text,
  description  text,
  is_published boolean NOT NULL DEFAULT true,
  created_at   timestamptz NOT NULL DEFAULT now(),
  updated_at   timestamptz NOT NULL DEFAULT now()
);

CREATE TRIGGER booking_pages_set_updated_at
BEFORE UPDATE ON booking_pages
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
