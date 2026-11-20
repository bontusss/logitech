DROP TABLE IF EXISTS delivery_timeline_events;
DROP INDEX IF EXISTS deliveries_tracking_code_unique;
DROP INDEX IF EXISTS deliveries_company_request_status_idx;
ALTER TABLE deliveries
  DROP COLUMN IF EXISTS request_status,
  DROP COLUMN IF EXISTS tracking_code,
  DROP COLUMN IF EXISTS rider_link_token,
  DROP COLUMN IF EXISTS declined_at,
  DROP COLUMN IF EXISTS accepted_at;
