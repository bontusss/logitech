DROP INDEX IF EXISTS deliveries_customer_created_idx;

ALTER TABLE deliveries
  DROP COLUMN IF EXISTS customer_id;
