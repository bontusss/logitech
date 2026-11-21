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
