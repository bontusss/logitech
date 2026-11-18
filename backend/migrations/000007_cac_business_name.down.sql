DROP INDEX IF EXISTS companies_review_queue_idx;

ALTER TABLE companies
  DROP COLUMN IF EXISTS cac_business_name;
