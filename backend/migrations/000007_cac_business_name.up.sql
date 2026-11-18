ALTER TABLE companies
  ADD COLUMN cac_business_name text;

CREATE INDEX companies_review_queue_idx ON companies (status, created_at)
  WHERE deleted_at IS NULL AND status IN ('pending', 'under_review');
